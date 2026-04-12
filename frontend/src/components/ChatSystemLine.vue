<script setup lang="ts">
import TwitchUserLink from './TwitchUserLink.vue';

const props = withDefaults(
  defineProps<{
    variant: 'join' | 'part' | 'gap';
    user: string;
    text: string;
    /** Secondary line (e.g. account date, or gap duration label) */
    detail?: string;
    chatterUserId?: number | null;
    highlightChannel?: string;
  }>(),
  {
    detail: '',
    chatterUserId: undefined,
    highlightChannel: '',
  },
);
</script>

<template>
  <li
    class="sys-line"
    :class="{
      'sys-line--join': variant === 'join',
      'sys-line--part': variant === 'part',
      'sys-line--gap': variant === 'gap',
    }"
  >
    <template v-if="variant === 'gap'">
      <span class="sys-gap-text">{{ text }}</span>
    </template>
    <template v-else>
      <span class="sys-badge" aria-hidden="true">{{ variant === 'join' ? '→' : '←' }}</span>
      <TwitchUserLink
        :login="user"
        :user-twitch-id="chatterUserId"
        :highlight-channel="highlightChannel"
        variant="system"
      />
      <span class="sys-msg">{{ text }}</span>
      <span v-if="detail" class="sys-detail">{{ detail }}</span>
    </template>
  </li>
</template>

<style scoped lang="scss">
.sys-line {
  list-style: none;
  margin: 0.35rem 0.15rem;
  padding: 0.35rem 0.5rem;
  border-radius: 0.3rem;
  font-size: 0.78rem;
  line-height: 1.35;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.35rem 0.5rem;
}

.sys-line--gap {
  justify-content: center;
  margin: 0.5rem 0.25rem;
  padding: 0.5rem 0.65rem;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(145, 71, 255, 0.12) 15%,
    rgba(145, 71, 255, 0.12) 85%,
    transparent 100%
  );
  border: 1px solid rgba(145, 71, 255, 0.22);
  border-radius: 0.35rem;
}

.sys-gap-text {
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

.sys-line--join {
  background: rgba(0, 245, 147, 0.06);
  border-left: 3px solid rgba(0, 245, 147, 0.45);
}

.sys-line--part {
  background: rgba(100, 181, 246, 0.07);
  border-left: 3px solid rgba(100, 181, 246, 0.45);
}

.sys-badge {
  font-size: 0.85rem;
  opacity: 0.85;
  flex-shrink: 0;
}

.sys-msg {
  color: var(--text);
  opacity: 0.92;
}

.sys-detail {
  width: 100%;
  flex-basis: 100%;
  margin-left: 1.5rem;
  font-size: 0.72rem;
  color: var(--text-muted);
}
</style>
