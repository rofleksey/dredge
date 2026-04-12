/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { onUnmounted, watch } from 'vue';
const props = withDefaults(defineProps(), { title: '', wide: false, extraWide: false });
const emit = defineEmits();
function onDocKeydown(e) {
    if (e.key === 'Escape') {
        emit('close');
    }
}
watch(() => props.open, (v) => {
    if (v) {
        document.addEventListener('keydown', onDocKeydown);
    }
    else {
        document.removeEventListener('keydown', onDocKeydown);
    }
}, { immediate: true });
onUnmounted(() => {
    document.removeEventListener('keydown', onDocKeydown);
});
const __VLS_defaults = { title: '', wide: false, extraWide: false };
const __VLS_ctx = {
    ...{},
    ...{},
    ...{},
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
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
if (__VLS_ctx.open) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "modal-root" },
    });
    /** @type {__VLS_StyleScopedClasses['modal-root']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.open))
                    return;
                __VLS_ctx.emit('close');
                // @ts-ignore
                [open, emit,];
            } },
        ...{ class: "modal-backdrop" },
        'aria-hidden': "true",
    });
    /** @type {__VLS_StyleScopedClasses['modal-backdrop']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "modal-panel" },
        ...{ class: ({ 'modal-panel--wide': __VLS_ctx.wide, 'modal-panel--extra-wide': __VLS_ctx.extraWide }) },
        role: "dialog",
        'aria-label': (__VLS_ctx.title || 'Dialog'),
    });
    /** @type {__VLS_StyleScopedClasses['modal-panel']} */ ;
    /** @type {__VLS_StyleScopedClasses['modal-panel--wide']} */ ;
    /** @type {__VLS_StyleScopedClasses['modal-panel--extra-wide']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.header, __VLS_intrinsics.header)({
        ...{ class: "modal-head" },
    });
    /** @type {__VLS_StyleScopedClasses['modal-head']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({
        ...{ class: "modal-title" },
    });
    /** @type {__VLS_StyleScopedClasses['modal-title']} */ ;
    var __VLS_6 = {};
    (__VLS_ctx.title);
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.open))
                    return;
                __VLS_ctx.emit('close');
                // @ts-ignore
                [emit, wide, extraWide, title, title,];
            } },
        type: "button",
        ...{ class: "btn-close" },
        'aria-label': "Close",
    });
    /** @type {__VLS_StyleScopedClasses['btn-close']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "modal-body" },
    });
    /** @type {__VLS_StyleScopedClasses['modal-body']} */ ;
    var __VLS_8 = {};
    if (__VLS_ctx.$slots.footer) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.footer, __VLS_intrinsics.footer)({
            ...{ class: "modal-foot" },
        });
        /** @type {__VLS_StyleScopedClasses['modal-foot']} */ ;
        var __VLS_10 = {};
    }
}
// @ts-ignore
[$slots,];
var __VLS_3;
// @ts-ignore
var __VLS_7 = __VLS_6, __VLS_9 = __VLS_8, __VLS_11 = __VLS_10;
// @ts-ignore
[];
const __VLS_base = (await import('vue')).defineComponent({
    __typeEmits: {},
    __defaults: __VLS_defaults,
    __typeProps: {},
});
const __VLS_export = {};
export default {};
