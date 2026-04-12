import { storeToRefs } from 'pinia';
import { computed, toValue } from 'vue';
import { useChannelsStore } from '../stores/channels';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';
export function normChannelLogin(c) {
    return c.replace(/^#/, '').toLowerCase();
}
/**
 * CSS classes for chat usernames (streamer / your accounts / monitored channels).
 * Matches message list styling (hl-streamer, hl-mine, hl-monitored).
 */
export function useTwitchUserHighlight(highlightChannel) {
    const channelsStore = useChannelsStore();
    const twitchStore = useTwitchAccountsStore();
    const { channels } = storeToRefs(channelsStore);
    const highlightClass = computed(() => {
        const ch = normChannelLogin(toValue(highlightChannel));
        return (login) => {
            const u = login.toLowerCase().trim();
            if (ch && u === ch) {
                return 'hl-streamer';
            }
            if (twitchStore.accounts.some((a) => a.username.toLowerCase() === u)) {
                return 'hl-mine';
            }
            if (channels.value.some((c) => normChannelLogin(c.username) === u)) {
                return 'hl-monitored';
            }
            return '';
        };
    });
    return { highlightClass };
}
