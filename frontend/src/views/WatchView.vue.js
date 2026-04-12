/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
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
const sendAccountId = ref(null);
const sendText = ref('');
const chatEl = ref(null);
const historyEntries = ref([]);
const channelSwitchTime = ref(0);
const channelLive = ref(null);
const monitoredSidebar = ref([]);
const loadingChannelMeta = ref(false);
const loadingMonitoredSidebar = ref(false);
const viewerModalOpen = ref(false);
const viewerChatters = ref([]);
const loadingViewerChatters = ref(false);
const viewerFilterQuery = ref('');
const viewerSort = ref('present_new');
const watchHints = ref(null);
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
function normCh(c) {
    return c.replace(/^#/, '').toLowerCase();
}
function formatGapMinutes(fromMs, toMs) {
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
function formatSessionClock(iso) {
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
function formatSecondsAsClock(sec) {
    const s = Math.max(0, Math.floor(sec));
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const r = s % 60;
    return [h, m, r].map((n) => String(n).padStart(2, '0')).join(':');
}
function formatAccountDate(iso) {
    if (!iso) {
        return '';
    }
    const t = Date.parse(iso);
    if (!Number.isFinite(t)) {
        return '';
    }
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(t);
}
function formatPresentElapsed(iso) {
    const t = Date.parse(iso);
    if (!Number.isFinite(t)) {
        return '—';
    }
    const sec = Math.max(0, Math.floor((sessionClock.value - t) / 1000));
    return formatSecondsAsClock(sec);
}
/** Helix viewer count with delta vs IRC chatter snapshot: e.g. 4 (+3) or 4 (-2). */
function formatViewerDisplay(live) {
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
const viewerPollMs = computed(() => (watchHints.value?.viewer_poll_interval_seconds ?? 10) * 1000);
const chattersSyncMs = computed(() => (watchHints.value?.channel_chatters_sync_interval_seconds ?? 10) * 1000);
const onlineMonitored = computed(() => [...monitoredSidebar.value.filter((f) => f.is_live)].sort((a, b) => a.broadcaster_login.localeCompare(b.broadcaster_login)));
const offlineMonitored = computed(() => [...monitoredSidebar.value.filter((f) => !f.is_live)].sort((a, b) => a.broadcaster_login.localeCompare(b.broadcaster_login)));
const displayedViewerChatters = computed(() => {
    let rows = [...viewerChatters.value];
    const q = viewerFilterQuery.value.trim().toLowerCase();
    if (q) {
        rows = rows.filter((c) => c.login.toLowerCase().includes(q));
    }
    const presentMs = (iso) => {
        const t = Date.parse(iso);
        return Number.isFinite(t) ? t : 0;
    };
    const accountMs = (iso) => {
        if (!iso) {
            return NaN;
        }
        const t = Date.parse(iso);
        return Number.isFinite(t) ? t : NaN;
    };
    rows.sort((a, b) => {
        const msgCount = (c) => c.message_count ?? 0;
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
async function refreshMonitoredSidebar(silent = false) {
    const list = channelsStore.monitoredChannels;
    if (!list.length) {
        monitoredSidebar.value = [];
        return;
    }
    if (!silent) {
        loadingMonitoredSidebar.value = true;
    }
    try {
        const results = await Promise.all(list.map((c) => DefaultService.getChannelLive({ requestBody: { login: normCh(c.username) } }).catch(() => null)));
        monitoredSidebar.value = list.map((c, i) => {
            const live = results[i];
            if (live && typeof live === 'object' && 'broadcaster_login' in live) {
                return live;
            }
            return {
                broadcaster_id: c.id,
                broadcaster_login: normCh(c.username),
                display_name: c.username,
                profile_image_url: '',
                is_live: false,
            };
        });
    }
    finally {
        if (!silent) {
            loadingMonitoredSidebar.value = false;
        }
    }
}
watch(() => channelsStore.monitoredChannels, () => {
    void refreshMonitoredSidebar();
}, { deep: true, immediate: true });
async function pollChannelLive() {
    const login = normCh(selectedChannel.value);
    if (!login) {
        return;
    }
    try {
        const live = await DefaultService.getChannelLive({ requestBody: { login } });
        if (live && typeof live === 'object' && 'broadcaster_login' in live) {
            channelLive.value = live;
        }
    }
    catch {
        /* keep previous channelLive */
    }
}
useIntervalFn(() => {
    void pollChannelLive();
    void refreshMonitoredSidebar(true);
}, viewerPollMs);
const displayLines = computed(() => {
    const ch = normCh(selectedChannel.value);
    if (!ch) {
        return [];
    }
    const hist = historyEntries.value.map((h) => {
        const t = Date.parse(h.created_at);
        const at = Number.isFinite(t) ? t : channelSwitchTime.value;
        const badgeTags = [...(h.badge_tags ?? [])];
        return {
            key: `db-${h.id}`,
            user: h.user,
            message: h.message,
            keyword: h.keyword_match,
            userMarked: h.chatter_marked,
            fromSent: h.source === ChatHistoryEntry.source.SENT,
            at,
            badgeTags,
            createdAtIso: h.created_at,
            chatterUserId: h.chatter_user_id ?? undefined,
        };
    });
    const since = channelSwitchTime.value;
    const live = [];
    for (let i = 0; i < events.value.length; i++) {
        const e = events.value[i];
        if (normCh(e.channel) !== ch || e.receivedAt < since) {
            continue;
        }
        if (e.type === 'chatter_join') {
            const fromIso = e.created_at ? Date.parse(e.created_at) : NaN;
            const at = Number.isFinite(fromIso) ? fromIso : e.receivedAt;
            live.push({
                key: `ws-${e.receivedAt}-${i}-chatter_join`,
                user: e.user,
                message: 'joined chat',
                keyword: false,
                userMarked: false,
                fromSent: false,
                at,
                badgeTags: [],
                createdAtIso: e.created_at,
                chatterUserId: e.user_twitch_id,
                system: 'join',
                accountCreatedAt: e.account_created_at,
            });
            continue;
        }
        if (e.type === 'chatter_part') {
            const fromIso = e.created_at ? Date.parse(e.created_at) : NaN;
            const at = Number.isFinite(fromIso) ? fromIso : e.receivedAt;
            live.push({
                key: `ws-${e.receivedAt}-${i}-chatter_part`,
                user: e.user,
                message: 'left chat',
                keyword: false,
                userMarked: false,
                fromSent: false,
                at,
                badgeTags: [],
                createdAtIso: e.created_at,
                chatterUserId: e.user_twitch_id,
                system: 'part',
                presentSeconds: e.present_seconds,
            });
            continue;
        }
        if (e.type === 'chat_message' || e.type === 'message_sent') {
            const fromIso = e.created_at ? Date.parse(e.created_at) : NaN;
            const at = Number.isFinite(fromIso) ? fromIso : e.receivedAt;
            const badgeTags = (e.badge_tags ?? []);
            live.push({
                key: `ws-${e.receivedAt}-${i}-${e.type}`,
                user: e.user,
                message: e.message,
                keyword: e.type === 'chat_message' && Boolean(e.keyword_match),
                userMarked: e.type === 'chat_message' && Boolean(e.chatter_marked),
                fromSent: e.type === 'message_sent',
                at,
                badgeTags,
                createdAtIso: e.created_at,
                chatterUserId: e.user_twitch_id,
            });
        }
    }
    const merged = [...hist, ...live];
    merged.sort((a, b) => a.at - b.at);
    return merged;
});
const displayRows = computed(() => {
    const lines = displayLines.value;
    if (lines.length === 0) {
        return [];
    }
    const gapMs = Math.max(1, chatGapMinutes.value) * 60_000;
    const rows = [];
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
        if (channelsStore.monitoredChannels.length) {
            selectedChannel.value = channelsStore.monitoredChannels[0].username;
        }
        if (twitchStore.accounts.length) {
            sendAccountId.value = twitchStore.accounts[0].id;
        }
    }
    catch {
        notify({
            id: 'watch-load',
            type: 'error',
            title: 'Watch',
            description: 'Could not load channels or Twitch accounts.',
        });
    }
    try {
        watchHints.value = await DefaultService.getWatchUiHints();
    }
    catch {
        watchHints.value = null;
    }
});
watch(() => channelsStore.monitoredChannels, (list) => {
    if (!selectedChannel.value && list.length) {
        selectedChannel.value = list[0].username;
    }
}, { deep: true });
watch(channelEntryModalOpen, (open) => {
    if (open) {
        manualChannelInput.value = selectedChannel.value ? normCh(selectedChannel.value) : '';
    }
});
watch(() => twitchStore.accounts, (list) => {
    if (!sendAccountId.value || !list.some((a) => a.id === sendAccountId.value)) {
        sendAccountId.value = list[0]?.id ?? null;
    }
}, { deep: true });
watch(selectedChannel, async (ch) => {
    if (!ch) {
        historyEntries.value = [];
        return;
    }
    const norm = normCh(ch);
    try {
        historyEntries.value = await DefaultService.listChatHistory({ channel: norm, limit: 80 });
    }
    catch (e) {
        historyEntries.value = [];
        if (e instanceof ApiError && e.status === 404) {
            /* channel not monitored — leave empty */
        }
    }
    channelSwitchTime.value = Date.now();
}, { flush: 'post' });
watch(selectedChannel, async (ch) => {
    channelLive.value = null;
    if (!ch) {
        return;
    }
    const login = normCh(ch);
    loadingChannelMeta.value = true;
    try {
        channelLive.value = await DefaultService.getChannelLive({ requestBody: { login } });
    }
    catch {
        channelLive.value = null;
    }
    finally {
        loadingChannelMeta.value = false;
    }
}, { flush: 'post' });
async function loadViewerChatters(silent) {
    const ch = normCh(selectedChannel.value);
    if (!ch || sendAccountId.value == null) {
        return;
    }
    if (!silent) {
        loadingViewerChatters.value = true;
    }
    try {
        const body = {
            account_id: sendAccountId.value,
            login: ch,
        };
        if (channelLive.value?.is_live && channelLive.value.started_at) {
            body.session_started_at = channelLive.value.started_at;
        }
        viewerChatters.value = await DefaultService.listChannelChatters({
            requestBody: body,
        });
    }
    catch (e) {
        if (!silent) {
            viewerChatters.value = [];
            notifyFromApiError(e, {
                id: 'watch-viewers-chatters',
                title: 'Viewers',
                fallbackDescription: 'Could not load chatter list (needs moderator:read:chatters).',
            });
        }
    }
    finally {
        if (!silent) {
            loadingViewerChatters.value = false;
        }
    }
}
const { pause: pauseViewerModalPoll, resume: resumeViewerModalPoll } = useIntervalFn(() => {
    void pollChannelLive();
    void loadViewerChatters(true);
}, chattersSyncMs, { immediate: false });
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
watch(displayRows, async () => {
    await nextTick();
    const el = chatEl.value;
    if (el) {
        el.scrollTop = el.scrollHeight;
    }
}, { deep: true });
function selectChannel(raw) {
    const login = normCh(raw);
    if (login) {
        selectedChannel.value = login;
    }
}
function applyManualChannel() {
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
function onComposerKeydown(e) {
    if (e.key !== 'Enter' || e.shiftKey) {
        return;
    }
    e.preventDefault();
    void sendChat();
}
async function sendChat() {
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
    }
    catch (e) {
        notifyFromApiError(e, {
            id: 'watch-send',
            title: 'Chat',
            fallbackDescription: 'Send failed. Check account token and channel name.',
        });
    }
    finally {
        sendingChat.value = false;
    }
}
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['follow-initial']} */ ;
/** @type {__VLS_StyleScopedClasses['watch-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['follows-sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['follows-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['stream-strip']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['dn']} */ ;
/** @type {__VLS_StyleScopedClasses['game-line']} */ ;
/** @type {__VLS_StyleScopedClasses['stream-stats']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "watch" },
});
/** @type {__VLS_StyleScopedClasses['watch']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "watch-layout" },
});
/** @type {__VLS_StyleScopedClasses['watch-layout']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.aside, __VLS_intrinsics.aside)({
    ...{ class: "follows-sidebar" },
    'aria-label': "Monitored channels",
});
/** @type {__VLS_StyleScopedClasses['follows-sidebar']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.channelEntryModalOpen = true;
            // @ts-ignore
            [channelEntryModalOpen,];
        } },
    type: "button",
    ...{ class: "follow-tile follow-tile--add" },
    title: "Go to channel…",
    'aria-label': "Go to channel by name",
});
/** @type {__VLS_StyleScopedClasses['follow-tile']} */ ;
/** @type {__VLS_StyleScopedClasses['follow-tile--add']} */ ;
if (__VLS_ctx.loadingMonitoredSidebar) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "muted tiny follows-hint" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
    /** @type {__VLS_StyleScopedClasses['follows-hint']} */ ;
}
else {
    for (const [f] of __VLS_vFor((__VLS_ctx.onlineMonitored))) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.loadingMonitoredSidebar))
                        return;
                    __VLS_ctx.selectChannel(f.broadcaster_login);
                    // @ts-ignore
                    [loadingMonitoredSidebar, onlineMonitored, selectChannel,];
                } },
            key: (f.broadcaster_id),
            type: "button",
            ...{ class: "follow-tile" },
            ...{ class: ({ 'follow-tile--active': __VLS_ctx.normCh(__VLS_ctx.selectedChannel) === __VLS_ctx.normCh(f.broadcaster_login) }) },
            title: (`#${f.broadcaster_login}`),
            'aria-label': (`Open channel ${f.broadcaster_login}`),
        });
        /** @type {__VLS_StyleScopedClasses['follow-tile']} */ ;
        /** @type {__VLS_StyleScopedClasses['follow-tile--active']} */ ;
        if (f.profile_image_url) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.img)({
                ...{ class: "follow-avatar" },
                src: (f.profile_image_url),
                alt: (''),
                width: "40",
                height: "40",
            });
            /** @type {__VLS_StyleScopedClasses['follow-avatar']} */ ;
        }
        else {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "follow-initial" },
                'aria-hidden': "true",
            });
            /** @type {__VLS_StyleScopedClasses['follow-initial']} */ ;
            (f.broadcaster_login.charAt(0).toUpperCase());
        }
        // @ts-ignore
        [normCh, normCh, selectedChannel,];
    }
    for (const [f] of __VLS_vFor((__VLS_ctx.offlineMonitored))) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.loadingMonitoredSidebar))
                        return;
                    __VLS_ctx.selectChannel(f.broadcaster_login);
                    // @ts-ignore
                    [selectChannel, offlineMonitored,];
                } },
            key: ('off-' + f.broadcaster_id),
            type: "button",
            ...{ class: "follow-tile follow-tile--offline" },
            ...{ class: ({ 'follow-tile--active': __VLS_ctx.normCh(__VLS_ctx.selectedChannel) === __VLS_ctx.normCh(f.broadcaster_login) }) },
            title: (`#${f.broadcaster_login} (offline)`),
            'aria-label': (`Open channel ${f.broadcaster_login}`),
        });
        /** @type {__VLS_StyleScopedClasses['follow-tile']} */ ;
        /** @type {__VLS_StyleScopedClasses['follow-tile--offline']} */ ;
        /** @type {__VLS_StyleScopedClasses['follow-tile--active']} */ ;
        if (f.profile_image_url) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.img)({
                ...{ class: "follow-avatar" },
                src: (f.profile_image_url),
                alt: (''),
                width: "40",
                height: "40",
            });
            /** @type {__VLS_StyleScopedClasses['follow-avatar']} */ ;
        }
        else {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "follow-initial" },
                'aria-hidden': "true",
            });
            /** @type {__VLS_StyleScopedClasses['follow-initial']} */ ;
            (f.broadcaster_login.charAt(0).toUpperCase());
        }
        // @ts-ignore
        [normCh, normCh, selectedChannel,];
    }
    if (!__VLS_ctx.monitoredSidebar.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted tiny follows-hint" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
        /** @type {__VLS_StyleScopedClasses['follows-hint']} */ ;
    }
}
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "watch-main" },
});
/** @type {__VLS_StyleScopedClasses['watch-main']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "grid" },
});
/** @type {__VLS_StyleScopedClasses['grid']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "video" },
});
/** @type {__VLS_StyleScopedClasses['video']} */ ;
const __VLS_0 = TwitchPlayer;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
    channel: (__VLS_ctx.selectedChannel),
}));
const __VLS_2 = __VLS_1({
    channel: (__VLS_ctx.selectedChannel),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
if (__VLS_ctx.channelLive) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
        ...{ class: "stream-strip compact" },
    });
    /** @type {__VLS_StyleScopedClasses['stream-strip']} */ ;
    /** @type {__VLS_StyleScopedClasses['compact']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.img)({
        ...{ class: "avatar" },
        src: (__VLS_ctx.channelLive.profile_image_url),
        alt: (''),
        width: "36",
        height: "36",
    });
    /** @type {__VLS_StyleScopedClasses['avatar']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "stream-meta" },
    });
    /** @type {__VLS_StyleScopedClasses['stream-meta']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "stream-title-row" },
    });
    /** @type {__VLS_StyleScopedClasses['stream-title-row']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "dn" },
    });
    /** @type {__VLS_StyleScopedClasses['dn']} */ ;
    (__VLS_ctx.channelLive.display_name);
    if (__VLS_ctx.channelLive.is_live) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "live-pill" },
        });
        /** @type {__VLS_StyleScopedClasses['live-pill']} */ ;
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "off-pill" },
        });
        /** @type {__VLS_StyleScopedClasses['off-pill']} */ ;
    }
    if (__VLS_ctx.channelLive.is_live && __VLS_ctx.channelLive.title) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "game-line" },
        });
        /** @type {__VLS_StyleScopedClasses['game-line']} */ ;
        (__VLS_ctx.channelLive.title);
    }
    if (__VLS_ctx.channelLive.is_live && __VLS_ctx.channelLive.game_name) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "game-line" },
        });
        /** @type {__VLS_StyleScopedClasses['game-line']} */ ;
        (__VLS_ctx.channelLive.game_name);
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "stream-stats" },
    });
    /** @type {__VLS_StyleScopedClasses['stream-stats']} */ ;
    if (__VLS_ctx.channelLive.is_live) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "uptime" },
        });
        /** @type {__VLS_StyleScopedClasses['uptime']} */ ;
        (__VLS_ctx.formatSessionClock(__VLS_ctx.channelLive.started_at));
    }
    if (__VLS_ctx.channelLive.is_live) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(__VLS_ctx.channelLive))
                        return;
                    if (!(__VLS_ctx.channelLive.is_live))
                        return;
                    __VLS_ctx.viewerModalOpen = true;
                    // @ts-ignore
                    [selectedChannel, monitoredSidebar, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, formatSessionClock, viewerModalOpen,];
                } },
            type: "button",
            ...{ class: "viewers-btn" },
        });
        /** @type {__VLS_StyleScopedClasses['viewers-btn']} */ ;
        (__VLS_ctx.formatViewerDisplay(__VLS_ctx.channelLive));
    }
}
else if (__VLS_ctx.selectedChannel && __VLS_ctx.loadingChannelMeta) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
        ...{ class: "stream-strip compact placeholder" },
    });
    /** @type {__VLS_StyleScopedClasses['stream-strip']} */ ;
    /** @type {__VLS_StyleScopedClasses['compact']} */ ;
    /** @type {__VLS_StyleScopedClasses['placeholder']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "chat" },
});
/** @type {__VLS_StyleScopedClasses['chat']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "pane-head chat-head" },
});
/** @type {__VLS_StyleScopedClasses['pane-head']} */ ;
/** @type {__VLS_StyleScopedClasses['chat-head']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "chat-head-main" },
});
/** @type {__VLS_StyleScopedClasses['chat-head-main']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
    ...{ class: "chat-title-main" },
});
/** @type {__VLS_StyleScopedClasses['chat-title-main']} */ ;
if (__VLS_ctx.selectedChannel) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "chat-channel-tag" },
    });
    /** @type {__VLS_StyleScopedClasses['chat-channel-tag']} */ ;
    (__VLS_ctx.normCh(__VLS_ctx.selectedChannel));
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "chat-channel-empty muted" },
    });
    /** @type {__VLS_StyleScopedClasses['chat-channel-empty']} */ ;
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
__VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
    ref: "chatEl",
    ...{ class: "lines" },
});
/** @type {__VLS_StyleScopedClasses['lines']} */ ;
for (const [row] of __VLS_vFor((__VLS_ctx.displayRows))) {
    (row.kind === 'gap' ? row.key : row.line.key);
    if (row.kind === 'gap') {
        const __VLS_5 = ChatSystemLine;
        // @ts-ignore
        const __VLS_6 = __VLS_asFunctionalComponent1(__VLS_5, new __VLS_5({
            variant: "gap",
            user: "",
            text: (row.label),
        }));
        const __VLS_7 = __VLS_6({
            variant: "gap",
            user: "",
            text: (row.label),
        }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    }
    else if (row.line.system === 'join') {
        const __VLS_10 = ChatSystemLine;
        // @ts-ignore
        const __VLS_11 = __VLS_asFunctionalComponent1(__VLS_10, new __VLS_10({
            variant: "join",
            user: (row.line.user),
            text: "joined chat",
            detail: (row.line.accountCreatedAt
                ? `Account since ${__VLS_ctx.formatAccountDate(row.line.accountCreatedAt)}`
                : ''),
            chatterUserId: (row.line.chatterUserId ?? undefined),
            highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
        }));
        const __VLS_12 = __VLS_11({
            variant: "join",
            user: (row.line.user),
            text: "joined chat",
            detail: (row.line.accountCreatedAt
                ? `Account since ${__VLS_ctx.formatAccountDate(row.line.accountCreatedAt)}`
                : ''),
            chatterUserId: (row.line.chatterUserId ?? undefined),
            highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
        }, ...__VLS_functionalComponentArgsRest(__VLS_11));
    }
    else if (row.line.system === 'part') {
        const __VLS_15 = ChatSystemLine;
        // @ts-ignore
        const __VLS_16 = __VLS_asFunctionalComponent1(__VLS_15, new __VLS_15({
            variant: "part",
            user: (row.line.user),
            text: "left chat",
            detail: (row.line.presentSeconds != null
                ? `In chat for ${__VLS_ctx.formatSecondsAsClock(row.line.presentSeconds)}`
                : ''),
            chatterUserId: (row.line.chatterUserId ?? undefined),
            highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
        }));
        const __VLS_17 = __VLS_16({
            variant: "part",
            user: (row.line.user),
            text: "left chat",
            detail: (row.line.presentSeconds != null
                ? `In chat for ${__VLS_ctx.formatSecondsAsClock(row.line.presentSeconds)}`
                : ''),
            chatterUserId: (row.line.chatterUserId ?? undefined),
            highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
        }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    }
    else {
        const __VLS_20 = ChatMessageLine;
        // @ts-ignore
        const __VLS_21 = __VLS_asFunctionalComponent1(__VLS_20, new __VLS_20({
            user: (row.line.user),
            message: (row.line.message),
            keyword: (row.line.keyword),
            userMarked: (row.line.userMarked),
            fromSent: (row.line.fromSent),
            badgeTags: (row.line.badgeTags),
            showTimestamp: (false),
            createdAt: (row.line.createdAtIso),
            chatterUserId: (row.line.chatterUserId ?? undefined),
            highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
        }));
        const __VLS_22 = __VLS_21({
            user: (row.line.user),
            message: (row.line.message),
            keyword: (row.line.keyword),
            userMarked: (row.line.userMarked),
            fromSent: (row.line.fromSent),
            badgeTags: (row.line.badgeTags),
            showTimestamp: (false),
            createdAt: (row.line.createdAtIso),
            chatterUserId: (row.line.chatterUserId ?? undefined),
            highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
        }, ...__VLS_functionalComponentArgsRest(__VLS_21));
    }
    // @ts-ignore
    [normCh, normCh, normCh, normCh, selectedChannel, selectedChannel, selectedChannel, selectedChannel, selectedChannel, selectedChannel, channelLive, formatViewerDisplay, loadingChannelMeta, displayRows, formatAccountDate, formatSecondsAsClock,];
}
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "composer" },
});
/** @type {__VLS_StyleScopedClasses['composer']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.select, __VLS_intrinsics.select)({
    value: (__VLS_ctx.sendAccountId),
});
for (const [a] of __VLS_vFor((__VLS_ctx.twitchStore.accounts))) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        key: (a.id),
        value: (a.id),
    });
    (a.username);
    // @ts-ignore
    [sendAccountId, twitchStore,];
}
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.textarea)({
    ...{ onKeydown: (__VLS_ctx.onComposerKeydown) },
    value: (__VLS_ctx.sendText),
    ...{ class: "composer-textarea" },
    maxlength: "500",
    rows: "3",
    name: "chat_message",
    autocomplete: "off",
    autocorrect: "off",
    autocapitalize: "off",
    spellcheck: "false",
    placeholder: "Say something… (Enter to send, Shift+Enter for newline)",
});
/** @type {__VLS_StyleScopedClasses['composer-textarea']} */ ;
const __VLS_25 = SubmitButton || SubmitButton;
// @ts-ignore
const __VLS_26 = __VLS_asFunctionalComponent1(__VLS_25, new __VLS_25({
    ...{ 'onClick': {} },
    nativeType: "button",
    ...{ class: "btn-send" },
    loading: (__VLS_ctx.sendingChat),
    disabled: (!__VLS_ctx.twitchStore.accounts.length),
}));
const __VLS_27 = __VLS_26({
    ...{ 'onClick': {} },
    nativeType: "button",
    ...{ class: "btn-send" },
    loading: (__VLS_ctx.sendingChat),
    disabled: (!__VLS_ctx.twitchStore.accounts.length),
}, ...__VLS_functionalComponentArgsRest(__VLS_26));
let __VLS_30;
const __VLS_31 = ({ click: {} },
    { onClick: (__VLS_ctx.sendChat) });
