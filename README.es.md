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

## Estado del Proyecto

**Estado**: Desarrollo activo  
**Lenguaje Principal**: Go (backend), TypeScript (frontend)  
**Despliegue**: Docker Compose (local/staging), Kubernetes (producción - planificado)

---

*Esta plataforma representa un enfoque integral para el análisis en cadena, donde cada componente ha sido cuidadosamente arquitecturado para equilibrar rendimiento, mantenibilidad y escalabilidad. La filosofía de diseño enfatiza la separación de preocupaciones, el manejo robusto de errores y las prácticas de desarrollo centradas en la observabilidad.*

