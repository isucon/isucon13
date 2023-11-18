import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import Pages from 'vite-plugin-pages';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), Pages({})],
  server: {
    port: 4000,
    https: {
      key: '../provisioning/ansible/roles/nginx/files/etc/nginx/tls/_.u.isucon.dev.key',
      cert: '../provisioning/ansible/roles/nginx/files/etc/nginx/tls/_.u.isucon.dev.crt',
    },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
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
