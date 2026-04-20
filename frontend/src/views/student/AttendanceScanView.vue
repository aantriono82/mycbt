<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  mdiQrcodeScan,
  mdiMapMarker,
  mdiCheckCircle,
  mdiAlertCircle,
  mdiClose,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import { Html5QrcodeScanner } from 'html5-qrcode'
import { api } from '@/services/api.js'

const router = useRouter()
const scanner = ref(null)
const scanResult = ref(null)
const isScanning = ref(true)
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const lat = ref(null)
const lon = ref(null)
const accuracy = ref(null)

const getLocation = () => {
  return new Promise((resolve) => {
    if (!navigator.geolocation) {
      resolve()
      return
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        lat.value = pos.coords.latitude
        lon.value = pos.coords.longitude
        accuracy.value = pos.coords.accuracy
        resolve()
      },
      (err) => {
        console.error(err)
        resolve()
      },
      { enableHighAccuracy: true, timeout: 5000 }
    )
  })
}

const onScanSuccess = async (decodedText) => {
  if (!isScanning.value) return
  isScanning.value = false
  if (scanner.value) {
    scanner.value.clear()
  }

  isLoading.value = true
  errorMessage.value = ''
  
  try {
    let payload = null
    try {
      payload = JSON.parse(decodedText)
    } catch (e) {
      throw new Error('QR Code tidak valid')
    }

    if (!payload.t || payload.v !== 'abs1') {
      throw new Error('Format QR Code tidak dikenali')
    }

    await getLocation()

    const { data } = await api.post('/api/v1/student/attendance/scan', {
      token: payload.t,
      lat: lat.value,
      lon: lon.value,
      accuracy: accuracy.value
    })

    successMessage.value = 'Presensi berhasil dicatat!'
    scanResult.value = data.data
  } catch (err) {
    errorMessage.value = err.response?.data?.error?.message || err.message || 'Gagal melakukan presensi'
    isScanning.value = true
    setTimeout(startScanner, 2000)
  } finally {
    isLoading.value = false
  }
}

const startScanner = () => {
  successMessage.value = ''
  errorMessage.value = ''
  isScanning.value = true
  
  const config = { 
    fps: 10, 
    qrbox: { width: 250, height: 250 },
    aspectRatio: 1.0
  }
  
  scanner.value = new Html5QrcodeScanner('qr-reader', config, false)
  scanner.value.render(onScanSuccess, (err) => {
    // ignore scan errors (they happen every frame if no QR)
  })
}

onMounted(() => {
  startScanner()
})

onUnmounted(() => {
  if (scanner.value) {
    scanner.value.clear()
  }
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiQrcodeScan" title="Scan QR Absensi" main />

      <div class="max-w-xl mx-auto">
        <CardBox v-if="!successMessage">
          <div id="qr-reader" class="overflow-hidden rounded-xl border-0"></div>
          
          <div class="mt-6 text-center">
            <p class="text-slate-500 dark:text-slate-400">
              Arahkan kamera ke QR Code yang ditampilkan oleh Guru/Admin.
            </p>
          </div>

          <div v-if="isLoading" class="mt-6 text-center text-info animate-pulse font-bold">
            Memproses presensi...
          </div>

          <div v-if="errorMessage" class="mt-6 p-4 rounded-xl bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 border border-red-100 dark:border-red-900/40 flex items-start gap-3">
            <mdi-alert-circle class="w-5 h-5 flex-shrink-0" />
            <div>{{ errorMessage }}</div>
          </div>
        </CardBox>

        <CardBox v-else class="text-center py-10">
          <div class="mb-6">
            <mdi-check-circle class="w-20 h-20 mx-auto text-emerald-500" />
          </div>
          <h2 class="text-2xl font-bold mb-2">Berhasil!</h2>
          <p class="text-slate-500 dark:text-slate-400 mb-8">
            {{ successMessage }}
          </p>
          
          <div v-if="scanResult" class="mb-8 p-4 bg-slate-50 dark:bg-slate-800 rounded-xl text-left inline-block w-full">
            <div class="text-xs text-slate-500 uppercase font-bold mb-1">Waktu</div>
            <div class="font-mono mb-3">{{ new Date(scanResult.attended_at).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short', hour12: false }).replace(/\./g, ':') + ' WIB' }}</div>
            
            <div v-if="scanResult.lat" class="text-xs text-slate-500 uppercase font-bold mb-1">Lokasi</div>
            <div v-if="scanResult.lat" class="text-xs font-mono text-slate-400 group">
              {{ scanResult.lat.toFixed(6) }}, {{ scanResult.lon.toFixed(6) }} 
              <span class="ml-1 text-[10px]">(akurasi: {{ scanResult.accuracy.toFixed(1) }}m)</span>
            </div>
          </div>

          <BaseButton 
            color="info" 
            label="Kembali ke Dashboard" 
            @click="router.push('/student/dashboard')" 
          />
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>

<style>
#qr-reader {
  border: none !important;
}
#qr-reader__scan_region {
  background: #f8fafc;
  border-radius: 12px;
}
.dark #qr-reader__scan_region {
  background: #0f172a;
}
#qr-reader__dashboard_section_csr button {
  padding: 8px 16px;
  background: #3b82f6;
  color: white;
  border-radius: 8px;
  font-weight: 600;
  margin-top: 10px;
}
</style>
