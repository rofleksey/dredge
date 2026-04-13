<script setup lang="ts">
import { useIntervalFn, useScroll } from '@vueuse/core';
import { storeToRefs } from 'pinia';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import AppModal from '../components/AppModal.vue';
import SubmitButton from '../components/SubmitButton.vue';
import TwitchPlayer from '../components/TwitchPlayer.vue';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import ChatSystemLine from '../components/ChatSystemLine.vue';
import TwitchUserLink from '../components/TwitchUserLink.vue';
import { ApiError, ChatHistoryEntry, DefaultService } from '../api/generated';
import type { ListChannelChattersRequest } from '../api/generated';
import type { ChannelChatterEntry, ChannelLive, IrcMonitorStatus, WatchUiHints } from '../api/generated';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { isChannelJoinedOnIrc } from '../lib/ircMonitorJoined';
import { effectiveChatterIsSus, effectiveSuspicionTitle } from '../lib/suspicionOverlay';
import { notifyFromApiError } from '../lib/clientNotice';
import { notify } from '../lib/notify';
import { useChannelsStore } from '../stores/channels';
import { useLiveSocketStore } from '../stores/liveSocket';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';
import { useWatchPreferencesStore } from '../stores/watchPreferences';

defineOptions({ name: 'WatchView' });

const channelsStore = useChannelsStore();
const twitchStore = useTwitchAccountsStore();

const liveSocket = useLiveSocketStore();
const { events } = storeToRefs(liveSocket);

const watchPrefs = useWatchPreferencesStore();
const { chatGapMinutes } = storeToRefs(watchPrefs);

const selectedChannel = ref('');
const sendAccountId = ref<number | null>(null);
const sendText = ref('');
const chatEl = ref<HTMLElement | null>(null);
const historyEntries = ref<ChatHistoryEntry[]>([]);
const channelSwitchTime = ref(0);

const channelLive = ref<ChannelLive | null>(null);
const monitoredSidebar = ref<ChannelLive[]>([]);
const loadingChannelMeta = ref(false);
const loadingMonitoredSidebar = ref(false);
const viewerModalOpen = ref(false);
const viewerChatters = ref<ChannelChatterEntry[]>([]);
const loadingViewerChatters = ref(false);
const viewerFilterQuery = ref('');
type ViewerSortMode =
  | 'present_new'
  | 'present_old'
  | 'login_az'
  | 'login_za'
  | 'account_new'
  | 'account_old'
  | 'message_high'
  | 'message_low';
const viewerSort = ref<ViewerSortMode>('present_new');
const watchHints = ref<WatchUiHints | null>(null);
const channelEntryModalOpen = ref(false);
const manualChannelInput = ref('');
const sendingChat = ref(false);

/** Advances every second so session uptime and related labels stay live. */
const sessionClock = ref(Date.now());
useIntervalFn(() => {
  sessionClock.value = Date.now();
}, 1000);

const twitchLoginRe = /^[a-zA-Z0-9_]{4,25}$/;

useScroll(chatEl);

type ChatLine = {
  key: string;
  user: string;
  message: string;
  keyword: boolean;
  userMarked: boolean;
  userIsSus: boolean;
  susTitle?: string;
  fromSent: boolean;
  firstMessage: boolean;
  at: number;
  badgeTags: ChatBadgeTag[];
  createdAtIso?: string;
  chatterUserId?: number | null;
};

type ChatRow =
  | { kind: 'gap'; key: string; label: string }
  | { kind: 'msg'; line: ChatLine };

