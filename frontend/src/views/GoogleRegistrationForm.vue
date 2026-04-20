<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import SectionFullScreen from '@/components/SectionFullScreen.vue'
import LayoutGuest from '@/layouts/LayoutGuest.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { 
  mdiAccount, mdiEmail, mdiBadgeAccount, mdiPhone, mdiGenderTransgender, 
  mdiCalendar, mdiSchool, mdiOfficeBuilding, mdiChevronRight, mdiChevronLeft,
  mdiCheckCircle, mdiPencil, mdiInformationOutline, mdiArrowRight
} from '@mdi/js'

const route = useRoute()
const router = useRouter()

const currentStep = ref(1)
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

// Lookups
const programs = ref([])
const levels = ref([])
const groups = ref([])

const form = reactive({
  role: route.query.role || 'student',
  name: '',
  email: '',
  
  // Step 1: Identitas
  nisn: '',
  nis: '',
  nip: '',
  birth_date: '',
  gender: '',
  phone: '',

  // Step 2: Info Sekolah
  jenjang: 'SMA',
  school_name: '',
  program_code: '',
  level_name: '',
  group_name: '',
  academic_year: new Date().getFullYear() + '/' + (new Date().getFullYear() + 1),
  nis_sekolah: '',
  note: ''
})

const nextStep = () => {
  // Simple validation for step 1
  if (currentStep.value === 1) {
    if (!form.name || !form.email || !form.phone) {
      errorMessage.value = 'Mohon lengkapi Nama, Email, dan No HP.'
      return
    }
    errorMessage.value = ''
  }
  if (currentStep.value < 3) currentStep.value++
}

const prevStep = () => {
  if (currentStep.value > 1) currentStep.value--
}

const goToStep = (step) => {
  currentStep.value = step
}

const submit = async () => {
  isLoading.value = true
  errorMessage.value = ''
  
  try {
    const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const response = await axios.post(`${apiUrl}/api/v1/auth/google/register`, {
      ...form,
      birth_date: form.birth_date ? new Date(form.birth_date).toISOString() : null
    })
    successMessage.value = response.data.message
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal mengirim pendaftaran.'
  } finally {
    isLoading.value = false
  }
}

onMounted(async () => {
  try {
    const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const [p, l, g] = await Promise.all([
      axios.get(`${apiUrl}/api/v1/lookups/programs`),
      axios.get(`${apiUrl}/api/v1/lookups/levels`),
      axios.get(`${apiUrl}/api/v1/lookups/groups`)
    ])
    programs.value = p.data?.data || []
    levels.value = l.data?.data || []
    groups.value = g.data?.data || []
  } catch (err) {
    console.error('Failed to load lookups', err)
  }
})

const genderOptions = [
  { id: 'L', label: 'Laki-laki' },
  { id: 'P', label: 'Perempuan' }
]

const jenjangOptions = [
  { id: 'SD', label: 'SD' },
  { id: 'SMP', label: 'SMP' },
  { id: 'SMA', label: 'SMA/SMK' }
]

const levelOptions = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12']

const groupOptions = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J']

</script>

