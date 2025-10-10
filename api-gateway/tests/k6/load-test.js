import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 }, // Spike to 100 users
    { duration: '1m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<300'], // 95% of requests should be below 300ms
    'http_req_failed': ['rate<0.01'],   // Error rate should be below 1%
    'errors': ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const credentials = {
  email: 'aezi@example.com',
  password: 'Aa@123456789',
};

let authToken = '';

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify(credentials),
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (loginRes.status === 200) {
    const body = JSON.parse(loginRes.body);
    return { token: body.accessToken };
  }

  console.error('Failed to login:', loginRes.status, loginRes.body);
  return { token: '' };
}

export default function(data) {
  const token = data.token;
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  let res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health check status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);

  res = http.get(`${BASE_URL}/api/v1/portfolios`, { headers });
  check(res, {
    'list portfolios status is 200': (r) => r.status === 200,
    'list portfolios response has items': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.hasOwnProperty('items');
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  sleep(1);

  const mockAddress = '0x742d35Cc6634C0532925a3b844Bc454e4438f44e';
  res = http.get(
    `${BASE_URL}/api/v1/wallets/${mockAddress}?chain=ethereum`,
    { headers }
  );
  check(res, {
    'get wallet status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);

  res = http.post(
    `${BASE_URL}/api/v1/portfolios`,
    JSON.stringify({
      address: mockAddress,
      chain: 'ethereum',
      nickname: 'Test Wallet',
    }),
    { headers }
  );
  check(res, {
    'add wallet status is 202 or 429': (r) => r.status === 202 || r.status === 429,
  }) || errorRate.add(1);

  sleep(2);
}

export function teardown(data) {
  console.log('Load test completed');
}
