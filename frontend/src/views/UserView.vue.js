/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import * as echarts from 'echarts';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import { ApiError, ChatHistoryEntry, DefaultService } from '../api/generated';
import { notify } from '../lib/notify';
defineOptions({ name: 'UserView' });
const route = useRoute();
const profile = ref(null);
const notFound = ref(false);
const messages = ref([]);
const loadingProfile = ref(true);
const loadingMessages = ref(true);
const loadingMore = ref(false);
const togglingMarked = ref(false);
const togglingSus = ref(false);
const manualSusNote = ref('');
const followingQuery = ref('');
const userTab = ref('overview');
const activity = ref([]);
const loadingActivity = ref(false);
const loadingActivityMore = ref(false);
const timelineSegments = ref([]);
const loadingTimeline = ref(false);
const chartEl = ref(null);
let chart = null;
const userId = computed(() => {
    const raw = route.params.id;
    const s = Array.isArray(raw) ? raw[0] : raw;
    const n = Number(s);
    return Number.isFinite(n) ? n : NaN;
});
async function loadProfile() {
    if (!Number.isFinite(userId.value)) {
        notFound.value = true;
        profile.value = null;
        loadingProfile.value = false;
        return;
    }
    loadingProfile.value = true;
    notFound.value = false;
    try {
        profile.value = await DefaultService.getTwitchUserProfile({ requestBody: { id: userId.value } });
    }
    catch (e) {
        profile.value = null;
        if (e instanceof ApiError && e.status === 404) {
            notFound.value = true;
        }
        else {
            notify({
                id: 'user-profile',
                type: 'error',
                title: 'User',
                description: 'Could not load profile.',
            });
        }
    }
    finally {
        loadingProfile.value = false;
    }
}
function mergeProfileFromTwitchUser(u) {
    if (!profile.value) {
        return;
    }
    profile.value = {
        ...profile.value,
        monitored: u.monitored,
        marked: u.marked,
        is_sus: u.is_sus,
        sus_type: u.sus_type,
        sus_description: u.sus_description,
        sus_auto_suppressed: u.sus_auto_suppressed,
    };
}
async function markSuspicious() {
    if (!profile.value || togglingSus.value) {
        return;
    }
    togglingSus.value = true;
    const desc = manualSusNote.value.trim() || 'Marked manually';
    try {
        const u = await DefaultService.updateTwitchUser({
            requestBody: {
                id: profile.value.id,
                is_sus: true,
                sus_type: 'manual',
                sus_description: desc,
                sus_auto_suppressed: false,
            },
        });
        mergeProfileFromTwitchUser(u);
        manualSusNote.value = '';
    }
    catch {
        notify({
            id: 'user-sus-mark',
            type: 'error',
            title: 'User',
            description: 'Could not update suspicion.',
        });
    }
    finally {
        togglingSus.value = false;
    }
}
async function clearSuspicion() {
    if (!profile.value || togglingSus.value) {
        return;
    }
    togglingSus.value = true;
    try {
        const u = await DefaultService.updateTwitchUser({
            requestBody: {
                id: profile.value.id,
                is_sus: false,
                sus_type: null,
                sus_description: null,
                sus_auto_suppressed: true,
            },
        });
        mergeProfileFromTwitchUser(u);
    }
    catch {
        notify({
            id: 'user-sus-clear',
            type: 'error',
            title: 'User',
            description: 'Could not clear suspicion.',
        });
    }
    finally {
        togglingSus.value = false;
    }
}
async function allowAutoSuspicionAgain() {
    if (!profile.value || togglingSus.value) {
        return;
    }
    togglingSus.value = true;
    try {
        const u = await DefaultService.updateTwitchUser({
            requestBody: {
                id: profile.value.id,
                sus_auto_suppressed: false,
            },
        });
        mergeProfileFromTwitchUser(u);
    }
    catch {
        notify({
            id: 'user-sus-auto',
            type: 'error',
            title: 'User',
            description: 'Could not update suppression flag.',
        });
    }
    finally {
        togglingSus.value = false;
    }
}
async function toggleMarked() {
    if (!profile.value || togglingMarked.value) {
        return;
    }
    togglingMarked.value = true;
    try {
        const u = await DefaultService.updateTwitchUser({
            requestBody: { id: profile.value.id, marked: !profile.value.marked },
        });
        mergeProfileFromTwitchUser(u);
    }
    catch {
        notify({
            id: 'user-marked',
            type: 'error',
            title: 'User',
            description: 'Could not update marked flag.',
        });
    }
    finally {
        togglingMarked.value = false;
    }
}
function buildMsgQuery(appendCursor) {
    const q = {
        limit: 80,
        chatterUserId: userId.value,
    };
    if (appendCursor && messages.value.length) {
        const last = messages.value[messages.value.length - 1];
        q.cursorCreatedAt = last.created_at;
        q.cursorId = last.id;
    }
    return q;
}
async function loadMessages() {
    if (!Number.isFinite(userId.value)) {
        messages.value = [];
        loadingMessages.value = false;
        return;
    }
    loadingMessages.value = true;
    try {
        messages.value = await DefaultService.listTwitchMessages(buildMsgQuery(false));
    }
    catch {
        messages.value = [];
        notify({
            id: 'user-msgs',
            type: 'error',
            title: 'User',
            description: 'Could not load messages.',
        });
    }
    finally {
        loadingMessages.value = false;
    }
}
async function loadMore() {
    if (!messages.value.length || loadingMore.value || !Number.isFinite(userId.value)) {
        return;
    }
    loadingMore.value = true;
    try {
        const next = await DefaultService.listTwitchMessages(buildMsgQuery(true));
        const seen = new Set(messages.value.map((m) => m.id));
        for (const m of next) {
            if (!seen.has(m.id)) {
                messages.value.push(m);
                seen.add(m.id);
            }
        }
    }
    catch {
        notify({
            id: 'user-msgs-more',
            type: 'error',
            title: 'User',
            description: 'Could not load more messages.',
        });
    }
    finally {
        loadingMore.value = false;
    }
}
function buildActivityQuery(appendCursor) {
    const q = {
        id: userId.value,
        limit: 50,
        cursor_created_at: undefined,
        cursor_id: undefined,
    };
    if (appendCursor && activity.value.length) {
        const last = activity.value[activity.value.length - 1];
        q.cursor_created_at = last.created_at;
        q.cursor_id = last.id;
    }
    return q;
}
async function loadActivityFirst() {
    if (!Number.isFinite(userId.value)) {
        activity.value = [];
        return;
    }
    loadingActivity.value = true;
    try {
        activity.value = await DefaultService.listTwitchUserActivity({
            requestBody: buildActivityQuery(false),
        });
    }
    catch {
        activity.value = [];
        notify({
            id: 'user-activity',
            type: 'error',
            title: 'User',
            description: 'Could not load activity.',
        });
    }
    finally {
        loadingActivity.value = false;
    }
}
async function loadActivityMore() {
    if (!activity.value.length || loadingActivityMore.value || !Number.isFinite(userId.value)) {
        return;
    }
    loadingActivityMore.value = true;
    try {
        const next = await DefaultService.listTwitchUserActivity({
            requestBody: buildActivityQuery(true),
        });
        const seen = new Set(activity.value.map((e) => e.id));
        for (const e of next) {
            if (!seen.has(e.id)) {
                activity.value.push(e);
                seen.add(e.id);
            }
        }
    }
    catch {
        notify({
            id: 'user-activity-more',
            type: 'error',
            title: 'User',
            description: 'Could not load more activity.',
        });
    }
    finally {
        loadingActivityMore.value = false;
    }
}
async function loadTimeline() {
    if (!Number.isFinite(userId.value)) {
        timelineSegments.value = [];
        return;
    }
    loadingTimeline.value = true;
    try {
        const to = new Date();
        const from = new Date(to.getTime() - 7 * 24 * 60 * 60 * 1000);
        timelineSegments.value = await DefaultService.getTwitchUserActivityTimeline({
            requestBody: {
                id: userId.value,
                from: from.toISOString(),
                to: to.toISOString(),
            },
        });
    }
    catch {
        timelineSegments.value = [];
        notify({
            id: 'user-timeline',
            type: 'error',
            title: 'User',
            description: 'Could not load activity timeline.',
        });
    }
    finally {
        loadingTimeline.value = false;
    }
}
function renderTimelineChart() {
    if (!chartEl.value) {
        return;
    }
    if (chart) {
        chart.dispose();
        chart = null;
    }
    const segs = timelineSegments.value;
    if (!segs.length) {
        return;
    }
    const channels = [...new Set(segs.map((s) => s.channel_login))].sort((a, b) => a.localeCompare(b));
    const data = segs.map((s) => {
        const yi = channels.indexOf(s.channel_login);
        const t0 = new Date(s.start).getTime();
        const t1 = new Date(s.end).getTime();
        return [yi, t0, t1];
    });
    chart = echarts.init(chartEl.value, undefined, { renderer: 'canvas' });
    chart.setOption({
        tooltip: {
            trigger: 'item',
            formatter: (p) => {
                const d = p;
                const v = d.value;
                if (!v || v.length < 3) {
                    return '';
                }
                const ch = channels[v[0]] ?? '';
                const a = new Date(v[1]).toLocaleString();
                const b = new Date(v[2]).toLocaleString();
                return `${ch}<br/>${a} — ${b}`;
            },
        },
        grid: { left: 140, right: 24, top: 16, bottom: 48 },
        xAxis: {
            type: 'time',
            scale: true,
            axisLabel: { fontSize: 10, color: '#aaa' },
        },
        yAxis: {
            type: 'category',
            data: channels,
            axisLabel: { fontSize: 11, color: '#ccc' },
            inverse: true,
        },
        dataZoom: [{ type: 'inside' }, { type: 'slider', height: 18 }],
        series: [
            {
                type: 'custom',
                renderItem(_params, api) {
                    const yIndex = api.value(0);
                    const t0 = api.value(1);
                    const t1 = api.value(2);
                    const start = api.coord([t0, yIndex]);
                    const end = api.coord([t1, yIndex]);
                    const sy = api.size([0, 1])[1] ?? 12;
                    const h = Math.max(6, sy * 0.55);
                    return {
                        type: 'rect',
                        shape: {
                            x: start[0],
                            y: start[1] - h / 2,
                            width: Math.max(1, end[0] - start[0]),
                            height: h,
                        },
                        style: {
                            fill: 'rgba(145, 71, 255, 0.45)',
                            stroke: 'rgba(145, 71, 255, 0.85)',
                            lineWidth: 1,
                        },
                    };
                },
                dimensions: ['y', 't0', 't1'],
                encode: { x: [1, 2], y: 0 },
                data,
            },
        ],
    });
}
function onResize() {
    chart?.resize();
}
watch(userTab, async (t) => {
    if (t === 'activity' && !loadingActivity.value && activity.value.length === 0 && Number.isFinite(userId.value)) {
        await loadActivityFirst();
    }
    if (t === 'graphs') {
        await loadTimeline();
        await nextTick();
        renderTimelineChart();
    }
});
watch(() => route.params.id, async () => {
    userTab.value = 'overview';
    followingQuery.value = '';
    activity.value = [];
    timelineSegments.value = [];
    if (chart) {
        chart.dispose();
        chart = null;
    }
    await loadProfile();
    await loadMessages();
});
onMounted(async () => {
    window.addEventListener('resize', onResize);
    await loadProfile();
    await loadMessages();
});
onBeforeUnmount(() => {
    window.removeEventListener('resize', onResize);
    if (chart) {
        chart.dispose();
        chart = null;
    }
});
function rowBadges(m) {
    return [...(m.badge_tags ?? [])];
}
function activityLabel(e) {
    const ch = e.channel ? `#${e.channel}` : '';
    switch (e.event_type) {
        case 'chat_online':
            return ch ? `Online in ${ch}` : 'Online in chat';
        case 'chat_offline':
            return ch ? `Offline in ${ch}` : 'Offline in chat';
        default:
            return e.event_type;
    }
}
function formatWhen(iso) {
    const t = Date.parse(iso);
    if (!Number.isFinite(t)) {
        return iso;
    }
    return new Intl.DateTimeFormat(undefined, {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    }).format(t);
}
const accountCreatedLabel = computed(() => {
    const raw = profile.value?.account_created_at;
    if (!raw) {
        return null;
    }
    return formatWhen(raw);
});
const sortedFollowedChannels = computed(() => {
    const p = profile.value;
    if (!p?.followed_channels?.length) {
        return [];
    }
    const q = followingQuery.value.trim().toLowerCase();
    let rows = [...p.followed_channels];
    if (q) {
        rows = rows.filter((r) => r.channel_login.toLowerCase().includes(q));
    }
    rows.sort((a, b) => {
        if (a.on_blacklist !== b.on_blacklist) {
            return a.on_blacklist ? -1 : 1;
        }
        return a.channel_login.localeCompare(b.channel_login);
    });
    return rows;
});
function formatPresenceWeek(sec) {
    if (!Number.isFinite(sec) || sec <= 0) {
        return '—';
    }
    const h = Math.floor(sec / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    if (h > 0) {
        return `${h}h ${m}m`;
    }
    if (m > 0) {
        return `${m}m ${s}s`;
    }
    return `${s}s`;
}
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "page user-page" },
});
/** @type {__VLS_StyleScopedClasses['page']} */ ;
/** @type {__VLS_StyleScopedClasses['user-page']} */ ;
if (__VLS_ctx.loadingProfile) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
else if (__VLS_ctx.notFound) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
else if (__VLS_ctx.profile) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
        ...{ class: "page-head" },
    });
    /** @type {__VLS_StyleScopedClasses['page-head']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h1, __VLS_intrinsics.h1)({
        ...{ class: "page-title" },
    });
    /** @type {__VLS_StyleScopedClasses['page-title']} */ ;
    (__VLS_ctx.profile.username);
    __VLS_asFunctionalElement1(__VLS_intrinsics.nav, __VLS_intrinsics.nav)({
        ...{ class: "user-tabs" },
        'aria-label': "User sections",
    });
    /** @type {__VLS_StyleScopedClasses['user-tabs']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingProfile))
                    return;
                if (!!(__VLS_ctx.notFound))
                    return;
                if (!(__VLS_ctx.profile))
                    return;
                __VLS_ctx.userTab = 'overview';
                // @ts-ignore
                [loadingProfile, notFound, profile, profile, userTab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.userTab === 'overview' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingProfile))
                    return;
                if (!!(__VLS_ctx.notFound))
                    return;
                if (!(__VLS_ctx.profile))
                    return;
                __VLS_ctx.userTab = 'following';
                // @ts-ignore
                [userTab, userTab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.userTab === 'following' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingProfile))
                    return;
                if (!!(__VLS_ctx.notFound))
                    return;
                if (!(__VLS_ctx.profile))
                    return;
                __VLS_ctx.userTab = 'messages';
                // @ts-ignore
                [userTab, userTab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.userTab === 'messages' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingProfile))
                    return;
                if (!!(__VLS_ctx.notFound))
                    return;
                if (!(__VLS_ctx.profile))
                    return;
                __VLS_ctx.userTab = 'activity';
                // @ts-ignore
                [userTab, userTab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.userTab === 'activity' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!!(__VLS_ctx.loadingProfile))
                    return;
                if (!!(__VLS_ctx.notFound))
                    return;
                if (!(__VLS_ctx.profile))
                    return;
                __VLS_ctx.userTab = 'graphs';
                // @ts-ignore
                [userTab, userTab,];
            } },
        type: "button",
        ...{ class: ({ active: __VLS_ctx.userTab === 'graphs' }) },
    });
    /** @type {__VLS_StyleScopedClasses['active']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.userTab === 'overview') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    if (__VLS_ctx.profile.is_sus || __VLS_ctx.profile.sus_description) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "sus-banner" },
        });
        /** @type {__VLS_StyleScopedClasses['sus-banner']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "sus-banner-head" },
        });
        /** @type {__VLS_StyleScopedClasses['sus-banner-head']} */ ;
        if (__VLS_ctx.profile.is_sus) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "sus-badge" },
            });
            /** @type {__VLS_StyleScopedClasses['sus-badge']} */ ;
        }
        if (__VLS_ctx.profile.sus_type) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "muted sus-type" },
            });
            /** @type {__VLS_StyleScopedClasses['muted']} */ ;
            /** @type {__VLS_StyleScopedClasses['sus-type']} */ ;
            (__VLS_ctx.profile.sus_type);
        }
        if (__VLS_ctx.profile.sus_description) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
                ...{ class: "sus-desc" },
            });
            /** @type {__VLS_StyleScopedClasses['sus-desc']} */ ;
            (__VLS_ctx.profile.sus_description);
        }
        if (__VLS_ctx.profile.sus_auto_suppressed) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
                ...{ class: "muted sus-hint" },
            });
            /** @type {__VLS_StyleScopedClasses['muted']} */ ;
            /** @type {__VLS_StyleScopedClasses['sus-hint']} */ ;
        }
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "sus-actions" },
    });
    /** @type {__VLS_StyleScopedClasses['sus-actions']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
        ...{ class: "sus-note" },
    });
    /** @type {__VLS_StyleScopedClasses['sus-note']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "label" },
    });
    /** @type {__VLS_StyleScopedClasses['label']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        value: (__VLS_ctx.manualSusNote),
        type: "text",
        autocomplete: "off",
        placeholder: "Reason or label",
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "sus-buttons" },
    });
    /** @type {__VLS_StyleScopedClasses['sus-buttons']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (__VLS_ctx.markSuspicious) },
        type: "button",
        ...{ class: "btn-sus" },
        disabled: (__VLS_ctx.togglingSus),
    });
    /** @type {__VLS_StyleScopedClasses['btn-sus']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (__VLS_ctx.clearSuspicion) },
        type: "button",
        ...{ class: "btn-sus-secondary" },
        disabled: (__VLS_ctx.togglingSus),
    });
    /** @type {__VLS_StyleScopedClasses['btn-sus-secondary']} */ ;
    if (__VLS_ctx.profile.sus_auto_suppressed) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (__VLS_ctx.allowAutoSuspicionAgain) },
            type: "button",
            ...{ class: "btn-sus-secondary" },
            disabled: (__VLS_ctx.togglingSus),
        });
        /** @type {__VLS_StyleScopedClasses['btn-sus-secondary']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.dl, __VLS_intrinsics.dl)({
        ...{ class: "meta" },
    });
    /** @type {__VLS_StyleScopedClasses['meta']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.profile.id);
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.profile.monitored ? 'Yes' : 'No');
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (__VLS_ctx.toggleMarked) },
        type: "button",
        ...{ class: "btn-toggle" },
        disabled: (__VLS_ctx.togglingMarked),
    });
    /** @type {__VLS_StyleScopedClasses['btn-toggle']} */ ;
    (__VLS_ctx.profile.marked ? 'Yes (click to clear)' : 'No (click to mark)');
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.profile.is_sus ? 'Yes' : 'No');
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.formatPresenceWeek(__VLS_ctx.profile.presence_seconds_this_week));
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.profile.message_count);
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dt, __VLS_intrinsics.dt)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.dd, __VLS_intrinsics.dd)({});
    (__VLS_ctx.accountCreatedLabel ?? '—');
    if (__VLS_ctx.profile.followed_monitored_channels?.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "follow-block" },
        });
        /** @type {__VLS_StyleScopedClasses['follow-block']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.h3, __VLS_intrinsics.h3)({
            ...{ class: "sub-title" },
        });
        /** @type {__VLS_StyleScopedClasses['sub-title']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
            ...{ class: "follow-list" },
        });
        /** @type {__VLS_StyleScopedClasses['follow-list']} */ ;
        for (const [f] of __VLS_vFor((__VLS_ctx.profile.followed_monitored_channels))) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
                key: (f.channel_id),
            });
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "ch" },
            });
            /** @type {__VLS_StyleScopedClasses['ch']} */ ;
            (f.channel_login);
            if (f.followed_at) {
                __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                    ...{ class: "muted" },
                });
                /** @type {__VLS_StyleScopedClasses['muted']} */ ;
                (__VLS_ctx.formatWhen(f.followed_at));
            }
            else {
                __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                    ...{ class: "muted" },
                });
                /** @type {__VLS_StyleScopedClasses['muted']} */ ;
            }
            // @ts-ignore
            [profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, profile, userTab, userTab, manualSusNote, markSuspicious, togglingSus, togglingSus, togglingSus, clearSuspicion, allowAutoSuspicionAgain, toggleMarked, togglingMarked, formatPresenceWeek, accountCreatedLabel, formatWhen,];
        }
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel following-panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.userTab === 'following') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    /** @type {__VLS_StyleScopedClasses['following-panel']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({
        ...{ class: "section-title" },
    });
    /** @type {__VLS_StyleScopedClasses['section-title']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted hint" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    /** @type {__VLS_StyleScopedClasses['hint']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
        ...{ class: "follow-filter" },
    });
    /** @type {__VLS_StyleScopedClasses['follow-filter']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "label" },
    });
    /** @type {__VLS_StyleScopedClasses['label']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        type: "search",
        autocomplete: "off",
        placeholder: "Channel login contains…",
    });
    (__VLS_ctx.followingQuery);
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "table-wrap" },
    });
    /** @type {__VLS_StyleScopedClasses['table-wrap']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.table, __VLS_intrinsics.table)({
        ...{ class: "follow-table" },
    });
    /** @type {__VLS_StyleScopedClasses['follow-table']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.thead, __VLS_intrinsics.thead)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.tr, __VLS_intrinsics.tr)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.tbody, __VLS_intrinsics.tbody)({});
    for (const [row] of __VLS_vFor((__VLS_ctx.sortedFollowedChannels))) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.tr, __VLS_intrinsics.tr)({
            key: (row.channel_id),
            ...{ class: ({ 'bl-row': row.on_blacklist }) },
        });
        /** @type {__VLS_StyleScopedClasses['bl-row']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "ch" },
        });
        /** @type {__VLS_StyleScopedClasses['ch']} */ ;
        (row.channel_login);
        if (row.on_blacklist) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "tag-bl" },
            });
            /** @type {__VLS_StyleScopedClasses['tag-bl']} */ ;
        }
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        (row.followed_at ? __VLS_ctx.formatWhen(row.followed_at) : '—');
        __VLS_asFunctionalElement1(__VLS_intrinsics.td)({});
        // @ts-ignore
        [userTab, formatWhen, followingQuery, sortedFollowedChannels,];
    }
    if (!__VLS_ctx.sortedFollowedChannels.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.userTab === 'messages') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({
        ...{ class: "section-title" },
    });
    /** @type {__VLS_StyleScopedClasses['section-title']} */ ;
    if (__VLS_ctx.loadingMessages) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
            ...{ class: "lines" },
        });
        /** @type {__VLS_StyleScopedClasses['lines']} */ ;
        for (const [m] of __VLS_vFor((__VLS_ctx.messages))) {
            const __VLS_0 = ChatMessageLine;
            // @ts-ignore
            const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
                key: (m.id),
                user: (m.user),
                message: (m.message),
                keyword: (m.keyword_match),
                fromSent: (m.source === __VLS_ctx.ChatHistoryEntry.source.SENT),
                badgeTags: (__VLS_ctx.rowBadges(m)),
                showTimestamp: (true),
                createdAt: (m.created_at),
                chatterUserId: (m.chatter_user_id ?? undefined),
                userMarked: (m.chatter_marked),
                userIsSus: (m.chatter_is_sus),
                showChannel: true,
                channelLogin: (m.channel),
            }));
            const __VLS_2 = __VLS_1({
                key: (m.id),
                user: (m.user),
                message: (m.message),
                keyword: (m.keyword_match),
                fromSent: (m.source === __VLS_ctx.ChatHistoryEntry.source.SENT),
                badgeTags: (__VLS_ctx.rowBadges(m)),
                showTimestamp: (true),
                createdAt: (m.created_at),
                chatterUserId: (m.chatter_user_id ?? undefined),
                userMarked: (m.chatter_marked),
                userIsSus: (m.chatter_is_sus),
                showChannel: true,
                channelLogin: (m.channel),
            }, ...__VLS_functionalComponentArgsRest(__VLS_1));
            // @ts-ignore
            [userTab, sortedFollowedChannels, loadingMessages, messages, ChatHistoryEntry, rowBadges,];
        }
    }
    if (!__VLS_ctx.loadingMessages && __VLS_ctx.messages.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "more-row" },
        });
        /** @type {__VLS_StyleScopedClasses['more-row']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (__VLS_ctx.loadMore) },
            type: "button",
            ...{ class: "btn-more" },
            disabled: (__VLS_ctx.loadingMore),
        });
        /** @type {__VLS_StyleScopedClasses['btn-more']} */ ;
        (__VLS_ctx.loadingMore ? 'Loading…' : 'Load more');
    }
    if (!__VLS_ctx.loadingMessages && !__VLS_ctx.messages.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.userTab === 'activity') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({
        ...{ class: "section-title" },
    });
    /** @type {__VLS_StyleScopedClasses['section-title']} */ ;
    if (__VLS_ctx.loadingActivity) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    else {
        __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
            ...{ class: "activity-list" },
        });
        /** @type {__VLS_StyleScopedClasses['activity-list']} */ ;
        for (const [e] of __VLS_vFor((__VLS_ctx.activity))) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
                key: (e.id),
            });
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "act-ts" },
            });
            /** @type {__VLS_StyleScopedClasses['act-ts']} */ ;
            (__VLS_ctx.formatWhen(e.created_at));
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "act-body" },
            });
            /** @type {__VLS_StyleScopedClasses['act-body']} */ ;
            (__VLS_ctx.activityLabel(e));
            // @ts-ignore
            [userTab, formatWhen, loadingMessages, loadingMessages, messages, messages, loadMore, loadingMore, loadingMore, loadingActivity, activity, activityLabel,];
        }
    }
    if (!__VLS_ctx.loadingActivity && __VLS_ctx.activity.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
            ...{ class: "more-row" },
        });
        /** @type {__VLS_StyleScopedClasses['more-row']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (__VLS_ctx.loadActivityMore) },
            type: "button",
            ...{ class: "btn-more" },
            disabled: (__VLS_ctx.loadingActivityMore),
        });
        /** @type {__VLS_StyleScopedClasses['btn-more']} */ ;
        (__VLS_ctx.loadingActivityMore ? 'Loading…' : 'Load more');
    }
    if (!__VLS_ctx.loadingActivity && !__VLS_ctx.activity.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
        ...{ class: "panel graphs-panel" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.userTab === 'graphs') }, null, null);
    /** @type {__VLS_StyleScopedClasses['panel']} */ ;
    /** @type {__VLS_StyleScopedClasses['graphs-panel']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({
        ...{ class: "section-title" },
    });
    /** @type {__VLS_StyleScopedClasses['section-title']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted hint" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    /** @type {__VLS_StyleScopedClasses['hint']} */ ;
    if (__VLS_ctx.loadingTimeline) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.div)({
        ref: "chartEl",
        ...{ class: "chart-host" },
    });
    __VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (!__VLS_ctx.loadingTimeline) }, null, null);
    /** @type {__VLS_StyleScopedClasses['chart-host']} */ ;
    if (!__VLS_ctx.loadingTimeline && !__VLS_ctx.timelineSegments.length) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    }
}
// @ts-ignore
[userTab, loadingActivity, loadingActivity, activity, activity, loadActivityMore, loadingActivityMore, loadingActivityMore, loadingTimeline, loadingTimeline, loadingTimeline, timelineSegments,];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
