package httpapi

import (
	"net/http"
	httppprof "net/http/pprof"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"mycbt/backend/internal/config"
	"mycbt/backend/internal/httpapi/handlers"
	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/repo/auditrepo"
	"mycbt/backend/internal/repo/examrepo"
	"mycbt/backend/internal/repo/loginlogrepo"
	"mycbt/backend/internal/repo/ltirepo"
	"mycbt/backend/internal/repo/masterrepo"
	"mycbt/backend/internal/repo/questionbankrepo"
	"mycbt/backend/internal/repo/studentexamrepo"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
	"mycbt/backend/internal/service/ltisvc"
	"mycbt/backend/internal/service/notificationsvc"
)

type Deps struct {
	Config config.Config
	Auth   *authsvc.Service
	Users  *userrepo.Repo
	Pool   *pgxpool.Pool
}

func NewHandler(deps Deps) http.Handler {
	cfg := deps.Config
	gin.SetMode(cfg.Env)
	const (
		maxImportExcelBody = 8 << 20  // 8 MB
		maxImportDocxBody  = 24 << 20 // 24 MB
		maxUploadImageBody = 8 << 20  // 8 MB
	)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.SecurityHeaders())
	if cfg.Env == "debug" || cfg.Env == "" {
		pp := r.Group("/debug/pprof")
		pp.GET("/", gin.WrapF(httppprof.Index))
		pp.GET("/cmdline", gin.WrapF(httppprof.Cmdline))
		pp.GET("/profile", gin.WrapF(httppprof.Profile))
		pp.POST("/symbol", gin.WrapF(httppprof.Symbol))
		pp.GET("/symbol", gin.WrapF(httppprof.Symbol))
		pp.GET("/trace", gin.WrapF(httppprof.Trace))
		pp.GET("/allocs", gin.WrapH(httppprof.Handler("allocs")))
		pp.GET("/block", gin.WrapH(httppprof.Handler("block")))
		pp.GET("/goroutine", gin.WrapH(httppprof.Handler("goroutine")))
		pp.GET("/heap", gin.WrapH(httppprof.Handler("heap")))
		pp.GET("/mutex", gin.WrapH(httppprof.Handler("mutex")))
		pp.GET("/threadcreate", gin.WrapH(httppprof.Handler("threadcreate")))
	}
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":   "MyCBT API",
			"status": "running",
			"env":    cfg.Env,
		})
	})
	if deps.Pool != nil {
		r.Use(middleware.AuditLogger(auditrepo.New(deps.Pool)))
	}

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	if cfg.Env == "debug" || cfg.Env == "" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = splitCSV(cfg.CORSOrigins)
	}
	r.Use(cors.New(corsConfig))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1")
	var notifSvc *notificationsvc.Service
	if deps.Pool != nil {
		notifSvc = notificationsvc.New(masterrepo.NewSettings(deps.Pool), cfg)
	}

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Public LTI endpoints
	if deps.Pool != nil && deps.Auth != nil && deps.Users != nil {
		ltiH := handlers.NewLTIHandler(ltirepo.New(deps.Pool), deps.Users, deps.Auth, cfg.AppURL)
		v1.GET("/lti/login", ltiH.LoginInitiation)
		v1.POST("/lti/launch", ltiH.Launch)
	}

	// Login
	if deps.Auth != nil && deps.Users != nil {
		var loginLogs *loginlogrepo.Repo
		if deps.Pool != nil {
			loginLogs = loginlogrepo.New(deps.Pool)
		}
		authH := handlers.NewAuthHandler(deps.Auth, deps.Users, loginLogs)
		v1.POST("/auth/login", middleware.RateLimit("auth_login", 10, time.Minute), authH.Login)
		v1.POST("/auth/logout", middleware.RequireAuth(deps.Auth), authH.Logout)
		v1.GET("/me", middleware.RequireAuth(deps.Auth), authH.Me)
		v1.POST("/me/photo", middleware.RequireAuth(deps.Auth), authH.UploadPhoto)
	}

	// Password Reset
	if deps.Pool != nil && deps.Users != nil && notifSvc != nil {
		resetTokenRepo := userrepo.NewPasswordResetRepo(deps.Pool)
		resetSvc := authsvc.NewPasswordResetService(deps.Users, resetTokenRepo, notifSvc, cfg.AppURL)
		resetH := handlers.NewPasswordResetHandler(resetSvc)
		v1.POST("/auth/forgot-password", resetH.ForgotPassword)
		v1.POST("/auth/reset-password", resetH.ResetPassword)
	}

	// Google OAuth
	if deps.Auth != nil && deps.Users != nil && deps.Pool != nil {
		googleH := handlers.NewGoogleAuthHandler(cfg, deps.Auth, deps.Users, masterrepo.NewRegistrations(deps.Pool))
		v1.GET("/auth/google/redirect", googleH.Redirect)
		v1.GET("/auth/google/callback", googleH.Callback)
		v1.POST("/auth/google/register", googleH.SubmitRegistration)
	}

	// Public registration (pending approval)
	if deps.Pool != nil {
		regs := masterrepo.NewRegistrations(deps.Pool)
		pub := handlers.NewRegistrationPublicHandler(regs, deps.Users)
		v1.POST("/registrations", pub.Register)
		v1.GET("/registrations/:id", pub.Status)
	}

	// Question bank (admin/teacher)
	if deps.Auth != nil && deps.Pool != nil {
		qb := questionbankrepo.New(deps.Pool)
		teacherSubs := masterrepo.NewTeacherSubjects(deps.Pool)
		h := handlers.NewQuestionBankHandler(qb, teacherSubs)
		up := handlers.NewUploadsHandler()

		qg := v1.Group("")
		qg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))

		qg.GET("/question-sets", h.ListSets)
		qg.POST("/question-sets", h.CreateSet)
		qg.GET("/question-sets/:id", h.GetSet)
		qg.PATCH("/question-sets/:id", h.PatchSet)
		qg.DELETE("/question-sets/:id", h.DeleteSet)
		qg.POST("/question-sets/:id/clone", h.CloneSet)
		qg.GET("/question-sets/:id/questions", h.ListQuestions)
		qg.POST("/question-sets/:id/questions", h.CreateQuestion)
		qg.POST("/question-sets/:id/import-docx/preview", middleware.LimitBodyBytes(maxImportDocxBody), h.ImportDocxPreview)
		qg.POST("/question-sets/:id/import-docx", middleware.LimitBodyBytes(maxImportDocxBody), h.ImportDocx)
		qg.GET("/questions/:id", h.GetQuestion)
		qg.DELETE("/questions/:id", h.DeleteQuestion)
		qg.PATCH("/questions/:id", h.PatchQuestion)

		// Upload assets for editor content (RichEditor images, etc.)
		qg.POST("/uploads/images", middleware.LimitBodyBytes(maxUploadImageBody), up.UploadImage)
	}

	// Exams (admin/teacher)
	if deps.Auth != nil && deps.Pool != nil {
		ex := examrepo.New(deps.Pool)
		lti := ltirepo.New(deps.Pool)
		teacherSubs := masterrepo.NewTeacherSubjects(deps.Pool)
		settingsRepo := masterrepo.NewSettings(deps.Pool)

		teacherGroups := masterrepo.NewTeacherGroups(deps.Pool)
		teacherLevels := masterrepo.NewTeacherLevels(deps.Pool)

		h := handlers.NewExamsHandler(ex, teacherSubs, teacherGroups, teacherLevels)
		mon := handlers.NewMonitorHandler(ex, studentexamrepo.New(deps.Pool))
		res := handlers.NewExamResultsHandler(ex, lti, ltisvc.NewAGSService(nil), studentexamrepo.New(deps.Pool), notifSvc)
		reset := handlers.NewResetLoginHandler(ex, studentexamrepo.New(deps.Pool), settingsRepo)

		eg := v1.Group("")
		eg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))

		eg.GET("/exams", h.List)
		eg.POST("/exams", h.Create)
		eg.GET("/exams/:id", h.Get)
		eg.PATCH("/exams/:id", h.Patch)
		eg.DELETE("/exams/:id", h.Delete)

		eg.GET("/exams/:id/targets", h.GetTargets)
		eg.PUT("/exams/:id/targets", h.PutTargets)
		eg.GET("/exams/:id/question-sets", h.GetQuestionSets)
		eg.PUT("/exams/:id/question-sets", h.PutQuestionSets)

		eg.GET("/exams/:id/tokens", h.ListTokens)
		eg.POST("/exams/:id/tokens", h.CreateToken)
		eg.POST("/exams/:id/tokens/deactivate-all", h.DeactivateAllTokens)
		eg.POST("/exams/:id/tokens/rotate", h.RotateToken)
		eg.PATCH("/tokens/:id", h.PatchToken)
		eg.DELETE("/tokens/:id", h.DeleteToken)

		eg.GET("/exams/:id/results", res.List)
		eg.POST("/exams/:id/results/blast", res.BlastResults)
		eg.POST("/exams/:id/lti/sync-scores", res.SyncLTIScores)
		eg.GET("/exams/:id/item-analysis", res.ItemAnalysis)
		eg.GET("/exams/:id/item-analysis/suggestions", res.ItemSuggestions)
		eg.GET("/exams/:id/score-distribution", res.ScoreDistribution)
		eg.GET("/exams/:id/export", res.Export)
		eg.GET("/exams/:id/item-analysis/export", res.ExportItemAnalysis)
		eg.GET("/exams/:id/attendance", res.Attendance)
		eg.GET("/exams/:id/sessions/:sessionId/essays", res.ListEssays)
		eg.POST("/exams/:id/sessions/:sessionId/essays/score", res.SaveEssayScore)
		eg.GET("/exams/:id/monitor/sessions", mon.ListSessions)
		eg.GET("/exams/:id/monitor/participants", mon.ListParticipants)
		eg.POST("/exams/:id/sessions/:sessionId/reset", reset.ResetSession)
		eg.POST("/exams/:id/sessions/:sessionId/force-submit", reset.ForceSubmitSession)
	}

	// Monitor stream (SSE): allow query token for EventSource.
	if deps.Auth != nil && deps.Pool != nil {
		ex := examrepo.New(deps.Pool)
		mon := handlers.NewMonitorHandler(ex, studentexamrepo.New(deps.Pool))

		mg := v1.Group("")
		mg.Use(middleware.RequireAuthHeaderOrQueryToken(deps.Auth), middleware.RequireRole("admin", "teacher"))
		mg.GET("/exams/:id/monitor/stream", mon.Stream)
	}

	// Student exam room
	if deps.Auth != nil && deps.Pool != nil {
		st := studentexamrepo.New(deps.Pool)
		h := handlers.NewStudentExamHandler(st, masterrepo.NewSettings(deps.Pool))
		nh := handlers.NewNotificationHandler(masterrepo.NewAnnouncements(deps.Pool), st)

		sg := v1.Group("/student")
		sg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("student"))

		sg.GET("/notifications/stream", middleware.RequireAuthHeaderOrQueryToken(deps.Auth), nh.Stream)
		sg.GET("/exams", h.ListExams)
		sg.GET("/exams/:id/session", h.GetActiveSessionByExam)
		sg.POST("/exams/:id/join", middleware.RateLimit("student_join_exam", 20, time.Minute), h.Join)
		sg.GET("/sessions/:id", h.GetSession)
		sg.POST("/sessions/:id/verify-token", h.VerifyToken)
		sg.GET("/sessions/:id/questions", h.GetQuestions)
		sg.GET("/sessions/:id/answers", h.GetAnswers)
		sg.POST("/sessions/:id/answers", h.SaveAnswer)
		sg.POST("/sessions/:id/submit", h.Submit)
		sg.POST("/sessions/:id/heartbeat", h.Heartbeat)
		sg.GET("/results", h.ListResults)
		sg.GET("/announcements", h.ListAnnouncements)
		sg.POST("/attendance", h.SubmitAttendance)
		sg.GET("/attendance/history", h.ListAttendanceHistory)

		// QR Scan attendance
		attendanceSessRepo := masterrepo.NewAttendanceSessions(deps.Pool)
		qrH := handlers.NewAttendanceHandler(attendanceSessRepo, st, masterrepo.NewSettings(deps.Pool))
		sg.POST("/attendance/scan", qrH.ScanAttendance)

		// Compatibility endpoints for legacy CBT flow contracts.
		compat := v1.Group("")
		compat.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("student"))
		compat.POST("/exams/:id/start", middleware.RateLimit("student_join_exam", 20, time.Minute), h.StartCompat)
		compat.GET("/exams/:id/questions", h.GetQuestionsByExamCompat)
		compat.POST("/sessions/:session_id/answers", h.SaveAnswerCompat)
		compat.POST("/sessions/:session_id/finish", h.FinishCompat)
	}

	// Lookups (admin/teacher)
	if deps.Auth != nil && deps.Pool != nil {
		h := handlers.NewLookupsHandler(
			masterrepo.NewSubjects(deps.Pool),
			masterrepo.NewSessions(deps.Pool),
			masterrepo.NewLevels(deps.Pool),
			masterrepo.NewGroups(deps.Pool),
			masterrepo.NewStudents(deps.Pool),
			masterrepo.NewTeachers(deps.Pool),
			masterrepo.NewPrograms(deps.Pool),
		)
		// Lookups
		lup := v1.Group("/lookups")
		{
			// Public registration lookups
			lup.GET("/programs", h.ListPrograms)
			lup.GET("/levels", h.ListLevels)
			lup.GET("/groups", h.ListGroups)

			// Authenticated shared lookups
			authLup := lup.Group("")
			authLup.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))
			authLup.GET("/subjects", h.ListSubjects)
			authLup.GET("/sessions", h.ListSessions)
			authLup.GET("/students", h.ListStudents)
			authLup.GET("/teachers", h.ListTeachers)
			authLup.GET("/my-assignments", h.ListMyAssignments)
		}
	}

	// Settings (admin)
	if deps.Auth != nil && deps.Pool != nil {
		sg := v1.Group("/settings")
		sg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"))

		settings := masterrepo.NewSettings(deps.Pool)
		h := handlers.NewSettingsHandler(settings)

		sg.GET("/school-identity", h.GetSchoolIdentity)
		sg.PUT("/school-identity", h.PutSchoolIdentity)
		sg.POST("/school-identity/logo", middleware.LimitBodyBytes(4<<20), h.UploadSchoolLogo)
		sg.GET("/system", h.GetSystem)
		sg.PUT("/system", h.PutSystem)
		sg.GET("/smtp", h.GetSMTP)
		sg.PUT("/smtp", h.PutSMTP)
		sg.GET("/whatsapp", h.GetWhatsApp)
		sg.PUT("/whatsapp", h.PutWhatsApp)
	}

	// Maintenance (admin)
	if deps.Auth != nil && deps.Pool != nil {
		mg := v1.Group("/maintenance")
		mg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"))

		h := handlers.NewMaintenanceHandler(deps.Config.DatabaseURL)
		mg.GET("/backup", h.Backup)
		mg.POST("/restore", h.Restore)
	}

	// LMS / Data Portability (admin/teacher)
	if deps.Auth != nil && deps.Pool != nil {
		lmsH := handlers.NewLMSExportHandler(deps.Pool, studentexamrepo.New(deps.Pool))
		lg := v1.Group("/lms")
		lg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))
		lg.GET("/summary", lmsH.Summary)
		lg.GET("/exams", lmsH.ListExams)
		lg.GET("/export/students", lmsH.ExportStudents)
		lg.GET("/export/results", lmsH.ExportResults)
	}

	// Analytics (admin/teacher)
	if deps.Auth != nil && deps.Pool != nil {
		h := handlers.NewAnalyticsHandler(deps.Pool, studentexamrepo.New(deps.Pool))
		ag := v1.Group("/analytics")
		ag.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))
		ag.GET("/dashboard", h.Dashboard)
	}

	// LTI Management
	if deps.Pool != nil && deps.Auth != nil && deps.Users != nil {
		h := handlers.NewLTIHandler(ltirepo.New(deps.Pool), deps.Users, deps.Auth, cfg.AppURL)
		lg := v1.Group("/lti")
		// Management is for admins, Deep Linking is for teachers (and admins)
		lg.GET("/platforms", middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"), h.ListPlatforms)
		lg.POST("/platforms", middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"), h.CreatePlatform)
		lg.GET("/platforms/:id", middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"), h.GetPlatform)
		lg.POST("/keys/generate", middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"), h.GenerateKeys)

		lg.POST("/deep-link", middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"), h.SubmitDeepLink)
	}

	// Shared Announcements (admin/teacher)
	if deps.Auth != nil && deps.Pool != nil {
		ag := v1.Group("/announcements")
		ag.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))

		h := handlers.NewAdminMasterHandler(deps.Pool, deps.Users,
			masterrepo.NewPrograms(deps.Pool),
			masterrepo.NewLevels(deps.Pool),
			masterrepo.NewGroups(deps.Pool),
			masterrepo.NewSubjects(deps.Pool),
			masterrepo.NewAnnouncements(deps.Pool),
			masterrepo.NewTeacherSubjects(deps.Pool),
			masterrepo.NewTeacherGroups(deps.Pool),
			masterrepo.NewTeacherLevels(deps.Pool),
			masterrepo.NewTeachers(deps.Pool),
			masterrepo.NewStudents(deps.Pool),
			masterrepo.NewRegistrations(deps.Pool),
			masterrepo.NewSessions(deps.Pool),
			masterrepo.NewLookups(deps.Pool),
			notifSvc,
		)

		ag.GET("", h.ListAnnouncements)
		ag.POST("", h.CreateAnnouncement)
		ag.POST("/:id/blast", h.BlastAnnouncement)
		ag.GET("/:id", h.GetAnnouncement)
		ag.PATCH("/:id", h.UpdateAnnouncement)
		ag.DELETE("/:id", h.DeleteAnnouncement)
	}

	// Attendance Sessions (QR)
	if deps.Auth != nil && deps.Pool != nil {
		attendanceSessRepo := masterrepo.NewAttendanceSessions(deps.Pool)
		h := handlers.NewAttendanceHandler(attendanceSessRepo, studentexamrepo.New(deps.Pool), masterrepo.NewSettings(deps.Pool))

		tg := v1.Group("")
		tg.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin", "teacher"))
		tg.POST("/exams/:id/attendance-session", h.CreateSession)
		tg.GET("/exams/:id/attendance-session", h.ListActiveSessions)
	}

	// Admin master data
	if deps.Auth != nil && deps.Pool != nil {
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireAuth(deps.Auth), middleware.RequireRole("admin"))

		programs := masterrepo.NewPrograms(deps.Pool)
		levels := masterrepo.NewLevels(deps.Pool)
		groups := masterrepo.NewGroups(deps.Pool)
		subjects := masterrepo.NewSubjects(deps.Pool)
		announcements := masterrepo.NewAnnouncements(deps.Pool)
		teacherSubs := masterrepo.NewTeacherSubjects(deps.Pool)
		teacherGroups := masterrepo.NewTeacherGroups(deps.Pool)
		teacherLevels := masterrepo.NewTeacherLevels(deps.Pool)
		teachers := masterrepo.NewTeachers(deps.Pool)
		students := masterrepo.NewStudents(deps.Pool)
		regs := masterrepo.NewRegistrations(deps.Pool)
		sessions := masterrepo.NewSessions(deps.Pool)
		lookups := masterrepo.NewLookups(deps.Pool)

		h := handlers.NewAdminMasterHandler(deps.Pool, deps.Users, programs, levels, groups, subjects, announcements, teacherSubs, teacherGroups, teacherLevels, teachers, students, regs, sessions, lookups, notifSvc)

		admin.GET("/dashboard/stats", h.GetDashboardStats)
		admin.GET("/programs", h.ListPrograms)
		admin.POST("/programs", h.CreateProgram)
		admin.GET("/programs/:id", h.GetProgram)
		admin.PATCH("/programs/:id", h.UpdateProgram)
		admin.DELETE("/programs/:id", h.DeleteProgram)
		admin.GET("/levels", h.ListLevels)
		admin.POST("/levels", h.CreateLevel)
		admin.GET("/levels/:id", h.GetLevel)
		admin.PATCH("/levels/:id", h.UpdateLevel)
		admin.DELETE("/levels/:id", h.DeleteLevel)
		admin.GET("/groups", h.ListGroups)
		admin.POST("/groups", h.CreateGroup)
		admin.GET("/groups/:id", h.GetGroup)
		admin.PATCH("/groups/:id", h.UpdateGroup)
		admin.DELETE("/groups/:id", h.DeleteGroup)
		admin.GET("/subjects", h.ListSubjects)
		admin.POST("/subjects", h.CreateSubject)
		admin.GET("/subjects/:id", h.GetSubject)
		admin.PATCH("/subjects/:id", h.UpdateSubject)
		admin.DELETE("/subjects/:id", h.DeleteSubject)

		admin.GET("/sessions", h.ListSessions)
		admin.POST("/sessions", h.CreateSession)
		admin.GET("/sessions/:id", h.GetSession)
		admin.PATCH("/sessions/:id", h.UpdateSession)
		admin.DELETE("/sessions/:id", h.DeleteSession)

		admin.GET("/teachers", h.ListTeachers)
		admin.POST("/teachers", h.CreateTeacher)
		admin.GET("/teachers/:id", h.GetTeacher)
		admin.GET("/teachers/:id/subjects", h.GetTeacherSubjects)
		admin.PUT("/teachers/:id/subjects", h.SetTeacherSubjects)
		admin.GET("/teachers/:id/groups", h.GetTeacherGroups)
		admin.PUT("/teachers/:id/groups", h.SetTeacherGroups)
		admin.GET("/teachers/:id/levels", h.GetTeacherLevels)
		admin.PUT("/teachers/:id/levels", h.SetTeacherLevels)
		admin.GET("/teachers/template", h.TeachersTemplate)
		admin.POST("/teachers/import", middleware.LimitBodyBytes(maxImportExcelBody), h.ImportTeachers)
		admin.PATCH("/teachers/:id", h.UpdateTeacher)
		admin.DELETE("/teachers/:id", h.DeleteTeacher)

		admin.GET("/students", h.ListStudents)
		admin.POST("/students", h.CreateStudent)
		admin.GET("/students/:id", h.GetStudent)
		admin.GET("/students/template", h.StudentsTemplate)
		admin.POST("/students/import", middleware.LimitBodyBytes(maxImportExcelBody), h.ImportStudents)
		admin.PATCH("/students/:id", h.UpdateStudent)
		admin.DELETE("/students/:id", h.DeleteStudent)

		// Switch role between teacher <-> student
		admin.POST("/users/:id/switch-role", h.SwitchUserRole)

		admin.GET("/registrations/pending", h.ListPendingRegistrations)
		admin.GET("/registrations", h.ListRegistrations)
		admin.POST("/registrations/approve-bulk", h.BulkApproveRegistrations)
		admin.GET("/registrations/:id", h.GetRegistration)
		admin.PATCH("/registrations/:id", h.PatchRegistration)
		admin.POST("/registrations/:id/approve", h.ApproveRegistration)
		admin.POST("/registrations/:id/reject", h.RejectRegistration)
		admin.POST("/registrations/:id/pending", h.ResetRegistration)

		// Activity logs (login)
		ll := handlers.NewLoginLogsHandler(loginlogrepo.New(deps.Pool))
		admin.GET("/login-logs", ll.List)
		admin.DELETE("/login-logs", ll.ClearAll)
		admin.DELETE("/login-logs/prune", ll.Prune)
		admin.DELETE("/login-logs/:id", ll.Delete)

		// Audit logs (mutations by admin/teacher)
		al := handlers.NewAuditLogsHandler(auditrepo.New(deps.Pool))
		admin.GET("/audit-logs", al.List)
		admin.GET("/audit-logs/export", al.ExportCSV)
		admin.DELETE("/audit-logs", al.ClearAll)
		admin.DELETE("/audit-logs/prune", al.Prune)
		admin.DELETE("/audit-logs/:id", al.Delete)
	}

	// Serve uploaded assets (logo sekolah, dll) from local storage.
	r.Static("/uploads", "./uploads")
	return r
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"http://localhost:5173"}
	}
	return out
}
