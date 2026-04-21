<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core';
import { storeToRefs } from 'pinia';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import {
  ApiError,
  CreateNotificationRequest,
  DefaultService,
  NotificationEntry,
  TwitchAccount,
  UpdateTwitchAccountPostRequest,
} from '../api/generated';
import type { PatchAiSettingsRequest } from '../api/generated';
import type {
  ChannelDiscoverySettings,
  DiscoveryCandidate,
  IrcMonitorSettings,
  IrcMonitorStatus,
} from '../api/generated';
import type { SuspicionSettings } from '../api/generated';
import type { Rule } from '../api/generated';
import AppModal from '../components/AppModal.vue';
import SubmitButton from '../components/SubmitButton.vue';
import { isChannelJoinedOnIrc } from '../lib/ircMonitorJoined';
import { notify } from '../lib/notify';
import { useChannelsStore } from '../stores/channels';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';
import { useWatchPreferencesStore } from '../stores/watchPreferences';

type SettingsTab =
  | 'channels'
  | 'rules'
  | 'notifications'
  | 'twitch'
  | 'channelDiscovery'
  | 'suspicion'
  | 'display'
  | 'ai';

const route = useRoute();
const router = useRouter();
const tab = ref<SettingsTab>('channels');
const channelsStore = useChannelsStore();
const { monitoredChannels } = storeToRefs(channelsStore);
const twitchStore = useTwitchAccountsStore();
const watchPrefs = useWatchPreferencesStore();
const { chatGapMinutes } = storeToRefs(watchPrefs);
const rules = ref<Rule[]>([]);
const rulesTotal = ref(0);
const twitchAccountsTotal = ref(0);
const notifications = ref<NotificationEntry[]>([]);
const channelBlacklist = ref<string[]>([]);
const newBlacklistLogin = ref('');
const blacklistFilter = ref('');
const blacklistSort = ref<'name-asc' | 'name-desc'>('name-asc');
const aiBaseUrl = ref('');
const aiModel = ref('');
const aiTokenDraft = ref('');
const aiHasToken = ref(false);
const aiTokenLast4 = ref('');
const savingAi = ref(false);

const suspicionDraft = reactive<SuspicionSettings>({
  auto_check_account_age: true,
  account_age_sus_days: 14,
  auto_check_blacklist: true,
  auto_check_low_follows: true,
  low_follows_threshold: 10,
  max_gql_follow_pages: 1,
});

const newChannel = reactive({ name: '' });
const channelsFilter = ref('');
const channelsSort = ref<'name-asc' | 'name-desc' | 'id-desc' | 'id-asc'>('name-asc');

const regexTestModalOpen = ref(false);
const regexTestPattern = ref('');
const regexTestSample = ref('');
const regexTestCaseInsensitive = ref(false);
const regexTestMatches = ref<boolean | null>(null);
const regexTestCompileError = ref<string | null>(null);

const notifModalOpen = ref(false);
const notifDraft = reactive({
  provider: 'telegram' as 'telegram' | 'webhook',
  enabled: true,
  bot_token: '',
  chat_id: '',
  webhook_url: '',
});

const addingChannel = ref(false);
const savingNotif = ref(false);
const savingBlacklist = ref(false);
const savingSuspicion = ref(false);
const savingIrcMonitor = ref(false);
const savingDiscovery = ref(false);
/** '' = anonymous IRC; otherwise linked account Twitch user id as string */
const ircOAuthAccountId = ref('');
const enrichmentCooldownHours = ref(24);

const discoveryDraft = reactive<ChannelDiscoverySettings>({
  enabled: false,
  poll_interval_seconds: 3600,
  game_id: '',
  min_live_viewers: 0,
  required_stream_tags: [],
  max_stream_pages_per_run: 20,
});
const discoveryTagsText = ref('');
const discoveryCandidates = ref<DiscoveryCandidate[]>([]);
let discoveryPollId: number | null = null;

const ircStatus = ref<IrcMonitorStatus | null>(null);
let ircPollId: number | null = null;

async function fetchIrcStatus(): Promise<void> {
  try {
    ircStatus.value = await DefaultService.getIrcMonitorStatus();
  } catch {
    ircStatus.value = null;
  }
}

function stopIrcPoll(): void {
  if (ircPollId !== null) {
    clearInterval(ircPollId);
    ircPollId = null;
  }
}

function startIrcPoll(): void {
  stopIrcPoll();
  if (tab.value !== 'channels') {
    return;
  }
  void fetchIrcStatus();
  ircPollId = window.setInterval(() => {
    void fetchIrcStatus();
  }, 4000);
}

async function loadDiscoveryCandidates(): Promise<void> {
  try {
    discoveryCandidates.value = await DefaultService.listChannelDiscoveryCandidates();
  } catch {
    discoveryCandidates.value = [];
  }
}

function stopDiscoveryPoll(): void {
  if (discoveryPollId !== null) {
    window.clearInterval(discoveryPollId);
    discoveryPollId = null;
  }
}

function startDiscoveryPoll(): void {
  stopDiscoveryPoll();
  discoveryPollId = window.setInterval(() => {
    void loadDiscoveryCandidates();
  }, 30_000);
}

