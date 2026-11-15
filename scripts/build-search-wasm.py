#!/usr/bin/env python3
"""Build the versioned browser search bundle; run from any directory."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[1]


def output(*args: str) -> str:
    return subprocess.check_output(args, cwd=ROOT, text=True).strip()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", type=Path, default=ROOT / "dist/search-wasm")
    args = parser.parse_args()
    dest = args.out.resolve()
    dest.mkdir(parents=True, exist_ok=True)
    go_version = output("go", "env", "GOVERSION")
    # wasm_exec.js and the compiler must come from the same toolchain.
    expected = next(line.split()[1] for line in (ROOT / "backend/go.mod").read_text().splitlines()
                    if line.startswith("toolchain "))
    if go_version != expected:
        raise SystemExit(f"Use {expected}, found {go_version}")
    dirty = bool(output("git", "status", "--porcelain", "--untracked-files=normal"))
    in_ci = os.environ.get("GITHUB_ACTIONS") == "true"
    if in_ci and dirty:
        raise SystemExit("Refusing to publish a dirty source tree")
    wasm = dest / "search.wasm"
    subprocess.run(["go", "build", "-mod=readonly", "-tags", "nomsgpack", "-trimpath",
                    "-ldflags=-s -w", "-o", str(wasm), "./cmd/searchwasm"],
                   cwd=ROOT / "backend", env=dict(os.environ, GOOS="js", GOARCH="wasm"), check=True)
    raw = wasm.read_bytes()
    packed = subprocess.check_output(["gzip", "-n", "-9", "-c", str(wasm)])
    if len(packed) > 6_000_000:
        raise SystemExit(f"Search wasm exceeds 6 MB gzip: {len(packed)} bytes")
    (dest / "search.wasm.gz").write_bytes(packed)
    wasm.unlink()
    migrations = [{"name": p.name, "sql": p.read_text()} for p in sorted((ROOT / "backend/migrations").glob("*.up.sql"))]
    (dest / "migrations.json").write_text(json.dumps(migrations) + "\n")
    goroot = Path(output("go", "env", "GOROOT"))
    shutil.copyfile(goroot / "lib/wasm/wasm_exec.js", dest / "wasm_exec.js")
    shutil.copyfile(goroot / "LICENSE", dest / "GO-LICENSE")
    files = {}
    for name in ["search.wasm.gz", "migrations.json", "wasm_exec.js", "GO-LICENSE"]:
        data = (dest / name).read_bytes()
        files[name] = {"bytes": len(data), "sha256": hashlib.sha256(data).hexdigest()}
    manifest = {
        "schemaVersion": 1,
        "repository": "subculture-collective/clpr",
        "commit": output("git", "rev-parse", "HEAD"),
        "date": output("git", "show", "-s", "--format=%cI", "HEAD"),
        "goVersion": go_version,
        "buildTags": ["nomsgpack"],
        "builtInClprCI": in_ci and os.environ.get("GITHUB_REPOSITORY") == "subculture-collective/clpr",
        "sourceDirty": dirty,
        "ciRun": (f"https://github.com/subculture-collective/clpr/actions/runs/{os.environ['GITHUB_RUN_ID']}" if in_ci else None),
        "wasm": {"bytes": len(raw), "gzipBytes": len(packed), "sha256": hashlib.sha256(raw).hexdigest()},
        "migrationCount": len(migrations),
        "files": files,
    }
    (dest / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"{manifest['commit']}: {len(packed):,} bytes gzip, {len(migrations)} migrations -> {dest}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
