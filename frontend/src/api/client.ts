import { OpenAPI } from './generated';

const TOKEN_KEY = 'dredge_token';

export function getToken(): string {
  if (typeof localStorage === 'undefined') {
    return '';
  }
  return localStorage.getItem(TOKEN_KEY) ?? '';
}

/** Call once at app startup (after env is available). */
export function configureApi(): void {
  // Generated client paths include `/api/v1`, so BASE should be origin-only.
  OpenAPI.BASE = import.meta.env.VITE_API_BASE ?? '';
  OpenAPI.TOKEN = async () => getToken();
}