watch(
  () => tab.value,
  (t) => {
    if (t === 'channels') {
      startIrcPoll();
    } else {
      stopIrcPoll();
    }

    if (t === 'channelDiscovery') {
      void loadDiscoveryCandidates();
      startDiscoveryPoll();
    } else {
      stopDiscoveryPoll();
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  stopIrcPoll();
  stopDiscoveryPoll();
});

async function refresh(): Promise<void> {
  const [rulesList, rCount, taCount, notifList, bl, sus, ircMon, disc, cand, aiSt] = await Promise.all([
    DefaultService.listRules(),
    DefaultService.countRules(),
    DefaultService.countTwitchAccounts(),
    DefaultService.listNotifications({}),
    DefaultService.listChannelBlacklist(),
    DefaultService.getSuspicionSettings(),
    DefaultService.getIrcMonitorSettings(),
    DefaultService.getChannelDiscoverySettings(),
    DefaultService.listChannelDiscoveryCandidates(),
    DefaultService.getAiSettings(),
  ]);
  await Promise.all([channelsStore.fetch(), twitchStore.fetch()]);
  rules.value = rulesList;
  rulesTotal.value = rCount.total;
  twitchAccountsTotal.value = taCount.total;
  notifications.value = notifList;
  channelBlacklist.value = bl;
  Object.assign(suspicionDraft, sus);
  applyIrcMonitorDraft(ircMon);
  applyDiscoveryDraft(disc);
  discoveryCandidates.value = cand;
  aiBaseUrl.value = aiSt.base_url ?? '';
  aiModel.value = aiSt.model ?? '';
  aiHasToken.value = Boolean(aiSt.has_token);
  aiTokenLast4.value = aiSt.token_last4 ?? '';
  aiTokenDraft.value = '';
}

function applyDiscoveryDraft(s: ChannelDiscoverySettings): void {
  discoveryDraft.enabled = s.enabled;
  discoveryDraft.poll_interval_seconds =
    typeof s.poll_interval_seconds === 'number' && s.poll_interval_seconds >= 60
      ? s.poll_interval_seconds
      : 3600;
  discoveryDraft.game_id = s.game_id ?? '';
  discoveryDraft.min_live_viewers =
    typeof s.min_live_viewers === 'number' && s.min_live_viewers >= 0 ? s.min_live_viewers : 0;
  discoveryDraft.required_stream_tags = Array.isArray(s.required_stream_tags)
    ? [...s.required_stream_tags]
    : [];
  discoveryDraft.max_stream_pages_per_run =
    typeof s.max_stream_pages_per_run === 'number' && s.max_stream_pages_per_run >= 1
      ? s.max_stream_pages_per_run
      : 20;
  discoveryTagsText.value = (discoveryDraft.required_stream_tags || []).join('\n');
}

function applyIrcMonitorDraft(s: IrcMonitorSettings): void {
  ircOAuthAccountId.value =
    s.oauth_twitch_account_id !== null && s.oauth_twitch_account_id !== undefined
      ? String(s.oauth_twitch_account_id)
      : '';
  enrichmentCooldownHours.value =
    typeof s.enrichment_cooldown_hours === 'number' && s.enrichment_cooldown_hours > 0
      ? s.enrichment_cooldown_hours
      : 24;
}

function queryOne(v: unknown): string | undefined {
  if (typeof v === 'string') {
    return v;
  }
  if (Array.isArray(v) && typeof v[0] === 'string') {
    return v[0];
  }
  return undefined;
}

function applyTwitchOAuthQuery(): void {
  const okParam = queryOne(route.query.twitch_oauth_ok);
  const errParam = queryOne(route.query.twitch_oauth_error);
  if (okParam === '1') {
    notify({
      id: 'settings-twitch-oauth-callback',
      type: 'success',
      title: 'Twitch',
      description: 'Twitch account linked.',
    });
  } else if (errParam) {
    notify({
      id: 'settings-twitch-oauth-callback',
      type: 'error',
      title: 'Twitch',
      description: errParam,
    });
  }
  if (okParam !== undefined || errParam !== undefined) {
    const next: Record<string, string | string[]> = {};
    for (const [k, v] of Object.entries(route.query)) {
      if (k === 'twitch_oauth_ok' || k === 'twitch_oauth_error') {
        continue;
      }
      next[k] = v as string | string[];
    }
    void router.replace({ path: route.path, query: next });
  }
}

function syncTabFromRoute(): void {
  const t = queryOne(route.query.tab);
  if (
    t === 'channels' ||
    t === 'rules' ||
    t === 'notifications' ||
    t === 'twitch' ||
    t === 'channelDiscovery' ||
    t === 'suspicion' ||
    t === 'display' ||
    t === 'ai'
  ) {
    tab.value = t;
  }
}

onMounted(async () => {
  syncTabFromRoute();
  applyTwitchOAuthQuery();
  try {
    await refresh();
  } catch {
    notify({
      id: 'settings-load',
      type: 'error',
      title: 'Settings',
      description: 'Failed to load settings (admin only).',
    });
  }
});

async function addChannel(): Promise<void> {
  if (addingChannel.value) {
    return;
  }
  addingChannel.value = true;
  try {
    const created = await DefaultService.createTwitchUser({
      requestBody: { name: newChannel.name.trim() },
    });
    newChannel.name = '';
    notify({
      id: 'settings-add-channel',
      type: 'success',
      title: 'Channels',
      description: 'Channel added.',
    });
    await channelsStore.fetch();
    await router.push({ name: 'user', params: { id: String(created.id) } });
  } catch (e: unknown) {
    if (e instanceof ApiError && e.status === 400 && e.body && typeof e.body === 'object' && 'message' in e.body) {
      notify({
        id: 'settings-add-channel',
        type: 'error',
        title: 'Channels',
        description: String((e.body as { message: string }).message),
      });
    } else {
      notify({
        id: 'settings-add-channel',
        type: 'error',
        title: 'Channels',
        description: 'Could not add channel.',
      });
    }
  } finally {
    addingChannel.value = false;
  }
}

const runRegexTestDebounced = useDebounceFn(async () => {
  regexTestMatches.value = null;
  regexTestCompileError.value = null;
  try {
    const res = await DefaultService.testRuleRegex({
      requestBody: {
        pattern: regexTestPattern.value,
        sample: regexTestSample.value,
        case_insensitive: regexTestCaseInsensitive.value,
      },
    });
    regexTestMatches.value = res.matches;
    regexTestCompileError.value = res.compile_error ?? null;
  } catch {
    regexTestCompileError.value = 'Request failed';
  }
}, 320);

watch([regexTestPattern, regexTestSample, regexTestCaseInsensitive], () => {
  if (!regexTestModalOpen.value) {
    return;
  }
  void runRegexTestDebounced();
});

function openRegexTestModal(): void {
  regexTestPattern.value = '';
  regexTestSample.value = '';
  regexTestCaseInsensitive.value = false;
  regexTestMatches.value = null;
  regexTestCompileError.value = null;
  regexTestModalOpen.value = true;
  void runRegexTestDebounced();
}

async function deleteRule(id: number): Promise<void> {
  try {
    await DefaultService.deleteRule({ requestBody: { id } });
    notify({ id: 'settings-del-rule', type: 'success', title: 'Rules', description: 'Rule removed.' });
    await refresh();
  } catch {
    notify({ id: 'settings-del-rule', type: 'error', title: 'Rules', description: 'Could not delete rule.' });
  }
}

function openNotifModal(): void {
  notifDraft.provider = 'telegram';
  notifDraft.enabled = true;
  notifDraft.bot_token = '';
  notifDraft.chat_id = '';
  notifDraft.webhook_url = '';
  notifModalOpen.value = true;
}

function notifSettingsBody(): Record<string, unknown> {
  if (notifDraft.provider === 'telegram') {
    return { bot_token: notifDraft.bot_token.trim(), chat_id: notifDraft.chat_id.trim() };
  }
  return { url: notifDraft.webhook_url.trim() };
}

async function saveNotifModal(): Promise<void> {
  if (savingNotif.value) {
    return;
  }
  savingNotif.value = true;
  try {
    await DefaultService.createNotification({
      requestBody: {
        provider:
          notifDraft.provider === 'telegram'
            ? CreateNotificationRequest.provider.TELEGRAM
            : CreateNotificationRequest.provider.WEBHOOK,
        settings: notifSettingsBody() as Record<string, unknown>,
        enabled: notifDraft.enabled,
      },
    });
    notifModalOpen.value = false;
    notify({ id: 'settings-notif', type: 'success', title: 'Notifications', description: 'Entry added.' });
    await refresh();
  } catch {
    notify({
      id: 'settings-notif',
      type: 'error',
      title: 'Notifications',
      description: 'Could not save notification.',
    });
  } finally {
    savingNotif.value = false;
  }
}

async function toggleNotifEnabled(n: NotificationEntry): Promise<void> {
  try {
    await DefaultService.updateNotification({
      requestBody: { id: n.id, enabled: !n.enabled },
    });
    await refresh();
  } catch {
    notify({
      id: 'settings-notif-toggle',
      type: 'error',
      title: 'Notifications',
      description: 'Could not update.',
    });
  }
}

async function deleteNotif(id: number): Promise<void> {
  try {
    await DefaultService.deleteNotification({ requestBody: { id } });
    notify({ id: 'settings-notif-del', type: 'success', title: 'Notifications', description: 'Removed.' });
    await refresh();
  } catch {
    notify({ id: 'settings-notif-del', type: 'error', title: 'Notifications', description: 'Could not remove.' });
  }
}

async function setTwitchAccountType(id: number, accountType: UpdateTwitchAccountPostRequest.account_type): Promise<void> {
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
  } catch {
    notify({
      id: 'settings-twitch-account-type',
      type: 'error',
      title: 'Twitch accounts',
      description: 'Could not update account type.',
    });
  }
}

