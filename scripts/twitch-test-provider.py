#!/usr/bin/env python3
"""Loopback-only Twitch fixture for disposable browser journeys, never OAuth evidence."""
import datetime
import http.server
import json
import sys
import urllib.parse


class Provider(http.server.BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def do_POST(self):
        if urllib.parse.urlparse(self.path).path != "/oauth2/token":
            self.send_error(404)
            return
        self.reply({"access_token": "disposable-provider-token", "expires_in": 3600, "token_type": "bearer"})

    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        if parsed.path != "/helix/clips":
            self.reply({"data": [], "pagination": {}})
            return
        ids = urllib.parse.parse_qs(parsed.query).get("id", [])
        self.reply({"data": [{
            "id": clip, "url": "https://clips.twitch.tv/" + clip,
            "embed_url": "https://clips.twitch.tv/embed?clip=" + clip,
            "title": "Candidate submission " + clip,
            "creator_name": "Fixture Creator", "creator_id": "fixture-creator",
            "broadcaster_name": "Fixture Channel", "broadcaster_id": "fixture-channel",
            "game_id": "release-game", "language": "en", "view_count": 100,
            "created_at": datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z"), "duration": 30,
        } for clip in ids if clip.startswith("clpr-e2e-")], "pagination": {}})

    def reply(self, data):
        body = json.dumps(data).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    http.server.ThreadingHTTPServer(("127.0.0.1", int(sys.argv[1])), Provider).serve_forever()
