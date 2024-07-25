<script setup lang="ts">
import { ref } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import TheWelcome from './components/TheWelcome.vue'
import HeaderLine from './components/HeaderLine.vue'
import axios from 'axios'
var username = ref('')
var showLogin = ref(true)
axios
  .get('/api/me')
  .then((res) => {
    username.value = res.data.user
    showLogin.value = false
  })
  .catch((err) => {
    if (err.response.status === 401) {
      showLogin.value = true
    }
  })
axios.post('/api/login', { username: 'admin' }).then((res) => {
  console.log(res.data)
  axios.get('/api/me').then((res) => {
    console.log(res.data)
  })
})
function logout() {
  axios.post('/api/logout').then((res) => {
    console.log(res.data)
    username.value = ''
    showLogin.value = true
  })
}
</script>

<template>
  <header>
    <!-- <img alt="Vue logo" class="logo" src="./assets/logo.svg" width="125" height="125" /> -->
    <!-- 
    <div class="wrapper">
      <HelloWorld msg="You did it!@!" />
    </div> -->
    <HeaderLine :username="username"></HeaderLine>
    <button @click="logout">logout</button>
  </header>
  <main>
    <div v-if="showLogin">abc</div>
    <div v-else>def</div>
    <TheWelcome />
  </main>
</template>

<style scoped>
header {
  line-height: 1.5;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>
