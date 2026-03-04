import { createClient } from '@connectrpc/connect'
import type { Interceptor } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { A2AService } from '@local/a2a-js'
import { useAgentConnectionStore } from '@/store/agentConnection'
import { useAgentHeadersStore } from '@/store/agentHeaders'
import { useAgentOAuthStore } from '@/store/agentOAuth'

// When empty, use same origin (BFF). Set VITE_API_URL for a custom API base.
const baseUrl = import.meta.env.VITE_API_URL ?? ''

const agentConfigInterceptor: Interceptor = (next) => (req) => {
  const connectionStore = useAgentConnectionStore()
  if (connectionStore.agentUrl) {
    req.header.set('X-A2A-Agent-URL', connectionStore.agentUrl)
  }
  req.header.set('X-A2A-Agent-Protocol', connectionStore.protocol)
  return next(req)
}

const agentOAuthBearerInterceptor: Interceptor = (next) => async (req) => {
  const oauthStore = useAgentOAuthStore()
  let token = oauthStore.getValidAccessToken()
  if (!token && oauthStore.tokens?.refreshToken) {
    await oauthStore.refresh()
    token = oauthStore.getValidAccessToken()
  }
  if (token) {
    req.header.set('Authorization', `Bearer ${token}`)
  }
  return next(req)
}

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
  interceptors: [agentConfigInterceptor, agentOAuthBearerInterceptor, agentHeadersInterceptor],
})

/** Connect-RPC client for the A2A service. Single local agent—no headers or tenant routing. */
export function createA2AClient() {
  return createClient(A2AService, transport)
}
