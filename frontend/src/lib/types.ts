// Mirror of the Go `internal/protocol` package. Keep these in sync with the
// backend wire format: every server frame is an Envelope { type, data }.

export interface Bot {
  id: string
  name: string
  persona: string
  manifesto: string
  color: string
}

export interface Message {
  id: string
  botId: string
  text: string
  ts: string // RFC3339 timestamp
}

export type EventType = 'snapshot' | 'topic_changed' | 'bot_typing' | 'message'

export interface SnapshotData {
  cycleId: string
  topic: string
  endsAt: string
  bots: Bot[]
  messages: Message[]
  typingBotId?: string
}

export interface TopicChangedData {
  cycleId: string
  topic: string
  endsAt: string
  bots: Bot[]
}

export interface BotTypingData {
  botId: string
}

export interface MessageData {
  message: Message
}

export interface Envelope {
  type: EventType
  data: unknown
}
