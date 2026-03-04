import type { ConversationMessage } from '@/pages/playground/store/messages'
import { defineStore } from 'pinia'
import { ref, type Ref } from 'vue'

const STORE_ID = 'agentPlayground'

// Drawer mode: message inspection only (no agent details)
export type DetailsDrawerMode = 'message'

export const useAgentPlaygroundStore = defineStore(STORE_ID, () => {
  const sidebarOpen: Ref<boolean> = ref(true)
  const detailsDrawerOpen: Ref<boolean> = ref(false)
  const detailsDrawerMode: Ref<DetailsDrawerMode> = ref('message')
  const detailsDrawerMessage: Ref<ConversationMessage | null> = ref(null)

  const setSidebarOpen = (open: boolean) => {
    sidebarOpen.value = open
  }

  const toggleSidebar = () => {
    sidebarOpen.value = !sidebarOpen.value
  }

  const toggleDetailsDrawer = () => {
    detailsDrawerOpen.value = !detailsDrawerOpen.value
  }

  const openDetailsDrawer = () => {
    detailsDrawerOpen.value = true
  }

  const closeDetailsDrawer = () => {
    detailsDrawerOpen.value = false
  }

  const setDetailsDrawerMessage = (message: ConversationMessage | null) => {
    detailsDrawerMessage.value = message
  }

  const openDetailsDrawerForMessage = (message: ConversationMessage) => {
    detailsDrawerMessage.value = message
    detailsDrawerMode.value = 'message'
    detailsDrawerOpen.value = true
  }

  const resetDrawerState = () => {
    detailsDrawerOpen.value = false
    detailsDrawerMode.value = 'message'
    detailsDrawerMessage.value = null
  }

  return {
    sidebarOpen,
    setSidebarOpen,
    toggleSidebar,
    detailsDrawerOpen,
    detailsDrawerMode,
    detailsDrawerMessage,
    toggleDetailsDrawer,
    openDetailsDrawer,
    closeDetailsDrawer,
    setDetailsDrawerMessage,
    openDetailsDrawerForMessage,
    resetDrawerState,
  }
})
