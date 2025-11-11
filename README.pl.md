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

## Status Projektu

**Status**: Aktywny rozwój  
**Główny Język**: Go (backend), TypeScript (frontend)  
**Wdrożenie**: Docker Compose (lokalne/staging), Kubernetes (produkcja - planowane)

---

*Ta platforma reprezentuje kompleksowe podejście do analityki on-chain, gdzie każdy komponent został starannie zaprojektowany, aby zrównoważyć wydajność, utrzymywalność i skalowalność. Filozofia projektowania kładzie nacisk na separację odpowiedzialności, solidną obsługę błędów i praktyki rozwoju zorientowane na obserwowalność.*