<template>
  <LayoutGuest>
    <SectionFullScreen v-slot="{ cardClass }" bg="cream">
      <div class="w-full max-w-lg mx-auto py-10 px-4 flex flex-col items-center">
        <!-- Logo Header -->
        <div class="flex items-center self-start mb-8 ml-2 cursor-pointer" @click="router.push('/')">
          <div class="bg-white w-14 h-14 rounded-xl flex items-center justify-center p-2 shadow-lg shadow-blue-500/20">
             <img src="/logo_a3_blue.png" alt="A3" class="w-full h-full" />
          </div>
          <div class="ml-4">
            <h1 class="text-2xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-blue-600 to-indigo-600">Atiga CBT</h1>
            <p class="text-[10px] text-slate-400 font-bold tracking-widest uppercase">Professional LMS</p>
          </div>
        </div>

        <!-- Success View -->
        <div v-if="successMessage" class="w-full bg-white rounded-2xl shadow-sm border border-slate-100 p-10 text-center">
             <div class="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-blue-100 mb-6 font-bold text-blue-600">
              <svg class="h-10 w-10 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h2 class="text-2xl font-bold text-slate-800 mb-4">Pendaftaran Berhasil!</h2>
            <p class="text-slate-500 mb-8">{{ successMessage }}</p>
            <BaseButton color="info" label="Kembali ke Beranda" to="/login" rounded-full class="px-8 w-full" />
        </div>

        <!-- Wizard Content -->
        <template v-else>
          <!-- Custom Stepper -->
          <div class="w-full max-w-sm mb-10">
            <div class="relative flex items-center justify-between">
              <div v-for="s in 3" :key="s" class="flex flex-col items-center z-10">
                <div 
                  class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300 border-2"
                  :class="[
                  currentStep >= s ? 'bg-blue-600 border-blue-600 text-white shadow-md shadow-blue-500/20' : 'bg-white border-slate-200 text-slate-400'
                  ]"
                >
                  {{ s }}
                </div>
                <span class="text-[10px] font-medium mt-2" :class="currentStep >= s ? 'text-blue-600 font-bold' : 'text-slate-400'">
                  {{ s === 1 ? 'Identitas diri' : s === 2 ? 'Info sekolah' : 'Konfirmasi' }}
                </span>
              </div>
              <!-- Connector Line -->
              <div class="absolute top-4 left-0 w-full h-[1px] bg-slate-200 -z-0"></div>
              <div 
                class="absolute top-4 left-0 h-[1px] bg-blue-600 transition-all duration-500 -z-0"
                :style="{ width: ((currentStep - 1) * 50) + '%' }"
              ></div>
            </div>
          </div>

          <!-- Form Card -->
          <div class="w-full bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
            <div class="p-8">
              <div v-if="errorMessage" class="mb-6 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-700 border border-red-100">
                {{ errorMessage }}
              </div>

              <!-- Step 1: Identitas Diri -->
              <div v-if="currentStep === 1" class="space-y-6">
                <div>
                  <h2 class="text-xl font-bold text-slate-800 mb-1">Identitas diri</h2>
                  <p class="text-sm text-slate-500">Silakan lengkapi data profil pribadi Anda.</p>
                </div>

                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-slate-700 mb-1.5">Nama lengkap <span class="text-rose-500">*</span></label>
                    <input v-model="form.name" type="text" placeholder="Masukkan Nama Lengkap" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-blue-500 transition-all text-slate-800 outline-none" />
                  </div>
 
                   <div>
                     <label class="block text-sm font-medium text-slate-700 mb-1.5">Email <span class="text-rose-500">*</span></label>
                    <input v-model="form.email" type="email" placeholder="contoh@gmail.com" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-blue-500 transition-all text-slate-800 outline-none" />
                  </div>
 
                   <div v-if="form.role === 'student'">
                     <label class="block text-sm font-medium text-slate-700 mb-1.5">NISN <span class="text-rose-500">*</span></label>
                    <input v-model="form.nisn" type="text" placeholder="Masukkan 10 digit NISN" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-slate-800 outline-none" />
                  </div>
                   <div v-if="form.role === 'teacher'">
                     <label class="block text-sm font-medium text-slate-700 mb-1.5">NIP <span class="text-rose-500">*</span></label>
                    <input v-model="form.nip" type="text" placeholder="Masukkan NIP (atau - jika belum ada)" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-slate-800 outline-none" />
                  </div>
 
                   <div class="grid grid-cols-2 gap-4">
                     <div>
                       <label class="block text-sm font-medium text-slate-700 mb-1.5">Tanggal lahir <span class="text-rose-500">*</span></label>
                      <input v-model="form.birth_date" type="date" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-slate-800 outline-none" />
                    </div>
                     <div>
                       <label class="block text-sm font-medium text-slate-700 mb-1.5">Jenis kelamin <span class="text-rose-500">*</span></label>
                      <select v-model="form.gender" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-slate-800 outline-none appearance-none bg-no-repeat bg-[right_1rem_center] bg-[length:1em_1em]" style="background-image: url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A//www.w3.org/2000/svg%22%20width%3D%22292.4%22%20height%3D%22292.4%22%3E%3Cpath%20fill%3D%22%2364748b%22%20d%3D%22M287%2069.4a17.6%2017.6%200%200%200-13-5.4H18.4c-5%200-9.3%201.8-12.9%205.4A17.6%2017.6%200%200%200%200%2082.2c0%205%201.8%209.3%205.4%2012.9l128%20127.9c3.6%203.6%207.8%205.4%2012.8%205.4s9.2-1.8%2012.8-5.4L287%2095c3.5-3.5%205.4-7.8%205.4-12.8%200-5-1.9-9.2-5.5-12.8z%22/%3E%3C/svg%3E');">
                         <option value="" disabled>Pilih...</option>
                         <option value="L">Laki-laki</option>
                         <option value="P">Perempuan</option>
                       </select>
                     </div>
                   </div>
 
                   <div>
                     <label class="block text-sm font-medium text-slate-700 mb-1.5">Nomor HP aktif <span class="text-rose-500">*</span></label>
                    <input v-model="form.phone" type="text" placeholder="08xxxxxxxxxx" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-slate-800 outline-none" />
                    <p class="text-[11px] text-slate-400 mt-1.5 ml-1">Untuk keperluan verifikasi dan notifikasi ujian</p>
                  </div>
                </div>

                <div class="pt-6">
                  <button 
                    @click="nextStep"
                    class="w-full flex items-center justify-center py-3.5 px-6 rounded-xl border-2 border-slate-100 text-slate-700 font-bold hover:bg-slate-50 transition-all gap-2"
                  >
                    Lanjut ke info sekolah <BaseIcon :path="mdiArrowRight" size="20" />
                  </button>
                </div>
              </div>

              <!-- Step 2: Info Sekolah -->
              <div v-if="currentStep === 2" class="space-y-6">
                <div>
                  <h2 class="text-xl font-bold text-slate-800 mb-1">Info sekolah</h2>
                  <p class="text-sm text-slate-500">Pilih identitas sekolah Anda saat ini.</p>
                </div>

                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-slate-700 mb-1.5">Nama Sekolah <span class="text-rose-500">*</span></label>
                    <input v-model="form.school_name" type="text" placeholder="Masukkan Nama Sekolah" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-blue-500 outline-none" />
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-slate-700 mb-1.5">Jenjang <span class="text-rose-500">*</span></label>
                      <select v-model="form.jenjang" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 outline-none">
                        <option v-for="opt in jenjangOptions" :key="opt.id" :value="opt.id">{{ opt.label }}</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-slate-700 mb-1.5">Kelas <span class="text-rose-500">*</span></label>
                      <select v-model="form.level_name" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 outline-none">
                        <option value="">Pilih...</option>
                        <option v-for="l in levels" :key="l.id" :value="l.name">{{ l.name }}</option>
                      </select>
                    </div>
                  </div>

                  <div class="grid grid-cols-2 gap-4">
                     <div>
                      <label class="block text-sm font-medium text-slate-700 mb-1.5">Rombel/Group <span class="text-rose-500">*</span></label>
                      <select v-model="form.group_name" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 outline-none">
                        <option value="">Pilih...</option>
                        <option v-for="g in groups" :key="g.id" :value="g.name">{{ g.name }}</option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-slate-700 mb-1.5">Tahun Ajaran <span class="text-rose-500">*</span></label>
                      <input v-model="form.academic_year" type="text" class="w-full px-4 py-2.5 rounded-xl border border-slate-200 outline-none" />
                    </div>
                  </div>
                </div>

                <div class="pt-6">
                  <button 
                    @click="nextStep"
                    class="w-full flex items-center justify-center py-3.5 px-6 rounded-xl border-2 border-slate-100 text-slate-700 font-bold hover:bg-slate-50 transition-all gap-2"
                  >
                    Lanjut ke konfirmasi <BaseIcon :path="mdiArrowRight" size="20" />
                  </button>
                  <button @click="prevStep" class="w-full text-sm text-slate-400 font-medium py-3 hover:text-slate-600 transition-colors mt-2">
                    Kembali
                  </button>
                </div>
              </div>

              <!-- Step 3: Konfirmasi -->
              <div v-if="currentStep === 3" class="space-y-6 text-center">
                 <div>
                  <h2 class="text-xl font-bold text-slate-800 mb-1">Konfirmasi data</h2>
                  <p class="text-sm text-slate-500">Pastikan seluruh data sudah benar.</p>
                </div>

                <div class="bg-slate-50 rounded-2xl p-6 text-left space-y-4 border border-slate-100">
                  <div class="flex items-center justify-between border-b border-slate-200 pb-3">
                    <span class="text-xs font-bold text-slate-400 uppercase tracking-widest">Identitas</span>
                    <button @click="goToStep(1)" class="text-blue-600 font-bold text-xs">Ubah</button>
                  </div>
                  <div class="space-y-2">
                    <div class="flex justify-between text-sm">
                      <span class="text-slate-500 text-xs">Nama Lengkap</span>
                      <span class="font-medium text-slate-800">{{ form.name }}</span>
                    </div>
                    <div class="flex justify-between text-sm">
                      <span class="text-slate-500 text-xs">Email</span>
                      <span class="font-medium text-slate-800 font-mono">{{ form.email }}</span>
                    </div>
                    <div v-if="form.role === 'student'" class="flex justify-between text-sm">
                      <span class="text-slate-500 text-xs">NISN</span>
                      <span class="font-medium text-slate-800">{{ form.nisn }}</span>
                    </div>
                    <div v-if="form.role === 'teacher'" class="flex justify-between text-sm">
                      <span class="text-slate-500 text-xs">NIP</span>
                      <span class="font-medium text-slate-800">{{ form.nip }}</span>
                    </div>
                  </div>

                  <div class="flex items-center justify-between border-b border-slate-200 pt-4 pb-3">
                    <span class="text-xs font-bold text-slate-400 uppercase tracking-widest">Sekolah</span>
                    <button @click="goToStep(2)" class="text-blue-600 font-bold text-xs">Ubah</button>
                  </div>
                   <div class="space-y-2">
                    <div class="flex justify-between text-sm">
                      <span class="text-slate-500 text-xs">Nama Sekolah</span>
                      <span class="font-medium text-slate-800">{{ form.school_name }}</span>
                    </div>
                    <div class="flex justify-between text-sm">
                       <span class="text-slate-500 text-xs">Kelas</span>
                       <span class="font-medium text-slate-800">{{ form.level_name }} - {{ form.group_name }}</span>
                    </div>
                  </div>
                </div>

                <div class="pt-6">
                  <button 
                    @click="submit"
                    :disabled="isLoading"
                    class="w-full flex items-center justify-center py-4 px-6 rounded-xl bg-blue-600 text-white font-bold hover:bg-blue-700 transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50"
                  >
                    {{ isLoading ? 'Memproses...' : 'Daftar Sekarang' }}
                  </button>
                   <button @click="prevStep" class="w-full text-sm text-slate-400 font-medium py-3 hover:text-slate-600 transition-colors mt-2">
                    Kembali
                  </button>
                </div>
              </div>

            </div>
          </div>
        </template>
      </div>
    </SectionFullScreen>
  </LayoutGuest>
