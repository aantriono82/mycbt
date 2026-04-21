<script setup>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { api } from '@/services/api.js'
import katex from 'katex'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  placeholder: {
    type: String,
    default: 'Tulis disini...',
  },
  height: {
    type: Number,
    default: 200,
  },
  toolbar: {
    type: String,
    default: 'undo redo | blocks | bold italic underline forecolor | bullist numlist | table image link | fullscreen',
  },
})

const emit = defineEmits(['update:modelValue'])

const editorRef = ref(null)
const editorInstance = ref(null)
const uid = ref(`tiny-${Math.random().toString(36).slice(2, 9)}`)
const mathMlElements = [
  'math', 'semantics', 'annotation', 'annotation-xml',
  'mrow', 'mi', 'mn', 'mo',
  'msqrt', 'mroot', 'mfrac',
  'msub', 'msup', 'msubsup',
  'munder', 'mover', 'munderover',
  'mtable', 'mtr', 'mtd',
  'mstyle', 'mspace', 'mtext',
  'mpadded', 'menclose', 'mfenced',
]

const extendedMathElements = mathMlElements.map(tag => `${tag}[*]`).join(',')
const customMathElements = mathMlElements.join(',')

const escapeAttr = (value) =>
  String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

const uploadImageFile = async (file) => {
  const formData = new FormData()
  formData.append('file', file)

  const { data } = await api.post('/api/v1/uploads/images', formData)
  let url = data?.data?.url
  if (!url) throw new Error('Upload gagal: URL kosong')
  if (url.startsWith('/uploads')) {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    url = `${baseUrl}${url}`
  }
  return url
}

const buildMathHtml = (latex, displayMode) => {
  const clean = String(latex || '').trim()
  if (!clean) throw new Error('LaTeX kosong')
  const rendered = katex.renderToString(clean, {
    throwOnError: true,
    displayMode: !!displayMode,
    // Render as native MathML to avoid dependency on KaTeX webfonts inside TinyMCE iframe.
    output: 'mathml',
  })

  if (displayMode) {
    return `<div class="math-tex" data-latex="${escapeAttr(clean)}" data-display="1" contenteditable="false">${rendered}</div><p></p>`
  }

  return `<span class="math-tex" data-latex="${escapeAttr(clean)}" data-display="0" contenteditable="false">${rendered}</span>&nbsp;`
}

const openMathDialog = (editor, initialLatex = '', initialDisplay = false, targetEl = null) => {
  editor.windowManager.open({
    title: targetEl ? 'Edit LaTeX' : 'Insert LaTeX',
    body: {
      type: 'panel',
      items: [
        { type: 'textarea', name: 'latex', label: 'LaTeX (tanpa tanda $)' },
        { type: 'checkbox', name: 'display', label: 'Display mode (blok)' },
      ],
    },
    initialData: {
      latex: initialLatex,
      display: !!initialDisplay,
    },
    buttons: [
      { type: 'cancel', text: 'Batal' },
      { type: 'submit', text: targetEl ? 'Simpan' : 'Insert', primary: true },
    ],
    onSubmit: (apiDialog) => {
      const data = apiDialog.getData()
      const latex = String(data.latex || '')
        .trim()
        // Normalize \\command → \command (common user mistake)
        .replace(/\\\\([a-zA-Z]+)/g, '\\$1')
        .replace(/^\$+|\$+$/g, '')
      const display = !!data.display

      try {
        const mathHtml = buildMathHtml(latex, display)
        if (targetEl) {
          targetEl.outerHTML = mathHtml
        } else {
          editor.insertContent(mathHtml)
        }
        apiDialog.close()
      } catch (err) {
        console.error('KaTeX render error:', err)
        alert(err?.message || 'Gagal render LaTeX. Periksa penulisan LaTeX Anda.')
      }
    },
  })
}

