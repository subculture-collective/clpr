---
title: "Native viewer build and beta prerequisites"
summary: "Initial native toolchain decisions, available build capabilities and remaining beta gates."
tags: ["product", "mobile", "planning"]
area: "product"
status: "draft"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-11-17
---

# Native viewer build and beta prerequisites

This is the build-readiness decision for
[issue #1441](https://github.com/subculture-collective/clpr/issues/1441), within
the [native viewer roadmap](https://github.com/subculture-collective/clpr/issues/1436).
It selects a starting build approach and records what still needs qualification.
No native workspace, signing profile, app registration or distribution channel
has been created by this investigation.

## Foundation decisions

| Area | iOS | Android |
| --- | --- | --- |
| App | Swift/SwiftUI in `apps/ios` | Kotlin/Jetpack Compose in `apps/android` |
| Initial minimum OS | iOS 18 | Android 8 / API 26 |
| Project/build | XcodeGen specification, generated Xcode project and shared schemes | Committed Gradle wrapper, Kotlin build scripts and version catalog |
| Networking/JSON | URLSession, structured concurrency and Codable | OkHttp, coroutines and Kotlin serialization |
| First tests | Swift unit tests and XCUITest against a selected simulator | JVM unit tests and Compose instrumentation tests against API 35 |
| First private channel | TestFlight after account/app/signing and tester access are confirmed | Owner-controlled signed APK distribution to a private tester list |

The initial OS targets bound the first compatibility matrix. They are product
choices for the foundation PRs, not claims that either minimum has passed tests.
Add the minimum-version simulator/emulator to that matrix when it is available;
record its absence if the installed runtime cannot exercise it. Do not download
a large runtime without checking the host's current capacity.

XcodeGen is already installed on the owner's Mac. Keeping its project
specification and shared schemes in source makes the native target reproducible
without hand-maintained project-file changes. URLSession avoids a second iOS
networking dependency. OkHttp provides the Android transport while coroutine and
serialization code remain explicit. Pin library/plugin/wrapper versions in each
foundation PR after proving the selected combination builds; no untested version
combination is represented as working here.

## Available capabilities

Read-only inventory on November 17, 2026 UTC established:

- The owner designated their Mac as the iOS build host. It reports macOS 26.6.2,
  Xcode 27.2, Swift 6.4.2, iOS/iOS Simulator SDK 27.2, XcodeGen 2.46.0 and
  Task 3.54.0. A physical iPhone 16 Plus is paired and reported available; an
  iPhone simulator is installed. Neither was driven during this investigation.
- The Mac reports one valid signing identity. This does not establish its
  ownership, team, entitlements, provisioning, app-specific suitability or
  Developer Program/App Store Connect access. No identity was used or exported.
- The Linux development host has OpenJDK 21, Android SDK platforms 35 and 36,
  build tools 34 and 35, command-line tools and an Android 15/API 35 emulator.
  A standalone Gradle installation is unnecessary once the foundation commits
  its qualified wrapper. No physical Android phone was identified.
- This thread's T3 Device panel currently exposes the Android emulator. The Mac
  still needs a supported T3 device-host route before interactive iOS work can
  be shown and driven through the device tools. SSH inventory is not device UI
  qualification.

The Mac had about 7.5 GiB free at inspection. Recheck space before downloads or
builds; this investigation installed nothing and removed no host data. Hostnames,
device identifiers, signing identities and credential locations belong in the
private operations receipt, not this public repository.

## Verification tasks to implement

The foundation issues introduce these commands; they do not exist yet:

| Task | Required behavior |
| --- | --- |
| `task mobile:ios:verify` | Generate the Xcode project using the pinned generator, build for an available simulator and run the unit-test target; fail on missing tools or tests |
| `task mobile:ios:ui` | Run the XCUITest target on an explicitly selected destination and retain its result bundle; distinguish simulator from physical-phone evidence |
| `task mobile:android:verify` | Use the committed wrapper to assemble the debug app, run JVM tests and Android lint; fail on missing SDK/plugin versions or tests |
| `task mobile:android:ui` | Run Compose instrumentation tests on an explicitly selected emulator/device and retain reports |

Both platform jobs also run `task mobile:contracts`, provided by
[issue #1444](https://github.com/subculture-collective/clpr/issues/1444).
Use deterministic fixtures for decoder/error tests and a real staging feed for
the foundation's networking acceptance.

Add independent path-aware GitHub workflows. Prefer disposable GitHub-hosted
macOS for unsigned iOS simulator CI and Linux for Android CI. The iOS foundation
must prove that its pinned Xcode version exists on the selected runner; the local
Mac inventory alone does not establish hosted-runner availability. Do not expose
the owner's Mac or signing credentials to untrusted pull-request jobs. Release
signing belongs in an explicitly authorized trusted workflow.

Interactive device work uses T3 `device_list`, `device_open` and the command
returned for that device. Project test runners can perform their normal test
actions; manual launches, links, input and screenshots use the bound device
command. Do not reuse or close another thread's device session.

## Signing and distribution gates

The project owner controls release approval. Ownership of the Mac is confirmed;
ownership/access for the following accounts and assets remains unverified:

| Gate | Evidence needed before use |
| --- | --- |
| Apple account and team | Authorized signing owner, Developer Program/App Store Connect access and the public team identifier |
| iOS app identity | Owner-approved bundle identifier, registered app, required capabilities and matching development/distribution profiles |
| Android identity | Owner-approved application ID, app-specific signing ownership and the public certificate fingerprint |
| Verified links | The approved public identities, callback/content paths and staging association-file readback from issue #1446 |
| Private channels | TestFlight app/tester setup and any required beta review; an authorized private APK destination, tester list and delivery/revocation instructions |
| Device matrix | This thread's supported iOS device route, an available physical Android phone, and dated pass/fail results for the required journeys |
| Privacy/account support | The selected channels' requirements-to-evidence assessment in issue #1443 |

Do not infer account enrollment or distribution permission from an installed
certificate. Public app/team IDs and certificate fingerprints can be recorded in
the association decision after approval; private keys and tokens stay in the
existing secret store. No paid account, signing credential or app registration
is provisioned by this plan.

Unsigned discovery foundations can proceed after this decision and the contract
pack merge. App-specific associations, authenticated qualification and private
distribution remain gated by their own evidence. Public store publication is a
later release decision.

## Sources

- [XcodeGen project specifications](https://github.com/yonaskolb/XcodeGen/blob/master/Docs/ProjectSpec.md)
- [URLSession](https://developer.apple.com/documentation/foundation/urlsession)
- [OkHttp](https://square.github.io/okhttp/)
- [AndroidX minimum SDK policy](https://developer.android.com/jetpack/androidx/versions)
- [TestFlight overview](https://developer.apple.com/help/app-store-connect/test-a-beta-version/testflight-overview/)
- [Android app signing](https://developer.android.com/studio/publish/app-signing)
