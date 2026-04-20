<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  mdiCalendarCheckOutline,
  mdiMapMarker,
  mdiRefresh,
  mdiArrowLeft,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import QrcodeVue from 'qrcode.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()
const examId = route.params.id

const exam = ref(null)
const session = ref(null)
const isLoading = ref(false)
const errorMessage = ref('')

const lat = ref(null)
const lon = ref(null)
const radius = ref(50)
const duration = ref(15)

const isLocating = ref(false)

const getLocation = () => {
  if (!navigator.geolocation) {
    alert('Geolocation tidak didukung oleh browser Anda')
    return
  }
  isLocating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      lat.value = pos.coords.latitude
      lon.value = pos.coords.longitude
      isLocating.value = false
    },
    (err) => {
      console.error(err)
      alert('Gagal mendapatkan lokasi: ' + err.message)
      isLocating.value = false
    }
  )
}

const loadExam = async () => {
  try {
    const { data } = await api.get(`/api/v1/exams/${examId}`)
    exam.value = data?.data
  } catch (err) {
    errorMessage.value = 'Gagal memuat detail ujian'
  }
}

const createSession = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.post(`/api/v1/exams/${examId}/attendance-session`, {
      lat: lat.value,
      lon: lon.value,
      radius_meters: parseInt(radius.value),
      duration_minutes: parseInt(duration.value)
    })
    session.value = data.data
    startTimer()
  } catch (err) {
    errorMessage.value = err.response?.data?.error?.message || 'Gagal membuat sesi absensi'
  } finally {
    isLoading.value = false
  }
}

const timeLeft = ref(0)
let timer = null

const startTimer = () => {
  if (!session.value) return
  const expires = new Date(session.value.expires_at).getTime()
  
  if (timer) clearInterval(timer)
  
  timer = setInterval(() => {
    const now = new Date().getTime()
    const diff = expires - now
    if (diff <= 0) {
      timeLeft.value = 0
      clearInterval(timer)
      session.value = null
    } else {
      timeLeft.value = Math.floor(diff / 1000)
    }
  }, 1000)
}

const formatTimeLeft = (sec) => {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

onMounted(() => {
  loadExam()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const qrValue = ref('')
const updateQrValue = () => {
  if (session.value) {
    // Generate QR payload
    const payload = {
      t: session.value.token,
      e: examId,
      v: 'abs1'
    }
    qrValue.value = JSON.stringify(payload)
  }
}

// Watch session to update QR
import { watch } from 'vue'
watch(session, updateQrValue)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiCalendarCheckOutline" title="Generate QR Absensi" main>
        <BaseButton 
          :icon="mdiArrowLeft" 
          label="Kembali" 
          color="whiteDark" 
          outline 
          @click="router.back()" 
        />
      </SectionTitleLineWithButton>

      <div class="grid gap-6 md:grid-cols-2">
        <CardBox title="Pengaturan Geofence & Sesi">
          <div v-if="exam" class="mb-4 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
            <div class="text-xs text-slate-500 uppercase font-bold">Ujian</div>
            <div class="font-bold">{{ exam.title }}</div>
          </div>

          <FormField label="Lokasi Presensi (Opsional)">
            <div class="flex gap-2">
              <FormControl 
                v-model="lat" 
                placeholder="Latitude" 
                type="number" 
                step="any"
                readonly
              />
              <FormControl 
                v-model="lon" 
                placeholder="Longitude" 
                type="number" 
                step="any"
                readonly
              />
              <BaseButton 
                :icon="mdiMapMarker" 
                color="info" 
                @click="getLocation" 
                :disabled="isLocating"
              />
            </div>
            <template #help>
              Klik ikon map untuk mengambil lokasi Anda saat ini. Kosongkan jika tidak ingin menggunakan geofence.
            </template>
          </FormField>

          <div class="grid gap-4 md:grid-cols-2">
            <FormField label="Radius (Meter)">
              <FormControl v-model="radius" type="number" min="5" max="1000" />
            </FormField>
            <FormField label="Durasi QR (Menit)">
              <FormControl v-model="duration" type="number" min="1" max="1440" />
            </FormField>
          </div>

          <div class="mt-6">
            <BaseButton 
              color="info" 
              label="Generate QR Code" 
              class="w-full" 
              :disabled="isLoading"
              @click="createSession"
            />
          </div>

          <div v-if="errorMessage" class="mt-4 text-sm text-red-500">
            {{ errorMessage }}
          </div>
        </CardBox>

        <CardBox class="flex flex-col items-center justify-center min-h-[400px]">
          <div v-if="session" class="text-center">
            <div class="mb-6 p-6 bg-white rounded-2xl shadow-xl inline-block border-8 border-slate-50">
              <QrcodeVue :value="qrValue" :size="280" level="H" />
            </div>
            
            <div class="mb-2 text-2xl font-mono font-bold text-info">
              {{ formatTimeLeft(timeLeft) }}
            </div>
            <div class="text-sm text-slate-500 uppercase tracking-widest font-bold">
              QR Berlaku Hingga
            </div>
            
            <div v-if="session.lat" class="mt-4 text-[10px] text-slate-400 font-mono">
              GEOFENCE ACTIVE: {{ session.radius_meters }}m @ {{ session.lat.toFixed(6) }}, {{ session.lon.toFixed(6) }}
            </div>
          </div>
          
          <div v-else class="text-center text-slate-400">
            <div class="mb-4">
              <mdi-calendar-check-outline class="w-16 h-16 mx-auto opacity-20" />
            </div>
            <p>Klik tombol Generate untuk membuat QR Code presensi</p>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
