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
const AUTH_EMAIL = __ENV.AUTH_EMAIL;
const AUTH_PASSWORD = __ENV.AUTH_PASSWORD;

const rateLimitExceeded = new Rate('rate_limit_exceeded');

export function setup() {
  if (!AUTH_EMAIL || !AUTH_PASSWORD) {
    throw new Error(
      'Missing required environment variables: AUTH_EMAIL and AUTH_PASSWORD must be set.\n' +
      'Example: k6 run -e AUTH_EMAIL=user@example.com -e AUTH_PASSWORD=secret stress-test.js'
    );
  }

  const credentials = {
    email: AUTH_EMAIL,
    password: AUTH_PASSWORD,
  };

  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify(credentials),
    { headers: { 'Content-Type': 'application/json' } }
  );

  if (loginRes.status === 200) {
    const body = JSON.parse(loginRes.body);
    if (!body.accessToken) {
      throw new Error('Login response missing accessToken field');
    }
    return { token: body.accessToken };
  }

  throw new Error(`Failed to login: ${loginRes.status} - ${loginRes.body}`);
}

export default function(data) {
  const token = data.token;
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  const endpoints = [
    { path: '/api/v1/portfolios', weight: 0.5, tag: 'portfolios' },
    { path: '/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc454e4438f44e?chain=ethereum', weight: 0.3, tag: 'wallet' },
    { path: '/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc454e4438f44e/transactions', weight: 0.2, tag: 'transactions' },
  ];

  const rand = Math.random();
  let cumulative = 0;
  let selected = endpoints[0];
  
  for (const ep of endpoints) {
    cumulative += ep.weight;
    if (rand <= cumulative) {
      selected = ep;
      break;
    }
  }

  const res = http.get(`${BASE_URL}${selected.path}`, { 
    headers,
    tags: { endpoint: selected.tag }
  });

  const is429 = res.status === 429;
  if (is429) {
    rateLimitExceeded.add(1);
  }

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'response time < 1s': (r) => r.timings.duration < 1000,
  }) || errorRate.add(1);

  sleep(0.5);
}
