#!/usr/bin/env python3
"""Bind the signed candidate to a frontend-only, host-owned deployment profile."""
import argparse
import base64
import datetime
import fcntl
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import sys
import tarfile
import tempfile
import traceback

if not __debug__:
    raise RuntimeError('Frontend delivery requires enabled validation checks')

REPOSITORY = 'subculture-collective/clpr'
SERVICE = 'clpr-frontend'
DIGEST = re.compile(r'sha256:[0-9a-f]{64}')
REVISION = re.compile(r'[0-9a-f]{40}')
CANDIDATE_CONTEXT = 'Immutable production candidate / Build, scan, sign, and publish immutable candidate (push)'
QUALITY_NAMES = ['Secret history gate', 'Documentation and OpenAPI gate',
                 'Browser gate (Chromium, Firefox, WebKit)', 'Image scan and SBOM gate']
SCHEMA_SQL = """BEGIN READ ONLY; SET LOCAL statement_timeout='10s';
SELECT version,dirty FROM schema_migrations;
SELECT coalesce(json_agg(item ORDER BY item)::text,'[]') FROM (
 SELECT 'column:'||table_schema||'.'||table_name||'.'||column_name||':'||
        data_type||':'||udt_name||':'||is_nullable||':'||coalesce(column_default,'') AS item
 FROM information_schema.columns WHERE table_schema='public'
 UNION ALL SELECT 'constraint:'||c.conrelid::regclass||':'||c.conname||':'||pg_get_constraintdef(c.oid)
 FROM pg_constraint c JOIN pg_namespace n ON n.oid=c.connamespace WHERE n.nspname='public'
 UNION ALL SELECT 'index:'||schemaname||'.'||tablename||'.'||indexname||':'||indexdef
 FROM pg_indexes WHERE schemaname='public'
 UNION ALL SELECT 'enum:'||n.nspname||'.'||t.typname||':'||e.enumsortorder||':'||e.enumlabel
 FROM pg_enum e JOIN pg_type t ON t.oid=e.enumtypid JOIN pg_namespace n ON n.oid=t.typnamespace
 WHERE n.nspname='public'
 UNION ALL SELECT 'view:'||schemaname||'.'||viewname||':'||definition FROM pg_views WHERE schemaname='public'
) structure; COMMIT;"""


def sha(data):
    return hashlib.sha256(data).hexdigest()


def run(args, timeout=60):
    result = subprocess.run(args, capture_output=True, timeout=timeout)
    if result.returncode:
        raise ValueError('Required command failed; inspect private host evidence')
    return result.stdout


def config_digest(image):
    # Docker/containerd manifest IDs differ after the receiver's classic import.
    # The saved config digest binds the same layer diff IDs on either engine.
    with tempfile.TemporaryDirectory() as folder:
        archive_path = str(Path(folder) / 'image.tar')
        run(['docker', 'save', '-o', archive_path, image])
        with tarfile.open(archive_path) as archive:
            manifest_member = archive.getmember('manifest.json')
            assert manifest_member.isfile() and manifest_member.size <= 1024 * 1024
            manifest = json.load(archive.extractfile(manifest_member))
            assert len(manifest) == 1
            member = archive.getmember(manifest[0]['Config'])
            assert member.isfile() and member.size <= 4 * 1024 * 1024
            return 'sha256:' + sha(archive.extractfile(member).read())


def validate_statement(statement, revision):
    assert REVISION.fullmatch(revision)
    assert set(statement) == {'schema_version', 'repository', 'source_commit',
                              'config_digest', 'candidate', 'evidence_sha256'}
    assert statement['schema_version'] == 1 and statement['repository'] == REPOSITORY
    assert statement['source_commit'] == revision and DIGEST.fullmatch(statement['config_digest'])
    candidate = statement['candidate']
    assert set(candidate) == {'schema_version', 'candidate_sha', 'backend_digest',
                              'frontend_digest', 'crawler_digest', 'built_at'}
    assert candidate['schema_version'] == 1 and candidate['candidate_sha'] == revision
    for kind in ['backend', 'frontend', 'crawler']:
        assert DIGEST.fullmatch(candidate[kind + '_digest'])
    assert datetime.datetime.fromisoformat(candidate['built_at']).tzinfo is not None
    assert set(statement['evidence_sha256']) == {'scan', 'sbom', 'http'}
    for value in statement['evidence_sha256'].values():
        assert re.fullmatch(r'[0-9a-f]{64}', value)
    return statement


