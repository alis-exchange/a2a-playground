export abstract class MessagePart {
  abstract readonly type: 'text' | 'file' | 'audio'
}

export class TextMessagePart extends MessagePart {
  readonly type = 'text' as const
  content: string // markdown text

  constructor(content: string) {
    super()
    this.content = content
  }
}

export class FileMessagePart extends MessagePart {
  readonly type = 'file' as const
  file: File
  mimeType: string
  fileName: string

  constructor(file: File, mimeType?: string, fileName?: string) {
    super()
    this.file = file
    this.mimeType = mimeType ?? file.type
    this.fileName = fileName ?? file.name
  }
}

export class AudioMessagePart extends MessagePart {
  readonly type = 'audio' as const
  audioBlob: Blob
  mimeType: string
  fileName: string

  constructor(audioBlob: Blob, mimeType?: string, fileName?: string) {
    super()
    this.audioBlob = audioBlob
    this.mimeType = mimeType ?? 'audio/mp4'
    this.fileName = fileName ?? 'recording.mp4'
  }
}
