import { defineConfig } from 'vitest/config';

export default defineConfig({
  resolve: {
    alias: [
      {
        find: '~/',
        replacement: '/src/',
      },
    ],
  },
});
