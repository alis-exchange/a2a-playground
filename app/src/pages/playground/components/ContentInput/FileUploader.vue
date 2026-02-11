<script setup lang="ts">
  // Vue & Store Imports
  import type { Ref } from 'vue'
  import { ref, watch } from 'vue'
  import { VFileInput } from 'vuetify/components'

  // Component Imports

  // Props
  const props = defineProps<{
    disabled?: boolean
  }>()

  // Define Stores

  // Define Emits
  const emit = defineEmits<{
    (e: 'update:files', files: File[] | undefined): void
    (e: 'removeFile', file: File): void
  }>()

  // Refs
  const uploadedFiles = defineModel<File[] | undefined>()
  const fileToUrl: Ref<Map<string, string>> = ref(new Map())
  const videoDurations: Ref<Map<string, number>> = ref(new Map())
  const hoveredFile: Ref<number | undefined> = ref(undefined)

  // Computed

  // Watchers
  watch(
    () => uploadedFiles.value,
    (files) => {
      if (!files) return
      for (const file of files) {
        if (file.type.startsWith('image/')) {
          const reader = new FileReader()
          reader.onload = (e) => {
            if (e.target?.result) {
              fileToUrl.value.set(file.name, e.target.result as string)
            }
          }
          reader.readAsDataURL(file)
        } else if (file.type.startsWith('video/')) {
          const video = document.createElement('video')
          video.preload = 'metadata'
          video.src = URL.createObjectURL(file)
          video.onloadedmetadata = () => {
            videoDurations.value.set(file.name, video.duration)
            video.currentTime = 0
          }
          video.onseeked = () => {
            const canvas = document.createElement('canvas')
            canvas.width = video.videoWidth
            canvas.height = video.videoHeight
            const ctx = canvas.getContext('2d')
            if (ctx) {
              ctx.drawImage(video, 0, 0, canvas.width, canvas.height)
              fileToUrl.value.set(file.name, canvas.toDataURL())
            }
            URL.revokeObjectURL(video.src)
          }
        }
      }
    },
    { deep: true, immediate: true },
  )

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
    emit('removeFile', fileToRemove)
  }

  function formatDuration(seconds: number): string {
    if (isNaN(seconds) || seconds < 0) {
      return ''
    }
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = Math.floor(seconds % 60)

    if (h > 0) {
      return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
    } else {
      return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
    }
  }

  function calcWidth(f: File) {
    if (f.type.includes('image') || f.type.includes('video')) {
      return '80px'
    }
    return '208px'
  }

  // Lifecycle Hooks
</script>

<template>
  <div
    v-if="uploadedFiles && uploadedFiles.length > 0"
    class="d-flex flex-wrap ga-2 pa-4"
  >
    <v-card
      v-for="(f, i) in uploadedFiles"
      :key="i"
      variant="flat"
      color="surfaceContainer"
      class="w-100"
      height="80px"
      :max-width="calcWidth(f)"
      @mouseover="hoveredFile = i"
      @mouseleave="hoveredFile = undefined"
    >
      <v-btn
        v-show="hoveredFile === i"
        class="position-absolute right-0 ma-2"
        size="x-small"
        variant="flat"
        color="white"
        elevation="1"
        icon
        @click="removeFile(f)"
      >
        <v-icon color="grey-darken-3"> close </v-icon>
      </v-btn>
      <v-img
        v-if="f.type.includes('image/') || f.type.includes('video/')"
        :src="fileToUrl.get(f.name)"
        cover
        class="h-100"
      >
        <v-btn
          v-show="hoveredFile === i"
          class="position-absolute right-0 ma-2"
          size="x-small"
          variant="flat"
          color="white"
          icon
          @click="removeFile(f)"
        >
          <v-icon color="grey-darken-3"> close </v-icon>
        </v-btn>
        <div
          v-if="f.type.includes('video/') && videoDurations.get(f.name)"
          class="d-flex align-end h-100"
        >
          <span class="font-weight-bold text-body-2 text-grey-darken-2 bg-white w-100 px-3 py-2 opacity-80">
            {{ formatDuration(videoDurations.get(f.name)!) }}
          </span>
        </div>
      </v-img>
      <v-card-text v-else>
        <div class="text-truncate text-body-2 text-grey-darken-3 font-weight-medium">
          {{ f.name.split('.')[0] }}
        </div>
        <div class="text-caption text-truncate text-grey-darken-1 mt-1">
          {{ f.name.split('.')[1]?.toUpperCase() }}
        </div>
      </v-card-text>
    </v-card>
  </div>
</template>

<style scoped></style>
