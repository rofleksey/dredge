/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { useDebounceFn } from '@vueuse/core';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import TwitchUserLink from '../components/TwitchUserLink.vue';
import { ApiError, DefaultService } from '../api/generated';
import { StreamLeaderboardSort } from '../api/generated/models/StreamLeaderboardSort';
import { notify } from '../lib/notify';
defineOptions({ name: 'StreamDetailView' });
const route = useRoute();
const streamId = computed(() => Number.parseInt(String(route.params.id), 10));
const meta = ref(null);
const loadingMeta = ref(false);
const tab = ref('leaderboard');
const leaderboard = ref([]);
const loadingLb = ref(false);
const lbSort = ref(StreamLeaderboardSort.PRESENCE_DESC);
const lbFilter = ref('');
const messages = ref([]);
const loadingMsg = ref(false);
const loadingMsgMore = ref(false);
const msgCursorAt = ref();
const msgCursorId = ref();
const msgHasMore = ref(true);
const activity = ref([]);
const loadingAct = ref(false);
const loadingActMore = ref(false);
const actCursorAt = ref();
const actCursorId = ref();
const actHasMore = ref(true);
function notifyErr(e, id, title) {
    const msg = e instanceof ApiError && e.body && typeof e.body.message === 'string'
        ? e.body.message
        : 'Request failed.';
    notify({ id, type: 'error', title, description: msg });
}
async function loadMeta() {
    if (!Number.isFinite(streamId.value)) {
        meta.value = null;
        return;
    }
    loadingMeta.value = true;
    try {
        meta.value = await DefaultService.getRecordedStream({ streamId: streamId.value });
    }
    catch (e) {
        meta.value = null;
        notifyErr(e, 'stream-meta', 'Stream');
    }
    finally {
        loadingMeta.value = false;
    }
}
async function loadLeaderboard() {
    if (!Number.isFinite(streamId.value)) {
        return;
    }
    loadingLb.value = true;
    try {
        leaderboard.value = await DefaultService.getRecordedStreamLeaderboard({
            streamId: streamId.value,
            sort: lbSort.value,
            q: lbFilter.value.trim() || undefined,
        });
    }
    catch (e) {
        leaderboard.value = [];
        notifyErr(e, 'stream-lb', 'Leaderboard');
    }
    finally {
        loadingLb.value = false;
    }
}
async function loadMessages(first) {
    if (!Number.isFinite(streamId.value)) {
        return;
    }
    if (first) {
        loadingMsg.value = true;
        msgCursorAt.value = undefined;
        msgCursorId.value = undefined;
        msgHasMore.value = true;
        messages.value = [];
    }
    try {
        const list = await DefaultService.listRecordedStreamMessages({
            streamId: streamId.value,
            limit: 80,
            cursorCreatedAt: first ? undefined : msgCursorAt.value,
            cursorId: first ? undefined : msgCursorId.value,
        });
        if (first) {
            messages.value = list;
        }
        else {
            messages.value = messages.value.concat(list);
        }
        if (list.length < 80) {
            msgHasMore.value = false;
        }
        else {
            const last = list[list.length - 1];
            msgCursorAt.value = last.created_at;
            msgCursorId.value = last.id;
        }
    }
    catch (e) {
        if (first) {
            messages.value = [];
        }
        notifyErr(e, 'stream-msg', 'Messages');
    }
    finally {
        loadingMsg.value = false;
        loadingMsgMore.value = false;
    }
}
async function loadActivity(first) {
    if (!Number.isFinite(streamId.value)) {
        return;
    }
    if (first) {
        loadingAct.value = true;
        actCursorAt.value = undefined;
        actCursorId.value = undefined;
        actHasMore.value = true;
        activity.value = [];
    }
    try {
        const list = await DefaultService.listRecordedStreamActivity({
            streamId: streamId.value,
            limit: 80,
            cursorCreatedAt: first ? undefined : actCursorAt.value,
            cursorId: first ? undefined : actCursorId.value,
        });
        if (first) {
            activity.value = list;
        }
        else {
            activity.value = activity.value.concat(list);
        }
        if (list.length < 80) {
            actHasMore.value = false;
        }
        else {
            const last = list[list.length - 1];
            actCursorAt.value = last.created_at;
            actCursorId.value = last.id;
        }
    }
    catch (e) {
        if (first) {
            activity.value = [];
        }
        notifyErr(e, 'stream-act', 'Activity');
    }
    finally {
        loadingAct.value = false;
        loadingActMore.value = false;
    }
}
function formatClock(sec) {
    const s = Math.max(0, Math.floor(sec));
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const r = s % 60;
    return [h, m, r].map((n) => String(n).padStart(2, '0')).join(':');
}
function formatWhen(iso) {
    const t = Date.parse(iso);
    if (!Number.isFinite(t)) {
        return iso;
    }
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(t);
}
function activityLabel(e) {
    const ch = e.channel ? `#${e.channel}` : '';
    switch (e.event_type) {
        case 'chat_online':
            return `Joined ${ch || 'chat'}`;
        case 'chat_offline':
            return `Left ${ch || 'chat'}`;
        default:
            return e.event_type;
    }
}
async function refreshTab() {
    if (tab.value === 'leaderboard') {
        await loadLeaderboard();
    }
    else if (tab.value === 'messages') {
        await loadMessages(true);
    }
    else {
        await loadActivity(true);
    }
}
watch(streamId, () => {
    void loadMeta();
    void refreshTab();
});
watch(tab, (t) => {
    if (t === 'leaderboard' && !loadingLb.value && !leaderboard.value.length) {
        void loadLeaderboard();
    }
    if (t === 'messages' && !messages.value.length && !loadingMsg.value) {
        void loadMessages(true);
    }
    if (t === 'activity' && !activity.value.length && !loadingAct.value) {
        void loadActivity(true);
    }
});
watch(lbSort, () => {
    if (tab.value === 'leaderboard') {
        void loadLeaderboard();
    }
});
const debouncedLbFilter = useDebounceFn(() => {
    if (tab.value === 'leaderboard') {
        void loadLeaderboard();
    }
}, 320);
watch(lbFilter, () => {
    debouncedLbFilter();
});
onMounted(() => {
    void loadMeta();
    void loadLeaderboard();
});
function loadMoreMessages() {
    if (loadingMsgMore.value || !msgHasMore.value) {
        return;
    }
    loadingMsgMore.value = true;
    void loadMessages(false);
}
function loadMoreActivity() {
    if (loadingActMore.value || !actHasMore.value) {
        return;
    }
    loadingActMore.value = true;
    void loadActivity(false);
}
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "stream-detail" },
});
/** @type {__VLS_StyleScopedClasses['stream-detail']} */ ;
if (__VLS_ctx.loadingMeta) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
else if (__VLS_ctx.meta) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
        ...{ class: "stream-head" },
    });
    /** @type {__VLS_StyleScopedClasses['stream-head']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h1, __VLS_intrinsics.h1)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "live-pill" },
        ...{ class: ({ 'live-pill--off': __VLS_ctx.meta.ended_at }) },
    });
    /** @type {__VLS_StyleScopedClasses['live-pill']} */ ;
    /** @type {__VLS_StyleScopedClasses['live-pill--off']} */ ;
    (__VLS_ctx.meta.ended_at ? 'Past' : 'Live');
    (__VLS_ctx.meta.channel_login);
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "meta-line" },
    });
    /** @type {__VLS_StyleScopedClasses['meta-line']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
    (__VLS_ctx.meta.title?.trim() || '—');
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    (__VLS_ctx.meta.game_name?.trim() || '—');
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted small" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    /** @type {__VLS_StyleScopedClasses['small']} */ ;
    (__VLS_ctx.formatWhen(__VLS_ctx.meta.started_at));
    if (__VLS_ctx.meta.ended_at) {
        (__VLS_ctx.formatWhen(__VLS_ctx.meta.ended_at));
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.nav, __VLS_intrinsics.nav)({
        ...{ class: "tabs" },
    });
    /** @type {__VLS_StyleScopedClasses['tabs']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingMeta))
                    return;
                if (!(__VLS_ctx.meta))
                    return;
                __VLS_ctx.tab = 'leaderboard';
                // @ts-ignore
                [loadingMeta, meta, meta, meta, meta, meta, meta, meta, meta, meta, formatWhen, formatWhen, tab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.tab === 'leaderboard' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingMeta))
                    return;
                if (!(__VLS_ctx.meta))
                    return;
                __VLS_ctx.tab = 'messages';
                // @ts-ignore
                [tab, tab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.tab === 'messages' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingMeta))
                    return;
                if (!(__VLS_ctx.meta))
                    return;
                __VLS_ctx.tab = 'activity';
                // @ts-ignore
                [tab, tab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.tab === 'activity' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'leaderboard') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "toolbar" },
    });
    /** @type {__VLS_StyleScopedClasses['toolbar']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
        ...{ class: "grow" },
    });
    /** @type {__VLS_StyleScopedClasses['grow']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "sr-only" },
    });
    /** @type {__VLS_StyleScopedClasses['sr-only']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        type: "search",
        placeholder: "Filter login…",
    });
    (__VLS_ctx.lbFilter);
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "sr-only" },
    });
    /** @type {__VLS_StyleScopedClasses['sr-only']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.select, __VLS_intrinsics.select)({
        value: (__VLS_ctx.lbSort),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.PRESENCE_DESC),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.PRESENCE_ASC),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.MESSAGES_DESC),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.MESSAGES_ASC),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.LOGIN_AZ),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.LOGIN_ZA),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.ACCOUNT_NEW),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
        value: (__VLS_ctx.StreamLeaderboardSort.ACCOUNT_OLD),
    });
    if (__VLS_ctx.loadingLb) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    else if (__VLS_ctx.leaderboard.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.table, __VLS_intrinsics.table)({
            ...{ class: "lb-table" },
        });
        /** @type {__VLS_StyleScopedClasses['lb-table']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.thead, __VLS_intrinsics.thead)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.tr, __VLS_intrinsics.tr)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.tbody, __VLS_intrinsics.tbody)({});
        for (const [row] of __VLS_vFor((__VLS_ctx.leaderboard))) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.tr, __VLS_intrinsics.tr)({
                key: (row.user_twitch_id),
            });
            __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
            const __VLS_0 = TwitchUserLink;
            // @ts-ignore
            const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
                login: (row.login),
                userTwitchId: (row.user_twitch_id),
                highlightChannel: (__VLS_ctx.meta.channel_login),
                variant: "chat",
            }));
            const __VLS_2 = __VLS_1({
                login: (row.login),
                userTwitchId: (row.user_twitch_id),
                highlightChannel: (__VLS_ctx.meta.channel_login),
                variant: "chat",
            }, ...__VLS_functionalComponentArgsRest(__VLS_1));
            __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
            (__VLS_ctx.formatClock(row.presence_seconds));
            __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
            (row.message_count);
            // @ts-ignore
            [meta, tab, tab, lbFilter, lbSort, StreamLeaderboardSort, StreamLeaderboardSort, StreamLeaderboardSort, StreamLeaderboardSort, StreamLeaderboardSort, StreamLeaderboardSort, StreamLeaderboardSort, StreamLeaderboardSort, loadingLb, leaderboard, leaderboard, formatClock,];
        }
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'messages') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    if (__VLS_ctx.loadingMsg) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    else if (__VLS_ctx.messages.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
            ...{ class: "msg-list" },
        });
        /** @type {__VLS_StyleScopedClasses['msg-list']} */ ;
        for (const [m] of __VLS_vFor((__VLS_ctx.messages))) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
                key: (m.id),
            });
            const __VLS_5 = ChatMessageLine;
            // @ts-ignore
            const __VLS_6 = __VLS_asFunctionalComponent1(__VLS_5, new __VLS_5({
                user: (m.user),
                message: (m.message),
                keyword: (m.keyword_match),
                userMarked: (m.chatter_marked),
                fromSent: (m.source === 'sent'),
                badgeTags: (m.badge_tags),
                showTimestamp: (true),
                createdAt: (m.created_at),
                chatterUserId: (m.chatter_user_id ?? null),
                highlightChannel: (__VLS_ctx.meta.channel_login),
            }));
            const __VLS_7 = __VLS_6({
                user: (m.user),
                message: (m.message),
                keyword: (m.keyword_match),
                userMarked: (m.chatter_marked),
                fromSent: (m.source === 'sent'),
                badgeTags: (m.badge_tags),
                showTimestamp: (true),
                createdAt: (m.created_at),
                chatterUserId: (m.chatter_user_id ?? null),
                highlightChannel: (__VLS_ctx.meta.channel_login),
            }, ...__VLS_functionalComponentArgsRest(__VLS_6));
            // @ts-ignore
            [meta, tab, loadingMsg, messages, messages,];
        }
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    if (__VLS_ctx.messages.length && __VLS_ctx.msgHasMore) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "more-row" },
        });
        /** @type {__VLS_StyleScopedClasses['more-row']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (__VLS_ctx.loadMoreMessages) },
            type: "button",
            ...{ class: "btn-more" },
            disabled: (__VLS_ctx.loadingMsgMore),
        });
        /** @type {__VLS_StyleScopedClasses['btn-more']} */ ;
        (__VLS_ctx.loadingMsgMore ? 'Loading…' : 'Load more');
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'activity') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    if (__VLS_ctx.loadingAct) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    else if (__VLS_ctx.activity.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
            ...{ class: "activity-list" },
        });
        /** @type {__VLS_StyleScopedClasses['activity-list']} */ ;
        for (const [e] of __VLS_vFor((__VLS_ctx.activity))) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
                key: (e.id),
            });
            __VLS_asFunctionalElement1(__VLS_intrinsics.time, __VLS_intrinsics.time)({
                ...{ class: "act-time" },
                datetime: (e.created_at),
            });
            /** @type {__VLS_StyleScopedClasses['act-time']} */ ;
            (__VLS_ctx.formatWhen(e.created_at));
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "act-user" },
            });
            /** @type {__VLS_StyleScopedClasses['act-user']} */ ;
            (e.username);
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "act-body" },
            });
            /** @type {__VLS_StyleScopedClasses['act-body']} */ ;
            (__VLS_ctx.activityLabel(e));
            // @ts-ignore
            [formatWhen, tab, messages, msgHasMore, loadMoreMessages, loadingMsgMore, loadingMsgMore, loadingAct, activity, activity, activityLabel,];
        }
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    if (__VLS_ctx.activity.length && __VLS_ctx.actHasMore) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "more-row" },
        });
        /** @type {__VLS_StyleScopedClasses['more-row']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (__VLS_ctx.loadMoreActivity) },
            type: "button",
            ...{ class: "btn-more" },
            disabled: (__VLS_ctx.loadingActMore),
        });
        /** @type {__VLS_StyleScopedClasses['btn-more']} */ ;
        (__VLS_ctx.loadingActMore ? 'Loading…' : 'Load more');
    }
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
// @ts-ignore
[activity, actHasMore, loadMoreActivity, loadingActMore, loadingActMore,];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
