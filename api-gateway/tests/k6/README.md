# k6 Load Tests for API Gateway

This directory contains k6 load testing scenarios for the Million Hunter API Gateway.

## Prerequisites

- [k6](https://k6.io/docs/getting-started/installation/) installed
- API Gateway running (default: `http://localhost:8080`)
- Auth service running and accessible
- Valid test credentials

## Test Scenarios

### load-test.js - Production Traffic Simulation

Simulates realistic production traffic patterns with gradual ramp-up and spike testing.

**Traffic Pattern:**
- Ramp up: 10 → 50 → 100 concurrent users
- Duration: ~5.5 minutes total
- Tests multiple endpoints (health, portfolios, wallets)
- Optional write operations (POST)

**SLO Validation:**
- **p95 latency** < 300ms (production requirement)
- **Error rate** < 1% (5xx responses)
- Rate limiting headers present on all responses

**Usage:**
```bash
# Basic run (read-only operations)
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e AUTH_EMAIL=user@example.com \
  -e AUTH_PASSWORD=your-password \
  api-gateway/tests/k6/load-test.js

# With write operations enabled
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e AUTH_EMAIL=user@example.com \
  -e AUTH_PASSWORD=your-password \
  -e ENABLE_WRITE_OPS=true \
  api-gateway/tests/k6/load-test.js
```

**What it tests:**
- Health check endpoint (`GET /healthz`)
- List portfolios (`GET /api/v1/portfolios`)
- Get wallet details (`GET /api/v1/wallets/:address`)
- Add wallet (`POST /api/v1/portfolios`) - optional

### stress-test.js - Stress & Capacity Testing

Tests system behavior under extreme load to identify breaking points and rate limiting effectiveness.

**Traffic Pattern:**
- Ramp up: 100 → 200 → 300 concurrent users
- Duration: ~9 minutes total
- Random endpoint selection with weighted distribution
- High request rate (0.5s sleep between requests)

**Stress Thresholds:**
- **p95 latency** < 500ms (relaxed for stress)
- **Error rate** < 5% (allows some degradation)
- Tracks rate limit exceeded (429) responses

**Usage:**
```bash
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e AUTH_EMAIL=user@example.com \
  -e AUTH_PASSWORD=your-password \
  api-gateway/tests/k6/stress-test.js
```

**What it tests:**
- Random endpoint selection (weighted by expected traffic)
- Rate limiting under extreme load
- System degradation patterns
- Recovery after load spike

**Endpoint Weights:**
- `GET /api/v1/portfolios` - 50% of traffic
- `GET /api/v1/wallets/:address` - 30% of traffic
- `GET /api/v1/wallets/:address/transactions` - 20% of traffic

## Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `BASE_URL` | API Gateway URL | `http://localhost:8080` |
| `AUTH_EMAIL` | Test user email | `user@example.com` |
| `AUTH_PASSWORD` | Test user password | `your-password` |
| `ENABLE_WRITE_OPS` | Enable write operations in load test (optional) | `true` |

**⚠️ Security Warning**: NEVER commit credentials to version control. Always pass them via environment variables at runtime.

## Interpreting Results

### Key Metrics

**http_req_duration**: Request latency
- `p(95)` - 95th percentile (SLO: <300ms for load, <500ms for stress)
- `p(99)` - 99th percentile (watch for outliers)
- `avg` - Average latency
- `max` - Worst case latency

**http_req_failed**: Error rate
- Should be < 1% for load tests
- Should be < 5% for stress tests
- Excludes 429 (rate limit) responses

**http_reqs**: Throughput
- Total requests per second
- Should meet minimum: 100 req/s per instance

**Custom Metrics:**
- `errors` - Failed checks (response validation)
- `rate_limit_exceeded` (stress test) - 429 responses

### Example Output

```
✓ health check status is 200
✓ list portfolios status is 200
✓ list portfolios response has items
✓ rate limit headers present

checks.........................: 100.00% ✓ 12000 ✗ 0
data_received..................: 15 MB   50 kB/s
data_sent......................: 3.2 MB  11 kB/s
http_req_duration..............: avg=245ms  p(95)=289ms p(99)=345ms
  { expected_response:true }...: avg=245ms  p(95)=289ms p(99)=345ms
http_req_failed................: 0.00%   ✓ 0    ✗ 12000
http_reqs......................: 12000   40/s
iteration_duration.............: avg=5.2s   min=4.1s max=6.8s
iterations.....................: 2400    8/s
```

**✅ Pass Criteria:**
- All thresholds green (✓)
- `http_req_duration p(95)` < SLO threshold
- `http_req_failed` < error threshold
- No unexpected errors in checks

**❌ Failure Indicators:**
- Red thresholds (✗)
- High `http_req_failed` rate
- Missing rate limit headers
- Authentication failures

## Authentication Payload

The login endpoint should return:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Tests validate presence of `accessToken` field and fail immediately if missing.

## Rate Limiting Behavior

The API Gateway enforces rate limits on all protected endpoints:

**Expected Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1697564890
Retry-After: 3  (on 429 responses only)
```

**429 Responses:**
- Normal under stress testing (indicates rate limiting works)
- Should be rare in load testing (indicates limits too low)
- Tests accept 429 as valid response for write operations

## Troubleshooting

### "Missing required environment variables"

**Problem**: k6 script throws error immediately

**Solution**: Provide `AUTH_EMAIL` and `AUTH_PASSWORD`:
```bash
k6 run -e AUTH_EMAIL=user@example.com -e AUTH_PASSWORD=pass load-test.js
```

### "Failed to login: 401"

**Problem**: Authentication fails

**Solution**: 
- Verify credentials are correct
- Check auth service is running
- Ensure `AUTH_SERVICE_URL` is configured in gateway
- Verify `JWT_SECRET` matches between gateway and auth service

### High latency (p95 > SLO)

**Problem**: Requests slower than expected

**Solution**:
- Check Redis connectivity and latency
- Review auth service response times
- Monitor system resources (CPU, memory)
- Check network latency between services
- Review Prometheus metrics at `/metrics`

### High error rate

**Problem**: Many failed requests (5xx errors)

**Solution**:
- Check API Gateway logs for errors
- Verify all dependencies are healthy (`/healthz`)
- Reduce concurrent users to isolate issue
- Check for resource exhaustion

### Rate limiting not working

**Problem**: No 429 responses even under stress

**Solution**:
- Verify `REDIS_URL` is set correctly
- Check rate limit configuration (`RATE_DEFAULT_RPS`, `RATE_DEFAULT_BURST`)
- Review rate limit metrics at `/metrics`
- Check Redis logs for errors

## Best Practices

### Running Tests Safely

1. **Never run against production** without explicit approval
2. **Use dedicated test credentials** - not real user accounts
3. **Start with load test** before stress testing
4. **Monitor system resources** during tests
5. **Have rollback plan** if issues discovered

### Test Environment

1. **Isolate test environment** from production
2. **Use realistic data** similar to production
3. **Match production configuration** (rate limits, etc.)
4. **Scale infrastructure** to match production capacity
5. **Enable observability** (metrics, logs, traces)

### Credential Management

```bash
# Good: Use environment file (not committed)
source .env.test && k6 run -e AUTH_EMAIL -e AUTH_PASSWORD load-test.js

# Good: Pass inline for CI/CD
k6 run -e AUTH_EMAIL=$TEST_EMAIL -e AUTH_PASSWORD=$TEST_PASSWORD load-test.js

# Bad: Hardcode in script
# const password = 'secret123'  ❌ NEVER DO THIS

# Bad: Commit to git
# echo "AUTH_PASSWORD=secret" >> .env.test
# git add .env.test  ❌ NEVER DO THIS
```

### CI/CD Integration

```yaml
# Example GitHub Actions
name: Load Tests
on: [pull_request]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install k6
        run: |
          curl https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz -L | tar xvz
          sudo mv k6-v0.47.0-linux-amd64/k6 /usr/local/bin
      
      - name: Start services
        run: docker-compose up -d
      
      - name: Run load test
        env:
          AUTH_EMAIL: ${{ secrets.TEST_USER_EMAIL }}
          AUTH_PASSWORD: ${{ secrets.TEST_USER_PASSWORD }}
        run: |
          k6 run \
            -e BASE_URL=http://localhost:8080 \
            -e AUTH_EMAIL \
            -e AUTH_PASSWORD \
            api-gateway/tests/k6/load-test.js
```

## Advanced Usage

### Custom Thresholds

Override default thresholds:
```bash
k6 run \
  --threshold 'http_req_duration{endpoint:wallets}=p(95)<200' \
  --threshold 'http_req_failed=rate<0.005' \
  load-test.js
```

### Output to File

Export results for analysis:
```bash
k6 run --out json=results.json load-test.js
k6 run --out influxdb=http://localhost:8086/k6 load-test.js
```

### Custom Scenarios

Modify test duration and users:
```javascript
export const options = {
  stages: [
    { duration: '2m', target: 50 },  // Custom ramp
  ],
};
```

## Further Reading

- [k6 Documentation](https://k6.io/docs/)
- [k6 HTTP Requests](https://k6.io/docs/using-k6/http-requests/)
- [k6 Thresholds](https://k6.io/docs/using-k6/thresholds/)
- [k6 Metrics](https://k6.io/docs/using-k6/metrics/)
- [API Gateway Monitoring](../README.md#health--observability)
