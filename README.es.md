# Million Dollar Hunter

Una plataforma de seguimiento de cartera de criptomonedas personal diseñada para monitorear tokens de blockchain y actividad de billeteras en múltiples cadenas. Construida con una arquitectura de microservicios usando el patrón CQRS para ingesta y consulta de datos de alto rendimiento.

> *"Todo gran cazador sabe que la paciencia y la precisión conducen a los descubrimientos más valiosos."* — Creado con meticulosa atención a los patrones arquitectónicos y el diseño de sistemas.

## Resumen

Million Dollar Hunter es un panel de criptomonedas en cadena de un solo usuario que permite a los individuos monitorear, consultar y analizar tokens de blockchain y actividad de billeteras en tiempo real. La plataforma proporciona análisis profundos para billeteras y tokens en las blockchains BSC, Solana, Ethereum y Polygon, con paneles personalizables y capacidades completas de exportación de datos.

El sistema utiliza una arquitectura basada en microservicios que separa las operaciones de escritura (ingesta de datos) de las operaciones de lectura (consultas de cartera) usando streaming de eventos Kafka, permitiendo escalado y optimización independiente de cada componente.

## Inicio Rápido

Ponga la aplicación en funcionamiento en 5 pasos:

### 1. Iniciar Servicios de Infraestructura

```bash
cd ops
docker-compose up -d
```

Esto inicia PostgreSQL, Redis y Kafka usando Docker Compose.

### 2. Ejecutar Migraciones de Base de Datos

```bash
# Servicio de autenticación
cd services/auth-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
migrate -path db/migrations -database "$DATABASE_URL" up

# Servicio de cartera
cd services/portfolio-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
make migrate-up

# Repetir para market-data-service e ingestion-service según sea necesario
```

### 3. Iniciar Servicios Backend

Abra terminales separadas para cada servicio:

```bash
# Terminal 1: Servicio de autenticación
cd services/auth-service && make run

# Terminal 2: Servicio de cartera
cd services/portfolio-service && make run

# Terminal 3: Servicio de datos de mercado
cd services/market-data-service && go run cmd/market-data-service/main.go

# Terminal 4: Servicio de ingesta
cd ingestion-service && go run cmd/ingestion-service/main.go

# Terminal 5: API Gateway
cd api-gateway && make run
```

### 4. Iniciar Frontend

```bash
cd frontend
npm install
npm run dev
```

### 5. Acceder a la Aplicación

- **Frontend**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **Verificación de Salud**: `curl http://localhost:8080/healthz`

**Credenciales de Autenticación MVP**:
- Nombre de usuario: `aezi`
- Contraseña: `Aa@123456789`

## Resumen de Arquitectura

### Microservicios con Patrón CQRS

La plataforma implementa Separación de Responsabilidades de Comando y Consulta (CQRS), separando las operaciones de escritura (ingestion-service) de las operaciones de lectura (portfolio-service) usando streaming de eventos Kafka.

### Componentes Principales

1. **Frontend (Next.js 15)**
   - Aplicación moderna React 19 con App Router
   - Biblioteca de componentes Material-UI
   - TanStack Query para gestión de estado del servidor
   - TypeScript con modo estricto habilitado

2. **API Gateway (Go/Fiber)**
   - Punto de entrada público único para todas las solicitudes del cliente
   - Middleware de autenticación JWT
   - Limitación de velocidad con token bucket de Redis
   - Enrutamiento de solicitudes a servicios gRPC backend

3. **Servicio de Autenticación (JWT/gRPC)**
   - Interfaz dual HTTP/gRPC
   - Generación y validación de tokens JWT
   - Rotación de tokens de actualización
   - Protección de bloqueo de inicio de sesión (3 fallos en 15 minutos)

4. **Servicio de Cartera (Modelo de Lectura CQRS)**
   - Servidor gRPC para consultas de cartera
   - Consumidor Kafka para eventos de transacciones
   - Agrega balances de billeteras del historial de transacciones
   - Proporciona resúmenes de cartera y funcionalidad de exportación

5. **Servicio de Datos de Mercado (Integración CoinGecko)**
   - Datos de precios de tokens en tiempo real de la API de CoinGecko
   - Caché Redis con TTL de 60 segundos
   - Worker en segundo plano para actualización de precios
   - Interfaz gRPC para consultas de precios

6. **Servicio de Ingesta (Modelo de Escritura CQRS)**
   - Obtiene datos de blockchain de las APIs de Alchemy/Moralis
   - Publica eventos de transacciones en Kafka
   - Maneja seguimiento de billeteras multi-cadena
   - Patrón de interruptor de circuito para resiliencia de API

