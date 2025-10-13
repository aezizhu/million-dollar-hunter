# JWT Key Rotation Runbook

**Service**: auth-service  
**Last Updated**: 2025-10-13  
**Owner**: Security Team

---

## Overview

This runbook covers JWT signing key rotation procedures for the auth-service. The service supports RS256 (RSA asymmetric) keys with multi-key rotation capabilities, ensuring zero-downtime key changes.

### Key Concepts

- **kid (Key ID)**: Unique identifier for each signing key
- **Active Key**: The current key used to sign new JWTs
- **Grace Period**: 24 hours after expiration where old keys can still validate tokens
- **Cleanup Period**: 7 days after expiration when keys are permanently removed

---

## Standard Key Rotation (Planned)

Use this procedure for routine 90-day key rotations.

### Prerequisites

- Access to auth-service deployment
- `keytool` CLI installed
- Backup of current keystore file

### Steps

#### 1. Generate New Key

```bash
cd /home/ubuntu/repos/million-dollar-hunter/services/auth-service

# Generate a new 2048-bit RSA key (expires in 90 days)
./bin/keytool -keystore ./config/keystore.json -generate -bits 2048 -expires 2160h

# Example output:
# Generated key: a1b2c3d4e5f6g7h8
```

**Note**: The new key is NOT active yet. This allows testing before activation.

#### 2. Verify New Key

```bash
# List all keys
./bin/keytool -keystore ./config/keystore.json -list
```

Expected output:
```
KID               ALGORITHM  ACTIVE  CREATED     EXPIRES
a1b2c3d4e5f6g7h8  RS256              2025-10-13  2026-01-11
f9e8d7c6b5a4g3h2  RS256      ✓       2025-07-15  2025-10-13
```

#### 3. Test New Key (Optional but Recommended)

Create a test token with the new key in a non-production environment first.

#### 4. Activate New Key

```bash
# Activate the new key
./bin/keytool -keystore ./config/keystore.json -activate a1b2c3d4e5f6g7h8

# Output:
# Activated key: a1b2c3d4e5f6g7h8
```

**Impact**: New JWTs will now be signed with the new key. Old tokens remain valid during grace period.

#### 5. Monitor

Watch for validation errors:

```bash
# Check Prometheus metrics
curl http://localhost:8081/metrics | grep jwt_validation

# Expected: jwt_validation_errors should remain at 0
```

#### 6. Wait for Grace Period

**Duration**: 24 hours minimum (recommendation: wait for access token TTL + grace = 24h 15min)

During this period:
- Old key validates existing tokens
- New key signs new tokens
- No user impact

#### 7. Clean Up Expired Keys

After grace period + safety margin (7 days recommended):

```bash
./bin/keytool -keystore ./config/keystore.json -cleanup

# Output:
# Removed 1 expired keys
```

#### 8. Verify

```bash
./bin/keytool -keystore ./config/keystore.json -list
```

Only the new key should remain active.

---

## Emergency Key Rotation (Compromise)

Use this procedure if a signing key is compromised.

### Immediate Actions (T+0 minutes)

#### 1. Generate and Activate New Key Immediately

```bash
# Generate and activate in one step
./bin/keytool -keystore ./config/keystore.json -generate -bits 4096 -expires 2160h -active

# Output:
# Generated key: z9y8x7w6v5u4t3s2
# Key is now active
```

**Impact**: All existing tokens will expire within access token TTL (15 minutes).

#### 2. Force Service Restart (Optional)

If immediate token invalidation is required:

```bash
# Kubernetes
kubectl rollout restart deployment/auth-service

# Docker Compose
docker-compose restart auth-service

# Systemd
systemctl restart auth-service
```

#### 3. Revoke All Refresh Tokens

```sql
-- Connect to auth database
UPDATE refresh_tokens SET revoked = true;
```

**Impact**: All users must re-login.

#### 4. Monitor User Impact

```bash
# Watch login rate
curl http://localhost:8081/metrics | grep login_total

# Watch for support tickets
```

### Post-Incident Actions (T+24 hours)

#### 1. Update JWKS Endpoint

The JWKS endpoint automatically serves only valid keys. Verify:

```bash
curl http://localhost:8081/.well-known/jwks.json | jq .
```

Expected: Only the new key should be listed.

#### 2. Update API Gateway Configuration

If the API Gateway caches JWKS:

```bash
# Clear gateway JWKS cache
kubectl exec deployment/api-gateway -- kill -USR1 1

# Or restart gateway
kubectl rollout restart deployment/api-gateway
```

#### 3. Document Incident

Record in security incident log:
- Time of compromise detection
- Time of key rotation
- Number of affected users
- Root cause
- Remediation steps

#### 4. Review and Update

- Update this runbook if needed
- Test rotation procedure in staging
- Schedule next rotation

---

## Rollback Procedure

If the new key causes validation errors:

### Quick Rollback

```bash
# Reactivate the previous key
./bin/keytool -keystore ./config/keystore.json -list  # Find old kid

./bin/keytool -keystore ./config/keystore.json -activate <old-kid>
```

### Verification

```bash
# Check active key
./bin/keytool -keystore ./config/keystore.json -list

# Verify token generation
# Test login via API
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'
```

---

## Keystore Backup and Recovery

### Backup

```bash
# Backup keystore file
cp config/keystore.json config/keystore.backup.$(date +%Y%m%d).json

# Encrypt for storage (recommended)
gpg -c config/keystore.backup.$(date +%Y%m%d).json
```

