// Run the published payload, including every migration, without an API or database server.
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { pathToFileURL } from 'node:url'
import { gunzipSync } from 'node:zlib'
import { createHash } from 'node:crypto'
import { PGlite } from '@electric-sql/pglite'
import { vector } from '@electric-sql/pglite-pgvector'
import { pg_trgm } from '@electric-sql/pglite/contrib/pg_trgm'
import { ltree } from '@electric-sql/pglite/contrib/ltree'
import { pgcrypto } from '@electric-sql/pglite/contrib/pgcrypto'

const dir = resolve(process.argv[2] ?? 'dist/search-wasm')
const manifest = JSON.parse(await readFile(`${dir}/manifest.json`, 'utf8'))
for (const [name, expected] of Object.entries(manifest.files)) {
  const data = await readFile(`${dir}/${name}`)
  assert.equal(data.length, expected.bytes)
  assert.equal(createHash('sha256').update(data).digest('hex'), expected.sha256)
}
const db = await PGlite.create({ extensions: { vector, pg_trgm, ltree, pgcrypto } })
await db.exec('CREATE ROLE clpr')
const migrations = JSON.parse(await readFile(`${dir}/migrations.json`, 'utf8'))
assert.equal(migrations.length, manifest.migrationCount)
for (const { name, sql } of migrations) {
  try { await db.exec(sql) } catch (error) { throw new Error(`Migration ${name}: ${error.message}`) }
}
await db.exec(`INSERT INTO clips(twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,created_at,vote_score)
  VALUES ('wasm-1','https://example.invalid/1','https://example.invalid/1','speedrun world record attempt','runner_a','runner_a','2026-11-01',41),
  ('wasm-2','https://example.invalid/2','https://example.invalid/2','chat reacts to the world record','streamer_b','streamer_b','2026-11-02',33)`)
globalThis.pglite = db
await import(pathToFileURL(`${dir}/wasm_exec.js`))
const go = new Go()
const raw = gunzipSync(await readFile(`${dir}/search.wasm.gz`))
assert.equal(raw.length, manifest.wasm.bytes)
assert.equal(createHash('sha256').update(raw).digest('hex'), manifest.wasm.sha256)
const { instance } = await WebAssembly.instantiate(raw, go.importObject)
void go.run(instance).catch(error => { console.error(error); process.exit(1) })
assert.equal(globalThis.goReady, true)
// After the executor has been released, repeated asynchronous calls must still resolve correctly.
for (let i = 0; i < 30; i++) {
  const response = await goHandle('GET', '/api/v1/search?q=world+record&type=clips')
  assert.equal(response.status, 200)
  const body = JSON.parse(response.body)
  assert.deepEqual(body.results.clips.map(c => c.title), ['speedrun world record attempt', 'chat reacts to the world record'])
}
assert.equal((await goHandle('GET', '/api/v1/search')).status, 400)
assert.equal((await goHandle('GET', '/api/v1/search?q=world&min_votes=x')).status, 400)
const failure = await goHandle('GET', '/demo/fault/hybrid?q=world')
assert.equal(failure.status, 503)
assert.equal(failure.headers['Retry-After'], '60')
console.log(`PASS: ${migrations.length} migrations, real pgx searches, repeated requests, validation and hybrid failure`)
process.exit(0) // The Go program deliberately waits for more requests; this process owns it.
