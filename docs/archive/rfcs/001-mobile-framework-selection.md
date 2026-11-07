---
title: "001 Mobile Framework Selection"
summary: "**Status:** Accepted"
tags: ["rfcs"]
area: "rfcs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# RFC 001: Mobile Framework Selection

**Status:** Accepted
**Date:** 2025-11-02
**Authors:** Clipper Team
**Decision:** React Native + Expo

## Summary

This RFC evaluates and selects the mobile app technology stack for Clipper iOS and Android applications. After comprehensive analysis, we have chosen **React Native with Expo** as our mobile framework.

## Context

Clipper is a Twitch clip curation platform currently built with:

- **Frontend**: React 19 + TypeScript + Vite
- **Backend**: Go + Gin + PostgreSQL
- **State Management**: TanStack Query + Zustand
- **Styling**: TailwindCSS (via NativeWind for mobile)

We need to build native mobile apps for iOS and Android to:

1. Provide a better mobile user experience than PWA alone
2. Support native features (push notifications, deep linking, biometric auth)
3. Reach users who prefer native apps over web
4. Leverage app store distribution channels

## Goals

1. **Code Sharing**: Maximize code reuse between web and mobile
2. **Developer Velocity**: Enable fast iteration and feature development
3. **Native Quality**: Deliver smooth, native-feeling experiences
4. **Team Alignment**: Leverage existing team skills and knowledge
5. **Ecosystem Maturity**: Choose battle-tested tools with strong communities
6. **Cost Efficiency**: Minimize infrastructure and maintenance overhead

## Evaluation Criteria

| Criterion | Weight | Description |
|-----------|--------|-------------|
| Code Sharing | 30% | Ability to share TypeScript models, utilities, and business logic |
| Developer Experience | 20% | Tooling, debugging, hot reload, testing |
| Performance | 15% | Runtime performance, startup time, memory usage |
| Native Capabilities | 15% | Access to platform APIs, sensors, OS features |
| Ecosystem & Community | 10% | Library availability, community size, documentation |
| Build & Deploy | 10% | CI/CD integration, OTA updates, app store automation |

## Options Considered

### Option 1: React Native + Expo (Recommended ✅)

**Overview:**
React Native is a JavaScript framework for building native mobile apps using React. Expo is a framework and platform built on top of React Native that provides additional tooling, services, and managed workflows.

**Pros:**

- ✅ **Maximum Code Sharing**: Share TypeScript interfaces, types, API clients, utilities, business logic with web
- ✅ **Team Expertise**: Team already proficient in React, TypeScript, and JavaScript ecosystem
- ✅ **Developer Experience**: Excellent with Expo CLI, hot reload, debugging, and dev tools
- ✅ **OTA Updates**: Expo Update Service (EAS Update) provides instant updates without app store review
- ✅ **Simplified Builds**: EAS Build handles iOS/Android builds in the cloud (no Mac required for development)
- ✅ **Rich Ecosystem**: Huge library ecosystem via npm (React Navigation, Reanimated, Gesture Handler, etc.)
- ✅ **Native Quality**: Near-native performance for most use cases with Hermes engine
- ✅ **Deep Linking**: Built-in support via Expo Router
- ✅ **Push Notifications**: Expo Notifications service with unified API
- ✅ **CI/CD Integration**: Easy GitHub Actions integration with EAS CLI
- ✅ **Cost Effective**: Free tier for small teams, reasonable pricing for production

**Cons:**

- ⚠️ **Learning Curve**: Need to learn React Native-specific components and patterns
- ⚠️ **Bridge Overhead**: JavaScript bridge can impact performance for CPU-intensive tasks
- ⚠️ **Bundle Size**: Larger initial bundle size compared to native
- ⚠️ **Custom Native Code**: Requires ejecting to bare workflow for complex native modules (though less common with Expo SDK)

**Architecture:**

```
┌─────────────────────────────────────────┐
│           Shared TypeScript             │
│  (types, API clients, utils, models)   │
└────────────┬────────────────────────────┘
             │
    ┌────────┴────────┐
    │                 │
┌───▼───┐         ┌───▼───────────┐
│  Web  │         │ Mobile (Expo) │
│ React │         │ React Native  │
│ Vite  │         │ EAS Build     │
└───────┘         └───────────────┘
```

**Tech Stack:**

- **Framework**: React Native 0.76+ (with Expo SDK 52+)
- **Navigation**: Expo Router (file-based routing, similar to Next.js)
- **State Management**:
  - TanStack Query (API state, cache, sync)
  - Zustand (global client state)
