

# **Million $ Hunter: A Comprehensive Technical Development Plan**

## **I. Strategic Technical Vision & System Architecture**

This document presents the definitive technical development plan for the "Million $ Hunter" dashboard. It translates high-level product requirements into a comprehensive, actionable blueprint designed for execution by a distributed engineering team, including collaborative AI agents. The plan outlines a full-stack architecture, specifies technologies, defines data models, and establishes development standards to ensure the creation of a robust, scalable, and secure platform for advanced crypto-portfolio analysis.

### **A. Architectural Paradigm: A Microservices-Oriented Approach**

The fundamental architectural choice for the Million $ Hunter platform is a distributed system based on a microservices paradigm. This decision is driven by the multifaceted nature of the application's domain, which encompasses distinct functional areas such as user authentication, real-time market data aggregation, historical on-chain data ingestion, and complex portfolio analysis.1 A monolithic architecture would inevitably lead to a tightly coupled system that is difficult to scale, maintain, and evolve.

The microservices approach provides several key strategic advantages for this project:

* **Independent Scalability:** Core components, such as the data ingestion service responsible for fetching vast transaction histories, will experience highly variable and potentially massive loads. A microservice architecture allows this service to be scaled independently of user-facing components like the authentication or portfolio services, optimizing resource allocation and cost-efficiency.1  
* **Technological Flexibility:** Each service can be built using the most appropriate technology for its specific task. While Go is the standard for the backend, this architecture allows for future services to be written in other languages if a compelling reason arises, without requiring a rewrite of the entire system.2  
* **Improved Resilience and Fault Isolation:** By designing for failure, the system can tolerate the outage of a non-critical service without causing a complete platform failure.3 For example, if the market data service experiences a temporary disruption, the rest of the application can remain operational, perhaps displaying cached data or gracefully degrading the user experience. This aligns with the principle of building resilient systems using patterns like circuit breakers and bulkheads.3  
* **Team Autonomy and Development Velocity:** Small, focused teams can own one or more services, developing, testing, and deploying them independently. This reduces coordination overhead and accelerates the development lifecycle.2

While this paradigm introduces complexities such as network latency, distributed data management, and a larger security surface, these challenges are deliberately addressed through the selection of a high-performance technology stack and the implementation of robust design patterns. The choice of Go, for instance, directly mitigates concerns about performance and network latency due to its efficient, compiled nature and built-in concurrency primitives.1

### **B. Core Technology Stack & Justification**

The selection of each technology is a deliberate decision aimed at achieving a balance of performance, developer productivity, security, and long-term maintainability. The following table provides the definitive technology stack for the project.

**Table 1: Definitive Technology Stack**

| Category | Technology/Library | Version | Justification & Key Snippets |
| :---- | :---- | :---- | :---- |
| Backend Language | Go | 1.21+ | Selected for its high performance, first-class concurrency support via goroutines, and a strong standard library ideal for building efficient, low-latency web services.1 |
| Backend Framework | Standard Library (net/http), Gin | Latest | The standard library provides a robust foundation, while Gin is used for its high-performance HTTP routing, middleware support, and simplicity, which accelerates API development.7 |
| Frontend Framework | Next.js (App Router) | 15+ | Chosen for its powerful server-side rendering (SSR) capabilities, scalable project structure, and support for modern React features like Server Components, ensuring a fast and SEO-friendly user experience.9 |
| Database | PostgreSQL | 15+ | A proven, reliable relational database system offering data integrity through ACID compliance, which is critical for financial data applications. Its robust feature set supports complex queries and data models.11 |
| Caching | Redis | 7+ | A high-performance in-memory data store used for caching frequently accessed data, such as token prices and aggregated portfolio snapshots, to reduce database load and minimize API latency. It is also essential for managing external API rate limits.13 |
| Comms (Sync) | gRPC | Latest | Used for high-performance, low-latency, and strongly-typed synchronous communication between internal microservices. Its use of Protobuf ensures efficient binary serialization, outperforming traditional JSON-over-HTTP.1 |
| Comms (Async) | Apache Kafka | Latest | A distributed event streaming platform that serves as the backbone for asynchronous communication. It decouples services, enables event-driven workflows, and ensures system resilience by buffering messages.1 |
| Authentication | golang-jwt/jwt | v5+ | The de facto standard library for implementing JSON Web Tokens (JWT) in Go, providing a secure and compliant way to handle user authentication and authorization.14 |
| Logging | rs/zerolog | Latest | A high-performance, zero-allocation structured JSON logger. Its efficiency minimizes the performance overhead of logging, and its structured output is essential for effective log aggregation and analysis in a distributed system.17 |
| UI Components | Material-UI (MUI) | Latest | A comprehensive and customizable library of React components that enables the rapid development of a consistent, professional, and accessible user interface, adhering to established design principles.20 |
| Charting | TradingView Lightweight Charts | Latest | A specialized, high-performance charting library designed for financial data visualization. Its interactivity and small footprint make it ideal for displaying complex portfolio performance and price history data without degrading frontend performance.22 |
| Containerization | Docker | Latest | The industry standard for containerization, used to package each microservice and its dependencies into an isolated, portable unit. This ensures consistency across development, staging, and production environments.3 |
| Orchestration | Kubernetes | Latest | A powerful container orchestration platform that automates the deployment, scaling, and management of the containerized microservices. It provides essential features like service discovery, load balancing, and self-healing.3 |

### **C. System Component Diagram & Data Flow**

The architecture is composed of a user-facing Next.js application, an API Gateway that serves as the single entry point, and a collection of backend microservices, each with its own dedicated database and specific domain responsibility. External data providers are integrated via dedicated services to isolate dependencies and manage their unique contracts.

A typical request flow for viewing a wallet's details proceeds as follows:

1. The authenticated user's browser sends an HTTP GET request to the API Gateway.  
2. The API Gateway validates the user's JWT access token.  
3. Upon successful validation, the gateway routes the request to the portfolio-service.  
4. The portfolio-service queries its own database for the pre-aggregated wallet data (assets, historical snapshots).  
5. To enrich this data with real-time prices, the portfolio-service makes a synchronous gRPC call to the market-data-service.  
6. The market-data-service returns the latest cached prices for the requested tokens.  
7. The portfolio-service synthesizes the final response object and returns it to the API Gateway.  
8. The API Gateway forwards the response to the Next.js application, which then renders the detailed wallet analysis view for the user.

Asynchronous data ingestion operates in the background. For example, when a user adds a new wallet to track, an event is published. The ingestion-service consumes this event and begins the long-running process of fetching historical transaction data from Alchemy and Moralis, persisting the raw data into its own database.

### **D. Foundational Design Patterns for Resilience and Scalability**

To manage the complexities of a distributed system and ensure its long-term viability, a set of foundational design patterns will be strictly enforced across the architecture.

