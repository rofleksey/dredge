import { defineStore } from 'pinia';
import { ref } from 'vue';

export type NotificationType = 'success' | 'info' | 'warning' | 'error';

export type ShowNotificationInput = {
  title: string;
  description: string;
  type: NotificationType;
  id?: string;
};

export type NotificationItem = {
  key: string;
  groupId?: string;
  title: string;
  description: string;
  type: NotificationType;
  durationMs: number;
};

function defaultDurationMs(type: NotificationType): number {
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

function randomKey(): string {
  return crypto.randomUUID();
}

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationItem[]>([]);
  const timers = new Map<string, ReturnType<typeof setTimeout>>();

  function clearTimer(key: string): void {
    const t = timers.get(key);
    if (t !== undefined) {
      clearTimeout(t);
      timers.delete(key);
    }
  }

  function removeByKey(key: string): void {
    clearTimer(key);
    items.value = items.value.filter((n) => n.key !== key);
  }

  function removeByGroupId(groupId: string): void {
    const toRemove = items.value.filter((n) => n.groupId === groupId);
    for (const n of toRemove) {
      clearTimer(n.key);
    }
    items.value = items.value.filter((n) => n.groupId !== groupId);
  }

  function show(input: ShowNotificationInput): void {
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
