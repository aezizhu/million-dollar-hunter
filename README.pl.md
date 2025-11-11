# Million Dollar Hunter

Osobista platforma śledzenia portfela kryptowalut zaprojektowana do monitorowania tokenów blockchain i aktywności portfeli w wielu łańcuchach. Zbudowana z architekturą mikroserwisów wykorzystującą wzorzec CQRS do wysokowydajnego pozyskiwania i zapytań danych.

> *"Każdy wielki myśliwy wie, że cierpliwość i precyzja prowadzą do najcenniejszych odkryć."* — Stworzone z drobiazgową uwagą na wzorce architektoniczne i projektowanie systemów.

## Przegląd

Million Dollar Hunter to jednoużytkownikowy panel kryptowalutowy on-chain, który umożliwia osobom monitorowanie, zapytania i analizę tokenów blockchain oraz aktywności portfeli w czasie rzeczywistym. Platforma zapewnia głębokie analityki dla portfeli i tokenów w łańcuchach BSC, Solana, Ethereum i Polygon, z konfigurowalnymi panelami i kompleksowymi możliwościami eksportu danych.

System wykorzystuje architekturę opartą na mikroserwisach, która oddziela operacje zapisu (pozyskiwanie danych) od operacji odczytu (zapytania portfela) przy użyciu strumieniowania zdarzeń Kafka, umożliwiając niezależne skalowanie i optymalizację każdego komponentu.

## Szybki Start

Uruchom aplikację w 5 krokach:

### 1. Uruchomienie Usług Infrastruktury

```bash
cd ops
docker-compose up -d
```

To uruchamia PostgreSQL, Redis i Kafka przy użyciu Docker Compose.

### 2. Uruchomienie Migracji Bazy Danych

```bash
# Usługa uwierzytelniania
cd services/auth-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
migrate -path db/migrations -database "$DATABASE_URL" up

# Usługa portfela
cd services/portfolio-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
make migrate-up

# Powtórz dla market-data-service i ingestion-service w razie potrzeby
```

### 3. Uruchomienie Usług Backend

Otwórz osobne terminale dla każdej usługi:

```bash
# Terminal 1: Usługa uwierzytelniania
cd services/auth-service && make run

# Terminal 2: Usługa portfela
cd services/portfolio-service && make run

# Terminal 3: Usługa danych rynkowych
cd services/market-data-service && go run cmd/market-data-service/main.go

# Terminal 4: Usługa pozyskiwania
cd ingestion-service && go run cmd/ingestion-service/main.go

# Terminal 5: Brama API
cd api-gateway && make run
```

### 4. Uruchomienie Frontendu

```bash
cd frontend
npm install
npm run dev
```

### 5. Dostęp do Aplikacji

- **Frontend**: http://localhost:3000
- **Brama API**: http://localhost:8080
- **Sprawdzenie Zdrowia**: `curl http://localhost:8080/healthz`

**Dane Uwierzytelniające MVP**:
- Nazwa użytkownika: `aezi`
- Hasło: `Aa@123456789`

## Przegląd Architektury

### Mikroserwisy ze Wzorcem CQRS

Platforma implementuje Rozdzielenie Odpowiedzialności Zapytań i Poleceń (CQRS), oddzielając operacje zapisu (ingestion-service) od operacji odczytu (portfolio-service) przy użyciu strumieniowania zdarzeń Kafka.

### Główne Komponenty

1. **Frontend (Next.js 15)**
   - Nowoczesna aplikacja React 19 z App Router
   - Biblioteka komponentów Material-UI
   - TanStack Query do zarządzania stanem serwera
   - TypeScript z włączonym trybem ścisłym

2. **Brama API (Go/Fiber)**
   - Pojedynczy publiczny punkt wejścia dla wszystkich żądań klienta
   - Middleware uwierzytelniania JWT
   - Ograniczanie szybkości z token bucket Redis
   - Routing żądań do usług gRPC backend

3. **Usługa Uwierzytelniania (JWT/gRPC)**
   - Podwójny interfejs HTTP/gRPC
   - Generowanie i walidacja tokenów JWT
   - Rotacja tokenów odświeżania
   - Ochrona blokady logowania (3 niepowodzenia w 15 minut)

