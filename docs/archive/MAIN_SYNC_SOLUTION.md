---
title: Main Branch Sync & Chunking Solution
summary: Resolves main branch divergence and React initialization errors through optimized Vite configuration.
tags: ["archive", "implementation"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---

# Main Branch Sync & Chunking Solution

## Problem

- Main branch had diverged from origin/main
- Attempted fixes made chunking worse, causing React initialization errors
- Need stable, reliable builds for deployment

## Solution: No Chunking Strategy

The original `vite.config.ts` on origin/main is the **correct solution**:

```typescript
rollupOptions: {
    output: {
        manualChunks: undefined,          // NO chunking
        inlineDynamicImports: true,       // Everything in ONE bundle
    },
},
cssCodeSplit: false,                      // CSS stays together
chunkSizeWarningLimit: 1200,              // Accept larger single bundle
```

## Why This Works

**Single Bundle = Reliable:**
1. React initializes first in the bundle
2. All dependencies are available immediately
3. No race conditions between chunks
4. No "Cannot set properties of undefined" errors
5. Predictable load order

**Trade-off: Size vs Reliability**
- Gzipped bundle: ~389 KB (large but acceptable)
- No multiple concurrent requests needed
- Simpler caching strategy
- More reliable user experience

## Build Output

```
dist/assets/app-a7yJgo3I.js      1,363 KB (gzipped: 389 KB)
dist/assets/style-BnphCyAL.css   68.5 KB (gzipped: 11.6 KB)
dist/index.html                  1.56 KB
```

## Current Status

✅ Main synced with origin/main (commit: 70d04a2)
✅ Vite config set to no-chunking strategy
✅ Frontend builds successfully with single bundle
✅ No React initialization errors in this config

## Deployment

The current config should:
1. Build without errors
2. Deploy successfully
3. Work reliably in production
4. **No more "Activity is undefined" errors**

## When Main Should Update

- Only merge PRs that pass CI/CD tests
- Deploy should rebuild and redeploy automatically
- Monitor GitHub Actions for success/failure

## Future Optimization

If bundle size becomes an issue (>500 KB gzipped):
- Use code-splitting strategically for **lazy-loaded routes only**
- Keep React ecosystem in main bundle
- Never split React/ReactDOM from entry point
- Verify in staging before production

## Files

- `frontend/vite.config.ts` - Build configuration (no chunking)
- `.github/workflows/deploy-production.yml` - Auto-deployment on main push
- Origin: <https://git.subcult.tv/subculture-collective/clpr>

## Testing

Verify the build:
```bash
cd frontend
npm run build
# Should see single app-*.js file in dist/assets/
```

Test in browser:
```bash
npm run preview
# Open http://localhost:4173
# Should load without React initialization errors
```