* **API Gateway:** All external client requests (from the Next.js frontend) will be routed through a single API Gateway.2 This gateway acts as a reverse proxy, providing a unified API surface and handling cross-cutting concerns such as JWT authentication, rate limiting, logging, and routing requests to the appropriate internal microservices. This pattern simplifies the client application by abstracting away the internal service topology and centralizes critical security and traffic management policies.26  
* **Database per Service:** This is a non-negotiable principle of the architecture. Each microservice will have exclusive ownership of its own database and schema.2 Direct database access between services is strictly forbidden; services must communicate only through their public APIs (gRPC or messaging). This enforces loose coupling, allowing each service's data model to evolve independently without creating cascading changes across the system.  
* **Command Query Responsibility Segregation (CQRS):** The system's data management will be explicitly divided into two distinct paths: commands (writes) and queries (reads).2 This pattern is a strategic necessity for the Million $ Hunter platform due to its fundamentally different data workloads. The "command" side of the system is the data ingestion pipeline, which is write-intensive, involves long-running processes, and handles vast amounts of raw historical data from external APIs like Alchemy and Moralis.28 The "query" side is the user-facing dashboard, which is read-intensive and demands low-latency responses to provide a fluid user experience. By separating these concerns, the ingestion-service can focus on efficiently writing normalized data to its database, while the portfolio-service can maintain denormalized, pre-aggregated "read models" specifically optimized for fast retrieval. This separation ensures that heavy data ingestion loads do not degrade the performance of the user dashboard.  
* **Saga Pattern:** For business processes that span multiple services, the Saga pattern will be used to maintain data consistency across distributed transactions.2 For example, the process of adding a new wallet involves several steps: creating a wallet record (portfolio-service), triggering a historical data fetch (ingestion-service), and calculating initial performance metrics. This workflow will be orchestrated as a saga, where each step is a local transaction that publishes an event upon completion, triggering the next step in the sequence. If a step fails, compensating transactions are executed to roll back the preceding changes. This event-driven approach avoids the need for complex and often unsupported two-phase commit protocols, ensuring eventual consistency in a resilient and decoupled manner.  
* **Resilience Patterns:** To build a fault-tolerant system, each microservice will implement a suite of resilience patterns.3 All network calls to other services or external APIs will be wrapped in a **Circuit Breaker** to prevent a single failing dependency from causing a cascading system failure. **Timeouts** will be strictly enforced on all outgoing requests to prevent indefinite blocking of resources. For transient failures, a **Retry** mechanism with **Exponential Backoff** will be implemented to allow the system to recover gracefully without overwhelming the downstream service.

## **II. Backend Development Specification: Go Microservices**

This section provides the definitive blueprint for the design and implementation of the backend microservices using Go. It establishes standards for project structure, development practices, service decomposition, communication, data persistence, security, and observability to ensure a consistent, high-quality, and maintainable codebase.

### **A. Canonical Project Structure and Development Standards**

Consistency in project structure is paramount for maintainability and developer onboarding. All Go microservices will adhere to the standard Go project layout, which provides a clear separation of concerns and enforces strong encapsulation principles.6

* **Project Layout:**  
  * cmd/: This directory will contain the entry point for the application, typically a main.go file within a subdirectory named after the service (e.g., cmd/auth-service/main.go). This structure allows for the potential inclusion of multiple binaries, such as a server and a CLI tool, within the same project.6  
  * internal/: This is where the vast majority of the application's core logic will reside. The Go toolchain prevents code within the internal directory from being imported by external projects, thereby enforcing strict encapsulation of the service's domain logic, API handlers, and data repositories.6  
  * pkg/: This directory is reserved for code that is intended to be shared and imported by other services within the project. This could include common utility functions or shared data transfer objects (DTOs), but its use will be minimized to avoid creating unnecessary coupling between services.6  
  * api/: This directory will house the formal API definitions, including .proto files for gRPC services and OpenAPI (Swagger) specifications for any RESTful interfaces exposed by the API Gateway.30  
* **Development Best Practices:** A rigorous set of development standards will be enforced through code reviews and automated CI checks to ensure the codebase is idiomatic, robust, and testable.  
  * **Error Handling:** All errors must be handled explicitly. The if err\!= nil pattern is mandatory. Errors returned from lower layers will be wrapped with additional context using fmt.Errorf("context: %w", err) to create a traceable error stack, which is invaluable for debugging.30  
  * **Dependency Injection:** Global state will be strictly avoided. All dependencies (such as database connections, loggers, or clients for other services) will be injected into components via their constructors. This practice, known as dependency inversion, decouples components and makes them significantly easier to test in isolation.30  
  * **Context Propagation:** The context.Context package will be used pervasively. It will be the first parameter in functions that cross process or API boundaries. This enables the propagation of request-scoped values, deadlines, and cancellation signals throughout the call chain, which is essential for building a resilient and efficient distributed system.30  
  * **Code Formatting and Linting:** The CI pipeline will automatically run go fmt and goimports to enforce a consistent code style. Additionally, golangci-lint will be used with a strict configuration to catch common programming errors, style inconsistencies, and potential performance issues before code is merged.25

### **B. Microservice Decomposition and Domain Boundaries**

The system is logically decomposed into five core microservices, each representing a distinct business capability or "bounded context." This clear separation of responsibilities is fundamental to the architecture's maintainability and scalability.3 The following matrix defines the scope and ownership of each service.

**Table 2: Microservice Domain & Responsibility Matrix**

| Service Name | Domain | Core Responsibilities | Data Ownership | Primary Communication |
| :---- | :---- | :---- | :---- | :---- |
| auth-service | User & Identity | Handles user registration, credentials verification, and the generation and validation of JWTs. Manages all user-related profile information. | Owns the users table, which stores user IDs, email addresses, hashed passwords, and profile metadata. | Exposes a gRPC interface for internal services. Communicates with the outside world via the API Gateway (REST). |
| portfolio-service | Portfolio Management & Querying | Manages the relationship between users and the wallets they track. Aggregates raw data into optimized "read models" for fast querying. Calculates portfolio metrics like net worth, profit & loss, and historical performance. | Owns the portfolios, wallets, assets, transactions\_view, and asset\_snapshots tables. These are denormalized, read-optimized data structures. | Exposes a gRPC interface for the API Gateway. Consumes events from the message broker to trigger data aggregation. |
| ingestion-service | External Data Ingestion & Normalization | Responsible for all communication with external blockchain data APIs (Alchemy, Moralis). Fetches historical transactions, transfers, and balances for tracked wallets. Normalizes the data from different sources into a canonical format and persists it. | Owns the raw\_transactions and raw\_balances tables, which serve as a write-optimized, durable store of the raw data fetched from external sources. | Primarily an asynchronous consumer of events (e.g., WalletTrackingRequested). Makes outbound REST/HTTP calls to external APIs. Publishes events upon completion (e.g., TransactionDataIngested). |
| market-data-service | Market Data Aggregation & Caching | Fetches token price data and other market metrics from CoinGecko. Caches this data in Redis and its own database to provide a fast, reliable internal price feed for other services and to avoid hitting external API rate limits. | Owns the token\_prices and market\_data tables, which function as a persistent cache for market information. | Exposes a gRPC interface for internal price queries. Makes outbound REST/HTTP calls to CoinGecko. |
| api-gateway | API Gateway & Aggregation | Acts as the single, public-facing entry point for the frontend application. Responsible for routing incoming requests, validating JWTs, enforcing rate limits, and sometimes aggregating responses from multiple internal services. | Stateless; does not own any data. | Exposes a public REST API. Communicates with internal services via gRPC and publishes events to the message broker. |

