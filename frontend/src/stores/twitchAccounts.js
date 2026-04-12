import { defineStore } from 'pinia';
import { ref } from 'vue';
import { DefaultService } from '../api/generated';
export const useTwitchAccountsStore = defineStore('twitchAccounts', () => {
    const accounts = ref([]);
    const loaded = ref(false);
    async function fetch() {
        accounts.value = await DefaultService.listTwitchAccounts();
        loaded.value = true;
    }
    return { accounts, loaded, fetch };
});
