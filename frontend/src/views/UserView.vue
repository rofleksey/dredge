<script setup lang="ts">
import * as echarts from 'echarts';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import { ApiError, ChatHistoryEntry, DefaultService } from '../api/generated';
import type {
  ActivityTimelineSegment,
  FollowedChannelEntry,
  TwitchUser,
  TwitchUserProfile,
  UpdateTwitchUserPostRequest,
} from '../api/generated';
import type { UserActivityEvent } from '../api/generated/models/UserActivityEvent';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { notify } from '../lib/notify';

defineOptions({ name: 'UserView' });

type UserTab = 'overview' | 'following' | 'messages' | 'activity' | 'graphs' | 'settings';

const route = useRoute();
const profile = ref<TwitchUserProfile | null>(null);
const notFound = ref(false);
const messages = ref<ChatHistoryEntry[]>([]);
const loadingProfile = ref(true);
const loadingMessages = ref(true);
const loadingMore = ref(false);
const togglingMarked = ref(false);
const savingUserSettings = ref(false);
const togglingSus = ref(false);
const manualSusNote = ref('');
const followingQuery = ref('');

const userTab = ref<UserTab>('overview');

const activity = ref<UserActivityEvent[]>([]);
const loadingActivity = ref(false);
const loadingActivityMore = ref(false);

const timelineSegments = ref<ActivityTimelineSegment[]>([]);
const loadingTimeline = ref(false);
const chartEl = ref<HTMLDivElement | null>(null);
let chart: echarts.ECharts | null = null;

const userId = computed(() => {
  const raw = route.params.id;
  const s = Array.isArray(raw) ? raw[0] : raw;
  const n = Number(s);
  return Number.isFinite(n) ? n : NaN;
});

async function loadProfile(): Promise<void> {
  if (!Number.isFinite(userId.value)) {
    notFound.value = true;
    profile.value = null;
    loadingProfile.value = false;
    return;
  }

  loadingProfile.value = true;
  notFound.value = false;

  try {
    profile.value = await DefaultService.getTwitchUserProfile({ requestBody: { id: userId.value } });
  } catch (e) {
    profile.value = null;
    if (e instanceof ApiError && e.status === 404) {
      notFound.value = true;
    } else {
      notify({
        id: 'user-profile',
        type: 'error',
        title: 'User',
        description: 'Could not load profile.',
      });
    }
  } finally {
    loadingProfile.value = false;
  }
}

function mergeProfileFromTwitchUser(u: TwitchUser): void {
  if (!profile.value) {
    return;
  }
  profile.value = {
    ...profile.value,
    monitored: u.monitored,
    marked: u.marked,
    is_sus: u.is_sus,
    sus_type: u.sus_type,
    sus_description: u.sus_description,
    sus_auto_suppressed: u.sus_auto_suppressed,
    irc_only_when_live: u.irc_only_when_live,
    notify_off_stream_messages: u.notify_off_stream_messages,
    notify_stream_start: u.notify_stream_start,
  };
}

async function patchUserProfile(patch: Omit<UpdateTwitchUserPostRequest, 'id'>): Promise<void> {
  if (!profile.value || savingUserSettings.value) {
    return;
  }
  savingUserSettings.value = true;
  try {
    const body: UpdateTwitchUserPostRequest = { id: profile.value.id, ...patch };
    if (body.irc_only_when_live === false) {
      body.notify_off_stream_messages = false;
    }
    const u = await DefaultService.updateTwitchUser({
      requestBody: body,
    });
    mergeProfileFromTwitchUser(u);
  } catch (e) {
    const msg =
      e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
        ? String((e.body as { message: string }).message)
        : 'Could not save settings.';
    notify({
      id: 'user-settings',
      type: 'error',
      title: 'User',
      description: msg,
    });
  } finally {
    savingUserSettings.value = false;
  }
}

async function markSuspicious(): Promise<void> {
  if (!profile.value || togglingSus.value) {
    return;
  }
  togglingSus.value = true;
  const desc = manualSusNote.value.trim() || 'Marked manually';
  try {
    const u = await DefaultService.updateTwitchUser({
      requestBody: {
        id: profile.value.id,
        is_sus: true,
        sus_type: 'manual',
        sus_description: desc,
        sus_auto_suppressed: false,
      },
    });
    mergeProfileFromTwitchUser(u);
    manualSusNote.value = '';
  } catch {
    notify({
      id: 'user-sus-mark',
      type: 'error',
      title: 'User',
      description: 'Could not update suspicion.',
    });
  } finally {
    togglingSus.value = false;
  }
}

