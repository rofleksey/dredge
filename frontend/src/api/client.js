import { OpenAPI } from './generated';
const TOKEN_KEY = 'dredge_token';
export function getToken() {
    if (typeof localStorage === 'undefined') {
        return '';
    }
    return localStorage.getItem(TOKEN_KEY) ?? '';
}
/** Call once at app startup (after env is available). */
export function configureApi() {
    OpenAPI.BASE = import.meta.env.VITE_API_BASE ?? '';
    OpenAPI.TOKEN = async () => getToken();
}