4. **Usługa Portfela (Model Odczytu CQRS)**
   - Serwer gRPC dla zapytań portfela
   - Konsument Kafka dla zdarzeń transakcji
   - Agreguje salda portfeli z historii transakcji
   - Zapewnia podsumowania portfela i funkcjonalność eksportu

5. **Usługa Danych Rynkowych (Integracja CoinGecko)**
   - Dane cen tokenów w czasie rzeczywistym z API CoinGecko
   - Cache Redis z TTL 60 sekund
   - Worker w tle do odświeżania cen
   - Interfejs gRPC dla zapytań cenowych

6. **Usługa Pozyskiwania (Model Zapisów CQRS)**
   - Pobiera dane blockchain z API Alchemy/Moralis
   - Publikuje zdarzenia transakcji do Kafka
   - Obsługuje śledzenie portfeli wielołańcuchowych
   - Wzorzec wyłącznika obwodu dla odporności API

### Przepływ Danych

1. **Usługa Pozyskiwania** pobiera transakcje blockchain z API Alchemy/Moralis
2. Publikuje zdarzenia `TransactionDataIngested` do Kafka
3. **Usługa Portfela** konsumuje zdarzenia i aktualizuje modele odczytu
4. **Usługa Danych Rynkowych** wzbogaca portfele o ceny tokenów w czasie rzeczywistym z CoinGecko
5. **Brama API** kieruje uwierzytelnione żądania i egzekwuje limity szybkości
6. **Usługa Uwierzytelniania** wydaje i waliduje tokeny JWT przez gRPC

## Stack Technologiczny

### Backend
- **Język**: Go 1.21+ (tryb workspace)
- **Bazy Danych**: PostgreSQL 15 (główna), Redis (cache, ograniczanie szybkości)
- **Komunikacja**: Apache Kafka (strumieniowanie zdarzeń)
- **API**: gRPC (między usługami), REST (publiczna brama)
- **Obserwowalność**: OpenTelemetry, Prometheus, strukturalne logowanie JSON (zerolog)

### Frontend
- **Framework**: Next.js 15 z App Router
- **Biblioteka UI**: React 19, TypeScript, Material-UI
- **Zarządzanie Stanem**: TanStack Query (stan serwera), React Context (stan UI)
- **Wykresy**: TradingView Lightweight Charts

### Zewnętrzne API
- **CoinGecko**: Dane rynkowe i ceny tokenów
- **Alchemy**: Dane transakcji blockchain (Ethereum, BSC, Polygon)
- **Moralis**: Salda portfeli wielołańcuchowych (Solana, zapasowe)

## Wymagania Wstępne

Przed rozpoczęciem rozwoju upewnij się, że masz zainstalowane następujące narzędzia:

- **Go 1.21+** (wsparcie trybu workspace)
- **Node.js 20+** i npm
- **Docker & Docker Compose** (dla zależności i testów integracyjnych)
- **Kompilator Protocol Buffers** (`protoc`) z wtyczkami Go:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **golang-migrate** CLI:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint** (do lintingu):
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **gofumpt** (do formatowania):
  ```bash
  go install mvdan.cc/gofumpt@latest
  ```
- **k6** (opcjonalne, do testów obciążeniowych):
  ```bash
  # macOS: brew install k6
  # Linux: Zobacz https://k6.io/docs/get-started/installation/
  ```

## Struktura Projektu

