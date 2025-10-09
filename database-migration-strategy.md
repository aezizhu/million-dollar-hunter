# Database Migration Strategy

## Tool Selection
- golang-migrate (migrate) is the standard tool.
- Rationale: battle-tested, works with PostgreSQL, simple CLI/CI usage, reversible up/down migrations, idempotent, container-friendly.

## Structure and Versioning
- Per-service migrations under: service-name/db/migrations
- File naming: NNN_description.up.sql and NNN_description.down.sql (NNN is zero-padded integer).
- One logical change per migration; avoid cross-service schema coupling.

## Execution
- Local: docker-compose runs migrate on service startup; repeatable and idempotent.
- CI: migration dry-run and apply steps before integration tests.
- Production: apply migrations during deployment with a pre-flight job; fail-fast on errors and avoid partial rollout.

## Rollbacks
- Use .down.sql to reverse changes.
- For destructive operations, require a two-step process: deploy with shadow columns/tables, backfill, then swap; rollback reverses the swap first.
- Always back up the database before running destructive migrations.

## Zero-Downtime Guidelines
- Backfill patterns and dual-write when necessary.
- Add columns as nullable or with safe defaults; populate in background; then add NOT NULL constraints.
- Avoid long exclusive locks; prefer online operations and batching.
- Use feature flags to gate code paths relying on new schema until backfill completes.

## Operational Safeguards
- Tag migration versions per release.
- Keep migrations immutable post-merge.
- Document data migrations separately from schema migrations.
## Future-State Schemas in MVP
- The auth-service users table is preparatory for multi-user and not required for MVP runtime.  
- Postpone enabling cross-service constraints that force usage until ENABLE_MULTI_USER=true.  
- Optionally seed a single admin user when transitioning to multi-user to validate auth flows and migrations.

## Holder Snapshots Schema
- Add holder_snapshots (ingestion-service) to support historical top holder tracking:
  - id BIGSERIAL PK
  - token_address TEXT
  - holder_address TEXT
  - balance NUMERIC
  - rank INT
  - timestamp TIMESTAMPTZ
- Migration guidance: create table with appropriate indexes on (token_address, timestamp) and (token_address, rank, timestamp) for time-series queries.
