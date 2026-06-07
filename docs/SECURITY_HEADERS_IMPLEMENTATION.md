# Security Headers Implementation Guide

This document covers the security improvements implemented based on ZAP baseline scan results.

## Summary

### Issues Addressed

1. ✅ **Content Security Policy (CSP)** - Added in report-only mode
2. ✅ **Site Isolation Headers** - COOP/CORP for Spectre mitigation
3. ✅ **Optimized Caching Policies** - Path-based caching for performance
4. ✅ **Comment Stripping** - Verified build pipeline strips comments

### Current Status

- **CSP**: Currently in **Report-Only** mode (safe monitoring phase)
- **COOP/CORP**: **Active** (safe to deploy immediately)
- **Caching**: **Active** with path-based rules
- **Build Pipeline**: Already strips comments via esbuild minification

---

## 1. Content Security Policy (CSP)

### Current Configuration (Report-Only)

```
Content-Security-Policy-Report-Only: default-src 'self';
  script-src 'self';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data:;
  font-src 'self' data:;
  connect-src 'self' https: wss:;
  object-src 'none';
  base-uri 'none';
  frame-ancestors 'none';
  upgrade-insecure-requests
```

### What This Means

- **Report-Only**: Violations are logged in the browser console but NOT blocked
- **Monitoring Period**: Run for 24-48 hours to identify any legitimate violations
- **Safe Deployment**: Will not break your site during monitoring

### Monitoring for Violations

1. **Open browser developer console** (F12)
2. **Navigate through your app** (test all features)
3. **Look for CSP violation warnings** - they'll appear as:
    ```
    [Report Only] Refused to load ... because it violates the following
    Content Security Policy directive: ...
    ```
4. **Record any violations** that are from legitimate resources

### Common Violations You Might See

| Violation     | Cause                   | Fix                                  |
| ------------- | ----------------------- | ------------------------------------ |
| `connect-src` | Third-party API calls   | Add specific domain to `connect-src` |
| `script-src`  | Inline scripts or CDN   | Add `'sha256-...'` hash or domain    |
| `style-src`   | Third-party stylesheets | Add domain to `style-src`            |
| `img-src`     | External images         | Add domain to `img-src`              |

### Activating Enforced Mode

After 24-48 hours of monitoring with **NO violations** (or violations resolved):

1. **Edit Caddyfile.vps** (or Caddyfile if not using VPS)
2. **Comment out the Report-Only line**:
    ```caddy
    # Content-Security-Policy-Report-Only "..."
    ```
3. **Uncomment the enforced CSP line**:
    ```caddy
    Content-Security-Policy "default-src 'self'; script-src 'self'; ..."
    ```
4. **Reload Caddy**:
    ```bash
    docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile
    ```
    Or if using systemd:
    ```bash
    sudo systemctl reload caddy
    ```

### Customizing CSP Directives

If you use **third-party services**, add them to the appropriate directives:

```caddy
# Example: Adding Sentry for error reporting
connect-src 'self' https://o*.ingest.sentry.io

# Example: Adding Google Fonts
font-src 'self' data: https://fonts.gstatic.com
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com

# Example: Adding analytics
script-src 'self' https://www.google-analytics.com
connect-src 'self' https: wss: https://www.google-analytics.com
```

> ⚠️ **Important**: Keep directives as restrictive as possible. Only add what you actually need.

---

## 2. Site Isolation Headers (COOP/CORP)

### Active Configuration

```
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Resource-Policy: same-origin
```

### What These Do

- **COOP**: Prevents other sites from accessing your window object
- **CORP**: Prevents other sites from embedding your resources
- **Spectre Mitigation**: Protects against timing attacks

### Safe to Deploy Immediately

These headers are **safe for most SPAs** unless you:

- Embed your site in iframes on other domains
- Share resources (images, fonts) with other origins
- Use cross-origin window.open() features

### COEP (Currently Commented Out)

```caddy
# Cross-Origin-Embedder-Policy "require-corp"  # Uncomment if needed
```

**When to enable COEP:**

- You need `SharedArrayBuffer` (high-performance computing)
- You want maximum isolation

**Why it's commented:**

- Can break third-party embeds (YouTube, Twitter, etc.)
- Requires all resources to have CORP or CORS headers
- More complex to implement correctly

---

## 3. Optimized Caching Policies

### Path-Based Caching Rules

| Content Type                                     | Cache Duration     | Rationale                               |
| ------------------------------------------------ | ------------------ | --------------------------------------- |
| **Hashed assets** (`/assets/*`, `*.js`, `*.css`) | 1 year (immutable) | Vite generates content-hashed filenames |
| **Manifest files** (`manifest.json`)             | 1 hour             | Can change with app updates             |
| **SEO files** (`robots.txt`, `sitemap.xml`)      | 1 day              | Change infrequently                     |
| **HTML files**                                   | No cache           | SPA shell must always be fresh          |
| **API responses**                                | No cache (default) | Dynamic data                            |

### Performance Impact

**Before**: All content sent `cache-control: no-cache, no-store`

- Every asset refetched on every page load
- Increased server load
- Slower page loads

**After**: Aggressive caching for immutable assets

- Hashed assets cached for 1 year
- Return visitors only fetch HTML (few KB)
- Reduced bandwidth costs
- Faster page loads

### Why This Is Safe

Vite generates **content-hashed filenames**:

```
assets/app-abc123def.js
assets/chunk-xyz789abc.js
assets/main-def456ghi.css
```

When code changes, **the hash changes**, so browsers fetch the new file automatically.

---

