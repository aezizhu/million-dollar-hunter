import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '1m', target: 100 },   // Ramp up to 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '3m', target: 300 },   // Ramp up to 300 users (stress point)
    { duration: '2m', target: 300 },   // Hold at 300 users
    { duration: '1m', target: 0 },     // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // Allow higher latency under stress
    'http_req_failed': ['rate<0.05'],   // Allow 5% error rate under stress
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({
      email: 'aezi@example.com',
      password: 'Aa@123456789',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  if (loginRes.status === 200) {
    const body = JSON.parse(loginRes.body);
    return { token: body.accessToken };
  }

  return { token: '' };
}

export default function(data) {
  const token = data.token;
  const headers = {
    'Authorization': `Bearer ${token}`,
  };

  const endpoints = [
    '/api/v1/portfolios',
    '/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc454e4438f44e?chain=ethereum',
    '/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc454e4438f44e/transactions',
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const res = http.get(`${BASE_URL}${endpoint}`, { headers });

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'response time < 1s': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);

  sleep(0.5); // High request rate
}