**CRITICAL**: Keystore contains private keys. Store backups securely:
- Encrypt at rest
- Limit access (IAM policies)
- Store in secrets manager (production)

### Recovery

```bash
# Restore from backup
cp config/keystore.backup.20251013.json config/keystore.json

# Restart service
systemctl restart auth-service

# Verify
./bin/keytool -keystore ./config/keystore.json -list
```

---

## Monitoring and Alerting

### Key Metrics

Monitor these Prometheus metrics:

```
# Token generation rate (should be stable)
rate(jwt_tokens_generated_total[5m])

# Validation errors (should be near 0)
rate(jwt_validation_errors_total[5m])

# Active key age (alert if > 85 days)
jwt_active_key_age_days

# Keys in keystore (should be 1-2 normally)
jwt_keystore_keys_total
```

### Recommended Alerts

```yaml
# Prometheus alert rules
groups:
  - name: jwt_rotation
    rules:
      - alert: JWTKeyNearingExpiration
        expr: jwt_active_key_age_days > 85
        for: 24h
        annotations:
          summary: "JWT signing key expires in < 5 days"
          description: "Schedule key rotation"

      - alert: JWTValidationErrors
        expr: rate(jwt_validation_errors_total[5m]) > 0.01
        for: 10m
        annotations:
          summary: "High JWT validation error rate"
          description: "Check for key rotation issues"

      - alert: MultipleActiveKeys
        expr: jwt_active_keys_total > 1
        for: 1h
        annotations:
          summary: "Multiple active JWT keys detected"
          description: "Verify rotation completed successfully"
```

---

## Troubleshooting

### Problem: "no active signing key found"

**Cause**: No key marked as active, or active key is expired.

**Solution**:
```bash
# List keys
./bin/keytool -keystore ./config/keystore.json -list

# Activate a valid key
./bin/keytool -keystore ./config/keystore.json -activate <kid>
```

### Problem: "key not found: <kid>"

**Cause**: Token signed with a deleted key.

**Solution**:
- Tokens with deleted keys cannot be validated
- Users must re-login
- Consider extending cleanup period if this happens frequently

### Problem: High validation errors after rotation

**Cause**: Possible causes:
1. New key activated before generation completed
2. Keystore file corrupted
3. Clock skew between services

**Solution**:
```bash
# 1. Check keystore integrity
./bin/keytool -keystore ./config/keystore.json -list

# 2. Verify new key exists and is valid
# 3. Check server time
date -u

# 4. If corrupted, restore from backup
cp config/keystore.backup.*.json config/keystore.json
systemctl restart auth-service
```

### Problem: JWKS endpoint returns empty keys

**Cause**: All keys expired or no keys in keystore.

**Solution**:
```bash
# Generate new key immediately
./bin/keytool -keystore ./config/keystore.json -generate -bits 2048 -expires 2160h -active

# Restart service
systemctl restart auth-service
```

---

## Configuration Reference

### Keystore File Format

Location: `config/keystore.json`

```json
{
  "keys": [
    {
      "id": "a1b2c3d4e5f6g7h8",
      "algorithm": "RS256",
      "private_pem": "-----BEGIN RSA PRIVATE KEY-----\n...",
      "public_pem": "-----BEGIN PUBLIC KEY-----\n...",
      "created_at": "2025-10-13T00:00:00Z",
      "expires_at": "2026-01-11T00:00:00Z",
      "active": true
    }
  ]
}
```

### Environment Variables

```bash
# Keystore location
KEYSTORE_PATH=/path/to/keystore.json

# Key rotation settings (planned)
KEY_ROTATION_ENABLED=true
KEY_ROTATION_SCHEDULE="0 0 * * 0"  # Weekly on Sunday
KEY_DEFAULT_EXPIRY=90d
KEY_GRACE_PERIOD=24h
KEY_CLEANUP_PERIOD=168h  # 7 days
```

---

## Security Best Practices

1. **Regular Rotation**: Rotate keys every 90 days minimum
2. **Audit Logging**: Log all key generation, activation, and deletion
3. **Access Control**: Limit keystore file access to auth-service process only
4. **Backup Encryption**: Always encrypt keystore backups
5. **Monitoring**: Alert on key expiration before it happens
6. **Testing**: Test rotation in staging before production
7. **Documentation**: Keep this runbook updated after each rotation

---

## Appendix: Key Tool Command Reference

### Generate Key
```bash
./bin/keytool -keystore <path> -generate [-bits 2048|4096] [-expires <duration>] [-active]
```

### List Keys
```bash
./bin/keytool -keystore <path> -list
```

### Activate Key
```bash
./bin/keytool -keystore <path> -activate <kid>
```

### Cleanup Expired Keys
```bash
./bin/keytool -keystore <path> -cleanup
```

### Examples

```bash
# Generate 4096-bit key expiring in 1 year, active immediately
./bin/keytool -keystore ./config/keystore.json -generate -bits 4096 -expires 8760h -active

# List all keys in table format
./bin/keytool -keystore ./config/keystore.json -list

# Activate specific key
./bin/keytool -keystore ./config/keystore.json -activate a1b2c3d4e5f6g7h8

# Clean up keys expired > 7 days ago
./bin/keytool -keystore ./config/keystore.json -cleanup
```

---

**End of Runbook**

For questions or issues, contact: Security Team  
Incident Response: [On-call rotation]  
Documentation Repository: `/docs/JWT-ROTATION-RUNBOOK.md`
