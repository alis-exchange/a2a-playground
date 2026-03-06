<script setup lang="ts">
  // Vue & Store Imports
  import { useAgentPlaygroundStore } from '@/pages/playground/store/agentPlayground'
  import { useSnackbarStore } from '@/store/snackbar'
  import { v7 as uuidv7 } from 'uuid'
  import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import { onBeforeRouteLeave } from 'vue-router'
  import { useDisplay } from 'vuetify'

  // Component Imports
  import ContentInput from '@/pages/playground/components/ContentInput/ContentInput.vue'
  import type { MessagePart } from '@/pages/playground/components/ContentInput/types'
  import MessageBubble from '@/pages/playground/components/MessageBubble.vue'
  import PendingFunctionResponseBlock from '@/pages/playground/components/PendingFunctionResponseBlock.vue'
  import PlaygroundSidebar from '@/pages/playground/components/PlaygroundSidebar.vue'
  import TaskStatusWidget from '@/pages/playground/components/TaskStatusWidget.vue'
  import { isFunctionCall, normalizeDataPartToObject } from '@/pages/playground/store/dataPartUtils'
  import { ConversationMessage, streamResponseToAgentMessage, useMessagesStore, userMessageToAgentMessage } from '@/pages/playground/store/messages'
  import { Data } from '@a2ui/lit/0.8'
  import type { Surface } from '@a2ui/lit/ui'
  import type { StreamResponse } from '@local/a2a-js'
  import { TaskState } from '@local/a2a-js'

  // Define Stores
  const snackbarStore = useSnackbarStore()
  const messagesStore = useMessagesStore()
  const agentPlaygroundStore = useAgentPlaygroundStore()

  // Refs
  const messagesContainer = ref<HTMLElement | null>(null)
  const playgroundWrapperRef = ref<HTMLElement | null>(null)

  // A2UI Processor (one surface per block; surfaces keyed by a2uiSurfaceId)
  const processor = Data.createSignalA2uiMessageProcessor()

  // Dynamic height calculation
  const { height: windowHeight, width: windowWidth } = useDisplay()
  const layoutTrigger = ref(0)
  const playgroundHeight = computed(() => {
    const _h = windowHeight.value
    const _w = windowWidth.value
    const _trigger = layoutTrigger.value

    if (!playgroundWrapperRef.value) return 400
    const rect = playgroundWrapperRef.value.getBoundingClientRect()
    const availableHeight = _h - rect.top - 24
    return Math.max(200, availableHeight)
  })

  const conversationMessages = computed(() => {
    const msgs = messagesStore.getMessages()
    return msgs.map((m) => ConversationMessage.fromUnifiedMessage(m))
  })

  /** Pending function calls from the latest INPUT_REQUIRED or AUTH_REQUIRED status update */
  const pendingFunctionCalls = computed(() => {
    const msgs = conversationMessages.value
    const result: Array<{ id: string; name: string; args: Record<string, unknown>; taskId: string; contextId: string }> = []
    for (let i = msgs.length - 1; i >= 0; i--) {
      const message = msgs[i]
      if (message?.getPayloadType() !== 'status_update') continue
      const state = message.getState()
      if (state !== TaskState.INPUT_REQUIRED && state !== TaskState.AUTH_REQUIRED) continue
      const parts = message.getParts()
      const contextId = message.getContextId()
      const taskId = message.getTaskId()
      for (const part of parts) {
        if (part?.part?.case !== 'data') continue
        const raw = (part.part as { value?: unknown }).value
        const obj = normalizeDataPartToObject(raw) as Record<string, unknown>
        if (!isFunctionCall(obj)) continue
        result.push({
          id: (obj as { id: string }).id,
          name: (obj as { name: string }).name,
          args: (obj.args as Record<string, unknown>) ?? {},
          taskId,
          contextId,
        })
      }
      break
    }
    return result
  })

  interface MessageItem {
    key: string
    type: 'message' | 'status' | 'pending_function_response' | 'a2ui'
    message?: ConversationMessage
    statusMessages?: ConversationMessage[]
    isUser?: boolean
    pendingCall?: { id: string; name: string; args: Record<string, unknown>; taskId: string; contextId: string }
    surfaceId?: string
  }

  const messageItems = computed((): MessageItem[] => {
    const msgs = conversationMessages.value
    const statusMessagesByTaskId = new Map<string, ConversationMessage[]>()
    const statusWidgetInserted = new Set<string>()
    const items: MessageItem[] = []
    const pendingCalls = pendingFunctionCalls.value

    for (const message of msgs) {
      if (message.getPayloadType() === 'status_update') {
        const statusUpdate = message.getStatusUpdate()
        const taskId = statusUpdate?.taskId || 'unknown'

        if (!statusMessagesByTaskId.has(taskId)) {
          statusMessagesByTaskId.set(taskId, [])
        }
        statusMessagesByTaskId.get(taskId)!.push(message)
      }
    }

    for (const message of msgs) {
      const payloadType = message.getPayloadType()

      if (payloadType === 'status_update') {
        const statusUpdate = message.getStatusUpdate()
        const taskId = statusUpdate?.taskId || 'unknown'

        if (!statusWidgetInserted.has(taskId)) {
          const statusMsgs = statusMessagesByTaskId.get(taskId)
          if (statusMsgs && statusMsgs.length > 0) {
            items.push({
              key: `status-${taskId}`,
              type: 'status',
              statusMessages: statusMsgs,
              isUser: false,
            })
            const latestStatus = statusMsgs[statusMsgs.length - 1]
            const statusA2uiSurfaceId = latestStatus?.getA2uiSurfaceId()
            if (statusA2uiSurfaceId) {
              items.push({
                key: `a2ui-${statusA2uiSurfaceId}`,
                type: 'a2ui',
                surfaceId: statusA2uiSurfaceId,
                isUser: false,
              })
            }
            statusWidgetInserted.add(taskId)
            const latestState = latestStatus?.getState()
            const showPending = (latestState === TaskState.INPUT_REQUIRED || latestState === TaskState.AUTH_REQUIRED) && pendingCalls.length > 0
            if (showPending) {
              for (let j = 0; j < pendingCalls.length; j++) {
                const call = pendingCalls[j]
                if (call && call.taskId === taskId) {
                  items.push({
                    key: `pending-response-${call.id}-${j}`,
                    type: 'pending_function_response',
                    pendingCall: call,
                    isUser: false,
                  })
                }
              }
            }
          }
        }
      } else {
        items.push({
          key: message.getKey(),
          type: 'message',
          message,
          isUser: message.isUser(),
        })
        const a2uiSurfaceId = message.getA2uiSurfaceId()
        if (a2uiSurfaceId) {
          items.push({
            key: `a2ui-${a2uiSurfaceId}`,
            type: 'a2ui',
            surfaceId: a2uiSurfaceId,
            isUser: false,
          })
        }
      }
    }

    return items
  })

  watch(messageItems, async () => {
    await nextTick()
    scrollToBottom()
  })

  const scrollToBottom = () => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  }

  const selectMessage = (message: ConversationMessage) => {
    agentPlaygroundStore.openDetailsDrawerForMessage(message)
  }

  const sendingFunctionResponse = ref(false)

  const onSendFunctionResponse = async (call: { id: string; name: string; taskId: string; contextId: string }, responsePayload: Record<string, unknown>) => {
    sendingFunctionResponse.value = true
    try {
      await messagesStore.sendFunctionResponse(call.contextId, call.taskId, call.id, call.name, responsePayload, (response: StreamResponse) => {
        const unifiedMessage = streamResponseToAgentMessage(response)
        const messageId = unifiedMessage?.messageId ?? uuidv7()
        const surfaceId = processA2UIParts(response, messageId)
        if (unifiedMessage) {
          if (surfaceId) unifiedMessage.a2uiSurfaceId = surfaceId
          messagesStore.addMessage(unifiedMessage)
        }
      })
    } catch (error) {
      console.error(error)
      snackbarStore.error('Failed to send response')
    } finally {
      sendingFunctionResponse.value = false
      await nextTick()
      scrollToBottom()
    }
  }

  const sendMessage = async (parts: MessagePart[]) => {
    try {
      const messageId = uuidv7()
      const sessionContextId = messagesStore.getOrCreateSessionContextId()

      const userAgentMessage = await userMessageToAgentMessage(parts, messageId, sessionContextId)
      const userUnifiedMessage = {
        type: 'agent_message' as const,
        agentMessage: userAgentMessage,
        timestamp: Date.now(),
        messageId,
      }
      messagesStore.addMessage(userUnifiedMessage)

      await messagesStore.sendStreamingMessage(
        parts,
        messageId,
        (response: StreamResponse) => {
          const unifiedMessage = streamResponseToAgentMessage(response)
          const messageId = unifiedMessage?.messageId ?? uuidv7()
          const surfaceId = processA2UIParts(response, messageId)
          if (unifiedMessage) {
            if (surfaceId) unifiedMessage.a2uiSurfaceId = surfaceId
            messagesStore.addMessage(unifiedMessage)
          }
        },
        sessionContextId,
      )
    } catch (error) {
      console.error(error)
      snackbarStore.error('Failed to send message')
    } finally {
      await nextTick()
      scrollToBottom()
    }
  }

  /**
   * Process A2UI data parts from a stream response into a dedicated surface (one per block).
   * Injects the given surfaceId into the first A2UI message so the processor creates that surface.
   * Returns the surfaceId used, or null if no A2UI parts were found.
   */
  const processA2UIParts = (response: StreamResponse, messageId: string): string | null => {
    const payload = response.payload
    if (!payload) return null

    let parts: unknown[] = []
    if (payload.case === 'msg' && payload.value?.parts) {
      parts = payload.value.parts
    } else if (payload.case === 'statusUpdate' && payload.value?.status?.update?.parts) {
      parts = payload.value.status.update.parts
    } else if (payload.case === 'artifactUpdate' && payload.value?.artifact?.parts) {
      parts = payload.value.artifact.parts
    }

    const a2uiMessages: Record<string, unknown>[] = []
    for (const part of parts) {
      const p = part as { part?: { case: string; value?: { data?: unknown } }; metadata?: { mimeType?: string } }
      if (p.part?.case === 'data' && p.metadata?.mimeType === 'application/json+a2ui') {
        const dataPart = p.part.value
        if (dataPart && typeof dataPart === 'object' && 'data' in dataPart && dataPart.data) {
          a2uiMessages.push(dataPart.data as Record<string, unknown>)
        }
      }
    }

    if (a2uiMessages.length === 0) return null

    const surfaceId = `a2ui-${messageId}`
    const first = a2uiMessages[0]
    if (first && typeof first === 'object') {
      first.surfaceId = surfaceId
      if (!('surface_id' in first)) (first as Record<string, unknown>).surface_id = surfaceId
    }
    processor.processMessages(a2uiMessages)
    return surfaceId
  }

  const getSurfaceForId = (surfaceId: string) => processor.getSurfaces().get(surfaceId) ?? null

  onBeforeRouteLeave(async (_to, _from, next) => {
    agentPlaygroundStore.resetDrawerState()
    await nextTick()
    next()
  })

  onBeforeUnmount(() => {
    agentPlaygroundStore.resetDrawerState()
  })

  onMounted(async () => {
    await nextTick()
    layoutTrigger.value++
    await nextTick()
    scrollToBottom()
  })
