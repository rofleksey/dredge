import { getActivePinia } from 'pinia';
import {
  useNotificationsStore,
  type ShowNotificationInput,
} from '../stores/notifications';

export type { ShowNotificationInput };

export function notify(input: ShowNotificationInput): void {
  const pinia = getActivePinia();
  if (!pinia) {
    console.warn('notify: no active Pinia instance');
    return;
  }
  useNotificationsStore(pinia).show(input);
}
