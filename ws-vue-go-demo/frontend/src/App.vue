<template>
  <div class="container py-4">
    <h2 class="text-center mb-4">💬 Chat Room</h2>
    <LoginForm v-if="!joined" @join="joinRoom" />
    <ChatRoom v-else :username="username" :roomID="roomID" />
  </div>
</template>

<script>
import LoginForm from './components/LoginForm.vue'
import ChatRoom from './components/ChatRoom.vue'

export default {
  components: { LoginForm, ChatRoom },
  data() {
    return {
      joined: false,
      username: '',
      roomID: ''
    }
  },
  methods: {
    joinRoom({ username, roomID }) {
      this.username = username
      this.roomID = roomID
      this.joined = true
    }
  }
}
</script>
