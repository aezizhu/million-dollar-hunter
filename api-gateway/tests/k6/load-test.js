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

const AUTH_EMAIL = __ENV.AUTH_EMAIL;
const AUTH_PASSWORD = __ENV.AUTH_PASSWORD;

export function setup() {
  if (!AUTH_EMAIL || !AUTH_PASSWORD) {
    throw new Error(
      'Missing required environment variables: AUTH_EMAIL and AUTH_PASSWORD must be set.\n' +
      'Example: k6 run -e AUTH_EMAIL=user@example.com -e AUTH_PASSWORD=secret load-test.js'
    );
  }

  const credentials = {
    email: AUTH_EMAIL,
    password: AUTH_PASSWORD,
  };

  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify(credentials),
    {
      headers: { 'Content-Type': 'application/json' },
    }
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

  let res = http.get(`${BASE_URL}/healthz`);
  check(res, {
    'health check status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.5 + Math.random() * 0.5); // Random jitter (0.5-1s)

  res = http.get(`${BASE_URL}/api/v1/portfolios`, { 
    headers,
    tags: { endpoint: 'portfolios' }
  });
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
    'rate limit headers present': (r) => 
      r.headers['X-Ratelimit-Limit'] !== undefined,
  }) || errorRate.add(1);

  sleep(0.5 + Math.random() * 1); // Random jitter (0.5-1.5s)

  const mockAddress = '0x742d35Cc6634C0532925a3b844Bc454e4438f44e';
  res = http.get(
    `${BASE_URL}/api/v1/wallets/${mockAddress}?chain=ethereum`,
    { 
      headers,
      tags: { endpoint: 'wallet-details' }
    }
  );
  check(res, {
    'get wallet status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(0.5 + Math.random() * 1); // Random jitter (0.5-1.5s)

  if (__ENV.ENABLE_WRITE_OPS === 'true') {
    res = http.post(
      `${BASE_URL}/api/v1/portfolios`,
      JSON.stringify({
        address: mockAddress,
        chain: 'ethereum',
        nickname: `Test-${Date.now()}`, // Unique to avoid conflicts
      }),
      { 
        headers,
        tags: { endpoint: 'add-wallet' }
      }
    );
    check(res, {
      'add wallet status is 202 or 429': (r) => r.status === 202 || r.status === 429,
    }) || errorRate.add(1);
  }

  sleep(1 + Math.random() * 1); // Random jitter (1-2s)
}

export function teardown(data) {
  console.log('Load test completed');
}