async function removeTwitchAccount(id: number): Promise<void> {
  try {
    await DefaultService.deleteTwitchAccount({ requestBody: { id } });
    notify({
      id: 'settings-remove-twitch-account',
      type: 'success',
      title: 'Twitch accounts',
      description: 'Account removed.',
    });
    await refresh();
  } catch {
    notify({
      id: 'settings-remove-twitch-account',
      type: 'error',
      title: 'Twitch accounts',
      description: 'Could not remove account.',
    });
  }
}

function twitchOAuthReturnUrl(): string {
  const path = route.fullPath.includes('?')
    ? `${route.fullPath}&tab=${tab.value}`
    : `${route.fullPath}?tab=${tab.value}`;
  return `${window.location.origin}${window.location.pathname}#${path}`;
}

async function connectTwitchInBrowser(): Promise<void> {
  try {
    const res = await DefaultService.startTwitchOAuth({
      requestBody: { return_url: twitchOAuthReturnUrl() },
    });
    window.location.assign(res.authorize_url);
  } catch {
    notify({
      id: 'settings-connect-twitch',
      type: 'error',
      title: 'Twitch',
      description: 'Could not start Twitch sign-in.',
    });
  }
}

function setTab(next: SettingsTab): void {
  tab.value = next;
  void router.replace({ path: route.path, query: { ...route.query, tab: next } });
}

async function addBlacklistEntry(): Promise<void> {
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
  } catch (e: unknown) {
    const msg =
      e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
        ? String((e.body as { message: string }).message)
        : 'Could not update blacklist.';
    notify({ id: 'bl-add', type: 'error', title: 'Blacklist', description: msg });
  } finally {
    savingBlacklist.value = false;
  }
}

async function removeBlacklistEntry(login: string): Promise<void> {
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
  } catch {
    notify({ id: 'bl-rm', type: 'error', title: 'Blacklist', description: 'Could not remove.' });
  } finally {
    savingBlacklist.value = false;
  }
}

async function saveIrcMonitorSettings(): Promise<void> {
  if (savingIrcMonitor.value) {
    return;
  }
  savingIrcMonitor.value = true;
  try {
    const oauthTwitchAccountId =
      ircOAuthAccountId.value === '' ? null : Number(ircOAuthAccountId.value);
    const body: IrcMonitorSettings = {
      oauth_twitch_account_id: oauthTwitchAccountId,
      enrichment_cooldown_hours: Math.max(1, Math.floor(enrichmentCooldownHours.value || 24)),
    };
    const s = await DefaultService.updateIrcMonitorSettings({ requestBody: body });
    applyIrcMonitorDraft(s);
    notify({
      id: 'irc-monitor-settings',
      type: 'success',
      title: 'IRC',
      description: 'IRC monitor settings saved.',
    });
  } catch {
    notify({
      id: 'irc-monitor-settings',
      type: 'error',
      title: 'IRC',
      description: 'Could not save IRC monitor settings.',
    });
  } finally {
    savingIrcMonitor.value = false;
  }
}

