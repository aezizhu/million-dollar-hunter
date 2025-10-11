# PR #14 Staging Validation Guide

This guide provides step-by-step instructions for validating the Kafka integration and TokenAddress extraction in a staging environment before merging to production.

## Prerequisites

- Docker and Docker Compose installed
- API keys for Alchemy and Moralis (optional for validation, but needed for real data)
- Access to a wallet address with known ERC-20 token transactions

## Quick Start

### 1. Set Environment Variables

Create a `.env` file in the `ops/` directory:

```bash
# Required for real blockchain data testing
ALCHEMY_API_KEY=your_alchemy_api_key
MORALIS_API_KEY=your_moralis_api_key

# Optional: override default URLs
ALCHEMY_BASE_URL=https://eth-mainnet.g.alchemy.com/v2
MORALIS_BASE_URL=https://deep-index.moralis.io/api/v2.2
```

### 2. Start the Stack

```bash
cd ops/
docker-compose up -d
```

Wait for all services to be healthy:

```bash
docker-compose ps
```

Expected output should show all services as "Up" and healthy.

### 3. Verify Services

Check ingestion service health:

```bash
curl http://localhost:8090/healthz | jq
```

Expected response:
```json
{
  "status": "ok",
  "kafka": "enabled",
  "queue_depth": 0,
  "queue_capacity": 64
}
```

### 4. Trigger Test Ingestion

**Option A: Using a test wallet with known ERC-20 transactions**