## 4. Comment Stripping & Source Maps

### Current Configuration ✅

Your build pipeline already handles this correctly:

1. **Minification**: `esbuild` strips comments automatically
2. **Source Maps**: Generated for debugging
3. **Source Map Upload**: Uploaded to Sentry for error tracking
4. **Source Map Deletion**: Removed after upload, not deployed to production

### Vite Configuration (Already Set)

```typescript
build: {
  sourcemap: true,  // Generate for Sentry
  minify: 'esbuild',  // Strips comments
}

// Sentry plugin
sentryVitePlugin({
  sourcemaps: {
    assets: './dist/**',
    filesToDeleteAfterUpload: ['**/*.js.map', '**/*.mjs.map'],  // Delete after upload
  },
})
```

### What This Prevents

- ❌ Source code exposure via source maps
- ❌ Build timestamps in comments
- ❌ Developer comments leaking information
- ✅ Clean, minified production bundles

### Verification

Check production deployment:

```bash
# Source maps should NOT exist in production
curl -I https://clpr.tv/assets/app-*.js.map
# Should return 404

# JS files should be minified (no comments)
curl https://clpr.tv/assets/app-*.js | head -n 10
# Should show minified code, no comments
```

---

## Deployment Checklist

### Phase 1: Initial Deployment (Now)

- [x] Update Caddyfile with security headers
- [x] Add CSP in **report-only** mode
- [x] Add COOP/CORP headers
- [x] Implement path-based caching
- [ ] Deploy to production
- [ ] Verify headers via curl

### Phase 2: CSP Monitoring (24-48 hours)

- [ ] Monitor browser console for CSP violations
- [ ] Test all site features (auth, uploads, API calls, etc.)
- [ ] Document any legitimate violations
- [ ] Adjust CSP directives if needed

### Phase 3: CSP Enforcement (After monitoring)

- [ ] Switch from `Content-Security-Policy-Report-Only` to `Content-Security-Policy`
- [ ] Reload Caddy configuration
- [ ] Test site functionality again
- [ ] Monitor for any breaking issues

### Phase 4: Verification (After enforcement)

- [ ] Run ZAP baseline scan again
- [ ] Verify WARN count decreased
- [ ] Check performance improvements (PageSpeed, Core Web Vitals)
- [ ] Monitor error rates in Sentry

---

## Verification Commands

### Check Current Headers

```bash
# Check all security headers
curl -sI https://clpr.tv/ | grep -E "(Content-Security|Cross-Origin|Cache-Control|X-)"

# Specific CSP check
curl -sI https://clpr.tv/ | grep Content-Security

# Specific caching check
curl -sI https://clpr.tv/assets/app-*.js | grep Cache-Control
```

### Expected Output

```
strict-transport-security: max-age=31536000; includeSubDomains; preload
x-content-type-options: nosniff
x-frame-options: DENY
x-xss-protection: 1; mode=block
cross-origin-opener-policy: same-origin
cross-origin-resource-policy: same-origin
content-security-policy-report-only: default-src 'self'; ...
```

### Reload Caddy (Docker)

```bash
# For docker-compose deployment
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile

# Verify Caddy logs
docker logs clpr-caddy --tail 50
```

### Reload Caddy (VPS with systemd)

```bash
# Reload configuration
sudo systemctl reload caddy

# Check status
sudo systemctl status caddy

# View logs
sudo journalctl -u caddy -n 50
```

---

## Expected ZAP Results After Deployment

### Before

- **FAIL**: 0
- **WARN**: ~10 (CSP, COOP/CORP, caching issues)

### After Phase 1 (Report-Only)

- **FAIL**: 0
- **WARN**: ~5 (CSP still report-only, minor issues)

### After Phase 3 (Enforced)

- **FAIL**: 0
- **WARN**: 0-2 (only informational items)

---

## Troubleshooting

### CSP Violations Breaking Functionality

If enforcing CSP breaks your site:

1. **Check browser console** for violation details
2. **Revert to report-only mode**:
    ```caddy
    Content-Security-Policy-Report-Only "..."
    # Content-Security-Policy "..."
    ```
3. **Identify the violating resource**
4. **Add to appropriate directive**
5. **Test again in report-only**
6. **Re-enforce**

### Site Looks Broken After Deployment

1. **Check if assets are loading**:
    ```bash
    # Open browser dev tools > Network tab
    # Look for 404s or blocked resources
    ```
2. **Verify caching isn't too aggressive**:
    ```bash
    # Force refresh (Ctrl+Shift+R or Cmd+Shift+R)
    ```
3. **Check Caddy logs**:
    ```bash
    docker logs clpr-caddy --tail 100
    ```

### Mixed Content Warnings

If you see "Mixed Content" errors after `upgrade-insecure-requests`:

1. **Find the HTTP resource** in console
2. **Update to HTTPS** in your code
3. **Or remove from CSP** if not critical

---

## Additional Resources

- [MDN: Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [MDN: Cross-Origin-Opener-Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cross-Origin-Opener-Policy)
- [Caddy Documentation: header](https://caddyserver.com/docs/caddyfile/directives/header)
- [CSP Evaluator](https://csp-evaluator.withgoogle.com/) - Test your CSP

---

## Summary

All security improvements have been implemented and are ready to deploy:

1. **CSP in report-only mode** - Safe to deploy, monitor for 24-48h
2. **COOP/CORP active** - Safe to deploy immediately
3. **Optimized caching** - Safe to deploy, improves performance
4. **Build pipeline** - Already secure (no changes needed)

**Next step**: Deploy to production and begin CSP monitoring phase.
