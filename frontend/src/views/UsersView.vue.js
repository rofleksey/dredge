/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { ApiError, DefaultService } from '../api/generated';
import { notify } from '../lib/notify';
defineOptions({ name: 'UsersView' });
const q = ref('');
const users = ref([]);
const loading = ref(false);
const loadingMore = ref(false);
const totalCount = ref(null);
let debounceTimer = null;
async function load(append = false) {
    if (append) {
        loadingMore.value = true;
    }
    else {
        loading.value = true;
    }
    try {
        const last = append && users.value.length ? users.value[users.value.length - 1] : undefined;
        const [list, cnt] = await Promise.all([
            DefaultService.listTwitchDirectoryUsers({
                username: q.value.trim() || undefined,
                limit: 100,
                cursorId: append && last ? last.id : undefined,
                cursorMarked: append && last ? last.marked : undefined,
            }),
            DefaultService.countTwitchDirectoryUsers({
                username: q.value.trim() || undefined,
            }),
        ]);
        if (append) {
            const seen = new Set(users.value.map((u) => u.id));
            for (const u of list) {
                if (!seen.has(u.id)) {
                    users.value.push(u);
                    seen.add(u.id);
                }
            }
        }
        else {
            users.value = list;
        }
        totalCount.value = cnt.total;
    }
    catch (e) {
        if (!append) {
            users.value = [];
        }
        totalCount.value = null;
        const msg = e instanceof ApiError && e.body && typeof e.body.message === 'string'
            ? e.body.message
            : 'Could not load users.';
        notify({
            id: 'users-load',
            type: 'error',
            title: 'Users',
            description: msg,
        });
    }
    finally {
        loading.value = false;
        loadingMore.value = false;
    }
}
async function loadMore() {
    if (!users.value.length || loadingMore.value || loading.value) {
        return;
    }
    await load(true);
}
onMounted(() => {
    void load();
});
watch(q, () => {
    if (debounceTimer) {
        clearTimeout(debounceTimer);
    }
    debounceTimer = setTimeout(() => {
        debounceTimer = null;
        void load(false);
    }, 300);
});
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['suspicious']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "page users-page" },
});
/** @type {__VLS_StyleScopedClasses['page']} */ ;
/** @type {__VLS_StyleScopedClasses['users-page']} */ ;
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
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "search" },
});
/** @type {__VLS_StyleScopedClasses['search']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
    ...{ class: "label" },
});
/** @type {__VLS_StyleScopedClasses['label']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "search",
    autocomplete: "off",
    placeholder: "Substring match…",
});
(__VLS_ctx.q);
if (__VLS_ctx.loading) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
        ...{ class: "user-list" },
    });
    /** @type {__VLS_StyleScopedClasses['user-list']} */ ;
    for (const [u] of __VLS_vFor((__VLS_ctx.users))) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
            key: (u.id),
            ...{ class: ({ suspicious: u.is_sus }) },
        });
        /** @type {__VLS_StyleScopedClasses['suspicious']} */ ;
        let __VLS_0;
        /** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
        RouterLink;
        // @ts-ignore
        const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
            ...{ class: "user-link" },
            ...{ class: ({ marked: u.marked, suspicious: u.is_sus }) },
            to: ({ name: 'user', params: { id: String(u.id) } }),
        }));
        const __VLS_2 = __VLS_1({
            ...{ class: "user-link" },
            ...{ class: ({ marked: u.marked, suspicious: u.is_sus }) },
            to: ({ name: 'user', params: { id: String(u.id) } }),
        }, ...__VLS_functionalComponentArgsRest(__VLS_1));
        /** @type {__VLS_StyleScopedClasses['user-link']} */ ;
        /** @type {__VLS_StyleScopedClasses['marked']} */ ;
        /** @type {__VLS_StyleScopedClasses['suspicious']} */ ;
        const { default: __VLS_5 } = __VLS_3.slots;
        (u.username);
        // @ts-ignore
        [totalCount, totalCount, q, loading, users,];
        var __VLS_3;
        if (u.is_sus) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "tag tag-sus" },
            });
            /** @type {__VLS_StyleScopedClasses['tag']} */ ;
            /** @type {__VLS_StyleScopedClasses['tag-sus']} */ ;
        }
        if (u.marked) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "tag tag-marked" },
            });
            /** @type {__VLS_StyleScopedClasses['tag']} */ ;
            /** @type {__VLS_StyleScopedClasses['tag-marked']} */ ;
        }
        if (u.monitored) {
            __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
                ...{ class: "tag" },
            });
            /** @type {__VLS_StyleScopedClasses['tag']} */ ;
        }
        // @ts-ignore
        [];
    }
}
if (!__VLS_ctx.loading && __VLS_ctx.users.length) {
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
if (!__VLS_ctx.loading && !__VLS_ctx.users.length) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
}
// @ts-ignore
[loading, loading, users, users, loadMore, loadingMore, loadingMore,];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
