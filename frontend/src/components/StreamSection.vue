<template>
  <section class="stream-section">
    <div class="stream-container">
      <div class="stream-header">
        <h3>Twitch Stream</h3>
        <span v-if="currentChannel">#{{ currentChannel }}</span>
      </div>
      <div class="stream-iframe">
        <div v-if="!currentChannel" class="empty-state">
          <i class="material-icons">live_tv</i>
          <p>No channel selected</p>
        </div>
        <iframe
            v-else
            :src="streamUrl"
            width="100%"
            height="100%"
            allowfullscreen>
        </iframe>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  currentChannel: string | null
}>()

// @ts-ignore
const FRAME_PARENT = encodeURIComponent(process.env.NODE_ENV === 'development' ? "localhost" : window.location.hostname)

const streamUrl = computed(() => {
  if (!props.currentChannel) {
    return ''
  }

  return `https://player.twitch.tv/?channel=${props.currentChannel}&parent=${FRAME_PARENT}&muted=true`
})
</script>

<style scoped>
.stream-section {
  width: 100%;
  max-width: 400px;
}

@media (min-width: 768px) {
  .stream-section {
    min-width: 300px;
  }
}

.stream-container {
  background-color: var(--surface-color);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  height: fit-content;
}

.stream-header {
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stream-iframe {
  width: 100%;
  aspect-ratio: 16/9;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text-secondary);
  text-align: center;
}

.empty-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}
</style>