### **C. Inter-Service Communication Contracts**

The communication strategy employs a hybrid approach, selecting the optimal protocol based on the nature of the interaction.

* **Synchronous Communication (gRPC):** For request-response interactions that require an immediate answer, gRPC is the mandated protocol.1 Its use of HTTP/2 and Protobuf for schema definition provides a high-performance, low-latency communication channel with strong typing, which is superior to traditional REST for internal communication. For example, when the portfolio-service needs the current price of a token, it will make a direct, blocking gRPC call to the market-data-service.  
* **Asynchronous Communication (Message Broker):** To decouple services and handle long-running or broadcast-style operations, an event-driven approach using a message broker like Apache Kafka is essential.1 This pattern is the backbone of the CQRS and Saga implementations. When a user requests to track a new wallet, the api-gateway (or portfolio-service) publishes a WalletTrackingRequested event. The ingestion-service, as a subscriber to this event, picks up the task asynchronously. This decouples the user's immediate request from the potentially lengthy process of fetching historical data, improving responsiveness and resilience.2

### **D. Data Persistence: Database Schemas per Service (PostgreSQL)**

Each service's database schema is designed specifically for its purpose, reflecting the principles of financial data modeling and the CQRS pattern.4 Using PostgreSQL provides the necessary relational integrity and robust features for this domain.11

* **auth-service Schema:**  
  * users: id (UUID, PK), email (TEXT, UNIQUE), password\_hash (TEXT), created\_at (TIMESTAMPTZ), updated\_at (TIMESTAMPTZ). A simple and secure schema focused solely on user identity.  
* **portfolio-service Schema (Read-Optimized):**  
  * wallets: id (UUID, PK), user\_id (UUID, FK to auth.users), address (TEXT), chain (TEXT), nickname (TEXT).  
  * assets: id (UUID, PK), wallet\_id (UUID, FK to wallets), token\_address (TEXT), symbol (TEXT), name (TEXT), current\_balance (NUMERIC).  
  * asset\_snapshots: id (BIGSERIAL, PK), asset\_id (UUID, FK to assets), timestamp (TIMESTAMPTZ), balance (NUMERIC), usd\_value (NUMERIC). This table is crucial for generating historical portfolio value charts efficiently.  
  * transactions\_view: A denormalized table or materialized view designed for fast querying of a wallet's transaction history. It will contain aggregated and formatted data, such as timestamp, type (e.g., SEND, RECEIVE, SWAP), from\_address, to\_address, asset\_symbol, amount, usd\_value\_at\_time, transaction\_hash.  
* **ingestion-service Schema (Write-Optimized):**  
  * ingestion\_jobs: id (UUID, PK), wallet\_address (TEXT), chain (TEXT), status (TEXT), last\_run\_timestamp (TIMESTAMPTZ), cursor (TEXT). Manages the state of data fetching for each wallet.  
  * raw\_transactions: id (BIGSERIAL, PK), source\_api (TEXT), wallet\_address (TEXT), data (JSONB), ingested\_at (TIMESTAMPTZ). Stores the complete, unmodified JSON response from external APIs. This provides an idempotent data source that can be reprocessed if the normalization logic changes, without needing to re-fetch from the external API.  
* **market-data-service Schema:**  
  * token\_prices: token\_address (TEXT), chain (TEXT), usd\_price (NUMERIC), last\_updated (TIMESTAMPTZ). A simple table acting as a persistent cache for token prices.

### **E. Security Architecture: JWT Implementation Deep-Dive**

The platform's security model is centered around JWTs for stateless authentication. The auth-service is the single source of truth for user identity and token issuance. The golang-jwt/jwt library (v5 or later) will be used for its security features and compliance with RFC 7519\.14

* **Token Generation Flow:**  
  1. A user submits credentials (email/password) to the /api/v1/auth/login endpoint on the API Gateway.  
  2. The gateway forwards the request to the auth-service.  
  3. The auth-service validates the credentials against the users table.  
  4. Upon success, it generates two tokens:  
     * **Access Token:** A short-lived JWT (e.g., 15-minute expiry) containing claims such as user\_id, exp (expiration time), iat (issued at), and iss (issuer).34 This token is used to authenticate subsequent API requests.  
     * **Refresh Token:** A long-lived, opaque token stored in the database and sent to the client (typically in a secure, HttpOnly cookie). This token is used to obtain a new access token without requiring the user to log in again.  
* **Token Validation Flow:**  
  1. The frontend client includes the access token in the Authorization: Bearer \<token\> header of every protected API request.  
  2. The API Gateway intercepts the request and inspects the header.  
  3. It validates the token's signature using the public key of the auth-service. It also verifies that the token is not expired and was issued by the correct authority (iss claim).34  
  4. If the token is valid, the gateway extracts the user\_id claim and injects it into a downstream request header (e.g., X-User-ID) for the internal services to use for authorization.  
  5. If the token is invalid or expired, the gateway immediately rejects the request with a 401 Unauthorized status.  
* **Security Best Practices:**  
  * The signing key for the JWTs will be a strong, randomly generated secret (for HS256) or a private key (for RS256) and will be managed securely via environment variables or a secret management system. It will never be hardcoded in the source code.34  
  
#### MVP Authentication Reconciliation
For the single-user MVP, the public endpoints under /api/v1/auth are stubbed and the API Gateway enforces a simple login gate using a hardcoded admin credential matching the PRD. JWT issuance/validation and multi-user flows (register, refresh) are scaffolded but disabled in production builds until multi-user is introduced. This preserves the JWT-compatible architecture while keeping the MVP operational simplicity.
  * All communication between the client, gateway, and services will be encrypted using TLS.

### **F. System Observability: A Framework for Logging, Tracing, and Metrics**

In a distributed microservices architecture, robust observability is not an optional feature but a foundational requirement for operational stability and effective debugging. A three-pillar approach—logging, tracing, and metrics—will be implemented from the outset.

