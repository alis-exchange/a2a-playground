/**
 * OAuth helpers for adk_request_credential flow.
 * Used when the agent requests user authorization (e.g. Google OAuth).
 */

const ADK_REQUEST_CREDENTIAL = 'adk_request_credential'

export interface PendingAuthCall {
  id: string
  name: string
  args: Record<string, unknown>
  taskId: string
  contextId: string
}

/** Auth config shape from backend (nested structure). */
interface AuthConfigShape {
  exchangedAuthCredential?: { oauth2?: { authUri?: string } }
}

/** True when the call is adk_request_credential with OAuth authConfig. */
export function isAuthRequestCall(obj: unknown): obj is PendingAuthCall {
  if (!obj || typeof obj !== 'object') return false
  const o = obj as Record<string, unknown>
  if (o.name !== ADK_REQUEST_CREDENTIAL) return false
  const args = o.args as Record<string, unknown> | undefined
  const authConfig = args?.authConfig as AuthConfigShape | undefined
  const authUri = authConfig?.exchangedAuthCredential?.oauth2?.authUri
  return typeof authUri === 'string' && authUri.length > 0
}

/** Extract auth URI from a pending auth call. */
export function getAuthUriFromCall(call: PendingAuthCall): string {
  const authConfig = call.args?.authConfig as AuthConfigShape | undefined
  const oauth2 = authConfig?.exchangedAuthCredential?.oauth2 as Record<string, unknown> | undefined
  return (oauth2?.authUri as string) ?? ''
}

/** Extract auth config from a pending auth call. */
export function getAuthConfigFromCall(call: PendingAuthCall): Record<string, unknown> {
  const authConfig = call.args?.authConfig as Record<string, unknown> | undefined
  return authConfig ?? {}
}

/**
 * Open OAuth popup, rewrite redirect_uri in auth URL, listen for postMessage.
 * Resolves with authResponseUrl on success, rejects on popup block.
 */
export function openOAuthPopup(authUri: string, redirectUri: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const url = new URL(authUri)
    url.searchParams.set('redirect_uri', redirectUri)
    const popup = window.open(url.toString(), 'oauthPopup', 'width=600,height=700')

    if (!popup) {
      reject(new Error('Popup blocked'))
      return
    }

    const listener = (event: MessageEvent) => {
      if (event.origin !== window.location.origin) return
      const { authResponseUrl } = event.data
      if (authResponseUrl && typeof authResponseUrl === 'string') {
        window.removeEventListener('message', listener)
        resolve(authResponseUrl)
      }
    }

    window.addEventListener('message', listener)
  })
}

/**
 * Build the response payload for adk_request_credential FunctionResponse.
 * Deep clones authConfig and sets authResponseUri and redirectUri.
 */
export function buildAuthResponsePayload(
  authConfig: Record<string, unknown>,
  authResponseUrl: string,
  redirectUri: string,
): Record<string, unknown> {
  const cfg = JSON.parse(JSON.stringify(authConfig)) as Record<string, unknown>
  let exchanged = cfg.exchangedAuthCredential as Record<string, unknown> | undefined
  if (!exchanged) {
    exchanged = {}
    cfg.exchangedAuthCredential = exchanged
  }
  let oauth2 = exchanged.oauth2 as Record<string, unknown> | undefined
  if (!oauth2) {
    oauth2 = {}
    exchanged.oauth2 = oauth2
  }
  oauth2.authResponseUri = authResponseUrl
  oauth2.redirectUri = redirectUri
  return cfg
}
