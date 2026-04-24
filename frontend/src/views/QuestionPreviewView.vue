<template>
  <div class="tka-theme min-h-screen bg-slate-100">
    <!-- ANBK TOP HEADER -->
    <header class="tka-topbar text-white px-6 py-3 flex items-center justify-between shadow sticky top-0 z-50">
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="bg-white/10 px-3 py-1.5 rounded hover:bg-white/15 active:scale-[0.99] transition"
          @click="backToList"
        >
          <BaseIcon :path="mdiArrowLeft" class="text-white" />
        </button>
        <div class="font-bold tracking-tight text-sm md:text-base uppercase select-none">
          {{ questionSet?.name || 'TO TKA MTK SMP LAMPUNG TENGAH' }}
        </div>
      </div>
      <div class="flex items-center gap-4">
        <div class="hidden sm:block text-right leading-tight">
          <div class="text-xs font-semibold">PRATINJAU</div>
        </div>
        <div class="tka-timer bg-white text-slate-900 px-4 py-1.5 rounded-full font-semibold flex items-center gap-2">
          <span class="text-xs">SISA WAKTU:</span>
          <span class="font-mono">{{ formattedTime }}</span>
        </div>
      </div>
    </header>

    <!-- LOADING / ERROR -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center min-h-[60vh]">
       <div class="h-16 w-16 border-4 border-[#0D47A1] border-t-transparent rounded-full animate-spin mb-6"></div>
       <p class="text-[#0D47A1] font-black uppercase tracking-[0.3em] animate-pulse">Memuat Pratinjau...</p>
    </div>

    <div v-else-if="errorMessage" class="max-w-xl mx-auto mt-20 p-12 bg-white rounded-3xl border-2 border-red-50 shadow-2xl text-center">
        <BaseIcon :path="mdiAlert" size="64" class="text-red-500 mb-6 mx-auto" />
        <h2 class="text-2xl font-black text-slate-800 mb-4 uppercase">Terjadi Kesalahan</h2>
        <p class="text-slate-600 mb-8 leading-relaxed">{{ errorMessage }}</p>
        <button @click="backToList" class="bg-[#0D47A1] text-white px-8 py-3 rounded-xl font-bold uppercase transition-all hover:scale-105 active:scale-95">Kembali</button>
    </div>

    <!-- MAIN CONTENT AREA -->
    <main v-else-if="currentQuestion" :key="currentQuestion.id" class="max-w-[1500px] mx-auto px-6 py-6 pb-24 animate-fade-in">
       <div class="grid lg:grid-cols-[1fr_340px] gap-6 items-start">
          
          <!-- LEFT COLUMN: The Question Card -->
          <div class="bg-white rounded-lg border border-slate-200 shadow-sm min-h-[75vh] max-h-[calc(100vh-180px)] flex flex-col">
             <!-- Card Header -->
             <div class="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-white">
                <span class="text-[#0B7EA1] font-bold uppercase text-sm tracking-wide">SOAL NOMOR: {{ currentIndex + 1 }}</span>
                <div class="flex items-center gap-4 text-[11px] font-bold text-slate-500">
                   <span class="uppercase opacity-60">Ukuran Font:</span>
                   <button class="h-6 w-8 flex items-center justify-center border border-slate-300 rounded bg-slate-50 text-slate-700 font-bold leading-none">A</button>
                   <button class="h-7 w-9 flex items-center justify-center border-2 border-[#0B7EA1] rounded bg-white text-[#0B7EA1] font-black leading-none">A</button>
                   <button class="h-7 w-10 flex items-center justify-center border border-slate-300 rounded bg-slate-50 text-slate-700 font-bold leading-none text-base">A</button>
                </div>
             </div>

             <div ref="cardScrollEl" class="flex-1 overflow-auto">
                <!-- CONTENT: Stimulus (Stem) -->
                <div class="p-6 prose prose-slate max-w-none text-base leading-relaxed text-slate-900">
                   <div class="flex justify-between items-start mb-4">
                      <div class="text-[10px] font-black uppercase text-slate-400 tracking-[0.2em]">STIMULUS / SOAL UTAMA</div>
                      <button class="text-xs font-semibold underline text-slate-500 hover:text-slate-700" @click="goToEditor">Edit</button>
                   </div>
                   <div v-html="renderHtml(currentQuestion.stem)"></div>
                </div>
                
                <!-- INTERACTION Area -->
                <div class="px-6 pb-10 pt-6 border-t border-slate-200 bg-white">
                <div class="mb-8 flex items-center justify-between">
                   <div class="text-[10px] font-black uppercase text-slate-400 tracking-[0.2em]">
                      {{ currentQuestion.type === 'true_false' ? 'DAFTAR PERNYATAAN & JAWABAN' : 'OPSI JAWABAN' }}
                   </div>
                   
                   <!-- Preview Key Toggle -->
                   <div class="flex items-center gap-3 bg-white px-3 py-1.5 rounded-xl border border-slate-200 shadow-sm">
                      <span class="text-[10px] font-black uppercase text-indigo-500 tracking-widest">TAMPILKAN KUNCI</span>
                      <label class="relative inline-flex items-center cursor-pointer scale-75">
                         <input type="checkbox" v-model="showCorrect" class="sr-only peer">
                         <div class="w-11 h-6 bg-slate-200 rounded-full peer peer-checked:bg-emerald-600 after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:after:translate-x-full shadow-inner"></div>
                      </label>
                   </div>
                </div>

                <!-- MC Single -->
                <div v-if="currentQuestion.type === 'mc_single'" class="space-y-4">
                   <label
                     v-for="opt in currentQuestion.options" 
                     :key="opt.id"
                     class="flex items-center gap-4 p-4 rounded border transition-colors cursor-pointer bg-white shadow-sm text-left w-full"
                     :class="[
                        answers[currentQuestion.id] === String(opt.id) ? 'border-[#0B7EA1] bg-[#0B7EA1]/[0.03]' : 'border-slate-200 hover:border-slate-300',
                        showCorrect && opt.is_correct ? 'border-emerald-600 bg-emerald-50/50 ring-4 ring-emerald-50' : ''
                     ]"
                   >
                      <input
                        type="radio"
                        :name="'mc-'+currentQuestion.id"
                        class="h-4 w-4 accent-slate-600"
                        :checked="answers[currentQuestion.id] === String(opt.id)"
                        @change="answers[currentQuestion.id] = String(opt.id)"
                      />
                      <div class="text-slate-800 font-medium text-sm leading-relaxed" v-html="renderHtml(opt.content)"></div>
                   </label>
                   <div v-if="!currentQuestion.options?.length" class="p-10 text-center text-slate-400 italic bg-white rounded-2xl border-2 border-dashed border-slate-100">
                      Tidak ada opsi pilihan ganda yang tersedia.
                   </div>
                </div>

                <!-- MC Multiple (PG Kompleks) -->
                <div v-else-if="currentQuestion.type === 'mc_multiple'" class="space-y-4">
                   <button
                     type="button"
                     v-for="opt in currentQuestion.options" 
                     :key="opt.id"
                     @click="toggleMulti(currentQuestion.id, opt.id)"
                     class="flex items-start gap-4 p-4 rounded border transition-colors cursor-pointer bg-white shadow-sm text-left w-full"
                     :aria-pressed="answers[currentQuestion.id]?.includes(String(opt.id)) ? 'true' : 'false'"
                     :class="[
                        answers[currentQuestion.id]?.includes(String(opt.id)) ? 'border-[#0B7EA1] bg-[#0B7EA1]/[0.03]' : 'border-slate-200 hover:border-slate-300',
                        showCorrect && opt.is_correct ? 'border-emerald-600 bg-emerald-50/50 ring-4 ring-emerald-50' : ''
                     ]"
                   >
                      <div class="pt-0.5">
                        <div
                          class="h-5 w-5 rounded border flex items-center justify-center"
                          :class="answers[currentQuestion.id]?.includes(String(opt.id))
                            ? 'bg-[#0B7EA1] border-[#0B7EA1] text-white'
                            : (showCorrect && opt.is_correct ? 'bg-emerald-600 border-emerald-600 text-white' : 'bg-white border-slate-300 text-transparent')"
                        >
                          <span class="text-xs font-black leading-none">✓</span>
                        </div>
                      </div>
                      <div class="text-slate-800 font-medium text-sm leading-relaxed" v-html="renderHtml(opt.content)"></div>
                   </button>
                   <div v-if="!currentQuestion.options?.length" class="p-10 text-center text-slate-400 italic bg-white rounded-2xl border-2 border-dashed border-slate-100">
                      Tidak ada opsi PG Kompleks yang tersedia.
                   </div>
                </div>

	                <!-- True/False Statements -->
	                <div v-else-if="currentQuestion.type === 'true_false'" class="space-y-4">
	                   <div v-if="currentQuestion.statements?.length" class="bg-white rounded-2xl border-2 border-slate-200 overflow-hidden shadow-sm">
	                      <table class="w-full table-fixed border-collapse text-sm border-2 border-slate-300">
	                         <colgroup>
	                           <col class="w-[76%]" />
	                           <col class="w-[12%]" />
	                           <col class="w-[12%]" />
	                         </colgroup>
	                         <thead class="bg-slate-50/80 font-black text-[10px] text-slate-500 uppercase tracking-widest">
	                            <tr>
	                               <th class="py-4 px-6 text-left border-2 border-slate-300">Pernyataan / Pertanyaan</th>
	                               <th class="py-4 px-3 text-center border-2 border-slate-300">Benar</th>
	                               <th class="py-4 px-3 text-center border-2 border-slate-300">Salah</th>
	                            </tr>
	                         </thead>
	                         <tbody>
	                            <tr v-for="st in currentQuestion.statements" :key="st.id" class="hover:bg-slate-50/30 transition-colors">
	                               <td class="py-6 px-6 text-slate-900 font-medium align-middle border-2 border-slate-300 leading-relaxed" v-html="renderHtml(st.content)"></td>
	                               <td class="py-6 px-3 text-center align-middle border-2 border-slate-300" :class="showCorrect && st.correct === true ? 'bg-emerald-50/50' : ''">
	                                  <div class="flex flex-col items-center gap-2">
	                                     <input type="radio" :name="'st-'+st.id" :checked="answers[currentQuestion.id]?.[st.id] === true" @change="setTFStatement(currentQuestion.id, st.id, true)" class="h-6 w-6 accent-[#0B7EA1]" />
	                                     <span v-if="showCorrect && st.correct === true" class="text-[8px] font-black text-emerald-600 uppercase">Kunci</span>
	                                  </div>
	                                </td>
	                                <td class="py-6 px-3 text-center align-middle border-2 border-slate-300" :class="showCorrect && st.correct === false ? 'bg-emerald-50/50' : ''">
	                                   <div class="flex flex-col items-center gap-2">
	                                      <input type="radio" :name="'st-'+st.id" :checked="answers[currentQuestion.id]?.[st.id] === false" @change="setTFStatement(currentQuestion.id, st.id, false)" class="h-6 w-6 accent-[#0B7EA1]" />
	                                      <span v-if="showCorrect && st.correct === false" class="text-[8px] font-black text-emerald-600 uppercase">Kunci</span>
	                                   </div>
	                                </td>
	                             </tr>
	                          </tbody>
	                       </table>
	                   </div>
	                   
	                   <!-- Legacy Single Statement -->
	                   <div v-else class="bg-white rounded-2xl border-2 border-slate-200 overflow-hidden shadow-sm">
	                      <table class="w-full table-fixed border-collapse text-sm border-2 border-slate-300">
	                         <colgroup>
	                           <col class="w-[76%]" />
	                           <col class="w-[12%]" />
	                           <col class="w-[12%]" />
	                         </colgroup>
	                         <thead class="bg-slate-50/80 font-black text-[10px] text-slate-500 uppercase tracking-widest">
	                            <tr>
	                               <th class="py-4 px-6 text-left border-2 border-slate-300">Pernyataan / Pertanyaan</th>
	                               <th class="py-4 px-3 text-center border-2 border-slate-300">Benar</th>
	                               <th class="py-4 px-3 text-center border-2 border-slate-300">Salah</th>
	                            </tr>
	                         </thead>
	                         <tbody>
	                            <tr>
	                               <!-- Don't duplicate stimulus/stem here; T/F statements should come from `statements`. -->
	                               <td class="py-8 px-6 text-slate-500 italic font-semibold align-middle border-2 border-slate-300 leading-relaxed">
	                                  Lihat stimulus / soal utama di atas.
	                               </td>
	                               <td class="py-8 px-3 text-center align-middle border-2 border-slate-300" :class="showCorrect && (currentQuestion.correct === true || currentQuestion.true_false?.correct === true) ? 'bg-emerald-50/50' : ''">
	                                  <div class="flex flex-col items-center gap-2">
	                                     <input
	                                       type="radio"
                                       :name="'tf-legacy-'+currentQuestion.id"
                                       :checked="answers[currentQuestion.id] === true"
                                       @change="answers[currentQuestion.id] = true"
                                       class="h-6 w-6 accent-[#0B7EA1]"
                                     />
	                                     <span v-if="showCorrect && (currentQuestion.correct === true || currentQuestion.true_false?.correct === true)" class="text-[8px] font-black text-emerald-600 uppercase">Kunci</span>
	                                  </div>
	                               </td>
	                               <td class="py-8 px-3 text-center align-middle border-2 border-slate-300" :class="showCorrect && (currentQuestion.correct === false || currentQuestion.true_false?.correct === false) ? 'bg-emerald-50/50' : ''">
	                                  <div class="flex flex-col items-center gap-2">
	                                     <input
	                                       type="radio"
                                       :name="'tf-legacy-'+currentQuestion.id"
                                       :checked="answers[currentQuestion.id] === false"
                                       @change="answers[currentQuestion.id] = false"
                                       class="h-6 w-6 accent-[#0B7EA1]"
                                     />
	                                     <span v-if="showCorrect && (currentQuestion.correct === false || currentQuestion.true_false?.correct === false)" class="text-[8px] font-black text-emerald-600 uppercase">Kunci</span>
	                                  </div>
	                               </td>
	                            </tr>
	                         </tbody>
	                      </table>
	                   </div>
	                </div>

                <!-- Short Answer -->
                <div v-else-if="currentQuestion.type === 'short_answer'" class="space-y-6">
                  <QuillEditor
                    v-model="shortAnswerEditorHtml"
                    :height="120"
                    :enable-math="true"
                    placeholder="Ketik jawaban isian singkat..."
                    @update:model-value="onShortAnswerEditorUpdate"
                  />
                  <div class="text-[11px] font-semibold text-slate-500">
                    Preview admin/guru: formula ditulis via tombol `fx`, lalu disimpan sebagai teks jawaban singkat.
                  </div>
                   <div v-if="showCorrect" class="p-8 rounded-2xl bg-emerald-50 border-2 border-emerald-100">
                      <div class="text-[10px] font-black uppercase text-emerald-600 tracking-widest mb-3">Kunci Jawaban (Diterima):</div>
                      <div class="space-y-2">
                         <div v-for="ans in currentQuestion.answers" :key="ans.id" class="flex items-center gap-3">
                            <BaseIcon :path="mdiCheckCircleOutline" size="18" class="text-emerald-500" />
                            <span class="text-xl font-black text-emerald-800" v-html="renderAcceptedAnswerHtml(ans.answer_text)"></span>
                         </div>
                      </div>
                   </div>
                </div>

                <!-- Essay -->
                <div v-else-if="currentQuestion.type === 'essay'" class="space-y-6">
                  <QuillEditor
                    v-model="essayEditorHtml"
                    :height="260"
                    :enable-math="true"
                    placeholder="Masukkan jawaban uraian secara detail..."
                    @update:model-value="onEssayEditorUpdate"
                  />
                   <div v-if="showCorrect && (currentQuestion.essay?.rubric_text || currentQuestion.essay?.max_score != null)" class="p-8 rounded-2xl bg-emerald-50 border-2 border-emerald-100">
                      <div class="text-[10px] font-black uppercase text-emerald-600 tracking-widest mb-3">Rubrik Penilaian</div>
                      <div v-if="currentQuestion.essay?.rubric_text" class="text-slate-800 font-semibold leading-relaxed" v-html="renderHtml(currentQuestion.essay.rubric_text)"></div>
                      <div v-if="currentQuestion.essay?.max_score != null" class="mt-4 text-[11px] font-black uppercase tracking-widest text-emerald-800">Max Score: {{ currentQuestion.essay.max_score }}</div>
                   </div>
                </div>

                <!-- Matching -->
                <div v-else-if="currentQuestion.type === 'matching'" class="space-y-6">
                   <div v-if="!currentQuestion.pairs?.length" class="p-10 bg-white rounded-2xl border-2 border-dashed border-slate-200 text-center text-slate-400 italic">
                      Tidak ada pasangan (pairs) untuk soal menjodohkan ini.
                   </div>

                   <div v-else class="bg-white rounded-2xl border-2 border-slate-100 shadow-inner overflow-hidden">
                      <div class="px-8 py-6 border-b border-slate-100 bg-slate-50/70 flex items-center justify-between">
                         <div class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500">MENJODOHKAN</div>
                         <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Pilih pasangan yang sesuai</div>
                      </div>

                      <div class="p-8 space-y-4">
                         <div
                           v-for="pair in currentQuestion.pairs"
                           :key="pair.id"
                           class="grid md:grid-cols-[1fr_24px_1fr] gap-4 items-center p-6 rounded-2xl border-2 border-slate-100 bg-white"
                           :class="[
                             showCorrect && answers[currentQuestion.id]?.[pair.id] && answers[currentQuestion.id]?.[pair.id] !== pair.id ? '!border-red-200 bg-red-50/30' : '',
                             showCorrect && answers[currentQuestion.id]?.[pair.id] === pair.id ? '!border-emerald-200 bg-emerald-50/30' : ''
                           ]"
                         >
                            <div class="font-black text-slate-900 text-lg leading-snug" v-html="renderHtml(pair.left_content)"></div>
                            <div class="text-center text-slate-300 font-black">→</div>

                            <div class="space-y-2">
                               <select
                                 class="w-full p-4 rounded-xl border-2 border-slate-100 bg-white focus:border-blue-500 outline-none transition-all font-bold text-slate-900"
                                 :value="answers[currentQuestion.id]?.[pair.id] || ''"
                                 @change="setMatching(currentQuestion.id, pair.id, $event.target.value)"
                               >
                                  <option value="" disabled>Pilih jawaban...</option>
                                  <option v-for="opt in matchingRightOptions" :key="opt.id" :value="opt.id">
                                     {{ stripHtml(opt.content).slice(0, 80) }}
                                  </option>
                               </select>

                               <div v-if="showCorrect" class="text-[11px] font-black uppercase tracking-widest"
                                 :class="answers[currentQuestion.id]?.[pair.id] === pair.id ? 'text-emerald-700' : 'text-slate-500'"
                               >
                                  Kunci: <span class="normal-case font-extrabold" v-html="renderHtml(pair.right_content)"></span>
                               </div>
                            </div>
                         </div>
                      </div>
                   </div>
                </div>

                <!-- Unexpected Type -->
                <div v-else class="p-10 bg-slate-100 rounded-2xl border-2 border-slate-200 text-center italic text-slate-400">
                   Tipe soal tidak dikenali: {{ currentQuestion.type }}
                </div>
                </div>
             </div>
          </div>

          <!-- RIGHT COLUMN: Sidebar (Navigator) -->
          <aside class="hidden lg:block bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden sticky top-[74px]">
             <div class="px-6 py-4 border-b border-slate-200 bg-white">
                <span class="text-slate-800 font-bold uppercase text-sm tracking-wide select-none">DAFTAR SOAL</span>
             </div>
             <div class="p-5">
                <div class="grid grid-cols-5 gap-2">
                   <button
                      v-for="(q, idx) in questions"
                      :key="q.id"
                      :data-qnav-idx="idx"
                      class="h-10 w-10 flex items-center justify-center rounded border font-semibold text-sm transition-colors"
                      :class="currentIndex === idx
                        ? 'border-[#0B7EA1] ring-1 ring-[#0B7EA1] text-[#0B7EA1] bg-white'
                        : (flagged[q.id]
                          ? 'bg-[#F4C20D] border-[#F4C20D] text-white'
                          : (isAnswered(q.id)
                            ? 'bg-emerald-500 border-emerald-500 text-white'
                            : 'bg-white border-slate-200 text-slate-600 hover:border-slate-300'))"
                      @click="setIndex(idx)"
                   >
                      {{ idx + 1 }}
                   </button>
                </div>
                
                <div class="mt-6 pt-4 border-t border-slate-200 grid grid-cols-2 gap-3 text-xs text-slate-600">
                  <div class="flex items-center gap-2"><span class="h-3 w-3 bg-emerald-500 rounded-sm"></span>Sudah</div>
                  <div class="flex items-center gap-2"><span class="h-3 w-3 bg-[#F4C20D] rounded-sm"></span>Ragu</div>
                  <div class="flex items-center gap-2"><span class="h-3 w-3 border-2 border-[#0B7EA1] rounded-sm"></span>Dibuka</div>
                  <div class="flex items-center gap-2"><span class="h-3 w-3 border border-slate-300 rounded-sm bg-white"></span>Belum</div>
                </div>
             </div>
          </aside>
       </div>
    </main>

    <div v-else class="max-w-xl mx-auto mt-20 p-12 bg-white rounded-3xl border-2 border-slate-100 shadow-2xl text-center">
      <BaseIcon :path="mdiAlert" size="64" class="text-amber-500 mb-6 mx-auto" />
      <h2 class="text-2xl font-black text-slate-800 mb-4 uppercase">Soal Tidak Ditemukan</h2>
      <p class="text-slate-600 mb-8 leading-relaxed">
        Index: {{ currentIndex + 1 }} / {{ questions.length }}. Klik tombol di bawah untuk kembali ke soal pertama.
      </p>
      <button
        @click="currentIndex = 0"
        class="bg-[#0D47A1] text-white px-8 py-3 rounded-xl font-bold uppercase transition-all hover:scale-105 active:scale-95"
      >
        Reset Ke Soal 1
      </button>
    </div>

    <!-- FOOTER NAVIGATION -->
    <footer v-if="!isLoading && currentQuestion" class="bg-white border-t border-slate-200 p-4 fixed bottom-0 left-0 right-0 z-[100] shadow-[0_-12px_30px_-18px_rgba(0,0,0,0.3)]">
       <div class="max-w-[1500px] mx-auto w-full flex items-center justify-between">
          <button 
             class="bg-[#E74C3C] text-white px-8 py-2.5 rounded font-bold uppercase text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all disabled:opacity-50"
             :disabled="currentIndex === 0"
             @click="goPrev"
          >
             Soal Sebelumnya
          </button>

          <button
            type="button"
            class="flex items-center gap-3 bg-[#F4C20D] text-white px-10 py-2.5 rounded font-bold uppercase text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all"
            @click="toggleFlagged"
          >
            <span class="h-4 w-4 rounded border border-white/70 bg-white/10 flex items-center justify-center">
              <span v-if="flagged[currentQuestion?.id]" class="text-[10px] font-black leading-none">✓</span>
            </span>
            Ragu-Ragu
          </button>

          <button 
             class="bg-[#0B7EA1] text-white px-8 py-2.5 rounded font-bold uppercase text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all"
             @click="goNext"
          >
             {{ currentIndex === questions.length - 1 ? 'Selesai' : 'Soal Berikutnya' }}
          </button>
       </div>
    </footer>
  </div>
