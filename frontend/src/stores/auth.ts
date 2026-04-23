import { useLocalStorage } from '@vueuse/core';
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { configureApi, withPendingAuthToken } from '../api/client';
import { DefaultService } from '../api/generated';
import type { Account } from '../api/generated';

export const useAuthStore = defineStore('auth', () => {
  const token = useLocalStorage('dredge_token', '');
  const user = ref<Account | null>(null);
  const bootstrapped = ref(false);

  const isAuthenticated = computed(() => !!token.value);

  async function bootstrap(): Promise<void> {
    configureApi();
    if (!token.value) {
      bootstrapped.value = true;
      return;
    }
    try {
      user.value = await DefaultService.me();
    } catch {
      logout();
    } finally {
      bootstrapped.value = true;
    }
  }

  async function login(email: string, password: string): Promise<void> {
    configureApi();
    const res = await DefaultService.login({ requestBody: { email, password } });
    await withPendingAuthToken(res.token, async () => {
      token.value = res.token;
      user.value = await DefaultService.me();
    });
  }

  function logout(): void {
    token.value = '';
    user.value = null;
  }

  return { token, user, bootstrapped, isAuthenticated, bootstrap, login, logout };
});
