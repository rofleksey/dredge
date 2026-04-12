import { defineStore } from 'pinia';
import { ref } from 'vue';
function defaultDurationMs(type) {
    switch (type) {
        case 'success':
            return 3000;
        case 'info':
            return 5000;
        case 'warning':
        case 'error':
            return 10000;
    }
}
function randomKey() {
    return crypto.randomUUID();
}
export const useNotificationsStore = defineStore('notifications', () => {
    const items = ref([]);
    const timers = new Map();
    function clearTimer(key) {
        const t = timers.get(key);
        if (t !== undefined) {
            clearTimeout(t);
            timers.delete(key);
        }
    }
    function removeByKey(key) {
        clearTimer(key);
        items.value = items.value.filter((n) => n.key !== key);
    }
    function removeByGroupId(groupId) {
        const toRemove = items.value.filter((n) => n.groupId === groupId);
        for (const n of toRemove) {
            clearTimer(n.key);
        }
        items.value = items.value.filter((n) => n.groupId !== groupId);
    }
    function show(input) {
        const durationMs = defaultDurationMs(input.type);
        if (input.id) {
            removeByGroupId(input.id);
        }
        const key = randomKey();
        items.value.push({
            key,
            groupId: input.id,
            title: input.title,
            description: input.description,
            type: input.type,
            durationMs,
        });
        const handle = setTimeout(() => removeByKey(key), durationMs);
        timers.set(key, handle);
    }
    return { items, show };
});