function normCh(c: string): string {
  return c.replace(/^#/, '').toLowerCase();
}

/** Rejects error-shaped JSON bodies so we do not keep a stale LIVE strip when the API returns `{ message }`. */
function parseChannelLivePayload(data: unknown): ChannelLive | null {
  if (!data || typeof data !== 'object') {
    return null;
  }
  const o = data as Record<string, unknown>;
  if (typeof o.broadcaster_login !== 'string') {
    return null;
  }
  return data as ChannelLive;
}

/** Prefer last opened channel from local storage when it is still a monitored channel or a valid manual login. */
function preferredWatchChannel(monitored: { username: string }[]): string {
  const last = watchPrefs.getLastWatchChannel();
  if (last) {
    const fromList = monitored.find((c) => normCh(c.username) === last);
    if (fromList) {
      return fromList.username;
    }
    if (twitchLoginRe.test(last)) {
      return last;
    }
  }
  return monitored[0]?.username ?? '';
}

const ircStatusForWatchDot = ref<IrcMonitorStatus | null>(null);

async function pullIrcStatusForWatchDot(): Promise<void> {
  try {
    ircStatusForWatchDot.value = await DefaultService.getIrcMonitorStatus();
  } catch {
    ircStatusForWatchDot.value = null;
  }
}

const ircChatConnected = computed(() => isChannelJoinedOnIrc(ircStatusForWatchDot.value, selectedChannel.value));

watch(
  () => normCh(selectedChannel.value),
  () => {
    void pullIrcStatusForWatchDot();
  },
  { immediate: true },
);

useIntervalFn(
  () => {
    if (normCh(selectedChannel.value)) {
      void pullIrcStatusForWatchDot();
    }
  },
  4000,
);

function formatGapMinutes(fromMs: number, toMs: number): string {
  const mins = Math.round((toMs - fromMs) / 60000);
  if (mins <= 0) {
    return '— pause —';
  }
  if (mins >= 60) {
    const h = Math.floor(mins / 60);
    const m = mins % 60;
    return m ? `— ${h}h ${m}m —` : `— ${h}h —`;
  }
  return `— ${mins} min —`;
}

/** Formats elapsed time as HH:MM:SS (live via sessionClock). */
function formatSessionClock(iso?: string | null): string {
  if (!iso) {
    return '—';
  }
  const t = Date.parse(iso);
  if (!Number.isFinite(t)) {
    return '—';
  }
  let sec = Math.floor((sessionClock.value - t) / 1000);
  if (sec < 0) {
    sec = 0;
  }
  return formatSecondsAsClock(sec);
}

function formatSecondsAsClock(sec: number): string {
  const s = Math.max(0, Math.floor(sec));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const r = s % 60;
  return [h, m, r].map((n) => String(n).padStart(2, '0')).join(':');
}

function formatAccountDate(iso?: string | null): string {
  if (!iso) {
    return '';
  }
  const t = Date.parse(iso);
  if (!Number.isFinite(t)) {
    return '';
  }
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(t);
}

function formatPresentElapsed(iso: string): string {
  const t = Date.parse(iso);
  if (!Number.isFinite(t)) {
    return '—';
  }
  const sec = Math.max(0, Math.floor((sessionClock.value - t) / 1000));
  return formatSecondsAsClock(sec);
}

/** Helix viewer count with delta vs IRC chatter snapshot: e.g. 4 (+3) or 4 (-2). */
function formatViewerDisplay(live: Pick<ChannelLive, 'viewer_count' | 'channel_chatter_count'>): string {
  const v = live.viewer_count;
  if (v == null) {
    return '—';
  }
  const c = live.channel_chatter_count;
  if (c == null || c === undefined) {
    return String(v);
  }
  const d = c - v;
  if (d === 0) {
    return `${v} (0)`;
  }
  const sign = d > 0 ? '+' : '';
  return `${v} (${sign}${d})`;
}

const viewerPollMs = computed(() =>
  Math.max(1000, (watchHints.value?.viewer_poll_interval_seconds ?? 10) * 1000),
);
const chattersSyncMs = computed(() =>
  Math.max(1000, (watchHints.value?.channel_chatters_sync_interval_seconds ?? 10) * 1000),
);

const onlineMonitored = computed(() =>
  [...monitoredSidebar.value.filter((f) => f.is_live)].sort((a, b) =>
    a.broadcaster_login.localeCompare(b.broadcaster_login),
  ),
);

const offlineMonitored = computed(() =>
  [...monitoredSidebar.value.filter((f) => !f.is_live)].sort((a, b) =>
    a.broadcaster_login.localeCompare(b.broadcaster_login),
  ),
);

const displayedViewerChatters = computed((): ChannelChatterEntry[] => {
  let rows = [...viewerChatters.value];
  const q = viewerFilterQuery.value.trim().toLowerCase();
  if (q) {
    rows = rows.filter((c) => c.login.toLowerCase().includes(q));
  }
  const presentMs = (iso: string): number => {
    const t = Date.parse(iso);
    return Number.isFinite(t) ? t : 0;
  };
  const accountMs = (iso?: string | null): number => {
    if (!iso) {
      return NaN;
    }
    const t = Date.parse(iso);
    return Number.isFinite(t) ? t : NaN;
  };
  rows.sort((a, b) => {
    const msgCount = (c: ChannelChatterEntry): number => c.message_count ?? 0;
    switch (viewerSort.value) {
      case 'present_new':
        return presentMs(b.present_since) - presentMs(a.present_since);
      case 'present_old':
        return presentMs(a.present_since) - presentMs(b.present_since);
      case 'message_high':
        return msgCount(b) - msgCount(a);
      case 'message_low':
        return msgCount(a) - msgCount(b);
      case 'login_az':
        return a.login.localeCompare(b.login);
      case 'login_za':
        return b.login.localeCompare(a.login);
      case 'account_new': {
        const tb = accountMs(b.account_created_at);
        const ta = accountMs(a.account_created_at);
        const hb = Number.isFinite(tb);
        const ha = Number.isFinite(ta);
        if (ha && hb) {
          return tb - ta;
        }
        if (ha !== hb) {
          return ha ? -1 : 1;
        }
        return presentMs(b.present_since) - presentMs(a.present_since);
      }
      case 'account_old': {
        const tb = accountMs(b.account_created_at);
        const ta = accountMs(a.account_created_at);
        const hb = Number.isFinite(tb);
        const ha = Number.isFinite(ta);
        if (ha && hb) {
          return ta - tb;
        }
        if (ha !== hb) {
          return ha ? -1 : 1;
        }
        return presentMs(a.present_since) - presentMs(b.present_since);
      }
      default:
        return 0;
    }
  });
  return rows;
});

async function refreshMonitoredSidebar(silent = false): Promise<void> {
  const list = channelsStore.monitoredChannels;
  if (!list.length) {
    monitoredSidebar.value = [];
    return;
  }
  if (!silent) {
    loadingMonitoredSidebar.value = true;
  }
  try {
    const results = await Promise.all(
      list.map((c) =>
        DefaultService.getChannelLive({ requestBody: { login: normCh(c.username) } }).catch(() => null),
      ),
    );
    monitoredSidebar.value = list.map((c, i) => {
      const live = results[i];
      const parsed = parseChannelLivePayload(live);
      if (parsed) {
        return parsed;
      }
      return {
        broadcaster_id: c.id,
        broadcaster_login: normCh(c.username),
        display_name: c.username,
        profile_image_url: '',
        is_live: false,
      };
    });
  } finally {
    if (!silent) {
      loadingMonitoredSidebar.value = false;
    }
  }
}

watch(
  () => channelsStore.monitoredChannels,
  () => {
    void refreshMonitoredSidebar();
  },
  { deep: true, immediate: true },
);

async function pollChannelLive(): Promise<void> {
  const login = normCh(selectedChannel.value);
  if (!login) {
    return;
  }
  try {
    const live = await DefaultService.getChannelLive({ requestBody: { login } });
    const parsed = parseChannelLivePayload(live);
    if (parsed) {
      channelLive.value = parsed;
    } else {
      channelLive.value = null;
    }
  } catch {
    /* keep previous channelLive */
  }
}

useIntervalFn(
  () => {
    void pollChannelLive();
    void refreshMonitoredSidebar(true);
  },
  viewerPollMs,
);

const displayLines = computed((): ChatLine[] => {
  const ch = normCh(selectedChannel.value);
  if (!ch) {
    return [];
  }

  const ov = liveSocket.suspicionByTwitchId;

  const hist: ChatLine[] = historyEntries.value.map((h) => {
    const t = Date.parse(h.created_at);
    const at = Number.isFinite(t) ? t : channelSwitchTime.value;
    const badgeTags = [...(h.badge_tags ?? [])] as ChatBadgeTag[];
    const uid = h.chatter_user_id ?? undefined;
    const userIsSus = effectiveChatterIsSus(uid, h.chatter_is_sus, ov);
    return {
      key: `db-${h.id}`,
      user: h.user,
      message: h.message,
      keyword: h.keyword_match,
      userMarked: h.chatter_marked,
      userIsSus,
      susTitle: effectiveSuspicionTitle(uid, userIsSus, ov),
      fromSent: h.source === ChatHistoryEntry.source.SENT,
      firstMessage: Boolean(h.first_message),
      at,
      badgeTags,
      createdAtIso: h.created_at,
      chatterUserId: uid,
    };
  });

  const since = channelSwitchTime.value;
  const live: ChatLine[] = [];

  for (let i = 0; i < events.value.length; i++) {
    const e = events.value[i];
    if (normCh(e.channel) !== ch || e.receivedAt < since) {
      continue;
    }

    if (e.type === 'chat_message' || e.type === 'message_sent') {
      const fromIso = e.created_at ? Date.parse(e.created_at) : NaN;
      const at = Number.isFinite(fromIso) ? fromIso : e.receivedAt;
      const badgeTags = (e.badge_tags ?? []) as ChatBadgeTag[];
      const uid = e.user_twitch_id;
      const fromPayload = e.type === 'chat_message' && Boolean(e.chatter_is_sus);
      const userIsSus =
        e.type === 'chat_message' ? effectiveChatterIsSus(uid, fromPayload, ov) : false;
      live.push({
        key: `ws-${e.receivedAt}-${i}-${e.type}`,
        user: e.user,
        message: e.message,
        keyword: e.type === 'chat_message' && Boolean(e.keyword_match),
        userMarked: e.type === 'chat_message' && Boolean(e.chatter_marked),
        userIsSus,
        susTitle:
          e.type === 'chat_message' ? effectiveSuspicionTitle(uid, userIsSus, ov) : undefined,
        fromSent: e.type === 'message_sent',
        firstMessage: e.type === 'chat_message' && Boolean(e.first_message),
        at,
        badgeTags,
        createdAtIso: e.created_at,
        chatterUserId: uid,
      });
    }
  }

  const merged = [...hist, ...live];
  merged.sort((a, b) => a.at - b.at);
  return merged;
});

const displayRows = computed((): ChatRow[] => {
  const lines = displayLines.value;
  if (lines.length === 0) {
    return [];
  }
  const gapMs = Math.max(1, chatGapMinutes.value) * 60_000;
  const rows: ChatRow[] = [];
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (i > 0 && line.at - lines[i - 1].at > gapMs) {
      rows.push({
        kind: 'gap',
        key: `gap-before-${line.key}`,
        label: formatGapMinutes(lines[i - 1].at, line.at),
      });
    }
    rows.push({ kind: 'msg', line });
  }
  return rows;
});

