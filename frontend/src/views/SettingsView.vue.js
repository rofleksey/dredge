/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/template-helpers.d.ts" />
/// <reference types="C:/Users/rofleksey/AppData/Local/npm-cache/_npx/2db181330ea4b15b/node_modules/@vue/language-core/types/props-fallback.d.ts" />
import { storeToRefs } from 'pinia';
import { computed, onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ApiError, CreateNotificationRequest, DefaultService, TwitchAccount, UpdateTwitchAccountPostRequest, } from '../api/generated';
import AppModal from '../components/AppModal.vue';
import SubmitButton from '../components/SubmitButton.vue';
import { notify } from '../lib/notify';
import { useChannelsStore } from '../stores/channels';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';
import { useWatchPreferencesStore } from '../stores/watchPreferences';
const route = useRoute();
const router = useRouter();
const tab = ref('channels');
const channelsStore = useChannelsStore();
const { monitoredChannels } = storeToRefs(channelsStore);
const twitchStore = useTwitchAccountsStore();
const watchPrefs = useWatchPreferencesStore();
const { chatGapMinutes } = storeToRefs(watchPrefs);
const rules = ref([]);
const rulesTotal = ref(0);
const twitchAccountsTotal = ref(0);
const notifications = ref([]);
const channelBlacklist = ref([]);
const newBlacklistLogin = ref('');
const suspicionDraft = reactive({
    auto_check_account_age: true,
    account_age_sus_days: 14,
    auto_check_blacklist: true,
    auto_check_low_follows: true,
    low_follows_threshold: 10,
    max_gql_follow_pages: 1,
});
const newChannel = reactive({ name: '' });
const ruleModalOpen = ref(false);
const ruleDraft = reactive({
    regex: '',
    included_users: '*',
    denied_users: '',
    included_channels: '*',
    denied_channels: '',
});
const notifModalOpen = ref(false);
const notifDraft = reactive({
    provider: 'telegram',
    enabled: true,
    bot_token: '',
    chat_id: '',
    webhook_url: '',
});
const addingChannel = ref(false);
const savingRule = ref(false);
const savingNotif = ref(false);
const savingBlacklist = ref(false);
const savingSuspicion = ref(false);
async function refresh() {
    const [rulesList, rCount, taCount, notifList, bl, sus] = await Promise.all([
        DefaultService.listRules(),
        DefaultService.countRules(),
        DefaultService.countTwitchAccounts(),
        DefaultService.listNotifications(),
        DefaultService.listChannelBlacklist(),
        DefaultService.getSuspicionSettings(),
    ]);
    await Promise.all([channelsStore.fetch(), twitchStore.fetch()]);
    rules.value = rulesList;
    rulesTotal.value = rCount.total;
    twitchAccountsTotal.value = taCount.total;
    notifications.value = notifList;
    channelBlacklist.value = bl;
    Object.assign(suspicionDraft, sus);
}
function queryOne(v) {
    if (typeof v === 'string') {
        return v;
    }
    if (Array.isArray(v) && typeof v[0] === 'string') {
        return v[0];
    }
    return undefined;
}
function applyTwitchOAuthQuery() {
    const okParam = queryOne(route.query.twitch_oauth_ok);
    const errParam = queryOne(route.query.twitch_oauth_error);
    if (okParam === '1') {
        notify({
            id: 'settings-twitch-oauth-callback',
            type: 'success',
            title: 'Twitch',
            description: 'Twitch account linked.',
        });
    }
    else if (errParam) {
        notify({
            id: 'settings-twitch-oauth-callback',
            type: 'error',
            title: 'Twitch',
            description: errParam,
        });
    }
    if (okParam !== undefined || errParam !== undefined) {
        const next = {};
        for (const [k, v] of Object.entries(route.query)) {
            if (k === 'twitch_oauth_ok' || k === 'twitch_oauth_error') {
                continue;
            }
            next[k] = v;
        }
        void router.replace({ path: route.path, query: next });
    }
}
function syncTabFromRoute() {
    const t = queryOne(route.query.tab);
    if (t === 'channels' ||
        t === 'rules' ||
        t === 'notifications' ||
        t === 'twitch' ||
        t === 'suspicion' ||
        t === 'display') {
        tab.value = t;
    }
}
onMounted(async () => {
    syncTabFromRoute();
    applyTwitchOAuthQuery();
    try {
        await refresh();
    }
    catch {
        notify({
            id: 'settings-load',
            type: 'error',
            title: 'Settings',
            description: 'Failed to load settings (admin only).',
        });
    }
});
async function addChannel() {
    if (addingChannel.value) {
        return;
    }
    addingChannel.value = true;
    try {
        await DefaultService.createTwitchUser({
            requestBody: { name: newChannel.name.trim() },
        });
        newChannel.name = '';
        notify({
            id: 'settings-add-channel',
            type: 'success',
            title: 'Channels',
            description: 'Channel added.',
        });
        await refresh();
    }
    catch (e) {
        if (e instanceof ApiError && e.status === 400 && e.body && typeof e.body === 'object' && 'message' in e.body) {
            notify({
                id: 'settings-add-channel',
                type: 'error',
                title: 'Channels',
                description: String(e.body.message),
            });
        }
        else {
            notify({
                id: 'settings-add-channel',
                type: 'error',
                title: 'Channels',
                description: 'Could not add channel.',
            });
        }
    }
    finally {
        addingChannel.value = false;
    }
}
async function setChannelMonitored(id, monitored) {
    try {
        await DefaultService.updateTwitchUser({
            requestBody: { id, monitored },
        });
        notify({
            id: 'settings-channel-monitored',
            type: 'success',
            title: 'Channels',
            description: monitored ? 'Monitoring enabled.' : 'Monitoring paused (history kept).',
        });
        await refresh();
    }
    catch {
        notify({
            id: 'settings-channel-monitored',
            type: 'error',
            title: 'Channels',
            description: 'Could not update channel.',
        });
    }
}
function openRuleModal() {
    ruleDraft.regex = '';
    ruleDraft.included_users = '*';
    ruleDraft.denied_users = '';
    ruleDraft.included_channels = '*';
    ruleDraft.denied_channels = '';
    ruleModalOpen.value = true;
}
async function saveRuleModal() {
    if (savingRule.value) {
        return;
    }
    savingRule.value = true;
    try {
        await DefaultService.createRule({
            requestBody: {
                regex: ruleDraft.regex.trim(),
                included_users: ruleDraft.included_users,
                denied_users: ruleDraft.denied_users,
                included_channels: ruleDraft.included_channels,
                denied_channels: ruleDraft.denied_channels,
            },
        });
        ruleModalOpen.value = false;
        notify({ id: 'settings-add-rule', type: 'success', title: 'Rules', description: 'Rule added.' });
        await refresh();
    }
    catch {
        notify({
            id: 'settings-add-rule',
            type: 'error',
            title: 'Rules',
            description: 'Could not add rule (invalid regex?).',
        });
    }
    finally {
        savingRule.value = false;
    }
}
async function deleteRule(id) {
    try {
        await DefaultService.deleteRule({ requestBody: { id } });
        notify({ id: 'settings-del-rule', type: 'success', title: 'Rules', description: 'Rule removed.' });
        await refresh();
    }
    catch {
        notify({ id: 'settings-del-rule', type: 'error', title: 'Rules', description: 'Could not delete rule.' });
    }
}
function openNotifModal() {
    notifDraft.provider = 'telegram';
    notifDraft.enabled = true;
    notifDraft.bot_token = '';
    notifDraft.chat_id = '';
    notifDraft.webhook_url = '';
    notifModalOpen.value = true;
}
function notifSettingsBody() {
    if (notifDraft.provider === 'telegram') {
        return { bot_token: notifDraft.bot_token.trim(), chat_id: notifDraft.chat_id.trim() };
    }
    return { url: notifDraft.webhook_url.trim() };
}
async function saveNotifModal() {
    if (savingNotif.value) {
        return;
    }
    savingNotif.value = true;
    try {
        await DefaultService.createNotification({
            requestBody: {
                provider: notifDraft.provider === 'telegram'
                    ? CreateNotificationRequest.provider.TELEGRAM
                    : CreateNotificationRequest.provider.WEBHOOK,
                settings: notifSettingsBody(),
                enabled: notifDraft.enabled,
            },
        });
        notifModalOpen.value = false;
        notify({ id: 'settings-notif', type: 'success', title: 'Notifications', description: 'Entry added.' });
        await refresh();
    }
    catch {
        notify({
            id: 'settings-notif',
            type: 'error',
            title: 'Notifications',
            description: 'Could not save notification.',
        });
    }
    finally {
        savingNotif.value = false;
    }
}
async function toggleNotifEnabled(n) {
    try {
        await DefaultService.updateNotification({
            requestBody: { id: n.id, enabled: !n.enabled },
        });
        await refresh();
    }
    catch {
        notify({
            id: 'settings-notif-toggle',
            type: 'error',
            title: 'Notifications',
            description: 'Could not update.',
        });
    }
}
async function deleteNotif(id) {
    try {
        await DefaultService.deleteNotification({ requestBody: { id } });
        notify({ id: 'settings-notif-del', type: 'success', title: 'Notifications', description: 'Removed.' });
        await refresh();
    }
    catch {
        notify({ id: 'settings-notif-del', type: 'error', title: 'Notifications', description: 'Could not remove.' });
    }
}
async function setTwitchAccountType(id, accountType) {
    try {
        await DefaultService.updateTwitchAccount({
            requestBody: { id, account_type: accountType },
        });
        notify({
            id: 'settings-twitch-account-type',
            type: 'success',
            title: 'Twitch accounts',
            description: 'Account type updated.',
        });
        await refresh();
    }
    catch {
        notify({
            id: 'settings-twitch-account-type',
            type: 'error',
            title: 'Twitch accounts',
            description: 'Could not update account type.',
        });
    }
}
async function removeTwitchAccount(id) {
    try {
        await DefaultService.deleteTwitchAccount({ requestBody: { id } });
        notify({
            id: 'settings-remove-twitch-account',
            type: 'success',
            title: 'Twitch accounts',
            description: 'Account removed.',
        });
        await refresh();
    }
    catch {
        notify({
            id: 'settings-remove-twitch-account',
            type: 'error',
            title: 'Twitch accounts',
            description: 'Could not remove account.',
        });
    }
}
function twitchOAuthReturnUrl() {
    const path = route.fullPath.includes('?')
        ? `${route.fullPath}&tab=${tab.value}`
        : `${route.fullPath}?tab=${tab.value}`;
    return `${window.location.origin}${window.location.pathname}#${path}`;
}
async function connectTwitchInBrowser() {
    try {
        const res = await DefaultService.startTwitchOAuth({
            requestBody: { return_url: twitchOAuthReturnUrl() },
        });
        window.location.assign(res.authorize_url);
    }
    catch {
        notify({
            id: 'settings-connect-twitch',
            type: 'error',
            title: 'Twitch',
            description: 'Could not start Twitch sign-in.',
        });
    }
}
function setTab(next) {
    tab.value = next;
    void router.replace({ path: route.path, query: { ...route.query, tab: next } });
}
async function addBlacklistEntry() {
    const login = newBlacklistLogin.value.trim().toLowerCase();
    if (!login || savingBlacklist.value) {
        return;
    }
    savingBlacklist.value = true;
    try {
        await DefaultService.setChannelBlacklist({
            requestBody: { login, add: true },
        });
        newBlacklistLogin.value = '';
        channelBlacklist.value = await DefaultService.listChannelBlacklist();
        notify({ id: 'bl-add', type: 'success', title: 'Blacklist', description: 'Channel added.' });
    }
    catch (e) {
        const msg = e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
            ? String(e.body.message)
            : 'Could not update blacklist.';
        notify({ id: 'bl-add', type: 'error', title: 'Blacklist', description: msg });
    }
    finally {
        savingBlacklist.value = false;
    }
}
async function removeBlacklistEntry(login) {
    if (savingBlacklist.value) {
        return;
    }
    savingBlacklist.value = true;
    try {
        await DefaultService.setChannelBlacklist({
            requestBody: { login, add: false },
        });
        channelBlacklist.value = await DefaultService.listChannelBlacklist();
        notify({ id: 'bl-rm', type: 'success', title: 'Blacklist', description: 'Removed.' });
    }
    catch {
        notify({ id: 'bl-rm', type: 'error', title: 'Blacklist', description: 'Could not remove.' });
    }
    finally {
        savingBlacklist.value = false;
    }
}
async function saveSuspicionSettings() {
    if (savingSuspicion.value) {
        return;
    }
    savingSuspicion.value = true;
    try {
        const s = await DefaultService.updateSuspicionSettings({
            requestBody: { ...suspicionDraft },
        });
        Object.assign(suspicionDraft, s);
        notify({
            id: 'sus-settings',
            type: 'success',
            title: 'Suspicion',
            description: 'Settings saved.',
        });
    }
    catch {
        notify({
            id: 'sus-settings',
            type: 'error',
            title: 'Suspicion',
            description: 'Could not save settings.',
        });
    }
    finally {
        savingSuspicion.value = false;
    }
}
/** Twitch IRC allows a limited number of concurrent channel joins per connection. */
const maxMonitoredChannels = 100;
const channelsHeading = computed(() => `Monitored channels (${monitoredChannels.value.length} / ${maxMonitoredChannels})`);
const rulesHeading = computed(() => `Rules (${rulesTotal.value})`);
const twitchHeading = computed(() => `Twitch accounts (${twitchAccountsTotal.value})`);
const __VLS_ctx = {
    ...{},
    ...{},
};
let __VLS_components;
let __VLS_intrinsics;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
/** @type {__VLS_StyleScopedClasses['row-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['tiny']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-form']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "settings-wrap" },
});
/** @type {__VLS_StyleScopedClasses['settings-wrap']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
    ...{ class: "settings" },
});
/** @type {__VLS_StyleScopedClasses['settings']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h1, __VLS_intrinsics.h1)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.nav, __VLS_intrinsics.nav)({
    ...{ class: "tabs" },
    'aria-label': "Settings sections",
});
/** @type {__VLS_StyleScopedClasses['tabs']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setTab('channels');
            // @ts-ignore
            [setTab,];
        } },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.tab === 'channels' }) },
});
/** @type {__VLS_StyleScopedClasses['active']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setTab('rules');
            // @ts-ignore
            [setTab, tab,];
        } },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.tab === 'rules' }) },
});
/** @type {__VLS_StyleScopedClasses['active']} */ ;
(__VLS_ctx.rulesHeading);
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setTab('notifications');
            // @ts-ignore
            [setTab, tab, rulesHeading,];
        } },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.tab === 'notifications' }) },
});
/** @type {__VLS_StyleScopedClasses['active']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setTab('twitch');
            // @ts-ignore
            [setTab, tab,];
        } },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.tab === 'twitch' }) },
});
/** @type {__VLS_StyleScopedClasses['active']} */ ;
(__VLS_ctx.twitchHeading);
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setTab('suspicion');
            // @ts-ignore
            [setTab, tab, twitchHeading,];
        } },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.tab === 'suspicion' }) },
});
/** @type {__VLS_StyleScopedClasses['active']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setTab('display');
            // @ts-ignore
            [setTab, tab,];
        } },
    type: "button",
    ...{ class: ({ active: __VLS_ctx.tab === 'display' }) },
});
/** @type {__VLS_StyleScopedClasses['active']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "panel" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'channels') }, null, null);
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({});
(__VLS_ctx.channelsHeading);
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "hint channels-limit-hint" },
});
/** @type {__VLS_StyleScopedClasses['hint']} */ ;
/** @type {__VLS_StyleScopedClasses['channels-limit-hint']} */ ;
(__VLS_ctx.maxMonitoredChannels);
__VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
    ...{ class: "chan-list" },
});
/** @type {__VLS_StyleScopedClasses['chan-list']} */ ;
for (const [c] of __VLS_vFor((__VLS_ctx.monitoredChannels))) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
        key: (c.id),
    });
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
    (c.username);
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.setChannelMonitored(c.id, false);
                // @ts-ignore
                [tab, tab, channelsHeading, maxMonitoredChannels, monitoredChannels, setChannelMonitored,];
            } },
        type: "button",
        ...{ class: "btn-danger" },
    });
    /** @type {__VLS_StyleScopedClasses['btn-danger']} */ ;
    // @ts-ignore
    [];
}
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.addChannel) },
    ...{ class: "row" },
});
/** @type {__VLS_StyleScopedClasses['row']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    placeholder: "channel name",
    required: true,
});
(__VLS_ctx.newChannel.name);
const __VLS_0 = SubmitButton || SubmitButton;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent1(__VLS_0, new __VLS_0({
    loading: (__VLS_ctx.addingChannel),
}));
const __VLS_2 = __VLS_1({
    loading: (__VLS_ctx.addingChannel),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
const { default: __VLS_5 } = __VLS_3.slots;
// @ts-ignore
[addChannel, newChannel, addingChannel,];
var __VLS_3;
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "panel" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'rules') }, null, null);
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({});
(__VLS_ctx.rulesHeading);
__VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
    ...{ class: "rule-list" },
});
/** @type {__VLS_StyleScopedClasses['rule-list']} */ ;
for (const [r] of __VLS_vFor((__VLS_ctx.rules))) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
        key: (r.id),
        ...{ class: "rule-row" },
    });
    /** @type {__VLS_StyleScopedClasses['rule-row']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.code, __VLS_intrinsics.code)({});
    (r.regex);
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "rule-meta muted small" },
    });
    /** @type {__VLS_StyleScopedClasses['rule-meta']} */ ;
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    /** @type {__VLS_StyleScopedClasses['small']} */ ;
    (r.included_users);
    (r.denied_users);
    (r.included_channels);
    (r.denied_channels);
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.deleteRule(r.id);
                // @ts-ignore
                [tab, rulesHeading, rules, deleteRule,];
            } },
        type: "button",
        ...{ class: "btn-danger" },
    });
    /** @type {__VLS_StyleScopedClasses['btn-danger']} */ ;
    // @ts-ignore
    [];
}
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "row-actions" },
});
/** @type {__VLS_StyleScopedClasses['row-actions']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (__VLS_ctx.openRuleModal) },
    type: "button",
    ...{ class: "btn-secondary" },
});
/** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "panel" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'notifications') }, null, null);
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
    ...{ class: "notif-list" },
});
/** @type {__VLS_StyleScopedClasses['notif-list']} */ ;
for (const [n] of __VLS_vFor((__VLS_ctx.notifications))) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
        key: (n.id),
        ...{ class: "notif-row" },
    });
    /** @type {__VLS_StyleScopedClasses['notif-row']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "tag" },
    });
    /** @type {__VLS_StyleScopedClasses['tag']} */ ;
    (n.provider);
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: ({ muted: !n.enabled }) },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    (n.enabled ? 'on' : 'off');
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.toggleNotifEnabled(n);
                // @ts-ignore
                [tab, openRuleModal, notifications, toggleNotifEnabled,];
            } },
        type: "button",
        ...{ class: "btn-secondary tiny" },
    });
    /** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
    /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.deleteNotif(n.id);
                // @ts-ignore
                [deleteNotif,];
            } },
        type: "button",
        ...{ class: "btn-danger tiny" },
    });
    /** @type {__VLS_StyleScopedClasses['btn-danger']} */ ;
    /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
    // @ts-ignore
    [];
}
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (__VLS_ctx.openNotifModal) },
    type: "button",
    ...{ class: "btn-secondary" },
});
/** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "panel" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'twitch') }, null, null);
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({});
(__VLS_ctx.twitchHeading);
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "hint" },
});
/** @type {__VLS_StyleScopedClasses['hint']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "row-actions" },
});
/** @type {__VLS_StyleScopedClasses['row-actions']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (__VLS_ctx.connectTwitchInBrowser) },
    type: "button",
    ...{ class: "btn-secondary" },
});
/** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
    ...{ class: "twitch-account-list" },
});
/** @type {__VLS_StyleScopedClasses['twitch-account-list']} */ ;
for (const [a] of __VLS_vFor((__VLS_ctx.twitchStore.accounts))) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
        key: (a.id),
        ...{ class: "twitch-account-row" },
    });
    /** @type {__VLS_StyleScopedClasses['twitch-account-row']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "twitch-account-meta" },
    });
    /** @type {__VLS_StyleScopedClasses['twitch-account-meta']} */ ;
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "twitch-account-name" },
    });
    /** @type {__VLS_StyleScopedClasses['twitch-account-name']} */ ;
    (a.username);
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({
        ...{ class: "twitch-account-kind" },
    });
    /** @type {__VLS_StyleScopedClasses['twitch-account-kind']} */ ;
    (a.account_type);
    __VLS_asFunctionalElement1(__VLS_intrinsics.div, __VLS_intrinsics.div)({
        ...{ class: "row-actions wrap" },
    });
    /** @type {__VLS_StyleScopedClasses['row-actions']} */ ;
    /** @type {__VLS_StyleScopedClasses['wrap']} */ ;
    if (a.account_type !== __VLS_ctx.TwitchAccount.account_type.BOT) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(a.account_type !== __VLS_ctx.TwitchAccount.account_type.BOT))
                        return;
                    __VLS_ctx.setTwitchAccountType(a.id, __VLS_ctx.UpdateTwitchAccountPostRequest.account_type.BOT);
                    // @ts-ignore
                    [tab, twitchHeading, openNotifModal, connectTwitchInBrowser, twitchStore, TwitchAccount, setTwitchAccountType, UpdateTwitchAccountPostRequest,];
                } },
            type: "button",
            ...{ class: "btn-secondary" },
        });
        /** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
    }
    if (a.account_type !== __VLS_ctx.TwitchAccount.account_type.MAIN) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(a.account_type !== __VLS_ctx.TwitchAccount.account_type.MAIN))
                        return;
                    __VLS_ctx.setTwitchAccountType(a.id, __VLS_ctx.UpdateTwitchAccountPostRequest.account_type.MAIN);
                    // @ts-ignore
                    [TwitchAccount, setTwitchAccountType, UpdateTwitchAccountPostRequest,];
                } },
            type: "button",
            ...{ class: "btn-secondary" },
        });
        /** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
    }
    __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.removeTwitchAccount(a.id);
                // @ts-ignore
                [removeTwitchAccount,];
            } },
        type: "button",
        ...{ class: "btn-danger" },
    });
    /** @type {__VLS_StyleScopedClasses['btn-danger']} */ ;
    // @ts-ignore
    [];
}
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "panel" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'suspicion') }, null, null);
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "hint" },
});
/** @type {__VLS_StyleScopedClasses['hint']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h3, __VLS_intrinsics.h3)({
    ...{ class: "subh" },
});
/** @type {__VLS_StyleScopedClasses['subh']} */ ;
if (!__VLS_ctx.channelBlacklist.length) {
    __VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
        ...{ class: "muted small" },
    });
    /** @type {__VLS_StyleScopedClasses['muted']} */ ;
    /** @type {__VLS_StyleScopedClasses['small']} */ ;
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.ul, __VLS_intrinsics.ul)({
        ...{ class: "bl-list" },
    });
    /** @type {__VLS_StyleScopedClasses['bl-list']} */ ;
    for (const [login] of __VLS_vFor((__VLS_ctx.channelBlacklist))) {
        __VLS_asFunctionalElement1(__VLS_intrinsics.li, __VLS_intrinsics.li)({
            key: (login),
            ...{ class: "bl-row" },
        });
        /** @type {__VLS_StyleScopedClasses['bl-row']} */ ;
        __VLS_asFunctionalElement1(__VLS_intrinsics.code, __VLS_intrinsics.code)({});
        (login);
        __VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
            ...{ onClick: (...[$event]) => {
                    if (!!(!__VLS_ctx.channelBlacklist.length))
                        return;
                    __VLS_ctx.removeBlacklistEntry(login);
                    // @ts-ignore
                    [tab, channelBlacklist, channelBlacklist, removeBlacklistEntry,];
                } },
            type: "button",
            ...{ class: "btn-danger tiny" },
            disabled: (__VLS_ctx.savingBlacklist),
        });
        /** @type {__VLS_StyleScopedClasses['btn-danger']} */ ;
        /** @type {__VLS_StyleScopedClasses['tiny']} */ ;
        // @ts-ignore
        [savingBlacklist,];
    }
}
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.addBlacklistEntry) },
    ...{ class: "row bl-add" },
});
/** @type {__VLS_StyleScopedClasses['row']} */ ;
/** @type {__VLS_StyleScopedClasses['bl-add']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    placeholder: "channel login",
    autocomplete: "off",
});
(__VLS_ctx.newBlacklistLogin);
const __VLS_6 = SubmitButton || SubmitButton;
// @ts-ignore
const __VLS_7 = __VLS_asFunctionalComponent1(__VLS_6, new __VLS_6({
    loading: (__VLS_ctx.savingBlacklist),
}));
const __VLS_8 = __VLS_7({
    loading: (__VLS_ctx.savingBlacklist),
}, ...__VLS_functionalComponentArgsRest(__VLS_7));
const { default: __VLS_11 } = __VLS_9.slots;
// @ts-ignore
[savingBlacklist, addBlacklistEntry, newBlacklistLogin,];
var __VLS_9;
__VLS_asFunctionalElement1(__VLS_intrinsics.h3, __VLS_intrinsics.h3)({
    ...{ class: "subh" },
});
/** @type {__VLS_StyleScopedClasses['subh']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.saveSuspicionSettings) },
    ...{ class: "stack sus-form" },
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['sus-form']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "row-inline" },
});
/** @type {__VLS_StyleScopedClasses['row-inline']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "checkbox",
});
(__VLS_ctx.suspicionDraft.auto_check_account_age);
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "stack gap-setting" },
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-setting']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "number",
    min: "0",
    max: "36500",
    required: true,
});
(__VLS_ctx.suspicionDraft.account_age_sus_days);
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "row-inline" },
});
/** @type {__VLS_StyleScopedClasses['row-inline']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "checkbox",
});
(__VLS_ctx.suspicionDraft.auto_check_blacklist);
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "row-inline" },
});
/** @type {__VLS_StyleScopedClasses['row-inline']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "checkbox",
});
(__VLS_ctx.suspicionDraft.auto_check_low_follows);
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "stack gap-setting" },
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-setting']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "number",
    min: "0",
    max: "1000000",
    required: true,
});
(__VLS_ctx.suspicionDraft.low_follows_threshold);
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "stack gap-setting" },
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-setting']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "number",
    min: "1",
    max: "500",
    required: true,
});
(__VLS_ctx.suspicionDraft.max_gql_follow_pages);
__VLS_asFunctionalElement1(__VLS_intrinsics.p, __VLS_intrinsics.p)({
    ...{ class: "row-actions" },
});
/** @type {__VLS_StyleScopedClasses['row-actions']} */ ;
const __VLS_12 = SubmitButton || SubmitButton;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent1(__VLS_12, new __VLS_12({
    loading: (__VLS_ctx.savingSuspicion),
}));
const __VLS_14 = __VLS_13({
    loading: (__VLS_ctx.savingSuspicion),
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
const { default: __VLS_17 } = __VLS_15.slots;
// @ts-ignore
[saveSuspicionSettings, suspicionDraft, suspicionDraft, suspicionDraft, suspicionDraft, suspicionDraft, suspicionDraft, savingSuspicion,];
var __VLS_15;
__VLS_asFunctionalElement1(__VLS_intrinsics.section, __VLS_intrinsics.section)({
    ...{ class: "panel" },
});
__VLS_asFunctionalDirective(__VLS_directives.vShow, {})(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tab === 'display') }, null, null);
/** @type {__VLS_StyleScopedClasses['panel']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.h2, __VLS_intrinsics.h2)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "stack gap-setting" },
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['gap-setting']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    ...{ onChange: (...[$event]) => {
            __VLS_ctx.watchPrefs.setChatGapMinutes(__VLS_ctx.chatGapMinutes);
            // @ts-ignore
            [tab, watchPrefs, chatGapMinutes,];
        } },
    ...{ onBlur: (...[$event]) => {
            __VLS_ctx.watchPrefs.setChatGapMinutes(__VLS_ctx.chatGapMinutes);
            // @ts-ignore
            [watchPrefs, chatGapMinutes,];
        } },
    type: "number",
    min: "1",
    max: "1440",
});
(__VLS_ctx.chatGapMinutes);
const __VLS_18 = AppModal || AppModal;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent1(__VLS_18, new __VLS_18({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.ruleModalOpen),
    title: "New rule",
}));
const __VLS_20 = __VLS_19({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.ruleModalOpen),
    title: "New rule",
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
let __VLS_23;
const __VLS_24 = ({ close: {} },
    { onClose: (...[$event]) => {
            __VLS_ctx.ruleModalOpen = false;
            // @ts-ignore
            [chatGapMinutes, ruleModalOpen, ruleModalOpen,];
        } });
const { default: __VLS_25 } = __VLS_21.slots;
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.saveRuleModal) },
    ...{ class: "stack modal-form" },
    autocomplete: "off",
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-form']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    name: "dredge_rule_regex",
    required: true,
    placeholder: "pattern",
    autocomplete: "off",
});
(__VLS_ctx.ruleDraft.regex);
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    name: "dredge_rule_included_users",
    autocomplete: "off",
});
(__VLS_ctx.ruleDraft.included_users);
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    name: "dredge_rule_denied_users",
    autocomplete: "off",
});
(__VLS_ctx.ruleDraft.denied_users);
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    name: "dredge_rule_included_channels",
    autocomplete: "off",
});
(__VLS_ctx.ruleDraft.included_channels);
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    name: "dredge_rule_denied_channels",
    autocomplete: "off",
});
(__VLS_ctx.ruleDraft.denied_channels);
__VLS_asFunctionalElement1(__VLS_intrinsics.footer, __VLS_intrinsics.footer)({
    ...{ class: "modal-actions" },
});
/** @type {__VLS_StyleScopedClasses['modal-actions']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.ruleModalOpen = false;
            // @ts-ignore
            [ruleModalOpen, saveRuleModal, ruleDraft, ruleDraft, ruleDraft, ruleDraft, ruleDraft,];
        } },
    type: "button",
    ...{ class: "btn-secondary" },
});
/** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
const __VLS_26 = SubmitButton || SubmitButton;
// @ts-ignore
const __VLS_27 = __VLS_asFunctionalComponent1(__VLS_26, new __VLS_26({
    loading: (__VLS_ctx.savingRule),
}));
const __VLS_28 = __VLS_27({
    loading: (__VLS_ctx.savingRule),
}, ...__VLS_functionalComponentArgsRest(__VLS_27));
const { default: __VLS_31 } = __VLS_29.slots;
// @ts-ignore
[savingRule,];
var __VLS_29;
// @ts-ignore
[];
var __VLS_21;
var __VLS_22;
const __VLS_32 = AppModal || AppModal;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent1(__VLS_32, new __VLS_32({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.notifModalOpen),
    title: "New notification",
}));
const __VLS_34 = __VLS_33({
    ...{ 'onClose': {} },
    open: (__VLS_ctx.notifModalOpen),
    title: "New notification",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
let __VLS_37;
const __VLS_38 = ({ close: {} },
    { onClose: (...[$event]) => {
            __VLS_ctx.notifModalOpen = false;
            // @ts-ignore
            [notifModalOpen, notifModalOpen,];
        } });
const { default: __VLS_39 } = __VLS_35.slots;
__VLS_asFunctionalElement1(__VLS_intrinsics.form, __VLS_intrinsics.form)({
    ...{ onSubmit: (__VLS_ctx.saveNotifModal) },
    ...{ class: "stack modal-form" },
});
/** @type {__VLS_StyleScopedClasses['stack']} */ ;
/** @type {__VLS_StyleScopedClasses['modal-form']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
__VLS_asFunctionalElement1(__VLS_intrinsics.select, __VLS_intrinsics.select)({
    value: (__VLS_ctx.notifDraft.provider),
});
__VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
    value: "telegram",
});
__VLS_asFunctionalElement1(__VLS_intrinsics.option, __VLS_intrinsics.option)({
    value: "webhook",
});
__VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({
    ...{ class: "row-inline" },
});
/** @type {__VLS_StyleScopedClasses['row-inline']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.input)({
    type: "checkbox",
});
(__VLS_ctx.notifDraft.enabled);
__VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
if (__VLS_ctx.notifDraft.provider === 'telegram') {
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        type: "password",
        autocomplete: "off",
    });
    (__VLS_ctx.notifDraft.bot_token);
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        autocomplete: "off",
    });
    (__VLS_ctx.notifDraft.chat_id);
}
else {
    __VLS_asFunctionalElement1(__VLS_intrinsics.label, __VLS_intrinsics.label)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.span, __VLS_intrinsics.span)({});
    __VLS_asFunctionalElement1(__VLS_intrinsics.input)({
        type: "url",
        name: "dredge_notif_webhook_url",
        autocomplete: "off",
    });
    (__VLS_ctx.notifDraft.webhook_url);
}
__VLS_asFunctionalElement1(__VLS_intrinsics.footer, __VLS_intrinsics.footer)({
    ...{ class: "modal-actions" },
});
/** @type {__VLS_StyleScopedClasses['modal-actions']} */ ;
__VLS_asFunctionalElement1(__VLS_intrinsics.button, __VLS_intrinsics.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.notifModalOpen = false;
            // @ts-ignore
            [notifModalOpen, saveNotifModal, notifDraft, notifDraft, notifDraft, notifDraft, notifDraft, notifDraft,];
        } },
    type: "button",
    ...{ class: "btn-secondary" },
});
/** @type {__VLS_StyleScopedClasses['btn-secondary']} */ ;
const __VLS_40 = SubmitButton || SubmitButton;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent1(__VLS_40, new __VLS_40({
    loading: (__VLS_ctx.savingNotif),
}));
const __VLS_42 = __VLS_41({
    loading: (__VLS_ctx.savingNotif),
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
const { default: __VLS_45 } = __VLS_43.slots;
// @ts-ignore
[savingNotif,];
var __VLS_43;
// @ts-ignore
[];
var __VLS_35;
var __VLS_36;
// @ts-ignore
[];
const __VLS_export = (await import('vue')).defineComponent({});
export default {};
