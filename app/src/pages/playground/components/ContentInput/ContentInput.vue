<template>
  <v-card-outlined
    variant="flat"
    rounded="xl"
    class="pt-2 w-100"
    :class="dropZoneClass"
    @dragenter="handleDragEnter"
    @dragleave="handleDragLeave"
    @dragover.prevent="handleDragOver"
    @drop.prevent="handleDrop"
    @paste="handlePaste"
  >
    <div
      v-if="isDragging"
      class="drop-zone-overlay d-flex flex-column align-center justify-center position-absolute w-100 h-100 top-0 left-0 text-primary rounded-xl"
    >
      <div class="mb-2">
        <v-icon>attach_file</v-icon>
      </div>
      Drop files here.
    </div>

    <div
      v-if="uploadedFiles && uploadedFiles.length > 0"
      class="px-4 pb-2"
    >
      <FileUploader
        v-model="uploadedFiles"
        :disabled="disabled"
        @remove-file="removeFile"
      />
    </div>
    <div
      v-if="!audioIsRecording"
      class="editor-wrapper px-4"
    >
      <RichTextInput
        v-model="textInput"
        v-model:mount-styling="mountTextStyling"
        :placeholder="placeholder"
        :disabled="disabled || !!audioIsRecording"
      />
    </div>
    <v-card-actions class="d-flex justify-end pt-0">
      <template v-if="!audioIsRecording">
        <v-btn
          variant="text"
          rounded="xl"
          icon
          @click="fileInputComponent?.click()"
          :disabled="disabled"
        >
          <v-icon>attach_file</v-icon>
          <v-tooltip activator="parent"> Upload a file </v-tooltip>
        </v-btn>
        <v-file-input
          v-show="false"
          v-model="uploadedFiles"
          multiple
          :disabled="disabled"
          ref="fileInputComponent"
        />
      </template>
      <v-spacer />
      <AudioRecorder
        v-if="textInput.trim().length === 0"
        v-model:is-recording="audioIsRecording"
        v-model:audio-blob="audioBlob"
        :disabled="disabled"
      />
      <v-btn
        v-else-if="!audioIsRecording"
        color="primary"
        variant="flat"
        rounded="xl"
        icon
        :disabled="!hasContent"
        @click="submit"
      >
        <v-icon>send</v-icon>
        <v-tooltip activator="parent"> Send message </v-tooltip>
      </v-btn>
    </v-card-actions>
  </v-card-outlined>
</template>