async function clearSuspicion(): Promise<void> {
  if (!profile.value || togglingSus.value) {
    return;
  }
  togglingSus.value = true;
  try {
    const u = await DefaultService.updateTwitchUser({
      requestBody: {
        id: profile.value.id,
        is_sus: false,
        sus_type: null,
        sus_description: null,
        sus_auto_suppressed: true,
      },
    });
    mergeProfileFromTwitchUser(u);
  } catch {
    notify({
      id: 'user-sus-clear',
      type: 'error',
      title: 'User',
      description: 'Could not clear suspicion.',
    });
  } finally {
    togglingSus.value = false;
  }
}

async function allowAutoSuspicionAgain(): Promise<void> {
  if (!profile.value || togglingSus.value) {
    return;
  }
  togglingSus.value = true;
  try {
    const u = await DefaultService.updateTwitchUser({
      requestBody: {
        id: profile.value.id,
        sus_auto_suppressed: false,
      },
    });
    mergeProfileFromTwitchUser(u);
  } catch {
    notify({
      id: 'user-sus-auto',
      type: 'error',
      title: 'User',
      description: 'Could not update suppression flag.',
    });
  } finally {
    togglingSus.value = false;
  }
}

async function toggleMarked(): Promise<void> {
  if (!profile.value || togglingMarked.value) {
    return;
  }
  togglingMarked.value = true;
  try {
    const u: TwitchUser = await DefaultService.updateTwitchUser({
      requestBody: { id: profile.value.id, marked: !profile.value.marked },
    });
    mergeProfileFromTwitchUser(u);
  } catch {
    notify({
      id: 'user-marked',
      type: 'error',
      title: 'User',
      description: 'Could not update marked flag.',
    });
  } finally {
    togglingMarked.value = false;
  }
}

function buildMsgQuery(appendCursor: boolean) {
  const q: Parameters<typeof DefaultService.listTwitchMessages>[0] = {
    limit: 80,
    chatterUserId: userId.value,
  };
  if (appendCursor && messages.value.length) {
    const last = messages.value[messages.value.length - 1];
    q.cursorCreatedAt = last.created_at;
    q.cursorId = last.id;
  }
  return q;
}

async function loadMessages(): Promise<void> {
  if (!Number.isFinite(userId.value)) {
    messages.value = [];
    loadingMessages.value = false;
    return;
  }

  loadingMessages.value = true;
  try {
    messages.value = await DefaultService.listTwitchMessages(buildMsgQuery(false));
  } catch {
    messages.value = [];
    notify({
      id: 'user-msgs',
      type: 'error',
      title: 'User',
      description: 'Could not load messages.',
    });
  } finally {
    loadingMessages.value = false;
  }
}

async function loadMore(): Promise<void> {
  if (!messages.value.length || loadingMore.value || !Number.isFinite(userId.value)) {
    return;
  }
  loadingMore.value = true;
  try {
    const next = await DefaultService.listTwitchMessages(buildMsgQuery(true));
    const seen = new Set(messages.value.map((m) => m.id));
    for (const m of next) {
      if (!seen.has(m.id)) {
        messages.value.push(m);
        seen.add(m.id);
      }
    }
  } catch {
    notify({
      id: 'user-msgs-more',
      type: 'error',
      title: 'User',
      description: 'Could not load more messages.',
    });
  } finally {
    loadingMore.value = false;
  }
}

function buildActivityQuery(appendCursor: boolean) {
  const q = {
    id: userId.value,
    limit: 50,
    cursor_created_at: undefined as string | undefined,
    cursor_id: undefined as number | undefined,
  };
  if (appendCursor && activity.value.length) {
    const last = activity.value[activity.value.length - 1];
    q.cursor_created_at = last.created_at;
    q.cursor_id = last.id;
  }
  return q;
}

async function loadActivityFirst(): Promise<void> {
  if (!Number.isFinite(userId.value)) {
    activity.value = [];
    return;
  }
  loadingActivity.value = true;
  try {
    activity.value = await DefaultService.listTwitchUserActivity({
      requestBody: buildActivityQuery(false),
    });
  } catch {
    activity.value = [];
    notify({
      id: 'user-activity',
      type: 'error',
      title: 'User',
      description: 'Could not load activity.',
    });
  } finally {
    loadingActivity.value = false;
  }
}

