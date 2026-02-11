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

      <!-- Confirmation-style shortcut -->
      <div
        v-if="showConfirmationButtons"
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

      <!-- Generic JSON response -->
      <div class="mt-2">
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

  const sending = ref(false)
  const responseText = ref('{}')
  const jsonError = ref('')

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