onMounted(async () => {
  try {
    await Promise.all([channelsStore.fetch(), twitchStore.fetch()]);
    if (twitchStore.accounts.length) {
      sendAccountId.value = twitchStore.accounts[0].id;
    }
  } catch {
    notify({
      id: 'watch-load',
      type: 'error',
      title: 'Watch',
      description: 'Could not load channels or Twitch accounts.',
    });
  }
  const pick = preferredWatchChannel(channelsStore.monitoredChannels);
  if (pick) {
    selectedChannel.value = pick;
  }
  try {
    watchHints.value = await DefaultService.getWatchUiHints();
  } catch {
    watchHints.value = null;
  }
});

watch(
  () => channelsStore.monitoredChannels,
  (list) => {
    if (selectedChannel.value || !list.length) {
      return;
    }
    const next = preferredWatchChannel(list);
    if (next) {
      selectedChannel.value = next;
    }
  },
  { deep: true },
);

watch(selectedChannel, (ch) => {
  const n = normCh(ch);
  if (n) {
    watchPrefs.setLastWatchChannel(n);
  }
});

watch(channelEntryModalOpen, (open) => {
  if (open) {
    manualChannelInput.value = selectedChannel.value ? normCh(selectedChannel.value) : '';
  }
});

watch(
  () => twitchStore.accounts,
  (list) => {
    if (!sendAccountId.value || !list.some((a) => a.id === sendAccountId.value)) {
      sendAccountId.value = list[0]?.id ?? null;
    }
  },
  { deep: true },
);

