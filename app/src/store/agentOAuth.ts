import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORE_ID = 'agentOAuth'

export interface OAuthConfig {
  clientId: string
  clientSecret: string
  authorizationUrl: string
  tokenUrl: string
  scope: string
}

export interface OAuthTokens {
  accessToken: string
  refreshToken: string
  expiresAt: number
}

export interface OAuthPresetSummary {
  nickname: string
}

/** Response shape from GET /api/oauth-presets/{nickname} (snake_case). */
interface OAuthPresetResponse {
  nickname: string
  client_id: string
  client_secret: string
  authorization_url: string
  token_url: string
  scope: string
}

function getApiBase(): string {
  return import.meta.env.VITE_API_URL ?? ''
}

export const useAgentOAuthStore = defineStore(STORE_ID, () => {
  const config = ref<OAuthConfig>({
    clientId: '',
    clientSecret: '',
    authorizationUrl: '',
    tokenUrl: '',
    scope: '',
  })

  const tokens = ref<OAuthTokens | null>(null)

  const setConfig = (c: Partial<OAuthConfig>) => {
    config.value = { ...config.value, ...c }
  }

  const setTokens = (t: OAuthTokens | null) => {
    tokens.value = t
  }

  const listPresets = async (): Promise<OAuthPresetSummary[]> => {
    const baseUrl = getApiBase()
    const res = await fetch(`${baseUrl}/api/oauth-presets`)
    if (!res.ok) throw new Error('Failed to list presets')
    const data = (await res.json()) as { presets: OAuthPresetSummary[] }
    return data.presets ?? []
  }

  const loadPreset = async (nickname: string): Promise<void> => {
    const baseUrl = getApiBase()
    const res = await fetch(`${baseUrl}/api/oauth-presets/${encodeURIComponent(nickname)}`)
    if (!res.ok) throw new Error('Failed to load preset')
    const p = (await res.json()) as OAuthPresetResponse
    setConfig({
      clientId: p.client_id ?? '',
      clientSecret: p.client_secret ?? '',
      authorizationUrl: p.authorization_url ?? '',
      tokenUrl: p.token_url ?? '',
      scope: p.scope ?? '',
    })
  }

  const savePreset = async (nickname: string, overwrite: boolean): Promise<'ok' | 'exists'> => {
    const baseUrl = getApiBase()
    const c = config.value
    const res = await fetch(`${baseUrl}/api/oauth-presets`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        nickname,
        client_id: c.clientId,
        client_secret: c.clientSecret,
        authorization_url: c.authorizationUrl,
        token_url: c.tokenUrl,
        scope: c.scope,
        overwrite,
      }),
    })
    if (res.status === 409) return 'exists'
    if (!res.ok) throw new Error('Failed to save preset')
    return 'ok'
  }

  const deletePreset = async (nickname: string): Promise<void> => {
    const baseUrl = getApiBase()
    const res = await fetch(`${baseUrl}/api/oauth-presets/${encodeURIComponent(nickname)}`, {
      method: 'DELETE',
    })
    if (!res.ok) throw new Error('Failed to delete preset')
  }

  const authorize = async () => {
    const { clientId, clientSecret, authorizationUrl, tokenUrl, scope } = config.value
    if (!authorizationUrl || !tokenUrl) return
    const baseUrl = getApiBase()
    const redirectOrigin = typeof window !== 'undefined' ? window.location.origin : ''
    const res = await fetch(`${baseUrl}/auth/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        client_id: clientId,
        client_secret: clientSecret,
        authorization_url: authorizationUrl,
        token_url: tokenUrl,
        scope,
        redirect_origin: redirectOrigin,
      }),
    })
    if (!res.ok) throw new Error('Failed to start auth')
    const data = (await res.json()) as { auth_url: string }
    if (!data.auth_url) return

    const popup = window.open(data.auth_url, 'oauthPopup', 'width=600,height=700')
    if (!popup) {
      throw new Error('Popup blocked. Please allow popups for this site and try again.')
    }

    const TIMEOUT_MS = 10 * 60 * 1000 // 10 minutes

    return new Promise<void>((resolve, reject) => {
      let resolved = false
      let timeoutId: ReturnType<typeof setTimeout> | null = null

      const cleanup = () => {
        window.removeEventListener('message', listener)
        if (timeoutId != null) {
          clearTimeout(timeoutId)
          timeoutId = null
        }
      }

      const listener = (event: MessageEvent) => {
        if (event.origin !== window.location.origin) return
        const payload = event.data as { type?: string; accessToken?: string; refreshToken?: string; expiresAt?: number }
        if (payload?.type === 'auth_callback' && payload.accessToken) {
          if (resolved) return
          resolved = true
          cleanup()
          const tokenPayload = {
            accessToken: payload.accessToken,
            refreshToken: payload.refreshToken ?? '',
            expiresAt: payload.expiresAt ?? Date.now() + 3600 * 1000,
          }
          setTokens(tokenPayload)
          // TODO: remove temporary debug log once OAuth flow is verified
          console.log('[OAuth] Flow completed successfully', { hasAccessToken: true, expiresAt: tokenPayload.expiresAt })
          resolve()
        }
      }
      window.addEventListener('message', listener)

      timeoutId = setTimeout(() => {
        cleanup()
        if (!resolved) {
          resolved = true
          reject(new Error('Authorization timed out'))
        }
      }, TIMEOUT_MS)
    })
  }

  const refresh = async (): Promise<boolean> => {
    const t = tokens.value
    if (!t?.refreshToken) return false
    const { clientId, clientSecret, tokenUrl } = config.value
    if (!tokenUrl) return false
    const baseUrl = import.meta.env.VITE_API_URL ?? ''
    const res = await fetch(`${baseUrl}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        refresh_token: t.refreshToken,
        client_id: clientId,
        client_secret: clientSecret,
        token_url: tokenUrl,
      }),
    })
    if (!res.ok) return false
    const data = (await res.json()) as { access_token: string; refresh_token?: string; expires_in?: number }
    const expiresIn = data.expires_in ?? 3600
    setTokens({
      accessToken: data.access_token,
      refreshToken: data.refresh_token ?? t.refreshToken,
      expiresAt: Date.now() + expiresIn * 1000,
    })
    return true
  }

  const signOut = () => {
    tokens.value = null
  }

  const getValidAccessToken = (): string | null => {
    const t = tokens.value
    if (!t?.accessToken) return null
    if (t.expiresAt && t.expiresAt <= Date.now()) {
      return null
    }
    return t.accessToken
  }

  return {
    config,
    tokens,
    setConfig,
    setTokens,
    listPresets,
    loadPreset,
    savePreset,
    deletePreset,
    authorize,
    refresh,
    signOut,
    getValidAccessToken,
  }
})