async function loadActivityMore(): Promise<void> {
  if (!activity.value.length || loadingActivityMore.value || !Number.isFinite(userId.value)) {
    return;
  }
  loadingActivityMore.value = true;
  try {
    const next = await DefaultService.listTwitchUserActivity({
      requestBody: buildActivityQuery(true),
    });
    const seen = new Set(activity.value.map((e) => e.id));
    for (const e of next) {
      if (!seen.has(e.id)) {
        activity.value.push(e);
        seen.add(e.id);
      }
    }
  } catch {
    notify({
      id: 'user-activity-more',
      type: 'error',
      title: 'User',
      description: 'Could not load more activity.',
    });
  } finally {
    loadingActivityMore.value = false;
  }
}

async function loadTimeline(): Promise<void> {
  if (!Number.isFinite(userId.value)) {
    timelineSegments.value = [];
    return;
  }
  loadingTimeline.value = true;
  try {
    const to = new Date();
    const from = new Date(to.getTime() - 7 * 24 * 60 * 60 * 1000);
    timelineSegments.value = await DefaultService.getTwitchUserActivityTimeline({
      requestBody: {
        id: userId.value,
        from: from.toISOString(),
        to: to.toISOString(),
      },
    });
  } catch {
    timelineSegments.value = [];
    notify({
      id: 'user-timeline',
      type: 'error',
      title: 'User',
      description: 'Could not load activity timeline.',
    });
  } finally {
    loadingTimeline.value = false;
  }
}

function renderTimelineChart(): void {
  if (!chartEl.value) {
    return;
  }

  if (chart) {
    chart.dispose();
    chart = null;
  }

  const segs = timelineSegments.value;
  if (!segs.length) {
    return;
  }

  const channels = [...new Set(segs.map((s) => s.channel_login))].sort((a, b) => a.localeCompare(b));
  const data: [number, number, number][] = segs.map((s) => {
    const yi = channels.indexOf(s.channel_login);
    const t0 = new Date(s.start).getTime();
    const t1 = new Date(s.end).getTime();
    return [yi, t0, t1];
  });

  chart = echarts.init(chartEl.value, undefined, { renderer: 'canvas' });

  chart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: (p: unknown) => {
        const d = p as { value?: [number, number, number]; name?: string };
        const v = d.value;
        if (!v || v.length < 3) {
          return '';
        }
        const ch = channels[v[0]] ?? '';
        const a = new Date(v[1]).toLocaleString();
        const b = new Date(v[2]).toLocaleString();
        return `${ch}<br/>${a} — ${b}`;
      },
    },
    grid: { left: 140, right: 24, top: 16, bottom: 48 },
    xAxis: {
      type: 'time',
      scale: true,
      axisLabel: { fontSize: 10, color: '#aaa' },
    },
    yAxis: {
      type: 'category',
      data: channels,
      axisLabel: { fontSize: 11, color: '#ccc' },
      inverse: true,
    },
    dataZoom: [{ type: 'inside' }, { type: 'slider', height: 18 }],
    series: [
      {
        type: 'custom',
        renderItem(_params: unknown, api: any) {
          const yIndex = api.value(0) as number;
          const t0 = api.value(1) as number;
          const t1 = api.value(2) as number;
          const start = api.coord([t0, yIndex]);
          const end = api.coord([t1, yIndex]);
          const sy = api.size([0, 1])[1] ?? 12;
          const h = Math.max(6, sy * 0.55);
          return {
            type: 'rect',
            shape: {
              x: start[0],
              y: start[1] - h / 2,
              width: Math.max(1, end[0] - start[0]),
              height: h,
            },
            style: {
              fill: 'rgba(145, 71, 255, 0.45)',
              stroke: 'rgba(145, 71, 255, 0.85)',
              lineWidth: 1,
            },
          };
        },
        dimensions: ['y', 't0', 't1'],
        encode: { x: [1, 2], y: 0 },
        data,
      },
    ],
  });
}

function onResize(): void {
  chart?.resize();
}

watch(userTab, async (t) => {
  if (t === 'activity' && !loadingActivity.value && activity.value.length === 0 && Number.isFinite(userId.value)) {
    await loadActivityFirst();
  }
  if (t === 'graphs') {
    await loadTimeline();
    await nextTick();
    renderTimelineChart();
  }
});

