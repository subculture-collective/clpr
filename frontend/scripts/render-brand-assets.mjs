// Renders clpr's raster brand assets from the SVG sources in public/ and the
// self-hosted fonts in src/assets/fonts. Run after changing the logo, icon, or
// social card: `node scripts/render-brand-assets.mjs`. Requires Playwright's
// Chromium (installed for the e2e suite).
import { chromium } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import path from 'node:path';

const root = path.resolve(import.meta.dirname, '..');
const publicDir = path.join(root, 'public');
const fontDir = path.join(root, 'src/assets/fonts');

const icon = await readFile(path.join(publicDir, 'icons/icon.svg'), 'utf8');
const maskable = await readFile(path.join(publicDir, 'icons/icon-maskable.svg'), 'utf8');
const favicon = await readFile(path.join(publicDir, 'favicon.svg'), 'utf8');
const logo = await readFile(path.join(publicDir, 'clpr-logo.svg'), 'utf8');

// setContent pages cannot read file:// URLs, so fonts are inlined as data URLs.
const fontFace = async (family, weight, file) => {
    const data = (await readFile(path.join(fontDir, file))).toString('base64');
    return `@font-face { font-family: '${family}'; font-weight: ${weight}; src: url(data:font/woff2;base64,${data}) format('woff2'); }`;
};
const fonts = (await Promise.all([
    fontFace('Barlow', 400, 'barlow-400.woff2'),
    fontFace('Barlow Condensed', 800, 'barlow-condensed-800.woff2'),
    fontFace('IBM Plex Mono', 500, 'ibm-plex-mono-500.woff2'),
])).join('\n');

const sized = (svg, width, height) => svg.replace('<svg ', `<svg width="${width}" height="${height}" `);

/** A wordmark on ink with an optional headline, used for the social card and banners. */
function card({ width, height, logoHeight, headline, meta }) {
    return `<!doctype html><html><head><style>
        ${fonts}
        * { margin: 0; box-sizing: border-box; }
        body { width: ${width}px; height: ${height}px; background: #0E0C13; color: #EEEDF7; position: relative; overflow: hidden; }
        .rule { position: absolute; inset: 0 0 auto 0; height: ${Math.round(height * 0.012)}px; background: #8C5CFF; }
        .wrap { position: absolute; inset: 0; display: flex; flex-direction: column; justify-content: center; padding: 0 ${Math.round(width * 0.07)}px; gap: ${Math.round(height * 0.05)}px; }
        .logo svg { height: ${logoHeight}px; width: auto; display: block; }
        h1 { font: 800 ${Math.round(height * 0.1)}px/0.95 'Barlow Condensed'; text-transform: uppercase; letter-spacing: 0.01em; max-width: 20ch; }
        .meta { font: 500 ${Math.round(height * 0.032)}px 'IBM Plex Mono'; letter-spacing: 0.12em; text-transform: uppercase; color: #8F8A9C; }
        .burn { position: absolute; right: ${Math.round(width * 0.04)}px; bottom: ${Math.round(height * 0.06)}px; font: 500 ${Math.round(height * 0.03)}px 'IBM Plex Mono'; background: #000000b8; color: #fff; padding: 4px 10px; }
    </style></head><body>
        <div class="rule"></div>
        <div class="wrap">
            <div class="logo">${logo}</div>
            ${headline ? `<h1>${headline}</h1>` : ''}
            ${meta ? `<p class="meta">${meta}</p>` : ''}
        </div>
        ${headline ? '<span class="burn">00:00:20:14</span>' : ''}
    </body></html>`;
}

const browser = await chromium.launch();

async function shoot(html, width, height, out) {
    const page = await browser.newPage({ viewport: { width, height }, deviceScaleFactor: 1 });
    await page.setContent(html);
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({ path: path.join(publicDir, out) });
    await page.close();
}

const svgPage = svg => `<html><body style="margin:0">${svg}</body></html>`;

for (const size of [72, 96, 128, 144, 152, 192, 384, 512]) {
    await shoot(svgPage(sized(icon, size, size)), size, size, `icons/icon-${size}x${size}.png`);
}
for (const size of [192, 512]) {
    await shoot(svgPage(sized(maskable, size, size)), size, size, `icons/icon-${size}x${size}-maskable.png`);
    await shoot(svgPage(sized(icon, size, size)), size, size, `favicon_io/android-chrome-${size}x${size}.png`);
}
await shoot(svgPage(sized(icon, 180, 180)), 180, 180, 'favicon_io/apple-touch-icon.png');
for (const size of [16, 32]) {
    await shoot(svgPage(sized(favicon, size, size)), size, size, `favicon_io/favicon-${size}x${size}.png`);
}

await shoot(
    card({
        width: 1200,
        height: 630,
        logoHeight: 120,
        headline: 'The moments shaping live culture',
        meta: 'clpr.tv · Twitch clips by creator, topic and tag',
    }),
    1200,
    630,
    'social-card.png',
);
await shoot(card({ width: 500, height: 500, logoHeight: 110 }), 500, 500, 'clpr-500px.png');
await shoot(card({ width: 1024, height: 1024, logoHeight: 230 }), 1024, 1024, 'clpr-1021px.png');
await shoot(card({ width: 500, height: 294, logoHeight: 90 }), 500, 294, 'clpr-banner-500px.png');
await shoot(card({ width: 1021, height: 601, logoHeight: 180 }), 1021, 601, 'clpr-banner-1021px.png');

await browser.close();
console.log('Brand assets rendered. favicon.ico is assembled separately from the favicon PNGs.');
