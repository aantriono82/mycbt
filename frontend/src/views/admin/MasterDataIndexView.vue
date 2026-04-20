<script setup>
import {
  mdiAccountCheckOutline,
  mdiBullhornOutline,
  mdiAccountSchool,
  mdiAccountTie,
  mdiBookEducationOutline,
  mdiFolderOutline,
  mdiLayersTripleOutline,
  mdiSchoolOutline,
  mdiAccountGroupOutline,
  mdiCalendarClockOutline,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiFolderOutline" title="Master Data" main />

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Guru</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola dan import data guru (template Excel).
              </div>
            </div>
            <BaseButton :icon="mdiAccountTie" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/guru" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Siswa</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola dan import data peserta didik (template Excel).
              </div>
            </div>
            <BaseButton :icon="mdiAccountSchool" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/siswa" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Verifikasi Pendaftaran</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Verifikasi akun/pendaftaran yang masuk sebelum aktif.
              </div>
            </div>
            <BaseButton :icon="mdiAccountCheckOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton
              to="/admin/master-data/verifikasi-pendaftaran"
              label="Buka"
              color="contrast"
            />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Pengumuman</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola pengumuman siswa, termasuk target level, group, atau siswa spesifik.
              </div>
            </div>
            <BaseButton :icon="mdiBullhornOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/pengumuman" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Program</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola jurusan atau program belajar seperti IPA, IPS, Bahasa.
              </div>
            </div>
            <BaseButton :icon="mdiSchoolOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/program" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Level</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola tingkatan kelas yang dipakai pada penjadwalan ujian.
              </div>
            </div>
            <BaseButton :icon="mdiLayersTripleOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/level" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Group</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola pembagian rombel atau kelompok siswa.
              </div>
            </div>
            <BaseButton :icon="mdiAccountGroupOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/group" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Mata Pelajaran</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola daftar mapel yang dipakai di bank soal dan ujian.
              </div>
            </div>
            <BaseButton :icon="mdiBookEducationOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/mata-pelajaran" label="Buka" color="contrast" />
          </div>
        </CardBox>

        <CardBox is-hoverable>
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="text-lg font-semibold dark:text-slate-100">Sesi</div>
              <div class="mt-1 text-sm text-slate-600 dark:text-slate-400">
                Kelola pembagian sesi waktu ujian (pagi/siang/sore).
              </div>
            </div>
            <BaseButton :icon="mdiCalendarClockOutline" color="info" small rounded-full />
          </div>
          <div class="mt-4">
            <BaseButton to="/admin/master-data/sesi" label="Buka" color="contrast" />
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
