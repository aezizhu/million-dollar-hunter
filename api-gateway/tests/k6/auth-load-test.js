import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const rateLimited = new Rate('rate_limited');

export const options = {
  scenarios: {
    auth: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    'http_req_duration{endpoint:auth_login}': ['p(50)<200', 'p(95)<500', 'p(99)<1000'],
    'http_req_duration{endpoint:auth_refresh}': ['p(50)<200', 'p(95)<500', 'p(99)<1000'],
    'http_req_failed': ['rate<0.02'],
    'rate_limited': ['rate<0.2'],
  },
  timeout: '10s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const AUTH_EMAIL = __ENV.AUTH_EMAIL;
const AUTH_PASSWORD = __ENV.AUTH_PASSWORD;

export function setup() {
  if (!AUTH_EMAIL || !AUTH_PASSWORD) {
    throw new Error('Set AUTH_EMAIL and AUTH_PASSWORD env vars');
  }
  return null;
}

function hasRateLimitHeaders(h) {
  return h['X-RateLimit-Limit'] !== undefined || h['RateLimit-Limit'] !== undefined;
}

export default function () {
  const creds = JSON.stringify({ email: AUTH_EMAIL, password: AUTH_PASSWORD });
  let res = http.post(`${BASE_URL}/api/v1/auth/login`, creds, { headers: { 'Content-Type': 'application/json' }, tags: { endpoint: 'auth_login' } });
  const okLogin = check(res, {
    'login status 200': (r) => r.status === 200,
    'rate limit headers present': (r) => hasRateLimitHeaders(r.headers),
  });
  if (!okLogin) errorRate.add(1);
  if (res.status === 429) rateLimited.add(1);

  let refreshToken = null;
  try {
    const body = JSON.parse(res.body);
    refreshToken = body.refreshToken || body.refresh_token;
  } catch (e) {
  }

  if (refreshToken) {
    const payload = JSON.stringify({ refresh_token: refreshToken });
    res = http.post(`${BASE_URL}/api/v1/auth/refresh`, payload, { headers: { 'Content-Type': 'application/json' }, tags: { endpoint: 'auth_refresh' } });
    const okRefresh = check(res, {
      'refresh status ok/allowed': (r) => r.status === 200 || r.status === 501 || r.status === 429,
      'rate limit headers present': (r) => hasRateLimitHeaders(r.headers),
    });
    if (!okRefresh) errorRate.add(1);
    if (res.status === 429) rateLimited.add(1);
  }

  sleep(0.1);
}

export function teardown() {
  console.log('Auth load test completed');
}