watch(
  selectedChannel,
  async (ch) => {
    if (!ch) {
      historyEntries.value = [];
      return;
    }
    const norm = normCh(ch);
    try {
      historyEntries.value = await DefaultService.listChatHistory({ channel: norm, limit: 80 });
    } catch (e) {
      historyEntries.value = [];
      if (e instanceof ApiError && e.status === 404) {
        /* channel not monitored — leave empty */
      }
    }
    channelSwitchTime.value = Date.now();
  },
  { flush: 'post' },
);

watch(
  selectedChannel,
  async (ch) => {
    channelLive.value = null;
    if (!ch) {
      return;
    }
    const login = normCh(ch);
    loadingChannelMeta.value = true;
    try {
      const live = await DefaultService.getChannelLive({ requestBody: { login } });
      channelLive.value = parseChannelLivePayload(live);
    } catch {
      channelLive.value = null;
    } finally {
      loadingChannelMeta.value = false;
    }
  },
  { flush: 'post' },
);

async function loadViewerChatters(silent: boolean): Promise<void> {
  const ch = normCh(selectedChannel.value);
  if (!ch || sendAccountId.value == null) {
    return;
  }
  if (!silent) {
    loadingViewerChatters.value = true;
  }
  try {
    const body: ListChannelChattersRequest = {
      account_id: sendAccountId.value,
      login: ch,
    };
    if (channelLive.value?.is_live && channelLive.value.started_at) {
      body.session_started_at = channelLive.value.started_at;
    }
    viewerChatters.value = await DefaultService.listChannelChatters({
      requestBody: body,
    });
  } catch (e) {
    if (!silent) {
      viewerChatters.value = [];
      notifyFromApiError(e, {
        id: 'watch-viewers-chatters',
        title: 'Viewers',
        fallbackDescription: 'Could not load chatter list (needs moderator:read:chatters).',
      });
    }
  } finally {
    if (!silent) {
      loadingViewerChatters.value = false;
    }
  }
}

