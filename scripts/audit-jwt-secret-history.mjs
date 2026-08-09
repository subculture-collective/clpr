#!/usr/bin/env node

import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = new URL('..', import.meta.url).pathname;
const markerLocations = [
  ['32086c6b4b7e0a62c5a0254745d010a5c6554ea1', 'backend/docs/authentication.md'],
  ['a5063080ec144dff329f9d422bae55f384b91469', 'backend/k8s/secret-backend.yaml'],
  ['82ccc39ff5d516d96ade425e7ccea0af6a8fc545', 'infrastructure/k8s/external-secrets/README.md'],
];

const git = (...args) => execFileSync('git', args, {
  cwd: repoRoot,
  encoding: 'utf8',
  maxBuffer: 64 * 1024 * 1024,
});

function auditPemMarkers() {
for (const [commit, markerPath] of markerLocations) {
  const content = git('show', `${commit}:${markerPath}`);
  const marker = /-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----/g;
  let match;
  let markers = 0;
  while ((match = marker.exec(content)) !== null) {
    markers += 1;
    const tail = content.slice(match.index + match[0].length, match.index + match[0].length + 8192);
    const end = tail.search(/-----END (?:[A-Z0-9 ]+ )?PRIVATE KEY-----/);
    const body = end >= 0 ? tail.slice(0, end) : tail.slice(0, 1024);
    if (/^[A-Za-z0-9+/]{64,}={0,2}$/m.test(body)) {
      throw new Error(`private-key payload found at reviewed history location ${commit}:${markerPath}`);
    }
  }
  if (markers === 0) {
    throw new Error(`expected reviewed private-key marker is missing at ${commit}:${markerPath}`);
  }
}
}

// Review every revision that changed a JWT signing configuration reference.
// Long literal right-hand values are rejected unless visibly externalized or a
// documented placeholder. Values are never printed.
const assignment = /JWT_(?:SECRET|PRIVATE_KEY(?:_B64)?|SIGNING_KEY(?:_B64)?)\s*[:=]\s*["']?([^\s,"'#]+)/gi;
const safeValue = /^(?:\$|\{\{|<|your|replace|change|example|test|fake|placeholder|secretkeyref|os\.getenv|getenv)/i;

export function auditJwtPatch(patch, commit = 'fixture') {
  let referencesReviewed = 0;
  for (const line of patch.split('\n')) {
    if (!line.startsWith('+') || line.startsWith('+++')) continue;
    assignment.lastIndex = 0;
    let match;
    while ((match = assignment.exec(line.slice(1))) !== null) {
      referencesReviewed += 1;
      const value = match[1].replace(/["'}\])]+$/, '');
      if (value.length >= 24 && !safeValue.test(value)) {
        throw new Error(`possible literal JWT signing material introduced by commit ${commit}`);
      }
    }
  }
  return referencesReviewed;
}

function main() {
  auditPemMarkers();
  const commits = git(
    'log', '--all', '--format=%H',
    '-G', 'JWT_(SECRET|PRIVATE_KEY(_B64)?|SIGNING_KEY(_B64)?)', '--',
  ).trim().split('\n').filter(Boolean);
  let referencesReviewed = 0;
  for (const commit of commits) {
    const patch = git('show', '--format=', '--no-ext-diff', '--unified=0', commit);
    referencesReviewed += auditJwtPatch(patch, commit);
  }
  console.log(`JWT/PEM history audit passed (${referencesReviewed} signing-config assignments and ${markerLocations.length} marker-only fixtures reviewed).`);
}

if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) main();
