<template>
  <div class="file-part">
    <v-img
      v-if="isImage && hasUri"
      :src="fileUri"
      width="300"
      height="auto"
      cover
      class="file-image"
    />
    <v-img
      v-else-if="isImage && hasBytes"
      :src="imageDataUrl"
      width="300"
      height="auto"
      cover
      class="file-image"
    />
    <v-card
      v-else
      variant="outlined"
      class="file-info"
    >
      <v-card-text>
        <div class="d-flex align-center">
          <v-icon class="mr-3"> description </v-icon>
          <div class="flex-grow-1">
            <div class="font-weight-bold">
              {{ fileName }}
            </div>
            <div
              v-if="mimeType"
              class="text-caption text-grey"
            >
              {{ mimeType }}
            </div>
          </div>
          <v-btn
            v-if="hasUri"
            icon
            variant="text"
            size="small"
            @click="downloadFile"
          >
            <v-icon>download</v-icon>
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  // Vue & Store Imports
  import type { Part } from '@local/a2a-js'
  import type { Ref } from 'vue'
  import { computed } from 'vue'

  // Component Imports

  // Props
  interface Props {
    part: Part
  }

  const props = defineProps<Props>()

  // Define Stores

  // Define Emits

  // Refs

  // Computed
  const fileUri = computed(() => {
    const c = props.part.content
    return c?.case === 'url' ? c.value : undefined
  })

  const fileBytes = computed(() => {
    const c = props.part.content
    return c?.case === 'raw' ? c.value : undefined
  })

  const mimeType = computed(() => props.part.mediaType ?? undefined)

  const fileName = computed(() => props.part.filename ?? undefined)

  const hasUri = computed(() => !!fileUri.value)
  const hasBytes = computed(() => !!fileBytes.value)

  const isImage = computed(() => {
    const mime = mimeType.value
    return mime ? mime.startsWith('image/') : false
  })

  const imageDataUrl = computed(() => {
    // If we have a URI, use it directly (handled in template)
    if (hasUri.value && fileUri.value) {
      return fileUri.value
    }

    // Otherwise, convert bytes to data URL
    if (!hasBytes.value) return ''
    const bytes = fileBytes.value
    if (!bytes) return ''

    const mime = mimeType.value || 'image/png'

    // Handle string format (already base64) - raw content is typically Uint8Array
    if (typeof bytes === 'string') {
      if ((bytes as string).startsWith('data:')) {
        return bytes as string
      }
      return `data:${mime};base64,${bytes as string}`
    }

    // Handle Uint8Array or number array
    let byteArray: Uint8Array | number[]
    if (bytes instanceof Uint8Array) {
      byteArray = bytes
    } else if (Array.isArray(bytes)) {
      byteArray = bytes
    } else {
      return ''
    }

    // Check if the bytes are already base64-encoded (ASCII characters)
    // Base64 only contains: A-Z, a-z, 0-9, +, /, =
    // Raw binary image data (like JPEG) starts with bytes > 127
    // If first bytes are all printable ASCII (32-126), it's likely already base64
    const isLikelyBase64 = byteArray.length > 0 && byteArray.slice(0, 20).every((b) => b >= 32 && b <= 126)

    if (isLikelyBase64) {
      // Bytes are already base64-encoded as ASCII - just convert to string
      const chunkSize = 8192
      let base64String = ''
      for (let i = 0; i < byteArray.length; i += chunkSize) {
        const chunk = byteArray.slice(i, i + chunkSize)
        base64String += String.fromCharCode.apply(null, Array.from(chunk))
      }
      return `data:${mime};base64,${base64String}`
    }

    // Raw binary bytes - need to convert to base64
    const chunkSize = 8192
    let binaryString = ''
    for (let i = 0; i < byteArray.length; i += chunkSize) {
      const chunk = byteArray.slice(i, i + chunkSize)
      binaryString += String.fromCharCode.apply(null, Array.from(chunk))
    }

    const base64 = btoa(binaryString)
    return `data:${mime};base64,${base64}`
  })

  // Watchers

  // Functions
  const downloadFile = () => {
    if (hasUri.value && fileUri.value) {
      window.open(fileUri.value, '_blank')
    }
  }

  // Lifecycle Hooks
</script>

<style lang="scss" scoped>
  .file-part {
    margin-top: 4px;
  }

  .file-image {
    border-radius: 16px;
    border: 1px solid rgba(0, 0, 0, 0.1);
  }

  .file-info {
    border-radius: 16px;
  }
</style>