* **Structured Logging:** The zerolog library is mandated for all services due to its high performance and zero-allocation characteristics, which are critical in a high-throughput environment to prevent logging from becoming a performance bottleneck.17 All log output will be in JSON format, written to stdout.19 This standard allows for seamless collection and parsing by container orchestration platforms like Kubernetes and ingestion into centralized logging systems (e.g., Elasticsearch, Loki, Datadog).  
* **Log Context and Correlation:** To make logs useful in a distributed context, every log entry must contain a set of consistent contextual fields. A middleware in the API Gateway will generate a unique trace\_id for each incoming request. This trace\_id will be propagated to all downstream services via request context and included in every log message related to that request. Additional context, such as service\_name and user\_id, will also be added to the logger's context using zerolog.Ctx().30 This enables the ability to filter the centralized log store for a single trace\_id and see the complete, ordered lifecycle of a request across all services.  
* **Distributed Tracing:** While correlated logs provide a narrative, distributed tracing provides a performance profile. OpenTelemetry will be integrated into all services to trace the execution flow of requests.30 Spans will be created at critical boundaries: upon receiving a request at the API Gateway, before and after a gRPC call, and before and after a database query. This data, when visualized in a tool like Jaeger or Zipkin, allows developers to pinpoint latency bottlenecks within and between services.  
* **Metrics and Alerting:** Each service will expose a /metrics endpoint in the Prometheus exposition format.1 Key application-level metrics (the "RED" method: Rate, Errors, Duration) will be instrumented for all API endpoints. This includes request counts, error rates (by status code), and request latency histograms. This data will be scraped by a Prometheus server, enabling the creation of dashboards (e.g., in Grafana) to monitor the real-time health of the system and configure automated alerts for critical conditions, such as a spike in 5xx errors or a significant increase in API latency.30

## **III. Frontend Development Specification: Next.js Application**

This section outlines the architectural and implementation strategy for the Million $ Hunter frontend. The plan leverages the full capabilities of the Next.js App Router to build a performant, scalable, and maintainable user interface.

### **A. Scalable Project Architecture with Next.js App Router**

The project's structure is designed for large-scale applications, emphasizing modularity, clear separation of concerns, and ease of navigation for developers.9

* **Directory Structure:** A feature-centric, colocation-first approach will be adopted within the src/ directory.38  
  * src/app/: The core of the application, utilizing the App Router's file-system-based routing.  
    * src/app/(main)/dashboard/page.tsx: The main dashboard route. The (main) directory is a route group, allowing routes to share a layout without affecting the URL structure.10  
    * src/app/(main)/wallets/\[address\]/page.tsx: A dynamic route for displaying the detailed analysis of a specific wallet address.  
    * src/app/auth/: A dedicated directory for authentication-related pages like login and registration.9  
    * src/app/api/: For Next.js API Routes that serve as a backend-for-frontend (BFF), handling tasks like proxying requests to the main API Gateway or managing authentication sessions.  
  * src/components/: A directory for shared components.  
    * src/components/ui/: Contains general-purpose, reusable UI components (e.g., Button, Card, Modal) built upon Material-UI. These are the fundamental building blocks of the design system.9  
    * src/components/features/: Contains composite components that are specific to a business feature but may be reused across different pages (e.g., WalletList, TransactionTable).  
  * src/lib/: Houses non-component code, such as utility functions, API client configurations, and constants.9  
  * src/hooks/: For custom React hooks that encapsulate reusable logic, such as useAuth for managing user sessions or useWalletData for fetching portfolio information.9  
  * src/types/: Centralizes all TypeScript type definitions, especially for API request and response payloads, ensuring type safety across the application.9  
* **React Server Components (RSCs) vs. Client Components:** The application will strategically utilize RSCs to optimize performance. By default, all components in the App Router are Server Components. They are used for fetching data and rendering static UI on the server, which reduces the amount of JavaScript shipped to the client. Components requiring interactivity, state, or browser-only APIs (like the TradingView chart) will be explicitly marked with the 'use client' directive and kept as "leaf" components in the component tree to minimize the size of the client-side bundle.39

### **B. Design System and Component Hierarchy (Material-UI)**

A consistent and professional design is crucial for user trust and application usability. Material-UI (MUI) is selected for its comprehensive component library and robust theming capabilities.20

* **Integration with Next.js App Router:** MUI will be integrated following the official guidelines to ensure correct behavior with server-side rendering and the App Router.21 This involves wrapping the root layout in the AppRouterCacheProvider to handle the injection of CSS generated on the server into the client's \<head\>, preventing style flickering and ensuring a smooth initial page load.21  
* **Custom Theming:** A centralized theme will be defined in src/theme.ts. This file will be a client component ('use client') and will export a theme object created with MUI's createTheme function.21 This theme will specify the application's color palette (primary, secondary, error, etc.), typography settings (integrated with Next.js font optimization via CSS variables), and default component styles. The entire application will be wrapped in a ThemeProvider in the root layout to make this theme available to all components.  
* **Component Hierarchy:**  
  1. **Base Components:** These are the raw components provided by @mui/material (e.g., Button, TextField, Table).  
  2. **UI Components (src/components/ui/):** These are lightly styled wrappers around base components to enforce application-specific design standards. For example, a \<AppButton\> component might wrap MUI's \<Button\> to apply a consistent size and variant.  
  3. **Feature Components (src/components/features/):** These are more complex components that combine multiple UI components to fulfill a specific business purpose, such as a \<PortfolioValueCard\> that displays a wallet's net worth and a performance sparkline.

### **C. State Management Strategy for Complex Data Environments**

The application will handle two primary types of state: server state and client state. A clear distinction and the right tools for each are essential for a scalable architecture.

* **Server State Management (React Query / TanStack Query):** For all data fetched from the backend API, TanStack Query will be the designated tool. It is not merely a data-fetching library but a powerful server-state manager. It provides out-of-the-box solutions for caching, background refetching, optimistic updates, and pagination. This is perfectly suited for the application's needs, such as keeping wallet balances and transaction lists synchronized with the backend with minimal boilerplate code.  
* **Client State Management (React Context / Zustand):** For global UI state that is not persisted on the server—such as the current theme (light/dark mode), notification messages, or the state of a multi-step modal—a lightweight solution will be used. React's built-in Context API is the first choice for simple state sharing. If state logic becomes more complex, a minimal library like Zustand will be adopted to avoid the significant boilerplate associated with more traditional state management libraries like Redux.

### **D. Financial Data Visualization Module (TradingView Charts Integration)**

Effective data visualization is a core feature of the Million $ Hunter dashboard. TradingView's Lightweight Charts library is chosen for its high performance with large datasets, responsive design, and rich feature set tailored for financial analysis.22

