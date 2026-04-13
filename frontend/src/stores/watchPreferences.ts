import { defineStore } from 'pinia';
import { ref, watch } from 'vue';

const STORAGE_KEY = 'dredge.chatGapMinutes';
const STORAGE_KEY_LAST_CHANNEL = 'dredge.lastWatchChannel';

function readStorage(key: string): string | null {
  try {
    if (typeof localStorage === 'undefined') {
      return null;
    }
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeStorage(key: string, value: string): void {
  try {
    if (typeof localStorage === 'undefined') {
      return;
    }
    localStorage.setItem(key, value);
  } catch {
    /* quota, security policy, or disabled storage */
  }
}

function clampGap(m: number): number {
  if (!Number.isFinite(m) || m < 1) {
    return 10;
  }
  return Math.min(1440, Math.floor(m));
}

function loadGap(): number {
  const raw = readStorage(STORAGE_KEY);
  if (raw == null) {
    return 10;
  }
  return clampGap(parseInt(raw, 10));
}

function normChannelLogin(raw: string): string {
  return raw.replace(/^#/, '').trim().toLowerCase();
}

export const useWatchPreferencesStore = defineStore('watchPreferences', () => {
  const chatGapMinutes = ref(loadGap());

  watch(chatGapMinutes, (v) => {
    if (typeof v !== 'number' || !Number.isFinite(v)) {
      return;
    }
    writeStorage(STORAGE_KEY, String(clampGap(v)));
  });

  function setChatGapMinutes(m: number): void {
    chatGapMinutes.value = clampGap(Number(m) || 10);
  }

  function getLastWatchChannel(): string {
    const raw = readStorage(STORAGE_KEY_LAST_CHANNEL);
    if (raw == null) {
      return '';
    }
    return normChannelLogin(raw);
  }

  function setLastWatchChannel(login: string): void {
    const n = normChannelLogin(login);
    if (!n) {
      return;
    }
    writeStorage(STORAGE_KEY_LAST_CHANNEL, n);
  }

  return { chatGapMinutes, setChatGapMinutes, getLastWatchChannel, setLastWatchChannel };
});