</template>

<style scoped>
/* Cream Background Gradient */
:deep(.cream) {
  background-color: #F8F9F5 !important;
  background-image: 
    radial-gradient(at 0% 0%, hsla(161, 48%, 95%, 1) 0, transparent 50%), 
    radial-gradient(at 50% 0%, hsla(161, 48%, 95%, 1) 0, transparent 50%),
    radial-gradient(at 100% 0%, hsla(161, 48%, 95%, 1) 0, transparent 50%) !important;
}

input::placeholder {
  color: #94a3b8;
}

select {
  background-image: url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A//www.w3.org/2000/svg%22%20width%3D%22292.4%22%20height%3D%22292.4%22%3E%3Cpath%20fill%3D%22%2364748b%22%20d%3D%22M287%2069.4a17.6%2017.6%200%200%200-13-5.4H18.4c-5%200-9.3%201.8-12.9%205.4A17.6%2017.6%200%200%200%200%2082.2c0%205%201.8%209.3%205.4%2012.9l128%20127.9c3.6%203.6%207.8%205.4%2012.8%205.4s9.2-1.8%2012.8-5.4L287%2095c3.5-3.5%205.4-7.8%205.4-12.8%200-5-1.9-9.2-5.5-12.8z%22/%3E%3C/svg%3E');
  background-repeat: no-repeat;
  background-position: right 1rem center;
  background-size: 11px;
}
</style>