### Flujo de Datos

1. **Servicio de Ingesta** obtiene transacciones de blockchain de las APIs de Alchemy/Moralis
2. Publica eventos `TransactionDataIngested` en Kafka
3. **Servicio de Cartera** consume eventos y actualiza modelos de lectura
4. **Servicio de Datos de Mercado** enriquece carteras con precios de tokens en tiempo real de CoinGecko
5. **API Gateway** enruta solicitudes autenticadas y aplica límites de velocidad
6. **Servicio de Autenticación** emite y valida tokens JWT vía gRPC

## Stack Tecnológico

### Backend
- **Lenguaje**: Go 1.21+ (modo workspace)
- **Bases de Datos**: PostgreSQL 15 (principal), Redis (caché, limitación de velocidad)
- **Mensajería**: Apache Kafka (streaming de eventos)
- **APIs**: gRPC (inter-servicio), REST (puerta de enlace pública)
- **Observabilidad**: OpenTelemetry, Prometheus, registro estructurado JSON (zerolog)

### Frontend
- **Framework**: Next.js 15 con App Router
- **Biblioteca UI**: React 19, TypeScript, Material-UI
- **Gestión de Estado**: TanStack Query (estado del servidor), React Context (estado UI)
- **Gráficos**: TradingView Lightweight Charts

### APIs Externas
- **CoinGecko**: Datos de mercado y precios de tokens
- **Alchemy**: Datos de transacciones blockchain (Ethereum, BSC, Polygon)
- **Moralis**: Balances de billeteras multi-cadena (Solana, respaldo)

## Prerrequisitos

Antes de comenzar el desarrollo, asegúrese de tener instaladas las siguientes herramientas:

- **Go 1.21+** (soporte de modo workspace)
- **Node.js 20+** y npm
- **Docker & Docker Compose** (para dependencias y pruebas de integración)
- **Compilador Protocol Buffers** (`protoc`) con plugins de Go:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **golang-migrate** CLI:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint** (para linting):
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **gofumpt** (para formateo):
  ```bash
  go install mvdan.cc/gofumpt@latest
  ```
- **k6** (opcional, para pruebas de carga):
  ```bash
  # macOS: brew install k6
  # Linux: Ver https://k6.io/docs/get-started/installation/
  ```

## Estructura del Proyecto

```
million-dollar-hunter/
├── api-gateway/                # Puerta de enlace API pública (Go, Fiber)
│   ├── cmd/api-gateway/        # Punto de entrada principal
│   ├── internal/               # Código de aplicación privado
│   ├── tests/k6/               # Pruebas de carga para limitación de velocidad
│   └── Makefile
├── services/
│   ├── auth-service/           # Autenticación JWT (Go, gRPC)
│   │   ├── cmd/auth-service/
│   │   ├── internal/           # Lógica de servicio, almacenamiento, manejadores
│   │   ├── tests/              # Pruebas de integración con Testcontainers
│   │   ├── db/migrations/      # Migraciones PostgreSQL
│   │   └── .golangci.yml       # Configuración de linter
│   ├── portfolio-service/      # Modelo de lectura CQRS (Go, gRPC)
│   │   ├── cmd/server/
│   │   ├── internal/           # Servicio, repositorio, consumidor Kafka
│   │   └── migrations/         # Migraciones de base de datos
│   ├── market-data-service/    # Integración CoinGecko (Go, gRPC)
│   │   ├── cmd/market-data-service/
│   │   ├── internal/           # Caché, cliente, manejador, repositorio, worker
│   │   └── docker-compose.yml  # Stack de desarrollo local
│   └── ingestion-service/      # Modelo de escritura CQRS (Go, HTTP)
│       ├── cmd/ingestion-service/
│       ├── internal/           # Clientes Alchemy/Moralis, limitación de velocidad
│       └── docker-compose.yml  # WireMock, Postgres, Redis, Kafka
├── frontend/                   # Next.js 15 App Router
│   ├── src/
│   │   ├── app/                # Páginas App Router
│   │   ├── components/         # Componentes React
│   │   ├── lib/                # Cliente API, utilidades
│   │   └── context/            # Contexto React (auth, etc.)
│   ├── tsconfig.json           # Configuración TypeScript (modo estricto)
│   └── eslint.config.mjs       # Configuración ESLint plana
├── proto/                      # Definiciones protobuf compartidas
├── docs/                       # Arquitectura, ADRs, guías de despliegue
├── ops/                        # Archivos de orquestación Docker compose
└── go.work                     # Definición de workspace Go
```

