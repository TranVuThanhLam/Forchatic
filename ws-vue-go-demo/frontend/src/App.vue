<template>
  <div style="max-width:720px;margin:20px auto;font-family:Arial,Helvetica,sans-serif;">
    <h2>Vue + WebSocket Demo</h2>
    <div style="margin-bottom:8px">Status: <strong>{{ status }}</strong></div>

    <div style="display:flex; gap:8px; margin-bottom:8px;">
      <input v-model="message" placeholder="Nhập tin nhắn..." style="flex:1;padding:8px"/>
      <button @click="send" style="padding:8px 12px">Gửi</button>
    </div>

    <div style="border:1px solid #ddd; padding:12px; min-height:200px;">
      <div v-for="(m, i) in chat" :key="i" style="margin:6px 0;">📩 {{ m }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const status = ref('Đang kết nối...')
const chat = ref([])
const message = ref('')
let ws = null

function buildWSUrl() {
  // Use secure WSS when served under HTTPS (e.g. Tailscale Funnel)
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  // When running dev (vite on 5173), location.host will be e.g. localhost:5173,
  // so we connect to backend at :8080 explicitly when host includes 5173.
  let host = window.location.host
  // If page served by vite dev server, default backend port is 8080
  if (host.includes(':5173')) {
    host = window.location.hostname + ':8080'
  }
  return `${protocol}://${host}/ws`
}

function connect() {
  const url = buildWSUrl()
  ws = new WebSocket(url)
  ws.onopen = () => {
    status.value = '✅ Đã kết nối'
    // optional: announce
    ws.send('Hello from client (Vue)')
  }
  ws.onmessage = (ev) => {
    chat.value.push(ev.data)
  }
  ws.onclose = () => {
    status.value = '❌ Mất kết nối'
    // simple reconnect attempt after 1s
    setTimeout(() => connect(), 1000)
  }
  ws.onerror = (e) => {
    console.error('ws error', e)
  }
}

onMounted(() => connect())
onBeforeUnmount(() => {
  if (ws) ws.close()
})

function send() {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(message.value)
    message.value = ''
  }
}
</script>
