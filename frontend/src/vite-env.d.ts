/// <reference types="vite/client" />

export {};

declare module 'vue-router' {
  interface RouteMeta {
    /** When true, the authenticated router outlet grows to fill remaining space (non-watch pages). */
    fillMainOutlet?: boolean;
  }
}

interface ImportMetaEnv {
  readonly VITE_API_BASE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
