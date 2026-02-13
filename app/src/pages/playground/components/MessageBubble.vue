<template>
  <!-- Don't render anything if there are no valid parts to display -->
  <v-hover v-if="hasValidParts">
    <template v-slot:default="{ isHovering, props: hoverProps }">
      <div
        v-bind="hoverProps"
        :class="['message-bubble-wrapper', { 'message-bubble-wrapper--user': isUser, 'message-bubble-wrapper--agent': !isUser }]"
      >
        <div :class="['message-bubble px-4 py-1', { 'bg-surfaceContainer message-user': isUser, 'message-agent': !isUser }]">
          <div
            v-for="(part, idx) in validParts"
            :key="`${message.getKey()}.${idx}`"
            class="message-part"
          >
            <TextPart
              v-if="hasText(part)"
              :part="part"
            />
            <FilePart
              v-else-if="hasFile(part)"
              :part="part"
            />
            <DataPart
              v-else-if="hasData(part)"
              :part="part"
            />
          </div>
        </div>
        <div
          v-if="clickable"
          :class="['actions-wrapper', { 'actions-wrapper--visible': isHovering, 'actions-wrapper--user': isUser, 'actions-wrapper--agent': !isUser }]"
        >
          <v-btn
            size="x-small"
            density="compact"
            icon
            variant="plain"
            @click.stop="handleCopy"
          >
            <v-icon size="16">content_copy</v-icon>
          </v-btn>
          <v-btn
            size="x-small"
            density="compact"
            icon
            variant="plain"
            @click.stop="handleOpenDetails"
          >
            <v-icon size="16">info</v-icon>
          </v-btn>
        </div>
      </div>
    </template>
  </v-hover>
</template>

<script setup lang="ts">
  // Vue & Store Imports
  import { normalizeDataPartToObject } from '@/pages/playground/store/dataPartUtils'
  import type { ConversationMessage } from '@/pages/playground/store/messages'
  import { useSnackbarStore } from '@/store/snackbar'
  import type { Part } from '@local/a2a-js'
  import { computed } from 'vue'

  // Component Imports
  import DataPart from './MessageParts/DataPart.vue'
  import FilePart from './MessageParts/FilePart.vue'
  import TextPart from './MessageParts/TextPart.vue'

  // Props
  interface Props {
    message: ConversationMessage
    clickable?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    clickable: false,
  })

  // Define Stores
  const snackbarStore = useSnackbarStore()

  // Define Emits
  const emit = defineEmits<{
    (e: 'click', message: ConversationMessage): void
  }>()

  // Computed
  const isUser = computed(() => props.message?.isUser() ?? false)

  const parts = computed((): Part[] => {
    if (!props.message) return []
    try {
      return props.message.getParts() ?? []
    } catch (error) {
      console.error('Error getting message parts:', error)
      return []
    }
  })

  // Filter to only parts that have actual content (text, file, or data)
  const validParts = computed((): Part[] => {
    return parts.value.filter((part) => part && (hasText(part) || hasFile(part) || hasData(part)))
  })

  // Check if there are any valid parts to display
  const hasValidParts = computed(() => validParts.value.length > 0)

  // Functions
  const handleOpenDetails = () => {
    if (props.clickable) {
      emit('click', props.message)
    }
  }

  const handleCopy = async () => {
    const segments: string[] = []

    // Text from text parts
    const textParts = validParts.value
      .filter((part) => hasText(part))
      .map((part) => (part?.part?.case === 'text' ? (part.part as { value?: string }).value : ''))
      .filter(Boolean)
    const text = textParts.join('\n').trim()
    if (text) segments.push(text)

    // JSON from data parts
    for (const part of validParts.value) {
      if (!hasData(part)) continue
      try {
        const raw = part?.part?.case === 'data' ? (part.part as { value?: unknown }).value : undefined
        const obj = normalizeDataPartToObject(raw)
        if (Object.keys(obj).length > 0) {
          segments.push(JSON.stringify(obj, null, 2))
        }
      } catch {
        // skip part if serialization fails
      }
    }

    const toCopy = segments.join('\n\n').trim()
    if (!toCopy) {
      snackbarStore.error('No text to copy')
      return
    }

    try {
      await navigator.clipboard.writeText(toCopy)
      snackbarStore.success('Copied to clipboard')
    } catch (error) {
      console.error('Failed to copy:', error)
      snackbarStore.error('Failed to copy to clipboard')
    }
  }

  const hasText = (part: Part | null | undefined): boolean => {
    if (!part) return false
    const p = part.part
    return p?.case === 'text' ? !!((p as { value?: string }).value && String((p as { value?: string }).value).trim().length > 0) : false
  }

  const hasFile = (part: Part | null | undefined): boolean => {
    if (!part) return false
    return part.part?.case === 'file'
  }

  const hasData = (part: Part | null | undefined): boolean => {
    if (!part) return false
    return part.part?.case === 'data'
  }
</script>

<style lang="scss" scoped>
  .message-bubble {
    word-wrap: break-word;
  }

  .message-user {
    max-width: 452px;
    border-radius: 24px 4px 24px 24px;
  }

  .message-part {
    margin-top: 4px;

    &:first-child {
      margin-top: 0;
    }
  }

  .message-bubble-wrapper {
    display: inline-flex;
    flex-direction: column;
    margin-bottom: 24px;

    &--user {
      align-items: flex-end;
    }

    &--agent {
      align-items: flex-start;
    }
  }

  .actions-wrapper {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 2px;
    margin-top: 4px;
    height: 24px;
    opacity: 0;
    transition: opacity 0.15s ease;

    &--visible {
      opacity: 1;
    }

    &--user {
      justify-content: flex-end;
    }

    &--agent {
      justify-content: flex-start;
    }
  }
</style>