* **Component Implementation:** The charting functionality will be encapsulated within a dedicated React component, e.g., src/components/charts/FinancialChart.tsx. This component will be responsible for initializing the chart, handling data updates, and managing user interactions.  
* **Server-Side Rendering (SSR) Handling:** Lightweight Charts is a client-side-only library that directly interacts with the browser's DOM and Canvas API.22 Therefore, it cannot be rendered on the server. To handle this, the FinancialChart component will be marked with the 'use client' directive. Furthermore, it will be loaded into pages using Next.js's dynamic import feature with SSR disabled. This ensures the component's code is only loaded and executed on the client side, preventing server-side errors.24  
  TypeScript  
  // In a page component (e.g., app/wallets/\[address\]/page.tsx)  
  import dynamic from 'next/dynamic';

  const FinancialChart \= dynamic(  
    () \=\> import('@/components/charts/FinancialChart'),  
    { ssr: false, loading: () \=\> \<p\>Loading chart...\</p\> }  
  );

* **Data Flow and Lifecycle:** The chart component will receive historical data (e.g., an array of { time, value } objects) as a prop. It will use a useRef to get a stable reference to the container div element. A useEffect hook will be used to manage the chart's lifecycle:  
  1. On the initial render, it will call createChart to initialize the chart instance within the referenced div.  
  2. It will add a series (e.g., addAreaSeries) and set the initial data.  
  3. It will set up a resize observer to ensure the chart responsively adjusts to its container's dimensions.  
  4. The useEffect cleanup function will call chart.remove() to properly dispose of the chart instance when the component unmounts, preventing memory leaks.24

### **E. Core Application Modules**

* **Unified Dashboard View (app/dashboard/page.tsx):** This will be the primary landing page for authenticated users. It will feature a grid or list of WalletCard components, each providing a high-level summary of a tracked wallet, including its address, nickname, current USD net worth, and a 24-hour performance metric. This data will be fetched using TanStack Query from the /api/v1/portfolios endpoint.  
* **In-Depth Wallet Analysis View (app/wallets/\[address\]/page.tsx):** This dynamic page will provide a comprehensive analysis of a single wallet. It will be composed of several feature components:  
  * A primary FinancialChart displaying the wallet's historical portfolio value over different timeframes (24h, 7d, 30d, All).  
  * An AssetHoldings table detailing all current ERC20 and NFT assets, including their quantity, current price, and total value.  
  * A TransactionHistory table that is paginated, searchable, and filterable, showing all historical transactions for the wallet.  
* **User Authentication and Session Flow (app/auth/):** This module will consist of pages for login, registration, and password recovery. It will use a dedicated useAuth hook to manage interactions with the Next.js API routes (e.g., /api/auth/login), which in turn will communicate with the backend auth-service. Client-side session management will be handled securely, storing JWTs in memory and using refresh tokens persisted in secure HttpOnly cookies to maintain sessions.

## **IV. API Contracts and Integrations**

This section provides the definitive specification for all API interactions, both with external third-party services and within the application via the API Gateway. These contracts are critical for enabling parallel development between the frontend and backend teams.

### **A. External API Integration Blueprints**

The system relies on three primary external data providers. The ingestion-service and market-data-service are responsible for encapsulating all interactions with these APIs, providing a stable internal interface that is insulated from external changes.

* **Alchemy (ingestion-service):** Alchemy will serve as the primary source for comprehensive transaction history. The alchemy\_getAssetTransfers endpoint is particularly valuable as it aggregates multiple types of transfers (external, internal, ERC20, ERC721, ERC1155) into a single, efficient API call.28 The service will be designed to handle the API's pagination mechanism, using the pageKey returned in each response to fetch subsequent pages of results until the entire history is retrieved.43 The service will also include logic to respect Alchemy's rate limits and implement appropriate backoff strategies.  
* **Moralis (ingestion-service):** Moralis will be used as a complementary data source, providing redundancy and access to specific high-level data points not easily available elsewhere. Key endpoints to be used include getWalletTokenBalancesPrice for a snapshot of all token balances with their current USD prices, getWalletNetWorth for a quick valuation, and its broad cross-chain support for future expansion to chains like Solana.29 The ingestion-service will contain adapters to normalize the data from both Alchemy and Moralis into a consistent internal format.  
* **CoinGecko (market-data-service):** CoinGecko is the designated provider for all cryptocurrency market data. The service will primarily use the /simple/price endpoint to fetch current prices for multiple tokens by their IDs and the /simple/token\_price/{id} endpoint to fetch prices by contract address on a specific chain.49 To optimize performance and adhere to API usage policies, the market-data-service will implement a robust caching layer (using Redis) with a Time-to-Live (TTL) that aligns with CoinGecko's data update frequency (e.g., 60 seconds), ensuring that repeated requests for the same token price within a short window are served from the cache.51

### **B. API Gateway: Public Endpoint Specification (REST)**

The API Gateway exposes a RESTful API to the Next.js frontend, providing a stable and secure interface. All endpoints, with the exception of those under /api/v1/auth, will be protected and require a valid JWT Bearer token in the Authorization header.

**Table 3: API Gateway Endpoint Definitions**