Use a well-known DeFi wallet address (e.g., Vitalik's address or a popular DeFi protocol):

```bash
TEST_WALLET="0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"  # Example: Vitalik's address
TEST_CHAIN="ethereum"
```

**Option B: Manual API call (if API endpoint exists)**

```bash
# Replace with actual ingestion API endpoint
curl -X POST http://localhost:8090/api/v1/ingest \
  -H "Content-Type: application/json" \
  -d '{"wallet": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "chain": "ethereum"}'
```

### 5. Run Validation Script

```bash
cd ops/
./validate-staging.sh
```

The script will automatically check:
- ✓ Service health
- ✓ Kafka message structure
- ✓ Database records with token_address
- ✓ Native ETH handling (0x0000...0000 address)
- ✓ No duplicate assets
- ✓ Transaction hash validation

## Manual Validation Steps

If you prefer manual validation:

### Step 1: Check Kafka Messages

```bash
# View recent messages on TransactionDataIngested topic
docker exec -it ops-kafka-1 kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic TransactionDataIngested \
  --from-beginning \
  --max-messages 10
```

**What to look for:**
- Each message should have `wallet_address`, `chain`, and `transactions` array
- Each transaction should have `token_address` field (not empty for ERC-20)
- Native ETH transfers should have `token_address: "0x0000000000000000000000000000000000000000"`

Example valid message:
```json
{
  "schema_version": "1.0.0",
  "event_id": "uuid-here",
  "wallet_address": "0x742d35Cc...",
  "chain": "ethereum",
  "transactions": [
    {
      "hash": "0xabc123...",
      "from": "0x111...",
      "to": "0x222...",
      "amount": "1000000",
      "symbol": "USDC",
      "token_address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
      "timestamp": "2024-01-01T00:00:00Z",
      "type": "erc20"
    }
  ]
}
```

### Step 2: Verify Database Records

Connect to portfolio database:

```bash
docker exec -it ops-postgres-1 psql -U portfolio -d portfolio
```

**Critical Validation Queries:**

1. Check assets have token addresses (not symbols):
```sql
SELECT 
  LEFT(wallet_id::text, 8) as wallet,
  token_address,
  symbol,
  current_balance
FROM assets
WHERE token_address IS NOT NULL AND token_address != ''
LIMIT 10;
```

**Expected**: `token_address` should be actual contract addresses like `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` (USDC), NOT symbols like "USDC".

2. Verify native ETH uses zero address:
```sql
SELECT token_address, symbol, current_balance
FROM assets
WHERE symbol = 'ETH';
```

**Expected**: `token_address` should be `0x0000000000000000000000000000000000000000`

3. Check for duplicate tokens (should be ZERO):
```sql
SELECT wallet_id, token_address, symbol, COUNT(*)
FROM assets
GROUP BY wallet_id, token_address, symbol
HAVING COUNT(*) > 1;
```

**Expected**: No results (if there are duplicates, data corruption occurred)

4. Verify distinct token addresses vs symbols:
```sql
SELECT 
  COUNT(DISTINCT token_address) as distinct_addresses,
  COUNT(DISTINCT symbol) as distinct_symbols
FROM assets
WHERE token_address != '';
```

**Expected**: `distinct_addresses` should be >= `distinct_symbols` (proves we're using addresses, not symbols)

### Step 3: Check Transaction Records

```sql
SELECT 
  LEFT(tx_hash, 15) as hash,
  asset_symbol,
  amount,
  type,
  ts
FROM transactions_view
LIMIT 10;
```

**What to verify:**
- No transactions with empty `tx_hash`
- Timestamps look reasonable (not all set to current time)
- Asset symbols match the transactions

### Step 4: Monitor Logs

Watch for any errors during ingestion:

```bash
# Ingestion service logs
docker-compose logs -f ingestion-service

# Portfolio service logs (consumer)
docker-compose logs -f portfolio-service
```

**Red flags to watch for:**
- "type assertion failed" warnings
- "empty token_address" errors
- "kafka publish failed" errors
- "unmarshal event" errors

## Success Criteria

The staging deployment is considered validated if:

- [x] ✅ Kafka messages contain `token_address` field for all transactions
- [x] ✅ Database `assets` table uses contract addresses, not symbols
- [x] ✅ Native ETH transfers have `0x0000...0000` address
- [x] ✅ No duplicate assets per wallet+token_address combination
- [x] ✅ All transactions have non-empty hashes
- [x] ✅ Multiple distinct token addresses exist (proving Symbol isn't used)
- [x] ✅ No errors in service logs
- [x] ✅ Health checks show "ok" or "degraded" (degraded is acceptable if Kafka is optional)

## Failure Scenarios & Troubleshooting

### Scenario 1: Empty `token_address` for ERC-20 tokens

**Symptom**: ERC-20 transfers in DB have empty `token_address`

**Cause**: Alchemy API response structure changed or transformer parsing failed

**Fix**: 
1. Check Alchemy API response format
2. Update transformer in `ingestion-service/internal/transformer/alchemy.go`
3. Add logging for type assertion failures

### Scenario 2: Symbol used as `token_address`

**Symptom**: `token_address` column contains "USDC" instead of "0xA0b86991..."

**Cause**: Critical bug - code is using Symbol field instead of TokenAddress

**Fix**: DO NOT MERGE - investigate portfolio-service consumer code

### Scenario 3: Kafka messages not appearing

**Symptom**: Kafka topic is empty or health check shows "unavailable"

**Possible causes**:
1. Kafka service not fully started (wait 30s and retry)
2. Ingestion service failed to connect to Kafka
3. Topic not created automatically

**Fix**:
```bash
# Check Kafka logs
docker-compose logs kafka

# Manually create topic if needed
docker exec ops-kafka-1 kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic TransactionDataIngested \
  --partitions 3 \
  --replication-factor 1
```

### Scenario 4: Portfolio service not consuming events

**Symptom**: Kafka has messages but database is empty

**Check**:
```bash
# Check consumer group lag
docker exec ops-kafka-1 kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group portfolio-service \
  --describe
```

**Possible fixes**:
1. Restart portfolio-service: `docker-compose restart portfolio-service`
2. Check portfolio-service logs for errors
3. Verify database connection

## Real-World Test Wallets

Use these addresses for comprehensive testing:

### Ethereum Mainnet
- **Vitalik's Address**: `0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045`
  - Has many ERC-20 token transfers (USDC, DAI, WETH, etc.)
  - Good for testing token address extraction

- **Uniswap V2 Router**: `0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D`
  - High-volume address with diverse tokens
  - Tests scalability

### BSC
- **PancakeSwap Router**: `0x10ED43C718714eb63d5aA57B78B54704E256024E`

### Polygon
- **QuickSwap Router**: `0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff`

## Post-Validation Checklist

Before merging PR #14:

- [ ] All validation script checks pass
- [ ] Manual database queries confirm correct token addresses
- [ ] Kafka messages have proper schema
- [ ] No errors in service logs
- [ ] Tested with at least 2 different wallets
- [ ] Tested with both ERC-20 and native ETH transactions
- [ ] Verified on multiple chains (Ethereum, BSC, or Polygon)

## Rollback Plan

If issues are discovered in staging:

1. Stop the stack: `docker-compose down`
2. Checkout previous stable commit: `git checkout main`
3. Rebuild and restart: `docker-compose up -d --build`
4. Report issues on PR #14

## Production Deployment Recommendations

After successful staging validation:

1. **Deploy with feature flag** - Consider adding `KAFKA_ENABLED=false` option for gradual rollout
2. **Monitor metrics** - Set up alerts for:
   - Kafka publish failures
   - Empty token_address rates
   - Consumer lag
3. **Database backup** - Take backup before deploying (though greenfield, good practice)
4. **Canary deployment** - Deploy to small percentage of traffic first

## Support

If you encounter issues during validation:
- Check service logs: `docker-compose logs -f [service-name]`
- Verify all environment variables are set correctly
- Ensure API keys have sufficient quota
- Report issues on PR #14 with logs and database query results
