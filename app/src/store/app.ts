import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORE_ID = 'app'

export const useAppStore = defineStore(STORE_ID, () => {
  const secondaryDrawerOpen = ref(true)

  const toggleSecondaryDrawer = () => {
    secondaryDrawerOpen.value = !secondaryDrawerOpen.value
  }

  return {
    secondaryDrawerOpen,
    toggleSecondaryDrawer,
  }
})