watch(
  () => route.params.id,
  async () => {
    userTab.value = 'overview';
    followingQuery.value = '';
    activity.value = [];
    timelineSegments.value = [];
    if (chart) {
      chart.dispose();
      chart = null;
    }
    await loadProfile();
    await loadMessages();
  },
);

onMounted(async () => {
  window.addEventListener('resize', onResize);
  await loadProfile();
  await loadMessages();
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize);
  if (chart) {
    chart.dispose();
    chart = null;
  }
});

function rowBadges(m: ChatHistoryEntry): ChatBadgeTag[] {
  return [...(m.badge_tags ?? [])] as ChatBadgeTag[];
}

function activityLabel(e: UserActivityEvent): string {
  const ch = e.channel ? `#${e.channel}` : '';
  switch (e.event_type) {
    case 'chat_online':
      return ch ? `Online in ${ch}` : 'Online in chat';
    case 'chat_offline':
      return ch ? `Offline in ${ch}` : 'Offline in chat';
    default:
      return e.event_type;
  }
}

function formatWhen(iso: string): string {
  const t = Date.parse(iso);
  if (!Number.isFinite(t)) {
    return iso;
  }
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(t);
}

const accountCreatedLabel = computed(() => {
  const raw = profile.value?.account_created_at;
  if (!raw) {
    return null;
  }
  return formatWhen(raw);
});

const profileAvatarUrl = computed((): string => {
  const u = profile.value?.profile_image_url;
  if (u == null || u === '') {
    return '';
  }
  return u;
});

const sortedFollowedChannels = computed((): FollowedChannelEntry[] => {
  const p = profile.value;
  if (!p?.followed_channels?.length) {
    return [];
  }
  const q = followingQuery.value.trim().toLowerCase();
  let rows = [...p.followed_channels];
  if (q) {
    rows = rows.filter((r) => r.channel_login.toLowerCase().includes(q));
  }
  rows.sort((a, b) => {
    if (a.on_blacklist !== b.on_blacklist) {
      return a.on_blacklist ? -1 : 1;
    }
    return a.channel_login.localeCompare(b.channel_login);
  });
  return rows;
});

function formatPresenceWeek(sec: number): string {
  if (!Number.isFinite(sec) || sec <= 0) {
    return '—';
  }
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  if (h > 0) {
    return `${h}h ${m}m`;
  }
  if (m > 0) {
    return `${m}m ${s}s`;
  }
  return `${s}s`;
}
</script>

