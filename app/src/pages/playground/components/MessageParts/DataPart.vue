<template>
  <v-card
    variant="outlined"
    class="data-part"
  >
    <v-expansion-panels>
      <v-expansion-panel>
        <v-expansion-panel-title>
          {{ title }}
        </v-expansion-panel-title>
        <v-expansion-panel-text>
          <pre class="json-content">{{ formattedJson }}</pre>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-card>
</template>

<script setup lang="ts">
  import { normalizeDataPartToObject, isFunctionCall } from '@/pages/playground/store/dataPartUtils'
  import type { Part } from '@local/a2a-js'
  import { computed } from 'vue'

  interface Props {
    part: Part
  }

  const props = defineProps<Props>()

  const parsed = computed(() => {
    try {
      const raw = props.part.content?.case === 'data' ? props.part.content.value : undefined
      if (!raw) return null
      return normalizeDataPartToObject(raw) as Record<string, unknown>
    } catch {
      return null
    }
  })

  const isFunctionCallPart = computed(() => parsed.value !== null && isFunctionCall(parsed.value))

  const title = computed(() => {
    if (isFunctionCallPart.value && parsed.value) {
      const name = (parsed.value as { name?: string }).name ?? 'unknown'
      const args = (parsed.value as { args?: Record<string, unknown> }).args
      const argsStr =
        args && typeof args === 'object'
          ? Object.entries(args)
              .slice(0, 3)
              .map(([k, v]) => `${k}: ${typeof v === 'object' ? JSON.stringify(v) : String(v)}`)
              .join(', ')
          : ''
      return argsStr ? `Function call: ${name}(${argsStr})` : `Function call: ${name}()`
    }
    return 'Structured Data'
  })

  const formattedJson = computed(() => {
    try {
      const raw = props.part.content?.case === 'data' ? props.part.content.value : undefined
      const obj = raw !== undefined ? normalizeDataPartToObject(raw) : {}
      return JSON.stringify(obj, null, 2)
    } catch {
      return 'Invalid data'
    }
  })

  // Watchers

  // Functions

  // Lifecycle Hooks
</script>

<style lang="scss" scoped>
  .data-part {
    margin-top: 4px;
    border-radius: 16px;
  }

  .json-content {
    font-family: monospace;
    font-size: 12px;
    background-color: rgba(0, 0, 0, 0.05);
    padding: 8px;
    border-radius: 8px;
    overflow-x: auto;
  }
</style>
