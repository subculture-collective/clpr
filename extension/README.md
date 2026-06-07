# Clipper Browser Extension

Share Twitch clips to [Clipper](https://clpr.gg) with one click.

## Features

- **Detects Twitch clips** automatically when you visit a clip page
- **Context menu** – right-click on any Twitch clip page to see "Share to Clipper"
- **Popup UI** with:
  - Clip thumbnail preview
  - Editable title (pre-filled from Twitch metadata)
  - Optional description
  - Multi-select tag picker
  - NSFW toggle
  - One-click submit
- **OAuth via Clipper** – authenticates using your existing Clipper account
- **Desktop notifications** on successful submission

## Supported Browsers

| Browser | Minimum version |
|---------|----------------|
| Chrome / Chromium | 99+ |
| Firefox | 109+ |
| Edge | 99+ |

## Development

### Prerequisites

- Node.js 18+
- npm 9+

### Install dependencies

```bash
cd extension
npm install
```

### Build

```bash
npm run build
```

The built extension will be in `extension/dist/`.

### Watch mode

```bash
npm run build:watch
```

### Type-check only

```bash
npm run typecheck
```

## Loading the extension

### Chrome / Edge

1. Navigate to `chrome://extensions`
2. Enable **Developer mode** (top-right toggle)
3. Click **Load unpacked**
4. Select the `extension/dist/` folder

### Firefox

1. Navigate to `about:debugging#/runtime/this-firefox`
2. Click **Load Temporary Add-on…**
3. Select any file inside `extension/dist/`

## Configuration

By default the extension points to the production Clipper API (`https://api.clpr.gg/api/v1`).

To point at a local backend, open the browser's extension storage inspector and set:

```json
{
  "apiBaseUrl": "http://localhost:8080/api/v1",
  "frontendUrl": "http://localhost:3000"
}
```

Or run this snippet in the background service worker's DevTools console:

```js
chrome.storage.local.set({
  apiBaseUrl: 'http://localhost:8080/api/v1',
  frontendUrl: 'http://localhost:3000',
});
```

## Authentication

The extension reads the `access_token` JWT cookie issued by the Clipper backend and sends it as a `Bearer` token.  To authenticate:

1. Click the extension icon
2. Click **Login with Twitch** – this opens `clpr.gg` in a new tab
3. Complete the Twitch OAuth flow on the website
4. Return to the extension popup – you should now be logged in

## Architecture

```
extension/
├── manifest.json           # Manifest V3
├── build.js                # esbuild orchestration script
├── package.json
├── tsconfig.json
├── scripts/
│   └── generate-icons.js   # Generates PNG icons from pure Node.js
├── src/
│   ├── background.ts       # Service worker – context menu & tab tracking
│   ├── content.ts          # Content script – Twitch SPA URL detection
│   ├── types.ts            # Shared TypeScript interfaces
│   ├── lib/
│   │   ├── api.ts          # Clipper API client
│   │   ├── auth.ts         # Auth token retrieval & verification
│   │   └── twitch.ts       # Twitch URL detection utilities
│   └── popup/
│       ├── popup.html      # Popup markup
│       ├── popup.ts        # Popup logic
│       └── popup.css       # Popup styles
└── icons/                  # Generated PNG icons
```

## Publishing

### Chrome Web Store

1. Run `npm run build`
2. Zip the `dist/` folder: `cd dist && zip -r ../clpr-extension.zip .`
3. Upload to the [Chrome Web Store Developer Dashboard](https://chrome.google.com/webstore/devconsole)

### Firefox Add-ons (AMO)

1. Run `npm run build`
2. Zip the `dist/` folder
3. Submit to [addons.mozilla.org](https://addons.mozilla.org/developers/)

> **Note:** Firefox requires the `browser_specific_settings.gecko.id` field in `manifest.json` (already included).