const initEditor = async () => {
  let attempts = 0
  while (!window.tinymce && attempts < 50) {
    await new Promise(r => setTimeout(r, 100))
    attempts++
  }
  if (!window.tinymce) return

  window.tinymce.init({
    selector: `#${uid.value}`,
    height: props.height,
    menubar: false,
    plugins: [
      'advlist', 'autolink', 'lists', 'link', 'image', 'charmap', 'preview',
      'anchor', 'searchreplace', 'visualblocks', 'code', 'fullscreen',
      'insertdatetime', 'media', 'table', 'help', 'wordcount', 'emoticons'
    ],
    toolbar: props.toolbar,
    toolbar_mode: 'wrap',
    min_height: props.height,
    custom_elements: customMathElements,
    extended_valid_elements: [
      'span[class|data-latex|data-display|contenteditable]',
      'div[class|data-latex|data-display|contenteditable]',
      extendedMathElements,
    ].join(','),
    valid_children: '+span[math],+div[math],+math[semantics|mrow|mi|mn|mo|msqrt|mroot|mfrac|msub|msup|msubsup|munder|mover|munderover|mtable|mtr|mtd|mstyle|mspace|mtext|mpadded|menclose|mfenced]',
    content_css: ['default'],
    content_style: `
      body { font-family:Inter,Helvetica,Arial,sans-serif; font-size:14px; color: #475569; padding: 10px; }
      .math-tex { display: inline-flex; align-items: center; vertical-align: middle; white-space: nowrap; cursor: pointer; }
      .math-tex math { font-size: 1em; line-height: 1.2; }
      .math-tex[data-display="1"] { display: block; white-space: normal; margin: 0.5rem 0; }
      .math-tex[data-display="1"] math { display: block; margin: 0; }
    `,
    automatic_uploads: true,
    file_picker_types: 'image',
    images_file_types: 'jpg,jpeg,png,gif,webp',
    images_upload_handler: async (blobInfo) => {
      const blob = blobInfo.blob()
      const file = new File([blob], blobInfo.filename(), { type: blob.type })
      return uploadImageFile(file)
    },
    file_picker_callback: (cb, _value, meta) => {
      if (meta.filetype !== 'image') return
      const input = document.createElement('input')
      input.type = 'file'
      input.accept = 'image/*'
      input.onchange = async () => {
        const file = input.files?.[0]
        if (!file) return
        try {
          const url = await uploadImageFile(file)
          cb(url, { title: file.name })
        } catch (err) {
          console.error('Upload image failed', err)
          alert('Gagal upload gambar. Coba lagi.')
        }
      }
      input.click()
    },
    setup: (editor) => {
      editorInstance.value = editor

      // Add custom Sigma icon
      editor.ui.registry.addIcon('sigma', '<svg width="24" height="24" viewBox="0 0 24 24"><path d="M18 6H8.83l6 6-6 6H18v2H6v-2l6-6-6-6V4h12v2z"/></svg>')

      // Add custom Math (Sigma) button
      editor.ui.registry.addButton('math', {
        text: '',
        icon: 'sigma',
        tooltip: 'Insert Math (LaTeX)',
        onAction: () => openMathDialog(editor),
      })

      // Double-click an existing math node to edit
      editor.on('DblClick', (e) => {
        const target = e?.target
        if (!target || !target.closest) return
        const mathEl = target.closest('.math-tex')
        if (!mathEl) return
        const latex = mathEl.getAttribute('data-latex') || ''
        const display = mathEl.getAttribute('data-display') === '1'
        openMathDialog(editor, latex, display, mathEl)
      })

      editor.on('change input undo redo', () => {
        emit('update:modelValue', editor.getContent())
      })
    },
    skin: 'oxide',
    placeholder: props.placeholder,
    promotion: false,
    branding: false,
    menubar: false,
    statusbar: false,
    border_width: 0,
  })
}

watch(() => props.modelValue, (newValue) => {
  if (editorInstance.value && newValue !== editorInstance.value.getContent()) {
    editorInstance.value.setContent(newValue)
  }
})

onMounted(() => {
  initEditor()
})

onBeforeUnmount(() => {
  if (editorInstance.value) {
    editorInstance.value.destroy()
  }
})
</script>

<template>
  <div class="rich-editor-container border border-slate-200 dark:border-slate-800 rounded-xl bg-white dark:bg-slate-900 focus-within:ring-4 focus-within:ring-blue-500/10 transition-all">
    <textarea :id="uid" ref="editorRef"></textarea>
  </div>
</template>

<style>
.tox-tinymce {
  border: none !important;
  border-radius: 12px !important;
  background: transparent !important;
}
.tox .tox-toolbar-overlord,
.tox .tox-toolbar__primary {
  background-color: #ffffff !important;
  border-bottom: 1px solid #f1f5f9 !important;
  padding: 2px 3px !important;
  display: flex !important;
  flex-wrap: nowrap !important;
  overflow-x: auto !important;
  gap: 2px !important;
}

/* Hide scrollbar for clean look */
.tox .tox-toolbar__primary::-webkit-scrollbar {
  display: none;
}

.tox .tox-toolbar__group {
  padding: 0 3px !important;
  border-right: 2px solid #f1f5f9 !important;
  display: flex !important;
  align-items: center !important;
  gap: 2px !important;
}

.tox .tox-toolbar__group:last-child {
  border-right: none !important;
}

.tox .tox-tbtn {
  height: 28px !important;
  width: 28px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 4px !important;
  margin: 0 1px !important;
  background-color: #ffffff !important;
}

.tox .tox-tbtn:hover {
  background-color: #f8fafc !important;
  border-color: #cbd5e1 !important;
}

.tox .tox-tbtn svg {
  transform: scale(0.8) !important;
}

.tox .tox-tbtn--select {
  width: auto !important;
  min-width: 55px !important;
  padding: 0 4px !important;
}

.tox .tox-tbtn--select span {
  font-size: 11px !important;
  font-weight: 700 !important;
}

.tox-tinymce {
  border: 1px solid #e2e8f0 !important;
  border-radius: 12px !important;
}

/* Dark mode overrides */
.dark .tox .tox-toolbar-overlord,
.dark .tox .tox-toolbar__primary {
  background-color: #1e293b !important;
  border-bottom: 1px solid #334155 !important;
}

.dark .tox .tox-tbtn:hover {
  background-color: #334155 !important;
}
</style>