const { pause: pauseViewerModalPoll, resume: resumeViewerModalPoll } = useIntervalFn(
  () => {
    void pollChannelLive();
    void loadViewerChatters(true);
  },
  chattersSyncMs,
  { immediate: false },
);

watch(viewerModalOpen, (open) => {
  if (!open) {
    pauseViewerModalPoll();
    viewerChatters.value = [];
    viewerFilterQuery.value = '';
    viewerSort.value = 'present_new';
    return;
  }
  void loadViewerChatters(false);
  resumeViewerModalPoll();
});

const chatScrollSig = computed(() => {
  const lines = displayLines.value;
  if (!lines.length) {
    return '';
  }
  return `${lines.length}:${lines[lines.length - 1].key}`;
});

watch(chatScrollSig, async () => {
  await nextTick();
  const el = chatEl.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
});

function selectChannel(raw: string): void {
  const login = normCh(raw);
  if (login) {
    selectedChannel.value = login;
  }
}

function applyManualChannel(): void {
  const login = normCh(manualChannelInput.value);
  if (!login || !twitchLoginRe.test(login)) {
    notify({
      id: 'watch-channel',
      type: 'error',
      title: 'Channel',
      description: 'Enter a valid Twitch username (4–25 letters, numbers, or underscores).',
    });
    return;
  }
  selectedChannel.value = login;
  channelEntryModalOpen.value = false;
}

function onComposerKeydown(e: KeyboardEvent): void {
  if (e.key !== 'Enter' || e.shiftKey) {
    return;
  }
  e.preventDefault();
  void sendChat();
}

async function sendChat(): Promise<void> {
  if (sendingChat.value) {
    return;
  }
  const ch = normCh(selectedChannel.value);
  if (!sendAccountId.value || !ch || !sendText.value.trim()) {
    notify({
      id: 'watch-send',
      type: 'error',
      title: 'Chat',
      description: 'Pick an account and enter a message.',
    });
    return;
  }
  sendingChat.value = true;
  try {
    await DefaultService.sendMessage({
      requestBody: {
        account_id: sendAccountId.value,
        channel: ch,
        message: sendText.value.trim(),
      },
    });
    sendText.value = '';
    notify({
      id: 'watch-send',
      type: 'success',
      title: 'Chat',
      description: 'Message sent.',
    });
  } catch (e) {
    notifyFromApiError(e, {
      id: 'watch-send',
      title: 'Chat',
      fallbackDescription: 'Send failed. Check account token and channel name.',
    });
  } finally {
    sendingChat.value = false;
  }
}
</script>

