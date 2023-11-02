import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import Pages from 'vite-plugin-pages';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), Pages({})],
  server: {
    port: 4000,
    proxy: {
      '/api': {
        target: 'http://localhost:12345',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
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
