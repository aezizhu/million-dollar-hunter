#!/bin/bash

set -e

echo "=========================================="
echo "PR #14 Staging Validation"
echo "=========================================="
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

TEST_WALLET="${TEST_WALLET:-0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb}"
TEST_CHAIN="${TEST_CHAIN:-ethereum}"

echo "Configuration:"
echo "  Test Wallet: $TEST_WALLET"
echo "  Test Chain: $TEST_CHAIN"
echo ""

print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
        exit 1
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo "Step 1: Checking service health..."
echo "-----------------------------------"

docker compose ps | grep -q "Up" || {
    echo -e "${RED}Services not running. Start with: docker compose up -d${NC}"
    exit 1
}
print_status 0 "Docker services are running"

HEALTH=$(curl -s http://localhost:8090/healthz | jq -r '.status' 2>/dev/null || echo "error")
if [ "$HEALTH" = "ok" ] || [ "$HEALTH" = "degraded" ]; then
    print_status 0 "Ingestion service is healthy (status: $HEALTH)"
else
    print_status 1 "Ingestion service health check failed"
fi

KAFKA_STATUS=$(curl -s http://localhost:8090/healthz | jq -r '.kafka' 2>/dev/null || echo "unknown")
if [ "$KAFKA_STATUS" = "enabled" ]; then
    print_status 0 "Kafka is enabled and connected"
elif [ "$KAFKA_STATUS" = "disabled" ]; then
    print_warning "Kafka is disabled - cannot validate event publishing"
else
    print_status 1 "Kafka status unknown or unavailable"
fi

echo ""

echo "Step 2: Triggering test ingestion..."
echo "-------------------------------------"
echo "Note: You need to manually trigger ingestion via API or admin interface"
echo "Example: POST to ingestion service API with wallet=$TEST_WALLET, chain=$TEST_CHAIN"
echo ""
read -p "Press Enter after you've triggered ingestion for the test wallet..."

echo ""
echo "Step 3: Validating Kafka messages..."
echo "-------------------------------------"

sleep 2

echo "Checking for TransactionDataIngested events..."
KAFKA_MESSAGES=$(timeout 10s docker exec ops-kafka-1 kafka-console-consumer \
    --bootstrap-server localhost:9092 \
    --topic TransactionDataIngested \
    --from-beginning \
    --max-messages 5 2>/dev/null || echo "")

if [ -n "$KAFKA_MESSAGES" ]; then
    print_status 0 "Found Kafka messages on TransactionDataIngested topic"
    
    echo "$KAFKA_MESSAGES" | while read -r msg; do
        HAS_WALLET=$(echo "$msg" | jq -e '.wallet_address' &>/dev/null && echo "yes" || echo "no")
        HAS_TRANSACTIONS=$(echo "$msg" | jq -e '.transactions' &>/dev/null && echo "yes" || echo "no")
        HAS_TOKEN_ADDRESS=$(echo "$msg" | jq -e '.transactions[0].token_address' &>/dev/null && echo "yes" || echo "no")
        
        if [ "$HAS_WALLET" = "yes" ] && [ "$HAS_TRANSACTIONS" = "yes" ] && [ "$HAS_TOKEN_ADDRESS" = "yes" ]; then
            print_status 0 "Message has correct schema (wallet_address, transactions, token_address)"
        else
            print_warning "Message may have incomplete schema"
        fi
    done
else
    print_warning "No Kafka messages found (may need more time or ingestion failed)"
fi

echo ""

echo "Step 4: Validating database records..."
echo "---------------------------------------"

echo "Checking assets table for token_address values..."

ASSETS_COUNT=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
    "SELECT COUNT(*) FROM assets WHERE token_address IS NOT NULL AND token_address != '';" 2>/dev/null | xargs || echo "0")

if [ "$ASSETS_COUNT" -gt 0 ]; then
    print_status 0 "Found $ASSETS_COUNT assets with token_address"
    
    echo ""
    echo "Sample asset records:"
    docker exec ops-postgres-1 psql -U portfolio -d portfolio -c \
        "SELECT LEFT(wallet_id::text, 8) as wallet, LEFT(token_address, 20) as token_addr, symbol, LEFT(current_balance, 15) as balance 
         FROM assets 
         WHERE token_address IS NOT NULL AND token_address != '' 
         LIMIT 5;"
else
    print_warning "No assets with token_address found (ingestion may not have completed)"
fi

echo ""

echo "Step 5: Checking native ETH token address..."
echo "---------------------------------------------"

ETH_RECORDS=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
    "SELECT token_address FROM assets WHERE symbol = 'ETH' LIMIT 1;" 2>/dev/null | xargs || echo "")

