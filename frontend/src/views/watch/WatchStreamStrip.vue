<script setup lang="ts">
import { RouterLink } from 'vue-router';
import type { ChannelLive } from '../../api/generated';

defineProps<{
  channelLive: ChannelLive | null;
  loadingChannelMeta: boolean;
  selectedChannel: string;
  formatSessionClock: (iso?: string | null) => string;
  formatViewerDisplay: (live: Pick<ChannelLive, 'viewer_count' | 'channel_chatter_count'>) => string;
}>();

const emit = defineEmits<{
  'open-viewers': [];
}>();
</script>

<template>
  <header v-if="channelLive" class="stream-strip compact">
    <RouterLink
      class="stream-user-link stream-user-link--avatar"
      :to="{ name: 'user', params: { id: String(channelLive.broadcaster_id) } }"
      :aria-label="`Open user ${channelLive.display_name}`"
    >
      <img
        class="avatar"
        :src="channelLive.profile_image_url"
        :alt="''"
        width="36"
        height="36"
      />
    </RouterLink>
    <div class="stream-meta">
      <div class="stream-title-row">
        <RouterLink
          class="dn stream-user-link"
          :to="{ name: 'user', params: { id: String(channelLive.broadcaster_id) } }"
        >
          {{ channelLive.display_name }}
        </RouterLink>
        <span v-if="channelLive.is_live" class="live-pill">LIVE</span>
        <span v-else class="off-pill">Offline</span>
      </div>
      <p v-if="channelLive.is_live && channelLive.title" class="game-line">{{ channelLive.title }}</p>
      <p v-if="channelLive.is_live && channelLive.game_name" class="game-line">{{ channelLive.game_name }}</p>
      <div class="stream-stats">
        <span v-if="channelLive.is_live" class="uptime">Session {{ formatSessionClock(channelLive.started_at) }}</span>
        <button
          v-if="channelLive.is_live"
          type="button"
          class="viewers-btn"
          @click="emit('open-viewers')"
        >
          {{ formatViewerDisplay(channelLive) }} viewers
        </button>
      </div>
    </div>
  </header>
  <header v-else-if="selectedChannel && loadingChannelMeta" class="stream-strip compact placeholder">
    <span class="muted muted--compact">Loading channel…</span>
  </header>
</template>

<style scoped lang="scss">
.stream-strip {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  padding: 0.5rem 0.65rem;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  flex-shrink: 0;

  &.compact {
    gap: 0.45rem;
    padding: 0.35rem 0.5rem;
    margin-top: 0.45rem;

    .game-line {
      font-size: 0.72rem;
      margin: 0.08rem 0 0;
    }

    .stream-stats {
      margin-top: 0.2rem;
      font-size: 0.72rem;
    }
  }

  &.placeholder {
    align-items: center;
    min-height: 2.25rem;
  }
}

.avatar {
  border-radius: 0.35rem;
  flex-shrink: 0;
}

.stream-meta {
  flex: 1;
  min-width: 0;
}

.stream-title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}

.dn {
  font-weight: 700;
  font-size: 0.95rem;
}

.stream-user-link {
  color: inherit;
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }

  &--avatar {
    line-height: 0;

    &:hover {
      text-decoration: none;
    }
  }
}

.stream-strip.compact .dn {
  font-size: 0.85rem;
}

.live-pill {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.12rem 0.35rem;
  border-radius: 0.2rem;
  background: #e53935;
  color: #fff;
}

.off-pill {
  font-size: 0.65rem;
  color: var(--text-muted);
}

.game-line {
  margin: 0.15rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.35;
}

.stream-stats {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.35rem;
  font-size: 0.78rem;
}

.uptime {
  color: var(--text-muted);
}

.viewers-btn {
  padding: 0.2rem 0.45rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-base);
  color: var(--accent-bright);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
  }
}

</style>
