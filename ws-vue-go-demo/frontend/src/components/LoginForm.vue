<template>
  <div class="card shadow-sm p-4">
    <h5 class="mb-3">Đăng nhập phòng chat</h5>
    <div class="mb-3">
      <label class="form-label">Mã phòng</label>
      <input type="text" class="form-control" v-model="roomID" placeholder="Nhập mã phòng" />
    </div>
    <div class="mb-3">
      <label class="form-label">Tên người dùng</label>
      <input type="text" class="form-control" v-model="username" placeholder="Nhập tên của bạn" />
    </div>
    <button class="btn btn-primary w-100" @click="submit">Vào phòng</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      roomID: '',
      username: ''
    }
  },
  methods: {
    submit() {
      if (this.roomID && this.username) {
        this.$emit('join', { username: this.username, roomID: this.roomID })
      }
    }
  }
}
</script>
