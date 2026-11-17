# Browser search artifact

`cmd/searchwasm` runs the same `internal/search` handler and repository as the API. Its
PostgreSQL connection uses pgx over a PGlite wire-protocol bridge. The demo uses the
PostgreSQL search path; its hybrid fault route uses a deliberately refusing provider.

The Search WebAssembly workflow builds and tests pull requests. On `main`, it also
publishes a `search-wasm-<full commit>` GitHub release. Consumers pin that exact release
and serve the files from their own origin. The portfolio never builds Clpr itself.

Each release contains `search.wasm.gz`, `wasm_exec.js` from the matching Go toolchain,
`GO-LICENSE`, every up migration in ordered `migrations.json`, and `manifest.json`.
The manifest records the source commit and date, Go version, build tags, migration
count, compressed and raw Wasm sizes, file hashes, and the CI run that produced it.
The build fails above 6,000,000 gzip bytes. Release assets are created once; rerunning
publication for an existing tag fails instead of replacing them.

Local reproduction requires Go 1.26.11, Python 3, gzip, and Node 24. From the repository root:

```sh
python3 scripts/build-search-wasm.py
npm ci --prefix scripts/search-wasm-test
node scripts/search-wasm-test/smoke.mjs dist/search-wasm
```

Local manifests say `builtInClprCI: false` and record whether source changes are
uncommitted. CI refuses a dirty source tree. The smoke test verifies the file hashes,
applies every bundled migration with vector, pg_trgm, ltree and pgcrypto, then checks
real search responses, repeated asynchronous requests, validation and hybrid failure.
This is artifact validation, not proof of a production deployment or complete
native-versus-browser parity.
