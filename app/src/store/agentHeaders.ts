import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const STORE_ID = 'agentHeaders'
const STORAGE_KEY = 'a2a-playground-agent-headers'

export interface HeaderEntry {
  key: string
  value: string
}

const PRESET_KEYS = ['Authorization', 'X-API-Key', 'X-Tenant-ID'] as const

export const useAgentHeadersStore = defineStore(STORE_ID, () => {
  const headers = ref<HeaderEntry[]>([])

  const addHeader = (entry?: Partial<HeaderEntry>) => {
    const newEntry: HeaderEntry = {
      key: entry?.key ?? '',
      value: entry?.value ?? '',
    }
    headers.value = [...headers.value, newEntry]
  }

  const removeHeader = (index: number) => {
    headers.value = headers.value.filter((_, i) => i !== index)
  }

  const updateHeader = (index: number, updates: Partial<HeaderEntry>) => {
    const next = [...headers.value]
    if (index >= 0 && index < next.length) {
      const current = next[index]!
      next[index] = {
        key: updates.key ?? current.key,
        value: updates.value ?? current.value,
      }
      headers.value = next
    }
  }

  /** Add a preset row with key pre-filled; user enters value. */
  const addPreset = (key: string) => {
    addHeader({ key, value: '' })
  }

  /** Get headers as a plain object for sending. Filters out empty keys. */
  const getHeadersObject = (): Record<string, string> => {
    const obj: Record<string, string> = {}
    for (const { key, value } of headers.value) {
      const k = key?.trim()
      if (k) obj[k] = value ?? ''
    }
    return obj
  }

  /** JSON string for X-A2A-Agent-Headers. Returns null if empty. */
  const getHeadersJson = (): string | null => {
    const obj = getHeadersObject()
    if (Object.keys(obj).length === 0) return null
    return JSON.stringify(obj)
  }

  // Optional: persist to localStorage
  const persistEnabled = ref(true)

  const loadFromStorage = () => {
    if (!persistEnabled.value || typeof localStorage === 'undefined') return
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (raw) {
        const parsed = JSON.parse(raw) as HeaderEntry[]
        if (Array.isArray(parsed)) {
          headers.value = parsed.filter((e) => e && typeof e.key === 'string')
        }
      }
    } catch {
      // ignore
    }
  }

  const saveToStorage = () => {
    if (!persistEnabled.value || typeof localStorage === 'undefined') return
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(headers.value))
    } catch {
      // ignore
    }
  }

  watch(
    headers,
    (val) => {
      saveToStorage()
    },
    { deep: true }
  )

  loadFromStorage()

  return {
    headers,
    presetKeys: PRESET_KEYS,
    addHeader,
    removeHeader,
    updateHeader,
    addPreset,
    getHeadersObject,
    getHeadersJson,
    persistEnabled,
    loadFromStorage,
  }
})