<script setup lang="ts">
  // Vue & Store Imports
  import { useMagicKeys } from '@vueuse/core'
  import type { Ref } from 'vue'
  import { computed, ref, watchEffect } from 'vue'
  import { VFileInput } from 'vuetify/components'
  import { AudioMessagePart, FileMessagePart, MessagePart, TextMessagePart } from './types'

  // Component Imports
  import AudioRecorder from './AudioRecorder.vue'
  import FileUploader from './FileUploader.vue'
  import RichTextInput from './RichTextInput.vue'

  // Props
  const props = defineProps<{
    placeholder?: string
    disabled?: boolean
  }>()

  // Define Stores

  // Define Emits
  interface Emits {
    (e: 'submit', parts: MessagePart[]): void
  }

  const emit = defineEmits<Emits>()

  // Refs
  const textInput: Ref<string> = ref('')
  const uploadedFiles: Ref<File[] | undefined> = ref(undefined)
  const audioBlob: Ref<Blob | undefined> = ref(undefined)
  const audioIsRecording: Ref<boolean> = ref(false)
  const fileInputComponent: Ref<VFileInput | null> = ref(null)
  const dragCounter: Ref<number> = ref(0)

  const mountTextStyling: Ref<boolean> = ref(false)

  watchEffect(() => {
    if (textInput.value.trim()) {
      mountTextStyling.value = true
    } else {
      mountTextStyling.value = false
    }
  })

  // Computed
  const isDragging = computed(() => dragCounter.value > 0)

  const dropZoneClass = computed(() => {
    return {
      'bg-white': !isDragging.value,
      'bg-primary': isDragging.value,
    }
  })

  const hasContent = computed(() => {
    return (textInput.value && textInput.value.trim().length > 0) || (uploadedFiles.value && uploadedFiles.value.length > 0) || !!audioBlob.value
  })

  // Watchers
  const { meta_enter, ctrl_enter, enter } = useMagicKeys()
  watchEffect(async () => {
    // Enter key for text-only submissions (no files, no audio)
    if (enter?.value && !meta_enter?.value && !ctrl_enter?.value) {
      if (!audioIsRecording.value && !uploadedFiles.value && !audioBlob.value && textInput.value?.trim()) {
        await submit()
      }
    }
    // Ctrl+Enter / Cmd+Enter for submissions with files/audio
    if (meta_enter?.value || ctrl_enter?.value) {
      if (!audioIsRecording.value && hasContent.value) {
        await submit()
      }
    }
  })

  // Functions
  function removeFile(fileToRemove: File) {
    if (!uploadedFiles.value) return
    const index = uploadedFiles.value.findIndex((f) => f.name === fileToRemove.name && f.size === fileToRemove.size)
    if (index > -1) {
      uploadedFiles.value.splice(index, 1)
    }
    if (uploadedFiles.value.length === 0) {
      uploadedFiles.value = undefined
    }
  }

  function handleDragEnter(event: DragEvent) {
    event.preventDefault()
    dragCounter.value++
  }

  function handleDragLeave(event: DragEvent) {
    event.preventDefault()
    dragCounter.value--
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault()
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault()
    dragCounter.value = 0
    if (event.dataTransfer && event.dataTransfer.files.length) {
      const files = Array.from(event.dataTransfer.files)
      const validFiles: File[] = []

      files.forEach((file) => {
        const isDuplicate = uploadedFiles.value?.some((existingFile) => existingFile.name === file.name && existingFile.size === file.size)
        if (!isDuplicate) {
          validFiles.push(file)
        }
      })

      if (validFiles.length > 0) {
        if (!uploadedFiles.value) {
          uploadedFiles.value = []
        }
        uploadedFiles.value.push(...validFiles)
      }
    }
  }

  function handlePaste(event: ClipboardEvent) {
    if (!event.clipboardData) return

    const items = event.clipboardData.items
    const validFiles: File[] = []
    let mediaPasted = false

    for (let i = 0; i < items.length; i++) {
      const item = items[i]
      if (item?.kind === 'file') {
        const file = item.getAsFile()
        if (file) {
          mediaPasted = true
          const randomString = Math.random().toString(36).substring(2, 10)
          const newFile = new File([file], `${randomString}_${file.name}`, {
            type: file.type,
            lastModified: file.lastModified,
          })
          validFiles.push(newFile)
        }
      }
    }

    if (mediaPasted) {
      event.preventDefault()
      if (validFiles.length > 0) {
        if (!uploadedFiles.value) {
          uploadedFiles.value = []
        }
        uploadedFiles.value.push(...validFiles)
      }
    }
  }

  async function submit() {
    const parts: MessagePart[] = []

    // Add text part if text exists
    if (textInput.value && textInput.value.trim().length > 0) {
      parts.push(new TextMessagePart(textInput.value.trim()))
    }

    // Add file parts
    if (uploadedFiles.value && uploadedFiles.value.length > 0) {
      for (const file of uploadedFiles.value) {
        parts.push(new FileMessagePart(file))
      }
    }

    // Add audio part
    if (audioBlob.value) {
      parts.push(new AudioMessagePart(audioBlob.value))
    }

    if (parts.length > 0) {
      emit('submit', parts)
      // Reset all inputs
      textInput.value = ''
      uploadedFiles.value = undefined
      audioBlob.value = undefined
      audioIsRecording.value = false
    }
  }

  // Lifecycle Hooks
</script>

<style lang="scss" scoped>
  .drop-zone-overlay {
    background-color: rgba(255, 255, 255, 0.9);
    border: 2px dashed #006383;
    z-index: 1;
  }
</style>