<template>
  <div class="page user-page">
    <template v-if="loadingProfile">
      <p class="muted">Loading…</p>
    </template>
    <template v-else-if="notFound">
      <p class="muted">User not found.</p>
    </template>
    <template v-else-if="profile">
      <header class="page-head">
        <div class="user-head-row">
          <img
            v-if="profileAvatarUrl"
            class="profile-avatar"
            :src="profileAvatarUrl"
            alt=""
            width="48"
            height="48"
            loading="lazy"
          />
          <h1 class="page-title">{{ profile.username }}</h1>
        </div>
        <nav class="user-tabs" aria-label="User sections">
          <button type="button" :class="{ active: userTab === 'overview' }" @click="userTab = 'overview'">Overview</button>
          <button type="button" :class="{ active: userTab === 'following' }" @click="userTab = 'following'">Following</button>
          <button type="button" :class="{ active: userTab === 'messages' }" @click="userTab = 'messages'">Messages</button>
          <button type="button" :class="{ active: userTab === 'activity' }" @click="userTab = 'activity'">Activity</button>
          <button type="button" :class="{ active: userTab === 'graphs' }" @click="userTab = 'graphs'">Graphs</button>
          <button type="button" :class="{ active: userTab === 'settings' }" @click="userTab = 'settings'">Settings</button>
        </nav>
      </header>

      <section v-show="userTab === 'overview'" class="panel">
        <div v-if="profile.is_sus || profile.sus_description" class="sus-banner">
          <div class="sus-banner-head">
            <span v-if="profile.is_sus" class="sus-badge">Suspicious</span>
            <span v-if="profile.sus_type" class="muted sus-type">{{ profile.sus_type }}</span>
          </div>
          <p v-if="profile.sus_description" class="sus-desc">{{ profile.sus_description }}</p>
          <p v-if="profile.sus_auto_suppressed" class="muted sus-hint">
            Automatic suspicion is off until you use “Allow automatic suspicion again”.
          </p>
        </div>

        <div class="sus-actions">
          <label class="sus-note">
            <span class="label">Note (optional, when marking manually)</span>
            <input v-model="manualSusNote" type="text" autocomplete="off" placeholder="Reason or label" />
          </label>
          <div class="sus-buttons">
            <button type="button" class="btn-sus" :disabled="togglingSus" @click="markSuspicious">Mark suspicious</button>
            <button type="button" class="btn-sus-secondary" :disabled="togglingSus" @click="clearSuspicion">Clear</button>
            <button
              v-if="profile.sus_auto_suppressed"
              type="button"
              class="btn-sus-secondary"
              :disabled="togglingSus"
              @click="allowAutoSuspicionAgain"
            >
              Allow automatic suspicion again
            </button>
          </div>
        </div>

        <dl class="meta">
          <div>
            <dt>Twitch id</dt>
            <dd>{{ profile.id }}</dd>
          </div>
          <div>
            <dt>Monitored channel</dt>
            <dd>{{ profile.monitored ? 'Yes' : 'No' }}</dd>
          </div>
          <div>
            <dt>Marked</dt>
            <dd>
              <button
                type="button"
                class="btn-toggle"
                :disabled="togglingMarked"
                @click="toggleMarked"
              >
                {{ profile.marked ? 'Yes (click to clear)' : 'No (click to mark)' }}
              </button>
            </dd>
          </div>
          <div>
            <dt>Suspicious</dt>
            <dd>{{ profile.is_sus ? 'Yes' : 'No' }}</dd>
          </div>
          <div>
            <dt>Chat presence this week (UTC)</dt>
            <dd>{{ formatPresenceWeek(profile.presence_seconds_this_week) }}</dd>
          </div>
          <div>
            <dt>Messages stored</dt>
            <dd>{{ profile.message_count }}</dd>
          </div>
          <div>
            <dt>Account created</dt>
            <dd>{{ accountCreatedLabel ?? '—' }}</dd>
          </div>
        </dl>
        <div v-if="profile.followed_monitored_channels?.length" class="follow-block">
          <h3 class="sub-title">Follows (monitored channels)</h3>
          <ul class="follow-list">
            <li v-for="f in profile.followed_monitored_channels" :key="f.channel_id">
              <span class="ch">#{{ f.channel_login }}</span>
              <span class="muted" v-if="f.followed_at">{{ formatWhen(f.followed_at) }}</span>
              <span class="muted" v-else>—</span>
            </li>
          </ul>
        </div>
      </section>

      <section v-show="userTab === 'following'" class="panel following-panel">
        <h2 class="section-title">Following</h2>
        <p class="muted hint">
          Channels this user follows (synced via enrichment). Blacklisted channels are listed first.
        </p>
        <label class="follow-filter">
          <span class="label">Filter</span>
          <input v-model="followingQuery" type="search" autocomplete="off" placeholder="Channel login contains…" />
        </label>
        <div class="table-wrap">
          <table class="follow-table">
            <thead>
              <tr>
                <th>Channel</th>
                <th>Followed</th>
                <th />
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in sortedFollowedChannels" :key="row.channel_id" :class="{ 'bl-row': row.on_blacklist }">
                <td>
                  <span class="ch">#{{ row.channel_login }}</span>
                  <span v-if="row.on_blacklist" class="tag-bl">blacklist</span>
                </td>
                <td class="muted">{{ row.followed_at ? formatWhen(row.followed_at) : '—' }}</td>
                <td />
              </tr>
            </tbody>
          </table>
        </div>
        <p v-if="!sortedFollowedChannels.length" class="muted">No follows loaded yet (run enrichment) or nothing matches the filter.</p>
      </section>

      <section v-show="userTab === 'messages'" class="panel">
        <h2 class="section-title">Messages</h2>
        <p v-if="loadingMessages" class="muted">Loading messages…</p>
        <ul v-else class="lines">
          <ChatMessageLine
            v-for="m in messages"
            :key="m.id"
            :user="m.user"
            :message="m.message"
            :keyword="m.keyword_match"
            :from-sent="m.source === ChatHistoryEntry.source.SENT"
            :badge-tags="rowBadges(m)"
            :show-timestamp="true"
            :created-at="m.created_at"
            :chatter-user-id="m.chatter_user_id ?? undefined"
            :user-marked="m.chatter_marked"
            :user-is-sus="m.chatter_is_sus"
            show-channel
            :channel-login="m.channel"
          />
        </ul>
        <div v-if="!loadingMessages && messages.length" class="more-row">
          <button type="button" class="btn-more" :disabled="loadingMore" @click="loadMore">
            {{ loadingMore ? 'Loading…' : 'Load more' }}
          </button>
        </div>
        <p v-if="!loadingMessages && !messages.length" class="muted">No messages for this user.</p>
      </section>

      <section v-show="userTab === 'activity'" class="panel">
        <h2 class="section-title">Activity</h2>
        <p v-if="loadingActivity" class="muted">Loading activity…</p>
        <ul v-else class="activity-list">
          <li v-for="e in activity" :key="e.id">
            <span class="act-ts">{{ formatWhen(e.created_at) }}</span>
            <span class="act-body">{{ activityLabel(e) }}</span>
          </li>
        </ul>
        <div v-if="!loadingActivity && activity.length" class="more-row">
          <button type="button" class="btn-more" :disabled="loadingActivityMore" @click="loadActivityMore">
            {{ loadingActivityMore ? 'Loading…' : 'Load more' }}
          </button>
        </div>
        <p v-if="!loadingActivity && !activity.length" class="muted">No activity recorded yet.</p>
      </section>

      <section v-show="userTab === 'settings'" class="panel">
        <h2 class="section-title">Monitoring</h2>
        <p class="muted hint">IRC and notification behavior for this channel (requires admin).</p>
        <ul class="settings-options">
          <li>
            <label class="check-row">
              <input
                type="checkbox"
                :checked="profile.monitored"
                :disabled="savingUserSettings"
                @change="patchUserProfile({ monitored: ($event.target as HTMLInputElement).checked })"
              />
              <span>Monitored channel</span>
            </label>
          </li>
          <li>
            <label class="check-row">
              <input
                type="checkbox"
                :checked="profile.irc_only_when_live"
                :disabled="savingUserSettings"
                @change="
                  patchUserProfile({
                    irc_only_when_live: ($event.target as HTMLInputElement).checked,
                  })
                "
              />
              <span>Only monitor (IRC) when this channel is online on Twitch</span>
            </label>
          </li>
          <li>
            <label class="check-row">
              <input
                type="checkbox"
                :checked="profile.notify_off_stream_messages"
                :disabled="savingUserSettings || !profile.irc_only_when_live"
                @change="
                  patchUserProfile({ notify_off_stream_messages: ($event.target as HTMLInputElement).checked })
                "
              />
              <span>Notify about off-stream messages (join IRC while offline)</span>
            </label>
            <p v-if="!profile.irc_only_when_live" class="muted small indent">
              Turn on “only when online” to allow off-stream IRC for alerts.
            </p>
          </li>
          <li>
            <label class="check-row">
              <input
                type="checkbox"
                :checked="profile.notify_stream_start"
                :disabled="savingUserSettings"
                @change="patchUserProfile({ notify_stream_start: ($event.target as HTMLInputElement).checked })"
              />
              <span>Notify when this channel starts streaming</span>
            </label>
          </li>
        </ul>
      </section>

      <section v-show="userTab === 'graphs'" class="panel graphs-panel">
        <h2 class="section-title">Activity timeline</h2>
        <p class="muted hint">
          In-chat presence from IRC join/leave events over the last 7 days, by channel.
        </p>
        <p v-if="loadingTimeline" class="muted">Loading chart…</p>
        <div v-show="!loadingTimeline" ref="chartEl" class="chart-host" />
        <p v-if="!loadingTimeline && !timelineSegments.length" class="muted">No timeline data in this range.</p>
      </section>
    </template>
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.page-head {
  margin-bottom: 0.75rem;
}

