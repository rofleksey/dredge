import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const target = 'http://localhost:8080';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      'form-data': fileURLToPath(new URL('./src/shims/form-data.ts', import.meta.url)),
    },
    // Default order prefers `.js` over `.ts`, so stray `foo.js` + `foo.ts` in `src/` loads stale JS. Prefer TS first.
    extensions: ['.mjs', '.mts', '.ts', '.tsx', '.vue', '.js', '.jsx', '.json'],
  },
  server: {
    proxy: {
      '/auth': { target, changeOrigin: true },
      '/me': { target, changeOrigin: true },
      '/users': { target, changeOrigin: true },
      '/settings': { target, changeOrigin: true },
      '/twitch': { target, changeOrigin: true },
      '/metrics': { target, changeOrigin: true },
      '/ws': { target: 'ws://localhost:8080', ws: true, changeOrigin: true },
    },
  },
});
