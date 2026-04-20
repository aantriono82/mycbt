<script setup>
import { onMounted, ref, reactive } from 'vue'
import {
  mdiLinkVariant,
  mdiPlus,
  mdiRefresh,
  mdiKeyVariant,
  mdiInformationOutline,
  mdiContentCopy,
  mdiTrashCanOutline,
  mdiPencil
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import NotificationBar from '@/components/NotificationBar.vue'
import { api } from '@/services/api.js'

const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const platforms = ref([])
const showModal = ref(false)

const form = reactive({
  name: '',
  issuer: '',
  client_id: '',
  deployment_id: '',
  oidc_auth_url: '',
  oidc_token_url: '',
  jwks_url: '',
  tool_private_key: '',
  tool_public_key: ''
})

const loadPlatforms = async () => {
  isLoading.value = true
  try {
    const { data } = await api.get('/api/v1/lti/platforms')
    platforms.value = data?.data || []
  } catch (error) {
    errorMessage.value = 'Gagal memuat daftar platform LTI'
  } finally {
    isLoading.value = false
  }
}

const generateKeys = async () => {
  try {
    const { data } = await api.post('/api/v1/lti/keys/generate')
    form.tool_private_key = data.private_key
    form.tool_public_key = data.public_key
  } catch (error) {
    alert('Gagal generate keys')
  }
}

const savePlatform = async () => {
  isSaving.value = true
  try {
    await api.post('/api/v1/lti/platforms', form)
    showModal.value = false
    loadPlatforms()
    Object.assign(form, {
      name: '', issuer: '', client_id: '', deployment_id: '',
      oidc_auth_url: '', oidc_token_url: '', jwks_url: '',
      tool_private_key: '', tool_public_key: ''
    })
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Gagal menyimpan platform'
  } finally {
    isSaving.value = false
  }
}

const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text)
  alert('Disalin ke clipboard')
}

const getBaseURL = () => window.location.origin + '/api/v1'

onMounted(loadPlatforms)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiLinkVariant" title="LTI 1.3 Platforms" main>
        <BaseButton
          :icon="mdiPlus"
          label="Tambah Platform"
          color="info"
          @click="showModal = !showModal"
        />
      </SectionTitleLineWithButton>

      <NotificationBar v-if="errorMessage" color="danger" :icon="mdiInformationOutline">
        {{ errorMessage }}
      </NotificationBar>

      <!-- Connection Details Card (ReadOnly for LMS Config) -->
      <CardBox class="mb-6 bg-slate-50 dark:bg-slate-900/50 border-dashed border-2 border-slate-200 dark:border-slate-800">
        <h4 class="font-bold mb-4 flex items-center">
          <BaseIcon :path="mdiInformationOutline" class="mr-2" />
          Detail Konfigurasi untuk LMS (Moodle, Canvas, dll)
        </h4>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div class="space-y-2">
            <p class="font-semibold text-slate-500 uppercase text-xs">OIDC Login Initiation URL</p>
            <div class="flex items-center bg-white dark:bg-slate-800 p-2 rounded border border-slate-200 dark:border-slate-700">
              <code class="flex-1 truncate">{{ getBaseURL() }}/lti/login</code>
              <BaseButton :icon="mdiContentCopy" small color="white" @click="copyToClipboard(getBaseURL() + '/lti/login')" />
            </div>
          </div>
          <div class="space-y-2">
            <p class="font-semibold text-slate-500 uppercase text-xs">Target Link URI (Launch URL)</p>
            <div class="flex items-center bg-white dark:bg-slate-800 p-2 rounded border border-slate-200 dark:border-slate-700">
              <code class="flex-1 truncate">{{ getBaseURL() }}/lti/launch</code>
              <BaseButton :icon="mdiContentCopy" small color="white" @click="copyToClipboard(getBaseURL() + '/lti/launch')" />
            </div>
          </div>
        </div>
      </CardBox>

      <!-- Platform List -->
      <CardBox has-table>
        <div v-if="isLoading" class="p-6 text-center text-slate-500 italic">Memuat data...</div>
        <div v-else-if="platforms.length === 0" class="p-6 text-center text-slate-500 italic">Belum ada platform terdaftar.</div>
        <table v-else>
          <thead>
            <tr>
              <th>Platform Name</th>
              <th>Issuer (iss)</th>
              <th>Client ID</th>
              <th>Created</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in platforms" :key="p.id">
              <td data-label="Name" class="font-semibold">{{ p.name }}</td>
              <td data-label="Issuer"><code>{{ p.issuer }}</code></td>
              <td data-label="Client ID">{{ p.client_id }}</td>
              <td data-label="Created">{{ new Date(p.created_at).toLocaleDateString() }}</td>
              <td class="before:hidden lg:w-1 whitespace-nowrap">
                <BaseButtons type="justify-start lg:justify-end" no-wrap>
                  <BaseButton color="info" :icon="mdiPencil" small />
                  <BaseButton color="danger" :icon="mdiTrashCanOutline" small />
                </BaseButtons>
              </td>
            </tr>
          </tbody>
        </table>
      </CardBox>

      <!-- Add Platform Modal / Form -->
      <CardBox v-if="showModal" is-form class="mt-6" @submit.prevent="savePlatform">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-xl font-bold">Registrasi Platform Baru</h3>
          <BaseButton :icon="mdiRefresh" small label="Generate Tool Keys" color="warning" @click="generateKeys" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField label="Friendly Name" help="Contoh: Moodle SMK Atiga">
            <FormControl v-model="form.name" required />
          </FormField>
          <FormField label="Platform Issuer (iss)" help="URL issuer dari platform LMS">
            <FormControl v-model="form.issuer" placeholder="https://moodle.yourserver.com" required />
          </FormField>
          <FormField label="Client ID" help="Client ID yang diberikan oleh LMS">
            <FormControl v-model="form.client_id" required />
          </FormField>
          <FormField label="Deployment ID" help="Deployment ID dari platform">
            <FormControl v-model="form.deployment_id" required />
          </FormField>
          <FormField label="OIDC Auth URL" help="LMS OIDC auth endpoint">
            <FormControl v-model="form.oidc_auth_url" required />
          </FormField>
           <FormField label="OIDC Token URL" help="LMS token endpoint">
            <FormControl v-model="form.oidc_token_url" required />
          </FormField>
          <FormField label="Public Keyset URL (JWKS)" help="LMS JWKS endpoint" class="md:col-span-2">
            <FormControl v-model="form.jwks_url" required />
          </FormField>
          
          <FormField label="Tool Private Key" class="md:col-span-1">
            <FormControl v-model="form.tool_private_key" type="textarea" placeholder="Generate atau paste private key PEM..." required />
          </FormField>
          <FormField label="Tool Public Key" class="md:col-span-1">
            <FormControl v-model="form.tool_public_key" type="textarea" placeholder="Generate atau paste public key PEM..." required />
          </FormField>
        </div>

        <template #footer>
          <BaseButtons>
            <BaseButton type="submit" color="info" label="Simpan Platform" :disabled="isSaving" />
            <BaseButton color="info" outline label="Batal" @click="showModal = false" />
          </BaseButtons>
        </template>
      </CardBox>

    </SectionMain>
  </LayoutAuthenticated>
</template>
