<script setup lang="ts">
  // Vue & Store Imports
  import Placeholder from '@tiptap/extension-placeholder'
  import Underline from '@tiptap/extension-underline'
  import StarterKit from '@tiptap/starter-kit'
  import { EditorContent, useEditor } from '@tiptap/vue-3'
  import { BubbleMenu } from '@tiptap/vue-3/menus'
  import { Markdown } from 'tiptap-markdown'
  import type { Ref } from 'vue'
  import { onBeforeUnmount, ref, watch } from 'vue'

  // Component Imports

  // Props
  const props = defineProps<{
    /**
     * Placeholder text to display when the editor is empty.
     */
    placeholder?: string
    /**
     * Disables the editor, making it read-only.
     */
    disabled?: boolean
  }>()

  // Define Stores

  // Define Emits

  // Refs
  /**
   * `v-model` for the editor's content, synced as a Markdown string.
   */
  const content = defineModel<string>()
  /**
   * `v-model` to control which toolbar is displayed.
   * `true` shows the fixed toolbar, `false` shows the floating bubble menu.
   */
  const mountStyling = defineModel<boolean>('mountStyling', { default: false })
  /**
   * Holds the active inline format states (e.g., 'bold', 'italic') for the current selection.
   * Used to toggle the buttons in the toolbar.
   */
  const textFormatState = ref<string[]>([])
  /**
   * Holds the active block-level format state (e.g., 'h1', 'bulletList').
   * Used to toggle the buttons in the toolbar.
   */
  const blockState = ref<string>('')
  /**
   * The core Tiptap editor instance.
   */
  const editor = useEditor({
    content: content.value || '',
    extensions: [
      StarterKit,
      Placeholder.configure({
        placeholder: props.placeholder || 'How can I assist you today?',
      }),
      Underline,
      Markdown.configure({
        html: false, // Opt-out of HTML support
        tightLists: true, // No <p> inside <li>
        linkify: true, // Auto-detect links
        breaks: true, // Render soft line breaks as <br> tags
      }),
    ],
    /**
     * When the editor content changes, emit the new content as Markdown.
     * If the editor is empty, emit an empty string for clean v-model handling.
     */
    onUpdate: ({ editor }) => {
      if (editor.isEmpty) {
        content.value = ''
      } else {
        content.value = (editor.storage as any).markdown.getMarkdown()
      }
    },
    /**
     * On every transaction (e.g., selection change, content change),
     * check the active formats and update the state for the toolbars.
     */
    onTransaction: ({ editor }) => {
      // Update text format state
      const formats: string[] = []
      if (editor.isActive('bold')) formats.push('bold')
      if (editor.isActive('italic')) formats.push('italic')
      if (editor.isActive('underline')) formats.push('underline')
      textFormatState.value = formats

      // Update block state
      if (editor.isActive('heading', { level: 1 })) {
        blockState.value = 'h1'
      } else if (editor.isActive('heading', { level: 2 })) {
        blockState.value = 'h2'
      } else if (editor.isActive('bulletList')) {
        blockState.value = 'bulletList'
      } else if (editor.isActive('orderedList')) {
        blockState.value = 'orderedList'
      } else if (editor.isActive('codeBlock')) {
        blockState.value = 'codeBlock'
      } else {
        blockState.value = ''
      }
    },
    editorProps: {
      attributes: {
        class: 'prose-mirror-editor tiptap-content',
      },
    },
  })

  // Computed

  // Watchers
  /**
   * Watch for external changes to the v-model (e.g., from a parent component)
   * and update the editor's content accordingly.
   */
  watch(
    () => content.value,
    (newVal) => {
      if (editor.value && (editor.value.storage as any).markdown.getMarkdown() !== newVal) {
        editor.value.commands.setContent(newVal || '')
      }
    },
  )

  /**
   * Dynamically update the editor's placeholder if the prop changes.
   */
  watch(
    () => props.placeholder,
    (newPlaceholder) => {
      if (editor.value) {
        const placeholderExtension = editor.value.extensionManager.extensions.find((extension) => extension.name === 'placeholder')
        if (placeholderExtension) {
          placeholderExtension.options.placeholder = newPlaceholder || ''
          editor.value.view.dispatch(editor.value.state.tr)
        }
      }
    },
  )

  /**
   * Dynamically enable or disable the editor's editable state if the prop changes.
   */
  watch(
    () => props.disabled,
    (isDisabled) => {
      if (editor.value && editor.value.isEditable === isDisabled) {
        editor.value.setEditable(!isDisabled)
      }
    },
    { immediate: true },
  )

  // Functions

  // Lifecycle Hooks
  /**
   * Clean up the editor instance when the component is unmounted to prevent memory leaks.
   */
  onBeforeUnmount(() => {
    editor.value?.destroy()
  })
</script>