<template>
  <div class="watch">
    <div class="watch-layout">
      <aside class="follows-sidebar" aria-label="Monitored channels">
        <button
          type="button"
          class="follow-tile follow-tile--add"
          title="Go to channel…"
          aria-label="Go to channel by name"
          @click="channelEntryModalOpen = true"
        >
          +
        </button>
        <template v-if="loadingMonitoredSidebar">
          <span class="muted tiny follows-hint">…</span>
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
            @click="selectChannel(f.broadcaster_login)"
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
            @click="selectChannel(f.broadcaster_login)"
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
          <p v-if="!monitoredSidebar.length" class="muted tiny follows-hint">No monitored channels</p>
        </template>
      </aside>

      <div class="watch-main">
        <div class="grid">
          <section class="video">
            <TwitchPlayer :channel="selectedChannel" />

            <header v-if="channelLive" class="stream-strip compact">
              <img
                class="avatar"
                :src="channelLive.profile_image_url"
                :alt="''"
                width="36"
                height="36"
              />
              <div class="stream-meta">
                <div class="stream-title-row">
                  <span class="dn">{{ channelLive.display_name }}</span>
                  <span v-if="channelLive.is_live" class="live-pill">LIVE</span>
                  <span v-else class="off-pill">Offline</span>
                </div>
                <p v-if="channelLive.is_live && channelLive.title" class="game-line">{{ channelLive.title }}</p>
                <p v-if="channelLive.is_live && channelLive.game_name" class="game-line">{{ channelLive.game_name }}</p>
                <div class="stream-stats">
                  <span v-if="channelLive.is_live" class="uptime"
                    >Session {{ formatSessionClock(channelLive.started_at) }}</span
                  >
                  <button
                    v-if="channelLive.is_live"
                    type="button"
                    class="viewers-btn"
                    @click="viewerModalOpen = true"
                  >
                    {{ formatViewerDisplay(channelLive) }} viewers
                  </button>
                </div>
              </div>
            </header>
            <header v-else-if="selectedChannel && loadingChannelMeta" class="stream-strip compact placeholder">
              <span class="muted">Loading channel…</span>
            </header>
          </section>

          <section class="chat">
        <div class="pane-head chat-head">
          <div class="chat-head-main">
            <span class="chat-title-main">Chat</span>
            <span v-if="selectedChannel" class="chat-channel-tag">#{{ normCh(selectedChannel) }}</span>
            <span v-else class="chat-channel-empty muted">—</span>
            <span
              v-if="selectedChannel"
              class="irc-link-dot"
              role="img"
              :class="{ 'irc-link-dot--on': ircChatConnected, 'irc-link-dot--off': !ircChatConnected }"
              :title="ircChatConnected ? 'IRC monitor joined this channel' : 'IRC monitor not in this channel'"
              :aria-label="
                ircChatConnected ? 'IRC monitor joined this channel' : 'IRC monitor not in this channel'
              "
            />
          </div>
        </div>
        <ul ref="chatEl" class="lines">
          <template v-for="row in displayRows" :key="row.kind === 'gap' ? row.key : row.line.key">
            <ChatSystemLine
              v-if="row.kind === 'gap'"
              variant="gap"
              user=""
              :text="row.label"
            />
            <ChatMessageLine
              v-else
              :user="row.line.user"
              :message="row.line.message"
              :keyword="row.line.keyword"
              :user-marked="row.line.userMarked"
              :user-is-sus="row.line.userIsSus"
              :suspicious-title="row.line.susTitle"
              :from-sent="row.line.fromSent"
              :first-message="row.line.firstMessage"
              :badge-tags="row.line.badgeTags"
              :show-timestamp="false"
              :created-at="row.line.createdAtIso"
              :chatter-user-id="row.line.chatterUserId ?? undefined"
              :highlight-channel="normCh(selectedChannel)"
            />
          </template>
        </ul>

        <div class="composer">
          <label>
            <span>Send as</span>
            <select v-model.number="sendAccountId">
              <option v-for="a in twitchStore.accounts" :key="a.id" :value="a.id">{{ a.username }}</option>
            </select>
          </label>
          <label>
            <span>Message</span>
            <textarea
              v-model="sendText"
              class="composer-textarea"
              maxlength="500"
              rows="3"
              name="chat_message"
              autocomplete="off"
              autocorrect="off"
              autocapitalize="off"
              spellcheck="false"
              placeholder="Say something… (Enter to send, Shift+Enter for newline)"
              @keydown="onComposerKeydown"
            />
          </label>
          <SubmitButton
            native-type="button"
            class="btn-send"
            :loading="sendingChat"
            :disabled="!twitchStore.accounts.length"
            @click="sendChat"
          >
            {{ sendingChat ? 'Sending…' : 'Chat' }}
          </SubmitButton>
        </div>
      </section>
        </div>
      </div>
    </div>

    <AppModal :open="channelEntryModalOpen" title="Go to channel" @close="channelEntryModalOpen = false">
      <form class="channel-entry-form" @submit.prevent="applyManualChannel">
        <label class="channel-entry-label">
          <span>Twitch username</span>
          <input
            v-model="manualChannelInput"
            type="text"
            name="channel_login"
            autocomplete="off"
            autocorrect="off"
            autocapitalize="off"
            spellcheck="false"
            placeholder="channel_name"
          />
        </label>
        <div class="channel-entry-actions">
          <button type="button" class="btn-cancel" @click="channelEntryModalOpen = false">Cancel</button>
          <button type="submit" class="btn-confirm">Go</button>
        </div>
      </form>
    </AppModal>

    <AppModal :open="viewerModalOpen" extra-wide title="Stream details" @close="viewerModalOpen = false">
      <dl v-if="channelLive" class="viewer-dl">
        <div>
          <dt>Title</dt>
          <dd>{{ channelLive.title ?? '—' }}</dd>
        </div>
        <div>
          <dt>Category</dt>
          <dd>{{ channelLive.game_name ?? '—' }}</dd>
        </div>
        <div>
          <dt>Viewers</dt>
          <dd>{{ formatViewerDisplay(channelLive) }}</dd>
        </div>
        <div>
          <dt>Live since</dt>
          <dd>{{ channelLive.started_at ?? '—' }}</dd>
        </div>
        <div>
          <dt>Session uptime</dt>
          <dd>{{ formatSessionClock(channelLive.started_at) }}</dd>
        </div>
      </dl>
      <div v-if="channelLive?.is_live" class="viewer-chatters">
        <h3 class="viewer-chatters-title">Chatters in channel</h3>
        <div v-if="viewerChatters.length" class="viewer-chatter-toolbar">
          <label class="viewer-filter">
            <span class="sr-only">Filter by name</span>
            <input
              v-model="viewerFilterQuery"
              type="search"
              name="viewer_filter"
              autocomplete="off"
              autocorrect="off"
              spellcheck="false"
              placeholder="Filter…"
            />
          </label>
          <label class="viewer-sort">
            <span class="sr-only">Sort</span>
            <select v-model="viewerSort" name="viewer_sort">
              <option value="present_new">New in chat first</option>
              <option value="present_old">Longest in chat first</option>
              <option value="login_az">Name A–Z</option>
              <option value="login_za">Name Z–A</option>
              <option value="account_new">Newest Twitch accounts</option>
              <option value="account_old">Oldest Twitch accounts</option>
              <option value="message_high">Most messages</option>
              <option value="message_low">Fewest messages</option>
            </select>
          </label>
        </div>
        <p v-if="loadingViewerChatters" class="muted tiny">Loading…</p>
        <ul v-else-if="displayedViewerChatters.length" class="viewer-chatter-list">
          <li v-for="c in displayedViewerChatters" :key="c.user_twitch_id" class="viewer-chatter-row">
            <TwitchUserLink
              :login="c.login"
              :user-twitch-id="c.user_twitch_id"
              :highlight-channel="normCh(selectedChannel)"
              variant="chat"
            />
            <span class="viewer-chatter-meta">Present {{ formatPresentElapsed(c.present_since) }}</span>
            <span v-if="c.message_count != null && c.message_count !== undefined" class="viewer-chatter-meta"
              >Messages {{ c.message_count }}</span
            >
            <span v-if="c.account_created_at" class="viewer-chatter-meta"
              >Account {{ formatAccountDate(c.account_created_at) }}</span
            >
          </li>
        </ul>
        <p v-else-if="viewerChatters.length && !displayedViewerChatters.length" class="muted tiny">
          No names match the filter.
        </p>
        <p v-else class="muted tiny">No names loaded.</p>
      </div>
    </AppModal>
  </div>
