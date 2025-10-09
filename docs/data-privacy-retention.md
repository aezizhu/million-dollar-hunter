# Data Privacy & Retention

## Scope
- No PII beyond email if/when multi-user is enabled; MVP stores no personal data.
- Primary data: on-chain analytics, watchlists, exports.

## Retention
- Application DB: retain portfolio and analytics data indefinitely unless owner deletes.
- Logs: 14–30 days depending on storage; redact sensitive fields.
- Exports: stored locally by owner; app does not retain.

## Data Rights
- Export formats: CSV/JSON for wallet, assets, and transactions.
- Deletion: Owner can remove wallets/assets; background purge jobs remove related data.

## Wallet Privacy
- Treat wallet addresses as sensitive; avoid sharing; mask in logs/URLs where feasible.
- Document that blockchain data is public by nature; app provides local privacy controls.

## Compliance
- GDPR/CCPA not in scope for single-user MVP; follow best practices regardless.
