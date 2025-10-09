# Security Hardening

## Secrets Management
- sops + age for encrypted config; Kubernetes Secrets or env files for runtime.
- GitHub Actions secrets for CI; no secrets in repo.

## Network Security
- TLS in transit; service-to-service mTLS when moving to Kubernetes.
- Restrict egress by namespace/network policy; firewall DB to cluster-only.

## Container Security
- Base images: distroless/alpine; minimal surface.
- Image scanning in CI (Trivy, Grype); fail on high CVEs.

## Dependency Hygiene
- Go: govulncheck; Frontend: npm audit
- Renovate/Bot for dependency updates.

## Data Protection
- Encryption at rest via PostgreSQL storage-level or cloud KMS disks.
- Exported files optionally encrypted locally.

## AuthN/Z
- MVP gate with hardcoded credential at API Gateway.
- JWT scaffolding ready; rotate keys; short-lived tokens; audience/issuer checks when enabled.

## Logging & Privacy
- Avoid sensitive data in logs; mask secrets; structured logs only.