</template>

<style scoped lang="scss">
.watch {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.75rem;
  flex: 1;
  min-height: 0;
}

.watch-layout {
  display: flex;
  flex: 1;
  min-height: 0;
  gap: 0.65rem;
  align-items: stretch;
}

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

.watch-main {
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  min-height: 0;
}

.channel-entry-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.channel-entry-label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  margin: 0;
  font-size: 0.82rem;
  color: var(--text-muted);

  input {
    padding: 0.45rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.9rem;
  }
}

.channel-entry-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 0.4rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text);
  font-size: 0.85rem;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
  }
}

.btn-confirm {
  padding: 0.4rem 0.85rem;
  border-radius: 0.25rem;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;

  &:hover {
    filter: brightness(1.08);
  }
}

@media (max-width: 639px) {
  .watch-layout {
    flex-direction: column;
  }

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

.muted {
  color: var(--text-muted);
  font-size: 0.85rem;
}

.viewer-chatters {
  margin-top: 0.75rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--border);
}

.viewer-chatters-title {
  margin: 0 0 0.4rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
}

.viewer-chatter-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;

  input,
  select {
    padding: 0.3rem 0.45rem;
    border-radius: 0.2rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.78rem;
    min-width: 0;
  }

  .viewer-filter {
    flex: 1 1 10rem;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .viewer-sort {
    flex: 0 1 auto;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.viewer-chatter-list {
  list-style: none;
  margin: 0;
  padding: 0;
  /* Fixed viewport so filtering does not resize the panel */
  height: min(36vh, 32rem);
  min-height: min(36vh, 32rem);
  max-height: min(36vh, 32rem);
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  font-size: 0.82rem;
}

.viewer-chatter-row {
  display: flex;
  flex-direction: column;
  gap: 0.12rem;
  padding: 0.35rem 0.25rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-base);
}

.viewer-chatter-meta {
  font-size: 0.72rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

.viewer-dl {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;

  div {
    display: grid;
    grid-template-columns: 7rem 1fr;
    gap: 0.35rem;
    align-items: baseline;
  }

  dt {
    margin: 0;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  dd {
    margin: 0;
    font-size: 0.85rem;
    word-break: break-word;
  }
}

.pane-head {
  flex-shrink: 0;
  min-height: 2.875rem;
  display: flex;
  align-items: center;
  box-sizing: border-box;
  padding: 0.35rem 0;
}

.grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  flex: 1;
  min-height: 0;
  overflow: hidden;

  @media (min-width: 900px) {
    flex-direction: row;
    align-items: stretch;
  }
}

.video {
  flex: 0 0 auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  min-height: 0;

  @media (min-width: 900px) {
    flex: 1 1 auto;
    min-width: 0;
  }
}

.chat {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  overflow: hidden;

  @media (min-width: 900px) {
    flex: 0 0 min(400px, 38vw);
    align-self: stretch;
    max-width: 400px;
    min-height: 0;
    max-height: 100%;
  }
}

.chat-head {
  width: 100%;
  justify-content: flex-start;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--border);
  margin: 0;
  padding-left: 0.5rem;
  padding-right: 0.5rem;
  font-size: 0.85rem;
  font-weight: 600;
  box-sizing: border-box;
}

.chat-head-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  flex: 1;
}