## Configuración del Entorno

### API Gateway

```bash
PORT=8080
JWT_SECRET=devsecret                    # Mínimo 32 bytes en producción
REDIS_URL=localhost:6379
AUTH_SERVICE_URL=http://localhost:9000
PORTFOLIO_SERVICE_URL=localhost:8081    # gRPC
MARKET_DATA_SERVICE_URL=localhost:50051 # gRPC
AUTH_VALIDATE_MODE=grpc                 # o "local"
AUTH_GRPC_ADDR=localhost:50051
FRONTEND_URL=http://localhost:3000      # Configuración CORS
RATE_DEFAULT_RPS=10
RATE_DEFAULT_BURST=20
OPENAPI_PATH=../docs/openapi.yaml
```

**Nota de Seguridad CORS**: NUNCA use `FRONTEND_URL=*` con solicitudes con credenciales. Siempre especifique orígenes exactos (separados por comas para múltiples orígenes).

### Servicio de Autenticación

```bash
PORT=8081
GRPC_PORT=50051
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
JWT_ISSUER=million-hunter
JWT_AUDIENCE=million-hunter-client
JWT_SIGNING_KEY=devsecret               # Igual que JWT_SECRET
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168               # 7 días
ENABLE_MULTI_USER=false                 # Modo MVP
PASSWORD_MIN_LENGTH=12
LOCKOUT_AFTER_FAILS=3
LOCKOUT_WINDOW_MIN=15
```

### Servicio de Cartera

```bash
GRPC_ADDR=:50052
DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=portfolio-service
TOPIC_TRANSACTION_INGESTED=TransactionDataIngested
EXPORT_DIR=/tmp/exports
EXPORT_CLEANUP_TTL=1h
```

### Servicio de Datos de Mercado

```bash
GRPC_PORT=50051
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_TTL=60s
DATABASE_URL=postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable
COINGECKO_API_KEY=                      # Opcional, aumenta límites de velocidad
COINGECKO_BASE_URL=https://api.coingecko.com/api/v3
COINGECKO_RATE_LIMIT=50                 # Solicitudes por minuto
WORKER_ENABLED=true
WORKER_REFRESH_INTERVAL=30s
```

### Servicio de Ingesta

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
NEXT_PUBLIC_API_URL=http://localhost:8080  # Apunta a API Gateway
```

Consulte los archivos `.env.example` en cada directorio de servicio para documentación completa.

## Comandos de Construcción y Pruebas

### Raíz del Repositorio (Go Workspace)

El proyecto usa un workspace de Go (`go.work`) que incluye todos los servicios Go:

```bash
# Sincronizar todas las dependencias de módulos
go work sync

