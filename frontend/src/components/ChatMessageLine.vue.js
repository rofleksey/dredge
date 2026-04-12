/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { computed } from 'vue';
import { badgeEmojis } from '../lib/chatBadges';
import TwitchUserLink from './TwitchUserLink.vue';
const props = withDefaults(defineProps(), {
    createdAt: undefined,
    chatterUserId: undefined,
    highlightChannel: '',
    showChannel: false,
    channelLogin: '',
    userMarked: false,
    userIsSus: false,
});
const badgeStr = computed(() => badgeEmojis(props.badgeTags));
const timeLabel = computed(() => {
    if (!props.showTimestamp || !props.createdAt) {
        return '';
    }
    const t = Date.parse(props.createdAt);
    if (!Number.isFinite(t)) {
        return '';
    }
    return new Intl.DateTimeFormat(undefined, {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
    }).format(t);
});
const __VLS_defaults = {
    createdAt: undefined,
    chatterUserId: undefined,
    highlightChannel: '',
    showChannel: false,
    channelLogin: '',
    userMarked: false,
    userIsSus: false,
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
/** @type {__VLS_StyleScopedClasses['kw']} */ ;
/** @type {__VLS_StyleScopedClasses['kw']} */ ;
/** @type {__VLS_StyleScopedClasses['kw']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
    ...{ class: ({
            kw: __VLS_ctx.keyword,
            sent: __VLS_ctx.fromSent && !__VLS_ctx.keyword,
            marked: __VLS_ctx.userMarked && !__VLS_ctx.keyword && !__VLS_ctx.userIsSus,
            suspicious: __VLS_ctx.userIsSus && !__VLS_ctx.keyword,
        }) },
});
/** @type {__VLS_StyleScopedClasses['kw']} */ ;
/** @type {__VLS_StyleScopedClasses['sent']} */ ;
/** @type {__VLS_StyleScopedClasses['marked']} */ ;
/** @type {__VLS_StyleScopedClasses['suspicious']} */ ;
if (__VLS_ctx.timeLabel) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "ts" },
    });
    /** @type {__VLS_StyleScopedClasses['ts']} */ ;
    (__VLS_ctx.timeLabel);
}
if (__VLS_ctx.showChannel && __VLS_ctx.channelLogin) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "chan" },
        title: "Channel",
    });
    /** @type {__VLS_StyleScopedClasses['chan']} */ ;
    (__VLS_ctx.channelLogin);
}
if (__VLS_ctx.badgeStr) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "badges" },
        'aria-hidden': "true",
    });
    /** @type {__VLS_StyleScopedClasses['badges']} */ ;
    (__VLS_ctx.badgeStr);
}
const __VLS_0 = TwitchUserLink;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
    login: (__VLS_ctx.user),
    userTwitchId: (__VLS_ctx.chatterUserId),
    highlightChannel: (__VLS_ctx.highlightChannel),
    variant: "chat",
}));
const __VLS_2 = __VLS_1({
    login: (__VLS_ctx.user),
    userTwitchId: (__VLS_ctx.chatterUserId),
    highlightChannel: (__VLS_ctx.highlightChannel),
    variant: "chat",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
    ...{ class: "txt" },
});
/** @type {__VLS_StyleScopedClasses['txt']} */ ;
(__VLS_ctx.message);
// @ts-ignore
[keyword, keyword, keyword, keyword, fromSent, userMarked, userIsSus, userIsSus, timeLabel, timeLabel, showChannel, channelLogin, channelLogin, badgeStr, badgeStr, user, chatterUserId, highlightChannel, message,];
const __VLS_export = (await import('vue')).defineComponent({
    __defaults: __VLS_defaults,
    __typeProps: {},
});
export default {};
