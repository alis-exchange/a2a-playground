<template>
  <v-navigation-drawer
    v-model="store.detailsDrawerOpen"
    location="right"
    width="380"
    temporary
    color="background"
  >
    <!-- Message Details Mode -->
    <template v-if="store.detailsDrawerMode === 'message' && store.detailsDrawerMessage">
      <div class="pa-4">
        <p class="text-h6 mb-4">Message Details</p>

        <v-list
          lines="two"
          density="compact"
          class="bg-transparent message-details-list"
        >
          <v-list-item class="px-0">
            <template #prepend>
              <v-icon
                color="grey"
                class="mr-2"
              >
                category
              </v-icon>
            </template>
            <v-list-item-title>Type</v-list-item-title>
            <v-list-item-subtitle>{{ store.detailsDrawerMessage.getPayloadTypeLabel() }}</v-list-item-subtitle>
          </v-list-item>

          <v-list-item
            v-if="store.detailsDrawerMessage.getMessageId()"
            class="px-0"
          >
            <template #prepend>
              <v-icon
                color="grey"
                class="mr-2"
              >
                fingerprint
              </v-icon>
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
              <v-icon
                color="grey"
                class="mr-2"
              >
                account_tree
              </v-icon>
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
              <v-icon
                color="grey"
                class="mr-2"
              >
                task
              </v-icon>
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
              <v-icon
                color="grey"
                class="mr-2"
              >
                person
              </v-icon>
            </template>
            <v-list-item-title>Role</v-list-item-title>
            <v-list-item-subtitle>{{ store.detailsDrawerMessage.getRole() }}</v-list-item-subtitle>
          </v-list-item>

          <v-list-item
            v-if="store.detailsDrawerMessage.getTimestamp()"
            class="px-0"
          >
            <template #prepend>
              <v-icon
                color="grey"
                class="mr-2"
              >
                schedule
              </v-icon>
            </template>
            <v-list-item-title>Timestamp</v-list-item-title>
            <v-list-item-subtitle>{{ formatTimestamp(store.detailsDrawerMessage.getTimestamp()) }}</v-list-item-subtitle>
          </v-list-item>

          <v-list-item
            v-if="store.detailsDrawerMessage.getStateText()"
            class="px-0"
          >
            <template #prepend>
              <v-icon
                color="grey"
                class="mr-2"
              >
                sync_alt
              </v-icon>
            </template>
            <v-list-item-title>State</v-list-item-title>
            <v-list-item-subtitle>{{ store.detailsDrawerMessage.getStateText() }}</v-list-item-subtitle>
          </v-list-item>

          <v-list-item
            v-if="store.detailsDrawerMessage.getPayloadType() === 'status_update'"
            class="px-0"
          >
            <template #prepend>
              <v-icon
                color="grey"
                class="mr-2"
              >
                flag
              </v-icon>
            </template>
            <v-list-item-title>Final</v-list-item-title>
            <v-list-item-subtitle>{{ store.detailsDrawerMessage.isFinal() ? 'Yes' : 'No' }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </div>
    </template>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
  import { useAgentPlaygroundStore } from '@/pages/playground/store/agentPlayground'
  import { useSnackbarStore } from '@/store/snackbar'

  const store = useAgentPlaygroundStore()
  const snackbarStore = useSnackbarStore()

  const formatTimestamp = (timestamp: number): string => {
    return new Date(timestamp).toLocaleString()
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
</script>

<style lang="scss" scoped>
  .text-truncate {
    max-width: 280px;
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
</style>
