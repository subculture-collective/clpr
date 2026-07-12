import http from 'k6/http';
import { check, fail, sleep } from 'k6';

const baseURL = __ENV.BASE_URL || 'http://host.docker.internal:8080';
const token = __ENV.AUTH_TOKEN || '';
const adminToken = __ENV.ADMIN_TOKEN || token;
const clipID = __ENV.CLIP_ID || '';
const searchQuery = encodeURIComponent(__ENV.SEARCH_QUERY || 'speedrun');
const profile = __ENV.PROFILE || 'baseline';

const profiles = {
  baseline: { vus: 5, duration: '1m' },
  stress: { stages: [{ duration: '2m', target: 25 }, { duration: '3m', target: 75 }, { duration: '1m', target: 0 }] },
  soak: { vus: 10, duration: __ENV.SOAK_DURATION || '30m' },
};
if (!profiles[profile]) throw new Error(`unknown PROFILE ${profile}`);

function scenario(exec) {
  if (profile === 'stress') return { executor: 'ramping-vus', startVUs: 0, stages: profiles.stress.stages, exec };
  return { executor: 'constant-vus', vus: profiles[profile].vus, duration: profiles[profile].duration, exec };
}

export const options = {
  discardResponseBodies: true,
  scenarios: {
    feed: scenario('feed'),
    clip_detail: scenario('clipDetail'),
    search: scenario('search'),
    comments: scenario('comments'),
    auth: scenario('auth'),
    submission: scenario('submission'),
    moderation: scenario('moderation'),
    rate_limit: { executor: 'per-vu-iterations', vus: 1, iterations: 1, exec: 'rateLimit' },
  },
  thresholds: {
    'http_req_failed{journey:feed}': ['rate<0.005'],
    'http_req_duration{journey:feed}': ['p(95)<500'],
    'http_req_failed{journey:clip_detail}': ['rate<0.005'],
    'http_req_duration{journey:clip_detail}': ['p(95)<500'],
    'http_req_failed{journey:search}': ['rate<0.005'],
    'http_req_duration{journey:search}': ['p(95)<750'],
    'http_req_failed{journey:comments}': ['rate<0.005'],
    'http_req_duration{journey:comments}': ['p(95)<500'],
    'http_req_failed{journey:auth}': ['rate<0.005'],
    'http_req_duration{journey:auth}': ['p(95)<500'],
    'http_req_duration{journey:submission}': ['p(95)<1000'],
    'http_req_duration{journey:moderation}': ['p(95)<1000'],
    checks: ['rate>0.99'],
  },
};

function headers(value = token) {
  return value ? { Authorization: `Bearer ${value}` } : {};
}

export function setup() {
  if (!clipID) fail('CLIP_ID must identify a repository-owned load fixture');
  if (!token) fail('AUTH_TOKEN must identify a disposable load-test user');
  if (!adminToken) fail('ADMIN_TOKEN must identify a disposable moderator/admin');
  if (__ENV.REQUIRE_MUTATIONS === 'true' && !__ENV.SUBMISSION_URL) fail('SUBMISSION_URL is required when mutations are enabled');
}

function expectStatus(response, statuses, name) {
  check(response, { [`${name} returns ${statuses.join('/')}`]: r => statuses.includes(r.status) });
  sleep(0.2);
}

export function feed() {
  expectStatus(http.get(`${baseURL}/api/v1/feeds/clips`, { tags: { journey: 'feed' } }), [200], 'feed');
}

export function clipDetail() {
  expectStatus(http.get(`${baseURL}/api/v1/clips/${clipID}`, { tags: { journey: 'clip_detail' } }), [200], 'clip detail');
}

export function search() {
  expectStatus(http.get(`${baseURL}/api/v1/search?q=${searchQuery}`, { tags: { journey: 'search' } }), [200], 'search');
}

export function comments() {
  expectStatus(http.get(`${baseURL}/api/v1/clips/${clipID}/comments`, { tags: { journey: 'comments' } }), [200], 'comments');
}

export function auth() {
  expectStatus(http.get(`${baseURL}/api/v1/users/me`, { headers: headers(), tags: { journey: 'auth' } }), [200], 'authenticated profile');
}

export function submission() {
  if (__ENV.REQUIRE_MUTATIONS !== 'true') {
    expectStatus(http.get(`${baseURL}/api/v1/submissions`, { headers: headers(), tags: { journey: 'submission' } }), [200], 'submission list');
    return;
  }
  const payload = JSON.stringify({ clip_url: __ENV.SUBMISSION_URL });
  expectStatus(http.post(`${baseURL}/api/v1/submissions`, payload, {
    headers: { ...headers(), 'Content-Type': 'application/json' },
    tags: { journey: 'submission' },
  }), [200, 201, 409, 429], 'submission');
}

export function moderation() {
  expectStatus(http.get(`${baseURL}/api/v1/admin/moderation/queue`, {
    headers: headers(adminToken), tags: { journey: 'moderation' },
  }), [200], 'moderation queue');
}

export function rateLimit() {
  let limited = false;
  for (let index = 0; index < 70; index += 1) {
    const response = http.get(`${baseURL}/api/v1/search?q=rate-limit-fixture`, { tags: { journey: 'rate_limit' } });
    if (response.status === 429) limited = true;
  }
  check(limited, { 'search rate limit returns 429': value => value });
}