.user-head-row {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.5rem;
}

.profile-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
}

.page-title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
}

.settings-options {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  max-width: 36rem;
}

.check-row {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  cursor: pointer;
  font-size: 0.9rem;

  input {
    margin-top: 0.15rem;
  }
}

.indent {
  margin: 0.25rem 0 0 1.5rem;
}

.user-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;

  button {
    padding: 0.35rem 0.65rem;
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
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin: 0;
  font-size: 0.85rem;

  div {
    margin: 0;
  }

  dt {
    margin: 0;
    font-size: 0.72rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  dd {
    margin: 0.15rem 0 0;
  }
}

.btn-toggle {
  padding: 0.25rem 0.5rem;
  font-size: 0.82rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-base);
  color: var(--text);
  cursor: pointer;

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    background: var(--bg-hover);
  }
}

.sub-title {
  margin: 0.75rem 0 0.35rem;
  font-size: 0.92rem;
  font-weight: 600;
}

.follow-list {
  list-style: none;
  margin: 0;
  padding: 0;
  font-size: 0.85rem;

  li {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: baseline;
    padding: 0.2rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  }

  .ch {
    font-weight: 600;
    color: var(--accent-bright);
    min-width: 8rem;
  }
}

.section-title {
  margin: 0 0 0.5rem;
  font-size: 1rem;
  font-weight: 600;
}

