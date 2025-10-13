# JWT Key Management Tool (keytool)

Command-line utility for managing JWT signing keys in the auth-service keystore.

## Installation

```bash
# Build from source
cd services/auth-service
go build -o bin/keytool ./cmd/keytool

# Make executable
chmod +x bin/keytool
```

## Usage

### Generate New Key

```bash
./bin/keytool -keystore ./config/keystore.json -generate [options]

Options:
  -bits int
        RSA key size: 2048 or 4096 (default 2048)
  -expires duration
        Key expiration duration (default 2160h = 90 days)
  -active
        Make the new key active immediately (default false)
```

**Examples:**

```bash
# Generate 2048-bit key, expires in 90 days, not active
./bin/keytool -keystore ./config/keystore.json -generate

# Generate 4096-bit key, expires in 1 year, active immediately
./bin/keytool -keystore ./config/keystore.json -generate -bits 4096 -expires 8760h -active
```

### List Keys

```bash
./bin/keytool -keystore ./config/keystore.json -list
```

Output format:
```
KID               ALGORITHM  ACTIVE  CREATED     EXPIRES
fabc3e6829bc3153  RS256      ✓       2025-10-13  2026-01-11
a1b2c3d4e5f6g7h8  RS256              2025-07-15  2025-10-13
```

### Activate Key

```bash
./bin/keytool -keystore ./config/keystore.json -activate <kid>
```

**Effect**: Deactivates all other keys and activates the specified key. New JWTs will be signed with this key.

### Clean Up Expired Keys

```bash
./bin/keytool -keystore ./config/keystore.json -cleanup
```

**Effect**: Removes keys that expired more than 7 days ago (past grace period).

## Keystore File

The keystore is a JSON file containing RSA private/public key pairs:

**Location**: `./config/keystore.json` (default)

**Format**:
```json
{
  "keys": [
    {
      "id": "fabc3e6829bc3153",
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

**Security**:
- File permissions: `0600` (read/write owner only)
- Contains private keys - protect like database credentials
- Back up encrypted with GPG or store in secrets manager

## Key Rotation Workflow

### Planned Rotation (90 days)

```bash
# 1. Generate new key (not active yet)
./bin/keytool -keystore ./config/keystore.json -generate -bits 2048 -expires 2160h

# 2. Test new key (optional but recommended)
# Deploy to staging, generate test tokens

# 3. Activate new key
./bin/keytool -keystore ./config/keystore.json -activate <new-kid>

# 4. Wait 24+ hours (grace period)

# 5. Clean up old keys
./bin/keytool -keystore ./config/keystore.json -cleanup
```

### Emergency Rotation (compromise)

```bash
# Generate and activate immediately
./bin/keytool -keystore ./config/keystore.json -generate -bits 4096 -expires 2160h -active

# Impact: All existing tokens expire within 15 minutes (access token TTL)
```

See `/docs/JWT-ROTATION-RUNBOOK.md` for detailed procedures.

## Integration with Auth Service

The auth-service automatically loads keys from the keystore on startup.

**Environment Variable**:
```bash
KEYSTORE_PATH=./config/keystore.json
```

**Behavior**:
- If keystore doesn't exist, service runs in legacy mode (HS256)
- If keystore exists but has no active keys, service fails to start
- JWKS endpoint `/.well-known/jwks.json` serves public keys from keystore

**Reload**: Restart auth-service after key changes:
```bash
systemctl restart auth-service
# or
kubectl rollout restart deployment/auth-service
```

## Troubleshooting

### Error: "failed to load keystore: permission denied"

**Solution**: Fix file permissions
```bash
chmod 600 config/keystore.json
```

### Error: "key size must be at least 2048 bits"

**Solution**: Use `-bits 2048` or `-bits 4096`

### Error: "no active signing key found"

**Cause**: No key marked as active, or active key expired

**Solution**: Activate a valid key
```bash
./bin/keytool -keystore ./config/keystore.json -list
./bin/keytool -keystore ./config/keystore.json -activate <kid>
```

## Security Best Practices

1. **Permissions**: Set keystore file to `0600`
2. **Backups**: Encrypt with `gpg -c keystore.json`
3. **Rotation**: Rotate every 90 days minimum
4. **Key Size**: Use 2048 bits minimum, 4096 for sensitive deployments
5. **Audit**: Log all key operations
6. **Access**: Limit who can run keytool (production)

## See Also

- `/docs/JWT-ROTATION-RUNBOOK.md` - Operational procedures
- `/docs/SECURITY-HARDENING.md` - Overall security architecture
- `services/auth-service/internal/jwt/keystore.go` - Implementation
