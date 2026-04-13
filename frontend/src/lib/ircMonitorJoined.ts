import type { IrcMonitorStatus } from '../api/generated';

function normLogin(c: string): string {
  return c.replace(/^#/, '').toLowerCase();
}

/** Same semantics as Settings → Channels per-row IRC dot (TCP up + reconciler joined for that login). */
export function isChannelJoinedOnIrc(st: IrcMonitorStatus | null | undefined, login: string): boolean {
  if (!st?.connected) {
    return false;
  }
  const low = normLogin(login);
  if (!low) {
    return false;
  }
  const row = st.channels.find((c) => normLogin(c.login) === low);
  const r = row as { irc_ok?: boolean; ircOk?: boolean } | undefined;
  return Boolean(r?.irc_ok ?? r?.ircOk);
}
