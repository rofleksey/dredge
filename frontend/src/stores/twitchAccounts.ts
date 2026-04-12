import { defineStore } from 'pinia';
import { ref } from 'vue';
import { DefaultService } from '../api/generated';
import type { TwitchAccount } from '../api/generated';

export const useTwitchAccountsStore = defineStore('twitchAccounts', () => {
  const accounts = ref<TwitchAccount[]>([]);
  const loaded = ref(false);

  async function fetch(): Promise<void> {
    accounts.value = await DefaultService.listTwitchAccounts();
    loaded.value = true;
  }

  return { accounts, loaded, fetch };
});
