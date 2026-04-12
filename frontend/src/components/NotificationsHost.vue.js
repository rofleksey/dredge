/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { storeToRefs } from 'pinia';
import { watch } from 'vue';
import { notify } from '../lib/notify';
import { useLiveSocketStore } from '../stores/liveSocket';
import { useNotificationsStore } from '../stores/notifications';
const notifications = useNotificationsStore();
const { items } = storeToRefs(notifications);
const live = useLiveSocketStore();
const { lastError } = storeToRefs(live);
watch(lastError, (msg) => {
    if (msg) {
        notify({
            id: 'live-ws',
            type: 'warning',
            title: 'Live connection',
            description: msg,
        });
    }
}, { immediate: true });
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['toast-progress']} */ ;
/** @type {__VLS_StyleScopedClasses['toast-progress']} */ ;
/** @type {__VLS_StyleScopedClasses['toast-progress']} */ ;
/** @type {__VLS_StyleScopedClasses['toast-progress']} */ ;
/** @type {__VLS_StyleScopedClasses['toast--success']} */ ;
/** @type {__VLS_StyleScopedClasses['toast--info']} */ ;
/** @type {__VLS_StyleScopedClasses['toast--warning']} */ ;
/** @type {__VLS_StyleScopedClasses['toast--error']} */ ;
let __VLS_0;
/** @ts-ignore @type {typeof __VLS_components.Teleport | typeof __VLS_components.Teleport} */
Teleport;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
    to: "body",
}));
const __VLS_2 = __VLS_1({
    to: "body",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
const { default: __VLS_5 } = __VLS_3.slots;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "notifications-host" },
    'aria-live': "polite",
});
/** @type {__VLS_StyleScopedClasses['notifications-host']} */ ;
for (const [n] of __VLS_vFor((__VLS_ctx.items))) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        key: (n.key),
        ...{ class: "toast" },
        ...{ class: (`toast--${n.type}`) },
        role: "status",
    });
    /** @type {__VLS_StyleScopedClasses['toast']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "toast-body" },
    });
    /** @type {__VLS_StyleScopedClasses['toast-body']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "toast-title" },
    });
    /** @type {__VLS_StyleScopedClasses['toast-title']} */ ;
    (n.title);
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "toast-desc" },
    });
    /** @type {__VLS_StyleScopedClasses['toast-desc']} */ ;
    (n.description);
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "toast-progress-wrap" },
        'aria-hidden': "true",
    });
    /** @type {__VLS_StyleScopedClasses['toast-progress-wrap']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div)({
        ...{ class: "toast-progress" },
        ...{ style: ({ animationDuration: `${n.durationMs}ms` }) },
    });
    /** @type {__VLS_StyleScopedClasses['toast-progress']} */ ;
    // @ts-ignore
    [items,];
}
// @ts-ignore
[];
var __VLS_3;
// @ts-ignore
[];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
