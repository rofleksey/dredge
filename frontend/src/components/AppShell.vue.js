/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { computed } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
const route = useRoute();
const router = useRouter();
const props = withDefaults(defineProps(), {
    liveConnected: false,
    liveError: null,
});
const auth = useAuthStore();
const liveTitle = computed(() => {
    if (props.liveConnected) {
        return 'Live: connected';
    }
    if (props.liveError) {
        return `Live: connection problem (${props.liveError})`;
    }
    return 'Live: connecting…';
});
function logout() {
    auth.logout();
    void router.push({ name: 'login' });
}
const __VLS_defaults = {
    liveConnected: false,
    liveError: null,
};
const __VLS_ctx = {
    ...{},
    ...{},
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "shell" },
});
/** @type {__VLS_StyleScopedClasses['shell']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
    ...{ class: "top" },
});
/** @type {__VLS_StyleScopedClasses['top']} */ ;
let __VLS_0;
/** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
RouterLink;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
    ...{ class: "brand" },
    to: "/",
}));
const __VLS_2 = __VLS_1({
    ...{ class: "brand" },
    to: "/",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
/** @type {__VLS_StyleScopedClasses['brand']} */ ;
const { default: __VLS_5 } = __VLS_3.slots;
var __VLS_3;
__VLS_asFunctionalElement1(__VLS_intrinsics.nav, __VLS_intrinsics.nav)({
    ...{ class: "nav" },
});
/** @type {__VLS_StyleScopedClasses['nav']} */ ;
let __VLS_6;
/** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
RouterLink;
// @ts-ignore
const __VLS_7 = __VLS_asFunctionalComponent1(__VLS_6, new __VLS_6({
    to: "/",
    activeClass: "active",
}));
const __VLS_8 = __VLS_7({
    to: "/",
    activeClass: "active",
}, ...__VLS_functionalComponentArgsRest(__VLS_7));
const { default: __VLS_11 } = __VLS_9.slots;
var __VLS_9;
let __VLS_12;
/** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
RouterLink;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent1(__VLS_12, new __VLS_12({
    to: "/messages",
    activeClass: "active",
}));
const __VLS_14 = __VLS_13({
    to: "/messages",
    activeClass: "active",
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const { default: __VLS_17 } = __VLS_15.slots;
var __VLS_15;
let __VLS_18;
/** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
RouterLink;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent1(__VLS_18, new __VLS_18({
    to: "/users",
    ...{ class: ({ active: __VLS_ctx.route.name === 'users' || __VLS_ctx.route.name === 'user' }) },
}));
const __VLS_20 = __VLS_19({
    to: "/users",
    ...{ class: ({ active: __VLS_ctx.route.name === 'users' || __VLS_ctx.route.name === 'user' }) },
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
/** @type {__VLS_StyleScopedClasses['active']} */ ;
const { default: __VLS_23 } = __VLS_21.slots;
// @ts-ignore
[route, route,];
var __VLS_21;
let __VLS_24;
/** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
RouterLink;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent1(__VLS_24, new __VLS_24({
    to: "/streams",
    ...{ class: ({ active: __VLS_ctx.route.name === 'streams' || __VLS_ctx.route.name === 'stream' }) },
}));
const __VLS_26 = __VLS_25({
    to: "/streams",
    ...{ class: ({ active: __VLS_ctx.route.name === 'streams' || __VLS_ctx.route.name === 'stream' }) },
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
/** @type {__VLS_StyleScopedClasses['active']} */ ;
const { default: __VLS_29 } = __VLS_27.slots;
// @ts-ignore
[route, route,];
var __VLS_27;
let __VLS_30;
/** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
RouterLink;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent1(__VLS_30, new __VLS_30({
    to: "/settings",
    activeClass: "active",
}));
const __VLS_32 = __VLS_31({
    to: "/settings",
    activeClass: "active",
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
const { default: __VLS_35 } = __VLS_33.slots;
// @ts-ignore
[];
var __VLS_33;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "user" },
});
/** @type {__VLS_StyleScopedClasses['user']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span)({
    ...{ class: "live-dot" },
    ...{ class: ({
            'live-dot--ok': __VLS_ctx.liveConnected,
            'live-dot--err': !__VLS_ctx.liveConnected && __VLS_ctx.liveError,
            'live-dot--pending': !__VLS_ctx.liveConnected && !__VLS_ctx.liveError,
        }) },
    role: "status",
    'aria-label': (__VLS_ctx.liveTitle),
    title: (__VLS_ctx.liveTitle),
});
/** @type {__VLS_StyleScopedClasses['live-dot']} */ ;
/** @type {__VLS_StyleScopedClasses['live-dot--ok']} */ ;
/** @type {__VLS_StyleScopedClasses['live-dot--err']} */ ;
/** @type {__VLS_StyleScopedClasses['live-dot--pending']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (__VLS_ctx.logout) },
    type: "button",
    ...{ class: "btn-ghost" },
});
/** @type {__VLS_StyleScopedClasses['btn-ghost']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.main, __VLS_intrinsics.main)({
    ...{ class: "main" },
});
/** @type {__VLS_StyleScopedClasses['main']} */ ;
var __VLS_36 = {};
// @ts-ignore
var __VLS_37 = __VLS_36;
// @ts-ignore
[liveConnected, liveConnected, liveConnected, liveError, liveError, liveTitle, liveTitle, logout,];
const __VLS_base = (await import('vue')).defineComponent({
    __defaults: __VLS_defaults,
    __typeProps: {},
});
const __VLS_export = {};
export default {};
