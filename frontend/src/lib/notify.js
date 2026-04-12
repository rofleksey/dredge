import { getActivePinia } from 'pinia';
import { useNotificationsStore, } from '../stores/notifications';
export function notify(input) {
    const pinia = getActivePinia();
    if (!pinia) {
        console.warn('notify: no active Pinia instance');
        return;
    }
    useNotificationsStore(pinia).show(input);
}
