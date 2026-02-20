<script setup lang="ts">
  import { onMounted, ref } from 'vue'

  const hasCode = ref(false)

  onMounted(() => {
    const params = new URLSearchParams(window.location.search)
    hasCode.value = params.has('code')
    if (hasCode.value) {
      const authResponseUrl = window.location.href
      window.opener?.postMessage({ authResponseUrl }, window.location.origin)
      window.close()
    }
  })
</script>

<template>
  <div class="oauth-callback d-flex flex-column align-center justify-center pa-8">
    <v-progress-circular
      indeterminate
      color="primary"
      size="48"
      width="4"
      class="mb-4"
    />
    <div class="text-body-1 text-medium-emphasis">
      Completing authorization...
    </div>
    <div
      v-if="!hasCode"
      class="text-caption text-disabled mt-2"
    >
      This window should close automatically. If it doesn't, you can close it manually.
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .oauth-callback {
    min-height: 200px;
  }
</style>
