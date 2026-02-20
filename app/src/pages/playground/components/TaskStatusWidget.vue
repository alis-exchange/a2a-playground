<template>
  <div
    class="task-status-widget px-4 mt-2"
    :class="{ 'task-status-clickable': clickable }"
    @click="handleClick"
  >
    <v-progress-circular
      v-if="isWorking"
      :size="16"
      :width="2"
      indeterminate
      color="grey-lighten-1"
      class="mr-2"
    />
    <v-icon
      v-else
      color="grey-lighten-1"
      size="small"
      class="mr-2"
    >
      {{ statusIcon }}
    </v-icon>
    <span class="text-body-2 text-grey">
      {{ statusText }}
    </span>
  </div>
</template>

<script setup lang="ts">
  // Vue & Store Imports
  import type { ConversationMessage } from '@/pages/playground/store/messages'
  import { TaskState } from '@local/a2a-js'
  import { computed } from 'vue'

  // Props
  interface Props {
    statusMessages: ConversationMessage[]
    clickable?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    clickable: false,
  })

  // Define Emits
  const emit = defineEmits<{
    (e: 'click', message: ConversationMessage): void
  }>()

  // Computed
  const latestMessage = computed((): ConversationMessage | null => {
    if (props.statusMessages.length === 0) {
      return null
    }
    return props.statusMessages[props.statusMessages.length - 1] ?? null
  })

  const latestTaskState = computed((): TaskState | undefined => {
    if (!latestMessage.value) return undefined
    const statusUpdate = latestMessage.value.getStatusUpdate()
    return statusUpdate?.status?.state
  })

  const statusText = computed(() => {
    if (!latestMessage.value) return 'Status update'
    const statusUpdate = latestMessage.value.getStatusUpdate()
    if (!statusUpdate) return 'Status update'

    const updateMessage = statusUpdate.status?.update
    const parts = updateMessage?.parts || []
    const textParts = parts
      .filter((p) => p?.part?.case === 'text' && (p.part as { value?: string }).value)
      .map((p) => (p.part?.case === 'text' ? (p.part as { value?: string }).value : ''))

    if (textParts.length > 0) {
      const text = textParts.join(' ').trim()
      if (text) {
        // If the text is from the enum (starts with "Task status: "), extract the state
        if (text.startsWith('Task status: ')) {
          const state = text.substring('Task status: '.length).trim()
          return state || (latestMessage.value!.getStateText() ?? 'Status update')
        }
        if (text.startsWith('Task created: ')) {
          return 'Created'
        }
        return text
      }
    }

    // Fallback: use state enum when update/parts are missing (common with a2a-js)
    return latestMessage.value.getStateText() ?? 'Status update'
  })

  const statusColor = computed(() => {
    const upperStatus = statusText.value.toUpperCase()

    if (upperStatus.includes('COMPLETED') || upperStatus.includes('SUBMITTED')) {
      return 'primary'
    } else if (upperStatus.includes('WORKING') || upperStatus.includes('PROGRESS')) {
      return 'secondary'
    } else if (upperStatus.includes('FAILED') || upperStatus.includes('REJECTED')) {
      return 'error'
    } else if (upperStatus.includes('CANCELLED')) {
      return 'grey'
    } else if (upperStatus.includes('INPUT_REQUIRED') || upperStatus.includes('AUTH_REQUIRED')) {
      return 'tertiary'
    }
    return 'surface-variant'
  })

  const statusIcon = computed(() => {
    const upperStatus = statusText.value.toUpperCase()

    if (upperStatus.includes('COMPLETED')) {
      return 'check_circle'
    } else if (upperStatus.includes('WORKING') || upperStatus.includes('PROGRESS')) {
      return 'avg_pace'
    } else if (upperStatus.includes('FAILED') || upperStatus.includes('REJECTED')) {
      return 'error'
    } else if (upperStatus.includes('CANCELLED')) {
      return 'cancel'
    } else if (upperStatus.includes('INPUT_REQUIRED') || upperStatus.includes('AUTH_REQUIRED')) {
      return 'info'
    } else if (upperStatus.includes('SUBMITTED') || upperStatus.includes('CREATED')) {
      return 'schedule'
    }
    return 'info'
  })

  const textColor = computed(() => {
    const color = statusColor.value
    if (color === 'error') {
      return 'rgb(var(--v-theme-error))'
    } else if (color === 'primary') {
      return 'rgb(var(--v-theme-primary))'
    } else if (color === 'secondary') {
      return 'rgb(var(--v-theme-secondary))'
    }
    return 'rgba(0, 0, 0, 0.6)'
  })

  const iconColor = computed(() => {
    return textColor.value
  })

  const isWorking = computed(() => {
    const state = latestTaskState.value
    return state === TaskState.WORKING
  })

  // Functions
  const handleClick = () => {
    if (props.clickable && latestMessage.value) {
      emit('click', latestMessage.value)
    }
  }
</script>

<style lang="scss" scoped>
  .task-status-widget {
    display: flex;
    align-items: center;
    max-width: 600px;

    &.task-status-clickable {
      cursor: pointer;
      transition: opacity 0.15s ease;
      border-radius: 8px;

      &:hover {
        opacity: 0.75;
        background-color: rgba(var(--v-theme-on-surface), 0.04);
      }
    }
  }
</style>
