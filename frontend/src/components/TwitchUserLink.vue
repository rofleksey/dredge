<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import { useTwitchUserHighlight } from '../composables/useTwitchUserHighlight';

const props = withDefaults(
  defineProps<{
    login: string;
    userTwitchId?: number | null;
    /** Normalized or raw channel login; used for streamer / monitored / mine highlights. */
    highlightChannel?: string;
    /** Chat line (default) vs system join/part line weight. */
    variant?: 'chat' | 'system';
  }>(),
  {
    userTwitchId: undefined,
    highlightChannel: '',
    variant: 'chat',
  },
);

const { highlightClass } = useTwitchUserHighlight(() => props.highlightChannel ?? '');

const extraClass = computed(() => highlightClass.value(props.login));
</script>

<template>
  <RouterLink
    v-if="userTwitchId != null && userTwitchId !== undefined"
    class="twitch-user-link"
    :class="[`twitch-user-link--${variant}`, extraClass]"
    :to="{ name: 'user', params: { id: String(userTwitchId) } }"
  >
    {{ login }}
  </RouterLink>
  <span
    v-else
    class="twitch-user-link"
    :class="[`twitch-user-link--${variant}`, extraClass]"
    >{{ login }}</span
  >
</template>

<style scoped lang="scss">
.twitch-user-link {
  color: var(--accent-bright);
  margin-right: 0.35rem;
  text-decoration: none;
  flex-shrink: 0;

  &--chat {
    font-weight: 600;
  }

  &--system {
    font-weight: 700;
  }

  &:hover {
    text-decoration: underline;
  }

  &.hl-streamer {
    color: #ffb74d;
  }

  &.hl-mine {
    color: #69f0ae;
  }

  &.hl-monitored {
    color: #64b5f6;
  }
}

span.twitch-user-link {
  cursor: default;

  &:hover {
    text-decoration: none;
  }
}
</style>
