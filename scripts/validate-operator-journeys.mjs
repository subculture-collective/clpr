#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';

const manifestPath = process.argv[2] || 'config/release/operator-journeys.json';
const manifest = JSON.parse(fs.readFileSync(path.resolve(manifestPath), 'utf8'));
const failures = [];

if (manifest.schema_version !== 1) failures.push('schema_version must equal 1');
if (manifest.target_environment !== 'disposable-staging') {
  failures.push('target_environment must be disposable-staging');
}

const expectedRoles = new Map([
  ['user', 'user'],
  ['moderator', 'moderator'],
  ['administrator', 'admin'],
]);
for (const role of manifest.roles || []) {
  if (expectedRoles.get(role.name) !== role.database_role) {
    failures.push(`invalid database role mapping for ${role.name || '<unnamed>'}`);
  }
  if (!/^[A-Z][A-Z0-9_]+$/.test(role.storage_state_env || '')) {
    failures.push(`invalid storage-state environment name for ${role.name || '<unnamed>'}`);
  }
  expectedRoles.delete(role.name);
}
for (const role of expectedRoles.keys()) failures.push(`missing role: ${role}`);

const projects = new Set(manifest.browser_projects || []);
for (const project of ['real-chromium', 'real-firefox', 'real-webkit']) {
  if (!projects.has(project)) failures.push(`missing browser project: ${project}`);
}

const journeys = new Set((manifest.journeys || []).filter(item => item.required).map(item => item.id));
for (const journey of [
  'oauth-login',
  'public-search-empty-and-populated',
  'clip-playback',
  'submission',
  'settings-and-privacy',
  'moderation-allow-and-deny',
  'administration-boundary',
  'logout-and-session-expiry',
]) {
  if (!journeys.has(journey)) failures.push(`missing required journey: ${journey}`);
}

for (const name of manifest.required_secret_names || []) {
  if (!/^[A-Z][A-Z0-9_]+$/.test(name)) failures.push(`invalid secret name: ${name}`);
}

if (failures.length) {
  console.error(`Operator journey manifest failed:\n- ${failures.join('\n- ')}`);
  process.exit(1);
}
console.log(`Operator journey manifest passed: ${journeys.size} journeys across ${projects.size} browsers`);
