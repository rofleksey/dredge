/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { storeToRefs } from 'pinia';
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import AppShell from '../components/AppShell.vue';
import WatchView from '../views/WatchView.vue';
import { useLiveSocketStore } from '../stores/liveSocket';
const route = useRoute();
const live = useLiveSocketStore();
const { connected: liveConnected, lastError: liveError } = storeToRefs(live);
const fillOutlet = computed(() => ['settings', 'messages', 'users', 'user', 'streams', 'stream'].includes(String(route.name)));
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
const __VLS_0 = AppShell || AppShell;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
    liveConnected: (__VLS_ctx.liveConnected),
    liveError: (__VLS_ctx.liveError),
}));
const __VLS_2 = __VLS_1({
    liveConnected: (__VLS_ctx.liveConnected),
    liveError: (__VLS_ctx.liveError),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_5 = {};
const { default: __VLS_6 } = __VLS_3.slots;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "auth-content" },
});
/** @type {__VLS_StyleScopedClasses['auth-content']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "watch-persist" },
    ...{ class: ({ 'watch-persist--offscreen': __VLS_ctx.route.name !== 'watch' }) },
});
/** @type {__VLS_StyleScopedClasses['watch-persist']} */ ;
/** @type {__VLS_StyleScopedClasses['watch-persist--offscreen']} */ ;
const __VLS_7 = WatchView;
// @ts-ignore
const __VLS_8 = __VLS_asFunctionalComponent1(__VLS_7, new __VLS_7({}));
const __VLS_9 = __VLS_8({}, ...__VLS_functionalComponentArgsRest(__VLS_8));
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "auth-outlet" },
    ...{ class: ({ 'auth-outlet--fill': __VLS_ctx.fillOutlet }) },
});
/** @type {__VLS_StyleScopedClasses['auth-outlet']} */ ;
/** @type {__VLS_StyleScopedClasses['auth-outlet--fill']} */ ;
let __VLS_12;
/** @ts-ignore @type {typeof __VLS_components.routerView | typeof __VLS_components.RouterView} */
routerView;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent1(__VLS_12, new __VLS_12({}));
const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
// @ts-ignore
[liveConnected, liveError, route, fillOutlet,];
var __VLS_3;
// @ts-ignore
[];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
