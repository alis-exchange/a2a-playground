import { createA2AClient } from '@/clients/a2aClient'
import {
  MessageSchema,
  PartSchema,
  Role,
  SendMessageRequestSchema,
  TaskState,
  type Message,
  type Part,
  type StreamResponse,
  type Task,
  type TaskArtifactUpdateEvent,
  type TaskStatusUpdateEvent,
} from '@local/a2a-js'
import type { JsonValue } from '@bufbuild/protobuf'
import { create, fromJson } from '@bufbuild/protobuf'
import { ValueSchema } from '@bufbuild/protobuf/wkt'
import { defineStore } from 'pinia'
import { v7 as uuidv7 } from 'uuid'
import { ref } from 'vue'
import type { MessagePart } from '../components/ContentInput/types'
import { AudioMessagePart, FileMessagePart, TextMessagePart } from '../components/ContentInput/types'

const STORE_ID = 'messages'

// Helper: convert plain object to Value (for Part content.data)
function valueFromObject(obj: Record<string, unknown>) {
  return fromJson(ValueSchema, obj as JsonValue)
}

function bytesToBase64(bytes: Uint8Array): string {
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]!)
  }
  return btoa(binary)
}

function stringToUtf8Bytes(str: string): Uint8Array {
  return new TextEncoder().encode(str)
}

async function fileToBase64Bytes(file: File): Promise<Uint8Array> {
  const rawBytes = new Uint8Array(await file.arrayBuffer())
  return stringToUtf8Bytes(bytesToBase64(rawBytes))
}

async function blobToBase64Bytes(blob: Blob): Promise<Uint8Array> {
  const rawBytes = new Uint8Array(await blob.arrayBuffer())
  return stringToUtf8Bytes(bytesToBase64(rawBytes))
}

function getStateText(state: TaskState): string {
  switch (state) {
    case TaskState.SUBMITTED:
      return 'Task submitted'
    case TaskState.WORKING:
      return 'Task in progress'
    case TaskState.COMPLETED:
      return 'Task completed'
    case TaskState.FAILED:
      return 'Task failed'
    case TaskState.CANCELED:
      return 'Task cancelled'
    case TaskState.INPUT_REQUIRED:
      return 'Input required'
    case TaskState.REJECTED:
      return 'Task rejected'
    case TaskState.AUTH_REQUIRED:
      return 'Authentication required'
    default:
      return 'Status update'
  }
}

// Ensure status update has displayable text (for @local/a2a-js TaskStatusUpdateEvent)
function ensureStatusHasText(statusEvent: TaskStatusUpdateEvent): void {
  const status = statusEvent.status
  if (!status?.message) return
  const parts = status.message.parts || []
  const hasText = parts.some((p) => p.content?.case === 'text' && p.content.value?.trim())
  if (hasText) return
  const stateText = getStateText(status.state ?? TaskState.UNSPECIFIED)
  // Add text part - we need to clone/modify. In Buf, messages are typically immutable.
  // For display we'll handle missing text in ConversationMessage.getParts() / getStateText.
}

export type UnifiedMessage = {
  type: 'agent_message' | 'task' | 'status_update' | 'artifact_update'
  agentMessage?: Message
  task?: Task
  statusUpdate?: TaskStatusUpdateEvent
  artifactUpdate?: TaskArtifactUpdateEvent
  timestamp: number
  messageId: string
}

