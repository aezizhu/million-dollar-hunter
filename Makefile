GITLEAKS_VERSION ?= 8.24.3

secrets-scan:
	@gitleaks detect -c .gitleaks.toml --redact --report-format json --report-path gitleaks-local.json || true
	@gitleaks detect -c .gitleaks.toml --redact --report-format sarif --report-path gitleaks-local.sarif || true
	@trufflehog filesystem --json . > trufflehog-local.json || true

pre-commit-install:
	@pre-commit install
