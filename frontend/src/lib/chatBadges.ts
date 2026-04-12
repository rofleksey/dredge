export type ChatBadgeTag = 'moderator' | 'vip' | 'bot' | 'other';

const BADGE_ORDER: ChatBadgeTag[] = ['moderator', 'vip', 'bot', 'other'];

const BADGE_EMOJI: Record<ChatBadgeTag, string> = {
  moderator: '⚔️',
  vip: '💎',
  bot: '🤖',
  other: '✨',
};

export function badgeEmojis(tags: readonly ChatBadgeTag[]): string {
  let s = '';
  const set = new Set(tags);
  for (const t of BADGE_ORDER) {
    if (set.has(t)) {
      s += BADGE_EMOJI[t];
    }
  }
  return s;
}
