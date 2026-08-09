# Frontend module and UI boundaries

These boundaries are release gates, not style suggestions. `npm run boundaries:check`,
ESLint, route tests, and the production bundle budget enforce them in source CI.

## Application shell

`src/main.tsx`, `src/App.tsx`, and layout modules import concrete files rather than broad
component barrels. ESLint rejects shell imports from the aggregate component, UI, video,
layout, guard, and consent entry points. This keeps expensive media and administrative
features outside the initial route.

## Versioned routes

Routes with an authorization contract live in `src/routes/v1`:

- `AccountRoutes.tsx` owns signed-in product routes and explicitly identifies the public
  playlist views that do not use `ProtectedRoute`.
- `AdminRoutes.tsx` owns administration and moderation routes; every entry is wrapped by
  `AdminRoute`.

The root app retains public marketing/content routes and injects lazy page components
into these route families. Route tests fail if representative protected or public
boundaries change.

## Shared UI primitives

- Transient dialogs use `Modal`, which supplies the accessible name, focus trap and
  restoration, Escape/backdrop behavior, body scroll lock, reduced motion, and
  overscroll containment.
- Button-styled links use `Button asChild`, producing one interactive element.
- Dense counters use `formatCompactNumber`; duration and relative/exact time use the
  utilities in `src/lib/utils.ts`. Plain `toLocaleString`/`toLocaleDateString` remains
  intentional for locale-sensitive tables and long-form administrative values.
- Complex forms may use native inputs directly when they require specialized layout,
  but must retain explicit labels and error associations. The shared `Input` remains the
  default for ordinary validated fields.
- Server state uses React Query hooks/API modules. Local `useState` is reserved for
  transient presentation state rather than duplicating remote caches.

The only approved custom dialog implementations are the consent decision surface,
emoji picker, and immersive theatre player. Full-screen theatre/loading surfaces and
the mobile chat backdrop are overlays rather than transient dialogs and are allowlisted
by the boundary check. Adding another custom dialog, overlay, local compact-number
formatter, or nested link/button fails CI until the implementation is consolidated or
this document and the narrowly reviewed allowlist are updated.
