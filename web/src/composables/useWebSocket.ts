/**
 * Vue3 WebSocket Composable
 * 接收事件名和回调，自动管理生命周期
 * 连接失败时返回降级标志，供调用方回退到轮询
 */

import { ref, onMounted, onUnmounted, type Ref } from 'vue'
import { getWsClient } from '@/utils/websocket'

export interface UseWebSocketOptions {
  /** WebSocket 事件名 */
  event: string
  /** 事件回调 */
  handler: (data: any) => void
  /** WebSocket 连接地址，不传则从环境变量读取 */
  wsUrl?: string
}

export interface UseWebSocketReturn {
  /** 是否已连接 */
  connected: Ref<boolean>
  /** 是否需要降级到轮询 */
  fallback: Ref<boolean>
  /** 发送消息 */
  send: (data: any) => void
}

export function useWebSocket(options: UseWebSocketOptions): UseWebSocketReturn {
  const connected = ref(false)
  const fallback = ref(false)
  const client = getWsClient(options.wsUrl)
  let unsubConnected: (() => void) | null = null
  let unsubDisconnected: (() => void) | null = null
  let unsubEvent: (() => void) | null = null
  let unsubReconnectFailed: (() => void) | null = null

  onMounted(() => {
    unsubConnected = client.on('__connected', () => {
      connected.value = true
      fallback.value = false
    })

    unsubDisconnected = client.on('__disconnected', () => {
      connected.value = false
    })

    unsubEvent = client.on(options.event, options.handler)

    unsubReconnectFailed = client.on('__reconnect_failed', () => {
      fallback.value = true
      connected.value = false
    })

    // 尝试连接
    client.connect()

    // 如果已经连接，同步状态
    if (client.isConnected) {
      connected.value = true
    }
  })

  onUnmounted(() => {
    unsubConnected?.()
    unsubDisconnected?.()
    unsubEvent?.()
    unsubReconnectFailed?.()
  })

  function send(data: any) {
    client.send(data)
  }

  return { connected, fallback, send }
}