export function streamResponseToAgentMessage(response: StreamResponse): UnifiedMessage | null {
  const now = Date.now()
  const payload = response.payload

  if (!payload || payload.case === undefined) return null

  switch (payload.case) {
    case 'message': {
      const msg = payload.value
      if (!msg) return null
      return {
        type: 'agent_message',
        agentMessage: msg,
        timestamp: now,
        messageId: msg.messageId || uuidv7(),
      }
    }
    case 'task': {
      const task = payload.value
      if (!task) return null
      return { type: 'task', task, timestamp: now, messageId: uuidv7() }
    }
    case 'statusUpdate': {
      const statusUpdate = payload.value
      if (!statusUpdate) return null
      ensureStatusHasText(statusUpdate)
      return {
        type: 'status_update',
        statusUpdate,
        timestamp: now,
        messageId: uuidv7(),
      }
    }
    case 'artifactUpdate': {
      const artifactUpdate = payload.value
      if (!artifactUpdate) return null
      return {
        type: 'artifact_update',
        artifactUpdate,
        timestamp: now,
        messageId: uuidv7(),
      }
    }
    default:
      return null
  }
}

export async function userMessageToAgentMessage(
  parts: MessagePart[],
  messageId?: string,
  contextId?: string
): Promise<Message> {
  const msgId = messageId || uuidv7()
  const agentParts: Part[] = []

  for (const part of parts) {
    if (part instanceof TextMessagePart) {
      agentParts.push(
        create(PartSchema, { content: { case: 'text', value: part.content } })
      )
    } else if (part instanceof FileMessagePart) {
      const arrayBuffer = await part.file.arrayBuffer()
      const rawBytes = new Uint8Array(arrayBuffer)
      agentParts.push(
        create(PartSchema, {
          content: { case: 'raw', value: rawBytes },
          mediaType: part.mimeType || '',
          filename: part.fileName || '',
        })
      )
    } else if (part instanceof AudioMessagePart) {
      const arrayBuffer = await part.audioBlob.arrayBuffer()
      const rawBytes = new Uint8Array(arrayBuffer)
      agentParts.push(
        create(PartSchema, {
          content: { case: 'raw', value: rawBytes },
          mediaType: part.mimeType || 'audio/mp4',
          filename: part.fileName || 'recording.mp4',
        })
      )
    }
  }

  return create(MessageSchema, {
    messageId: msgId,
    contextId: contextId || '',
    taskId: '',
    role: Role.USER,
    parts: agentParts,
    extensions: [],
    referenceTaskIds: [],
  })
}

export class ConversationMessage {
  key: string
  unifiedMessage: UnifiedMessage
  payloadType: 'task' | 'msg' | 'status_update' | 'artifact_update'

  constructor(unifiedMessage: UnifiedMessage) {
    this.unifiedMessage = unifiedMessage
    this.payloadType =
      unifiedMessage.type === 'agent_message' ? 'msg' : unifiedMessage.type
    this.key = `${unifiedMessage.messageId}-${uuidv7()}`
  }

  getKey(): string {
    return this.key
  }

  getPayloadType(): 'task' | 'msg' | 'status_update' | 'artifact_update' {
    return this.payloadType
  }

  isUser(): boolean {
    const u = this.unifiedMessage
    if (u.type === 'agent_message' && u.agentMessage) {
      return u.agentMessage.role === Role.USER
    }
    return false
  }

  getParts(): Part[] {
    const u = this.unifiedMessage
    if (u.type === 'agent_message' && u.agentMessage) {
      return u.agentMessage.parts || []
    }
    if (u.type === 'artifact_update' && u.artifactUpdate?.artifact) {
      return u.artifactUpdate.artifact.parts || []
    }
    if (u.type === 'status_update' && u.statusUpdate?.status?.message) {
      return u.statusUpdate.status.message.parts || []
    }
    if (u.type === 'task' && u.task?.status?.message) {
      return u.task.status.message.parts || []
    }
    return []
  }

  getMessageId(): string {
    return this.unifiedMessage.messageId
  }

  getContextId(): string {
    const u = this.unifiedMessage
    if (u.type === 'agent_message' && u.agentMessage) return u.agentMessage.contextId || ''
    if (u.type === 'status_update' && u.statusUpdate) return u.statusUpdate.contextId || ''
    if (u.type === 'task' && u.task) return u.task.contextId || ''
    if (u.type === 'artifact_update' && u.artifactUpdate) return u.artifactUpdate.contextId || ''
    return ''
  }

