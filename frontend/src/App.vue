<template>
  <RouterView />
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<style>
:root {
  --primary-color: #9147ff;
  --primary-dark: #772ce8;
  --surface-color: #1f1f23;
  --background-color: #0e0e10;
  --text-primary: #efeff1;
  --text-secondary: #adadb8;
  --error-color: #eb0400;
  --success-color: #00d95f;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Roboto', sans-serif;
  background-color: var(--background-color);
  color: var(--text-primary);
  line-height: 1.5;
}

#app {
  min-height: 100vh;
}
</style>