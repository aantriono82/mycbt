<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import Quill from 'quill/core'
import { ClassAttributor, Scope, StyleAttributor } from 'parchment'
import SnowTheme from 'quill/themes/snow'
import Toolbar from 'quill/modules/toolbar'
import Bold from 'quill/formats/bold'
import Italic from 'quill/formats/italic'
import Underline from 'quill/formats/underline'
import Link from 'quill/formats/link'
import Formula from 'quill/formats/formula'
import { AlignClass, AlignStyle } from 'quill/formats/align'
import { DirectionClass, DirectionStyle } from 'quill/formats/direction'
import 'quill/dist/quill.snow.css'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  placeholder: {
    type: String,
    default: 'Tulis di sini...',
  },
  height: {
    type: Number,
    default: 200,
  },
  enableMath: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue', 'blur'])

const editorHostRef = ref(null)
const toolbarRef = ref(null)
const selectedFont = ref('arial')
let quillInstance = null
let textChangeHandler = null
let selectionChangeHandler = null
let applyingExternalValue = false
let quillRegistered = false

const fontWhitelist = ['arial', 'calibri', 'times-new-roman', 'courier-new', 'amiri', 'noto-naskh-arabic']
const FontClass = new ClassAttributor('font', 'ql-font', {
  scope: Scope.INLINE,
  whitelist: fontWhitelist,
})
const FontStyle = new StyleAttributor('font', 'font-family', {
  scope: Scope.INLINE,
  whitelist: fontWhitelist,
})

const getNormalizedHtml = (value) => String(value || '').trim()

const fontFamilyMap = {
  arial: 'Arial, Helvetica, sans-serif',
  calibri: 'Calibri, Candara, Segoe, "Segoe UI", Optima, Arial, sans-serif',
  'times-new-roman': '"Times New Roman", Times, serif',
  'courier-new': '"Courier New", Courier, monospace',
  amiri: '"Amiri", "Noto Naskh Arabic", serif',
  'noto-naskh-arabic': '"Noto Naskh Arabic", "Amiri", serif',
}

const applyFont = (fontValue) => {
  if (!quillInstance) return
  selectedFont.value = fontValue
  quillInstance.focus()
  quillInstance.format('font', fontValue)
}

onMounted(() => {
  if (!editorHostRef.value) return

  if (!quillRegistered) {
    Quill.register(
      {
        'themes/snow': SnowTheme,
        'modules/toolbar': Toolbar,
        'formats/bold': Bold,
        'formats/italic': Italic,
        'formats/underline': Underline,
        'formats/link': Link,
        'formats/font': FontClass,
        'attributors/class/font': FontClass,
        'attributors/style/font': FontStyle,
        'formats/formula': Formula,
        'formats/align': AlignClass,
        'formats/direction': DirectionClass,
        'attributors/class/align': AlignClass,
        'attributors/style/align': AlignStyle,
        'attributors/class/direction': DirectionClass,
        'attributors/style/direction': DirectionStyle,
      },
      true,
    )
    quillRegistered = true
  }

  quillInstance = new Quill(editorHostRef.value, {
    theme: 'snow',
    placeholder: props.placeholder,
    modules: {
      toolbar: {
        container: toolbarRef.value,
      },
    },
  })

  const editor = quillInstance.root
  editor.style.minHeight = `${props.height}px`
  editor.style.fontFamily = fontFamilyMap.arial
  editor.innerHTML = getNormalizedHtml(props.modelValue)

  textChangeHandler = () => {
    if (!quillInstance || applyingExternalValue) return
    emit('update:modelValue', quillInstance.root.innerHTML)
  }
  quillInstance.on('text-change', textChangeHandler)

  selectionChangeHandler = (range, oldRange) => {
    if (range && quillInstance) {
      const format = quillInstance.getFormat(range)
      selectedFont.value = typeof format.font === 'string' ? format.font : 'arial'
    }
    if (!oldRange || range) return
    emit('blur')
  }
  quillInstance.on('selection-change', selectionChangeHandler)
})

