<script setup lang="ts">
import { computed } from 'vue';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { badgeEmojis } from '../lib/chatBadges';
import TwitchUserLink from './TwitchUserLink.vue';

const props = withDefaults(
  defineProps<{
    user: string;
    message: string;
    keyword: boolean;
    fromSent: boolean;
    badgeTags: ChatBadgeTag[];
    showTimestamp: boolean;
    createdAt?: string;
    chatterUserId?: number | null;
    /** Channel login for hl-streamer / hl-mine / hl-monitored (e.g. Watch view). */
    highlightChannel?: string;
    /** When true, show channel login (e.g. global/user message lists; not Watch chat). */
    showChannel?: boolean;
    channelLogin?: string;
    /** Highlight chatter as marked (dredge) */
    userMarked?: boolean;
    /** Highlight chatter as suspicious */
    userIsSus?: boolean;
    /** Tooltip when suspicious (e.g. reason from live updates) */
    suspiciousTitle?: string;
  }>(),
  {
    createdAt: undefined,
    chatterUserId: undefined,
    highlightChannel: '',
    showChannel: false,
    channelLogin: '',
    userMarked: false,
    userIsSus: false,
    suspiciousTitle: '',
  },
);

const badgeStr = computed(() => badgeEmojis(props.badgeTags));

const timeLabel = computed(() => {
  if (!props.showTimestamp || !props.createdAt) {
    return '';
  }
  const t = Date.parse(props.createdAt);
  if (!Number.isFinite(t)) {
    return '';
  }
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(t);
});
</script>

<template>
  <li
    :class="{
      kw: keyword,
      sent: fromSent && !keyword,
      marked: userMarked && !keyword,
    }"
  >
    <span v-if="timeLabel" class="ts">{{ timeLabel }}</span>
    <span v-if="showChannel && channelLogin" class="chan" title="Channel">#{{ channelLogin }}</span>
    <span v-if="badgeStr" class="badges" aria-hidden="true">{{ badgeStr }}</span>
    <TwitchUserLink
      :login="user"
      :user-twitch-id="chatterUserId"
      :highlight-channel="highlightChannel"
      :suspicious="userIsSus"
      :link-title="suspiciousTitle"
      variant="chat"
    />
    <span class="txt">{{ message }}</span>
  </li>
</template>

<style scoped lang="scss">
li {
  list-style: none;
  padding: 0.2rem 0.15rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);

  &.kw {
    background: rgba(145, 71, 255, 0.12);
  }

  &.sent:not(.kw) {
    background: rgba(0, 245, 147, 0.06);
  }

  &.marked:not(.kw) {
    background: rgba(255, 193, 7, 0.1);
    border-left: 2px solid rgba(255, 193, 7, 0.55);
  }
}

.ts {
  display: inline-block;
  min-width: 7.5rem;
  margin-right: 0.35rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
  vertical-align: baseline;
}

.chan {
  display: inline-block;
  margin-right: 0.35rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  font-weight: 600;
  min-width: 4.5rem;
}

.badges {
  margin-right: 0.12rem;
  font-size: 0.92em;
  vertical-align: middle;
}

.txt {
  word-break: break-word;
}
</style>