- **Styling**: NativeWind (TailwindCSS for React Native)
- **HTTP Client**: Axios (shared with web)
- **Authentication**: Expo SecureStore (encrypted local storage)
- **Push Notifications**: Expo Notifications
- **Deep Linking**: Expo Linking
- **OTA Updates**: EAS Update
- **Build System**: EAS Build
- **Analytics**: Expo Analytics + PostHog (shared with web)
- **Error Tracking**: Sentry (shared with web)
- **Testing**:
  - Jest + React Native Testing Library (unit/integration)
  - Detox (E2E)
  - Maestro (UI testing alternative)

**Cost Estimate:**

- EAS Build: Free for development, ~$29/month for team plan
- EAS Update: Included with EAS Build
- EAS Submit: Included with EAS Build
- Push Notifications: Free up to 1M/month

### Option 2: Flutter

**Overview:**
Flutter is Google's UI toolkit for building natively compiled applications using Dart language.

**Pros:**

- ✅ **Excellent Performance**: Compiles to native ARM code, no JavaScript bridge
- ✅ **Beautiful UI**: Rich widget library, smooth animations out of the box
- ✅ **Single Codebase**: True cross-platform with consistent UI
- ✅ **Fast Development**: Hot reload, excellent tooling (Flutter DevTools)
- ✅ **Growing Ecosystem**: Rapidly expanding package ecosystem

**Cons:**

- ❌ **Zero Code Sharing**: Cannot reuse TypeScript code from web frontend
- ❌ **New Language**: Team must learn Dart (different from TypeScript/JavaScript)
- ❌ **Separate State Management**: Cannot share Zustand stores or TanStack Query logic
- ❌ **Duplicate API Clients**: Must rewrite API layer in Dart
- ❌ **Duplicate Type Definitions**: Must maintain separate Dart models
- ❌ **Team Split**: Creates knowledge silos between web and mobile teams
- ❌ **Less Familiar**: Limited team experience with Flutter/Dart ecosystem

**Impact:**

- 2-3x development time due to code duplication
- Higher maintenance burden (fix bugs twice)
- Slower feature parity between web and mobile

### Option 3: Native (Swift/Kotlin)

**Overview:**
Build separate native apps using Swift for iOS and Kotlin for Android.

**Pros:**

- ✅ **Best Performance**: Native platform performance
- ✅ **Platform Best Practices**: Follow platform conventions exactly
- ✅ **Full Platform Access**: Unrestricted access to all APIs

**Cons:**

- ❌ **Zero Code Sharing**: Separate codebases for iOS and Android
- ❌ **3x Development Time**: Build everything twice (plus web)
- ❌ **Team Expertise**: Requires iOS and Android specialists
- ❌ **Maintenance Burden**: Fix bugs in 3 places (web, iOS, Android)
- ❌ **Cost**: Requires larger team to maintain

**Verdict:** Not viable for small team or MVP phase.

## Decision

**We choose React Native + Expo** based on:

1. **Code Sharing (30% weight, 9/10 score)**:
   - Share TypeScript types, API clients, utilities, business logic
   - Reuse existing Zustand stores and TanStack Query hooks
   - Single source of truth for data models
   - Estimated 40-60% code sharing potential

2. **Developer Experience (20% weight, 9/10 score)**:
   - Team already expert in React and TypeScript
   - Expo CLI provides excellent developer experience
   - Hot reload, debugging, and dev tools are mature
   - EAS Build eliminates local build complexity

3. **Performance (15% weight, 7/10 score)**:
   - Hermes engine provides good performance for our use case
   - Our app is content-heavy (clips, lists, comments), not CPU-intensive
   - 60fps animations achievable with Reanimated
   - Acceptable tradeoff vs native for velocity gains

4. **Native Capabilities (15% weight, 8/10 score)**:
   - Expo SDK provides most features we need out of the box
   - Push notifications, deep linking, secure storage all supported
   - Can eject to bare workflow if needed for custom native code

5. **Ecosystem (10% weight, 9/10 score)**:
   - Massive library ecosystem via npm
   - Strong community (GitHub: 119k stars)
   - Expo has excellent documentation and support

6. **Build & Deploy (10% weight, 9/10 score)**:
   - EAS Build simplifies CI/CD dramatically
   - OTA updates enable instant bug fixes
   - Fastlane integration available
   - GitHub Actions workflows straightforward