watch(
  () => props.modelValue,
  (nextValue) => {
    if (!quillInstance) return
    const incoming = getNormalizedHtml(nextValue)
    const current = getNormalizedHtml(quillInstance.root.innerHTML)
    if (incoming === current) return

    applyingExternalValue = true
    quillInstance.root.innerHTML = incoming
    applyingExternalValue = false
  },
)

onBeforeUnmount(() => {
  if (quillInstance && textChangeHandler) {
    quillInstance.off('text-change', textChangeHandler)
  }
  if (quillInstance && selectionChangeHandler) {
    quillInstance.off('selection-change', selectionChangeHandler)
  }
  quillInstance = null
  textChangeHandler = null
  selectionChangeHandler = null
})
</script>

<template>
  <div class="max-w-full min-w-0 overflow-hidden rounded-xl border border-slate-200 bg-white p-2 dark:border-slate-800 dark:bg-slate-900">
    <div class="mycbt-toolbar-scroll mb-2 flex max-w-full min-w-0 flex-nowrap items-center gap-2 overflow-x-auto overflow-y-hidden rounded-t-md border border-slate-200 bg-white px-2 py-2">
      <select
        class="mycbt-font-select rounded-md border border-slate-200 bg-white px-3 py-1.5 text-sm text-slate-700"
        :value="selectedFont"
        @change="applyFont($event.target.value)"
      >
        <option value="arial">Arial</option>
        <option value="calibri">Calibri</option>
        <option value="times-new-roman">Times New Roman</option>
        <option value="courier-new">Courier New</option>
        <option value="amiri">Arabic Traditional (Amiri)</option>
        <option value="noto-naskh-arabic">Noto Naskh Arabic</option>
      </select>
      <span ref="toolbarRef" class="mycbt-quill-toolbar ql-toolbar ql-snow">
        <span class="ql-formats">
          <button class="ql-bold"></button>
          <button class="ql-italic"></button>
          <button class="ql-underline"></button>
        </span>
        <span class="ql-formats">
          <select class="ql-align"></select>
          <button class="ql-direction" value="rtl"></button>
        </span>
        <span class="ql-formats" v-if="enableMath">
          <button class="ql-formula"></button>
        </span>
        <span class="ql-formats">
          <button class="ql-link"></button>
          <button class="ql-clean"></button>
        </span>
      </span>
    </div>
    <div ref="editorHostRef"></div>
  </div>
</template>

<style scoped>
.mycbt-font-select {
  flex: 0 0 auto;
  min-width: 190px;
  height: 36px;
  line-height: 1.2;
}

.mycbt-toolbar-scroll {
  scrollbar-width: thin;
}

.mycbt-toolbar-scroll::-webkit-scrollbar {
  height: 8px;
}

.mycbt-toolbar-scroll::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgb(203 213 225);
}

.mycbt-quill-toolbar {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  padding: 0 !important;
  border: 1px solid rgb(226 232 240) !important;
  border-radius: 0.375rem;
  background: white;
}

.mycbt-quill-toolbar :deep(.ql-formats) {
  display: inline-flex;
  align-items: center;
  margin-right: 0;
  padding: 0 4px;
  border-right: 1px solid rgb(226 232 240);
}

.mycbt-quill-toolbar :deep(.ql-formats:last-child) {
  border-right: 0;
}

.mycbt-quill-toolbar :deep(button) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.mycbt-quill-toolbar :deep(.ql-picker.ql-align) {
  width: 28px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.mycbt-quill-toolbar :deep(.ql-picker.ql-align .ql-picker-label) {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 0;
}

.mycbt-quill-toolbar :deep(.ql-picker.ql-align .ql-picker-label svg) {
  width: 18px;
  height: 18px;
  position: static;
  margin: 0;
}

.mycbt-quill-toolbar :deep(.ql-picker.ql-align .ql-picker-item) {
  display: flex;
  align-items: center;
  justify-content: center;
}

.mycbt-quill-toolbar :deep(.ql-picker.ql-align .ql-picker-item svg) {
  width: 18px;
  height: 18px;
  position: static;
}
</style>