async function saveChannelDiscoverySettings(): Promise<void> {
  if (savingDiscovery.value) {
    return;
  }
  savingDiscovery.value = true;
  try {
    const tags = discoveryTagsText.value
      .split('\n')
      .map((x) => x.trim())
      .filter((x) => x.length > 0);
    const body: ChannelDiscoverySettings = {
      enabled: discoveryDraft.enabled,
      poll_interval_seconds: Math.max(60, Math.floor(discoveryDraft.poll_interval_seconds) || 3600),
      game_id: discoveryDraft.game_id.trim(),
      min_live_viewers: Math.max(0, Math.floor(discoveryDraft.min_live_viewers) || 0),
      required_stream_tags: tags,
      max_stream_pages_per_run: Math.max(1, Math.floor(discoveryDraft.max_stream_pages_per_run) || 20),
    };
    const s = await DefaultService.updateChannelDiscoverySettings({ requestBody: body });
    applyDiscoveryDraft(s);
    notify({
      id: 'channel-discovery-settings',
      type: 'success',
      title: 'Channel discovery',
      description: 'Settings saved.',
    });
  } catch (e: unknown) {
    const msg =
      e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
        ? String((e.body as { message: string }).message)
        : 'Could not save channel discovery settings.';
    notify({ id: 'channel-discovery-settings', type: 'error', title: 'Channel discovery', description: msg });
  } finally {
    savingDiscovery.value = false;
  }
}

async function approveDiscoveryUser(userId: number): Promise<void> {
  try {
    await DefaultService.approveChannelDiscoveryCandidate({ twitchUserId: userId });
    await channelsStore.fetch();
    await loadDiscoveryCandidates();
    notify({
      id: 'discovery-approve',
      type: 'success',
      title: 'Channel discovery',
      description: 'Channel is now monitored.',
    });
  } catch {
    notify({
      id: 'discovery-approve',
      type: 'error',
      title: 'Channel discovery',
      description: 'Could not approve.',
    });
  }
}

async function denyDiscoveryUser(userId: number): Promise<void> {
  try {
    await DefaultService.denyChannelDiscoveryCandidate({ twitchUserId: userId });
    await loadDiscoveryCandidates();
    notify({
      id: 'discovery-deny',
      type: 'success',
      title: 'Channel discovery',
      description: 'Channel denied for future discovery.',
    });
  } catch {
    notify({
      id: 'discovery-deny',
      type: 'error',
      title: 'Channel discovery',
      description: 'Could not deny.',
    });
  }
}

async function saveAiSettings(): Promise<void> {
  if (savingAi.value) {
    return;
  }
  savingAi.value = true;
  try {
    const body: PatchAiSettingsRequest = {};
    if (aiBaseUrl.value.trim()) {
      body.base_url = aiBaseUrl.value.trim();
    }
    if (aiModel.value.trim()) {
      body.model = aiModel.value.trim();
    }
    if (aiTokenDraft.value.trim()) {
      body.api_token = aiTokenDraft.value.trim();
    }
    const s = await DefaultService.patchAiSettings({ requestBody: body });
    aiBaseUrl.value = s.base_url ?? '';
    aiModel.value = s.model ?? '';
    aiHasToken.value = Boolean(s.has_token);
    aiTokenLast4.value = s.token_last4 ?? '';
    aiTokenDraft.value = '';
    notify({ id: 'ai-settings', type: 'success', title: 'AI', description: 'Saved.' });
  } catch {
    notify({ id: 'ai-settings', type: 'error', title: 'AI', description: 'Could not save AI settings.' });
  } finally {
    savingAi.value = false;
  }
}

async function saveSuspicionSettings(): Promise<void> {
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
  } catch {
    notify({
      id: 'sus-settings',
      type: 'error',
      title: 'Suspicion',
      description: 'Could not save settings.',
    });
  } finally {
    savingSuspicion.value = false;
  }
}

/** Twitch IRC allows a limited number of concurrent channel joins per connection. */
const maxIrcJoinedChannels = 100;

const channelsHeading = computed(() => `Monitored channels (${monitoredChannels.value.length})`);

const monitoredCount = computed(() => monitoredChannels.value.length);
const filteredMonitoredChannels = computed(() => {
  const q = channelsFilter.value.trim().toLowerCase();
  const channels = monitoredChannels.value.filter((c) => (q ? c.username.toLowerCase().includes(q) : true));
  const sorted = [...channels];
  if (channelsSort.value === 'name-asc') {
    sorted.sort((a, b) => a.username.localeCompare(b.username));
  } else if (channelsSort.value === 'name-desc') {
    sorted.sort((a, b) => b.username.localeCompare(a.username));
  } else if (channelsSort.value === 'id-desc') {
    sorted.sort((a, b) => b.id - a.id);
  } else {
    sorted.sort((a, b) => a.id - b.id);
  }
  return sorted;
});

const ircJoinedCount = computed(() => {
  const st = ircStatus.value;
  if (!st?.connected || !st.channels?.length) {
    return 0;
  }
  return st.channels.filter((c) => isChannelJoinedOnIrc(st, c.login)).length;
});