</template>

<script setup>
import { computed, nextTick, onErrorCaptured, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  mdiArrowLeft,
  mdiClockOutline,
  mdiCheckCircleOutline,
  mdiAlert
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import QuillEditor from '@/components/QuillEditor.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()

const setId = computed(() => String(route.params.id || ''))
const currentIndex = ref(0)
const isLoading = ref(true)
const showCorrect = ref(false)
const errorMessage = ref('')
const questions = ref([])
const questionSet = ref(null)
const answers = reactive({})
const flagged = reactive({})
const timeLeft = ref(3599) // 59:59 in seconds

const formattedTime = computed(() => {
  const m = Math.floor(timeLeft.value / 60)
  const s = timeLeft.value % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

let timerInterval = null
const startTimer = () => {
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(() => {
    if (timeLeft.value > 0) {
      timeLeft.value--
    } else {
      clearInterval(timerInterval)
    }
  }, 1000)
}

const currentQuestion = computed(() => questions.value[currentIndex.value] || null)
const cardScrollEl = ref(null)
const shortAnswerEditorHtml = ref('')
const essayEditorHtml = ref('')

const stripHtml = (html) => String(html || '').replace(/<[^>]*>/g, '').trim()
const escapeHtml = (value) =>
  String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')

const looksLikeLatex = (value) => {
  const text = String(value || '').trim()
  if (!text) return false
  return /\\[a-zA-Z]+|[_^{}]|\\frac|\\sqrt|\\times|\\cdot|\\left|\\right|\\sum|\\int|\\pi|\\alpha|\\beta|\\theta/.test(text)
}

const renderHtml = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return '<p><em>(Konten soal kosong)</em></p>'
  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    doc.querySelectorAll('script, iframe, object, embed').forEach(n => n.remove())
    doc.querySelectorAll('*').forEach(el => {
      for (const attr of Array.from(el.attributes || [])) {
        if (/^on/i.test(attr.name)) el.removeAttribute(attr.name)
      }
      const style = el.getAttribute && el.getAttribute('style')
      if (style) {
        const parts = style.split(';').map(s => s.trim()).filter(Boolean)
        const filtered = parts.filter(p => {
          const name = (p.split(':')[0] || '').trim().toLowerCase()
          return name && name !== 'color' && name !== 'background' && name !== 'background-color'
        })
        if (filtered.length) el.setAttribute('style', filtered.join('; '))
        else el.removeAttribute('style')
      }
    })
    return doc.body.innerHTML
  } catch {
    return raw
  }
}

const renderAcceptedAnswerHtml = (value) => {
  const text = String(value || '').trim()
  if (!text) return '<span class="italic text-slate-400">(Kunci kosong)</span>'
  if (looksLikeLatex(text)) return `\\(${escapeHtml(fixCommonLatexCommands(text))}\\)`
  return escapeHtml(text)
}

const textToQuillHtml = (text) => {
  const raw = String(text || '')
  if (!raw.trim()) return ''
  return `<p>${escapeHtml(raw)}</p>`
}

const quillHtmlToPlainText = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return ''
  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    doc.querySelectorAll('.ql-formula[data-value]').forEach((node) => {
      const latex = String(node.getAttribute('data-value') || '').trim()
      node.replaceWith(doc.createTextNode(latex))
    })
    const text = String(doc.body.textContent || '')
    return text.replace(/\s+/g, ' ').trim()
  } catch {
    return stripHtml(raw)
  }
}

