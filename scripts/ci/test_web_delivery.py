import base64
import copy
import importlib.util
import json
import io
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('delivery', Path(__file__).with_name('web-delivery.py'))
delivery = importlib.util.module_from_spec(spec)
spec.loader.exec_module(delivery)
REVISION = 'a' * 40
DIGEST = 'sha256:' + 'b' * 64


def statement():
    return {'schema_version': 1, 'repository': delivery.REPOSITORY, 'source_commit': REVISION,
            'config_digest': DIGEST, 'candidate': {'schema_version': 1, 'candidate_sha': REVISION,
             'backend_digest': DIGEST, 'frontend_digest': DIGEST, 'crawler_digest': DIGEST,
             'built_at': '2026-11-06T00:00:00Z'},
            'evidence_sha256': {'scan': 'c' * 64, 'sbom': 'd' * 64, 'http': 'e' * 64}}


class StatementTests(unittest.TestCase):
    def test_optimized_python_cannot_disable_delivery_guards(self):
        result = subprocess.run([sys.executable, '-O', str(Path(__file__).with_name('web-delivery.py')), '--help'], capture_output=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(b'requires enabled validation checks', result.stderr)

    def test_accepts_source_bound_candidate(self):
        self.assertEqual(delivery.validate_statement(statement(), REVISION), statement())

    def test_rejects_foreign_source_or_changed_digest_shape(self):
        for key, value in [('repository', 'other/app'), ('source_commit', 'f' * 40),
                           ('config_digest', 'latest'), ('schema_version', 2)]:
            item = statement()
            item[key] = value
            with self.subTest(key=key), self.assertRaises(AssertionError):
                delivery.validate_statement(item, REVISION)

    def test_rejects_mixed_candidate_or_missing_evidence(self):
        cases = []
        item = statement(); item['candidate']['candidate_sha'] = 'f' * 40; cases.append(item)
        item = statement(); item['candidate']['frontend_digest'] = 'candidate-main'; cases.append(item)
        item = statement(); item['candidate']['built_at'] = '2026-11-06T00:00:00'; cases.append(item)
        item = statement(); item['evidence_sha256'].pop('scan'); cases.append(item)
        item = statement(); item['evidence_sha256']['http'] = 'passed'; cases.append(item)
        item = statement(); item['public_key'] = 'untrusted caller key'; cases.append(item)
        for i, item in enumerate(cases):
            with self.subTest(case=i), self.assertRaises(AssertionError):
                delivery.validate_statement(item, REVISION)

    def test_packet_is_bounded_and_cannot_supply_its_own_trust_key(self):
        packet = {'statement': base64.b64encode(b'{}').decode(), 'bundle': base64.b64encode(b'{}').decode()}
        self.assertEqual(delivery.decode_packet(json.dumps(packet).encode()), {'statement': b'{}', 'bundle': b'{}'})
        for raw in [b'x' * (1024 * 1024 + 1), json.dumps(dict(packet, key='caller key')).encode(),
                    json.dumps(dict(packet, bundle=base64.b64encode(b'x' * (512 * 1024 + 1)).decode())).encode(),
                    json.dumps(dict(packet, statement=base64.b64encode(b'x' * 65537).decode())).encode()]:
            with self.subTest(size=len(raw)), self.assertRaises(AssertionError):
                delivery.decode_packet(raw)


class NativeRuntimeTests(unittest.TestCase):
    def setUp(self):
        self.current = {'Id': 'native-id', 'Image': 'native-image',
                        'State': {'Running': True, 'Restarting': False, 'StartedAt': 'before',
                                  'Health': {'Status': 'healthy'}},
                        'Config': {'Env': ['NATIVE_CONFIG=fixture', 'PUBLIC_MODE=free']},
                        'Mounts': [], 'guard': 'registered-shape'}
        self.contract = {'native_containers': {'native-api': {
            'id': 'native-id', 'image': 'native-image', 'healthy': True, 'mounts': [],
            'environment_sha256': delivery.sha(json.dumps(sorted(self.current['Config']['Env'])).encode()),
            'runtime_guard': 'registered-shape'}}}

    def snapshot(self):
        current = self.current
        class Receiver:
            @staticmethod
            def inspect(name):
                assert name == 'native-api'
                return current
            @staticmethod
            def runtime_guard(container):
                return container['guard']
        return delivery.native_snapshot(self.contract, Receiver)

    def test_records_identity_and_start_time_for_post_deployment_check(self):
        self.assertEqual(self.snapshot(), {'native-api': {'id': 'native-id', 'image': 'native-image', 'started_at': 'before'}})
        before = self.snapshot()
        self.current['State']['StartedAt'] = 'restarted'
        self.assertNotEqual(self.snapshot(), before)

    def test_rejects_native_identity_health_env_mount_or_runtime_drift(self):
        original = copy.deepcopy(self.current)
        changes = [('id', lambda c: c.update(Id='other-id')),
                   ('image', lambda c: c.update(Image='other-image')),
                   ('stopped', lambda c: c['State'].update(Running=False)),
                   ('restarting', lambda c: c['State'].update(Restarting=True)),
                   ('unhealthy', lambda c: c['State']['Health'].update(Status='unhealthy')),
                   ('environment', lambda c: c['Config']['Env'].append('UNEXPECTED=override')),
                   ('mount', lambda c: c['Mounts'].append({'Destination': '/other'})),
                   ('isolation', lambda c: c.update(guard='different-shape'))]
        for name, mutate in changes:
            self.current = copy.deepcopy(original)
            mutate(self.current)
            with self.subTest(change=name), self.assertRaises(AssertionError):
                self.snapshot()


class AttestationStorageTests(unittest.TestCase):
    @staticmethod
    def packet(item):
        return io.BytesIO(json.dumps({'statement': base64.b64encode(json.dumps(item).encode()).decode(),
                                      'bundle': base64.b64encode(b'fixture bundle').decode()}).encode())

    @staticmethod
    def verified_fixture(profile, folder, revision):
        # Crypto acceptance/rejection is exercised separately with pinned Cosign.
        item = delivery.validate_statement(json.loads((folder / 'statement.json').read_text()), revision)
        (folder / 'verification.log').write_text('Signature fixture accepted\n')
        return item

    class Receiver:
        @staticmethod
        def assert_current_main(profile, revision):
            assert revision == REVISION

    def test_atomic_first_receipt_and_retry_keep_the_first_candidate(self):
        with tempfile.TemporaryDirectory() as root, patch.object(delivery, 'verify_bundle', self.verified_fixture):
            profile = {'state': root}
            delivery.receive_attestation(profile, self.Receiver, REVISION, self.packet(statement()))
            folder = Path(root) / 'attestations' / REVISION
            before = (folder / 'statement.json').read_bytes()
            second = statement()
            second['candidate']['built_at'] = '2026-11-06T00:01:00Z'
            second['candidate']['frontend_digest'] = 'sha256:' + 'f' * 64
            delivery.receive_attestation(profile, self.Receiver, REVISION, self.packet(second))
            self.assertEqual((folder / 'statement.json').read_bytes(), before)
            self.assertEqual([p.name for p in folder.parent.iterdir()], [REVISION])

    def test_different_image_for_the_same_revision_cannot_replace_receipt(self):
        with tempfile.TemporaryDirectory() as root, patch.object(delivery, 'verify_bundle', self.verified_fixture):
            profile = {'state': root}
            delivery.receive_attestation(profile, self.Receiver, REVISION, self.packet(statement()))
            target = Path(root) / 'attestations' / REVISION / 'statement.json'
            original = target.read_bytes()
            changed = statement()
            changed['config_digest'] = 'sha256:' + 'f' * 64
            with self.assertRaises(AssertionError):
                delivery.receive_attestation(profile, self.Receiver, REVISION, self.packet(changed))
            self.assertEqual(target.read_bytes(), original)


if __name__ == '__main__':
    unittest.main()