function channelIrcJoined(login: string): boolean {
  return isChannelJoinedOnIrc(ircStatus.value, login);
}
const rulesHeading = computed(() => `Rules (${rulesTotal.value})`);
const twitchHeading = computed(() => `Twitch accounts (${twitchAccountsTotal.value})`);
const filteredChannelBlacklist = computed(() => {
  const q = blacklistFilter.value.trim().toLowerCase();
  const filtered = channelBlacklist.value.filter((login) => (q ? login.toLowerCase().includes(q) : true));
  const sorted = [...filtered];
  if (blacklistSort.value === 'name-desc') {
    sorted.sort((a, b) => b.localeCompare(a));
  } else {
    sorted.sort((a, b) => a.localeCompare(b));
  }
  return sorted;
});
</script>

<template>
  <div class="settings-wrap">
    <div class="settings">
      <h1>Settings</h1>

      <nav class="tabs" aria-label="Settings sections">
        <button type="button" :class="{ active: tab === 'channels' }" @click="setTab('channels')">Channels</button>
        <button type="button" :class="{ active: tab === 'rules' }" @click="setTab('rules')">{{ rulesHeading }}</button>
        <button type="button" :class="{ active: tab === 'notifications' }" @click="setTab('notifications')">
          Notifications
        </button>
        <button type="button" :class="{ active: tab === 'twitch' }" @click="setTab('twitch')">{{ twitchHeading }}</button>
        <button type="button" :class="{ active: tab === 'channelDiscovery' }" @click="setTab('channelDiscovery')">
          Channel discovery
        </button>
        <button type="button" :class="{ active: tab === 'suspicion' }" @click="setTab('suspicion')">Suspicion</button>
        <button type="button" :class="{ active: tab === 'ai' }" @click="setTab('ai')">AI</button>
        <button type="button" :class="{ active: tab === 'display' }" @click="setTab('display')">Display</button>
      </nav>

      <section v-show="tab === 'channels'" class="panel">
        <h2>{{ channelsHeading }}</h2>
        <p class="hint">
          Monitored: {{ monitoredCount }} ·
          <RouterLink class="irc-joined-link" :to="{ name: 'irc-joined-graph' }">
            IRC joined: {{ ircJoinedCount }} / {{ maxIrcJoinedChannels }}
          </RouterLink>
          <span v-if="!ircStatus?.connected" class="muted"> (IRC not connected)</span>
        </p>
        <p v-if="ircJoinedCount > maxIrcJoinedChannels" class="hint hint-warn">
          IRC joined count is above {{ maxIrcJoinedChannels }}. Twitch limits concurrent joins per connection; some
          channels may not receive chat reliably until you reduce joins (e.g. live-only monitoring per channel on the User
          page).
        </p>
        <p v-else class="hint channels-limit-hint">
          Twitch IRC enforces a hard limit of about {{ maxIrcJoinedChannels }} concurrent channel joins per connection. The
          monitored list can be longer; the joined count is what matters for IRC.
        </p>
        <form class="row" @submit.prevent="addChannel">
          <input v-model="newChannel.name" placeholder="channel name" required />
          <SubmitButton :loading="addingChannel">Add channel</SubmitButton>
        </form>
        <div class="row channels-controls">
          <input v-model="channelsFilter" type="text" placeholder="Filter channels" autocomplete="off" />
          <select v-model="channelsSort" aria-label="Sort monitored channels">
            <option value="name-asc">Sort: name A-Z</option>
            <option value="name-desc">Sort: name Z-A</option>
            <option value="id-desc">Sort: newest first</option>
            <option value="id-asc">Sort: oldest first</option>
          </select>
        </div>
        <ul class="chan-list">
          <li v-for="c in filteredMonitoredChannels" :key="c.id">
            <span
              class="irc-dot"
              :class="channelIrcJoined(c.username) ? 'irc-dot--on' : 'irc-dot--off'"
              :title="channelIrcJoined(c.username) ? 'Joined on IRC' : 'Not joined on IRC'"
            />
            <RouterLink class="chan-link" :to="{ name: 'user', params: { id: String(c.id) } }">{{ c.username }}</RouterLink>
          </li>
        </ul>
      </section>

      <section v-show="tab === 'rules'" class="panel">
        <h2>{{ rulesHeading }}</h2>
        <ul class="rule-list">
          <li v-for="r in rules" :key="r.id" class="rule-row">
            <button type="button" class="btn-danger btn-inline-x" @click="deleteRule(r.id)">x</button>
            <div class="rule-row-main">
              <RouterLink class="rule-edit-link" :to="{ name: 'rule-edit', params: { id: String(r.id) } }">
                <span class="muted small rule-id">#{{ r.id }}</span>
                <span class="rule-name">{{ r.name }}</span>
              </RouterLink>
            </div>
          </li>
        </ul>
        <p class="row-actions">
          <RouterLink class="btn-secondary btn-link" :to="{ name: 'rule-new' }">Add rule</RouterLink>
          <button type="button" class="btn-secondary" @click="openRegexTestModal">Test regex</button>
        </p>
      </section>

      <section v-show="tab === 'notifications'" class="panel">
        <h2>Notifications</h2>
        <ul class="notif-list">
          <li v-for="n in notifications" :key="n.id" class="notif-row">
            <span class="tag">{{ n.provider }}</span>
            <span :class="{ muted: !n.enabled }">{{ n.enabled ? 'on' : 'off' }}</span>
            <button type="button" class="btn-secondary tiny" @click="toggleNotifEnabled(n)">Toggle</button>
            <button type="button" class="btn-danger tiny" @click="deleteNotif(n.id)">Remove</button>
          </li>
        </ul>
        <button type="button" class="btn-secondary" @click="openNotifModal">Add notification</button>
      </section>

      <section v-show="tab === 'twitch'" class="panel">
        <h2>{{ twitchHeading }}</h2>
        <p class="hint">Use “Connect with Twitch” to authorize in the browser and save a refresh token.</p>
        <p class="row-actions">
          <button type="button" class="btn-secondary" @click="connectTwitchInBrowser">Connect with Twitch</button>
        </p>

        <h3 class="subh">IRC monitor</h3>
        <p class="hint muted small">
          Default is anonymous read-only IRC. Pick a linked account only if you want the monitor to connect with that
          account's OAuth token.
        </p>
        <form class="stack gap-setting" @submit.prevent="saveIrcMonitorSettings">
          <label class="stack gap-setting">
            <span>Identity for chat ingestion</span>
            <select v-model="ircOAuthAccountId" :disabled="savingIrcMonitor">
              <option value="">Anonymous (read-only)</option>
              <option v-for="a in twitchStore.accounts" :key="a.id" :value="String(a.id)">
                {{ a.username }} ({{ a.account_type }})
              </option>
            </select>
          </label>
          <label class="stack gap-setting">
            <span>Enrichment cooldown (hours)</span>
            <input v-model.number="enrichmentCooldownHours" type="number" min="1" max="168" required />
          </label>
          <p class="row-actions">
            <SubmitButton :loading="savingIrcMonitor">Save IRC settings</SubmitButton>
          </p>
        </form>

        <ul class="twitch-account-list">
          <li v-for="a in twitchStore.accounts" :key="a.id" class="twitch-account-row">
            <div class="twitch-account-meta">
              <span class="twitch-account-name">{{ a.username }}</span>
              <span class="twitch-account-kind">{{ a.account_type }}</span>
            </div>
            <div class="row-actions wrap">
              <button
                v-if="a.account_type !== TwitchAccount.account_type.BOT"
                type="button"
                class="btn-secondary"
                @click="setTwitchAccountType(a.id, UpdateTwitchAccountPostRequest.account_type.BOT)"
              >
                Mark bot
              </button>
              <button
                v-if="a.account_type !== TwitchAccount.account_type.MAIN"
                type="button"
                class="btn-secondary"
                @click="setTwitchAccountType(a.id, UpdateTwitchAccountPostRequest.account_type.MAIN)"
              >
                Mark main
              </button>
              <button type="button" class="btn-danger" @click="removeTwitchAccount(a.id)">Remove</button>
            </div>
          </li>
        </ul>
      </section>

      <section v-show="tab === 'channelDiscovery'" class="panel">
        <h2>Channel discovery</h2>
        <p class="hint">
          Periodically scans Twitch live streams (Helix) for a game. Matching channels appear below until you approve
          (monitor) or deny (never suggest again). Required tags use AND semantics (every listed tag must appear on the
          stream). Minimum viewers uses Helix concurrent viewer count.
        </p>
        <form class="stack gap-setting" @submit.prevent="saveChannelDiscoverySettings">
          <label class="row-inline">
            <input v-model="discoveryDraft.enabled" type="checkbox" />
            <span>Enable discovery</span>
          </label>
          <label class="stack gap-setting">
            <span>Poll interval (seconds)</span>
            <input v-model.number="discoveryDraft.poll_interval_seconds" type="number" min="60" max="86400" required />
          </label>
          <label class="stack gap-setting">
            <span>Twitch game id (Helix category id)</span>
            <input v-model="discoveryDraft.game_id" type="text" placeholder="e.g. 33214" autocomplete="off" />
          </label>
          <label class="stack gap-setting">
            <span>Minimum live viewers</span>
            <input v-model.number="discoveryDraft.min_live_viewers" type="number" min="0" max="2000000" required />
          </label>
          <label class="stack gap-setting">
            <span>Max stream pages per run (100 streams each)</span>
            <input v-model.number="discoveryDraft.max_stream_pages_per_run" type="number" min="1" max="500" required />
          </label>
          <label class="stack gap-setting">
            <span>Required stream tags (one per line; leave empty for no tag filter)</span>
            <textarea v-model="discoveryTagsText" rows="5" placeholder="One Twitch stream tag per line" />
          </label>
          <p class="row-actions">
            <SubmitButton :loading="savingDiscovery">Save discovery settings</SubmitButton>
          </p>
        </form>

        <h3 class="subh">Pending channels</h3>
        <p v-if="!discoveryCandidates.length" class="muted small">No pending suggestions.</p>
        <table v-else class="discovery-table">
          <thead>
            <tr>
              <th>Channel</th>
              <th>Viewers</th>
              <th>Tags</th>
              <th>Title</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in discoveryCandidates" :key="c.user.id">
              <td>
                <RouterLink class="chan-link" :to="{ name: 'user', params: { id: String(c.user.id) } }">
                  {{ c.user.username }}
                </RouterLink>
              </td>
              <td>{{ c.viewer_count ?? '—' }}</td>
              <td class="discovery-tags">{{ (c.stream_tags || []).join(', ') }}</td>
              <td class="discovery-title">{{ c.title || '—' }}</td>
              <td class="row-actions wrap">
                <button type="button" class="btn-secondary" @click="approveDiscoveryUser(c.user.id)">Approve</button>
                <button type="button" class="btn-danger" @click="denyDiscoveryUser(c.user.id)">Deny</button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <section v-show="tab === 'suspicion'" class="panel">
        <h2>Suspicion</h2>
        <p class="hint">Blacklist applies to outgoing follows enrichment; automatic rules run after GQL sync.</p>

        <h3 class="subh">Channel blacklist</h3>
        <form class="row bl-add" @submit.prevent="addBlacklistEntry">
          <input v-model="newBlacklistLogin" placeholder="channel login" autocomplete="off" />
          <SubmitButton :loading="savingBlacklist">Add</SubmitButton>
        </form>
        <div class="row bl-controls">
          <input v-model="blacklistFilter" type="text" placeholder="Filter blacklisted channels" autocomplete="off" />
          <select v-model="blacklistSort" aria-label="Sort blacklisted channels">
            <option value="name-asc">Sort: name A-Z</option>
            <option value="name-desc">Sort: name Z-A</option>
          </select>
        </div>
        <p v-if="!channelBlacklist.length" class="muted small">No channels blacklisted.</p>
        <p v-else-if="!filteredChannelBlacklist.length" class="muted small">No channels match current filter.</p>
        <ul v-else class="bl-list">
          <li v-for="login in filteredChannelBlacklist" :key="login" class="bl-row">
            <button
              type="button"
              class="btn-danger btn-inline-x"
              :disabled="savingBlacklist"
              @click="removeBlacklistEntry(login)"
            >
              x
            </button>
            <code>{{ login }}</code>
          </li>
        </ul>

        <h3 class="subh">Automatic rules</h3>
        <form class="stack sus-form" @submit.prevent="saveSuspicionSettings">
          <label class="row-inline">
            <input v-model="suspicionDraft.auto_check_account_age" type="checkbox" />
            <span>Flag accounts newer than N days</span>
          </label>
          <label class="stack gap-setting">
            <span>Account age threshold (days)</span>
            <input v-model.number="suspicionDraft.account_age_sus_days" type="number" min="0" max="36500" required />
          </label>
          <label class="row-inline">
            <input v-model="suspicionDraft.auto_check_blacklist" type="checkbox" />
            <span>Flag if user follows any blacklisted channel</span>
          </label>
          <label class="row-inline">
            <input v-model="suspicionDraft.auto_check_low_follows" type="checkbox" />
            <span>Flag low follow count</span>
          </label>
          <label class="stack gap-setting">
            <span>Low follows threshold (suspicious when total follows &lt; this)</span>
            <input v-model.number="suspicionDraft.low_follows_threshold" type="number" min="0" max="1000000" required />
          </label>
          <label class="stack gap-setting">
            <span>Max GQL follow pages per user (safety cap)</span>
            <input v-model.number="suspicionDraft.max_gql_follow_pages" type="number" min="1" max="500" required />
          </label>
          <p class="row-actions">
            <SubmitButton :loading="savingSuspicion">Save thresholds</SubmitButton>
          </p>
        </form>
      </section>

      <section v-show="tab === 'ai'" class="panel">
        <h2>AI assistant</h2>
        <p class="hint">
          OpenAI-compatible API (e.g. <code>https://api.openai.com/v1</code>). Token is stored encrypted on the server.
        </p>
        <form class="stack gap-setting" @submit.prevent="saveAiSettings">
          <label class="stack gap-setting">
            <span>Base URL</span>
            <input v-model="aiBaseUrl" type="url" autocomplete="off" placeholder="https://api.openai.com/v1" />
          </label>
          <label class="stack gap-setting">
            <span>Model</span>
            <input v-model="aiModel" type="text" autocomplete="off" placeholder="gpt-4o-mini" />
          </label>
          <label class="stack gap-setting">
            <span>API token</span>
            <input v-model="aiTokenDraft" type="password" autocomplete="off" placeholder="Leave blank to keep existing" />
            <span v-if="aiHasToken" class="muted small">Stored token ends with …{{ aiTokenLast4 }}</span>
          </label>
          <p class="row-actions">
            <SubmitButton :loading="savingAi">Save AI settings</SubmitButton>
          </p>
        </form>
      </section>

      <section v-show="tab === 'display'" class="panel">
        <h2>Watch / chat</h2>
        <label class="stack gap-setting">
          <span>Chat gap delimiter (minutes)</span>
          <input
            v-model.number="chatGapMinutes"
            type="number"
            min="1"
            max="1440"
            @change="watchPrefs.setChatGapMinutes(chatGapMinutes)"
            @blur="watchPrefs.setChatGapMinutes(chatGapMinutes)"
          />
        </label>
      </section>
    </div>

    <AppModal :open="regexTestModalOpen" title="Test regex" @close="regexTestModalOpen = false">
      <div class="stack modal-form">
        <label>
          <span>Pattern</span>
          <input v-model="regexTestPattern" autocomplete="off" placeholder="regular expression" />
        </label>
        <label>
          <span>Sample text</span>
          <textarea v-model="regexTestSample" rows="3" autocomplete="off" placeholder="string to match" />
        </label>
        <label class="row-inline">
          <input v-model="regexTestCaseInsensitive" type="checkbox" />
          <span>Case insensitive (match_regex)</span>
        </label>
        <p v-if="regexTestCompileError" class="regex-test-err">{{ regexTestCompileError }}</p>
        <p v-else-if="regexTestMatches !== null" class="regex-test-ok">
          {{ regexTestMatches ? 'Matches.' : 'Does not match.' }}
        </p>
        <p v-else class="muted small">Enter a pattern and sample…</p>
        <footer class="modal-actions">
          <button type="button" class="btn-secondary" @click="regexTestModalOpen = false">Close</button>
        </footer>
      </div>
    </AppModal>

    <AppModal :open="notifModalOpen" title="New notification" @close="notifModalOpen = false">
      <form class="stack modal-form" @submit.prevent="saveNotifModal">
        <label>
          <span>Provider</span>
          <select v-model="notifDraft.provider">
            <option value="telegram">telegram</option>
            <option value="webhook">webhook</option>
          </select>
        </label>
        <label class="row-inline">
          <input v-model="notifDraft.enabled" type="checkbox" />
          <span>Enabled</span>
        </label>
        <template v-if="notifDraft.provider === 'telegram'">
          <label>
            <span>Bot token</span>
            <input v-model="notifDraft.bot_token" type="password" autocomplete="off" />
          </label>
          <label>
            <span>Chat id</span>
            <input v-model="notifDraft.chat_id" autocomplete="off" />
          </label>
        </template>
        <template v-else>
          <label>
            <span>Webhook URL</span>
            <input v-model="notifDraft.webhook_url" type="url" name="dredge_notif_webhook_url" autocomplete="off" />
          </label>
        </template>
        <footer class="modal-actions">
          <button type="button" class="btn-secondary" @click="notifModalOpen = false">Cancel</button>
          <SubmitButton :loading="savingNotif">Save</SubmitButton>
        </footer>
      </form>
    </AppModal>
  </div>
</template>

<style scoped lang="scss">
.settings-wrap {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 1rem 1rem 2rem;
  min-height: 0;
  overflow-y: auto;
}

.settings {
  width: 100%;
  max-width: 640px;
}

h1 {
  font-size: 1.25rem;
  margin: 0 0 0.75rem;
  text-align: center;
}

.tabs {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.4rem;
  margin: 0 0 1rem;

  button {
    padding: 0.4rem 0.7rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    color: var(--text-muted);
    font-size: 0.82rem;
    cursor: pointer;

    &:hover {
      color: var(--text);
      background: var(--bg-hover);
    }

    &.active {
      color: var(--text);
      border-color: var(--accent);
      background: rgba(145, 71, 255, 0.12);
    }
  }
}

.panel {
  padding-bottom: 0.5rem;
}

h2 {
  font-size: 1rem;
  margin: 0 0 0.5rem;
  color: var(--accent-bright);
}

.hint {
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.45;
  margin: 0 0 0.65rem;

  a {
    color: var(--accent-bright);
  }
}

.irc-joined-link {
  color: var(--accent-bright);
  font-weight: 600;
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }
}