const syncPreviewEditorModel = () => {
  const q = currentQuestion.value
  if (!q?.id) {
    shortAnswerEditorHtml.value = ''
    essayEditorHtml.value = ''
    return
  }
  const qid = String(q.id || '')
  const value = answers[qid]
  if (q.type === 'short_answer') {
    shortAnswerEditorHtml.value = textToQuillHtml(String(value || ''))
    essayEditorHtml.value = ''
    return
  }
  if (q.type === 'essay') {
    essayEditorHtml.value = String(value || '')
    shortAnswerEditorHtml.value = ''
    return
  }
  shortAnswerEditorHtml.value = ''
  essayEditorHtml.value = ''
}

const onShortAnswerEditorUpdate = (html) => {
  shortAnswerEditorHtml.value = String(html || '')
  const q = currentQuestion.value
  if (!q?.id) return
  answers[String(q.id || '')] = quillHtmlToPlainText(shortAnswerEditorHtml.value)
}

const onEssayEditorUpdate = (html) => {
  essayEditorHtml.value = String(html || '')
  const q = currentQuestion.value
  if (!q?.id) return
  answers[String(q.id || '')] = essayEditorHtml.value
}

const matchingRightOptions = computed(() => {
  const q = currentQuestion.value
  if (!q || q.type !== 'matching') return []
  const pairs = Array.isArray(q.pairs) ? q.pairs : []
  return pairs.map(p => ({ id: p.id, content: p.right_content }))
})

