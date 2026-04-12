/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { computed } from 'vue';
const props = defineProps();
const src = computed(() => {
    const ch = props.channel.replace(/^#/, '').trim();
    if (!ch) {
        return '';
    }
    const q = new URLSearchParams({ channel: ch });
    for (const p of new Set([window.location.hostname, 'localhost', '127.0.0.1'])) {
        q.append('parent', p);
    }
    return `https://player.twitch.tv/?${q.toString()}`;
});
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
    ...{ class: "player-wrap" },
});
/** @type {__VLS_StyleScopedClasses['player-wrap']} */ ;
if (__VLS_ctx.src) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.iframe)({
        ...{ class: "player" },
        src: (__VLS_ctx.src),
        allowfullscreen: true,
        allow: "autoplay; fullscreen",
        title: "Twitch stream",
    });
    /** @type {__VLS_StyleScopedClasses['player']} */ ;
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "placeholder" },
    });
    /** @type {__VLS_StyleScopedClasses['placeholder']} */ ;
}
// @ts-ignore
[src, src,];
const __VLS_export = (await import('vue')).defineComponent({
    __typeProps: {},
});
export default {};
