<script setup lang="ts">
import type { ChannelLive } from '../../api/generated';

defineProps<{
  loading: boolean;
  onlineMonitored: ChannelLive[];
  offlineMonitored: ChannelLive[];
  monitoredSidebar: ChannelLive[];
  selectedChannel: string;
  normCh: (c: string) => string;
}>();

const emit = defineEmits<{
  select: [login: string];
  'open-channel': [];
}>();
</script>

<template>
  <aside class="follows-sidebar" aria-label="Monitored channels">
    <button
      type="button"
      class="follow-tile follow-tile--add"
      title="Go to channel…"
      aria-label="Go to channel by name"
      @click="emit('open-channel')"
    >
      +
    </button>
    <template v-if="loading">
      <span class="muted muted--compact tiny follows-hint">…</span>
    </template>
    <template v-else>
      <button
        v-for="f in onlineMonitored"
        :key="f.broadcaster_id"
        type="button"
        class="follow-tile"
        :class="{ 'follow-tile--active': normCh(selectedChannel) === normCh(f.broadcaster_login) }"
        :title="`#${f.broadcaster_login}`"
        :aria-label="`Open channel ${f.broadcaster_login}`"
        @click="emit('select', f.broadcaster_login)"
      >
        <img
          v-if="f.profile_image_url"
          class="follow-avatar"
          :src="f.profile_image_url"
          :alt="''"
          width="40"
          height="40"
        />
        <span v-else class="follow-initial" aria-hidden="true">{{ f.broadcaster_login.charAt(0).toUpperCase() }}</span>
      </button>
      <button
        v-for="f in offlineMonitored"
        :key="'off-' + f.broadcaster_id"
        type="button"
        class="follow-tile follow-tile--offline"
        :class="{ 'follow-tile--active': normCh(selectedChannel) === normCh(f.broadcaster_login) }"
        :title="`#${f.broadcaster_login} (offline)`"
        :aria-label="`Open channel ${f.broadcaster_login}`"
        @click="emit('select', f.broadcaster_login)"
      >
        <img
          v-if="f.profile_image_url"
          class="follow-avatar"
          :src="f.profile_image_url"
          :alt="''"
          width="40"
          height="40"
        />
        <span v-else class="follow-initial" aria-hidden="true">{{ f.broadcaster_login.charAt(0).toUpperCase() }}</span>
      </button>
      <p v-if="!monitoredSidebar.length" class="muted muted--compact tiny follows-hint">No monitored channels</p>
    </template>
  </aside>
</template>

<style scoped lang="scss">
.follows-sidebar {
  flex: 0 0 auto;
  width: 3.35rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.45rem;
  padding: 0.4rem 0.35rem;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.follows-hint {
  text-align: center;
  line-height: 1.2;
  max-width: 100%;
}

.follow-tile {
  flex-shrink: 0;
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
  border: 2px solid transparent;
  border-radius: 50%;
  background: var(--bg-base);
  cursor: pointer;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text);

  &:hover {
    background: var(--bg-hover);
  }

  &--active {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px rgba(145, 70, 255, 0.35);
  }

  &--add {
    border-radius: 0.35rem;
    border: 1px dashed var(--border);
    font-size: 1.35rem;
    font-weight: 600;
    line-height: 1;
    color: var(--accent-bright);

    &:hover {
      border-color: var(--accent);
      color: var(--accent);
    }
  }

  &--offline {
    filter: grayscale(1);
    opacity: 0.88;

    .follow-initial {
      color: var(--text-muted);
    }

    &:hover {
      opacity: 1;
    }
  }
}

.follow-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.follow-initial {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--accent-bright);
}

.tiny {
  font-size: 0.72rem;
}

@media (max-width: 639px) {
  .follows-sidebar {
    flex-direction: row;
    flex-wrap: nowrap;
    width: 100%;
    max-height: 3.6rem;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 0.35rem 0.45rem;
    align-items: center;
  }

  .follows-hint {
    flex-shrink: 0;
  }
}
</style>
