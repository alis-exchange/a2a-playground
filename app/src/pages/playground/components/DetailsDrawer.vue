<template>
  <v-dialog
    :model-value="store.detailsDrawerOpen"
    max-width="500"
    scrollable
    persistent
    content-class="details-dialog"
    transition="dialog-transition"
    @update:model-value="(v) => { if (!v) store.closeDetailsDrawer() }"
  >
    <v-card>
      <v-card-title class="d-flex align-center pa-4">
        <span class="text-h6">Message Details</span>
        <v-spacer />
        <v-btn
          icon
          variant="text"
          size="small"
          @click="store.closeDetailsDrawer()"
        >
          <v-icon>close</v-icon>
        </v-btn>
      </v-card-title>
      <v-divider />
      <v-card-text class="pa-4">
        <template v-if="store.detailsDrawerMessage">
          <!-- Message metadata -->
          <v-list
            lines="two"
            density="compact"
            class="bg-transparent message-details-list mb-4"
          >
            <v-list-item class="px-0">
              <template #prepend>
                <v-icon color="grey" class="mr-2">category</v-icon>
              </template>
              <v-list-item-title>Type</v-list-item-title>
              <v-list-item-subtitle>{{ store.detailsDrawerMessage.getPayloadTypeLabel() }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getMessageId()"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">fingerprint</v-icon>
              </template>
              <v-list-item-title class="d-flex align-center">
                Message ID
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  class="ml-1"
                  @click="copyToClipboard(store.detailsDrawerMessage!.getMessageId())"
                >
                  <v-icon size="14">content_copy</v-icon>
                </v-btn>
              </v-list-item-title>
              <v-list-item-subtitle class="text-truncate">
                {{ store.detailsDrawerMessage.getMessageId() }}
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getContextId()"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">account_tree</v-icon>
              </template>
              <v-list-item-title class="d-flex align-center">
                Context ID
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  class="ml-1"
                  @click="copyToClipboard(store.detailsDrawerMessage!.getContextId())"
                >
                  <v-icon size="14">content_copy</v-icon>
                </v-btn>
              </v-list-item-title>
              <v-list-item-subtitle class="text-truncate">
                {{ store.detailsDrawerMessage.getContextId() }}
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getTaskId()"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">task</v-icon>
              </template>
              <v-list-item-title class="d-flex align-center">
                Task ID
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  class="ml-1"
                  @click="copyToClipboard(store.detailsDrawerMessage!.getTaskId())"
                >
                  <v-icon size="14">content_copy</v-icon>
                </v-btn>
              </v-list-item-title>
              <v-list-item-subtitle class="text-truncate">
                {{ store.detailsDrawerMessage.getTaskId() }}
              </v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getRole()"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">person</v-icon>
              </template>
              <v-list-item-title>Role</v-list-item-title>
              <v-list-item-subtitle>{{ store.detailsDrawerMessage.getRole() }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getTimestamp()"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">schedule</v-icon>
              </template>
              <v-list-item-title>Timestamp</v-list-item-title>
              <v-list-item-subtitle>{{ formatTimestamp(store.detailsDrawerMessage.getTimestamp()) }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getStateText()"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">sync_alt</v-icon>
              </template>
              <v-list-item-title>State</v-list-item-title>
              <v-list-item-subtitle>{{ store.detailsDrawerMessage.getStateText() }}</v-list-item-subtitle>
            </v-list-item>

            <v-list-item
              v-if="store.detailsDrawerMessage.getPayloadType() === 'status_update'"
              class="px-0"
            >
              <template #prepend>
                <v-icon color="grey" class="mr-2">flag</v-icon>
              </template>
              <v-list-item-title>Final</v-list-item-title>
              <v-list-item-subtitle>{{ store.detailsDrawerMessage.isFinal() ? 'Yes' : 'No' }}</v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <!-- Task Details (when task ID present) -->
          <div
            v-if="store.detailsDrawerMessage.getTaskId()"
            class="task-details-section"
          >
            <p class="text-subtitle-1 font-weight-medium mb-2">Task Details</p>
            <div
              v-if="taskLoading"
              class="d-flex align-center justify-center py-8"
            >
              <v-progress-circular
                indeterminate
                color="primary"
                size="32"
              />
              <span class="ml-3">Loading task...</span>
            </div>
            <div
              v-else-if="taskError"
              class="py-4"
            >
              <v-alert
                type="error"
                density="compact"
                variant="tonal"
              >
                {{ taskError }}
              </v-alert>
            </div>
            <div
              v-else-if="task"
              class="task-details-content"
            >
              <v-list
                lines="two"
                density="compact"
                class="bg-transparent"
              >
                <v-list-item class="px-0">
                  <template #prepend>
                    <v-icon color="grey" class="mr-2">info</v-icon>
                  </template>
                  <v-list-item-title>Status</v-list-item-title>
                  <v-list-item-subtitle>{{ formatTaskState(task.status?.state) }}</v-list-item-subtitle>
                </v-list-item>

                <v-list-item
                  v-if="task.artifacts && task.artifacts.length > 0"
                  class="px-0"
                >
                  <template #prepend>
                    <v-icon color="grey" class="mr-2">folder_special</v-icon>
                  </template>
                  <v-list-item-title>Artifacts</v-list-item-title>
                  <v-list-item-subtitle>
                    <span
                      v-for="(art, i) in task.artifacts"
                      :key="i"
                    >
                      {{ art.name || 'Unnamed' }}{{ i < (task.artifacts?.length ?? 0) - 1 ? ', ' : '' }}
                    </span>
                  </v-list-item-subtitle>
                </v-list-item>

                <v-list-item
                  v-if="task.history && task.history.length > 0"
                  class="px-0"
                >
                  <template #prepend>
                    <v-icon color="grey" class="mr-2">history</v-icon>
                  </template>
                  <v-list-item-title>History</v-list-item-title>
                  <v-list-item-subtitle>
                    {{ task.history.length }} message(s)
                  </v-list-item-subtitle>
                </v-list-item>
              </v-list>
            </div>
          </div>
        </template>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
  import { create } from '@bufbuild/protobuf'
  import { GetTaskRequestSchema } from '@local/a2a-js'
  import type { Task } from '@local/a2a-js'
  import { TaskState } from '@local/a2a-js'
  import { createA2AClient } from '@/clients/a2aClient'
  import { useAgentPlaygroundStore } from '@/pages/playground/store/agentPlayground'
  import { useSnackbarStore } from '@/store/snackbar'
  import { watch, ref } from 'vue'

  const store = useAgentPlaygroundStore()
  const snackbarStore = useSnackbarStore()

  const task = ref<Task | null>(null)
  const taskLoading = ref(false)
  const taskError = ref<string | null>(null)

  const formatTimestamp = (timestamp: number): string => {
    return new Date(timestamp).toLocaleString()
  }

  const formatTaskState = (state?: TaskState): string => {
    if (state == null) return 'Unknown'
    switch (state) {
      case TaskState.SUBMITTED:
        return 'Submitted'
      case TaskState.WORKING:
        return 'Working'
      case TaskState.COMPLETED:
        return 'Completed'
      case TaskState.FAILED:
        return 'Failed'
      case TaskState.CANCELLED:
        return 'Canceled'
      case TaskState.INPUT_REQUIRED:
        return 'Input required'
      case TaskState.REJECTED:
        return 'Rejected'
      case TaskState.AUTH_REQUIRED:
        return 'Auth required'
      default:
        return 'Unknown'
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      snackbarStore.success('Copied to clipboard')
    } catch (error) {
      console.error('Failed to copy:', error)
      snackbarStore.error('Failed to copy to clipboard')
    }
  }

  const fetchTask = async (taskId: string) => {
    task.value = null
    taskError.value = null
    taskLoading.value = true
    try {
      const client = createA2AClient()
      const request = create(GetTaskRequestSchema, {
        name: `tasks/${taskId}`,
      })
      const response = await client.getTask(request)
      task.value = response
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      taskError.value = msg || 'Failed to load task'
    } finally {
      taskLoading.value = false
    }
  }

  watch(
    () => [store.detailsDrawerOpen, store.detailsDrawerMessage] as const,
    ([open, message]) => {
      if (open && message?.getTaskId()) {
        fetchTask(message.getTaskId())
      } else {
        task.value = null
        taskError.value = null
      }
    },
    { immediate: true }
  )
</script>

<style lang="scss" scoped>
  .text-truncate {
    max-width: 360px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .message-details-list {
    :deep(.v-list-item__content) {
      text-align: left;
    }

    :deep(.v-list-item-title),
    :deep(.v-list-item-subtitle) {
      text-align: left;
    }
  }

  .task-details-section {
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    padding-top: 16px;
  }
</style>
