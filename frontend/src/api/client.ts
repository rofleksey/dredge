import { OpenAPI } from './generated';

const TOKEN_KEY = 'dredge_token';

/** While logging in, `useLocalStorage` may not have flushed yet; `getToken()` reads storage directly. */
let pendingAuthToken: string | null = null;

export function getToken(): string {
  if (pendingAuthToken !== null) {
    return pendingAuthToken;
  }
  if (typeof localStorage === 'undefined') {
    return '';
  }
  return localStorage.getItem(TOKEN_KEY) ?? '';
}

/** Runs `fn` with `getToken()` returning `token` (for the first authenticated request right after login). */
export async function withPendingAuthToken<T>(token: string, fn: () => Promise<T>): Promise<T> {
  pendingAuthToken = token;
  try {
    return await fn();
  } finally {
    pendingAuthToken = null;
  }
}

/** Call once at app startup (after env is available). */
export function configureApi(): void {
  // Generated client paths include `/api/v1`, so BASE should be origin-only.
  OpenAPI.BASE = import.meta.env.VITE_API_BASE ?? '';
  OpenAPI.TOKEN = async () => getToken();
}
