/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { computed, onMounted, ref } from 'vue';
import AppModal from '../components/AppModal.vue';
import SubmitButton from '../components/SubmitButton.vue';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import { ApiError, ChatHistoryEntry, DefaultService } from '../api/generated';
import { notify } from '../lib/notify';
defineOptions({ name: 'MessagesView' });
const loading = ref(false);
const loadingMore = ref(false);
const filtersApplying = ref(false);
const messages = ref([]);
const filtersOpen = ref(false);
const totalCount = ref(null);
const applied = ref({
    username: '',
    text: '',
    channel: '',
    createdFrom: '',
    createdTo: '',
});
const draft = ref({ ...applied.value });
const filterSummary = computed(() => {
    const parts = [];
    if (applied.value.username) {
        parts.push(`user: ${applied.value.username}`);
    }
    if (applied.value.text) {
        parts.push(`text: ${applied.value.text}`);
    }
    if (applied.value.channel) {
        parts.push(`#${applied.value.channel}`);
    }
    if (applied.value.createdFrom) {
        parts.push(`from ${applied.value.createdFrom}`);
    }
    if (applied.value.createdTo) {
        parts.push(`to ${applied.value.createdTo}`);
    }
    return parts.length ? parts.join(' · ') : 'None';
});
function toIsoFromLocal(dtLocal) {
    if (!dtLocal.trim()) {
        return undefined;
    }
    const d = new Date(dtLocal);
    if (!Number.isFinite(d.getTime())) {
        return undefined;
    }
    return d.toISOString();
}
function buildQuery(appendCursor) {
    const q = {
        limit: 80,
        username: applied.value.username.trim() || undefined,
        text: applied.value.text.trim() || undefined,
        channel: applied.value.channel.replace(/^#/, '').trim().toLowerCase() || undefined,
        createdFrom: toIsoFromLocal(applied.value.createdFrom),
        createdTo: toIsoFromLocal(applied.value.createdTo),
    };
    if (appendCursor && messages.value.length) {
        const last = messages.value[messages.value.length - 1];
        q.cursorCreatedAt = last.created_at;
        q.cursorId = last.id;
    }
    return q;
}
function countQuery() {
    return {
        username: applied.value.username.trim() || undefined,
        text: applied.value.text.trim() || undefined,
        channel: applied.value.channel.replace(/^#/, '').trim().toLowerCase() || undefined,
        createdFrom: toIsoFromLocal(applied.value.createdFrom),
        createdTo: toIsoFromLocal(applied.value.createdTo),
    };
}
async function fetchFirst() {
    loading.value = true;
    try {
        const [list, cnt] = await Promise.all([
            DefaultService.listTwitchMessages(buildQuery(false)),
            DefaultService.countTwitchMessages(countQuery()),
        ]);
        messages.value = list;
        totalCount.value = cnt.total;
    }
    catch (e) {
        messages.value = [];
        totalCount.value = null;
        const msg = e instanceof ApiError && e.body && typeof e.body.message === 'string'
            ? e.body.message
            : 'Could not load messages.';
        notify({
            id: 'messages-load',
            type: 'error',
            title: 'Messages',
            description: msg,
        });
    }
    finally {
        loading.value = false;
    }
}
async function refreshCount() {
    try {
        const cnt = await DefaultService.countTwitchMessages(countQuery());
        totalCount.value = cnt.total;
    }
    catch {
        totalCount.value = null;
    }
}
async function fetchMore() {
    if (!messages.value.length || loadingMore.value) {
        return;
    }
    loadingMore.value = true;
    try {
        const next = await DefaultService.listTwitchMessages(buildQuery(true));
        const seen = new Set(messages.value.map((m) => m.id));
        for (const m of next) {
            if (!seen.has(m.id)) {
                messages.value.push(m);
                seen.add(m.id);
            }
        }
        void refreshCount();
    }
    catch {
        notify({
            id: 'messages-more',
            type: 'error',
            title: 'Messages',
            description: 'Could not load more.',
        });
    }
    finally {
        loadingMore.value = false;
    }
}
function openFilters() {
    draft.value = { ...applied.value };
    filtersOpen.value = true;
}
async function applyFilters() {
    if (filtersApplying.value) {
        return;
    }
    filtersApplying.value = true;
    try {
        applied.value = { ...draft.value };
        await fetchFirst();
        filtersOpen.value = false;
    }
    finally {
        filtersApplying.value = false;
    }
}
function clearFilters() {
    const empty = { username: '', text: '', channel: '', createdFrom: '', createdTo: '' };
    draft.value = { ...empty };
    applied.value = { ...empty };
    filtersOpen.value = false;
    void fetchFirst();
}
onMounted(() => {
    void fetchFirst();
});
function rowBadges(m) {
    return [...(m.badge_tags ?? [])];
}
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "page messages-page" },
});
/** @type {__VLS_StyleScopedClasses['page']} */ ;
/** @type {__VLS_StyleScopedClasses['messages-page']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
    ...{ class: "page-head" },
});
/** @type {__VLS_StyleScopedClasses['page-head']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h1, __VLS_intrinsics.h1)({
    ...{ class: "page-title" },
});
/** @type {__VLS_StyleScopedClasses['page-title']} */ ;
if (__VLS_ctx.totalCount != null) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "count-pill" },
    });
    /** @type {__VLS_StyleScopedClasses['count-pill']} */ ;
    (__VLS_ctx.totalCount.toLocaleString());
}
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "toolbar" },
});
/** @type {__VLS_StyleScopedClasses['toolbar']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (__VLS_ctx.openFilters) },
    type: "button",
    ...{ class: "btn-filter" },
});
/** @type {__VLS_StyleScopedClasses['btn-filter']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
    ...{ class: "filter-hint" },
    title: (__VLS_ctx.filterSummary),
});
/** @type {__VLS_StyleScopedClasses['filter-hint']} */ ;
(__VLS_ctx.filterSummary);
if (__VLS_ctx.loading) {
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
            showChannel: true,
            channelLogin: (m.channel),
        }, ...__VLS_functionalComponentArgsRest(__VLS_1));
        // @ts-ignore
        [totalCount, totalCount, openFilters, filterSummary, filterSummary, loading, messages, ChatHistoryEntry, rowBadges,];
    }
}
if (!__VLS_ctx.loading && __VLS_ctx.messages.length) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "more-row" },
    });
    /** @type {__VLS_StyleScopedClasses['more-row']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (__VLS_ctx.fetchMore) },
        type: "button",
        ...{ class: "btn-more" },
        disabled: (__VLS_ctx.loadingMore),
    });
    /** @type {__VLS_StyleScopedClasses['btn-more']} */ ;
    (__VLS_ctx.loadingMore ? 'Loading…' : 'Load more');
}
const __VLS_5 = AppModal || AppModal;
// @ts-ignore
const __VLS_6 = __VLS_asFunctionalComponent1(__VLS_5, new __VLS_5({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.filtersOpen),
    title: "Message filters",
}));
const __VLS_7 = __VLS_6({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.filtersOpen),
    title: "Message filters",
}, ...__VLS_functionalComponentArgsRest(__VLS_6));
let __VLS_10;
const __VLS_11 = ({ close: {} },
    { onClose: (...[$event]) => {
            __VLS_ctx.filtersOpen = false;
            // @ts-ignore
            [loading, messages, fetchMore, loadingMore, loadingMore, filtersOpen, filtersOpen,];
        } });
