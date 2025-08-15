<template>
  <div class="card shadow-sm p-4">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h5 class="mb-0">Phòng: {{ roomID }}</h5>
      <small class="text-muted">Bạn: <strong>{{ username }}</strong></small>
    </div>

    <div class="chat-box mb-3 p-3" ref="chatBox">
      <template v-for="(group, gi) in groupedMessages" :key="gi">
        <div :class="['mb-2 group-wrap', group.sender === username ? 'text-end' : 'text-start']">
          <!-- Sender label (italic, faded) -->
          <div v-if="group.sender" :class="['sender-label', group.sender === username ? 'me-label' : 'other-label']">
            <em>{{ group.sender }}</em>
          </div>

          <!-- messages in group -->
          <div
            v-for="(m, mi) in group.items"
            :key="m.id"
            :class="[
              'msg-bubble',
              group.sender === username ? 'msg-me' : 'msg-other',
              mi > 0 ? 'grouped' : ''
            ]"
          >
            {{ m.content }}
            <div class="msg-time text-muted">
              {{ formatTime(m.ts) }}
            </div>
          </div>
        </div>
      </template>
    </div>

    <div class="input-group">
      <input type="text" class="form-control" v-model="input" placeholder="Nhập tin nhắn..."
        @keyup.enter="sendMessage" />
      <button class="btn btn-success" @click="sendMessage">Gửi</button>
    </div>
  </div>
</template>

<script>
export default {
  props: ['username', 'roomID'],
  data() {
    return {
      ws: null,
      input: '',
      rawMessages: [], // flat array of messages in chronological order
    }
  },
  computed: {
    // group consecutive messages by same sender
    groupedMessages() {
      const groups = []
      for (const m of this.rawMessages) {
        const last = groups[groups.length - 1]
        if (last && last.sender === m.sender) {
          last.items.push(m)
        } else {
          groups.push({ sender: m.sender, items: [m] })
        }
      }
      return groups
    }
  },
  mounted() {
    // load history (page 1)
    fetch(`/history?room=${encodeURIComponent(this.roomID)}&page=1&limit=100`)
      .then(r => r.json())
      .then(arr => {
        // arr is array of messages in chronological order (older -> newer)
        this.rawMessages = arr.map(a => ({
          id: a.id,
          sender: a.sender,
          content: a.content,
          ts: a.ts
        }))
        this.$nextTick(this.scrollBottom)
      })
      .catch(err => console.error('load history', err))

    // connect WS
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    let host = window.location.host
    // when dev server (vite) is used on :5173, connect backend on :8080
    if (host.includes(':5173')) host = window.location.hostname + ':8080'
    const url = `${protocol}://${host}/ws?username=${encodeURIComponent(this.username)}&room=${encodeURIComponent(this.roomID)}`
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log('ws open')
    }
    this.ws.onmessage = (ev) => {
      try {
        const payload = JSON.parse(ev.data)
        if (payload.type === 'message' && payload.message) {
          const m = payload.message
          // append (chronological order)
          this.rawMessages.push({
            id: m.id,
            sender: m.sender,
            content: m.content,
            ts: m.ts
          })
          this.$nextTick(this.scrollBottom)
        } else if (payload.type === 'system') {
          // optional system notifications
        }
      } catch (e) {
        console.error('invalid payload', e)
      }
    }
    this.ws.onclose = () => {
      console.log('ws closed, will try reconnect in 1s')
      setTimeout(() => location.reload(), 1000) // simple approach
    }
  },
  methods: {
    sendMessage() {
      const text = this.input.trim()
      if (!text) return
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(text)
        this.input = '' // clear after send
      }
    },
    scrollBottom() {
      const el = this.$refs.chatBox
      if (el) el.scrollTop = el.scrollHeight
    },
    formatTime(ts) {
      const d = new Date(ts)
      return d.toLocaleTimeString()
    }
  }
}
</script>

<style scoped>
.chat-box {
  height: 400px;
  overflow-y: auto;
  background: #f5f7fb;
  border-radius: 8px;
  padding-bottom: 8px;
}

/* message bubble common */
.msg-bubble {
  display: block;           /* mỗi tin 1 dòng */
  padding: 10px 12px;
  border-radius: 12px;
  margin: 10px 0;          /* khoảng cách giữa các nhóm */
  max-width: 70%;
  word-wrap: break-word;
  box-shadow: 0 1px 2px rgba(16,24,40,0.04);
  line-height: 1.35;
  position: relative;
}

/* khi là tin nhắn trong cùng nhóm thì khoảng cách nhỏ hơn */
.msg-bubble.grouped {
  margin-top: 6px;
  margin-bottom: 6px;
  border-top-left-radius: 10px;
  border-top-right-radius: 10px;
}

/* my messages (right) - màu hồng */
.msg-me {
  background: linear-gradient(90deg, #ff9acb, #ff6fa6);
  color: #111;
  margin-left: auto;
  text-align: right;
}

/* others (left) */
.msg-other {
  background: #ffffff;
  border: 1px solid #e6e9ef;
  color: #111827;
  margin-right: auto;
  text-align: left;
}

/* sender label: italic & faded and positioned above the group */
.sender-label {
  font-style: italic;
  color: rgba(33,37,41,0.6);
  margin-bottom: 6px;
  font-size: 0.9rem;
}

/* if sender is me, align label to right slightly */
.me-label {
  text-align: right;
}

/* message time */
.msg-time {
  opacity: 0.75;
  font-size: 0.75rem;
  margin-top: 6px;
  display: block;
}

/* ensure groups wrap correctly on small screens */
.group-wrap {
  width: 100%;
}

/* responsive */
@media (max-width: 576px) {
  .msg-bubble { max-width: 90%; }
}
</style>
