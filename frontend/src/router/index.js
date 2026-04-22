import { createRouter, createWebHashHistory } from 'vue-router'
import { homeRouteForRole, routeAllowedForRole, useAuthStore } from '@/stores/auth.js'
import Placeholder from '@/views/PlaceholderView.vue'

import {
  mdiHomeOutline,
  mdiFolderOutline,
  mdiAccountTie,
  mdiAccountSchool,
  mdiAccountCheckOutline,
  mdiBullhornOutline,
  mdiSchoolOutline,
  mdiLayersTripleOutline,
  mdiAccountGroupOutline,
  mdiBookEducationOutline,
  mdiBookOpenVariant,
  mdiFileDocumentOutline,
  mdiClipboardTextOutline,
  mdiCalendarClockOutline,
  mdiKeyVariant,
  mdiMonitorEye,
  mdiAccountSearchOutline,
  mdiCalendarCheckOutline,
  mdiAccountSwitchOutline,
  mdiChartBoxOutline,
  mdiPrinterOutline,
  mdiCogOutline,
  mdiDatabaseExportOutline,
  mdiChartLine,
  mdiLinkVariant,
  mdiHistory,
  mdiClipboardTextSearchOutline,
} from '@mdi/js'

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
    component: Placeholder,
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
      icon: mdiHistory,
    },
    path: '/admin/settings/activity-log',
    name: 'admin-activity-log',
    component: () => import('@/views/admin/AdminActivityLogView.vue'),
  },
  {
    meta: {
      title: 'Audit Log',
      icon: mdiClipboardTextSearchOutline,
    },
    path: '/admin/settings/audit-log',
    name: 'admin-audit-log',
    component: () => import('@/views/admin/AdminAuditLogView.vue'),
  },
  {
    meta: {
      title: 'Master Data',
      icon: mdiFolderOutline,
    },
    path: '/admin/master-data',
    alias: '/admin/master-data/',
    name: 'admin-master-data',
    component: () => import('@/views/admin/MasterDataIndexView.vue'),
  },
  {
    meta: {
      title: 'Guru',
      icon: mdiAccountTie,
    },
    path: '/admin/master-data/guru',
    name: 'admin-master-guru',
    component: () => import('@/views/admin/AdminTeachersView.vue'),
  },
  {
    meta: {
      title: 'Siswa',
      icon: mdiAccountSchool,
    },
    path: '/admin/master-data/siswa',
    name: 'admin-master-siswa',
    component: () => import('@/views/admin/AdminStudentsView.vue'),
  },
  {
    meta: {
      title: 'Verifikasi Pendaftaran',
      icon: mdiAccountCheckOutline,
    },
    path: '/admin/master-data/verifikasi-pendaftaran',
    name: 'admin-master-verifikasi-pendaftaran',
    component: () => import('@/views/admin/AdminRegistrationsView.vue'),
  },
  {
    meta: {
      title: 'Pengumuman',
      icon: mdiBullhornOutline,
    },
    path: '/admin/pengumuman',
    name: 'admin-announcements',
    component: () => import('@/views/admin/AdminAnnouncementsView.vue'),
  },
  {
    meta: {
      title: 'Program',
      icon: mdiSchoolOutline,
      resourceConfig: {
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
      title: 'Level',
      icon: mdiLayersTripleOutline,
      resourceConfig: {
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
      title: 'Group',
      icon: mdiAccountGroupOutline,
      resourceConfig: {
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
      title: 'Mata Pelajaran',
      icon: mdiBookEducationOutline,
      resourceConfig: {
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
      title: 'Sesi',
      icon: mdiCalendarClockOutline,
      resourceConfig: {
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
      icon: mdiBookOpenVariant,
    },
    path: '/admin/bank-soal',
    name: 'admin-bank-soal',
    component: () => import('@/views/QuestionBankView.vue'),
  },
  {
    meta: {
      title: 'Buat Bank Soal',
    },
    path: '/admin/bank-soal/new',
    name: 'admin-bank-soal-new',
    component: () => import('@/views/QuestionBankNewView.vue'),
  },
  {
    meta: {
      title: 'Impor Soal',
      icon: mdiFileDocumentOutline,
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
      icon: mdiCalendarClockOutline,
    },
    path: '/admin/ujian/jadwal',
    name: 'admin-ujian-jadwal',
    component: () => import('@/views/ExamsView.vue'),
  },
  {
    meta: {
      title: 'Token',
      icon: mdiKeyVariant,
    },
    path: '/admin/ujian/token',
    name: 'admin-ujian-token',
    component: () => import('@/views/ExamTokensView.vue'),
  },
  {
    meta: {
      title: 'Monitor Ujian',
      icon: mdiMonitorEye,
    },
    path: '/admin/ujian/monitor-ujian',
    name: 'admin-ujian-monitor',
    component: () => import('@/views/monitor/ExamMonitorView.vue'),
  },
  {
    meta: {
      title: 'Monitor Peserta',
      icon: mdiAccountSearchOutline,
    },
    path: '/admin/ujian/monitor-peserta',
    name: 'admin-ujian-monitor-peserta',
    component: () => import('@/views/monitor/ParticipantMonitorView.vue'),
  },
  {
    meta: {
      title: 'Absensi Peserta',
      icon: mdiCalendarCheckOutline,
    },
    path: '/admin/ujian/absensi',
    name: 'admin-ujian-absensi',
    component: () => import('@/views/monitor/AttendanceMonitorView.vue'),
  },
  {
    meta: { title: 'Generate QR Absensi' },
    path: '/admin/ujian/absensi/qr/:id',
    name: 'admin-ujian-absensi-qr',
    component: () => import('@/views/monitor/AttendanceQRView.vue'),
  },
  {
    meta: {
      title: 'Reset Login',
      icon: mdiAccountSwitchOutline,
    },
    path: '/admin/ujian/reset-login',
    name: 'admin-ujian-reset-login',
    component: () => import('@/views/ResetLoginView.vue'),
  },
  {
    meta: {
      title: 'Evaluasi / Hasil Nilai',
      icon: mdiChartBoxOutline,
    },
    path: '/admin/evaluasi',
    name: 'admin-evaluasi',
    component: () => import('@/views/EvaluationView.vue'),
  },
  {
    meta: {
      title: 'Advanced Analytics',
      icon: mdiChartLine,
    },
    path: '/admin/analytics',
    name: 'admin-analytics',
    component: () => import('@/views/admin/AdminAnalyticsView.vue'),
  },
  {
    meta: {
      title: 'Cetak',
      icon: mdiPrinterOutline,
    },
    path: '/admin/cetak',
    name: 'admin-cetak',
    component: () => import('@/views/admin/AdminPrintView.vue'),
  },
  {
    meta: {
      title: 'Config / Settings',
      icon: mdiCogOutline,
    },
    path: '/admin/settings',
    name: 'admin-settings',
    component: () => import('@/views/admin/AdminSettingsView.vue'),
  },
  {
    meta: {
      title: 'Integrasi LMS',
      icon: mdiDatabaseExportOutline,
    },
    path: '/admin/lms',
    name: 'admin-lms',
    component: () => import('@/views/admin/AdminLMSView.vue'),
  },
  {
    meta: {
      title: 'LTI Platforms',
      icon: mdiLinkVariant,
    },
    path: '/admin/lti',
    name: 'admin-lti',
    component: () => import('@/views/admin/AdminLTIView.vue'),
  },

  // Teacher
  {
    meta: { title: 'Dashboard Guru', icon: mdiHomeOutline },
    path: '/teacher/dashboard',
    name: 'teacher-dashboard',
    component: () => import('@/views/dashboards/TeacherDashboardView.vue'),
  },
  {
    meta: { title: 'Semua Bank Soal', icon: mdiBookOpenVariant },
    path: '/teacher/bank-soal',
    name: 'teacher-bank-soal',
    component: () => import('@/views/QuestionBankView.vue'),
  },
  {
    meta: { title: 'Buat Bank Soal' },
    path: '/teacher/bank-soal/new',
    name: 'teacher-bank-soal-new',
    component: () => import('@/views/QuestionBankNewView.vue'),
  },
  {
    meta: { title: 'Impor Soal', icon: mdiFileDocumentOutline },
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
    meta: { title: 'Jadwal Ujian', icon: mdiCalendarClockOutline },
    path: '/teacher/ujian/jadwal',
    name: 'teacher-ujian-jadwal',
    component: () => import('@/views/ExamsView.vue'),
  },
  {
    meta: { title: 'Token', icon: mdiKeyVariant },
    path: '/teacher/ujian/token',
    name: 'teacher-ujian-token',
    component: () => import('@/views/ExamTokensView.vue'),
  },
  {
    meta: { title: 'Monitor Ujian', icon: mdiMonitorEye },
    path: '/teacher/ujian/monitor-ujian',
    name: 'teacher-ujian-monitor',
    component: () => import('@/views/monitor/ExamMonitorView.vue'),
  },
  {
    meta: { title: 'Monitor Peserta', icon: mdiAccountSearchOutline },
    path: '/teacher/ujian/monitor-peserta',
    name: 'teacher-ujian-monitor-peserta',
    component: () => import('@/views/monitor/ParticipantMonitorView.vue'),
  },
  {
    meta: { title: 'Absensi Peserta', icon: mdiCalendarCheckOutline },
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
    meta: { title: 'Reset Login', icon: mdiAccountSwitchOutline },
    path: '/teacher/ujian/reset-login',
    name: 'teacher-ujian-reset-login',
    component: () => import('@/views/ResetLoginView.vue'),
  },
  {
    meta: { title: 'Evaluasi', icon: mdiChartBoxOutline },
    path: '/teacher/evaluasi',
    name: 'teacher-evaluasi',
    component: () => import('@/views/EvaluationView.vue'),
  },
  {
    meta: { title: 'Profil', icon: mdiAccountTie },
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
    meta: { title: 'Dashboard Siswa', icon: mdiHomeOutline },
    path: '/student/dashboard',
    name: 'student-dashboard',
    component: () => import('@/views/dashboards/StudentDashboardView.vue'),
  },
  {
    meta: { title: 'Daftar Ujian', icon: mdiClipboardTextOutline },
    path: '/student/ujian',
    name: 'student-ujian',
    component: () => import('@/views/student/StudentExamsView.vue'),
  },
  {
    meta: { title: 'Kerjakan Ujian', icon: mdiClipboardTextOutline },
    path: '/student/kerjakan',
    name: 'student-kerjakan-home',
    redirect: '/student/ujian',
  },
  {
    meta: { title: 'Kerjakan Ujian', icon: mdiClipboardTextOutline },
    path: '/student/workspace/:sessionId',
    name: 'student-workspace',
    component: () => import('@/views/student/StudentExamWorkspaceView.vue'),
  },
  {
    meta: { title: 'Hasil Ujian', icon: mdiChartBoxOutline },
    path: '/student/hasil',
    name: 'student-hasil',
    component: () => import('@/views/ResultsView.vue'),
  },
  {
    meta: { title: 'Pengumuman', icon: mdiBullhornOutline },
    path: '/student/pengumuman',
    name: 'student-pengumuman',
    component: () => import('@/views/student/StudentAnnouncementsView.vue'),
  },
  {
    meta: { title: 'Profil', icon: mdiAccountSchool },
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

export default router