**Total Weighted Score**: 8.5/10

## Architecture

### Application Architecture

```
┌──────────────────────────────────────────────────────┐
│                  Clipper Mobile                      │
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌────────────────────────────────────────────┐    │
│  │         Presentation Layer                  │    │
│  │  (React Native Components + NativeWind)     │    │
│  └────────────────┬───────────────────────────┘    │
│                   │                                  │
│  ┌────────────────▼───────────────────────────┐    │
│  │           Navigation Layer                  │    │
│  │            (Expo Router)                    │    │
│  └────────────────┬───────────────────────────┘    │
│                   │                                  │
│  ┌────────────────▼───────────────────────────┐    │
│  │            State Layer                      │    │
│  │  ┌──────────────┐  ┌─────────────────┐    │    │
│  │  │ TanStack     │  │    Zustand      │    │    │
│  │  │ Query        │  │  (Global State) │    │    │
│  │  │ (API State)  │  └─────────────────┘    │    │
│  │  └──────────────┘                          │    │
│  └────────────────┬───────────────────────────┘    │
│                   │                                  │
│  ┌────────────────▼───────────────────────────┐    │
│  │          Services Layer                     │    │
│  │  (API Client, Auth, Storage, Push, etc.)   │    │
│  └────────────────┬───────────────────────────┘    │
│                   │                                  │
└───────────────────┼──────────────────────────────────┘
                    │
         ┌──────────▼──────────┐
         │   Backend API       │
         │  (Go + Gin)         │
         └─────────────────────┘
```

### Data Flow

```
User Action → Component
    ↓
Navigation (Expo Router)
    ↓
State Management (TanStack Query / Zustand)
    ↓
API Service (Axios + Shared Types)
    ↓
Backend (Go API)
    ↓
Database (PostgreSQL)
```

### Caching Strategy

```
┌─────────────────────────────────────────────┐
│         Client Cache Layers                 │
├─────────────────────────────────────────────┤
│                                             │
│  1. Memory (TanStack Query)                │
│     - TTL: 5 minutes (staleTime)           │
│     - Max entries: 50                       │
│     - Used for: API responses               │
│                                             │
│  2. Secure Storage (Expo SecureStore)      │
│     - Auth tokens, user session             │
│     - Encrypted by OS                       │
│                                             │
│  3. AsyncStorage                            │
│     - User preferences, settings            │
│     - Offline queue for mutations           │
│     - Unencrypted key-value store           │
│                                             │
│  4. File System (expo-file-system)         │
│     - Media caching (images, videos)        │
│     - TTL: 7 days                           │
│                                             │
└─────────────────────────────────────────────┘
```

## Monorepo Structure

We will integrate the mobile app into the existing monorepo:

```
clpr/
├── frontend/              # React web app
│   ├── src/
│   └── package.json
├── backend/               # Go API
│   └── cmd/api/
├── mobile/                # React Native + Expo (NEW)
│   ├── app/              # Expo Router file-based routes
│   │   ├── (tabs)/       # Bottom tab navigation
│   │   │   ├── index.tsx        # Home/Feed
│   │   │   ├── search.tsx       # Search
│   │   │   ├── favorites.tsx    # Favorites
│   │   │   └── profile.tsx      # Profile
│   │   ├── clip/
│   │   │   └── [id].tsx         # Clip detail
│   │   ├── _layout.tsx          # Root layout
│   │   └── +not-found.tsx       # 404
│   ├── src/
│   │   ├── components/          # React Native components
│   │   ├── hooks/               # Custom hooks
│   │   ├── services/            # API services (shared types)
│   │   ├── stores/              # Zustand stores
│   │   └── utils/               # Utilities
│   ├── assets/                  # Images, fonts, icons
│   ├── app.json                 # Expo configuration
│   ├── eas.json                 # EAS Build configuration
│   ├── package.json
│   └── tsconfig.json
├── shared/                # Shared TypeScript code (NEW)
│   ├── types/            # Shared type definitions
│   │   ├── api.ts       # API request/response types
│   │   ├── models.ts    # Data models (Clip, User, Comment)
│   │   └── index.ts
│   ├── constants/        # Shared constants
│   ├── utils/            # Shared utilities
│   └── package.json
├── docs/
└── package.json          # Root workspace package.json
```

### Workspace Configuration

**Root `package.json`:**

