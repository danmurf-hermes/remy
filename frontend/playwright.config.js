import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: 0,
  use: {
    headless: true,
  },
  webServer: {
    command: process.env.REMY_BINARY || '../../build/remy',
    port: 8080,
    reuseExistingServer: true,
  },
});
