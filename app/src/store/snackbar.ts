import { defineStore } from 'pinia'
import { v4 as uuidv4 } from 'uuid'
import { ref } from 'vue'

const STORE_ID = 'snackbar'

export interface SnackbarOptions {
  text: string
  color?: string
  timeout?: number
  location?: 'top' | 'bottom' | 'top start' | 'top end' | 'top center' | 'bottom start' | 'bottom end' | 'bottom center'
  variant?: 'text' | 'flat' | 'elevated' | 'tonal' | 'outlined'
  icon?: string
  closable?: boolean
  id?: string
}

export const useSnackbarStore = defineStore(STORE_ID, () => {
  const queue = ref<SnackbarOptions[]>([])

  const show = (options: SnackbarOptions) => {
    queue.value.push({
      id: uuidv4(),
      timeout: 5000,
      location: 'top end',
      variant: 'flat',
      closable: true,
      ...options,
    })
  }

  const error = (text: string, options?: Omit<SnackbarOptions, 'text' | 'color'>) => {
    show({
      text,
      color: 'error',
      icon: 'error',
      ...options,
    })
  }

  const success = (text: string, options?: Omit<SnackbarOptions, 'text' | 'color'>) => {
    show({
      text,
      color: 'success',
      icon: 'check_circle',
      ...options,
    })
  }

  const warn = (text: string, options?: Omit<SnackbarOptions, 'text' | 'color'>) => {
    show({
      text,
      color: 'warning',
      icon: 'warning',
      ...options,
    })
  }

  const info = (text: string, options?: Omit<SnackbarOptions, 'text' | 'color'>) => {
    show({
      text,
      color: 'info',
      icon: 'info',
      ...options,
    })
  }

  const remove = (index: number) => {
    queue.value.splice(index, 1)
  }

  const clear = () => {
    queue.value = []
  }

  return {
    queue,
    show,
    error,
    success,
    warn,
    info,
    remove,
    clear,
  }
})

// Export helper functions for convenience
export const useSnackbar = () => {
  const store = useSnackbarStore()
  return {
    error: store.error,
    success: store.success,
    warn: store.warn,
    info: store.info,
    show: store.show,
  }
}