.channels-limit-hint {
  color: rgba(255, 193, 7, 0.92);
}

.muted {
  color: var(--text-muted);
}

.small {
  font-size: 0.78rem;
}

.row-actions {
  margin: 0 0 0.75rem;
}

.btn-secondary {
  padding: 0.45rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
    border-color: var(--accent);
  }

  &.tiny {
    padding: 0.15rem 0.45rem;
    font-size: 0.75rem;
  }
}

section.panel {
  border-bottom: 1px solid var(--border);
  padding-bottom: 1rem;
  margin-bottom: 0.5rem;
}

ul {
  margin: 0 0 0.65rem;
  padding-left: 1.1rem;
  font-size: 0.9rem;
  color: var(--text-muted);

  &.chan-list li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    list-style: disc;
  }

  &.twitch-account-list {
    list-style: none;
    padding-left: 0;
  }
}

.chan-link {
  color: var(--accent-bright);
  font-weight: 600;
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }
}

.hint-warn {
  color: var(--accent-bright);
  border-left: 3px solid var(--accent);
  padding-left: 0.5rem;
}

.irc-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  flex-shrink: 0;

  &--on {
    background: #2ecc71;
  }

  &--off {
    background: #c0392b;
  }
}

.rule-list,
.notif-list {
  list-style: none;
  padding-left: 0;
}

