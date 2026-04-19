import type { CreateRuleRequest } from '../api/generated/models/CreateRuleRequest';
import type { Rule } from '../api/generated/models/Rule';
import type { RuleActionType } from '../api/generated/models/RuleActionType';
import type { RuleEventType } from '../api/generated/models/RuleEventType';
import type { RuleMiddleware } from '../api/generated/models/RuleMiddleware';
import type { UpdateRulePostRequest } from '../api/generated/models/UpdateRulePostRequest';
import { RuleActionType as AT } from '../api/generated/models/RuleActionType';
import { RuleEventType as ET } from '../api/generated/models/RuleEventType';

export const MIDDLEWARE_TYPES = [
  'filter_channel',
  'filter_user',
  'match_regex',
  'contains_word',
  'cooldown',
] as const;

export type MiddlewareKind = (typeof MIDDLEWARE_TYPES)[number];

export type MiddlewareFormRow = {
  key: string;
  type: MiddlewareKind;
  includeLogins: string;
  excludeLogins: string;
  requireOnline: boolean;
  pattern: string;
  caseInsensitive: boolean;
  words: string;
  seconds: string;
};

function newRowKey(): string {
  return typeof crypto !== 'undefined' && crypto.randomUUID
    ? crypto.randomUUID()
    : `mw-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export function defaultMiddlewareRow(type: MiddlewareKind = 'match_regex'): MiddlewareFormRow {
  return {
    key: newRowKey(),
    type,
    includeLogins: '',
    excludeLogins: '',
    requireOnline: false,
    pattern: '',
    caseInsensitive: false,
    words: '',
    seconds: '',
  };
}

/** Split comma- or newline-separated logins; trim and lowercase (matches server trimLower). */
export function parseLoginList(s: string): string[] {
  const parts = s.split(/[\s,]+/u);
  const out: string[] = [];
  for (const p of parts) {
    const x = p.trim().toLowerCase();
    if (x) {
      out.push(x);
    }
  }
  return out;
}

export function parseWords(s: string): string[] {
  const parts = s.split(/[,\n]+/u);
  const out: string[] = [];
  for (const p of parts) {
    const w = p.trim();
    if (w) {
      out.push(w);
    }
  }
  return out;
}

export function middlewareRowToApi(row: MiddlewareFormRow): RuleMiddleware {
  switch (row.type) {
    case 'filter_channel': {
      const settings: Record<string, unknown> = {};
      const inc = parseLoginList(row.includeLogins);
      const exc = parseLoginList(row.excludeLogins);
      if (inc.length) {
        settings.include_logins = inc;
      }
      if (exc.length) {
        settings.exclude_logins = exc;
      }
      if (row.requireOnline) {
        settings.require_online = true;
      }
      return { type: row.type, settings };
    }
    case 'filter_user': {
      const settings: Record<string, unknown> = {};
      const inc = parseLoginList(row.includeLogins);
      const exc = parseLoginList(row.excludeLogins);
      if (inc.length) {
        settings.include_logins = inc;
      }
      if (exc.length) {
        settings.exclude_logins = exc;
      }
      return { type: row.type, settings };
    }
    case 'match_regex':
      return {
        type: row.type,
        settings: {
          pattern: row.pattern,
          ...(row.caseInsensitive ? { case_insensitive: true } : {}),
        },
      };
    case 'contains_word':
      return {
        type: row.type,
        settings: {
          words: parseWords(row.words),
          case_insensitive: row.caseInsensitive,
        },
      };
    case 'cooldown': {
      const sec = Number.parseFloat(row.seconds);
      return {
        type: row.type,
        settings: { seconds: Number.isFinite(sec) ? sec : 0 },
      };
    }
    default:
      return { type: 'match_regex', settings: { pattern: '' } };
  }
}

function apiRowToForm(mw: RuleMiddleware): MiddlewareFormRow {
  const s = mw.settings ?? {};
  const row = defaultMiddlewareRow(mw.type as MiddlewareKind);
  row.type = (MIDDLEWARE_TYPES as readonly string[]).includes(mw.type)
    ? (mw.type as MiddlewareKind)
    : 'match_regex';

  if (row.type === 'filter_channel' || row.type === 'filter_user') {
    const inc = s.include_logins;
    const exc = s.exclude_logins;
    row.includeLogins = Array.isArray(inc) ? (inc as unknown[]).filter((x) => typeof x === 'string').join(' ') : '';
    row.excludeLogins = Array.isArray(exc) ? (exc as unknown[]).filter((x) => typeof x === 'string').join(' ') : '';
    row.requireOnline = Boolean(s.require_online);
  }
  if (row.type === 'match_regex') {
    row.pattern = typeof s.pattern === 'string' ? s.pattern : '';
    row.caseInsensitive = Boolean(s.case_insensitive);
  }
  if (row.type === 'contains_word') {
    const w = s.words;
    if (Array.isArray(w)) {
      row.words = (w as unknown[]).filter((x) => typeof x === 'string').join('\n');
    } else {
      row.words = '';
    }
    row.caseInsensitive = typeof s.case_insensitive === 'boolean' ? s.case_insensitive : true;
  }
  if (row.type === 'cooldown') {
    const sec = s.seconds;
    if (typeof sec === 'number') {
      row.seconds = String(sec);
    } else if (typeof sec === 'string') {
      row.seconds = sec;
    }
  }
  return row;
}

export type RuleFormState = {
  enabled: boolean;
  useSharedPool: boolean;
  eventType: RuleEventType;
  intervalSeconds: string;
  intervalChannel: string;
  middlewares: MiddlewareFormRow[];
  actionType: RuleActionType;
  notifyText: string;
  sendAccountId: string;
  sendChannel: string;
  sendMessage: string;
};

export function defaultRuleForm(): RuleFormState {
  return {
    enabled: true,
    useSharedPool: true,
    eventType: ET.CHAT_MESSAGE,
    intervalSeconds: '60',
    intervalChannel: '',
    middlewares: [defaultMiddlewareRow('match_regex')],
    actionType: AT.NOTIFY,
    notifyText: '[$CHANNEL] $USERNAME: $TEXT',
    sendAccountId: '',
    sendChannel: '',
    sendMessage: '',
  };
}

export function ruleToFormState(r: Rule): RuleFormState {
  const st = defaultRuleForm();
  st.enabled = r.enabled;
  st.useSharedPool = r.use_shared_pool;
  st.eventType = r.event_type;
  const es = r.event_settings ?? {};
  if (r.event_type === ET.INTERVAL) {
    const sec = es.interval_seconds;
    if (typeof sec === 'number') {
      st.intervalSeconds = String(sec);
    } else if (typeof sec === 'string') {
      st.intervalSeconds = sec;
    }
    st.intervalChannel = typeof es.channel === 'string' ? es.channel : '';
  }
  st.middlewares = r.middlewares?.length ? r.middlewares.map(apiRowToForm) : [defaultMiddlewareRow()];
  st.actionType = r.action_type;
  const as = r.action_settings ?? {};
  if (r.action_type === AT.NOTIFY) {
    st.notifyText = typeof as.text === 'string' ? as.text : '';
  }
  if (r.action_type === AT.SEND_CHAT) {
    const aid = as.account_id;
    if (typeof aid === 'number') {
      st.sendAccountId = String(aid);
    } else if (typeof aid === 'string') {
      st.sendAccountId = aid;
    }
    st.sendChannel = typeof as.channel === 'string' ? as.channel : '';
    st.sendMessage = typeof as.message === 'string' ? as.message : '';
  }
  return st;
}

function buildEventSettings(st: RuleFormState): Record<string, unknown> {
  if (st.eventType === ET.INTERVAL) {
    const sec = Number.parseFloat(st.intervalSeconds);
    return {
      interval_seconds: Number.isFinite(sec) ? sec : 0,
      channel: st.intervalChannel.trim(),
    };
  }
  return {};
}

function buildActionSettings(st: RuleFormState): Record<string, unknown> {
  if (st.actionType === AT.NOTIFY) {
    const t = st.notifyText.trim();
    return t ? { text: t } : {};
  }
  const aid = Number.parseInt(st.sendAccountId.trim(), 10);
  return {
    account_id: Number.isFinite(aid) ? aid : 0,
    channel: st.sendChannel.trim(),
    message: st.sendMessage,
  };
}

export function formStateToCreateRequest(st: RuleFormState): CreateRuleRequest {
  const middlewares = st.middlewares.map(middlewareRowToApi);
  return {
    enabled: st.enabled,
    event_type: st.eventType,
    event_settings: buildEventSettings(st),
    middlewares,
    action_type: st.actionType,
    action_settings: buildActionSettings(st),
    use_shared_pool: st.useSharedPool,
  };
}

export function formStateToUpdateRequest(id: number, st: RuleFormState): UpdateRulePostRequest {
  const middlewares = st.middlewares.map(middlewareRowToApi);
  return {
    id,
    enabled: st.enabled,
    event_type: st.eventType,
    event_settings: buildEventSettings(st),
    middlewares,
    action_type: st.actionType,
    action_settings: buildActionSettings(st),
    use_shared_pool: st.useSharedPool,
  };
}

export function validateRuleForm(st: RuleFormState): string | null {
  if (st.eventType === ET.INTERVAL) {
    const sec = Number.parseFloat(st.intervalSeconds);
    if (!Number.isFinite(sec) || sec <= 0) {
      return 'Interval rules need a positive interval (seconds).';
    }
    if (!st.intervalChannel.trim()) {
      return 'Interval rules need a channel login.';
    }
  }

  if (st.actionType === AT.SEND_CHAT) {
    const aid = Number.parseInt(st.sendAccountId.trim(), 10);
    if (!Number.isFinite(aid) || aid <= 0) {
      return 'send_chat requires a positive Twitch account id.';
    }
    if (!st.sendChannel.trim()) {
      return 'send_chat requires a channel template.';
    }
    if (!st.sendMessage.trim()) {
      return 'send_chat requires a message template.';
    }
  }

  for (let i = 0; i < st.middlewares.length; i++) {
    const row = st.middlewares[i];
    const err = validateMiddlewareRow(row, i);
    if (err) {
      return err;
    }
  }

  return null;
}

export function validateMiddlewareRow(row: MiddlewareFormRow, index: number): string | null {
  const n = index + 1;
  switch (row.type) {
    case 'match_regex': {
      if (!row.pattern.trim()) {
        return `Middleware #${n} (match_regex): pattern is required.`;
      }
      try {
        // eslint-disable-next-line no-new
        new RegExp(row.pattern);
      } catch {
        return `Middleware #${n} (match_regex): invalid regular expression.`;
      }
      return null;
    }
    case 'contains_word': {
      const words = parseWords(row.words);
      if (words.length === 0) {
        return `Middleware #${n} (contains_word): add at least one word (line or comma separated).`;
      }
      return null;
    }
    case 'cooldown': {
      const sec = Number.parseFloat(row.seconds);
      if (!Number.isFinite(sec) || sec <= 0) {
        return `Middleware #${n} (cooldown): seconds must be a positive number.`;
      }
      return null;
    }
    default:
      return null;
  }
}
