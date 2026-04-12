import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
const STORAGE_KEY = 'dredge.chatGapMinutes';
function clampGap(m) {
    if (!Number.isFinite(m) || m < 1) {
        return 10;
    }
    return Math.min(1440, Math.floor(m));
}
function loadGap() {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw == null) {
        return 10;
    }
    return clampGap(parseInt(raw, 10));
}
export const useWatchPreferencesStore = defineStore('watchPreferences', () => {
    const chatGapMinutes = ref(loadGap());
    watch(chatGapMinutes, (v) => {
        if (typeof v !== 'number' || !Number.isFinite(v)) {
            return;
        }
        localStorage.setItem(STORAGE_KEY, String(clampGap(v)));
    });
    function setChatGapMinutes(m) {
        chatGapMinutes.value = clampGap(Number(m) || 10);
    }
    return { chatGapMinutes, setChatGapMinutes };
});