```
million-dollar-hunter/
├── api-gateway/                # Publiczna brama API (Go, Fiber)
│   ├── cmd/api-gateway/        # Główny punkt wejścia
│   ├── internal/               # Prywatny kod aplikacji
│   ├── tests/k6/               # Testy obciążeniowe dla ograniczania szybkości
│   └── Makefile
├── services/
│   ├── auth-service/           # Uwierzytelnianie JWT (Go, gRPC)
│   │   ├── cmd/auth-service/
│   │   ├── internal/           # Logika serwisu, magazyn, obsługi
│   │   ├── tests/              # Testy integracyjne z Testcontainers
│   │   ├── db/migrations/      # Migracje PostgreSQL
│   │   └── .golangci.yml       # Konfiguracja lintera
│   ├── portfolio-service/      # Model odczytu CQRS (Go, gRPC)
│   │   ├── cmd/server/
│   │   ├── internal/           # Serwis, repozytorium, konsument Kafka
│   │   └── migrations/         # Migracje bazy danych
│   ├── market-data-service/    # Integracja CoinGecko (Go, gRPC)
│   │   ├── cmd/market-data-service/
│   │   ├── internal/           # Cache, klient, obsługa, repozytorium, worker
│   │   └── docker-compose.yml  # Stos deweloperski lokalny
│   └── ingestion-service/      # Model zapisu CQRS (Go, HTTP)
│       ├── cmd/ingestion-service/
│       ├── internal/           # Klienci Alchemy/Moralis, ograniczanie szybkości
│       └── docker-compose.yml  # WireMock, Postgres, Redis, Kafka
├── frontend/                   # Next.js 15 App Router
│   ├── src/
│   │   ├── app/                # Strony App Router
│   │   ├── components/         # Komponenty React
│   │   ├── lib/                # Klient API, narzędzia
│   │   └── context/            # Kontekst React (auth, itp.)
│   ├── tsconfig.json           # Konfiguracja TypeScript (tryb ścisły)
│   └── eslint.config.mjs       # Płaska konfiguracja ESLint
├── proto/                      # Wspólne definicje protobuf
├── docs/                       # Architektura, ADRs, przewodniki wdrożenia
├── ops/                        # Pliki orkiestracji Docker compose
└── go.work                     # Definicja workspace Go
```

## Konfiguracja Środowiska

### Brama API

```bash
PORT=8080
JWT_SECRET=devsecret                    # Min 32 bajty w produkcji
REDIS_URL=localhost:6379
AUTH_SERVICE_URL=http://localhost:9000
PORTFOLIO_SERVICE_URL=localhost:8081    # gRPC
MARKET_DATA_SERVICE_URL=localhost:50051 # gRPC
AUTH_VALIDATE_MODE=grpc                 # lub "local"
AUTH_GRPC_ADDR=localhost:50051
FRONTEND_URL=http://localhost:3000      # Konfiguracja CORS
RATE_DEFAULT_RPS=10
RATE_DEFAULT_BURST=20
OPENAPI_PATH=../docs/openapi.yaml
```

**Uwaga Bezpieczeństwa CORS**: NIGDY nie używaj `FRONTEND_URL=*` z żądaniami z poświadczeniami. Zawsze określ dokładne źródła (oddzielone przecinkami dla wielu źródeł).

### Usługa Uwierzytelniania

```bash
PORT=8081
GRPC_PORT=50051
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
JWT_ISSUER=million-hunter
JWT_AUDIENCE=million-hunter-client
JWT_SIGNING_KEY=devsecret               # To samo co JWT_SECRET
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168               # 7 dni
ENABLE_MULTI_USER=false                 # Tryb MVP
PASSWORD_MIN_LENGTH=12
LOCKOUT_AFTER_FAILS=3
LOCKOUT_WINDOW_MIN=15
```

### Usługa Portfela

```bash
GRPC_ADDR=:50052
DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=portfolio-service
TOPIC_TRANSACTION_INGESTED=TransactionDataIngested
EXPORT_DIR=/tmp/exports
EXPORT_CLEANUP_TTL=1h
```

### Usługa Danych Rynkowych

```bash
GRPC_PORT=50051
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_TTL=60s
DATABASE_URL=postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable
COINGECKO_API_KEY=                      # Opcjonalne, zwiększa limity szybkości
COINGECKO_BASE_URL=https://api.coingecko.com/api/v3
COINGECKO_RATE_LIMIT=50                 # Żądania na minutę
WORKER_ENABLED=true
WORKER_REFRESH_INTERVAL=30s
```

