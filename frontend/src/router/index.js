import { createRouter, createWebHashHistory } from 'vue-router'
import { homeRouteForRole, routeAllowedForRole, useAuthStore } from '@/stores/auth.js'

const ROUTE_CHUNK_RELOAD_KEY = 'atigacbt:route-chunk-reload'

const routes = [
  {
    meta: {
      title: 'Dashboard',
    },
    path: '/',
    name: 'root',
    redirect: '/dashboard',
  },
  {
    path: '/auth/google/:type',
    name: 'auth-google-callback',
    component: () => import('@/views/GoogleAuthCallback.vue'),
    meta: {
      title: 'Google Auth',
    },
  },
  {
    path: '/auth/google/register',
    name: 'auth-google-register',
    component: () => import('@/views/GoogleRegistrationForm.vue'),
    meta: {
      title: 'Pendaftaran Google',
    },
  },
  {
    meta: {
      title: 'Dashboard',
    },
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/views/PlaceholderView.vue'),
  },
  {
    meta: {
      title: 'Dashboard Admin',
    },
    path: '/admin/dashboard',
    name: 'admin-dashboard',
    component: () => import('@/views/dashboards/AdminDashboardView.vue'),
  },
  {
    meta: {
      title: 'Log Aktivitas',
    },
    path: '/admin/settings/activity-log',
    name: 'admin-activity-log',
    component: () => import('@/views/admin/AdminActivityLogView.vue'),
  },
  {
    meta: {
      title: 'Audit Log',
    },
    path: '/admin/settings/audit-log',
    name: 'admin-audit-log',
    component: () => import('@/views/admin/AdminAuditLogView.vue'),
  },
  {
    meta: {
      title: 'Master Data',
    },
    path: '/admin/master-data',
    alias: '/admin/master-data/',
    name: 'admin-master-data',
    component: () => import('@/views/admin/MasterDataIndexView.vue'),
  },
  {
    meta: {
      title: 'Menu Master Data',
    },
    path: '/admin/master-data-menu',
    name: 'admin-master-data-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: {
      title: 'Guru',
    },
    path: '/admin/master-data/guru',
    name: 'admin-master-guru',
    component: () => import('@/views/admin/AdminTeachersView.vue'),
  },
  {
    meta: {
      title: 'Siswa',
    },
    path: '/admin/master-data/siswa',
    name: 'admin-master-siswa',
    component: () => import('@/views/admin/AdminStudentsView.vue'),
  },
  {
    meta: {
      title: 'Verifikasi Pendaftaran',
    },
    path: '/admin/master-data/verifikasi-pendaftaran',
    name: 'admin-master-verifikasi-pendaftaran',
    component: () => import('@/views/admin/AdminRegistrationsView.vue'),
  },
  {
    meta: {
      title: 'Pengumuman',
    },
    path: '/admin/pengumuman',
    name: 'admin-announcements',
    component: () => import('@/views/admin/AdminAnnouncementsView.vue'),
  },
  {
    meta: {
      title: 'Program',resourceConfig: {
        endpoint: '/api/v1/admin/programs',
        itemLabel: 'Program',
        fields: [
          { key: 'code', label: 'Kode', placeholder: 'IPA' },
          { key: 'name', label: 'Nama', placeholder: 'Ilmu Pengetahuan Alam' },
        ],
      },
    },
    path: '/admin/master-data/program',
    name: 'admin-master-program',
    component: () => import('@/views/admin/AdminSimpleCrudView.vue'),
  },
  {
    meta: {
      title: 'Level',resourceConfig: {
        endpoint: '/api/v1/admin/levels',
        itemLabel: 'Level',
        fields: [
          { key: 'name', label: 'Nama', placeholder: 'Kelas 10' },
          {
            key: 'kelas',
            label: 'Kelas',
            type: 'select',
            placeholder: 'Pilih kelas',
            options: [
              { value: 1, label: 'Kelas 1' },
              { value: 2, label: 'Kelas 2' },
              { value: 3, label: 'Kelas 3' },
              { value: 4, label: 'Kelas 4' },
              { value: 5, label: 'Kelas 5' },
              { value: 6, label: 'Kelas 6' },
              { value: 7, label: 'Kelas 7' },
              { value: 8, label: 'Kelas 8' },
              { value: 9, label: 'Kelas 9' },
              { value: 10, label: 'Kelas 10' },
              { value: 11, label: 'Kelas 11' },
              { value: 12, label: 'Kelas 12' },
            ],
          },
        ],
      },
    },
    path: '/admin/master-data/level',
    name: 'admin-master-level',
    component: () => import('@/views/admin/AdminSimpleCrudView.vue'),
  },
  {
    meta: {
      title: 'Group',resourceConfig: {
        endpoint: '/api/v1/admin/groups',
        itemLabel: 'Group',
        fields: [{ key: 'name', label: 'Nama', placeholder: 'X IPA 1' }],
      },
    },
    path: '/admin/master-data/group',
    name: 'admin-master-group',
    component: () => import('@/views/admin/AdminSimpleCrudView.vue'),
  },
  {
    meta: {
      title: 'Mata Pelajaran',resourceConfig: {
        endpoint: '/api/v1/admin/subjects',
        itemLabel: 'Mata Pelajaran',
        fields: [
          { key: 'code', label: 'Kode', placeholder: 'MTK' },
          { key: 'name', label: 'Nama', placeholder: 'Matematika' },
        ],
      },
    },
    path: '/admin/master-data/mata-pelajaran',
    name: 'admin-master-mapel',
    component: () => import('@/views/admin/AdminSimpleCrudView.vue'),
  },
  {
    meta: {
      title: 'Sesi',resourceConfig: {
        endpoint: '/api/v1/admin/sessions',
        itemLabel: 'Sesi',
        fields: [
          { key: 'name', label: 'Nama', placeholder: 'Sesi 1' },
          { key: 'start_time', label: 'Jam Mulai', placeholder: '07:30' },
          { key: 'end_time', label: 'Jam Selesai', placeholder: '09:30' },
        ],
      },
    },
    path: '/admin/master-data/sesi',
    name: 'admin-master-sesi',
    component: () => import('@/views/admin/AdminSimpleCrudView.vue'),
  },
  {
    meta: {
      title: 'Semua Bank Soal',
    },
    path: '/admin/bank-soal-menu',
    name: 'admin-bank-soal-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: {
      title: 'Menu Ujian',
    },
    path: '/admin/ujian-menu',
    name: 'admin-ujian-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: {
      title: 'Menu Evaluasi',
    },
    path: '/admin/evaluasi-menu',
    name: 'admin-evaluasi-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: {
      title: 'Menu Cetak',
    },
    path: '/admin/cetak-menu',
    name: 'admin-cetak-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: {
      title: 'Menu Config',
    },
    path: '/admin/config-menu',
    name: 'admin-config-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: {
      title: 'Semua Bank Soal',
    },
    path: '/admin/bank-soal',
    name: 'admin-bank-soal',
    component: () => import('@/views/QuestionBankView.vue'),
  },
  {
    meta: {
      title: 'Buat Bank Soal',
    },
    path: '/admin/bank-soal/new/:editorType?',
    name: 'admin-bank-soal-new',
    component: () => import('@/views/QuestionBankNewView.vue'),
  },
  {
    meta: {
      title: 'Impor Soal',
    },
    path: '/admin/bank-soal/import',
    name: 'admin-bank-soal-import',
    component: () => import('@/views/QuestionBankImportView.vue'),
  },
  {
    meta: { title: 'Pratinjau Bank Soal' },
    path: '/admin/bank-soal/preview/:id',
    name: 'admin-bank-soal-preview',
    component: () => import('@/views/QuestionPreviewView.vue'),
  },
  {
    meta: {
      title: 'Jadwal Ujian',
    },
    path: '/admin/ujian/jadwal',
    name: 'admin-ujian-jadwal',
    component: () => import('@/views/ExamsView.vue'),
  },
  {
    meta: {
      title: 'Token',
    },
    path: '/admin/ujian/token',
    name: 'admin-ujian-token',
    component: () => import('@/views/ExamTokensView.vue'),
  },
  {
    meta: {
      title: 'Monitor Ujian',
    },
    path: '/admin/ujian/monitor-ujian',
    name: 'admin-ujian-monitor',
    component: () => import('@/views/monitor/ExamMonitorView.vue'),
  },
  {
    meta: {
      title: 'Monitor Peserta',
    },
    path: '/admin/ujian/monitor-peserta',
    name: 'admin-ujian-monitor-peserta',
    component: () => import('@/views/monitor/ParticipantMonitorView.vue'),
  },
  {
    meta: {
      title: 'Reset Login',
    },
    path: '/admin/ujian/reset-login',
    name: 'admin-ujian-reset-login',
    component: () => import('@/views/ResetLoginView.vue'),
  },
  {
    meta: {
      title: 'Evaluasi / Hasil Nilai',
    },
    path: '/admin/evaluasi',
    name: 'admin-evaluasi',
    component: () => import('@/views/EvaluationView.vue'),
  },
  {
    meta: {
      title: 'Advanced Analytics',
    },
    path: '/admin/analytics',
    name: 'admin-analytics',
    component: () => import('@/views/admin/AdminAnalyticsView.vue'),
  },
  {
    meta: {
      title: 'Cetak',
    },
    path: '/admin/cetak',
    name: 'admin-cetak',
    component: () => import('@/views/admin/AdminPrintView.vue'),
  },
  {
    meta: {
      title: 'Config / Settings',
    },
    path: '/admin/settings',
    name: 'admin-settings',
    component: () => import('@/views/admin/AdminSettingsView.vue'),
  },
  {
    meta: {
      title: 'Integrasi LMS',
    },
    path: '/admin/lms',
    name: 'admin-lms',
    component: () => import('@/views/admin/AdminLMSView.vue'),
  },

  // Teacher
  {
    meta: { title: 'Dashboard Guru'},
    path: '/teacher/dashboard',
    name: 'teacher-dashboard',
    component: () => import('@/views/dashboards/TeacherDashboardView.vue'),
  },
  {
    meta: { title: 'Semua Bank Soal'},
    path: '/teacher/bank-soal-menu',
    name: 'teacher-bank-soal-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: { title: 'Menu Ujian'},
    path: '/teacher/ujian-menu',
    name: 'teacher-ujian-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: { title: 'Menu Evaluasi'},
    path: '/teacher/evaluasi-menu',
    name: 'teacher-evaluasi-menu',
    component: () => import('@/views/SubmenuIndexView.vue'),
  },
  {
    meta: { title: 'Semua Bank Soal'},
    path: '/teacher/bank-soal',
    name: 'teacher-bank-soal',
    component: () => import('@/views/QuestionBankView.vue'),
  },
  {
    meta: { title: 'Buat Bank Soal' },
    path: '/teacher/bank-soal/new/:editorType?',
    name: 'teacher-bank-soal-new',
    component: () => import('@/views/QuestionBankNewView.vue'),
  },
  {
    meta: { title: 'Impor Soal'},
    path: '/teacher/bank-soal/import',
    name: 'teacher-bank-soal-import',
    component: () => import('@/views/QuestionBankImportView.vue'),
  },
  {
    meta: { title: 'Pratinjau Bank Soal' },
    path: '/teacher/bank-soal/preview/:id',
    name: 'teacher-bank-soal-preview',
    component: () => import('@/views/QuestionPreviewView.vue'),
  },
  {
    meta: { title: 'Jadwal Ujian'},
    path: '/teacher/ujian/jadwal',
    name: 'teacher-ujian-jadwal',
    component: () => import('@/views/ExamsView.vue'),
  },
  {
    meta: { title: 'Token'},
    path: '/teacher/ujian/token',
    name: 'teacher-ujian-token',
    component: () => import('@/views/ExamTokensView.vue'),
  },
  {
    meta: { title: 'Monitor Ujian'},
    path: '/teacher/ujian/monitor-ujian',
    name: 'teacher-ujian-monitor',
    component: () => import('@/views/monitor/ExamMonitorView.vue'),
  },
  {
    meta: { title: 'Monitor Peserta'},
    path: '/teacher/ujian/monitor-peserta',
    name: 'teacher-ujian-monitor-peserta',
    component: () => import('@/views/monitor/ParticipantMonitorView.vue'),
  },
  {
    meta: { title: 'Absensi Peserta'},
    path: '/teacher/ujian/absensi',
    name: 'teacher-ujian-absensi',
    component: () => import('@/views/monitor/AttendanceMonitorView.vue'),
  },
  {
    meta: { title: 'Generate QR Absensi' },
    path: '/teacher/ujian/absensi/qr/:id',
    name: 'teacher-ujian-absensi-qr',
    component: () => import('@/views/monitor/AttendanceQRView.vue'),
  },
  {
    meta: { title: 'Reset Login'},
    path: '/teacher/ujian/reset-login',
    name: 'teacher-ujian-reset-login',
    component: () => import('@/views/ResetLoginView.vue'),
  },
  {
    meta: { title: 'Evaluasi'},
    path: '/teacher/evaluasi',
    name: 'teacher-evaluasi',
    component: () => import('@/views/EvaluationView.vue'),
  },
  {
    meta: { title: 'Profil'},
    path: '/teacher/profil',
    name: 'teacher-profil',
    component: () => import('@/views/ProfileView.vue'),
  },
  {
    meta: { title: 'LTI Picker' },
    path: '/teacher/lti/picker',
    name: 'teacher-lti-picker',
    component: () => import('@/views/teacher/LTIPickerView.vue'),
  },

  // Student
  {
    meta: { title: 'Dashboard Siswa'},
    path: '/student/dashboard',
    name: 'student-dashboard',
    component: () => import('@/views/dashboards/StudentDashboardView.vue'),
  },
  {
    meta: { title: 'Daftar Ujian'},
    path: '/student/ujian',
    name: 'student-ujian',
    component: () => import('@/views/student/StudentExamsView.vue'),
  },
  {
    meta: { title: 'Token Ujian'},
    path: '/student/ujian/:examId/token',
    name: 'student-ujian-token-gate',
    component: () => import('@/views/student/StudentExamTokenGateView.vue'),
  },
  {
    meta: { title: 'Kerjakan Ujian'},
    path: '/student/kerjakan',
    name: 'student-kerjakan-home',
    redirect: '/student/ujian',
  },
  {
    meta: { title: 'Kerjakan Ujian'},
    path: '/student/workspace/:sessionId',
    name: 'student-workspace',
    component: () => import('@/views/student/StudentExamWorkspaceView.vue'),
  },
  {
    meta: { title: 'Hasil Ujian'},
    path: '/student/hasil',
    name: 'student-hasil',
    component: () => import('@/views/ResultsView.vue'),
  },
  {
    meta: { title: 'Pengumuman'},
    path: '/student/pengumuman',
    name: 'student-pengumuman',
    component: () => import('@/views/student/StudentAnnouncementsView.vue'),
  },
  {
    meta: { title: 'Profil'},
    path: '/student/profil',
    name: 'student-profil',
    component: () => import('@/views/ProfileView.vue'),
  },

  // Keep template pages (not in menu) for now
  {
    meta: { title: 'Profile' },
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/ProfileView.vue'),
  },
  {
    meta: { title: 'Login' },
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
  },
  {
    meta: { title: 'Lupa Kata Sandi' },
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('@/views/ForgotPasswordView.vue'),
  },
  {
    meta: { title: 'Atur Ulang Kata Sandi' },
    path: '/auth/reset-password',
    name: 'reset-password',
    component: () => import('@/views/ResetPasswordView.vue'),
  },
  {
    meta: { title: 'Error' },
    path: '/error',
    name: 'error',
    component: () => import('@/views/ErrorView.vue'),
  },
  // Fallback: prevent blank page on unknown hashes/old routes
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    return savedPosition || { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  const isPublicRoute = to.path === '/login' ||
    to.path === '/forgot-password' ||
    to.path === '/auth/reset-password' ||
    to.path === '/error' ||
    to.path.startsWith('/auth/google/')

  if (authStore.token && !authStore.user) {
    await authStore.loadMe()
  }

  if (!authStore.isAuthenticated && !isPublicRoute) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (authStore.isAuthenticated && to.path === '/login') {
    return homeRouteForRole(authStore.role)
  }

  if (to.path === '/' || to.path === '/dashboard') {
    return homeRouteForRole(authStore.role)
  }

  if (authStore.isAuthenticated && !routeAllowedForRole(authStore.role, to.path)) {
    return homeRouteForRole(authStore.role)
  }

  return true
})

router.onError((error, to) => {
  const message = String(error?.message || '')
  const isChunkLoadError =
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Importing a module script failed') ||
    message.includes('Unable to preload CSS for') ||
    message.includes('Loading chunk')

  if (!isChunkLoadError || typeof window === 'undefined') {
    console.error('Router navigation error:', error)
    return
  }

  const currentTarget = String(to?.fullPath || window.location.hash || '')
  const previousTarget = sessionStorage.getItem(ROUTE_CHUNK_RELOAD_KEY) || ''
  if (previousTarget === currentTarget) {
    sessionStorage.removeItem(ROUTE_CHUNK_RELOAD_KEY)
    console.error('Router chunk reload retry failed:', error)
    return
  }

  sessionStorage.setItem(ROUTE_CHUNK_RELOAD_KEY, currentTarget)
  window.location.reload()
})

export default router
