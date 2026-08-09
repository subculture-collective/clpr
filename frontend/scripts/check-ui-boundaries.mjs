import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';

async function sourceFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  return (await Promise.all(entries.map(async (entry) => {
    const target = path.join(directory, entry.name);
    if (entry.isDirectory()) return sourceFiles(target);
    return /\.(?:ts|tsx)$/.test(target) && !/\.test\./.test(target) ? [target] : [];
  }))).flat();
}

const files = await sourceFiles('src');
const allowedCustomDialogs = new Set([
  'src/components/consent/ConsentBanner.tsx',
  'src/components/ui/EmojiPicker.tsx',
  'src/components/ui/Modal.tsx',
  'src/components/video/TheatreMode.tsx',
]);
const allowedImmersiveOverlays = new Set([
  'src/components/playlist/PlaylistTheatreMode.tsx',
  'src/pages/ChatPage.tsx',
  'src/pages/PlaylistTheatrePage.tsx',
  'src/pages/QueueTheatrePage.tsx',
  'src/pages/StreamerClipRoomPage.tsx',
]);

const failures = [];
for (const file of files) {
  const source = await readFile(file, 'utf8');
  if (/role=["'](?:alert)?dialog["']/.test(source) && !allowedCustomDialogs.has(file)) {
    failures.push(`${file}: use the shared Modal for transient dialogs`);
  }
  if (/fixed inset-0[^"']*bg-black/.test(source) &&
      !allowedCustomDialogs.has(file) && !allowedImmersiveOverlays.has(file)) {
    failures.push(`${file}: unapproved custom full-screen overlay`);
  }
  if (/\b(?:function|const)\s+formatNumber\b/.test(source)) {
    failures.push(`${file}: use formatCompactNumber or locale formatting utilities`);
  }
  if (/<Link\b[^>]*>\s*<Button\b/s.test(source)) {
    failures.push(`${file}: compose links with <Button asChild> instead of nesting controls`);
  }
}

if (failures.length > 0) {
  console.error(`UI boundary violations:\n${failures.map((failure) => `- ${failure}`).join('\n')}`);
  process.exit(1);
}
console.log(`UI boundaries pass across ${files.length} source modules.`);