/** @type {__VLS_StyleScopedClasses['btn-send']} */ ;
const { default: __VLS_32 } = __VLS_28.slots;
(__VLS_ctx.sendingChat ? 'Sending…' : 'Chat');
// @ts-ignore
[twitchStore, onComposerKeydown, sendText, sendingChat, sendingChat, sendChat,];
var __VLS_28;
var __VLS_29;
const __VLS_33 = AppModal || AppModal;
// @ts-ignore
const __VLS_34 = __VLS_asFunctionalComponent1(__VLS_33, new __VLS_33({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.channelEntryModalOpen),
    title: "Go to channel",
}));
const __VLS_35 = __VLS_34({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.channelEntryModalOpen),
    title: "Go to channel",
}, ...__VLS_functionalComponentArgsRest(__VLS_34));
let __VLS_38;
const __VLS_39 = ({ close: {} },
    { onClose: (...[$event]) => {
            __VLS_ctx.channelEntryModalOpen = false;
            // @ts-ignore
            [channelEntryModalOpen, channelEntryModalOpen,];
        } });
const { default: __VLS_40 } = __VLS_36.slots;
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.applyManualChannel) },
    ...{ class: "channel-entry-form" },
});
/** @type {__VLS_StyleScopedClasses['channel-entry-form']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "channel-entry-label" },
});
/** @type {__VLS_StyleScopedClasses['channel-entry-label']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    value: (__VLS_ctx.manualChannelInput),
    type: "text",
    name: "channel_login",
    autocomplete: "off",
    autocorrect: "off",
    autocapitalize: "off",
    spellcheck: "false",
    placeholder: "channel_name",
});
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "channel-entry-actions" },
});
/** @type {__VLS_StyleScopedClasses['channel-entry-actions']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.channelEntryModalOpen = false;
            // @ts-ignore
            [channelEntryModalOpen, applyManualChannel, manualChannelInput,];
        } },
    type: "button",
    ...{ class: "btn-cancel" },
});
/** @type {__VLS_StyleScopedClasses['btn-cancel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    type: "submit",
    ...{ class: "btn-confirm" },
});
/** @type {__VLS_StyleScopedClasses['btn-confirm']} */ ;
// @ts-ignore
[];
var __VLS_36;
var __VLS_37;
const __VLS_41 = AppModal || AppModal;
// @ts-ignore
const __VLS_42 = __VLS_asFunctionalComponent1(__VLS_41, new __VLS_41({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.viewerModalOpen),
    extraWide: true,
    title: "Stream details",
}));
const __VLS_43 = __VLS_42({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.viewerModalOpen),
    extraWide: true,
    title: "Stream details",
}, ...__VLS_functionalComponentArgsRest(__VLS_42));
let __VLS_46;
const __VLS_47 = ({ close: {} },
    { onClose: (...[$event]) => {
            __VLS_ctx.viewerModalOpen = false;
            // @ts-ignore
            [viewerModalOpen, viewerModalOpen,];
        } });
