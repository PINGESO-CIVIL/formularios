<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <main class="app">
    <RouterView />
  </main>
</template>

<style>
body {
  margin: 0;
  font-family: var(--forms-font);
  color: var(--forms-color-text);
  background: var(--forms-color-bg);
}

.app {
  padding: var(--forms-spacing);
  max-width: 64rem;
  margin: 0 auto;
}
</style>
