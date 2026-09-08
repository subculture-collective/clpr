import http from 'k6/http';
import { check, fail, sleep } from 'k6';

const baseURL = __ENV.BASE_URL || 'http://host.docker.internal:8080';
const token = __ENV.AUTH_TOKEN || '';
const adminToken = __ENV.ADMIN_TOKEN || token;
const clipID = __ENV.CLIP_ID || '';
const searchQuery = encodeURIComponent(__ENV.SEARCH_QUERY || 'speedrun');
const profile = __ENV.PROFILE || 'baseline';
// Each VU owns its refresh tokens throughout the 30-minute soak.
const fixtureProfiles = __ENV.AUTH_FIXTURES_FILE ? JSON.parse(open(__ENV.AUTH_FIXTURES_FILE)) : null;
const sessions = {};

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
    release_traffic: scenario('releaseTraffic'),
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
    'http_req_failed{journey:submission}': ['rate<0.005'],
    'http_req_duration{journey:submission}': ['p(95)<1000'],
    'http_req_failed{journey:moderation}': ['rate<0.005'],
    'http_req_duration{journey:moderation}': ['p(95)<1000'],
    checks: ['rate>0.99'],
  },
};

const releaseJourneys = [feed, clipDetail, search, comments, auth, submission, moderation];

// Keep the declared VU count global to the profile. Previously every journey
// received its own 5/25/75-VU scenario, multiplying the intended load by seven.
export function releaseTraffic() {
  releaseJourneys[__ITER % releaseJourneys.length]();
}

function headers(role = 'member') {
  let value = role === 'admin' ? adminToken : token;
  if (fixtureProfiles) {
    const fixture = fixtureProfiles[profile]?.[__VU - 1]?.[role];
    if (!fixture?.refresh_token) fail('Missing per-VU authenticated load fixture');
    const session = sessions[role] || (sessions[role] = { refresh: fixture.refresh_token, refreshAt: 0 });
    if (Date.now() >= session.refreshAt) {
      // Keep member and admin refresh cookies from overriding each other's body.
      http.cookieJar().clear(baseURL);
      const response = http.post(`${baseURL}/api/v1/auth/refresh`, JSON.stringify({ refresh_token: session.refresh }), {
        headers: { 'Content-Type': 'application/json' }, responseType: 'text', tags: { journey: 'auth' },
      });
      if (!check(response, { 'load session refresh succeeds': r => r.status === 200 })) fail('Load session refresh failed');
      const renewed = response.json();
      if (!renewed.access_token || !renewed.refresh_token) fail('Load session refresh omitted credentials');
      session.access = renewed.access_token;
      session.refresh = renewed.refresh_token;
      session.refreshAt = Date.now() + 14 * 60 * 1000;
      http.cookieJar().clear(baseURL);
    }
    value = session.access;
  }
  return value ? { Authorization: `Bearer ${value}` } : {};
}

export function setup() {
  if (!clipID) fail('CLIP_ID must identify a repository-owned load fixture');
  if (!fixtureProfiles && !token) fail('AUTH_TOKEN must identify a disposable load-test user');
  if (!fixtureProfiles && !adminToken) fail('ADMIN_TOKEN must identify a disposable moderator/admin');
  if (__ENV.REQUIRE_MUTATIONS === 'true' && !__ENV.SUBMISSION_URL) fail('SUBMISSION_URL is required when mutations are enabled');
}

function expectStatus(response, statuses, name) {
  check(response, { [`${name} returns ${statuses.join('/')}`]: r => statuses.includes(r.status) });
  // Sustained user traffic must fit the ordinary 1,000-request/hour IP budget.
  // Four seconds between actions caps each VU at 900 requests/hour; the
  // separate rateLimit scenario still deliberately exercises burst rejection.
  sleep(4);
}

export function feed() {
  const periods = ['hour', 'day', 'week', 'month', 'year', 'all'];
  const period = periods[(__VU + __ITER) % periods.length];
  expectStatus(http.get(`${baseURL}/api/v1/feeds/clips?sort=trending&timeframe=${period}`, { tags: { journey: 'feed' } }), [200], 'feed');
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
  expectStatus(http.get(`${baseURL}/api/v1/auth/me`, { headers: headers(), tags: { journey: 'auth' } }), [200], 'authenticated profile');
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
    headers: headers('admin'), tags: { journey: 'moderation' },
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