const toggleMulti = (questionId, optId) => {
  const qid = String(questionId || '')
  if (!qid) return
  const id = String(optId || '')
  if (!id) return
  if (!Array.isArray(answers[qid])) answers[qid] = []
  const arr = answers[qid]
  const i = arr.indexOf(id)
  if (i >= 0) arr.splice(i, 1)
  else arr.push(id)
}

const setTFStatement = (questionId, statementId, val) => {
  const qid = String(questionId || '')
  const sid = String(statementId || '')
  if (!qid || !sid) return
  if (!answers[qid] || Array.isArray(answers[qid])) answers[qid] = {}
  answers[qid][sid] = !!val
}

const setMatching = (questionId, leftPairId, rightPickId) => {
  const qid = String(questionId || '')
  const pid = String(leftPairId || '')
  if (!qid || !pid) return
  if (!answers[qid] || Array.isArray(answers[qid])) answers[qid] = {}
  answers[qid][pid] = String(rightPickId || '')
}

const setIndex = (idx) => {
  const n = Number(idx)
  if (!Number.isFinite(n)) return
  if (n < 0 || n >= questions.value.length) return
  currentIndex.value = n
}

const goPrev = () => { if (currentIndex.value > 0) currentIndex.value -= 1 }
const goNext = () => { if (currentIndex.value < questions.value.length - 1) currentIndex.value += 1 }

