#!/usr/bin/env python3
"""Check the candidate's actual HTTP bytes before signing or publication."""
import json
from pathlib import PurePosixPath
import re
import subprocess
import sys
import time


def run(args):
    return subprocess.check_output(args, stderr=subprocess.PIPE)


def main():
    image, revision = sys.argv[1:]
    assert re.fullmatch(r'[0-9a-f]{40}', revision), 'Expected a complete source revision'
    info = json.loads(run(['docker', 'image', 'inspect', image]))[0]
    labels = info['Config']['Labels']
    assert labels['org.opencontainers.image.revision'] == revision
    assert labels['org.opencontainers.image.source'] == 'https://git.subcult.tv/subculture-collective/clpr'
    assert info['Config']['User'] == '101:101'
    container = run(['docker', 'run', '-d', '--network', 'none', image]).decode().strip()

    def fetch(path, status=200):
        result = subprocess.run(['docker', 'exec', container, 'wget', '-S', '-O', '-',
                                 'http://127.0.0.1:8080' + path], capture_output=True)
        assert (result.returncode == 0) == (status == 200), path
        assert ('HTTP/1.1 ' + str(status)).encode() in result.stderr, path
        headers = {}
        for line in result.stderr.splitlines():
            name, separator, value = line.strip().partition(b':')
            if separator:
                headers[name.lower()] = value.strip()
        return result.stdout, headers

    try:
        deadline = time.monotonic() + 30
        while True:
            try:
                fetch('/health.html')
                break
            except AssertionError:
                assert time.monotonic() < deadline, 'Candidate nginx failed to start'
                time.sleep(0.25)

        for query in ['', '?v=' + revision, '?v=stale']:
            body, headers = fetch('/build-info.json' + query)
            assert json.loads(body) == {'source_commit': revision, 'repository': 'subculture-collective/clpr'}
            assert headers.get(b'cache-control') == b'no-store'

        # These files deliberately have public aliases and reject their source paths.
        aliases = {'seo-pages/home.html': '/', 'seo-pages/about.html': '/about',
                   'seo-pages/support.html': '/support'}
        # The native API owns this response; the isolated image has no backend.
        proxy_paths = {'sitemap.xml'}
        files = run(['docker', 'exec', container, 'find', '/usr/share/nginx/html', '-type', 'f']).decode().splitlines()
        assets, denied_maps, internal_pages = 0, 0, 0
        for filename in files:
            relative = str(PurePosixPath(filename).relative_to('/usr/share/nginx/html'))
            path = '/' + relative
            if relative in proxy_paths:
                continue
            if relative.startswith('seo-pages/'):
                assert relative in aliases, 'Review a new internal metadata page'
                fetch(path, 404)
                path = aliases[relative]
                internal_pages += 1
            elif relative.endswith('.map'):
                fetch(path, 404)
                denied_maps += 1
                continue
            elif relative == 'health.html':
                assert fetch(path)[0] == b'healthy\n'
                continue
            body, _ = fetch(path)
            assert body == run(['docker', 'exec', container, 'cat', filename]), path
            assets += 1

        assert assets > 10 and internal_pages == len(aliases)
        root, _ = fetch('/')
        assert b'https://clpr.tv/' in root and b'og:site_name' in root and b'summary_large_image' in root
        fetch('/.clpr-ci-unrecognized', 404)
        fetch('/50x.html', 404)
        print(json.dumps({'result': 'PASS', 'source_commit': revision, 'image': info['Id'],
                          'served_files': assets, 'source_maps_denied': denied_maps,
                          'internal_pages_denied': internal_pages, 'uncached_marker_queries': 3,
                          'native_proxy_paths': sorted('/' + p for p in proxy_paths),
                          'public_metadata': 'PASS', 'runtime_user': '101:101'}))
    finally:
        run(['docker', 'rm', '-f', container])


if __name__ == '__main__':
    main()
