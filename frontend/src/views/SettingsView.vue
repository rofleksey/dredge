<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  ApiError,
  CreateNotificationRequest,
  DefaultService,
  NotificationEntry,
  TwitchAccount,
  UpdateTwitchAccountPostRequest,
} from '../api/generated';
import type { IrcMonitorStatus } from '../api/generated';
import type { SuspicionSettings } from '../api/generated';
import type { Rule } from '../api/generated';
import AppModal from '../components/AppModal.vue';
import SubmitButton from '../components/SubmitButton.vue';
import { notify } from '../lib/notify';
import { useChannelsStore } from '../stores/channels';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';
import { useWatchPreferencesStore } from '../stores/watchPreferences';

type SettingsTab = 'channels' | 'rules' | 'notifications' | 'twitch' | 'suspicion' | 'display';

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
const suspicionDraft = reactive<SuspicionSettings>({
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
  provider: 'telegram' as 'telegram' | 'webhook',
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

watch(
  () => tab.value,
  (t) => {
    if (t === 'channels') {
      startIrcPoll();
    } else {
      stopIrcPoll();
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  stopIrcPoll();
});

async function refresh(): Promise<void> {
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
    t === 'suspicion' ||
    t === 'display'
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

async function setChannelMonitored(id: number, monitored: boolean): Promise<void> {
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
  } catch {
    notify({
      id: 'settings-channel-monitored',
      type: 'error',
      title: 'Channels',
      description: 'Could not update channel.',
    });
  }
}

function openRuleModal(): void {
  ruleDraft.regex = '';
  ruleDraft.included_users = '*';
  ruleDraft.denied_users = '';
  ruleDraft.included_channels = '*';
  ruleDraft.denied_channels = '';
  ruleModalOpen.value = true;
}

async function saveRuleModal(): Promise<void> {
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
  } catch {
    notify({
      id: 'settings-add-rule',
      type: 'error',
      title: 'Rules',
      description: 'Could not add rule (invalid regex?).',
    });
  } finally {
    savingRule.value = false;
  }
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

const ircJoinedCount = computed(() => {
  const st = ircStatus.value;
  if (!st?.connected || !st.channels?.length) {
    return 0;
  }
  return st.channels.filter((c) => c.irc_ok).length;
});

function channelIrcJoined(login: string): boolean {
  const st = ircStatus.value;
  if (!st?.connected) {
    return false;
  }
  const low = login.toLowerCase();
  const row = st.channels.find((c) => c.login.toLowerCase() === low);
  return row?.irc_ok ?? false;
}
const rulesHeading = computed(() => `Rules (${rulesTotal.value})`);
const twitchHeading = computed(() => `Twitch accounts (${twitchAccountsTotal.value})`);
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
        <button type="button" :class="{ active: tab === 'suspicion' }" @click="setTab('suspicion')">Suspicion</button>
        <button type="button" :class="{ active: tab === 'display' }" @click="setTab('display')">Display</button>
      </nav>

      <section v-show="tab === 'channels'" class="panel">
        <h2>{{ channelsHeading }}</h2>
        <p class="hint">
          Monitored: {{ monitoredCount }} · IRC joined:
          {{ ircJoinedCount }}
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
        <ul class="chan-list">
          <li v-for="c in monitoredChannels" :key="c.id">
            <span
              class="irc-dot"
              :class="channelIrcJoined(c.username) ? 'irc-dot--on' : 'irc-dot--off'"
              :title="channelIrcJoined(c.username) ? 'Joined on IRC' : 'Not joined on IRC'"
            />
            <span>{{ c.username }}</span>
            <button type="button" class="btn-danger" @click="setChannelMonitored(c.id, false)">Stop monitoring</button>
          </li>
        </ul>
        <form class="row" @submit.prevent="addChannel">
          <input v-model="newChannel.name" placeholder="channel name" required />
          <SubmitButton :loading="addingChannel">Add channel</SubmitButton>
        </form>
      </section>

      <section v-show="tab === 'rules'" class="panel">
        <h2>{{ rulesHeading }}</h2>
        <ul class="rule-list">
          <li v-for="r in rules" :key="r.id" class="rule-row">
            <div>
              <code>{{ r.regex }}</code>
              <div class="rule-meta muted small">
                users: {{ r.included_users }} / {{ r.denied_users }} · ch: {{ r.included_channels }} / {{ r.denied_channels }}
              </div>
            </div>
            <button type="button" class="btn-danger" @click="deleteRule(r.id)">Delete</button>
          </li>
        </ul>
        <p class="row-actions">
          <button type="button" class="btn-secondary" @click="openRuleModal">Add rule</button>
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

      <section v-show="tab === 'suspicion'" class="panel">
        <h2>Suspicion</h2>
        <p class="hint">Blacklist applies to outgoing follows enrichment; automatic rules run after GQL sync.</p>

        <h3 class="subh">Channel blacklist</h3>
        <p v-if="!channelBlacklist.length" class="muted small">No channels blacklisted.</p>
        <ul v-else class="bl-list">
          <li v-for="login in channelBlacklist" :key="login" class="bl-row">
            <code>{{ login }}</code>
            <button
              type="button"
              class="btn-danger tiny"
              :disabled="savingBlacklist"
              @click="removeBlacklistEntry(login)"
            >
              Remove
            </button>
          </li>
        </ul>
        <form class="row bl-add" @submit.prevent="addBlacklistEntry">
          <input v-model="newBlacklistLogin" placeholder="channel login" autocomplete="off" />
          <SubmitButton :loading="savingBlacklist">Add</SubmitButton>
        </form>

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

    <AppModal :open="ruleModalOpen" title="New rule" @close="ruleModalOpen = false">
      <form class="stack modal-form" autocomplete="off" @submit.prevent="saveRuleModal">
        <label>
          <span>Regex</span>
          <input v-model="ruleDraft.regex" name="dredge_rule_regex" required placeholder="pattern" autocomplete="off" />
        </label>
        <label>
          <span>Included users (* = all)</span>
          <input v-model="ruleDraft.included_users" name="dredge_rule_included_users" autocomplete="off" />
        </label>
        <label>
          <span>Denied users (empty = none)</span>
          <input v-model="ruleDraft.denied_users" name="dredge_rule_denied_users" autocomplete="off" />
        </label>
        <label>
          <span>Included channels (* = all)</span>
          <input v-model="ruleDraft.included_channels" name="dredge_rule_included_channels" autocomplete="off" />
        </label>
        <label>
          <span>Denied channels (empty = none)</span>
          <input v-model="ruleDraft.denied_channels" name="dredge_rule_denied_channels" autocomplete="off" />
        </label>
        <footer class="modal-actions">
          <button type="button" class="btn-secondary" @click="ruleModalOpen = false">Cancel</button>
          <SubmitButton :loading="savingRule">Save</SubmitButton>
        </footer>
      </form>
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
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
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
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
}

.bl-add {
  margin-bottom: 0.5rem;
  max-width: 28rem;
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

  input[type='text'] {
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
</style>