| Method | Endpoint | Description | Request Body | Success Response (200 OK) |
| :---- | :---- | :---- | :---- | :---- |
| POST | /api/v1/auth/register | Registers a new user account. | { "email": "string", "password": "string" } | { "userId": "uuid", "message": "User created successfully" } |
| POST | /api/v1/auth/login | Authenticates a user and returns access and refresh tokens. | { "email": "string", "password": "string" } | { "accessToken": "string", "refreshToken": "string" } |
| POST | /api/v1/auth/refresh | Obtains a new access token using a valid refresh token. | { "refreshToken": "string" } | { "accessToken": "string" } |
| GET | /api/v1/portfolios | Retrieves a summary list of all wallets being tracked by the authenticated user. | N/A | \`\` |
| POST | /api/v1/portfolios/wallets | Adds a new wallet address for the user to track. This initiates an asynchronous data ingestion process. | { "address": "string", "chain": "string", "nickname": "string" } | { "address": "string", "status": "tracking\_initiated" } |
| GET | /api/v1/portfolios/wallets/{address} | Fetches detailed, aggregated data for a specific tracked wallet. | N/A | { "address": "string", "chain": "string", "netWorthUsd": "number", "assets": \[...\], "historicalValue": \[...\] } |
| GET | /api/v1/portfolios/wallets/{address}/transactions | Retrieves a paginated list of historical transactions for a specific wallet. Supports filtering and sorting via query parameters. | N/A (Query params: page, limit, filterByType) | { "transactions": \[...\], "pagination": { "currentPage": "number", "totalPages": "number", "totalItems": "number" } } |
| GET | /api/v1/market/prices | Retrieves the current USD price for a list of token contract addresses on a specific chain. | N/A (Query params: chain, token\_addresses) | { "0x...": { "usd": 123.45 }, "0x...": { "usd": 678.90 } } |

## **V. Deployment and Operational Strategy**

This section details the methodology for packaging, deploying, and operating the Million $ Hunter platform in a production environment. The strategy prioritizes automation, reliability, and scalability from the ground up.

### **A. Containerization (Docker) and Orchestration (Kubernetes)**

The entire backend system will be built for a containerized environment to ensure consistency and portability.

* **Containerization (Docker):** Each microservice will have its own Dockerfile. A multi-stage build process will be used to create minimal, production-ready container images.3 The first stage will use a Go build image to compile the application binary, and the second stage will copy only the compiled binary into a lightweight base image (like alpine or distroless). This practice significantly reduces the image size and attack surface of the final container.  
* **Orchestration (Kubernetes):** The application will be deployed to a Kubernetes cluster.3 Kubernetes will be responsible for managing the entire lifecycle of the application's containers. Key Kubernetes resources to be used include:  
  * **Deployments:** To declare the desired state for each microservice, including the container image and number of replicas.  
  * **Services:** To provide stable network endpoints for inter-service communication and to expose the API Gateway to the internet via a LoadBalancer.  
  * **ConfigMaps and Secrets:** To externalize configuration and manage sensitive data like API keys and database credentials, keeping them out of the container images.  
  * **Horizontal Pod Autoscalers (HPAs):** To automatically scale services like the ingestion-service based on CPU or memory usage, ensuring the system can handle variable loads.

### **B. Continuous Integration & Deployment (CI/CD) Pipeline**

A fully automated CI/CD pipeline is essential for achieving development velocity and maintaining a high standard of quality. This pipeline will be implemented using a tool like GitHub Actions.

* **Continuous Integration (CI):** This stage will be triggered on every git push to any branch. The pipeline will execute the following automated steps for each microservice:  
  1. **Lint and Format:** Run golangci-lint and go fmt to enforce code quality and style standards.30  
  2. **Unit and Integration Testing:** Execute all tests to verify the correctness of the code. Integration tests will use test containers to spin up temporary database instances, ensuring tests run in an isolated and realistic environment.1  
  3. **Security Scanning:** Integrate static application security testing (SAST) tools to scan for common vulnerabilities in the code and its dependencies.  
  4. **Build and Push Image:** If all previous steps pass, the pipeline will build the Docker container image, tag it with the commit hash, and push it to a central container registry (e.g., Docker Hub, Google Container Registry).  
* **Continuous Deployment (CD):** This stage will be triggered by a merge into the main branch.  
  1. **Deploy to Staging:** The newly built container image will be automatically deployed to a dedicated staging environment that mirrors production. Automated end-to-end tests can be run against this environment.  
  2. **Manual Promotion to Production:** The pipeline will pause for a manual approval step before deploying to the production environment. This provides a final gate for quality assurance and allows for deployments to be scheduled during low-traffic periods. The deployment to production will use a rolling update strategy to ensure zero downtime.
  
### **C. Secrets and Configuration Management**
- Tooling: Use sops + age for encrypting configuration files in-repo, and Kubernetes Secrets (or Docker Compose env files for local) for runtime injection. For cloud or future scaling, integrate with HashiCorp Vault or AWS Secrets Manager; interface via environment variables at runtime.
- Local Development: Provide .env.example files; actual .env kept out of VCS. Use docker-compose to inject env vars into services. For encrypted values, maintain config.enc.yaml managed by sops; decrypt during CI with repository-level age key.
- CI/CD: Store sensitive values as encrypted GitHub Actions secrets. CI jobs pass secrets as env vars to build/deploy steps only. Never commit secrets or keys.
- Rotation: Document key rotation procedures and short TTLs for API keys. Prefer per-environment credentials.
- Access Controls: Principle of least privilege for API keys (Alchemy, Moralis, CoinGecko). Network egress restricted at cluster level where applicable.

## **VI. Implementation Roadmap and Strategic Recommendations**

This section provides a phased implementation roadmap to guide the development process and offers key strategic recommendations to ensure the project's success and long-term health.

### **Implementation Roadmap**

The development will be structured in four distinct phases, allowing for iterative progress and the delivery of value at each stage.

* **Phase 1: Foundational Backend & Core Services (Weeks 1-4)**  
  * **Objective:** Establish the core architectural backbone and security layer.  
  * **Key Tasks:**  
    * Implement the auth-service, including user registration, login, and JWT generation/validation logic.  
    * Scaffold the api-gateway with basic request routing and JWT authentication middleware.  
    * Define and implement the initial PostgreSQL database schemas for all microservices using a migration tool.  
    * Migration Tool: Use golang-migrate (migrate) for schema versioning across services. Rationale: widely adopted in Go ecosystems, simple CLI/CI usage, supports PostgreSQL, reversible up/down migrations, and works well with containerized workflows. Migrations are stored per service under db/migrations with numbered up/down files and applied via CI and on container startup with idempotent execution.
    * Set up the foundational CI/CD pipeline to automate testing and container builds for the initial services.  
    * Establish the baseline Kubernetes configuration for deploying the core services.  
* **Phase 2: Data Ingestion and Processing (Weeks 5-8)**  
  * **Objective:** Build the data pipeline that powers the application.  
  * **Key Tasks:**  
    * Implement the market-data-service with CoinGecko API integration and a robust Redis caching layer.  
    * Develop the ingestion-service, focusing initially on the Alchemy alchemy\_getAssetTransfers integration for a single EVM chain (e.g., Ethereum).  
    * Set up the Apache Kafka message broker and implement the event-driven flow for the CQRS pattern (e.g., publishing a WalletTrackingRequested event and having the ingestion-service consume it).  
    * Implement the initial data normalization and persistence logic in the ingestion-service.  
* **Phase 3: Frontend Scaffolding and Core UI (Weeks 9-12)**  
  * **Objective:** Create the user-facing application and its primary views.  
  * **Key Tasks:**  
    * Initialize the Next.js project with the defined scalable folder structure.  
    * Integrate Material-UI and set up the custom application theme.  
    * Build the complete user authentication flow, including login, registration pages, and client-side session management.  
    * Develop the main dashboard and detailed wallet view pages with static or placeholder data, focusing on component structure and layout.  
* **Phase 4: End-to-End Integration and Visualization (Weeks 13-16)**  
  * **Objective:** Connect all components, implement data visualization, and prepare for launch.  
  * **Key Tasks:**  
    * Integrate the Next.js frontend with the live API Gateway endpoints, replacing all placeholder data with real data.  
    * Implement the TradingView Lightweight Charts component and feed it with historical data from the portfolio-service.  
    * Conduct thorough end-to-end testing across the entire system.  
    * Perform performance tuning on both the backend (database queries, API response times) and frontend (bundle size, rendering performance).  
    * Finalize documentation and prepare for the initial production launch.

### **Strategic Recommendations**

* **Prioritize Observability:** The implementation of the three pillars of observability—structured logging, distributed tracing, and metrics—should not be deferred. It must be treated as a core feature and built into every service from the very beginning. In a distributed system, the ability to debug and monitor effectively is a prerequisite for operational stability.  
* **Proactive External API Management:** The services that interact with external APIs (ingestion-service, market-data-service) must be designed defensively. Implement robust caching, request batching, and rate-limiting logic to stay within the usage quotas of providers like Alchemy, Moralis, and CoinGecko. This will prevent service disruptions and control operational costs.  
* **Adopt an Iterative Expansion Model:** The initial launch should focus on delivering a high-quality experience for a single, well-supported ecosystem (e.g., Ethereum Mainnet). Once the core platform is stable and has received user feedback, it can be expanded iteratively. This could include adding support for other EVM chains, integrating Solana data, or developing more advanced analytical features. This phased approach mitigates risk, allows the architecture to be validated in a real-world environment, and ensures that future development is guided by user needs.

#### **Works cited**

1. Building Microservices with Go: Practical Examples and Expert Tips \- Apriorit, accessed on October 9, 2025, [https://www.apriorit.com/dev-blog/building-microservices-with-golang](https://www.apriorit.com/dev-blog/building-microservices-with-golang)  
2. Microservice Architecture pattern \- Microservices.io, accessed on October 9, 2025, [https://microservices.io/patterns/microservices.html](https://microservices.io/patterns/microservices.html)  
3. 9 Best Practices for Building Microservices \- ByteByteGo, accessed on October 9, 2025, [https://bytebytego.com/guides/9-best-practices-for-building-microservices/](https://bytebytego.com/guides/9-best-practices-for-building-microservices/)  
4. Top 10 Microservices Design Patterns and How to Choose | Codefresh, accessed on October 9, 2025, [https://codefresh.io/learn/microservices/top-10-microservices-design-patterns-and-how-to-choose/](https://codefresh.io/learn/microservices/top-10-microservices-design-patterns-and-how-to-choose/)  
5. How to Write Microservices in Go \- Camunda, accessed on October 9, 2025, [https://camunda.com/resources/microservices/go/](https://camunda.com/resources/microservices/go/)  
6. Building Microservices with Go: A Step-by-Step Guide \- DEV ..., accessed on October 9, 2025, [https://dev.to/adi73/building-microservices-with-go-a-step-by-step-guide-5dla](https://dev.to/adi73/building-microservices-with-go-a-step-by-step-guide-5dla)  
7. JWT Authentication in Golang using Gin Web Framework-Tutorial, accessed on October 9, 2025, [https://www.golang.company/blog/jwt-authentication-in-golang-using-gin-web-framework](https://www.golang.company/blog/jwt-authentication-in-golang-using-gin-web-framework)  
8. Authentication in Golang and React using JWTs \- Auth0, accessed on October 9, 2025, [https://auth0.com/blog/authentication-in-golang/](https://auth0.com/blog/authentication-in-golang/)  
9. Mastering Next.js: Structuring Your Large-Scale Project for Success ..., accessed on October 9, 2025, [https://medium.com/@khalil.abid.tn/mastering-next-js-structuring-your-large-scale-project-for-success-6135773e21b0](https://medium.com/@khalil.abid.tn/mastering-next-js-structuring-your-large-scale-project-for-success-6135773e21b0)  
10. Getting Started: Project Structure | Next.js, accessed on October 9, 2025, [https://nextjs.org/docs/app/getting-started/project-structure](https://nextjs.org/docs/app/getting-started/project-structure)  
11. How To Build A Containerized Microservice in Golang: A Step-by-step Guide with Example Use-Case \- DEV Community, accessed on October 9, 2025, [https://dev.to/nikl/how-to-build-a-containerized-microservice-in-golang-a-step-by-step-guide-with-example-use-case-5ea8](https://dev.to/nikl/how-to-build-a-containerized-microservice-in-golang-a-step-by-step-guide-with-example-use-case-5ea8)  
12. How To Develop A Crypto Portfolio Tracker App? \- Idea Usher, accessed on October 9, 2025, [https://ideausher.com/blog/how-to-build-a-crypto-portfolio-tracker-app/](https://ideausher.com/blog/how-to-build-a-crypto-portfolio-tracker-app/)  
13. Go project layout for microservices : r/golang \- Reddit, accessed on October 9, 2025, [https://www.reddit.com/r/golang/comments/1j9vgw6/go\_project\_layout\_for\_microservices/](https://www.reddit.com/r/golang/comments/1j9vgw6/go_project_layout_for_microservices/)  
14. golang-jwt/jwt: Go implementation of JSON Web Tokens (JWT). \- GitHub, accessed on October 9, 2025, [https://github.com/golang-jwt/jwt](https://github.com/golang-jwt/jwt)  
15. Implementing JWT Authentication In Go \- Permify, accessed on October 9, 2025, [https://permify.co/post/jwt-authentication-go/](https://permify.co/post/jwt-authentication-go/)  
16. Creating a New JWT \- golang-jwt docs, accessed on October 9, 2025, [https://golang-jwt.github.io/jwt/usage/create/](https://golang-jwt.github.io/jwt/usage/create/)  
17. Logging in Go: A Comparison of the Top 9 Libraries | Better Stack Community, accessed on October 9, 2025, [https://betterstack.com/community/guides/logging/best-golang-logging-libraries/](https://betterstack.com/community/guides/logging/best-golang-logging-libraries/)  
18. rs/zerolog: Zero Allocation JSON Logger \- GitHub, accessed on October 9, 2025, [https://github.com/rs/zerolog](https://github.com/rs/zerolog)  
19. ZeroLog Tutorial \- Logging in Go \- Kelche, accessed on October 9, 2025, [https://www.kelche.co/blog/go/zerolog/](https://www.kelche.co/blog/go/zerolog/)  
20. Next.js App Router \- Toolpad Core \- MUI, accessed on October 9, 2025, [https://mui.com/toolpad/core/integrations/nextjs-approuter/](https://mui.com/toolpad/core/integrations/nextjs-approuter/)  
21. Next.js integration \- Material UI \- MUI, accessed on October 9, 2025, [https://mui.com/material-ui/integrations/nextjs/](https://mui.com/material-ui/integrations/nextjs/)  
22. Getting started | Lightweight Charts \- GitHub Pages, accessed on October 9, 2025, [https://tradingview.github.io/lightweight-charts/docs](https://tradingview.github.io/lightweight-charts/docs)  
23. Free Charting Library by TradingView, accessed on October 9, 2025, [https://www.tradingview.com/free-charting-libraries/](https://www.tradingview.com/free-charting-libraries/)  
24. Basic React example | Lightweight Charts \- GitHub Pages, accessed on October 9, 2025, [https://tradingview.github.io/lightweight-charts/tutorials/react/simple](https://tradingview.github.io/lightweight-charts/tutorials/react/simple)  
25. How Are You Structuring Your Go Microservices? | by Tai Vong \- Better Programming, accessed on October 9, 2025, [https://betterprogramming.pub/how-are-you-structuring-your-go-microservices-a355d6293932](https://betterprogramming.pub/how-are-you-structuring-your-go-microservices-a355d6293932)  
26. 19 Essential Microservices Patterns for System Design Interviews \- Design Gurus, accessed on October 9, 2025, [https://www.designgurus.io/blog/19-essential-microservices-patterns-for-system-design-interviews](https://www.designgurus.io/blog/19-essential-microservices-patterns-for-system-design-interviews)  
27. Go and Microservices : r/golang \- Reddit, accessed on October 9, 2025, [https://www.reddit.com/r/golang/comments/16fi4sb/go\_and\_microservices/](https://www.reddit.com/r/golang/comments/16fi4sb/go_and_microservices/)  
28. Transfers API \- Ethereum, Polygon, Optimism, Arbitrum \- Alchemy, accessed on October 9, 2025, [https://www.alchemy.com/transfers-api](https://www.alchemy.com/transfers-api)  
29. Wallet API | Moralis API Documentation, accessed on October 9, 2025, [https://docs.moralis.com/web3-data-api/evm/reference/wallet-api](https://docs.moralis.com/web3-data-api/evm/reference/wallet-api)  
30. Go Backend Development Best Practices for Microservices rule by Ehsan Davari, accessed on October 9, 2025, [https://cursor.directory/go-microservices](https://cursor.directory/go-microservices)  
31. Best Practices for Structuring Large Go Projects? \- Getting Help \- Go Forum, accessed on October 9, 2025, [https://forum.golangbridge.org/t/best-practices-for-structuring-large-go-projects/38392](https://forum.golangbridge.org/t/best-practices-for-structuring-large-go-projects/38392)  
32. Financial Data Modeling: Examples & Tips \- Solvexia, accessed on October 9, 2025, [https://www.solvexia.com/blog/financial-data-modeling](https://www.solvexia.com/blog/financial-data-modeling)  
33. golang-jwt docs, accessed on October 9, 2025, [https://golang-jwt.github.io/jwt/](https://golang-jwt.github.io/jwt/)  
34. Create Versatile Microservices in Golang — Part 4 (Authentication With JWT) \- DZone, accessed on October 9, 2025, [https://dzone.com/articles/create-versatile-microservices-in-golang-part-4-au](https://dzone.com/articles/create-versatile-microservices-in-golang-part-4-au)  
35. Securing a Go Microservice with JWT \- FusionAuth, accessed on October 9, 2025, [https://fusionauth.io/blog/securing-golang-microservice](https://fusionauth.io/blog/securing-golang-microservice)  
36. A Complete Guide to Logging in Go with Zerolog : r/golang \- Reddit, accessed on October 9, 2025, [https://www.reddit.com/r/golang/comments/x6hn15/a\_complete\_guide\_to\_logging\_in\_go\_with\_zerolog/](https://www.reddit.com/r/golang/comments/x6hn15/a_complete_guide_to_logging_in_go_with_zerolog/)  
37. Logging Errors in Go with ZeroLog: A Simple Guide \- Last9, accessed on October 9, 2025, [https://last9.io/blog/logging-errors-in-go-with-zerolog/](https://last9.io/blog/logging-errors-in-go-with-zerolog/)  
38. Next.js Enterprise Project Structure \- Dennis O'Keeffe, accessed on October 9, 2025, [https://www.dennisokeeffe.com/blog/2021-12-06-nextjs-enterprise-project-structure](https://www.dennisokeeffe.com/blog/2021-12-06-nextjs-enterprise-project-structure)  
39. Sharing my go-to project structure for Next.js \- colocation-first approach : r/nextjs \- Reddit, accessed on October 9, 2025, [https://www.reddit.com/r/nextjs/comments/1kkpqtm/sharing\_my\_goto\_project\_structure\_for\_nextjs/](https://www.reddit.com/r/nextjs/comments/1kkpqtm/sharing_my_goto_project_structure_for_nextjs/)  
40. Next.js App Router \- MUI Base, accessed on October 9, 2025, [https://mui.com/base-ui/guides/next-js-app-router/](https://mui.com/base-ui/guides/next-js-app-router/)  
41. Integrating Material UI into a React NextJS app | by Elana Olson | Medium, accessed on October 9, 2025, [https://medium.com/@elanaolson/integrating-material-ui-into-a-react-nextjs-app-55a95e15d767](https://medium.com/@elanaolson/integrating-material-ui-into-a-react-nextjs-app-55a95e15d767)  
42. Lightweight Charts™ library \- TradingView, accessed on October 9, 2025, [https://www.tradingview.com/lightweight-charts/](https://www.tradingview.com/lightweight-charts/)  
43. Transfers API Overview | Alchemy Docs, accessed on October 9, 2025, [https://www.alchemy.com/docs/reference/transfers-api-quickstart](https://www.alchemy.com/docs/reference/transfers-api-quickstart)  
44. alchemy\_getAssetTransfers | Alchemy Docs, accessed on October 9, 2025, [https://www.alchemy.com/docs/data/transfers-api/transfers-endpoints/alchemy-get-asset-transfers](https://www.alchemy.com/docs/data/transfers-api/transfers-endpoints/alchemy-get-asset-transfers)  
45. Guides to Wallet API \- Moralis Docs, accessed on October 9, 2025, [https://docs.moralis.com/web3-data-api/evm/wallet-api](https://docs.moralis.com/web3-data-api/evm/wallet-api)  
46. Crypto Wallet API for Web3 Dapps | Moralis for Developers, accessed on October 9, 2025, [https://moralis.com/api/wallet/](https://moralis.com/api/wallet/)  
47. Get Native & ERC20 Token Balances by Wallet | Moralis API Documentation, accessed on October 9, 2025, [https://docs.moralis.com/web3-data-api/evm/reference/wallet-api/get-wallet-token-balances-price](https://docs.moralis.com/web3-data-api/evm/reference/wallet-api/get-wallet-token-balances-price)  
48. Guides to Wallet API | Moralis API Documentation, accessed on October 9, 2025, [https://docs.moralis.io/web3-data-api/evm/wallet-api](https://docs.moralis.io/web3-data-api/evm/wallet-api)  
49. Coin Price by Token Addresses \- CoinGecko API, accessed on October 9, 2025, [https://docs.coingecko.com/reference/simple-token-price](https://docs.coingecko.com/reference/simple-token-price)  
50. Coin Price by IDs \- CoinGecko API, accessed on October 9, 2025, [https://docs.coingecko.com/reference/simple-price](https://docs.coingecko.com/reference/simple-price)  
51. Token Price by Token Addresses \- CoinGecko API, accessed on October 9, 2025, [https://docs.coingecko.com/v3.0.1/reference/onchain-simple-price](https://docs.coingecko.com/v3.0.1/reference/onchain-simple-price)  
52. Coin Price by Token Addresses \- CoinGecko API, accessed on October 9, 2025, [https://docs.coingecko.com/v3.0.1/reference/simple-token-price](https://docs.coingecko.com/v3.0.1/reference/simple-token-price)
