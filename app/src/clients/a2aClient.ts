import { createClient } from '@connectrpc/connect'
import type { Interceptor } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { A2AService } from '@local/a2a-js'
import { useAgentHeadersStore } from '@/store/agentHeaders'

// When empty, use same origin (BFF). Set VITE_API_URL for a custom API base.
const baseUrl = import.meta.env.VITE_API_URL ?? ''

const agentHeadersInterceptor: Interceptor = (next) => (req) => {
  const store = useAgentHeadersStore()
  const json = store.getHeadersJson()
  if (json) {
    req.header.set('X-A2A-Agent-Headers', json)
  }
  return next(req)
}

const transport = createConnectTransport({
  baseUrl,
  interceptors: [agentHeadersInterceptor],
})

/** Connect-RPC client for the A2A service. Single local agent—no headers or tenant routing. */
export function createA2AClient() {
  return createClient(A2AService, transport)
}