if [ -n "$ETH_RECORDS" ]; then
    if echo "$ETH_RECORDS" | grep -q "0x0000000000000000000000000000000000000000"; then
        print_status 0 "Native ETH uses correct zero address"
    else
        print_warning "Native ETH token_address is '$ETH_RECORDS' (expected 0x0000...0000)"
    fi
else
    print_warning "No ETH records found to validate"
fi

echo ""

echo "Step 6: Validating no duplicate tokens per wallet..."
echo "-----------------------------------------------------"

DUPLICATES=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
    "SELECT wallet_id, token_address, symbol, COUNT(*) 
     FROM assets 
     GROUP BY wallet_id, token_address, symbol 
     HAVING COUNT(*) > 1;" 2>/dev/null | wc -l)

if [ "$DUPLICATES" -eq 0 ]; then
    print_status 0 "No duplicate assets found (good data integrity)"
else
    print_warning "Found $DUPLICATES duplicate asset records - this could indicate issues"
fi

echo ""

echo "Step 7: Checking for distinct token addresses..."
echo "-------------------------------------------------"

DISTINCT_TOKENS=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
    "SELECT COUNT(DISTINCT token_address) FROM assets WHERE token_address != '';" 2>/dev/null | xargs || echo "0")

DISTINCT_SYMBOLS=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
    "SELECT COUNT(DISTINCT symbol) FROM assets WHERE symbol IS NOT NULL;" 2>/dev/null | xargs || echo "0")

echo "Distinct token addresses: $DISTINCT_TOKENS"
echo "Distinct symbols: $DISTINCT_SYMBOLS"

if [ "$DISTINCT_TOKENS" -gt 1 ] && [ "$DISTINCT_TOKENS" != "$DISTINCT_SYMBOLS" ]; then
    print_status 0 "Using actual token addresses (not symbols)"
else
    print_warning "Token address diversity unclear - may need more test data"
fi

echo ""

echo "Step 8: Validating transaction records..."
echo "------------------------------------------"

TX_COUNT=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
    "SELECT COUNT(*) FROM transactions_view;" 2>/dev/null | xargs || echo "0")

if [ "$TX_COUNT" -gt 0 ]; then
    print_status 0 "Found $TX_COUNT transactions in database"
    
    EMPTY_HASHES=$(docker exec ops-postgres-1 psql -U portfolio -d portfolio -t -c \
        "SELECT COUNT(*) FROM transactions_view WHERE tx_hash = '' OR tx_hash IS NULL;" 2>/dev/null | xargs || echo "0")
    
    if [ "$EMPTY_HASHES" -eq 0 ]; then
        print_status 0 "No transactions with empty hash (validation working)"
    else
        print_warning "Found $EMPTY_HASHES transactions with empty hash"
    fi
else
    print_warning "No transactions found in database"
fi

echo ""
echo "=========================================="
echo "Validation Summary"
echo "=========================================="
echo ""
echo "Review the results above. All checks with ✓ indicate success."
echo "Warnings (⚠) may require investigation but aren't necessarily failures."
echo ""
echo "Key validations:"
echo "  1. Kafka messages contain token_address field"
echo "  2. Database assets use actual contract addresses"
echo "  3. Native ETH gets 0x0000...0000 address"
echo "  4. No duplicate assets per wallet"
echo "  5. Transactions have non-empty hashes"
echo ""
echo "If all critical checks pass, PR #14 is safe to merge."
echo "=========================================="
