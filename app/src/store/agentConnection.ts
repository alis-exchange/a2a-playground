import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const STORE_ID = 'agentConnection'
const STORAGE_KEY = 'a2a-playground-agent-connection'

export type AgentProtocol = 'grpc' | 'jsonrpc'

const DEFAULT_AGENT_URL = 'localhost:8080'
const DEFAULT_PROTOCOL: AgentProtocol = 'grpc'

export const useAgentConnectionStore = defineStore(STORE_ID, () => {
  const agentUrl = ref(DEFAULT_AGENT_URL)
  const protocol = ref<AgentProtocol>(DEFAULT_PROTOCOL)

  const persistEnabled = ref(true)

  const loadFromStorage = () => {
    if (!persistEnabled.value || typeof localStorage === 'undefined') return
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (raw) {
        const parsed = JSON.parse(raw) as { agentUrl?: string; protocol?: AgentProtocol }
        if (typeof parsed.agentUrl === 'string' && parsed.agentUrl) {
          agentUrl.value = parsed.agentUrl
        }
        if (parsed.protocol === 'grpc' || parsed.protocol === 'jsonrpc') {
          protocol.value = parsed.protocol
        }
      }
    } catch {
      // ignore
    }
  }

  const saveToStorage = () => {
    if (!persistEnabled.value || typeof localStorage === 'undefined') return
    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({ agentUrl: agentUrl.value, protocol: protocol.value }),
      )
    } catch {
      // ignore
    }
  }

  watch([agentUrl, protocol], () => {
    saveToStorage()
  })

  loadFromStorage()

  return {
    agentUrl,
    protocol,
    persistEnabled,
    loadFromStorage,
  }
})
