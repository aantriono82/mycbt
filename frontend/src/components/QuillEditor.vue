<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import Quill from 'quill/core'
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
let quillInstance = null
let textChangeHandler = null
let selectionChangeHandler = null
let applyingExternalValue = false
let quillRegistered = false

const getNormalizedHtml = (value) => String(value || '').trim()

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
        'formats/formula': Formula,
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
      toolbar: props.enableMath
        ? [
            ['bold', 'italic', 'underline'],
            [{ align: [] }, { direction: 'rtl' }],
            ['formula', 'link', 'clean'],
          ]
        : [
            ['bold', 'italic', 'underline'],
            [{ align: [] }, { direction: 'rtl' }],
            ['link', 'clean'],
          ],
    },
  })

  const editor = quillInstance.root
  editor.style.minHeight = `${props.height}px`
  editor.style.fontFamily = 'Arial, "Noto Naskh Arabic", "Amiri", sans-serif'
  editor.innerHTML = getNormalizedHtml(props.modelValue)

  textChangeHandler = () => {
    if (!quillInstance || applyingExternalValue) return
    emit('update:modelValue', quillInstance.root.innerHTML)
  }
  quillInstance.on('text-change', textChangeHandler)

  selectionChangeHandler = (range, oldRange) => {
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
  <div class="rounded-xl border border-slate-200 bg-white p-2 dark:border-slate-800 dark:bg-slate-900">
    <div ref="editorHostRef"></div>
  </div>
</template>
