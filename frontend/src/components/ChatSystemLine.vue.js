/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import TwitchUserLink from './TwitchUserLink.vue';
const props = withDefaults(defineProps(), {
    detail: '',
    chatterUserId: undefined,
    highlightChannel: '',
});
const __VLS_defaults = {
    detail: '',
    chatterUserId: undefined,
    highlightChannel: '',
};
const __VLS_ctx = {
    ...{},
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
__VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
    ...{ class: "sys-line" },
    ...{ class: ({
            'sys-line--join': __VLS_ctx.variant === 'join',
            'sys-line--part': __VLS_ctx.variant === 'part',
            'sys-line--gap': __VLS_ctx.variant === 'gap',
        }) },
});
/** @type {__VLS_StyleScopedClasses['sys-line']} */ ;
/** @type {__VLS_StyleScopedClasses['sys-line--join']} */ ;
/** @type {__VLS_StyleScopedClasses['sys-line--part']} */ ;
/** @type {__VLS_StyleScopedClasses['sys-line--gap']} */ ;
if (__VLS_ctx.variant === 'gap') {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "sys-gap-text" },
    });
    /** @type {__VLS_StyleScopedClasses['sys-gap-text']} */ ;
    (__VLS_ctx.text);
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "sys-badge" },
        'aria-hidden': "true",
    });
    /** @type {__VLS_StyleScopedClasses['sys-badge']} */ ;
    (__VLS_ctx.variant === 'join' ? '→' : '←');
    const __VLS_0 = TwitchUserLink;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
        login: (__VLS_ctx.user),
        userTwitchId: (__VLS_ctx.chatterUserId),
        highlightChannel: (__VLS_ctx.highlightChannel),
        variant: "system",
    }));
    const __VLS_2 = __VLS_1({
        login: (__VLS_ctx.user),
        userTwitchId: (__VLS_ctx.chatterUserId),
        highlightChannel: (__VLS_ctx.highlightChannel),
        variant: "system",
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "sys-msg" },
    });
    /** @type {__VLS_StyleScopedClasses['sys-msg']} */ ;
    (__VLS_ctx.text);
    if (__VLS_ctx.detail) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
            ...{ class: "sys-detail" },
        });
        /** @type {__VLS_StyleScopedClasses['sys-detail']} */ ;
        (__VLS_ctx.detail);
    }
}
// @ts-ignore
[variant, variant, variant, variant, variant, text, text, user, chatterUserId, highlightChannel, detail, detail,];
const __VLS_export = (await import('vue')).defineComponent({
    __defaults: __VLS_defaults,
    __typeProps: {},
});
export default {};
