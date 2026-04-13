import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { getToken } from '../api/client';
import type { ChatBadgeTag } from '../lib/chatBadges';
import { useAuthStore } from './auth';

export type { ChatBadgeTag };

export type LiveEvent =
  | {
      type: 'chat_message';
      channel: string;
      user: string;
      message: string;
      keyword_match?: boolean;
      chatter_marked?: boolean;
      badge_tags?: ChatBadgeTag[];
      created_at?: string;
      receivedAt: number;
      user_twitch_id?: number;
    }
  | {
      type: 'message_sent';
      channel: string;
      user: string;
      message: string;
      badge_tags?: ChatBadgeTag[];
      created_at?: string;
      receivedAt: number;
      user_twitch_id?: number;
    };

function normChannel(c: string): string {
  return c.replace(/^#/, '').toLowerCase();
}

const badgeTagSet = new Set<string>(['moderator', 'vip', 'bot', 'other']);

function parseBadgeTags(raw: unknown): ChatBadgeTag[] | undefined {
  if (!Array.isArray(raw)) {
    return undefined;
  }
  const out: ChatBadgeTag[] = [];
  for (const x of raw) {
    if (typeof x === 'string' && badgeTagSet.has(x)) {
      out.push(x as ChatBadgeTag);
    }
  }
  return out.length ? out : undefined;
}

function buildWsUrl(): string {
  const tok = getToken();
  const base = import.meta.env.VITE_API_BASE;
  if (base) {
    const u = new URL(base);
    const wsProto = u.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${wsProto}//${u.host}/ws?token=${encodeURIComponent(tok)}`;
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(tok)}`;
}

export const useLiveSocketStore = defineStore('liveSocket', () => {
  const auth = useAuthStore();
  const connected = ref(false);
  const lastError = ref<string | null>(null);
  const events = ref<LiveEvent[]>([]);

  let ws: ReconnectingWebSocket | null = null;
  let hadSuccessfulOpen = false;

  function pushEvent(raw: unknown): void {
    if (!raw || typeof raw !== 'object') {
      return;
    }
    const o = raw as Record<string, unknown>;
    const t = o.type;

    const ts = Date.now();
    if (t === 'chat_message') {
      const uid = o.user_twitch_id;
      events.value.push({
        type: 'chat_message',
        channel: String(o.channel ?? ''),
        user: String(o.user ?? ''),
        message: String(o.message ?? ''),
        keyword_match: Boolean(o.keyword_match),
        chatter_marked: Boolean(o.chatter_marked),
        badge_tags: parseBadgeTags(o.badge_tags),
        created_at: typeof o.created_at === 'string' ? o.created_at : undefined,
        receivedAt: ts,
        user_twitch_id: typeof uid === 'number' && Number.isFinite(uid) ? uid : undefined,
      });
      return;
    }
    if (t === 'message_sent') {
      const uid = o.user_twitch_id;
      events.value.push({
        type: 'message_sent',
        channel: String(o.channel ?? ''),
        user: String(o.user ?? ''),
        message: String(o.message ?? ''),
        badge_tags: parseBadgeTags(o.badge_tags),
        created_at: typeof o.created_at === 'string' ? o.created_at : undefined,
        receivedAt: ts,
        user_twitch_id: typeof uid === 'number' && Number.isFinite(uid) ? uid : undefined,
      });
    }
  }

  function trimEvents(max = 500): void {
    if (events.value.length > max) {
      events.value = events.value.slice(-max);
    }
  }

  function connect(): void {
    hadSuccessfulOpen = false;
    disconnect();
    if (!getToken()) {
      return;
    }
    ws = new ReconnectingWebSocket(buildWsUrl, [], {
      maxReconnectionDelay: 10000,
      minReconnectionDelay: 1000,
      reconnectionDelayGrowFactor: 1.3,
      maxRetries: Infinity,
    });

    ws.addEventListener('open', () => {
      connected.value = true;
      lastError.value = null;
      hadSuccessfulOpen = true;
    });
    ws.addEventListener('close', () => {
      connected.value = false;
      if (hadSuccessfulOpen && auth.isAuthenticated) {
        lastError.value = 'Connection lost';
      }
    });
    ws.addEventListener('error', () => {
      lastError.value = 'WebSocket error';
    });
    ws.addEventListener('message', (ev) => {
      try {
        const data = JSON.parse(ev.data as string);
        pushEvent(data);
        trimEvents();
      } catch {
        /* ignore */
      }
    });
  }

  function disconnect(): void {
    if (ws) {
      ws.close();
      ws = null;
    }
    connected.value = false;
  }

  function clearEvents(): void {
    events.value = [];
  }

  function clearEventsForChannel(channel: string): void {
    const want = normChannel(channel);
    if (!want) {
      return;
    }
    events.value = events.value.filter((e) => normChannel(e.channel) !== want);
  }

  watch(
    () => auth.isAuthenticated,
    (v) => {
      if (v) {
        connect();
      } else {
        disconnect();
        events.value = [];
        lastError.value = null;
      }
    },
    { immediate: true },
  );

  return {
    connected,
    lastError,
    events,
    connect,
    disconnect,
    clearEvents,
    clearEventsForChannel,
  };
});
