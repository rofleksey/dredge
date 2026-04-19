import type { ChannelLive, TwitchUser } from '../../api/generated';

export const TWITCH_LOGIN_RE = /^[a-zA-Z0-9_]{4,25}$/;

export function normCh(c: string): string {
  return c.replace(/^#/, '').toLowerCase();
}

/** Rejects error-shaped JSON bodies so we do not keep a stale LIVE strip when the API returns `{ message }`. */
export function parseChannelLivePayload(data: unknown): ChannelLive | null {
  if (!data || typeof data !== 'object') {
    return null;
  }
  const o = data as Record<string, unknown>;
  if (typeof o.broadcaster_login !== 'string') {
    return null;
  }
  return data as ChannelLive;
}

export function directoryRowToChannelLive(row: TwitchUser): ChannelLive {
  const parsed = row.channel_live ? parseChannelLivePayload(row.channel_live) : null;
  if (parsed) {
    return parsed;
  }
  return {
    broadcaster_id: row.id,
    broadcaster_login: normCh(row.username),
    display_name: row.username,
    profile_image_url: row.profile_image_url ?? '',
    is_live: false,
  };
}