const toggleFlagged = () => {
  const q = currentQuestion.value
  if (!q) return
  const id = String(q.id || '')
  if (!id) return
  flagged[id] = !flagged[id]
}

const isAnswered = (qid) => {
  const id = String(qid || '')
  const v = answers[id]
  if (v === undefined || v === null) return false
  if (typeof v === 'string') {
    const q = questions.value.find((item) => String(item?.id || '') === id)
    const t = String(q?.type || '')
    if (t === 'essay') return stripHtml(v).trim() !== ''
    if (t === 'short_answer') return quillHtmlToPlainText(textToQuillHtml(v)).trim() !== ''
    return v.trim() !== ''
  }
  if (typeof v === 'boolean') return true
  if (Array.isArray(v)) return v.length > 0
  if (typeof v === 'object') return Object.keys(v).length > 0
  return false
}

const loadPreviewData = async () => {
  if (!setId.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const questionsRes = await api.get(`/api/v1/question-sets/${setId.value}/questions`)
    questions.value = Array.isArray(questionsRes.data?.data) ? questionsRes.data.data : []
    if (currentIndex.value < 0) currentIndex.value = 0
    if (questions.value.length > 0 && currentIndex.value >= questions.value.length) {
      currentIndex.value = questions.value.length - 1
    }
    
    if (questions.value.length > 0) {
      const setRes = await api.get(`/api/v1/question-sets/${setId.value}`)
      questionSet.value = setRes.data?.data
    } else {
       errorMessage.value = 'Tidak ada soal dalam paket ini.'
    }
    scheduleRenderMath()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat pratinjau soal'
  } finally {
    isLoading.value = false
  }
}