.muted {
  color: var(--text-muted);
  font-size: 0.88rem;
}

.hint {
  margin: 0 0 0.5rem;
  font-size: 0.8rem;
}

.lines {
  list-style: none;
  margin: 0;
  padding: 0;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  font-size: 0.82rem;
  line-height: 1.35;
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
}

.activity-list {
  list-style: none;
  margin: 0;
  padding: 0;
  font-size: 0.82rem;
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
  max-height: min(70vh, 520px);
  overflow-y: auto;

  li {
    display: grid;
    grid-template-columns: 9rem 1fr;
    gap: 0.35rem 0.75rem;
    padding: 0.35rem 0.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
    align-items: start;
  }

  .act-ts {
    font-size: 0.72rem;
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .act-body {
    color: var(--text);
  }

  .act-preview {
    grid-column: 2;
    color: var(--text-muted);
    word-break: break-word;
  }
}

.graphs-panel .chart-host {
  width: 100%;
  height: min(70vh, 420px);
  min-height: 280px;
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
}

.more-row {
  margin-top: 0.5rem;
}

.btn-more {
  padding: 0.4rem 0.85rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-base);
  color: var(--text);
  font-size: 0.85rem;
  cursor: pointer;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.sus-banner {
  margin-bottom: 0.75rem;
  padding: 0.65rem 0.75rem;
  border-radius: 0.35rem;
  border: 1px solid rgba(220, 53, 69, 0.45);
  background: rgba(220, 53, 69, 0.1);
}

.sus-banner-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.sus-badge {
  font-size: 0.78rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #ff6b7a;
}

.sus-type {
  font-size: 0.78rem;
}

.sus-desc {
  margin: 0.35rem 0 0;
  font-size: 0.88rem;
  line-height: 1.4;
}

.sus-hint {
  margin: 0.35rem 0 0;
  font-size: 0.78rem;
}

.sus-actions {
  margin-bottom: 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.sus-note {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  max-width: 28rem;

  .label {
    font-size: 0.72rem;
    color: var(--text-muted);
  }

  input {
    padding: 0.35rem 0.45rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.85rem;
  }
}

.sus-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.btn-sus {
  padding: 0.35rem 0.65rem;
  border-radius: 0.25rem;
  border: 1px solid rgba(220, 53, 69, 0.55);
  background: rgba(220, 53, 69, 0.2);
  color: #ffb3bc;
  font-size: 0.82rem;
  cursor: pointer;

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    background: rgba(220, 53, 69, 0.32);
  }
}

.btn-sus-secondary {
  padding: 0.35rem 0.65rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font-size: 0.82rem;
  cursor: pointer;

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    background: var(--bg-hover);
  }
}

.following-panel .follow-filter {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  margin-bottom: 0.65rem;
  max-width: 22rem;

  .label {
    font-size: 0.72rem;
    color: var(--text-muted);
  }

  input {
    padding: 0.4rem 0.45rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.85rem;
  }
}

.table-wrap {
  overflow-x: auto;
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
}

.follow-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.82rem;

  th,
  td {
    padding: 0.4rem 0.55rem;
    text-align: left;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  }

  th {
    font-size: 0.72rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  tr.bl-row {
    background: rgba(220, 53, 69, 0.08);
  }

  .tag-bl {
    margin-left: 0.35rem;
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: #ff6b7a;
  }
}
</style>
