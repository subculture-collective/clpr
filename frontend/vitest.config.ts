import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react-swc';
import path from 'path';
import criticalCoverage from './config/critical-coverage.json';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      buffer: 'buffer/',
    },
  },
  define: {
    global: 'globalThis',
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
    exclude: ['node_modules', 'dist', 'e2e'],
    css: false,
    forbidOnly: true,
    // Bound memory use consistently on developer machines and CI runners.
    maxWorkers: 4,
    pool: 'forks',
    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        'e2e/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/types/',
      ],
      // Report the whole application; gate critical modules independently so
      // unrelated UI coverage cannot conceal a security or state regression.
      thresholds: criticalCoverage,
    },
  },
});
