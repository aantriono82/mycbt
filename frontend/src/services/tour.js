import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'
import { useAuthStore } from '@/stores/auth'

export const startAdminTour = () => {
  const driverObj = driver({
    showProgress: true,
    steps: [
      {
        popover: {
          title: 'Selamat Datang di AtigaCBT',
          description: 'Ini adalah dashboard utama Anda. Mari kita lihat fitur-fitur penting yang tersedia.',
          position: 'center'
        }
      },
      {
        element: '#app > div > div > main > div:nth-child(3)', // Stats row
        popover: {
          title: 'Statistik Sistem',
          description: 'Pantau jumlah siswa aktif, bank soal, dan rata-rata nilai secara real-time di sini.',
          position: 'bottom'
        }
      },
      {
        element: '#app > div > div > main > div.grid > div.lg\\:col-span-2 > div.grid', // Quick menu
        popover: {
          title: 'Akses Cepat',
          description: 'Gunakan menu ini untuk melompat langsung ke pengaturan penting seperti Data Siswa atau Bank Soal.',
          position: 'top'
        }
      },
      {
        element: '#app > div > div > main > div.grid > div:nth-child(2) > div.flex.flex-col > div.mt-6', // Activity feed
        popover: {
          title: 'Aktivitas Live',
          description: 'Pantau semua perubahan dan login terbaru dari Admin dan Guru secara real-time.',
          position: 'left'
        }
      },
      {
        popover: {
          title: 'Selesai!',
          description: 'Anda siap menggunakan AtigaCBT dengan maksimal.',
          position: 'center'
        }
      }
    ],
    nextBtnText: 'Lanjut ➔',
    prevBtnText: '🠔 Kembali',
    doneBtnText: 'Selesai ✓',
  })

  driverObj.drive()
}

export const checkAndStartTour = () => {
  const authStore = useAuthStore()
  if (!authStore.isAuthenticated) return

  const hasSeenTour = localStorage.getItem(`tour_seen_${authStore.user?.id}`)
  if (!hasSeenTour) {
    setTimeout(() => {
      if (authStore.user?.role === 'admin') {
        startAdminTour()
      }
      // Can add student/teacher tours here later
      localStorage.setItem(`tour_seen_${authStore.user?.id}`, 'true')
    }, 1500)
  }
}
