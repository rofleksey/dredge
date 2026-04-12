import { useLocalStorage } from '@vueuse/core';
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { configureApi } from '../api/client';
import { DefaultService } from '../api/generated';
export const useAuthStore = defineStore('auth', () => {
    const token = useLocalStorage('dredge_token', '');
    const user = ref(null);
    const bootstrapped = ref(false);
    const isAuthenticated = computed(() => !!token.value);
    async function bootstrap() {
        configureApi();
        if (!token.value) {
            bootstrapped.value = true;
            return;
        }
        try {
            user.value = await DefaultService.me();
        }
        catch {
            logout();
        }
        finally {
            bootstrapped.value = true;
        }
    }
    async function login(email, password) {
        configureApi();
        const res = await DefaultService.login({ requestBody: { email, password } });
        token.value = res.token;
        user.value = await DefaultService.me();
    }
    function logout() {
        token.value = '';
        user.value = null;
    }
    return { token, user, bootstrapped, isAuthenticated, bootstrap, login, logout };
});
