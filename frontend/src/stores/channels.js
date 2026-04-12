import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { DefaultService } from '../api/generated';
export const useChannelsStore = defineStore('channels', () => {
    const channels = ref([]);
    const loaded = ref(false);
    const monitoredChannels = computed(() => channels.value.filter((c) => c.monitored));
    async function fetch() {
        channels.value = await DefaultService.listTwitchUsers();
        loaded.value = true;
    }
    return { channels, monitoredChannels, loaded, fetch };
});