def prepare(evidence, image, revision):
    candidate = json.loads((evidence / 'candidate-manifest.json').read_text())
    info = json.loads(run(['docker', 'image', 'inspect', image]))[0]
    assert info['Config']['User'] == '101:101'
    assert info['Config']['Labels']['org.opencontainers.image.revision'] == revision
    assert info['Config']['Labels']['org.opencontainers.image.source'] == 'https://git.subcult.tv/' + REPOSITORY
    http = json.loads((evidence / 'frontend-http.json').read_text())
    assert http['result'] == 'PASS' and http['source_commit'] == revision
    assert http['image'] == info['Id'] and http['uncached_marker_queries'] == 3
    assert http['served_files'] > 10 and http['internal_pages_denied'] == 3
    scan = json.loads((evidence / 'frontend-trivy.json').read_text())
    assert scan['Results'] and all(not result.get('Vulnerabilities') for result in scan['Results'])
    sbom = json.loads((evidence / 'frontend.spdx.json').read_text())
    assert sbom['spdxVersion'].startswith('SPDX-') and sbom['packages']
    statement = {'schema_version': 1, 'repository': REPOSITORY, 'source_commit': revision,
                 'config_digest': config_digest(image), 'candidate': candidate,
                 'evidence_sha256': {key: sha((evidence / name).read_bytes()) for key, name in
                                    [('scan', 'frontend-trivy.json'), ('sbom', 'frontend.spdx.json'),
                                     ('http', 'frontend-http.json')]}}
    return validate_statement(statement, revision)


def load_profile(path):
    profile = json.loads(path.read_text())
    assert profile['name'] == 'clpr-web' and profile['repository'] == REPOSITORY
    assert set(profile['services']) == {SERVICE}
    assert sha(Path(__file__).read_bytes()) == profile['delivery_sha256']
    receiver = Path(profile['receiver_path'])
    assert sha(receiver.read_bytes()) == profile['receiver_sha256']
    spec = importlib.util.spec_from_file_location('receiver', receiver)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return profile, module


def verify_signature(profile, folder):
    public_key = Path(profile['signing_public_key'])
    tool = Path(profile['cosign_path'])
    assert sha(public_key.read_bytes()) == profile['signing_public_key_sha256']
    assert sha(tool.read_bytes()) == profile['cosign_sha256']
    result = subprocess.run([str(tool), 'verify-blob', '--offline', '--key', str(public_key),
                             '--bundle', str(folder / 'bundle.json'), str(folder / 'statement.json')],
                            capture_output=True, timeout=60)
    (folder / 'verification.log').write_bytes(result.stdout + b'\n' + result.stderr)
    assert result.returncode == 0, 'Cosign bundle verification failed'


def verify_bundle(profile, folder, revision):
    statement = validate_statement(json.loads((folder / 'statement.json').read_text()), revision)
    verify_signature(profile, folder)
    return statement


def decode_packet(raw):
    assert len(raw) <= 1024 * 1024
    packet = json.loads(raw)
    assert set(packet) == {'statement', 'bundle'}
    decoded = {name: base64.b64decode(value, validate=True) for name, value in packet.items()}
    assert 0 < len(decoded['statement']) <= 64 * 1024
    assert 0 < len(decoded['bundle']) <= 512 * 1024
    return decoded


