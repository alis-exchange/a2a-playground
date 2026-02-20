<template>
  <div class="pending-function-response-block px-4 py-3 mt-2 mb-2">
    <v-card
      variant="outlined"
      class="pa-3"
    >
      <div class="text-body-2 font-weight-medium mb-2">
        Function call: {{ payload.name }}
      </div>
      <div
        v-if="argsSummary"
        class="text-caption text-grey mb-2"
      >
        {{ argsSummary }}
      </div>
      <div
        v-if="hint"
        class="text-caption mb-2"
      >
        {{ hint }}
      </div>

      <!-- Auth request: Authorize button primary, JSON as expandable fallback -->
      <template v-if="isAuthRequest">
        <div class="d-flex gap-2 mb-2">
          <v-btn
            size="small"
            color="primary"
            :disabled="sending || props.sending || authorizing"
            :loading="authorizing"
            @click="onAuthorize"
          >
            {{ authorizing ? 'Authorizing...' : 'Authorize' }}
          </v-btn>
        </div>
        <v-expansion-panels class="mt-2">
          <v-expansion-panel>
            <v-expansion-panel-title>Manual response (fallback)</v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-textarea
                v-model="responseText"
                label="Response (strict JSON object)"
                variant="outlined"
                density="compact"
                rows="3"
                hide-details
                class="mb-2"
                :error-messages="jsonError"
              />
              <v-btn
                size="small"
                :disabled="sending || props.sending || !isValidJson"
                @click="sendCustomResponse"
              >
                {{ sending || props.sending ? 'Sending...' : 'Send response' }}
              </v-btn>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </template>

      <!-- Confirmation-style shortcut (non-auth) -->
      <div
        v-else-if="showConfirmationButtons"
        class="d-flex gap-2 mb-2"
      >
        <v-btn
          size="small"
          color="primary"
          :disabled="sending || props.sending"
          @click="sendConfirmed(true)"
        >
          Approve
        </v-btn>
        <v-btn
          size="small"
          variant="outlined"
          :disabled="sending || props.sending"
          @click="sendConfirmed(false)"
        >
          Reject
        </v-btn>
      </div>

      <!-- Generic JSON response (non-auth) -->
      <div
        v-else
        class="mt-2"
      >
        <v-textarea
          v-model="responseText"
          label="Response (strict JSON object)"
          variant="outlined"
          density="compact"
          rows="3"
          hide-details
          class="mb-2"
          :error-messages="jsonError"
        />
        <v-btn
          size="small"
          :disabled="sending || props.sending || !isValidJson"
          @click="sendCustomResponse"
        >
          {{ sending || props.sending ? 'Sending...' : 'Send response' }}
        </v-btn>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { useSnackbarStore } from '@/store/snackbar'
  import {
    buildAuthResponsePayload,
    getAuthConfigFromCall,
    getAuthUriFromCall,
    isAuthRequestCall,
    openOAuthPopup,
  } from '@/utils/oauth'

  interface PendingCall {
    id: string
    name: string
    args: Record<string, unknown>
    taskId: string
    contextId: string
  }

  const props = withDefaults(
    defineProps<{
      payload: PendingCall
      sending?: boolean
    }>(),
    { sending: false },
  )

  const emit = defineEmits<{
    (e: 'send', response: Record<string, unknown>): void
  }>()

  const snackbarStore = useSnackbarStore()
  const sending = ref(false)
  const authorizing = ref(false)
  const responseText = ref('{}')
  const jsonError = ref('')

  const isAuthRequest = computed(() => isAuthRequestCall(props.payload))

  const argsSummary = computed(() => {
    const a = props.payload.args
    if (!a || typeof a !== 'object') return ''
    try {
      const entries = Object.entries(a)
        .filter(([, v]) => v !== undefined && v !== null)
        .map(([k, v]) => `${k}: ${typeof v === 'object' ? JSON.stringify(v) : String(v)}`)
      return entries.length ? entries.join(', ') : ''
    } catch {
      return ''
    }
  })

  /** Hint from args (e.g. toolConfirmation.Hint or args.hint). */
  const hint = computed(() => {
    const a = props.payload.args as Record<string, unknown>
    if (!a) return ''
    const tc = a.toolConfirmation as Record<string, unknown> | undefined
    if (tc && typeof tc.Hint === 'string') return tc.Hint
    if (typeof a.hint === 'string') return a.hint
    return ''
  })

  /** Show Approve/Reject when args look like a confirmation (e.g. toolConfirmation). */
  const showConfirmationButtons = computed(() => {
    const a = props.payload.args as Record<string, unknown>
    if (!a) return false
    return 'toolConfirmation' in a || ('Confirmed' in a && Object.keys(a).length <= 2)
  })

  function validateAndParse(): { ok: true; value: Record<string, unknown> } | { ok: false; error: string } {
    const t = responseText.value.trim()
    if (!t) return { ok: false, error: 'Enter a JSON object' }
    try {
      const parsed = JSON.parse(t)
      if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
        return { ok: false, error: 'Must be a JSON object (not array or primitive)' }
      }
      return { ok: true, value: parsed as Record<string, unknown> }
    } catch {
      return { ok: false, error: 'Invalid JSON' }
    }
  }

  const isValidJson = computed(() => validateAndParse().ok)

  watch(
    () => responseText.value,
    () => {
      const r = validateAndParse()
      jsonError.value = r.ok ? '' : r.error
    },
  )

  async function onAuthorize() {
    if (!isAuthRequest.value) return
    authorizing.value = true
    try {
      const authUri = getAuthUriFromCall(props.payload)
      const authConfig = getAuthConfigFromCall(props.payload)
      const redirectUri = `${window.location.origin}/oauth-callback`
      const authResponseUrl = await openOAuthPopup(authUri, redirectUri)
      const payload = buildAuthResponsePayload(authConfig, authResponseUrl, redirectUri)
      emit('send', payload)
    } catch (err) {
      snackbarStore.error('Popup blocked. Use manual response fallback.')
    } finally {
      authorizing.value = false
    }
  }

  async function sendConfirmed(confirmed: boolean) {
    sending.value = true
    try {
      emit('send', { Confirmed: confirmed })
    } finally {
      sending.value = false
    }
  }

  async function sendCustomResponse() {
    const r = validateAndParse()
    if (!r.ok) {
      jsonError.value = r.error
      return
    }
    sending.value = true
    try {
      emit('send', r.value)
    } finally {
      sending.value = false
    }
  }

  defineExpose({ sending })
</script>

<style lang="scss" scoped>
  .pending-function-response-block {
    max-width: 600px;
  }
</style>