const { default: __VLS_48 } = __VLS_44.slots;
if (__VLS_ctx.channelLive) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.dl, __VLS_intrinsics.dl)({
        ...{ class: "viewer-dl" },
    });
    /** @type {__VLS_StyleScopedClasses['viewer-dl']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.channelLive.title ?? '—');
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.channelLive.game_name ?? '—');
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.formatViewerDisplay(__VLS_ctx.channelLive));
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.channelLive.started_at ?? '—');
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.formatSessionClock(__VLS_ctx.channelLive.started_at));
}
if (__VLS_ctx.channelLive?.is_live) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "viewer-chatters" },
    });
    /** @type {__VLS_StyleScopedClasses['viewer-chatters']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h3, __VLS_intrinsics.h3)({
        ...{ class: "viewer-chatters-title" },
    });
    /** @type {__VLS_StyleScopedClasses['viewer-chatters-title']} */ ;
    if (__VLS_ctx.viewerChatters.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "viewer-chatter-toolbar" },
        });
        /** @type {__VLS_StyleScopedClasses['viewer-chatter-toolbar']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
            ...{ class: "viewer-filter" },
        });
        /** @type {__VLS_StyleScopedClasses['viewer-filter']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "sr-only" },
        });
        /** @type {__VLS_StyleScopedClasses['sr-only']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
            type: "search",
            name: "viewer_filter",
            autocomplete: "off",
            autocorrect: "off",
            spellcheck: "false",
            placeholder: "Filter…",
        });
        (__VLS_ctx.viewerFilterQuery);
        __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
            ...{ class: "viewer-sort" },
        });
        /** @type {__VLS_StyleScopedClasses['viewer-sort']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "sr-only" },
        });
        /** @type {__VLS_StyleScopedClasses['sr-only']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.select, __VLS_intrinsics.select)({
            value: (__VLS_ctx.viewerSort),
            name: "viewer_sort",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "present_new",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "present_old",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "login_az",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "login_za",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "account_new",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "account_old",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "message_high",
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
            value: "message_low",
        });
    }
    if (__VLS_ctx.loadingViewerChatters) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted tiny" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
    }
    else if (__VLS_ctx.displayedViewerChatters.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
            ...{ class: "viewer-chatter-list" },
        });
        /** @type {__VLS_StyleScopedClasses['viewer-chatter-list']} */ ;
        for (const [c] of __VLS_vFor((__VLS_ctx.displayedViewerChatters))) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
                key: (c.user_twitch_id),
                ...{ class: "viewer-chatter-row" },
            });
            /** @type {__VLS_StyleScopedClasses['viewer-chatter-row']} */ ;
            const __VLS_49 = TwitchUserLink;
            // @ts-ignore
            const __VLS_50 = __VLS_asFunctionalComponent1(__VLS_49, new __VLS_49({
                login: (c.login),
                userTwitchId: (c.user_twitch_id),
                highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
                variant: "chat",
            }));
            const __VLS_51 = __VLS_50({
                login: (c.login),
                userTwitchId: (c.user_twitch_id),
                highlightChannel: (__VLS_ctx.normCh(__VLS_ctx.selectedChannel)),
                variant: "chat",
            }, ...__VLS_functionalComponentArgsRest(__VLS_50));
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "viewer-chatter-meta" },
            });
            /** @type {__VLS_StyleScopedClasses['viewer-chatter-meta']} */ ;
            (__VLS_ctx.formatPresentElapsed(c.present_since));
            if (c.message_count != null && c.message_count !== undefined) {
                __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                    ...{ class: "viewer-chatter-meta" },
                });
                /** @type {__VLS_StyleScopedClasses['viewer-chatter-meta']} */ ;
                (c.message_count);
            }
            if (c.account_created_at) {
                __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                    ...{ class: "viewer-chatter-meta" },
                });
                /** @type {__VLS_StyleScopedClasses['viewer-chatter-meta']} */ ;
                (__VLS_ctx.formatAccountDate(c.account_created_at));
            }
            // @ts-ignore
            [normCh, selectedChannel, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, channelLive, formatSessionClock, formatViewerDisplay, formatAccountDate, viewerChatters, viewerFilterQuery, viewerSort, loadingViewerChatters, displayedViewerChatters, displayedViewerChatters, formatPresentElapsed,];
        }
    }
    else if (__VLS_ctx.viewerChatters.length && !__VLS_ctx.displayedViewerChatters.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted tiny" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted tiny" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
    }
}
// @ts-ignore
[viewerChatters, displayedViewerChatters,];
var __VLS_44;
var __VLS_45;
// @ts-ignore
[];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