def receive_attestation(profile, receiver, revision, stream):
    receiver.assert_current_main(profile, revision)
    packet = decode_packet(stream.read(1024 * 1024 + 1))
    root = Path(profile['state']) / 'attestations'
    root.mkdir(parents=True, exist_ok=True, mode=0o700)
    with tempfile.TemporaryDirectory(dir=root) as temporary:
        folder = Path(temporary)
        for name, data in packet.items():
            (folder / (name + '.json')).write_bytes(data)
        incoming = verify_bundle(profile, folder, revision)
        target = root / revision
        if target.exists():
            accepted = verify_bundle(profile, target, revision)
            assert accepted['config_digest'] == incoming['config_digest'], 'A different image was already accepted for this revision'
        else:
            folder.rename(target)
        receiver.assert_current_main(profile, revision)
    return {'attested': revision, 'config_digest': incoming['config_digest']}


def native_snapshot(contract, receiver):
    result = {}
    for name, expected in contract['native_containers'].items():
        current = receiver.inspect(name)
        assert current['Id'] == expected['id'] and current['Image'] == expected['image'], name
        assert current['State']['Running'] and not current['State']['Restarting'], name
        if expected['healthy']:
            assert current['State']['Health']['Status'] == 'healthy', name
        assert sha(json.dumps(sorted(current['Config']['Env'])).encode()) == expected['environment_sha256'], name
        assert current['Mounts'] == expected['mounts'], name
        assert receiver.runtime_guard(current) == expected['runtime_guard'], name
        result[name] = {'id': current['Id'], 'image': current['Image'], 'started_at': current['State']['StartedAt']}
    return result


def schema_state():
    rows = run(['docker', 'exec', '-i', 'pg17-clustr', 'psql', '-X', '-U', 'clpr', '-d', 'clpr_db',
                '-v', 'ON_ERROR_STOP=1', '-Atc', SCHEMA_SQL], timeout=15).decode().splitlines()
    ledger = [line for line in rows if re.fullmatch(r'\d+\|[tf]', line)]
    structure = [line for line in rows if line.startswith('[')]
    assert len(ledger) == len(structure) == 1
    return {'ledger': ledger[0], 'structure_sha256': sha(structure[0].encode())}


def check_native(profile, receiver, revision):
    path = Path(profile['native_contract'])
    assert sha(path.read_bytes()) == profile['native_contract_sha256']
    contract = json.loads(path.read_text())
    assert Path(contract['current_pointer']).resolve() == Path(contract['native_release'])
    for filename, digest in contract['files'].items():
        assert sha(Path(filename).read_bytes()) == digest, filename
    # ls-tree's backend tree ID also covers migration and release-source drift.
    tree_id = run(['git', '--git-dir', profile['source_cache'], 'rev-parse', revision + ':backend']).decode().strip()
    assert tree_id == contract['candidate_backend_tree'], 'Backend source changed; review native compatibility first'
    snapshot = native_snapshot(contract, receiver)
    assert schema_state() == contract['schema'], 'Native schema changed'
    web = receiver.inspect('clpr-frontend')
    assert web['HostConfig']['Tmpfs'] == contract['web_tmpfs']
    assert {key: web['Config'].get(key) for key in contract['web_startup']} == contract['web_startup']
    assert web['Config']['Healthcheck'] == contract['web_healthcheck']
    return snapshot


