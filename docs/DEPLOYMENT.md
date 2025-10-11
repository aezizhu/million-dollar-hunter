# Deployment Runbook: Portfolio Service

## Overview
This runbook documents the required deployment steps for the Portfolio Service, including database migrations, protobuf compatibility, and staged rollout guidance for client changes.

## Prerequisites
- Access to the production/staging PostgreSQL instance
- Access to CI/CD or manual deploy tooling for the portfolio-service and API Gateway

## 1) Apply DB migrations BEFORE restarting portfolio-service
Migrations must be applied in order and BEFORE the portfolio-service is restarted to prevent runtime query errors.

Required migrations (as of this PR):
1. services/portfolio-service/migrations/0001_init.sql
2. services/portfolio-service/migrations/0002_add_token_address_to_transactions.sql
3. services/portfolio-service/migrations/0003_add_token_and_snapshot_columns.sql

These migrations ensure:
- wallets.user_id is present and indexed
- transactions_view has token_address
- asset_snapshots has wallet_id and token_address with supporting indexes

Example (replace with your migration tool/invocation):
- Run migrations up to head
- Verify indexes exist:
  - wallets(user_id)
  - transactions_view(wallet_id, token_address, ts)
  - asset_snapshots(wallet_id, token_address, ts DESC)

Only after successful migration apply, restart the portfolio-service.

## 2) Proto change and staged rollout
Proto change: GetWalletDetailsRequest now requires user_id to enforce ownership.

Wire compatibility:
- Added a new field; no renumbering. Older clients won’t populate user_id, so the service will deny requests missing user_id.

Rollout plan:
- Ensure API Gateway populates user_id from JWT claims when calling GetWalletDetails.
- Deploy Gateway first (no DB dependency).
- Validate Gateway requests include user_id in staging.
- Deploy portfolio-service after confirming Gateway clients pass user_id.
- Monitor 4xx for PermissionDenied to catch unexpected callers missing user_id.

## 3) Authorization enforcement summary
- Export and GetWalletDetails enforce ownership via repository.VerifyWalletOwnership(user_id, wallet_id/address).
- PermissionDenied is returned when the user does not own the wallet or user_id is missing.

## 4) Integration tests (recommended)
Add integration tests in services/portfolio-service with a test DB to cover:
- GetWalletDetails with mismatched user_id -> PermissionDenied
- Export with mismatched user_id -> PermissionDenied

## 5) Rollback
If errors occur:
- Rollback portfolio-service first (no schema change rollback needed if queries remain compatible).
- If required, revert to previous Gateway version.
- Schema is additive and indexed, so compatibility is preserved.

## 6) Operational notes
- Ensure logs capture PermissionDenied counts and wallet identifiers are not logged in full to avoid PII leakage.
- Maintain a dashboard for error rates (4xx/5xx) and latency for GetWalletDetails/Export.
