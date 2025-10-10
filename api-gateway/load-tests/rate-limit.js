import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 20,
  duration: '15s',
  thresholds: {
    http_req_failed: ['rate<0.20'],
    http_req_duration: ['p(95)<300'],
  },
};

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const USER = __ENV.ADMIN_USER || 'aezi';
const PASS = __ENV.ADMIN_PASS || 'Aa@123456789';

export function setup() {
  const loginRes = http.post(
    `${BASE}/api/v1/auth/login`,
    JSON.stringify({ email: USER, password: PASS }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(loginRes, { 'login 200': (r) => r.status === 200 });
  return { token: loginRes.json('accessToken') };
}

export default function (data) {
  const params = { headers: { Authorization: `Bearer ${data.token}` } };

  const res = http.get(`${BASE}/api/v1/portfolios`, params);

  const ok = check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'rate limit headers present': (r) =>
      r.headers['X-RateLimit-Limit'] && r.headers['X-RateLimit-Remaining'] && r.headers['X-RateLimit-Reset'],
    'retry-after when 429': (r) => (r.status !== 429) || !!r.headers['Retry-After'],
  });

  sleep(0.1);
}
