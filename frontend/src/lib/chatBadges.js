const BADGE_ORDER = ['moderator', 'vip', 'bot', 'other'];
const BADGE_EMOJI = {
    moderator: '⚔️',
    vip: '💎',
    bot: '🤖',
    other: '✨',
};
export function badgeEmojis(tags) {
    let s = '';
    const set = new Set(tags);
    for (const t of BADGE_ORDER) {
        if (set.has(t)) {
            s += BADGE_EMOJI[t];
        }
    }
    return s;
}
