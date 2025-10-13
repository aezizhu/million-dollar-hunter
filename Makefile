.PHONY: load-auth
load-auth:
	BASE_URL?=http://localhost:8080 AUTH_EMAIL?=test@example.com AUTH_PASSWORD?=password k6 run api-gateway/tests/k6/auth-load-test.js