```json
{
  "name": "clpr",
  "private": true,
  "workspaces": [
    "frontend",
    "mobile",
    "shared"
  ],
  "scripts": {
    "dev:web": "npm -w frontend run dev",
    "dev:mobile": "npm -w mobile run start",
    "build:web": "npm -w frontend run build",
    "build:mobile": "npm -w mobile run build",
    "test": "npm -w frontend run test && npm -w mobile run test"
  }
}
```

## Tooling Selection

### Navigation

**Choice:** Expo Router (file-based routing)

**Rationale:**

- Similar to Next.js/Remix (familiar to web devs)
- Type-safe navigation
- Deep linking built-in
- Automatic code splitting

**Alternative Considered:** React Navigation (library-based)

- More manual setup
- Less type-safe
- More boilerplate

### State Management

**Choice:** TanStack Query + Zustand

**Rationale:**

- Already used in web app (consistency)
- TanStack Query handles API state, caching, sync
- Zustand for global client state (auth, theme, etc.)
- Lightweight and performant

**Alternative Considered:** Redux Toolkit

- More boilerplate
- Heavier bundle size
- Overkill for our needs

### Styling

**Choice:** NativeWind (TailwindCSS for React Native)

**Rationale:**

- Familiar utility-first approach (matches web)
- Can reuse Tailwind config from web
- Good performance with compile-time extraction
- Responsive design support

**Alternative Considered:** Styled Components, React Native StyleSheet

- More verbose
- Less consistency with web

### Testing

**Unit/Integration:**

- Jest + React Native Testing Library
- Same testing patterns as web

**E2E:**

- Detox (primary)
- Maestro (evaluation)

**Rationale:**

- Detox is mature, well-documented
- Maestro has better DX but less mature

### Analytics

**Choice:** PostHog (unified with web)

**Rationale:**

- Already integrated in web app
- React Native SDK available
- Self-hosted option for privacy
- Event tracking, session replay, feature flags

**Alternative Considered:** Expo Analytics

- Basic metrics only
- Less powerful than PostHog

### Error Tracking

**Choice:** Sentry (unified with web)

**Rationale:**

- Already integrated in web app
- Excellent React Native support
- Source maps work well with EAS Build
- Performance monitoring included

### OTA Updates