let mathRaf = 0
const fixCommonLatexCommands = (input) => {
  const s = String(input || '')
  // Auto-fix common TeX commands when author forgot leading backslash, e.g. "frac{1}{2}".
  return s.replace(
    /(^|[^\\a-zA-Z])(frac|sqrt|times|cdot|pm|mp|div|leq|geq|neq|approx|sum|prod|int|lim|infty|pi|alpha|beta|gamma|theta|lambda|mu|sigma|omega|sin|cos|tan|log|ln)\b/g,
    '$1\\\\$2',
  )
}
const renderMath = (rootEl) => {
  if (!rootEl) return
  if (window.renderMathInElement) {
    window.renderMathInElement(rootEl, {
      delimiters: [
        { left: '$$', right: '$$', display: true },
        { left: '$', right: '$', display: false },
        { left: '\\(', right: '\\)', display: false },
        { left: '\\[', right: '\\]', display: true },
      ],
      throwOnError: false,
      preProcess: fixCommonLatexCommands,
    })
  }
}
const scheduleRenderMath = async () => {
  if (mathRaf) cancelAnimationFrame(mathRaf)
  await nextTick()
  mathRaf = requestAnimationFrame(() => {
    try {
      renderMath(cardScrollEl.value)
    } catch (e) {
      // Avoid crashing render; keep console signal for debugging.
      console.warn('renderMath failed:', e)
    }
  })
}

