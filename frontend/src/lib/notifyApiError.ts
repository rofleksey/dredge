import { ApiError } from '../api/generated';
import { notify } from './notify';

/** Extract a human-readable API error message, or `fallback` if none. */
export function apiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.body && typeof error.body === 'object' && error.body !== null) {
    const msg = (error.body as { message?: unknown }).message;
    if (typeof msg === 'string' && msg.length > 0) {
      return msg;
    }
  }
  return fallback;
}

export function notifyApiError(
  error: unknown,
  options: { id: string; title: string; fallbackMessage: string },
): void {
  notify({
    id: options.id,
    type: 'error',
    title: options.title,
    description: apiErrorMessage(error, options.fallbackMessage),
  });
}
