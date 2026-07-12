import react from '@vitejs/plugin-react-swc';
import path from 'path';
import { defineConfig, type Plugin } from 'vite';
import { sentryVitePlugin } from '@sentry/vite-plugin';
import { visualizer } from 'rollup-plugin-visualizer';

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
    // Optimize dependencies to handle CommonJS modules like lucide-react
    optimizeDeps: {
        include: ['lucide-react', '@tanstack/react-query', 'react-router-dom'],
    },
    plugins: [
        react(),
        // Upload source maps to Sentry on production builds
        sentryVitePlugin({
            org: process.env.SENTRY_ORG,
            project: process.env.SENTRY_PROJECT,
            authToken: process.env.SENTRY_AUTH_TOKEN,
            // Only upload on production builds when auth token is set
            // Use Vite's provided `mode` instead of NODE_ENV for reliability
            disable: !process.env.SENTRY_AUTH_TOKEN || mode !== 'production',
            // Align the upload release with runtime init
            release: (() => {
                const name = process.env.VITE_SENTRY_RELEASE || process.env.SENTRY_RELEASE;
                return name ? { name } : undefined;
            })(),
            sourcemaps: {
                assets: './dist/**',
                filesToDeleteAfterUpload: ['**/*.js.map', '**/*.mjs.map'],
            },
        }),
        // Bundle analyzer - only in analyze mode
        ...(mode === 'analyze' ? [visualizer({
            filename: './dist/stats.html',
            open: false,
            gzipSize: true,
            brotliSize: true,
        }) as Plugin] : []),
    ],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },
    server: {
        port: 5173,
        strictPort: true,
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
    build: {
        // Only generate source maps when Sentry will upload them (hidden = no sourceMappingURL comment in output)
        sourcemap: process.env.SENTRY_AUTH_TOKEN ? 'hidden' : false,
        // Eliminate React init races by bundling entry as a single JS file
        // so React and ReactDOM initialize in the same execution unit.
        rollupOptions: {
            output: {
                // Create vendor chunks for better caching and preloading
                manualChunks: {
                    'query-vendor': ['@tanstack/react-query'],
                    'router-vendor': ['react-router-dom'],
                    'icons-vendor': ['lucide-react'],
                },
                // Enable dynamic imports as separate chunks
                inlineDynamicImports: false,
                entryFileNames: 'assets/app-[hash].js',
                chunkFileNames: 'assets/chunk-[hash].js',
                assetFileNames: 'assets/[name]-[hash][extname]',
            },
        },
        // Split CSS per lazy-loaded route for smaller initial payloads
        cssCodeSplit: true,
        // Increase chunk size warning limit since we may have a larger single bundle
        chunkSizeWarningLimit: 1200,
        // Use esbuild minifier (default, faster than terser)
        minify: 'esbuild',
    },
}));