### Usługa Pozyskiwania

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ingestion?sslmode=disable
REDIS_ADDR=localhost:6379
ALCHEMY_BASE_URL=https://eth-mainnet.g.alchemy.com/v2/
ALCHEMY_API_KEY=your_key_here
MORALIS_BASE_URL=https://deep-index.moralis.io/api/v2/
MORALIS_API_KEY=your_key_here
HTTP_PORT=8090
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_TX_INGESTED=TransactionDataIngested
```

### Frontend

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080  # Wskazuje na Bramę API
```

Zobacz pliki `.env.example` w każdym katalogu serwisu dla pełnej dokumentacji.

## Polecenia Budowania i Testowania

### Korzeń Repozytorium (Go Workspace)

Projekt używa workspace Go (`go.work`), który obejmuje wszystkie usługi Go:

```bash
# Synchronizuj wszystkie zależności modułów
go work sync

# Uruchom wszystkie testy w workspace
go test ./...
```

### Brama API

```bash
cd api-gateway
make build              # Zbuduj binarny
make test               # Uruchom testy
make validate-openapi   # Waliduj zgodność specyfikacji OpenAPI
make k6                 # Uruchom testy obciążeniowe
```

### Usługa Uwierzytelniania

```bash
cd services/auth-service
make check              # Formatuj, lint i uruchom krótkie testy (zalecane pre-commit)
make fmt                # Formatuj kod z gofumpt
make lint               # Uruchom golangci-lint
make test               # Uruchom wszystkie testy (wymaga Docker dla testów integracyjnych)
make test-short         # Uruchom tylko testy jednostkowe (pomiń integrację)
make build              # Zbuduj binarny do bin/auth-service
```

**Uwaga**: Testy integracyjne wymagają działającego daemona Docker (używa Testcontainers dla PostgreSQL).

### Usługa Portfela

```bash
cd services/portfolio-service
make build              # Zbuduj binarny do bin/portfolio-service
make proto              # Regeneruj kod gRPC z plików proto
make migrate-up         # Uruchom migracje bazy danych (wymaga DATABASE_URL)
make migrate-down       # Cofnij migracje
make run                # Uruchom serwis lokalnie
```

### Usługa Danych Rynkowych

```bash
cd services/market-data-service
make build              # Zbuduj binarny
go test ./internal/... -v                      # Testy jednostkowe
docker-compose up -d postgres redis            # Uruchom zależności
go test ./tests -v -run Integration            # Testy integracyjne
go test ./tests -v -run Load                   # Testy obciążeniowe
go test ./tests -bench=. -benchmem             # Benchmarki
```

### Usługa Pozyskiwania

```bash
cd ingestion-service
make build              # Zbuduj binarny
make test               # Uruchom testy (wymaga docker-compose up)
make lint               # Uruchom golangci-lint
make bench              # Uruchom benchmarki wydajności (cel ≥100 tx/s)
make up                 # Uruchom zależności docker-compose
make down               # Zatrzymaj i usuń kontenery
```

### Frontend (Next.js)

```bash
cd frontend
npm run dev             # Uruchom serwer deweloperski (port 3000)
npm run build           # Budowa produkcyjna
npm start               # Uruchom serwer produkcyjny
npm run lint            # Uruchom ESLint
npm test                # Uruchom testy Jest
npm run test:watch      # Uruchom testy w trybie watch
npm run test:coverage   # Wygeneruj raport pokrycia
```

**Uwaga**: Tryb ścisły TypeScript jest włączony. Alias ścieżki `@/*` mapuje na `./src/*`.

## Integracje Zewnętrznych API

### CoinGecko (Usługa Danych Rynkowych)

- **API**: `https://api.coingecko.com/api/v3/`
- **Limit Szybkości**: 50 żądań/min (warstwa darmowa), wyższy z kluczem API
- **Obsługiwane Łańcuchy**: BSC, Solana, Ethereum, Polygon
- **Cache**: Redis z TTL 60 sekund, celowanie w ≥80% wskaźnik trafień cache

### Alchemy (Usługa Pozyskiwania)

- **API**: Endpoint transferów aktywów z paginacją
- **Cel**: Pobierz transakcje blockchain (Ethereum, BSC, Polygon, Arbitrum, Optimism)
- **Ograniczanie Szybkości**: Algorytm token bucket z Redis
- **Funkcje**: Pojedynczy endpoint dla transferów ERC20, ERC721, ERC1155