</script>

<template>
  <div
    ref="playgroundWrapperRef"
    class="playground-root"
    :style="{ height: playgroundHeight + 'px' }"
  >
    <div class="playground-chat">
      <div
        ref="messagesContainer"
        class="messages-scroll"
      >
        <div
          v-for="item in messageItems"
          :key="item.key"
          :class="['message-row', { 'message-row-user': item.isUser, 'message-row-agent': !item.isUser }]"
        >
          <MessageBubble
            v-if="item.type === 'message' && item.message"
            :message="item.message"
            clickable
            @click="selectMessage"
          />
          <TaskStatusWidget
            v-else-if="item.type === 'status' && item.statusMessages"
            :status-messages="item.statusMessages"
            :clickable="!!(item.statusMessages && item.statusMessages.length > 0)"
            @click="selectMessage"
          />
          <PendingFunctionResponseBlock
            v-else-if="item.type === 'pending_function_response' && item.pendingCall"
            :payload="item.pendingCall"
            :sending="sendingFunctionResponse"
            @send="(response) => onSendFunctionResponse(item.pendingCall!, response)"
          />
          <a2ui-surface
            v-else-if="item.type === 'a2ui' && item.surfaceId"
            :surface-id="item.surfaceId"
            :processor="processor"
            :surface="getSurfaceForId(item.surfaceId)"
          />
        </div>

        <div
          v-if="messageItems.length === 0"
          class="empty-state"
        >
          <div class="d-flex flex-column align-center justify-center h-100 text-center">
            <v-icon
              size="80"
              color="primary"
              class="mb-8"
            >
              smart_toy
            </v-icon>
            <div class="text-h4 mb-4">Agent Playground</div>
            <div
              class="text-body-1 text-medium-emphasis mb-8"
              style="max-width: 600px"
            >
              Use the sidebar to configure your agent's connection parameters and security settings. Start the conversation by sending a message below.
            </div>
          </div>
        </div>
      </div>

      <div class="input-sticky">
        <v-card
          class="w-100 main-textarea-shadow"
          max-width="760px"
        >
          <ContentInput @submit="sendMessage" />
        </v-card>
      </div>
    </div>
    <PlaygroundSidebar v-if="agentPlaygroundStore.sidebarOpen" />
    <v-btn
      v-else
      class="sidebar-reopen-btn"
      icon
      variant="tonal"
      color="primary"
      aria-label="Open sidebar"
      @click="agentPlaygroundStore.setSidebarOpen(true)"
    >
      <v-icon>menu</v-icon>
    </v-btn>
  </div>
</template>

<style lang="scss" scoped>
  .playground-root {
    position: relative;
    min-height: 0;
    display: flex;
    flex-direction: row;
    overflow: hidden;
  }

  .playground-chat {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    max-width: 760px;
    margin: 0 auto;
    width: 100%;
    padding: 0 2rem;
  }

  .messages-scroll {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 4rem 0;
    -webkit-mask-image: linear-gradient(to bottom, black 95%, transparent 100%);
    mask-image: linear-gradient(to bottom, black 95%, transparent 100%);
  }

  .input-sticky {
    flex-shrink: 0;
    padding: 0.5rem 0 1rem;
    display: flex;
    justify-content: center;
  }

  .message-row {
    display: grid;
    grid-template-columns: 1fr;
    width: 100%;
    margin-bottom: 4px;
  }

  .message-row-user {
    justify-items: end;
  }

  .message-row-agent {
    justify-items: start;
  }

  .empty-state {
    height: 100%;
  }

  .sidebar-reopen-btn {
    position: absolute;
    top: 1rem;
    right: 1rem;
    z-index: 4;
  }
</style>