  getTaskId(): string {
    const u = this.unifiedMessage
    if (u.type === 'agent_message' && u.agentMessage) return u.agentMessage.taskId || ''
    if (u.type === 'status_update' && u.statusUpdate) return u.statusUpdate.taskId || ''
    if (u.type === 'task' && u.task) {
      const name = u.task.id || ''
      return name.startsWith('tasks/') ? name.slice(6) : name
    }
    if (u.type === 'artifact_update' && u.artifactUpdate) return u.artifactUpdate.taskId || ''
    return ''
  }

  getTimestamp(): number {
    return this.unifiedMessage.timestamp
  }

  getState(): TaskState | undefined {
    if (this.unifiedMessage.type === 'status_update' && this.unifiedMessage.statusUpdate) {
      return this.unifiedMessage.statusUpdate.status?.state
    }
    return undefined
  }

  getStateText(): string | undefined {
    const state = this.getState()
    return state !== undefined ? getStateText(state) : undefined
  }

  getPayloadTypeLabel(): string {
    switch (this.payloadType) {
      case 'msg':
        return 'Message'
      case 'task':
        return 'Task'
      case 'status_update':
        return 'Status Update'
      case 'artifact_update':
        return 'Artifact Update'
      default:
        return 'Unknown'
    }
  }

  getRole(): string {
    const u = this.unifiedMessage
    if (u.type === 'agent_message' && u.agentMessage) {
      return u.agentMessage.role === Role.USER ? 'user' : 'agent'
    }
    return ''
  }

  getStatusUpdate(): TaskStatusUpdateEvent | undefined {
    const u = this.unifiedMessage
    return u.type === 'status_update' ? u.statusUpdate : undefined
  }

  isFinal(): boolean {
    return false
  }

  getSourceMessage(): Message | undefined {
    const u = this.unifiedMessage
    if (u.type === 'agent_message') return u.agentMessage
    if (u.type === 'status_update' && u.statusUpdate?.status?.message) return u.statusUpdate.status.message
    return undefined
  }

  static fromUnifiedMessage(unifiedMessage: UnifiedMessage): ConversationMessage {
    return new ConversationMessage(unifiedMessage)
  }
}

interface MessagesStoreReturn {
  messages: ReturnType<typeof ref<UnifiedMessage[]>>
  sendStreamingMessage: (
    parts: MessagePart[],
    messageId: string,
    onStreamMessage?: (response: StreamResponse) => void,
    contextId?: string
  ) => Promise<string>
  sendFunctionResponse: (
    contextId: string,
    taskId: string,
    callId: string,
    callName: string,
    responsePayload: Record<string, unknown>,
    onStreamMessage?: (response: StreamResponse) => void
  ) => Promise<string>
  addMessage: (unifiedMessage: UnifiedMessage) => void
  getMessages: () => UnifiedMessage[]
  clearMessages: () => void
  getOrCreateSessionContextId: () => string
  clearSessionContextId: () => void
}

