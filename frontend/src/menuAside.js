import {
  mdiHomeOutline,
  mdiFolderOutline,
  mdiAccountTie,
  mdiAccountSchool,
  mdiAccountCheckOutline,
  mdiBullhornOutline,
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
  mdiAccountCircleOutline,
  mdiLogout,
  mdiPlus,
  mdiDatabaseExportOutline,
  mdiLinkVariant,
  mdiChartLine,
  mdiHistory,
  mdiClipboardTextSearchOutline,
} from '@mdi/js'

import { ROLES } from '@/stores/auth.js'

export const menuAsideBottom = [
  {
    label: 'Logout',
    icon: mdiLogout,
    color: 'info',
    isLogout: true,
  },
]

export const getMenuAsideMain = (role) => {
  if (role === ROLES.TEACHER) {
    return [
      { to: '/teacher/dashboard', icon: mdiHomeOutline, label: 'Dashboard' },
      {
        label: 'Bank Soal',
        icon: mdiBookOpenVariant,
        menu: [
          { to: '/teacher/bank-soal', label: 'Bank Soal', icon: mdiBookOpenVariant },
          { to: '/teacher/bank-soal/new', label: 'Buat Soal', icon: mdiPlus },
          { to: '/teacher/bank-soal/import', label: 'Impor Soal', icon: mdiFileDocumentOutline },
        ],
      },
      {
        label: 'Ujian',
        icon: mdiClipboardTextOutline,
        menu: [
          { to: '/teacher/ujian/jadwal', label: 'Jadwal Ujian', icon: mdiCalendarClockOutline },
          { to: '/teacher/ujian/token', label: 'Token', icon: mdiKeyVariant },
          { to: '/teacher/ujian/monitor-ujian', label: 'Monitor Ujian', icon: mdiMonitorEye },
          {
            to: '/teacher/ujian/monitor-peserta',
            label: 'Monitor Peserta',
            icon: mdiAccountSearchOutline,
          },
          { to: '/teacher/ujian/reset-login', label: 'Reset Login', icon: mdiAccountSwitchOutline },
        ],
      },
      {
        label: 'Evaluasi',
        icon: mdiChartBoxOutline,
        menu: [
          { to: '/teacher/evaluasi', label: 'Hasil Ujian', icon: mdiChartBoxOutline },
          { to: '/teacher/evaluasi', label: 'Analitik & Tren', icon: mdiChartLine },
        ],
      },
    ]
  }

  if (role === ROLES.STUDENT) {
    return [
      { to: '/student/dashboard', icon: mdiHomeOutline, label: 'Dashboard' },
      { to: '/student/ujian', icon: mdiClipboardTextOutline, label: 'Daftar Ujian' },
      { to: '/student/hasil', icon: mdiChartBoxOutline, label: 'Hasil Ujian' },
      { to: '/student/pengumuman', icon: mdiBullhornOutline, label: 'Pengumuman' },
    ]
  }

  // Admin (default)
  return [
    { to: '/admin/dashboard', icon: mdiHomeOutline, label: 'Dashboard' },
    { to: '/admin/pengumuman', label: 'Pengumuman', icon: mdiBullhornOutline },
    {
      label: 'Master Data',
      icon: mdiFolderOutline,
      menu: [
        { to: '/admin/master-data/guru', label: 'Guru', icon: mdiAccountTie },
        { to: '/admin/master-data/siswa', label: 'Siswa', icon: mdiAccountSchool },
        {
          to: '/admin/master-data/verifikasi-pendaftaran',
          label: 'Registrasi',
          icon: mdiAccountCheckOutline,
        },
        {
          to: '/admin/master-data/sesi',
          label: 'Sesi',
          icon: mdiCalendarClockOutline,
        },
      ],
    },
    {
      label: 'Bank Soal',
      icon: mdiBookOpenVariant,
      menu: [
        { to: '/admin/bank-soal', label: 'Bank Soal', icon: mdiBookOpenVariant },
        { to: '/admin/bank-soal/new', label: 'Buat Soal', icon: mdiPlus },
        { to: '/admin/bank-soal/import', label: 'Impor Soal', icon: mdiFileDocumentOutline },
      ]
    },
    {
      label: 'Ujian',
      icon: mdiClipboardTextOutline,
      menu: [
        { to: '/admin/ujian/jadwal', label: 'Jadwal Ujian', icon: mdiCalendarClockOutline },
        { to: '/admin/ujian/token', label: 'Token', icon: mdiKeyVariant },
        { to: '/admin/ujian/monitor-ujian', label: 'Monitor Ujian', icon: mdiMonitorEye },
        { to: '/admin/ujian/monitor-peserta', label: 'Mon. Peserta', icon: mdiAccountSearchOutline },
        { to: '/admin/ujian/reset-login', label: 'Reset Login', icon: mdiAccountSwitchOutline },
      ],
    },
    {
      label: 'Evaluasi',
      icon: mdiChartBoxOutline,
      menu: [
        { to: '/admin/evaluasi', label: 'Hasil Ujian', icon: mdiChartBoxOutline },
        { to: '/admin/analytics', label: 'Analitik & Tren', icon: mdiChartLine }
      ]
    },
    { to: '/admin/cetak', label: 'Cetak', icon: mdiPrinterOutline },
    {
      label: 'Config',
      icon: mdiCogOutline,
      menu: [
        { to: '/admin/settings', label: 'General Settings', icon: mdiCogOutline },
        { to: '/admin/settings/activity-log', label: 'Log Aktivitas', icon: mdiHistory },
        { to: '/admin/settings/audit-log', label: 'Audit Log', icon: mdiClipboardTextSearchOutline },
        { to: '/admin/lms', label: 'Integrasi LMS', icon: mdiDatabaseExportOutline },
      ]
    },
  ]
}
