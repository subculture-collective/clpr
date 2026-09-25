# Security Headers Deployment Quick Reference

## 🚀 Deployment Steps

### 1. Deploy Changes (Now)

```bash
# Pull latest changes
git pull origin onnwee/local-dev

# If using docker-compose
docker-compose down
docker-compose up -d

# Reload Caddy only (faster)
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile
```

### 2. Verify Headers (Immediately After)

```bash
# Check all security headers are present
curl -sI https://clpr.tv/ | grep -E "(Content-Security|Cross-Origin|Cache-Control)"

# Expected to see:
# - content-security-policy-report-only: (with full CSP)
# - cross-origin-opener-policy: same-origin
# - cross-origin-resource-policy: same-origin
# - cache-control: no-store, must-revalidate (for HTML)

# Check static asset caching
curl -sI https://clpr.tv/assets/app-*.js | grep cache-control
# Expected: cache-control: public, max-age=31536000, immutable
```

### 3. Monitor CSP Violations (24-48 hours)

1. **Open your site**: https://clpr.tv
2. **Open DevTools**: F12 or Cmd+Option+I
3. **Go to Console tab**
4. **Use the site normally** (browse, search, watch clips, login, etc.)
5. **Look for CSP warnings**: They'll say "[Report Only] Refused to load..."
6. **Document any violations** you see

**If you see violations from legitimate resources:**

- Note the domain and resource type
- Update Caddyfile CSP directive to allow it
- Reload Caddy
- Test again

**Example violation fix:**

```
Violation: Refused to connect to 'https://api.example.com' because it violates
           the following Content Security Policy directive: "connect-src 'self' https: wss:"
```

```caddy
# Add the specific domain to connect-src
connect-src 'self' https: wss: https://api.example.com
```

### 4. Enforce CSP (After 24-48h of Clean Monitoring)

```bash
# Edit Caddyfile
nano Caddyfile

# Comment out report-only line:
# Content-Security-Policy-Report-Only "..."

# Uncomment enforced line:
Content-Security-Policy "default-src 'self'; script-src 'self'; ..."

# Save and reload
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile
```

### 5. Run ZAP Scan Again (After Enforcement)

```bash
# Re-run baseline scan
docker run --rm -v $(pwd):/zap/wrk/:rw \
  ghcr.io/zaproxy/zaproxy:stable \
  zap-baseline.py -t https://clpr.tv -r zap-report-after.html

# Compare WARN counts before/after
```

---

## 📊 Expected Results

### Headers Timeline

| Phase                      | CSP Status     | COOP/CORP  | Caching       | ZAP WARNs |
| -------------------------- | -------------- | ---------- | ------------- | --------- |
| Before                     | ❌ Missing     | ❌ Missing | ⚠️ Too strict | ~10       |
| Phase 1 (Now)              | 🟡 Report-Only | ✅ Active  | ✅ Optimized  | ~5        |
| Phase 2 (After monitoring) | ✅ Enforced    | ✅ Active  | ✅ Optimized  | 0-2       |

---

## 🔍 Quick Checks

### Are security headers working?

```bash
curl -sI https://clpr.tv/ | sed -n '1,30p'
```

Look for:

- ✅ `cross-origin-opener-policy: same-origin`
- ✅ `cross-origin-resource-policy: same-origin`
- ✅ `content-security-policy-report-only:` (initially)
- ✅ `strict-transport-security:`

### Is caching working?

```bash
# Static assets should cache for 1 year
curl -sI https://clpr.tv/assets/app-*.js | grep cache-control
# Should show: public, max-age=31536000, immutable

# HTML should NOT cache
curl -sI https://clpr.tv/ | grep cache-control
# Should show: no-store, must-revalidate
```

### Are source maps hidden?

```bash
# Should return 404
curl -I https://clpr.tv/assets/app-*.js.map
```

---

## ⚠️ Rollback Plan

If something breaks:

### 1. Quick Rollback (Disable CSP Only)

```bash
# Edit Caddyfile
nano Caddyfile

# Comment out the CSP line that's causing issues
# Content-Security-Policy "..."

# Reload Caddy
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile
```

### 2. Full Rollback (Revert All Changes)

```bash
git checkout main -- Caddyfile
docker exec clpr-caddy caddy reload --config /etc/caddy/Caddyfile
```

---

## 📝 Monitoring Checklist

- [ ] Headers deployed successfully
- [ ] curl verification shows correct headers
- [ ] Site loads normally in browser
- [ ] No console errors on page load
- [ ] CSP violations logged (if any) and reviewed
- [ ] Monitored for 24-48 hours
- [ ] CSP enforced (after monitoring)
- [ ] ZAP scan re-run showing improvement
- [ ] Performance metrics stable or improved

---

## 🎯 Success Criteria

✅ **Deployment successful when:**

- curl shows all security headers
- Site loads and functions normally
- Browser console shows CSP report-only (not blocking)
- Static assets cache properly (check Network tab)

✅ **Ready to enforce CSP when:**

- 24-48 hours of monitoring complete
- No CSP violations (or all violations resolved)
- All site features tested and working
- Team agrees to enforce

✅ **Final verification when:**

- ZAP scan shows 0-2 WARNs (down from ~10)
- Site performance maintained or improved
- Error rates in Sentry normal
- No user-reported issues

---

## 🔗 Quick Links

- **CSP Evaluator**: https://csp-evaluator.withgoogle.com/
- **Header Test**: https://securityheaders.com/?q=clpr.tv
- **ZAP Baseline Tool**: https://www.zaproxy.org/docs/docker/baseline-scan/
