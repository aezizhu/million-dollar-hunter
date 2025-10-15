GITLEAKS_VERSION ?= 8.24.3

.PHONY: secrets-scan
secrets-scan:
	@gitleaks detect -c .gitleaks.toml --redact --report-format json --report-path gitleaks-local.json || true
	@gitleaks detect -c .gitleaks.toml --redact --report-format sarif --report-path gitleaks-local.sarif || true
	@trufflehog filesystem --json . > trufflehog-local.json || true

.PHONY: pre-commit-install
pre-commit-install:
	@pre-commit install

.PHONY: load-auth
load-auth:
	BASE_URL?=http://localhost:8080 AUTH_EMAIL?=test@example.com AUTH_PASSWORD?=password k6 run api-gateway/tests/k6/auth-load-test.js
