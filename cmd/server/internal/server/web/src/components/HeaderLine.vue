<script setup lang="ts">
defineProps<{
  username: string
}>()
</script>
<template>
  <div>{{ username }}</div>
</template>