### Moralis (Usługa Pozyskiwania)

- **API**: Salda portfeli wielołańcuchowych
- **Obsługiwane Łańcuchy**: BSC, Solana, Ethereum
- **Cel**: Dostawca zapasowy i wsparcie Solana
- **Ograniczanie Szybkości**: Wzorzec wyłącznika obwodu dla odporności

## Rozwój z Docker Compose

Dla zintegrowanego rozwoju lokalnego ze wszystkimi zależnościami i usługami:

```bash
# Uruchom pełny stos (z korzenia repozytorium)
docker-compose -f ops/docker-compose.yml up -d --build

# Zobacz logi
docker-compose -f ops/docker-compose.yml logs -f

# Zatrzymaj usługi
docker-compose -f ops/docker-compose.yml down

# Zatrzymaj i usuń wolumeny
docker-compose -f ops/docker-compose.yml down -v
```

Konfiguracja Docker Compose obejmuje:
- Wszystkie usługi backend (auth, portfolio, market-data, ingestion, api-gateway)
- Aplikacja frontend
- Infrastruktura (PostgreSQL, Redis, Kafka, Zookeeper)
- Kontrole zdrowia i automatyczne restartowanie

Konfiguracja jest zarządzana przez zmienne środowiskowe w `ops/docker-compose.yml`.

## Ważne Uwagi

### Tryb Go Workspace

Ten projekt używa trybu workspace Go (`go.work`) do zarządzania wieloma modułami Go w monorepo:

- Zawsze uruchamiaj `go work sync` z korzenia repozytorium po pobraniu zmian
- Każda usługa ma swój własny plik `go.mod`
- Tryb workspace umożliwia importy między usługami podczas rozwoju
- CI/CD buduje każdą usługę niezależnie z `GOWORK=off`

### Testy Integracyjne

Testy integracyjne wymagają Docker dla Testcontainers:

- Testy automatycznie uruchamiają kontenery PostgreSQL
- Użyj `make test-short` lub `go test -short` aby pominąć testy integracyjne
- Upewnij się, że daemon Docker działa przed uruchomieniem pełnego zestawu testów
- Konflikty portów mogą wystąpić, jeśli usługi już działają

### Bezpieczeństwo CORS

Brama API ustawia `Access-Control-Allow-Credentials: true`:

- **Produkcja**: Użyj określonego źródła (np. `https://app.million-hunter.com`)
- **Wiele Źródeł**: Lista oddzielona przecinkami
- **Rozwój**: `http://localhost:3000`
- **NIGDY**: Nie używaj `FRONTEND_URL=*` (narusza specyfikację CORS z poświadczeniami)

### Wzorzec CQRS

System implementuje CQRS dla optymalnej wydajności:

- **Model Zapisów** (Usługa Pozyskiwania): Zoptymalizowany dla wysokowydajnego pozyskiwania danych
- **Model Odczytu** (Usługa Portfela): Zoptymalizowany dla szybkich zapytań i agregacji
- **Strumieniowanie Zdarzeń**: Kafka łączy modele zapisu i odczytu
- **Przechowywanie Surowych Danych**: Przechowywanie JSONB umożliwia ponowne przetwarzanie bez ponownego pobierania

### Projekt MVP Jednego Użytkownika

Obecna implementacja jest zaprojektowana dla operacji jednego użytkownika:

- Zakodowane dane uwierzytelniające (nazwa użytkownika: `aezi`, hasło: `Aa@123456789`)
- Architektura JWT jest przygotowana, ale uproszczona dla MVP
- Pełne wsparcie wielu użytkowników można włączyć ustawiając `ENABLE_MULTI_USER=true`
- Brak obciążenia zgodności GDPR dla MVP

## Dodatkowa Dokumentacja

Aby uzyskać bardziej szczegółowe informacje, zapoznaj się z tymi plikami dokumentacji:

- **[AGENTS.md](AGENTS.md)** - Kompleksowe odniesienie dla deweloperów i briefing agenta AI
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Instrukcje wdrożenia i konfiguracja produkcyjna
- **[docs/TECHNICAL-DECISIONS.md](docs/TECHNICAL-DECISIONS.md)** - Decyzje architektoniczne i uzasadnienie
- **[docs/testing-strategy.md](docs/testing-strategy.md)** - Podejście do testowania i cele pokrycia
- **[TESTING-PHASE1-SUMMARY.md](TESTING-PHASE1-SUMMARY.md)** - Podsumowanie pokrycia testów
- **[docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md](docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md)** - Dokument wymagań produktu
- **[docs/external-api-integrations.md](docs/external-api-integrations.md)** - Szczegóły integracji zewnętrznych API

## Cele Wydajności

### Brama API
- Przepustowość: ≥100 żądań/s (pojedyncza instancja)
- Opóźnienie p95: ≤300ms (buforowane żądania)
- Dokładność limitu szybkości: <1% wskaźnik błędów

### Usługa Danych Rynkowych
- Wskaźnik trafień cache: ≥80% (okna 5-minutowe)
- Opóźnienie p95: ≤300ms (buforowane), ≤2s (brak trafienia cache + CoinGecko)
- Przepustowość: ≥100 żądań/s

### Usługa Pozyskiwania
- Przepustowość transakcji: ≥100 tx/s (model zapisu)
- Opóźnienie WireMock: <100ms (lokalne mocki)
- Wyłącznik obwodu zewnętrznego API: próg awarii 50%

### Usługa Portfela
- Opóźnienie modelu odczytu: <200ms p95
- Opóźnienie konsumenta Kafka: <10s przy normalnym obciążeniu

### Frontend
- Czas do Interaktywności (TTI): <3s (buforowane)
- Pierwsza Malowanie Treści (FCP): <1.5s
- Pokrycie testów: ≥70% instrukcji/gałęzi

## Współtworzenie

### Kontrole Pre-Commit

Przed zatwierdzeniem zmian:

1. Uruchom `make check` (auth-service) lub `make lint` + `make test` (inne usługi)
2. Upewnij się, że testy przechodzą: `make test` lub `go test ./...`
3. Frontend: `npm run lint` i `npm test`
4. Zweryfikuj brak niezatwierdzonych sekretów lub wrażliwych danych

### Konwencje Commitów

Użyj konwencjonalnych commitów dla jasności:

- `feat:` Nowa funkcja
- `fix:` Naprawa błędu
- `test:` Dodaj lub zaktualizuj testy
- `refactor:` Refaktoryzacja kodu bez zmian funkcjonalnych
- `docs:` Zmiany w dokumentacji
- `chore:` Zadania konserwacyjne (zależności, konfiguracja CI)

Przykład: `feat(auth): add refresh token rotation`

### Wymagania Pull Request

Każdy PR musi zawierać:

1. ✅ Wszystkie testy przechodzą (`make test` lub `npm test`)
2. ✅ Linting przechodzi (`make lint` lub `npm run lint`)
3. ✅ Sprawdzanie typów przechodzi (TypeScript: `npm run build`)
4. ✅ Diff ograniczony do odpowiednich plików (unikaj niepowiązanych zmian)
5. ✅ Dowód artefaktu (testy lub dowód testu ręcznego)
6. ✅ Opis jednego akapitu: Co się zmieniło, dlaczego i wszelkie pułapki
7. ✅ Brak spadku pokrycia (sprawdź raporty CI)

## Licencja

Umowa Podwójnej Licencji (Użycie Osobiste / Użycie Komercyjne)

Zobacz [LICENSE](LICENSE) dla pełnych warunków.

## Status Projektu

**Status**: Aktywny rozwój  
**Główny Język**: Go (backend), TypeScript (frontend)  
**Wdrożenie**: Docker Compose (lokalne/staging), Kubernetes (produkcja - planowane)

---

*Ta platforma reprezentuje kompleksowe podejście do analityki on-chain, gdzie każdy komponent został starannie zaprojektowany, aby zrównoważyć wydajność, utrzymywalność i skalowalność. Filozofia projektowania kładzie nacisk na separację odpowiedzialności, solidną obsługę błędów i praktyki rozwoju zorientowane na obserwowalność.*
