<script setup>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'

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

const initEditor = async () => {
  // Wait for tinymce to be available
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
    content_css: [
      'default',
      'https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.css'
    ],
    content_style: 'body { font-family:Inter,Helvetica,Arial,sans-serif; font-size:14px; color: #475569; padding: 10px; }',
    setup: (editor) => {
      editorInstance.value = editor
      
      // Add custom Sigma icon
      editor.ui.registry.addIcon('sigma', '<svg width="24" height="24" viewBox="0 0 24 24"><path d="M18 6H8.83l6 6-6 6H18v2H6v-2l6-6-6-6V4h12v2z"/></svg>');

      // Add custom Math (Sigma) button
      editor.ui.registry.addButton('math', {
        text: '',
        icon: 'sigma',
        tooltip: 'Insert Math (LaTeX)',
        onAction: () => {
          let latex = prompt('Masukkan notasi LaTeX (Tanpa tanda $).\nContoh: x^2 + y^2 = z^2')
          if (latex && window.katex) {
            // Strip $ if user accidentally included them
            latex = latex.replace(/^\$|\$$/g, '')
            try {
              const html = window.katex.renderToString(latex, {
                throwOnError: false,
                displayMode: false
              })
              editor.insertContent(`<span class="math-tex" data-latex="${latex}" contenteditable="false" style="display: inline-block; padding: 0 4px;">${html}</span>&nbsp;`)
            } catch (err) {
              console.error('KaTeX error:', err)
            }
          }
        }
      })

      // Auto-render feature: scan for $...$ and replace
      editor.on('KeyUp', (e) => {
        if (e.key === '$') {
          const content = editor.getContent()
          const regex = /\$([^\$]+)\$/g
          if (regex.test(content)) {
            const newContent = content.replace(regex, (match, latex) => {
              try {
                const html = window.katex.renderToString(latex, {
                  throwOnError: false,
                  displayMode: false
                })
                return `<span class="math-tex" data-latex="${latex}" contenteditable="false" style="display: inline-block; padding: 0 4px;">${html}</span>&nbsp;`
              } catch (err) {
                return match
              }
            })
            if (newContent !== content) {
              editor.setContent(newContent)
            }
          }
        }
      })

      editor.on('change input undo redo', () => {
        emit('update:modelValue', editor.getContent())
      })
    },
    skin: 'oxide',
    content_css: 'default',
    placeholder: props.placeholder,
    promotion: false,
    branding: false,
    menubar: false,
    statusbar: false, // Hide status bar for cleaner look as in the image
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
  flex-wrap: nowrap !important; /* Force single line */
  overflow-x: auto !important; /* Allow scroll if still too wide */
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
  border: 1px solid #e2e8f0 !important; /* Visible border for each icon */
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
  font-size: 11px !important; /* Enlarged font size text */
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
