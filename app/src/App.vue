<template>
  <app-layout>
    <router-view />
    <v-snackbar-queue
      v-model="snackbarStore.queue"
      location="top end"
    />
  </app-layout>
</template>

<script setup lang="ts">
  // Vue & Store Imports
  import { onBeforeMount, onErrorCaptured } from 'vue'
  import { useAppStore } from '@/store/app'
  import { useSnackbarStore } from '@/store/snackbar'

  // Component Imports

  // Props

  // Define Stores
  const store = useAppStore()
  const snackbarStore = useSnackbarStore()

  // Define Emits

  // Refs

  // Computed

  // Watchers

  // Functions

  // Lifecycle Hooks
  onBeforeMount(() => {
    // No user service in standalone app
  })

  onErrorCaptured((error) => {
    console.error('Uncaught error:', error)
    snackbarStore.warn('An unexpected error occurred. Please try again later.')
    return false // Prevent the error from being thrown
  })
</script>

<style lang="scss">
  /* Inline code overrides
      allows to use eg. <code> console.log('Hello World') </code>
      with nice styling
  */
  :not(pre) > code {
    border-radius: 4px;
    padding: 3px 6px;
    color: #476582;
  }
  /* Vuetify does not have a border bottom helper classes,
  therfore this is a globally defined class */
  .debug {
    outline: #777777 1px solid;
  }

  .text-ai-gradient {
    background: linear-gradient(135deg, #f20000, #006383 30%, #c5e4ce);
    background-clip: text;
    color: transparent;
  }

  .text-ai-gradient-alternative {
    background: linear-gradient(135deg, #b2f4f7, #5ec6de 75%, #e25c5c);
    background-clip: text;
    color: transparent;
  }

  .text-outlined {
    -webkit-text-stroke: 1px #000;
    color: transparent;
    -webkit-text-fill-color: transparent;
  }

  .text-code {
    font-family: 'Courier Prime', monospace;
    font-weight: 550;
    background: linear-gradient(135deg, #f20000, #006383 90%, #c5e4ce);
    background-clip: text;
    color: transparent;
    position: relative;
  }

  .bg-gradient-code {
    background: linear-gradient(135deg, #006383 5%, #6ba587 50%, #c5e4ce);
    color: white;
  }
  .text-dark-primary {
    color: #003c4f;
  }

  .main-textarea-shadow {
    box-shadow: 0 2px 8px -2px color(from #444746 srgb r g b / 0.16);
    transition: box-shadow 0.1s;
  }
</style>
