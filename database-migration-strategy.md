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