watch(currentIndex, () => {
  syncPreviewEditorModel()
  scheduleRenderMath()
  if (cardScrollEl.value && typeof cardScrollEl.value.scrollTo === 'function') {
    cardScrollEl.value.scrollTo({ top: 0, behavior: 'smooth' })
  } else if (cardScrollEl.value) {
    cardScrollEl.value.scrollTop = 0
  }
  // Keep the active number visible in the sidebar navigator when list is long.
  setTimeout(() => {
    const el = document.querySelector(`[data-qnav-idx="${currentIndex.value}"]`)
    if (el && typeof el.scrollIntoView === 'function') {
      el.scrollIntoView({ block: 'nearest', inline: 'nearest' })
    }
  }, 50)
}, { flush: 'post' })

watch(
  () => currentQuestion.value?.id,
  () => {
    syncPreviewEditorModel()
  },
  { immediate: true },
)

watch(showCorrect, () => {
  scheduleRenderMath()
}, { flush: 'post' })

onMounted(() => {
  loadPreviewData()
  startTimer()
})
onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
  if (mathRaf) cancelAnimationFrame(mathRaf)
})

onErrorCaptured((err) => {
  errorMessage.value = err?.message ? String(err.message) : 'Terjadi error saat merender soal'
  // Let Vue handle the error too (visible in console), but keep UI usable.
  return false
})

const backToList = () => {
  const role = route.path.startsWith('/admin') ? 'admin' : 'teacher'
  router.push(`/${role}/bank-soal`)
}

const goToEditor = () => {
  const role = route.path.startsWith('/admin') ? 'admin' : 'teacher'
  router.push(`/${role}/bank-soal/new?id=${setId.value}`)
}
</script>

<style scoped>
.tka-theme { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; }
.tka-topbar { background: #0b7ea1; }
.tka-timer { box-shadow: inset 0 0 0 2px #f59e0b; }
.animate-fade-in { animation: fadeIn 0.4s ease-out; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
:deep(.prose) { max-width: none; }
:deep(.prose img) { border-radius: 16px; display: block; margin: 32px auto; box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1); border: 8px solid #f8fafc; }
</style>
