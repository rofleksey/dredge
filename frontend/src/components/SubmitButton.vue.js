/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
const __VLS_props = withDefaults(defineProps(), {
    loading: false,
    nativeType: 'submit',
    disabled: false,
});
const __VLS_defaults = {
    loading: false,
    nativeType: 'submit',
    disabled: false,
};
const __VLS_ctx = {
    ...{},
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    type: (__VLS_ctx.nativeType),
    disabled: (__VLS_ctx.loading || __VLS_ctx.disabled),
    'aria-busy': (__VLS_ctx.loading),
    ...{ class: "submit-btn" },
});
/** @type {__VLS_StyleScopedClasses['submit-btn']} */ ;
if (__VLS_ctx.loading) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.span)({
        ...{ class: "btn-submit-spinner" },
        'aria-hidden': "true",
    });
    /** @type {__VLS_StyleScopedClasses['btn-submit-spinner']} */ ;
}
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
    ...{ class: "submit-btn-label" },
});
/** @type {__VLS_StyleScopedClasses['submit-btn-label']} */ ;
var __VLS_0 = {};
// @ts-ignore
var __VLS_1 = __VLS_0;
// @ts-ignore
[nativeType, loading, loading, loading, disabled,];
const __VLS_base = (await import('vue')).defineComponent({
    __defaults: __VLS_defaults,
    __typeProps: {},
});
const __VLS_export = {};
export default {};
