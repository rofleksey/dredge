<script setup lang="ts">
import { computed } from 'vue';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { badgeEmojis } from '../lib/chatBadges';
import { formatDateTime } from '../lib/dateTime';
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
    /** Twitch first message in channel (IRC) */
    firstMessage?: boolean;
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
    firstMessage: false,
  },
);

const badgeStr = computed(() => badgeEmojis(props.badgeTags));

const timeLabel = computed(() => {
  if (!props.showTimestamp || !props.createdAt) {
    return '';
  }
  return formatDateTime(props.createdAt);
});
</script>

<template>
  <li
    :class="{
      kw: keyword,
      sent: fromSent && !keyword,
      marked: userMarked && !keyword,
      'first-msg': firstMessage && !keyword,
    }"
  >
    <span v-if="timeLabel" class="ts">{{ timeLabel }}</span>
    <span v-if="showChannel && channelLogin" class="chan" title="Channel">#{{ channelLogin }}</span>
    <span v-if="firstMessage" class="first-msg-pill" title="First message in this channel">1st</span>
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

  &.first-msg:not(.kw) {
    background: rgba(56, 189, 248, 0.08);
    border-left: 2px solid rgba(56, 189, 248, 0.45);
  }
}

.first-msg-pill {
  display: inline-block;
  margin-right: 0.25rem;
  padding: 0.05rem 0.28rem;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--accent, #38bdf8);
  border: 1px solid rgba(56, 189, 248, 0.45);
  border-radius: 0.2rem;
  vertical-align: middle;
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