.irc-link-dot {
  flex-shrink: 0;
  width: 0.55rem;
  height: 0.55rem;
  margin-left: auto;
  border-radius: 50%;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.25);

  &--on {
    background: #2ecc71;
  }

  &--off {
    background: #c0392b;
  }
}

.chat-title-main {
  flex-shrink: 0;
  line-height: 1.25;
}

.chat-channel-tag {
  font-weight: 700;
  color: var(--accent-bright);
  letter-spacing: 0.02em;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-channel-empty {
  font-size: 0.85rem;
}

.lines {
  list-style: none;
  margin: 0;
  padding: 0.4rem;
  flex: 1 1 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
  font-size: 0.82rem;
  line-height: 1.35;

}

.composer {
  border-top: 1px solid var(--border);
  padding: 0.5rem;
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.4rem;
  font-size: 0.78rem;
  flex-shrink: 0;

  label {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    color: var(--text-muted);

    span {
      font-size: 0.72rem;
    }
  }

  select,
  input,
  textarea {
    padding: 0.35rem 0.4rem;
    border-radius: 0.2rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.85rem;
  }

  .composer-textarea {
    resize: vertical;
    min-height: 3.25rem;
    line-height: 1.35;
    font-family: inherit;
  }

  .btn-send {
    grid-column: 1 / -1;
    padding: 0.45rem;
    border: none;
    border-radius: 0.25rem;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
    cursor: pointer;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }
}
</style>
