import http from 'k6/http';
import { sleep, check } from 'k6';

export let options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate&lt;0.01'],
    http_req_duration: ['p(95)&lt;300'],
  },
};

const BASE = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const loginRes = http.post(`${BASE}/api/v1/auth/login`, JSON.stringify({email: 'aezi', password: 'Aa@123456789'}), { headers: {'Content-Type': 'application/json'}});
  check(loginRes, { 'login 200': (r) => r.status === 200 });
  const token = loginRes.json('accessToken') || 'dev-access';
  const params = { headers: { Authorization: `Bearer ${token}` } };

  const r1 = http.get(`${BASE}/api/v1/portfolios`, params);
  check(r1, { 'portfolios 200': (r) => r.status === 200 });

  const r2 = http.get(`${BASE}/metrics`);
  check(r2, { 'metrics 200': (r) => r.status === 200 });

  sleep(1);
}
