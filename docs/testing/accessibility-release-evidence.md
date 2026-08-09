# Accessibility release evidence

Date: 2026-07-12

Owner: Frontend/QA

Standard: WCAG 2.1 AA, with critical and serious axe findings treated as release blockers

## Automated browser coverage

Run:

```bash
cd frontend
npx playwright test e2e/mocked/accessibility.spec.ts --project=mocked-chromium
```

The suite covers login, search, clip detail, pricing, forum, submission authentication,
settings authentication, and moderation authorization boundaries. It also checks:

- zero critical or serious axe violations;
- skip-link focus transfer;
- mobile navigation touch targets of at least 44 CSS pixels;
- Escape dismissal and focus restoration;
- 320 CSS pixel reflow without horizontal document overflow;
- the reduced-motion media preference.

The shared Modal, Button, Input, ReplyComposer, PaywallModal, ClipGridCard, and
OptimizedImage component suites separately cover accessible names, error association,
busy/pressed state, focus trapping/restoration, and semantic interactive structure.
Automated account deletion is intentionally absent from the launch surface; destructive
confirmation semantics remain covered by the shared modal/confirmation component tests.

## Chromium keyboard and accessibility-tree check

The same release candidate was inspected in Chromium at desktop and 375×812 mobile
viewports. Accessibility-tree snapshots confirmed one focusable element per global
navigation action, labelled search/sort controls, a labelled forum topic group, and a
labelled mobile navigation region. Keyboard checks confirmed Escape closes the mobile
menu and restores focus to its trigger. Browser history restores the prior scroll
position through a separately tested layout behavior.

## Findings resolved during the check

- Invalid nested link/button navigation produced duplicate keyboard stops across the
  application. The shared Button now composes links as one element and the detected
  occurrences were migrated.
- `/forum` crashed with `ReferenceError: threads is not defined`; it now uses the sorted
  query result and has a page regression test.
- Search and forum sort selects lacked accessible names.
- The global link color and pricing badges failed AA contrast. Link text now uses a
  lighter semantic token, and pricing badges use contrast-safe foreground/background
  pairs.

No lower-severity axe debt is approved for the journeys listed above. Authenticated
business behavior remains validated by the real-backend smoke tier; the mocked suite is
used only for deterministic accessibility presentation and authorization-boundary checks.