**Choice:** EAS Update (Expo's update service)

**Rationale:**

- Seamless integration with EAS Build
- Instant bug fixes without app store review
- Staged rollouts support
- Rollback capability
- Free tier generous for MVP

**Alternative Considered:** CodePush (Microsoft)

- More complex setup
- Being deprecated in favor of App Center

### Build & Distribution

**Choice:** EAS Build + EAS Submit

**Rationale:**

- Cloud-based builds (no Mac required for dev)
- Handles iOS/Android certificates automatically
- GitHub Actions integration
- Automatic submission to app stores
- Build artifacts stored securely

**Alternative Considered:** Manual builds with Xcode/Android Studio

- Requires local tooling setup
- More manual certificate management
- Slower CI/CD pipeline

### Push Notifications

**Choice:** Expo Notifications + FCM

**Rationale:**

- Unified API for iOS and Android
- Handles tokens and permissions
- Works with EAS Build
- Free for reasonable volume

**Alternative Considered:** Firebase Cloud Messaging (FCM) directly

- More manual setup
- Platform-specific code

## Implementation Plan

### Phase 1: Foundation (Week 1)

**Deliverables:**

- [ ] Initialize Expo project in monorepo
- [ ] Configure EAS Build and EAS Update
- [ ] Set up shared TypeScript package
- [ ] Implement authentication flow (OAuth redirect)
- [ ] Configure NativeWind and design tokens
- [ ] Set up CI/CD with GitHub Actions
- [ ] Test build on iOS and Android simulators

**Acceptance Criteria:**

- App runs on iOS simulator (Xcode Simulator)
- App runs on Android simulator (Android Emulator)
- User can authenticate via Twitch OAuth
- NativeWind styles render correctly

### Phase 2: Core Features (Weeks 2-3)

**Deliverables:**

- [ ] Home feed with infinite scroll
- [ ] Clip detail view
- [ ] Search functionality
- [ ] Favorites/saved clips
- [ ] User profile
- [ ] Comments section
- [ ] Vote (upvote/downvote)

### Phase 3: Native Features (Week 4)

**Deliverables:**

- [ ] Deep linking (open clips from URLs)
- [ ] Push notifications (new clips, mentions)
- [ ] Share sheet integration
- [ ] Biometric authentication (FaceID/TouchID)
- [ ] Offline mode (cached content)

### Phase 4: Polish & Launch (Week 5-6)

**Deliverables:**

- [ ] E2E tests with Detox
- [ ] Performance optimization
- [ ] Analytics integration
- [ ] Error tracking setup
- [ ] App store assets (screenshots, descriptions)
- [ ] Beta testing with TestFlight/Google Play Internal Testing
- [ ] Submit to App Store and Google Play

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: Mobile CI/CD

on:
  push:
    branches: [main, develop]
    paths:
      - 'mobile/**'
      - 'shared/**'
  pull_request:
    paths:
      - 'mobile/**'
      - 'shared/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm -w mobile run lint

      - name: Run type check
        run: npm -w mobile run type-check

      - name: Run tests
        run: npm -w mobile run test

  build-preview:
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - uses: expo/expo-github-action@v8
        with:
          expo-version: latest
          eas-version: latest
          token: ${{ secrets.EXPO_TOKEN }}

      - name: Install dependencies
        run: npm ci

      - name: Build preview
        run: eas build --profile preview --platform all --non-interactive

  build-production:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - uses: expo/expo-github-action@v8
        with:
          expo-version: latest
          eas-version: latest
          token: ${{ secrets.EXPO_TOKEN }}

      - name: Install dependencies
        run: npm ci

      - name: Build production
        run: eas build --profile production --platform all --non-interactive

      - name: Submit to stores
        run: eas submit --platform all --latest
```

## Security Considerations

1. **Secure Storage:**
   - Use Expo SecureStore for tokens (encrypted by OS keychain)
   - Never store secrets in AsyncStorage

2. **API Security:**
   - Same JWT authentication as web
   - Certificate pinning for API calls (future consideration)
   - Refresh token rotation

3. **Deep Linking:**
   - Validate all incoming deep link parameters
   - Use Universal Links (iOS) and App Links (Android)

4. **Code Security:**
   - Hermes bytecode compilation obfuscates JS
   - Enable ProGuard/R8 for Android
   - Monitor security advisories for dependencies

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| App Launch (cold) | < 2s | Time to interactive |
| App Launch (warm) | < 500ms | Time to interactive |
| Screen Transition | 60fps | React Navigation metrics |
| API Response Time | < 100ms | TanStack Query devtools |
| Bundle Size (iOS) | < 50MB | App Store Connect |
| Bundle Size (Android) | < 30MB | APK size |
| Memory Usage | < 200MB | Xcode Instruments |

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Performance issues | High | Low | Profile early, optimize hot paths, use Hermes |
| Native module needs | Medium | Low | Expo SDK covers most needs, can eject if necessary |
| EAS Build downtime | Medium | Low | Cache builds locally, have Fastlane backup |
| Team learning curve | Low | Medium | Allocate 1-2 weeks for learning, pair programming |

## Success Metrics

### Technical Metrics

- [ ] App builds successfully on iOS and Android
- [ ] < 5 crashes per 1000 sessions
- [ ] > 80% code sharing with web
- [ ] All E2E tests passing

### Business Metrics

- [ ] > 10,000 downloads in first month
- [ ] > 30% D1 retention
- [ ] > 4.0 star rating on app stores
- [ ] < 5% uninstall rate

## Alternatives Rejected

### Why not PWA only?

- Push notifications limited on iOS
- No home screen icon prominence
- Limited offline capabilities
- Cannot access native APIs (biometric auth, etc.)

### Why not Capacitor?

- Less mature than Expo
- Smaller ecosystem
- More complex setup
- Web-first approach doesn't leverage native components

### Why not Ionic?

- Web components, not native components
- Performance concerns
- Less modern than React Native

## References

- [React Native Documentation](https://reactnative.dev/)
- [Expo Documentation](https://docs.expo.dev/)
- [EAS Build Documentation](https://docs.expo.dev/build/introduction/)
- [React Navigation Documentation](https://reactnavigation.org/)
- [NativeWind Documentation](https://www.nativewind.dev/)
- [TanStack Query Documentation](https://tanstack.com/query/latest)

## Conclusion

React Native with Expo is the clear choice for Clipper's mobile apps. It maximizes code sharing, leverages our team's existing expertise, provides excellent developer experience, and delivers native-quality performance. The Expo ecosystem provides all the tooling we need for building, deploying, and maintaining production mobile apps.

**Decision Date:** 2025-11-02
**Review Date:** 2026-05-02 (6 months)

---

**Approved By:** Clipper Engineering Team