.rule-row,
.notif-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.rule-row-main {
  flex: 1;
  min-width: 0;
}

.rule-edit-link {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.35rem;
  color: var(--accent-bright);
  text-decoration: none;
  font-weight: 600;

  &:hover {
    text-decoration: underline;
  }

  .rule-id {
    flex-shrink: 0;
  }

  .rule-name {
    font-weight: 600;
    min-width: 0;
  }
}

a.btn-link {
  display: inline-block;
  text-decoration: none;
  text-align: center;
}

.rule-meta {
  margin-top: 0.15rem;
}

.tag {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--accent-bright);
}

.twitch-account-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.subh {
  font-size: 0.92rem;
  margin: 1rem 0 0.5rem;
  color: var(--text);
  font-weight: 600;
}

.bl-list {
  list-style: none;
  padding-left: 0;
  margin: 0 0 0.65rem;
}

.bl-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
}

.bl-add {
  margin-bottom: 0.5rem;
  max-width: 28rem;
}

.bl-controls {
  margin-bottom: 0.65rem;
}

.sus-form {
  max-width: 28rem;
}

.twitch-account-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.35rem;
}

.twitch-account-name {
  color: var(--text);
  font-weight: 600;
}

.twitch-account-kind {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}

.row-actions.wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
}

.btn-danger {
  padding: 0.15rem 0.45rem;
  font-size: 0.75rem;
  border-radius: 0.2rem;
  border: 1px solid var(--border);
  background: transparent;
  color: #ff6b6b;
  cursor: pointer;

  &:hover {
    background: rgba(255, 107, 107, 0.12);
  }

  &.tiny {
    padding: 0.1rem 0.35rem;
    font-size: 0.7rem;
  }
}

