/**
 * WebSocket 连接管理
 * - 自动重连（最多3次）
 * - 心跳保活（30s）
 * - 消息分发 EventEmitter
 */

type EventHandler = (data: any) => void

class WsClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 3
  private heartbeatInterval = 30000
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private listeners: Map<string, Set<EventHandler>> = new Map()
  private intentionalClose = false

  constructor(url: string) {
    this.url = url
  }

  connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return
    }

    this.intentionalClose = false

    try {
      this.ws = new WebSocket(this.url)
    } catch (e) {
      console.error('[WebSocket] 创建连接失败:', e)
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      console.log('[WebSocket] 已连接')
      this.reconnectAttempts = 0
      this.startHeartbeat()
      this.emit('__connected', null)
    }

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg = JSON.parse(event.data)
        const eventName = msg.event || msg.type || 'message'
        this.emit(eventName, msg.data ?? msg)
      } catch {
        this.emit('message', event.data)
      }
    }

    this.ws.onerror = () => {
      console.error('[WebSocket] 连接错误')
    }

    this.ws.onclose = () => {
      this.stopHeartbeat()
      this.emit('__disconnected', null)
      if (!this.intentionalClose) {
        this.scheduleReconnect()
      }
    }
  }

  disconnect(): void {
    this.intentionalClose = true
    this.stopHeartbeat()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.reconnectAttempts = 0
  }

  send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(typeof data === 'string' ? data : JSON.stringify(data))
    } else {
      console.warn('[WebSocket] 未连接，无法发送消息')
    }
  }

  on(event: string, handler: EventHandler): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(handler)

    return () => {
      this.listeners.get(event)?.delete(handler)
    }
  }

  off(event: string, handler: EventHandler): void {
    this.listeners.get(event)?.delete(handler)
  }

  get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  private emit(event: string, data: any): void {
    this.listeners.get(event)?.forEach((handler) => {
      try {
        handler(data)
      } catch (e) {
        console.error('[WebSocket] 事件处理器错误:', e)
      }
    })
  }

  private startHeartbeat(): void {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, this.heartbeatInterval)
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('[WebSocket] 重连次数已达上限，停止重连')
      this.emit('__reconnect_failed', null)
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 10000)
    console.log(`[WebSocket] ${delay / 1000}s 后进行第 ${this.reconnectAttempts} 次重连`)

    setTimeout(() => {
      this.connect()
    }, delay)
  }
}

// 单例工厂
let instance: WsClient | null = null

export function getWsClient(wsUrl?: string): WsClient {
  if (!instance) {
    const url = wsUrl || import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
    instance = new WsClient(url)
  }
  return instance
}

export function destroyWsClient(): void {
  if (instance) {
    instance.disconnect()
    instance = null
  }
}

export { WsClient }
export default getWsClient