# Ejecutar todas las pruebas en el workspace
go test ./...
```

### API Gateway

```bash
cd api-gateway
make build              # Construir binario
make test               # Ejecutar pruebas
make validate-openapi   # Validar cumplimiento de especificación OpenAPI
make k6                 # Ejecutar pruebas de carga
```

### Servicio de Autenticación

```bash
cd services/auth-service
make check              # Formatear, lint y ejecutar pruebas cortas (recomendado pre-commit)
make fmt                # Formatear código con gofumpt
make lint               # Ejecutar golangci-lint
make test               # Ejecutar todas las pruebas (requiere Docker para pruebas de integración)
make test-short         # Ejecutar solo pruebas unitarias (omitir integración)
make build              # Construir binario a bin/auth-service
```

**Nota**: Las pruebas de integración requieren que el daemon de Docker esté en ejecución (usa Testcontainers para PostgreSQL).

### Servicio de Cartera

```bash
cd services/portfolio-service
make build              # Construir binario a bin/portfolio-service
make proto              # Regenerar código gRPC desde archivos proto
make migrate-up         # Ejecutar migraciones de base de datos (requiere DATABASE_URL)
make migrate-down       # Revertir migraciones
make run                # Ejecutar servicio localmente
```

### Servicio de Datos de Mercado

```bash
cd services/market-data-service
make build              # Construir binario
go test ./internal/... -v                      # Pruebas unitarias
docker-compose up -d postgres redis            # Iniciar dependencias
go test ./tests -v -run Integration            # Pruebas de integración
go test ./tests -v -run Load                   # Pruebas de carga
go test ./tests -bench=. -benchmem             # Benchmarks
```

### Servicio de Ingesta

```bash
cd ingestion-service
make build              # Construir binario
make test               # Ejecutar pruebas (requiere docker-compose up)
make lint               # Ejecutar golangci-lint
make bench              # Ejecutar benchmarks de rendimiento (objetivo ≥100 tx/s)
make up                 # Iniciar dependencias docker-compose
make down               # Detener y eliminar contenedores
```

### Frontend (Next.js)

```bash
cd frontend
npm run dev             # Iniciar servidor de desarrollo (puerto 3000)
npm run build           # Construcción de producción
npm start               # Iniciar servidor de producción
npm run lint            # Ejecutar ESLint
npm test                # Ejecutar pruebas Jest
npm run test:watch      # Ejecutar pruebas en modo watch
npm run test:coverage   # Generar informe de cobertura
```

**Nota**: El modo estricto de TypeScript está habilitado. El alias de ruta `@/*` se mapea a `./src/*`.

## Integraciones de API Externas

### CoinGecko (Servicio de Datos de Mercado)

- **API**: `https://api.coingecko.com/api/v3/`
- **Límite de Velocidad**: 50 req/min (nivel gratuito), mayor con clave API
- **Cadenas Soportadas**: BSC, Solana, Ethereum, Polygon
- **Caché**: Redis con TTL de 60 segundos, objetivo de tasa de aciertos de caché ≥80%

### Alchemy (Servicio de Ingesta)

- **API**: Endpoint de transferencias de activos con paginación
- **Propósito**: Obtener transacciones blockchain (Ethereum, BSC, Polygon, Arbitrum, Optimism)
- **Limitación de Velocidad**: Algoritmo de token bucket con Redis
- **Características**: Endpoint único para transferencias ERC20, ERC721, ERC1155

### Moralis (Servicio de Ingesta)

- **API**: Balances de billeteras multi-cadena
- **Cadenas Soportadas**: BSC, Solana, Ethereum
- **Propósito**: Proveedor de respaldo y soporte Solana
- **Limitación de Velocidad**: Patrón de interruptor de circuito para resiliencia

## Desarrollo con Docker Compose

Para desarrollo local integrado con todas las dependencias y servicios:

```bash
# Iniciar stack completo (desde la raíz del repositorio)
docker-compose -f ops/docker-compose.yml up -d --build

# Ver logs
docker-compose -f ops/docker-compose.yml logs -f

# Detener servicios
docker-compose -f ops/docker-compose.yml down

# Detener y eliminar volúmenes
docker-compose -f ops/docker-compose.yml down -v
```

La configuración de Docker Compose incluye:
- Todos los servicios backend (auth, portfolio, market-data, ingestion, api-gateway)
- Aplicación frontend
- Infraestructura (PostgreSQL, Redis, Kafka, Zookeeper)
- Verificaciones de salud y reinicios automáticos

La configuración se gestiona mediante variables de entorno en `ops/docker-compose.yml`.

## Notas Importantes

### Modo Go Workspace

Este proyecto usa el modo workspace de Go (`go.work`) para gestionar múltiples módulos Go en un monorepo:

- Siempre ejecute `go work sync` desde la raíz del repositorio después de obtener cambios
- Cada servicio tiene su propio archivo `go.mod`
- El modo workspace permite importaciones entre servicios durante el desarrollo
- CI/CD construye cada servicio independientemente con `GOWORK=off`

### Pruebas de Integración

Las pruebas de integración requieren Docker para Testcontainers:

- Las pruebas inician automáticamente contenedores PostgreSQL
- Use `make test-short` o `go test -short` para omitir pruebas de integración
- Asegúrese de que el daemon de Docker esté en ejecución antes de ejecutar la suite completa de pruebas
- Pueden ocurrir conflictos de puertos si los servicios ya están en ejecución

### Seguridad CORS

El API Gateway establece `Access-Control-Allow-Credentials: true`:

- **Producción**: Use origen específico (ej., `https://app.million-hunter.com`)
- **Múltiples Orígenes**: Lista separada por comas
- **Desarrollo**: `http://localhost:3000`
- **NUNCA**: Use `FRONTEND_URL=*` (viola la especificación CORS con credenciales)

### Patrón CQRS

El sistema implementa CQRS para un rendimiento óptimo:

- **Modelo de Escritura** (Servicio de Ingesta): Optimizado para ingesta de datos de alto rendimiento
- **Modelo de Lectura** (Servicio de Cartera): Optimizado para consultas y agregaciones rápidas
- **Streaming de Eventos**: Kafka conecta modelos de escritura y lectura
- **Almacenamiento de Datos Crudos**: El almacenamiento JSONB permite reprocesamiento sin re-obtener

### Diseño MVP de Usuario Único

La implementación actual está diseñada para operación de usuario único:

- Credenciales de autenticación codificadas (nombre de usuario: `aezi`, contraseña: `Aa@123456789`)
- La arquitectura JWT está preparada pero simplificada para MVP
- El soporte completo multi-usuario se puede habilitar estableciendo `ENABLE_MULTI_USER=true`
- Sin carga de cumplimiento GDPR para MVP

## Documentación Adicional

Para información más detallada, consulte estos archivos de documentación:

- **[AGENTS.md](AGENTS.md)** - Referencia completa para desarrolladores y resumen para agentes AI
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Instrucciones de despliegue y configuración de producción
- **[docs/TECHNICAL-DECISIONS.md](docs/TECHNICAL-DECISIONS.md)** - Decisiones arquitectónicas y justificación
- **[docs/testing-strategy.md](docs/testing-strategy.md)** - Enfoque de pruebas y objetivos de cobertura
- **[TESTING-PHASE1-SUMMARY.md](TESTING-PHASE1-SUMMARY.md)** - Resumen de cobertura de pruebas
- **[docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md](docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md)** - Documento de requisitos del producto
- **[docs/external-api-integrations.md](docs/external-api-integrations.md)** - Detalles de integración de API externas

## Objetivos de Rendimiento

### API Gateway
- Rendimiento: ≥100 req/s (instancia única)
- Latencia p95: ≤300ms (solicitudes en caché)
- Precisión de límite de velocidad: <1% tasa de error

### Servicio de Datos de Mercado
- Tasa de aciertos de caché: ≥80% (ventanas de 5 minutos)
- Latencia p95: ≤300ms (en caché), ≤2s (fallo de caché + CoinGecko)
- Rendimiento: ≥100 req/s

### Servicio de Ingesta
- Rendimiento de transacciones: ≥100 tx/s (modelo de escritura)
- Latencia WireMock: <100ms (mocks locales)
- Interruptor de circuito de API externa: umbral de fallo del 50%

### Servicio de Cartera
- Latencia del modelo de lectura: <200ms p95
- Retraso del consumidor Kafka: <10s bajo carga normal

### Frontend
- Tiempo hasta Interactivo (TTI): <3s (en caché)
- Primera Pintura de Contenido (FCP): <1.5s
- Cobertura de pruebas: ≥70% declaraciones/ramas

## Contribuir

### Verificaciones Pre-Commit

Antes de confirmar cambios:

1. Ejecute `make check` (auth-service) o `make lint` + `make test` (otros servicios)
2. Asegúrese de que las pruebas pasen: `make test` o `go test ./...`
3. Frontend: `npm run lint` y `npm test`
4. Verifique que no haya secretos o datos sensibles sin confirmar

### Convenciones de Commit

Use commits convencionales para claridad:

- `feat:` Nueva característica
- `fix:` Corrección de error
- `test:` Agregar o actualizar pruebas
- `refactor:` Refactorización de código sin cambios funcionales
- `docs:` Cambios en documentación
- `chore:` Tareas de mantenimiento (dependencias, configuración CI)

Ejemplo: `feat(auth): add refresh token rotation`

### Requisitos de Pull Request

Cada PR debe incluir:

1. ✅ Todas las pruebas pasando (`make test` o `npm test`)
2. ✅ Linting pasa (`make lint` o `npm run lint`)
3. ✅ Verificación de tipos pasa (TypeScript: `npm run build`)
4. ✅ Diff confinado a archivos relevantes (evitar cambios no relacionados)
5. ✅ Artefacto de prueba (pruebas o evidencia de prueba manual)
6. ✅ Descripción de un párrafo: Qué cambió, por qué y cualquier advertencia
7. ✅ Sin caída en cobertura (verificar informes CI)

## Licencia

Acuerdo de Licencia Dual (Uso Personal / Uso Comercial)

Consulte [LICENSE](LICENSE) para los términos completos.

## Estado del Proyecto

**Estado**: Desarrollo activo  
**Lenguaje Principal**: Go (backend), TypeScript (frontend)  
**Despliegue**: Docker Compose (local/staging), Kubernetes (producción - planificado)

---

*Esta plataforma representa un enfoque integral para el análisis en cadena, donde cada componente ha sido cuidadosamente arquitecturado para equilibrar rendimiento, mantenibilidad y escalabilidad. La filosofía de diseño enfatiza la separación de preocupaciones, el manejo robusto de errores y las prácticas de desarrollo centradas en la observabilidad.*
