import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import Pages from 'vite-plugin-pages';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), Pages({})],
  server: {
    port: 4000,
    proxy: {},
  },
  resolve: {
    alias: [
      {
        find: '~/',
        replacement: '/src/',
      },
    ],
  },
});
