import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { DefaultService } from '../api/generated';
import type { TwitchUser } from '../api/generated';

export const useChannelsStore = defineStore('channels', () => {
  const channels = ref<TwitchUser[]>([]);
  const loaded = ref(false);

  const monitoredChannels = computed(() => channels.value.filter((c) => c.monitored));

  async function fetch(): Promise<void> {
    channels.value = await DefaultService.listTwitchUsers({ monitoredOnly: true });
    loaded.value = true;
  }

  return { channels, monitoredChannels, loaded, fetch };
});
