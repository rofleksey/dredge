/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { onMounted, ref } from 'vue';
import { RouterLink } from 'vue-router';
import { ApiError, DefaultService } from '../api/generated';
import { notify } from '../lib/notify';
defineOptions({ name: 'StreamsView' });
const streams = ref([]);
const loading = ref(false);
const loadingMore = ref(false);
const channelFilter = ref('');
const cursorStartedAt = ref();
const cursorId = ref();
const hasMore = ref(true);
async function loadFirst() {
    if (loading.value) {
        return;
    }
    loading.value = true;
    cursorStartedAt.value = undefined;
    cursorId.value = undefined;
    hasMore.value = true;
    try {
        await loadPage(false);
    }
    finally {
        loading.value = false;
    }
}
async function loadPage(append) {
    const list = await DefaultService.listRecordedStreams({
        channelLogin: channelFilter.value.trim() || undefined,
        limit: 50,
        cursorStartedAt: append ? cursorStartedAt.value : undefined,
        cursorId: append ? cursorId.value : undefined,
    });
    if (append) {
        streams.value = streams.value.concat(list);
    }
    else {
        streams.value = list;
    }
    if (list.length < 50) {
        hasMore.value = false;
    }
    else {
        const last = list[list.length - 1];
        cursorStartedAt.value = last.started_at;
        cursorId.value = last.id;
        hasMore.value = true;
    }
}
async function loadMore() {
    if (loadingMore.value || !hasMore.value || !streams.value.length) {
        return;
    }
    loadingMore.value = true;
    try {
        await loadPage(true);
    }
    catch (e) {
        notifyFromErr(e, 'streams-more');
    }
    finally {
        loadingMore.value = false;
    }
}
function notifyFromErr(e, id) {
    const msg = e instanceof ApiError && e.body && typeof e.body.message === 'string'
        ? e.body.message
        : 'Request failed.';
    notify({ id, type: 'error', title: 'Streams', description: msg });
}
async function applyFilter() {
    try {
        await loadFirst();
    }
    catch (e) {
        streams.value = [];
        notifyFromErr(e, 'streams-load');
    }
}
onMounted(() => {
    void loadFirst().catch((e) => {
        streams.value = [];
        notifyFromErr(e, 'streams-load');
    });
});
function formatWhen(iso) {
    const t = Date.parse(iso);
    if (!Number.isFinite(t)) {
        return iso;
    }
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(t);
}
function statusLabel(s) {
    return s.ended_at ? 'Ended' : 'Live';
}
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "streams-page" },
});
/** @type {__VLS_StyleScopedClasses['streams-page']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
    ...{ class: "streams-head" },
});
/** @type {__VLS_StyleScopedClasses['streams-head']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h1, __VLS_intrinsics.h1)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "muted" },
});
/** @type {__VLS_StyleScopedClasses['muted']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.applyFilter) },
    ...{ class: "streams-filter" },
});
/** @type {__VLS_StyleScopedClasses['streams-filter']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "field" },
});
/** @type {__VLS_StyleScopedClasses['field']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
    ...{ class: "label" },
});
/** @type {__VLS_StyleScopedClasses['label']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "search",
    name: "channel",
    placeholder: "Filter by login…",
    autocomplete: "off",
});
(__VLS_ctx.channelFilter);
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    type: "submit",
    ...{ class: "btn-apply" },
});
/** @type {__VLS_StyleScopedClasses['btn-apply']} */ ;
if (__VLS_ctx.loading) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
else if (__VLS_ctx.streams.length) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.table, __VLS_intrinsics.table)({
        ...{ class: "streams-table" },
    });
    /** @type {__VLS_StyleScopedClasses['streams-table']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.thead, __VLS_intrinsics.thead)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.tr, __VLS_intrinsics.tr)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.th, __VLS_intrinsics.th)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.tbody, __VLS_intrinsics.tbody)({});
    for (const [s] of __VLS_vFor((__VLS_ctx.streams))) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.tr, __VLS_intrinsics.tr)({
            key: (s.id),
        });
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
        let __VLS_0;
        /** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
        RouterLink;
        // @ts-ignore
        const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
            ...{ class: "link" },
            to: ({ name: 'stream', params: { id: String(s.id) } }),
        }));
        const __VLS_2 = __VLS_1({
            ...{ class: "link" },
            to: ({ name: 'stream', params: { id: String(s.id) } }),
        }, ...__VLS_functionalComponentArgsRest(__VLS_1));
        /** @type {__VLS_StyleScopedClasses['link']} */ ;
        const { default: __VLS_5 } = __VLS_3.slots;
        (s.channel_login);
        // @ts-ignore
        [applyFilter, channelFilter, loading, streams, streams,];
        var __VLS_3;
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
        let __VLS_6;
        /** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
        RouterLink;
        // @ts-ignore
        const __VLS_7 = __VLS_asFunctionalComponent1(__VLS_6, new __VLS_6({
            ...{ class: "link title" },
            to: ({ name: 'stream', params: { id: String(s.id) } }),
        }));
        const __VLS_8 = __VLS_7({
            ...{ class: "link title" },
            to: ({ name: 'stream', params: { id: String(s.id) } }),
        }, ...__VLS_functionalComponentArgsRest(__VLS_7));
        /** @type {__VLS_StyleScopedClasses['link']} */ ;
        /** @type {__VLS_StyleScopedClasses['title']} */ ;
        const { default: __VLS_11 } = __VLS_9.slots;
        (s.title?.trim() || '—');
        // @ts-ignore
        [];
        var __VLS_9;
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        (s.game_name?.trim() || '—');
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({
            ...{ class: "muted" },
        });
        /** @type {__VLS_StyleScopedClasses['muted']} */ ;
        (__VLS_ctx.formatWhen(s.started_at));
        __VLS_asFunctionalElement1(__VLS_intrinsics.td, __VLS_intrinsics.td)({});
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: (['status', { 'status--live': !s.ended_at }]) },
        });
        /** @type {__VLS_StyleScopedClasses['status']} */ ;
        /** @type {__VLS_StyleScopedClasses['status--live']} */ ;
        (__VLS_ctx.statusLabel(s));
        // @ts-ignore
        [formatWhen, statusLabel,];
    }
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
if (__VLS_ctx.streams.length && __VLS_ctx.hasMore) {
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
// @ts-ignore
[streams, hasMore, loadMore, loadingMore, loadingMore,];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
