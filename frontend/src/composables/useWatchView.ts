import { useIntervalFn } from '@vueuse/core';
import { storeToRefs } from 'pinia';
import { computed, onMounted, ref, watch } from 'vue';
import { ApiError, ChatHistoryEntry, DefaultService } from '../api/generated';
import type { ListChannelChattersRequest } from '../api/generated';
import type {
  ChannelChatterEntry,
  ChannelLive,
  IrcMonitorStatus,
  WatchUiHints,
} from '../api/generated';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { isChannelJoinedOnIrc } from '../lib/ircMonitorJoined';
import { formatDateTime } from '../lib/dateTime';
import { effectiveChatterIsSus, effectiveSuspicionTitle } from '../lib/suspicionOverlay';
import { notifyFromApiError } from '../lib/clientNotice';
import { notify } from '../lib/notify';
import { useChannelsStore } from '../stores/channels';
import { useLiveSocketStore } from '../stores/liveSocket';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';
import { useWatchPreferencesStore } from '../stores/watchPreferences';
import {
  directoryRowToChannelLive,
  normCh,
  parseChannelLivePayload,
  TWITCH_LOGIN_RE,
} from '../views/watch/channelHelpers';
import type { ChatLine, ChatRow, ViewerSortMode } from '../views/watch/types';

export function useWatchView() {
  const channelsStore = useChannelsStore();
  const twitchStore = useTwitchAccountsStore();

  const liveSocket = useLiveSocketStore();
  const { events } = storeToRefs(liveSocket);

  const watchPrefs = useWatchPreferencesStore();
  const { chatGapMinutes } = storeToRefs(watchPrefs);

  const selectedChannel = ref('');
  const sendAccountId = ref<number | null>(null);
  const sendText = ref('');
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

  function isMonitoredLogin(login: string): boolean {
    return channelsStore.monitoredChannels.some((c) => normCh(c.username) === login);
  }

  /** Prefer last opened channel from local storage when it is still a monitored channel or a valid manual login. */
  function preferredWatchChannel(monitored: { username: string }[]): string {
    const last = watchPrefs.getLastWatchChannel();
    if (last) {
      const fromList = monitored.find((c) => normCh(c.username) === last);
      if (fromList) {
        return fromList.username;
      }
      if (TWITCH_LOGIN_RE.test(last)) {
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

  const ircChatConnected = computed(() =>
    isChannelJoinedOnIrc(ircStatusForWatchDot.value, selectedChannel.value),
  );

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
    return formatDateTime(iso);
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
  const monitoredLivePollMs = computed(() =>
    Math.max(1000, (watchHints.value?.monitored_live_poll_interval_seconds ?? 60) * 1000),
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
      const rows = await DefaultService.listTwitchDirectoryUsers({ monitoredOnly: true, limit: 200 });
      const byLogin = new Map(rows.map((r) => [normCh(r.username), r] as const));
      monitoredSidebar.value = list.map((c) => {
        const row = byLogin.get(normCh(c.username));
        if (row) {
          return directoryRowToChannelLive(row);
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
    if (isMonitoredLogin(login)) {
      const fromSidebar = monitoredSidebar.value.find((x) => normCh(x.broadcaster_login) === login);
      if (fromSidebar) {
        channelLive.value = fromSidebar;
      }
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

  watch(
    monitoredSidebar,
    (rows) => {
      const login = normCh(selectedChannel.value);
      if (!login || !isMonitoredLogin(login)) {
        return;
      }
      const hit = rows.find((x) => normCh(x.broadcaster_login) === login);
      if (hit) {
        channelLive.value = hit;
      }
    },
    { deep: true },
  );

  useIntervalFn(
    () => {
      void refreshMonitoredSidebar(true);
    },
    monitoredLivePollMs,
  );

  useIntervalFn(
    () => {
      void pollChannelLive();
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
        if (isMonitoredLogin(login)) {
          await refreshMonitoredSidebar(true);
          const fromSidebar = monitoredSidebar.value.find((x) => normCh(x.broadcaster_login) === login);
          channelLive.value = fromSidebar ?? null;
          return;
        }
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

  function selectChannel(raw: string): void {
    const login = normCh(raw);
    if (login) {
      selectedChannel.value = login;
    }
  }

  function applyManualChannel(): void {
    const login = normCh(manualChannelInput.value);
    if (!login || !TWITCH_LOGIN_RE.test(login)) {
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

  return {
    twitchStore,
    selectedChannel,
    sendAccountId,
    sendText,
    channelLive,
    loadingChannelMeta,
    loadingMonitoredSidebar,
    viewerModalOpen,
    viewerChatters,
    loadingViewerChatters,
    viewerFilterQuery,
    viewerSort,
    channelEntryModalOpen,
    manualChannelInput,
    sendingChat,
    onlineMonitored,
    offlineMonitored,
    monitoredSidebar,
    ircChatConnected,
    displayedViewerChatters,
    displayRows,
    chatScrollSig,
    formatSessionClock,
    formatViewerDisplay,
    formatDateTime,
    formatPresentElapsed,
    formatAccountDate,
    normCh,
    selectChannel,
    applyManualChannel,
    onComposerKeydown,
    sendChat,
  };
}
