<script setup lang="ts">
function login() {
  axios.post('/api/logout').then((res) => {
    console.log(res.data)
    username.value = ''
    showLogin.value = true
  })
}
function logout() {
  axios.post('/api/logout').then((res) => {
    console.log(res.data)
    username.value = ''
    showLogin.value = true
  })
}
</script>

<template>
  <input type="text" v-model="username" />
  <button @click="login">登录</button>
  <button @click="logout">logout</button>
</template>
<style scoped></style>