def predeploy(profile, receiver):
    assert sha(Path(profile['compose_file']).read_bytes()) == profile['compose_sha256']
    revision = run(['git', 'ls-remote', '--exit-code', 'https://git.subcult.tv/' + REPOSITORY + '.git', 'refs/heads/main']).decode().split()[0]
    cache = Path(profile['source_cache'])
    if not cache.exists():
        run(['git', 'init', '--bare', str(cache)])
    run(['git', '--git-dir', str(cache), 'fetch', '--depth=1', 'https://git.subcult.tv/' + REPOSITORY + '.git', revision])
    receiver.assert_current_main(profile, revision)
    statement = verify_bundle(profile, Path(profile['state']) / 'attestations' / revision, revision)
    image = receiver.image_tag(profile, SERVICE, revision)
    assert config_digest(image) == statement['config_digest'], 'Imported image differs from the signed candidate'
    contract = json.loads(Path(profile['native_contract']).read_text())
    candidate = receiver.inspect(image, 'image')
    assert {key: candidate['Config'].get(key) for key in contract['web_startup']} == contract['web_startup']
    assert candidate['Config']['Healthcheck'] == contract['web_healthcheck']
    assert sorted(candidate['Config']['Env']) == contract['web_image_environment']
    status = json.loads(run(['tea', 'api', '--login', 'subcult-agent', '/repos/' + REPOSITORY + '/commits/' + revision + '/status']))
    contexts = {s['context']: s for s in status['statuses']}
    for name in QUALITY_NAMES:
        assert contexts['Authoritative release gates / ' + name + ' (push)']['status'] == 'success', name
    target = contexts[CANDIDATE_CONTEXT]['target_url']
    match = re.search(r'/runs/(\d+)/jobs/(\d+)', target)
    assert match
    jobs = json.loads(run(['tea', 'api', '--login', 'subcult-agent', '/repos/' + REPOSITORY + '/actions/runs/' + match[1] + '/jobs']))['jobs']
    job = next(j for j in jobs if j['id'] == int(match[2]))
    assert job['head_sha'] == revision
    steps = {step['name']: step for step in job['steps']}
    for name in ['Converge candidate source before image generation', 'Build each image once, scan, SBOM, publish, and sign', 'Sign frontend delivery receipt']:
        assert steps[name]['conclusion'] == 'success', name
    native = check_native(profile, receiver, revision)
    receiver.assert_current_main(profile, revision)
    receiver.write_json(Path(profile['state']) / ('native-before-' + revision + '.json'), native)
    return {'result': 'PASS', 'source_commit': revision, 'signed_config_digest': statement['config_digest'], 'native_services_and_schema': 'PASS'}


def main():
    os.umask(0o077)
    parser = argparse.ArgumentParser()
    modes = parser.add_subparsers(dest='mode', required=True)
    prepare_args = modes.add_parser('prepare')
    prepare_args.add_argument('--evidence', type=Path, required=True)
    prepare_args.add_argument('--image', required=True)
    prepare_args.add_argument('--revision', required=True)
    for mode in ['receiver', 'predeploy']:
        modes.add_parser(mode).add_argument('--profile', type=Path, required=True)
    args = parser.parse_args()
    if args.mode == 'prepare':
        result = prepare(args.evidence, args.image, args.revision)
    else:
        profile, receiver = load_profile(args.profile)
        if args.mode == 'predeploy':
            result = predeploy(profile, receiver)
        else:
            command = shlex.split(os.environ.get('SSH_ORIGINAL_COMMAND', ''))
            if command and command[0] in ['attest', 'verify-native']:
                assert len(command) == 2 and REVISION.fullmatch(command[1])
                with (Path(profile['state']) / 'deploy.lock').open('a') as lock:
                    fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
                    if command[0] == 'attest':
                        result = receive_attestation(profile, receiver, command[1], sys.stdin.buffer)
                    else:
                        receiver.assert_current_main(profile, command[1])
                        expected = json.loads((Path(profile['state']) / ('native-before-' + command[1] + '.json')).read_text())
                        assert check_native(profile, receiver, command[1]) == expected
                        result = {'result': 'PASS', 'native_runtime_unchanged': True}
            else:
                original_argv = sys.argv
                try:
                    sys.argv = [str(receiver.__file__), '--profile', str(args.profile)]
                    receiver.main()
                finally:
                    sys.argv = original_argv
                return
    print(json.dumps(result, sort_keys=True))


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        state = Path.home() / '.local/state/fleet-deploy/clpr-web'
        if len(sys.argv) > 1 and sys.argv[1] == 'prepare':
            frame = traceback.extract_tb(error.__traceback__)[-1]
            print(f'Candidate evidence rejected at {frame.name}:{frame.lineno} ({type(error).__name__})', file=sys.stderr)
        if len(sys.argv) > 1 and sys.argv[1] in ['receiver', 'predeploy'] and state.is_dir():
            (state / 'last-delivery-error.log').write_text(traceback.format_exc())
        print('Frontend delivery refused or failed; inspect private host evidence', file=sys.stderr)
        sys.exit(1)