export const useMessagesStore = defineStore(STORE_ID, (): MessagesStoreReturn => {
  const messages = ref<UnifiedMessage[]>([])
  const sessionContextId = ref<string | null>(null)

  const getOrCreateSessionContextId = (): string => {
    if (sessionContextId.value) return sessionContextId.value
    const id = uuidv7()
    sessionContextId.value = id
    return id
  }

  const clearSessionContextId = () => {
    sessionContextId.value = null
  }

  const addMessage = (unifiedMessage: UnifiedMessage) => {
    const exists = messages.value.some((m) => m.messageId === unifiedMessage.messageId)
    if (exists) return
    const msgTime = unifiedMessage.timestamp
    let insertIndex = messages.value.length
    for (let i = 0; i < messages.value.length; i++) {
      if ((messages.value[i]?.timestamp ?? 0) > msgTime) {
        insertIndex = i
        break
      }
    }
    const next = [...messages.value]
    next.splice(insertIndex, 0, unifiedMessage)
    messages.value = next
  }

  const getMessages = (): UnifiedMessage[] => {
    return messages.value
  }

  const clearMessages = () => {
    messages.value = []
  }

  const sendStreamingMessage = async (
    parts: MessagePart[],
    messageId: string,
    onStreamMessage?: (response: StreamResponse) => void,
    contextId?: string
  ): Promise<string> => {
    const client = createA2AClient()
    const sessionId = contextId || getOrCreateSessionContextId()

    const agentParts: Part[] = []
    for (const part of parts) {
      if (part instanceof TextMessagePart) {
        agentParts.push(create(PartSchema, { content: { case: 'text', value: part.content } }))
      } else if (part instanceof FileMessagePart) {
        const base64Bytes = await fileToBase64Bytes(part.file)
        agentParts.push(
          create(PartSchema, {
            content: { case: 'raw', value: base64Bytes },
            mediaType: part.mimeType || '',
            filename: part.fileName || '',
          })
        )
      } else if (part instanceof AudioMessagePart) {
        const base64Bytes = await blobToBase64Bytes(part.audioBlob)
        agentParts.push(
          create(PartSchema, {
            content: { case: 'raw', value: base64Bytes },
            mediaType: part.mimeType || 'audio/mp4',
            filename: part.fileName || 'recording.mp4',
          })
        )
      }
    }

    const agentMessage = create(MessageSchema, {
      messageId: messageId || uuidv7(),
      contextId: sessionId,
      taskId: '',
      role: Role.USER,
      parts: agentParts,
      extensions: [],
      referenceTaskIds: [],
    })

    const request = create(SendMessageRequestSchema, {
      tenant: '',
      message: agentMessage,
    })

    let resolvedContextId = sessionId
    const stream = client.sendStreamingMessage(request)

    for await (const response of stream) {
      if (onStreamMessage) onStreamMessage(response)
      const payload = response.payload
      if (payload?.case === 'task' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
      if (payload?.case === 'message' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
      if (payload?.case === 'statusUpdate' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
      if (payload?.case === 'artifactUpdate' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
    }

    return resolvedContextId
  }

  const sendFunctionResponse = async (
    contextId: string,
    taskId: string,
    callId: string,
    callName: string,
    responsePayload: Record<string, unknown>,
    onStreamMessage?: (response: StreamResponse) => void
  ): Promise<string> => {
    const client = createA2AClient()
    const dataPayload = { id: callId, name: callName, response: responsePayload }
    const dataValue = valueFromObject(dataPayload)
    const part = create(PartSchema, {
      content: { case: 'data', value: dataValue },
      metadata: { adk_type: 'function_response' },
    })

    const agentMessage = create(MessageSchema, {
      messageId: uuidv7(),
      contextId,
      taskId: taskId || '',
      role: Role.USER,
      parts: [part],
      extensions: [],
      referenceTaskIds: taskId ? [taskId] : [],
    })

    const request = create(SendMessageRequestSchema, {
      tenant: '',
      message: agentMessage,
    })

    let resolvedContextId = contextId
    const stream = client.sendStreamingMessage(request)

    for await (const response of stream) {
      if (onStreamMessage) onStreamMessage(response)
      const payload = response.payload
      if (payload?.case === 'task' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
      if (payload?.case === 'message' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
      if (payload?.case === 'statusUpdate' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
      if (payload?.case === 'artifactUpdate' && payload.value?.contextId) {
        resolvedContextId = payload.value.contextId
        sessionContextId.value = resolvedContextId
      }
    }

    return resolvedContextId
  }

  return {
    messages,
    sendStreamingMessage,
    sendFunctionResponse,
    addMessage,
    getMessages,
    clearMessages,
    getOrCreateSessionContextId,
    clearSessionContextId,
  }
})
