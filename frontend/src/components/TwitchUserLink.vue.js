/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import { useTwitchUserHighlight } from '../composables/useTwitchUserHighlight';
const props = withDefaults(defineProps(), {
    userTwitchId: undefined,
    highlightChannel: '',
    variant: 'chat',
});
const { highlightClass } = useTwitchUserHighlight(() => props.highlightChannel ?? '');
const extraClass = computed(() => highlightClass.value(props.login));
const __VLS_defaults = {
    userTwitchId: undefined,
    highlightChannel: '',
    variant: 'chat',
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
/** @type {__VLS_StyleScopedClasses['twitch-user-link']} */ ;
if (__VLS_ctx.userTwitchId != null && __VLS_ctx.userTwitchId !== undefined) {
    let __VLS_0;
    /** @ts-ignore @type {typeof __VLS_components.RouterLink | typeof __VLS_components.RouterLink} */
    RouterLink;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
        ...{ class: "twitch-user-link" },
        ...{ class: ([`twitch-user-link--${__VLS_ctx.variant}`, __VLS_ctx.extraClass]) },
        to: ({ name: 'user', params: { id: String(__VLS_ctx.userTwitchId) } }),
    }));
    const __VLS_2 = __VLS_1({
        ...{ class: "twitch-user-link" },
        ...{ class: ([`twitch-user-link--${__VLS_ctx.variant}`, __VLS_ctx.extraClass]) },
        to: ({ name: 'user', params: { id: String(__VLS_ctx.userTwitchId) } }),
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    var __VLS_5 = {};
    /** @type {__VLS_StyleScopedClasses['twitch-user-link']} */ ;
    const { default: __VLS_6 } = __VLS_3.slots;
    (__VLS_ctx.login);
    // @ts-ignore
    [userTwitchId, userTwitchId, userTwitchId, variant, extraClass, login,];
    var __VLS_3;
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "twitch-user-link" },
        ...{ class: ([`twitch-user-link--${__VLS_ctx.variant}`, __VLS_ctx.extraClass]) },
    });
    /** @type {__VLS_StyleScopedClasses['twitch-user-link']} */ ;
    (__VLS_ctx.login);
}
// @ts-ignore
[variant, extraClass, login,];
const __VLS_export = (await import('vue')).defineComponent({
    __defaults: __VLS_defaults,
    __typeProps: {},
});
export default {};
