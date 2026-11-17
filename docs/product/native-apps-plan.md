---
title: "Separate native iOS and Android apps"
summary: "Planning scope, API prerequisites and verification for two native CLPR clients."
tags: ["product", "mobile", "planning"]
area: "product"
status: "draft"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-16
---

# Separate native iOS and Android apps

## Goal and current state

On November 16, 2026 the owner requested planning for two separate native apps.
The owner confirmed a viewer-first private beta: discovery, search, Twitch
playback, saved clips/playlists and submitting an existing clip. Creator tools
and clip-room workflows come later, after that journey is qualified.

There is no buildable native app in this repository. The earlier React Native
and Expo material is [historical planning](../archive/mobile/index.md). This
document starts a new effort; it does not establish a release date or authorize
store publication. [Roadmap #1436](https://github.com/subculture-collective/clpr/issues/1436)
tracks the shared prerequisites and separate platform issues. Implementation
issues remain blocked until the relevant spikes settle their contracts.

## Architecture decisions

| Area | iOS | Android |
| --- | --- | --- |
| Application | Swift and SwiftUI, proposed workspace `apps/ios` | Kotlin and Jetpack Compose, proposed workspace `apps/android` |
| Navigation | Native navigation stack, tabs, sheets and system share sheet | Native navigation, back behavior, sheets and Android sharesheet |
| Networking | URLSession with typed request/response models | Platform-native HTTP client with coroutines and typed models; select the library in the build spike |
| Authentication browser | ASWebAuthenticationSession | Auth Tab with Custom Tabs fallback |
| Sensitive local state | Keychain-backed credentials | Credentials encrypted with a key protected by Android Keystore |
| UI state | Feature-owned observable state and structured concurrency | ViewModel-owned state, StateFlow and unidirectional data flow |
| Verification | Swift unit tests, XCUITest and real iPhone checks | Kotlin unit tests, Compose UI tests and real Android checks |

Both apps use the existing Go API and its authorization rules. Share OpenAPI
contracts, JSON fixtures, localization keys, brand tokens and acceptance journeys.
The [native API contract pack](../openapi/mobile/index.md) provides the shared
response fixtures and `task mobile:contracts` check.
Each app owns its UI, navigation, storage and platform integration. Choose the
minimum supported OS versions, signing owners, application identifiers and CI
build hosts during the build spike; no choice is represented as configured.

Use CLPR's current [design system](../design-system.md) for identity, typography
and content emphasis, with each platform's native controls. Translate web tokens
into semantic platform colors and accessible type sizes. Keep loading, empty,
offline, removed-content and permission-denied states visible in the designs.

## First private beta

| Journey | Beta contract |
| --- | --- |
| Anonymous discovery | Browse ranked clips, creator/topic/category pages and countdown pages with paging, refresh and recoverable errors |
| Search | Text and transcript phrase search, filters, result snippets and shareable search state |
| Playback | Poster-first detail screen; Twitch playback only through a qualified embed or Twitch link, with a visible fallback |
| Sign-in and account | Twitch login, existing MFA challenge when required, session refresh, logout and account/support access |
| Library | Save/unsave clips, create/edit a personal playlist, add/remove clips and open a shared public playlist |
| Submission and sharing | Submit an existing Twitch clip URL; accept a shared URL into a reviewable draft; use the system share sheet |
| Safety | Existing report and block actions where exposed, moderation/removed-content states, and an accessible support destination |
| Lifecycle | Cold/warm deep links, background/foreground, interrupted login, process death and network loss |

Initial navigation proposal: Discover, Search and Library, with submission in the
system share flow and a visible in-app action. Anonymous browsing stays useful;
signed-in actions explain why login is needed and return to the interrupted task.

Keep platform administration, countdown rendering/publishing, streamer clip-room
management, push notifications, offline video downloads and background playback
outside this beta. Cache bounded public metadata and thumbnails; clearly mark
stale/offline data and require a connection for writes and video playback.
Comments and voting can follow the first saved-clip/playlist journey; do not claim
full web parity for the initial beta.

## Gate 1: authentication and account lifecycle spike

Existing source already has PKCE initiation and completion, bearer authentication,
refresh-token bodies and a shared MFA gate:

- `backend/internal/handlers/auth_handler.go`
- `backend/internal/handlers/login_session.go`
- `backend/internal/middleware/auth_middleware.go`
- `backend/internal/middleware/csrf_middleware.go`
- `backend/cmd/api/routes_auth.go`

The browser callback currently redirects to the configured web frontend. PKCE
completion sets session cookies, and MFA uses a pending browser challenge.
Those pieces are not evidence of a complete native sign-in flow.

The spike must record and test the smallest extension needed to return control
to each app and establish its own CLPR session. Prefer the system browser, a
verified HTTPS callback and a one-use, short-lived code bound to PKCE/state;
keep long-lived credentials out of URLs and Twitch client secrets on the server.
Do not assume provider support for a new OAuth parameter from CLPR's own PKCE
implementation. Prove the actual Twitch exchange in staging.

Acceptance: cold/warm callback, cancellation, invalid state/verifier, replay,
expired code, MFA enabled and disabled, refresh rotation, revoked credentials,
logout, and an account ban. Ordinary API responses must never expose provider
tokens. Confirm cookie-authenticated web requests keep CSRF protection when
native bearer sessions are introduced. Document key storage, log redaction and
how a lost device's session is revoked before shipping authenticated features.

Automated account deletion and full personal-data export remain
[disabled](../LAUNCH_FEATURE_INVENTORY.md). Define the supported account/privacy
request path for the beta. Before store submission, review current account
deletion, login, user-generated-content and privacy requirements and close the
actual gaps. Store approval and exceptions are not assumed.

## Gate 2: Twitch playback and deep-link spike

Twitch documents a minimum clip embed size of 400 by 300 CSS pixels and a
required `parent` host. A narrow portrait viewport cannot simply shrink that
player. The current web implementation already falls back when it does not fit
(`frontend/src/components/clip/TwitchEmbed.tsx`).

Evaluate a small HTTPS-hosted player surface in the platform web view, with the
real parent host and an unobstructed player, and a clear Watch on Twitch fallback.
Do not build a native player around extracted Twitch media URLs. This is a spike
proposal, not a claim that native web-view embedding has been accepted by Twitch.

Acceptance on a real iPhone and Android phone: portrait and landscape, measured
embed dimensions, clip unavailable/deleted, Twitch app installed and absent,
slow network, user-initiated audio, navigation away/back and interruption.
Verify HTTPS clip and playlist links launch the correct app screen when installed
and retain a usable web destination when absent. Record associated-domain/App
Links configuration and rejection of unsupported external URLs.

## Proposed delivery sequence

1. **Shared prerequisites:** complete the two spikes, API error/paging fixtures,
   account/privacy gap assessment, and a list of required API changes. Each
   backend change is independently reviewed and keeps the web client working.
2. **Independent foundations:** one iOS PR and one Android PR establish build,
   signing configuration, native navigation, brand tokens, networking and CI.
   Each proves anonymous discovery against staging on its own platform.
3. **Read-only journeys:** separate platform PRs add search, clip detail,
   qualified playback, countdown pages and verified deep links. Shared fixtures
   cover missing/null fields, paging, rate limits and removed content.
4. **Account and library:** after the auth contract lands, separate platform PRs
   add MFA-aware login, refresh/logout, saved clips and playlists. Prove revoked
   sessions and cross-account cache isolation.
5. **Submission and safety:** separate platform PRs add reviewed URL submission,
   system sharing, report/block/support and accessible error recovery.
6. **Private beta qualification:** distribute through the available private
   platform channels only after signing and tester access are established.
   Store submission is a later, explicit release decision.

Each platform can progress independently after shared prerequisites. A PR for
one app does not claim verification of the other. Do not create implementation
workspaces merely to make this plan look started.

## Beta exit evidence

- Reproducible signed builds and passing platform CI on the exact beta commits.
- Real staging API journeys; mocked tests recorded separately.
- VoiceOver/TalkBack, large text, contrast, keyboard/focus, reduced motion and
  touch-target checks on both apps.
- Physical-phone login/MFA, keyboard, rotation, playback, deep links and lifecycle
  checks. Simulator/emulator results remain useful but separate evidence.
- No tokens or personal content in logs/crash reports; permission and analytics
  choices documented; logout/account switch clears private local state.
- A dated beta issue list and feedback path, plus independent stop/rollback and
  session-revocation instructions for each platform.

## Sources and decisions still to make

Official guidance checked November 16, 2026:

- [Apple web authentication sessions](https://developer.apple.com/documentation/authenticationservices/authenticating-a-user-through-a-web-service)
- [Compose UI architecture](https://developer.android.com/develop/ui/compose/architecture)
- [Android Auth Tab and fallback](https://developer.chrome.com/docs/android/custom-tabs/guide-auth-tab)
- [Android deep links and App Links](https://developer.android.com/training/app-links)
- [Twitch clip embeds](https://dev.twitch.tv/docs/embed/video-and-clips/)
- [Twitch OAuth token flows](https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/)
- [Apple App Review Guidelines](https://developer.apple.com/app-store/review/guidelines/)

The first beta audience is confirmed: viewers, with creator tools later.
Remaining decisions: supported OS/device range,
signing/distribution ownership, the native session callback contract, and the
qualified playback approach. The first two spikes turn the technical unknowns
into evidence before implementation dates are committed.