.btn-inline-x {
  width: 1.15rem;
  height: 1.15rem;
  min-width: 1.15rem;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.68rem;
  line-height: 1;
}

code {
  font-size: 0.85rem;
}

.row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;

  input.grow {
    flex: 1 1 180px;
  }

  input[type='text'],
  input:not([type]),
  select {
    padding: 0.4rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    color: var(--text);
  }

  button {
    padding: 0.4rem 0.65rem;
    border-radius: 0.25rem;
    border: none;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
    cursor: pointer;

    &:disabled {
      opacity: 0.65;
      cursor: not-allowed;
    }
  }
}

.channels-controls {
  margin-bottom: 0.65rem;
}

.gap-setting {
  max-width: 280px;
}

.stack {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  max-width: 420px;

  input,
  select {
    padding: 0.45rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    color: var(--text);
  }

  button {
    padding: 0.45rem;
    border-radius: 0.25rem;
    border: none;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
    cursor: pointer;
    align-self: flex-start;

    &:disabled {
      opacity: 0.65;
      cursor: not-allowed;
    }
  }
}

.modal-form {
  max-width: none;
}

.modal-form label span {
  display: block;
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-bottom: 0.2rem;
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 0.5rem;
}

.row-inline {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-direction: row;
}

.discovery-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8rem;
  margin-top: 0.5rem;

  th,
  td {
    border-bottom: 1px solid var(--border);
    padding: 0.35rem 0.4rem;
    text-align: left;
    vertical-align: top;
  }

  th {
    color: var(--text-muted);
    font-weight: 600;
  }
}

.discovery-tags {
  max-width: 12rem;
  word-break: break-word;
}

.discovery-title {
  max-width: 14rem;
  word-break: break-word;
}
</style>
