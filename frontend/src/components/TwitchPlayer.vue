<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  channel: string;
}>();

const src = computed(() => {
  const ch = props.channel.replace(/^#/, '').trim();
  if (!ch) {
    return '';
  }
  const q = new URLSearchParams({ channel: ch });
  for (const p of new Set([window.location.hostname, 'localhost', '127.0.0.1'])) {
    q.append('parent', p);
  }
  return `https://player.twitch.tv/?${q.toString()}`;
});
</script>

<template>
  <div class="player-wrap">
    <iframe
      v-if="src"
      class="player"
      :src="src"
      allowfullscreen
      allow="autoplay; fullscreen"
      title="Twitch stream"
    />
    <div v-else class="placeholder">Pick a channel to load the player.</div>
  </div>
</template>

<style scoped lang="scss">
.player-wrap {
  position: relative;
  width: min(100%, calc((100dvh - 11rem) * 16 / 9));
  max-width: 100%;
  max-height: calc(100dvh - 11rem);
  margin: 0 auto;
  aspect-ratio: 16 / 9;
  background: #000;
  border-radius: 0.35rem;
  overflow: hidden;
}

.player {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  border: 0;
}

.placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 0.9rem;
}
</style>
