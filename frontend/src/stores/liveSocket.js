import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { getToken } from '../api/client';
import { useAuthStore } from './auth';
function normChannel(c) {
    return c.replace(/^#/, '').toLowerCase();
}
const badgeTagSet = new Set(['moderator', 'vip', 'bot', 'other']);
function parseBadgeTags(raw) {
    if (!Array.isArray(raw)) {
        return undefined;
    }
    const out = [];
    for (const x of raw) {
        if (typeof x === 'string' && badgeTagSet.has(x)) {
            out.push(x);
        }
    }
    return out.length ? out : undefined;
}
function buildWsUrl() {
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
    const lastError = ref(null);
    const events = ref([]);
    let ws = null;
    let hadSuccessfulOpen = false;
    function pushEvent(raw) {
        if (!raw || typeof raw !== 'object') {
            return;
        }
        const o = raw;
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
            return;
        }
        if (t === 'chatter_join') {
            const uid = o.user_twitch_id;
            if (typeof uid !== 'number' || !Number.isFinite(uid)) {
                return;
            }
            events.value.push({
                type: 'chatter_join',
                channel: String(o.channel ?? ''),
                user: String(o.user ?? ''),
                user_twitch_id: uid,
                present_since: typeof o.present_since === 'string' ? o.present_since : '',
                account_created_at: typeof o.account_created_at === 'string' ? o.account_created_at : undefined,
                created_at: typeof o.created_at === 'string' ? o.created_at : undefined,
                receivedAt: ts,
            });
            return;
        }
        if (t === 'chatter_part') {
            const uid = o.user_twitch_id;
            const ps = o.present_seconds;
            if (typeof uid !== 'number' || !Number.isFinite(uid)) {
                return;
            }
            let sec = 0;
            if (typeof ps === 'number' && Number.isFinite(ps)) {
                sec = Math.floor(ps);
            }
            events.value.push({
                type: 'chatter_part',
                channel: String(o.channel ?? ''),
                user: String(o.user ?? ''),
                user_twitch_id: uid,
                present_seconds: sec,
                present_since: typeof o.present_since === 'string' ? o.present_since : '',
                created_at: typeof o.created_at === 'string' ? o.created_at : undefined,
                receivedAt: ts,
            });
        }
    }
    function trimEvents(max = 500) {
        if (events.value.length > max) {
            events.value = events.value.slice(-max);
        }
    }
    function connect() {
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
                const data = JSON.parse(ev.data);
                pushEvent(data);
                trimEvents();
            }
            catch {
                /* ignore */
            }
        });
    }
    function disconnect() {
        if (ws) {
            ws.close();
            ws = null;
        }
        connected.value = false;
    }
    function clearEvents() {
        events.value = [];
    }
    function clearEventsForChannel(channel) {
        const want = normChannel(channel);
        if (!want) {
            return;
        }
        events.value = events.value.filter((e) => normChannel(e.channel) !== want);
    }
    watch(() => auth.isAuthenticated, (v) => {
        if (v) {
            connect();
        }
        else {
            disconnect();
            events.value = [];
            lastError.value = null;
        }
    }, { immediate: true });
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
