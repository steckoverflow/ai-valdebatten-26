// Reactive debate state plus the WebSocket connection that feeds it.
//
// Uses Svelte 5 runes in a .svelte.ts module so the state is a single shared
// reactive singleton. The WebSocket auto-reconnects; because the server sends a
// full snapshot on (re)connect, a dropped connection self-heals with no special
// catch-up logic here.

import type {
  Bot,
  Message,
  Envelope,
  SnapshotData,
  TopicChangedData,
  BotTypingData,
  MessageData,
} from './types'

class DebateStore {
  topic = $state('')
  cycleId = $state('')
  endsAt = $state<Date | null>(null)
  bots = $state<Bot[]>([])
  messages = $state<Message[]>([])
  typingBotId = $state<string | null>(null)
  connected = $state(false)

  #socket: WebSocket | null = null
  #reconnectDelay = 1000 // ms, backs off up to a cap
  #botIndex = $derived.by(() => {
    const m = new Map<string, Bot>()
    for (const b of this.bots) m.set(b.id, b)
    return m
  })

  /** Look up a bot by id (for name/color rendering). */
  bot(id: string): Bot | undefined {
    return this.#botIndex.get(id)
  }

  /** The bot currently typing, if any. */
  get typingBot(): Bot | undefined {
    return this.typingBotId ? this.bot(this.typingBotId) : undefined
  }

  /** Open the WebSocket and keep it alive. Call once on app start. */
  connect() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/ws`

    const ws = new WebSocket(url)
    this.#socket = ws

    ws.onopen = () => {
      this.connected = true
      this.#reconnectDelay = 1000
    }

    ws.onmessage = (ev) => {
      try {
        const env = JSON.parse(ev.data) as Envelope
        this.#apply(env)
      } catch (err) {
        console.error('bad frame', err)
      }
    }

    ws.onclose = () => {
      this.connected = false
      this.#scheduleReconnect()
    }

    ws.onerror = () => {
      // onclose will follow and handle reconnection.
      ws.close()
    }
  }

  #scheduleReconnect() {
    setTimeout(() => this.connect(), this.#reconnectDelay)
    this.#reconnectDelay = Math.min(this.#reconnectDelay * 2, 10000)
  }

  #apply(env: Envelope) {
    switch (env.type) {
      case 'snapshot': {
        const d = env.data as SnapshotData
        this.cycleId = d.cycleId
        this.topic = d.topic
        this.endsAt = new Date(d.endsAt)
        this.bots = d.bots
        this.messages = d.messages ?? []
        this.typingBotId = d.typingBotId ?? null
        break
      }
      case 'topic_changed': {
        const d = env.data as TopicChangedData
        this.cycleId = d.cycleId
        this.topic = d.topic
        this.endsAt = new Date(d.endsAt)
        this.bots = d.bots
        this.messages = []
        this.typingBotId = null
        break
      }
      case 'bot_typing': {
        const d = env.data as BotTypingData
        this.typingBotId = d.botId
        break
      }
      case 'message': {
        const d = env.data as MessageData
        this.messages = [...this.messages, d.message]
        this.typingBotId = null
        break
      }
    }
  }
}

export const debate = new DebateStore()