const { default: __VLS_12 } = __VLS_8.slots;
{
    const { default: __VLS_13 } = __VLS_8.slots;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "fields" },
    });
    /** @type {__VLS_StyleScopedClasses['fields']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        value: (__VLS_ctx.draft.username),
        type: "text",
        autocomplete: "off",
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        value: (__VLS_ctx.draft.text),
        type: "text",
        autocomplete: "off",
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        value: (__VLS_ctx.draft.channel),
        type: "text",
        placeholder: "channel login",
        autocomplete: "off",
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        type: "datetime-local",
    });
    (__VLS_ctx.draft.createdFrom);
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        type: "datetime-local",
    });
    (__VLS_ctx.draft.createdTo);
    // @ts-ignore
    [draft, draft, draft, draft, draft,];
}
{
    const { footer: __VLS_14 } = __VLS_8.slots;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (__VLS_ctx.clearFilters) },
        type: "button",
        ...{ class: "btn-ghost" },
    });
    /** @type {__VLS_StyleScopedClasses['btn-ghost']} */ ;
    const __VLS_15 = SubmitButton || SubmitButton;
    // @ts-ignore
    const __VLS_16 = __VLS_asFunctionalComponent1(__VLS_15, new __VLS_15({
        ...{ 'onClick': {} },
        nativeType: "button",
        ...{ class: "btn-primary" },
        loading: (__VLS_ctx.filtersApplying),
    }));
    const __VLS_17 = __VLS_16({
        ...{ 'onClick': {} },
        nativeType: "button",
        ...{ class: "btn-primary" },
        loading: (__VLS_ctx.filtersApplying),
    }, ...__VLS_functionalComponentArgsRest(__VLS_16));
    let __VLS_20;
    const __VLS_21 = ({ click: {} },
        { onClick: (__VLS_ctx.applyFilters) });
    /** @type {__VLS_StyleScopedClasses['btn-primary']} */ ;
    const { default: __VLS_22 } = __VLS_18.slots;
    (__VLS_ctx.filtersApplying ? 'Applying…' : 'Apply');
    // @ts-ignore
    [clearFilters, filtersApplying, filtersApplying, applyFilters,];
    var __VLS_18;
    var __VLS_19;
    // @ts-ignore
    [];
}
// @ts-ignore
[];
var __VLS_8;
var __VLS_9;
// @ts-ignore
[];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