<template>
  <v-slide-y-reverse-transition>
    <div
      v-if="mountStyling"
      class="d-flex align-center"
    >
      <v-btn-toggle
        :model-value="textFormatState"
        density="comfortable"
        color="primary"
        variant="text"
        multiple
      >
        <v-btn
          variant="text"
          value="bold"
          icon
          @click="editor?.chain().focus().toggleBold().run()"
          size="small"
        >
          <v-icon color="primary">format_bold</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="italic"
          @click="editor?.chain().focus().toggleItalic().run()"
          size="small"
        >
          <v-icon color="primary">format_italic</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="underline"
          @click="editor?.chain().focus().toggleUnderline().run()"
          size="small"
        >
          <v-icon color="primary">format_underlined</v-icon>
        </v-btn>
      </v-btn-toggle>

      <v-divider
        vertical
        class="mx-2"
        inset
      ></v-divider>

      <!-- Block-level elements -->
      <v-btn-toggle
        :model-value="blockState"
        density="comfortable"
        color="primary"
        variant="text"
      >
        <v-btn
          icon
          value="h1"
          variant="text"
          @click="editor?.chain().focus().toggleHeading({ level: 1 }).run()"
          size="small"
        >
          <span class="text-caption font-weight-bold">H1</span>
        </v-btn>
        <v-btn
          icon
          value="h2"
          variant="text"
          @click="editor?.chain().focus().toggleHeading({ level: 2 }).run()"
          size="small"
        >
          <span class="text-caption font-weight-bold">H2</span>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="bulletList"
          @click="editor?.chain().focus().toggleBulletList().run()"
          size="small"
        >
          <v-icon color="primary">format_list_bulleted</v-icon>
        </v-btn>
        <v-btn
          variant="text"
          value="orderedList"
          @click="editor?.chain().focus().toggleOrderedList().run()"
          size="small"
        >
          <v-icon color="primary">format_list_numbered</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="codeBlock"
          @click="editor?.chain().focus().toggleCodeBlock().run()"
          size="small"
        >
          <v-icon color="primary">code</v-icon>
        </v-btn>
      </v-btn-toggle>
    </div>
  </v-slide-y-reverse-transition>
  <BubbleMenu
    v-if="editor"
    :editor="editor"
    :should-show="({ view, from, to }) => view.hasFocus() && from !== to && !mountStyling"
    style="z-index: 2"
  >
    <v-card
      class="pa-1 d-flex overflow-x-auto opacity-1 bg-white"
      rounded="pill"
      color="white"
      elevation="4"
    >
      <!-- Bold, Italic, Underline -->
      <v-btn-toggle
        :model-value="textFormatState"
        color="primary"
        variant="text"
        multiple
        density="compact"
      >
        <v-btn
          icon
          variant="text"
          value="bold"
          @click="editor?.chain().focus().toggleBold().run()"
          size="x-small"
        >
          <v-icon color="primary">format_bold</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="italic"
          @click="editor?.chain().focus().toggleItalic().run()"
          size="x-small"
        >
          <v-icon color="primary">format_italic</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="underline"
          @click="editor?.chain().focus().toggleUnderline().run()"
          size="x-small"
        >
          <v-icon color="primary">format_underlined</v-icon>
        </v-btn>
      </v-btn-toggle>
      <v-divider
        vertical
        class="mx-2"
      ></v-divider>

      <!-- Block-level elements -->
      <v-btn-toggle
        :model-value="blockState"
        color="primary"
        variant="text"
        density="compact"
      >
        <v-btn
          value="h1"
          icon
          variant="text"
          @click="editor?.chain().focus().toggleHeading({ level: 1 }).run()"
          size="x-small"
        >
          <span class="text-caption font-weight-bold">H1</span>
        </v-btn>
        <v-btn
          icon
          value="h2"
          variant="text"
          @click="editor?.chain().focus().toggleHeading({ level: 2 }).run()"
          size="x-small"
        >
          <span class="text-caption font-weight-bold">H2</span>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="bulletList"
          @click="editor?.chain().focus().toggleBulletList().run()"
          size="x-small"
        >
          <v-icon color="primary">format_list_numbered</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="orderedList"
          @click="editor?.chain().focus().toggleOrderedList().run()"
          size="x-small"
        >
          <v-icon color="primary">format_list_numbered</v-icon>
        </v-btn>
        <v-btn
          icon
          variant="text"
          value="codeBlock"
          @click="editor?.chain().focus().toggleCodeBlock().run()"
          size="x-small"
        >
          <v-icon color="primary">code</v-icon>
        </v-btn>
      </v-btn-toggle>
    </v-card>
  </BubbleMenu>
  <EditorContent :editor="editor" />
</template>

<style scoped>
  :deep(.prose-mirror-editor) {
    outline: none;
    width: 100%;
    min-height: 62px;
    padding: 12px;
    border-radius: 0;
    line-height: 1.5rem;
    max-height: 280px;
    overflow-y: auto;
  }

  @media (max-width: 1280px) {
    :deep(.prose-mirror-editor) {
      max-height: 72px;
    }
  }

  :deep(.prose-mirror-editor p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    float: left;
    color: rgb(148, 148, 148);
    pointer-events: none;
    height: 0;
  }
</style>
