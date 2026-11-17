import assert from 'node:assert/strict';
import fs from 'node:fs';
import { test } from 'node:test';
import { checkContracts, loadContracts, responseValidator } from '../verify-mobile-contracts.mjs';

const { document, manifest } = loadContracts();
function sample(name) {
  const fixture = manifest.find(value => value.file === `${name}.json`);
  assert.ok(fixture, `Missing ${name}`);
  const body = JSON.parse(fs.readFileSync(new URL(`../../docs/openapi/mobile/${fixture.file}`, import.meta.url), 'utf8'));
  return { body, validate: responseValidator(document, fixture) };
}

test('all exported native fixtures match current routes and schemas', () => {
  assert.ok(checkContracts() > 0);
});

test('clip consumers require the wire envelope, typed counters, UUIDs and dates', () => {
  const { body, validate } = sample('clips-populated');
  for (const mutate of [
    value => delete value.data,
    value => { value.data[0].view_count = '42'; },
    value => { value.data[0].id = 'not-a-uuid'; },
    value => { value.data[0].created_at = 'yesterday'; },
  ]) {
    const invalid = structuredClone(body);
    mutate(invalid);
    assert.equal(validate(invalid), false);
  }
  body.future_field = true;
  body.data[0].future_metadata = { accepted: true };
  assert.equal(validate(body), true, 'response additions remain compatible');
});

test('search collections cannot become null or disappear', () => {
  const { body, validate } = sample('search-empty');
  for (const key of ['clips', 'creators', 'games', 'twitch_categories', 'tags']) {
    const invalid = structuredClone(body);
    invalid.results[key] = null;
    assert.equal(validate(invalid), false, key);
    delete invalid.results[key];
    assert.equal(validate(invalid), false, key);
  }
});

test('empty taxonomy and category pages preserve nullable repository results', () => {
  for (const [empty, populated, key] of [
    ['tags-empty', 'tags-populated', 'tags'],
    ['categories-empty', 'categories-populated', 'games'],
  ]) {
    const { body, validate } = sample(empty);
    assert.equal(body[key], null, `${empty} must reflect the nil repository slice`);
    assert.equal(validate(body), true);
    body[key] = [];
    assert.equal(validate(body), true);
    body[key] = sample(populated).body[key];
    assert.equal(validate(body), true);
    body[key] = {};
    assert.equal(validate(body), false);
    delete body[key];
    assert.equal(validate(body), false, `${key} must still be present`);
  }
});

test('countdown cursor is required, accepts null, and rejects an invalid date', () => {
  const { body, validate } = sample('countdowns-empty');
  assert.equal(body.meta.next_before, null);
  assert.equal(validate(body), true);
  body.meta.next_before = '2026-99-99';
  assert.equal(validate(body), false);
  delete body.meta.next_before;
  assert.equal(validate(body), false);
});

test('string and structured errors validate without accepting arbitrary objects', () => {
  for (const name of ['search-invalid', 'clip-removed', 'search-rate-limited']) {
    const { body, validate } = sample(name);
    assert.equal(validate(body), true, name);
    body.error = { message: 'missing code' };
    assert.equal(validate(body), false, name);
  }
});

test('empty playlists omit optional clips rather than promising an array', () => {
  const { body, validate } = sample('playlist-empty');
  assert.equal(Object.hasOwn(body.data, 'clips'), false);
  assert.equal(validate(body), true);
});
