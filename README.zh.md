# Million Dollar Hunter

一个用于监控区块链代币和钱包活动的个人加密货币投资组合跟踪平台。采用微服务架构，使用 CQRS 模式实现高性能数据摄取和查询。

> *"每一位优秀的猎手都知道，耐心和精准会带来最有价值的发现。"* — 以对架构模式和系统设计的细致关注精心打造。

## 概述

Million Dollar Hunter 是一个单用户链上加密货币仪表板，使个人能够实时监控、查询和分析区块链代币和钱包活动。该平台为 BSC、Solana、Ethereum 和 Polygon 区块链上的钱包和代币提供深度分析，具有可自定义的仪表板和全面的数据导出功能。

系统采用基于微服务的架构，使用 Kafka 事件流将写操作（数据摄取）与读操作（投资组合查询）分离，使每个组件能够独立扩展和优化。

## 快速开始

通过 5 个步骤启动应用程序：

### 1. 启动基础设施服务

```bash
cd ops
docker-compose up -d
```

这将启动 PostgreSQL、Redis 和 Kafka。

### 2. 运行数据库迁移

```bash
# 认证服务
cd services/auth-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
migrate -path db/migrations -database "$DATABASE_URL" up

# 投资组合服务
cd services/portfolio-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
make migrate-up

# 根据需要为 market-data-service 和 ingestion-service 重复上述步骤
```

### 3. 启动后端服务

为每个服务打开单独的终端：

```bash
# 终端 1: 认证服务
cd services/auth-service && make run

# 终端 2: 投资组合服务
cd services/portfolio-service && make run

# 终端 3: 市场数据服务
cd services/market-data-service && go run cmd/market-data-service/main.go

# 终端 4: 摄取服务
cd ingestion-service && go run cmd/ingestion-service/main.go

# 终端 5: API 网关
cd api-gateway && make run
```

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

### 5. 访问应用程序

- **前端**: http://localhost:3000
- **API 网关**: http://localhost:8080
- **健康检查**: `curl http://localhost:8080/healthz`

**MVP 认证凭据**:
- 用户名: `aezi`
- 密码: `Aa@123456789`

## 架构概述

### 使用 CQRS 模式的微服务

该平台实现了命令查询责任分离（CQRS），使用 Kafka 事件流将写操作（ingestion-service）与读操作（portfolio-service）分离。

### 核心组件

1. **前端 (Next.js 15)**
   - 使用 App Router 的现代 React 19 应用程序
   - Material-UI 组件库
   - TanStack Query 用于服务器状态管理
   - 启用严格模式的 TypeScript

2. **API 网关 (Go/Fiber)**
   - 所有客户端请求的单一公共入口点
   - JWT 认证中间件
   - 使用 Redis 令牌桶进行速率限制
   - 请求路由到后端 gRPC 服务

3. **认证服务 (JWT/gRPC)**
   - 双 HTTP/gRPC 接口
   - JWT 令牌生成和验证
   - 刷新令牌轮换
   - 登录锁定保护（15 分钟内 3 次失败）

4. **投资组合服务 (CQRS 读模型)**
   - 用于投资组合查询的 gRPC 服务器
   - 用于交易事件的 Kafka 消费者
   - 从交易历史聚合钱包余额
   - 提供投资组合摘要和导出功能

5. **市场数据服务 (CoinGecko 集成)**
   - 来自 CoinGecko API 的实时代币价格数据
   - 60 秒 TTL 的 Redis 缓存
   - 用于价格刷新的后台工作器
   - 用于价格查询的 gRPC 接口

6. **摄取服务 (CQRS 写模型)**
   - 从 Alchemy/Moralis API 获取区块链数据
   - 将交易事件发布到 Kafka
   - 处理多链钱包跟踪
   - 用于 API 弹性的断路器模式

### 数据流

1. **摄取服务** 从 Alchemy/Moralis API 获取区块链交易
2. 将 `TransactionDataIngested` 事件发布到 Kafka
3. **投资组合服务** 消费事件并更新读模型
4. **市场数据服务** 使用来自 CoinGecko 的实时代币价格丰富投资组合
5. **API 网关** 路由经过认证的请求并强制执行速率限制
6. **认证服务** 通过 gRPC 颁发和验证 JWT 令牌

## 技术栈

### 后端
- **语言**: Go 1.21+（工作区模式）
- **数据库**: PostgreSQL 15（主数据库），Redis（缓存、速率限制）
- **消息传递**: Apache Kafka（事件流）
- **API**: gRPC（服务间），REST（公共网关）
- **可观测性**: OpenTelemetry、Prometheus、结构化 JSON 日志（zerolog）

### 前端
- **框架**: 使用 App Router 的 Next.js 15
- **UI 库**: React 19、TypeScript、Material-UI
- **状态管理**: TanStack Query（服务器状态）、React Context（UI 状态）
- **图表**: TradingView Lightweight Charts

### 外部 API
- **CoinGecko**: 市场数据和代币价格
- **Alchemy**: 区块链交易数据（Ethereum、BSC、Polygon）
- **Moralis**: 多链钱包余额（Solana、备用）

## 先决条件

在开始开发之前，请确保已安装以下工具：

- **Go 1.21+**（工作区模式支持）
- **Node.js 20+** 和 npm
- **Docker & Docker Compose**（用于依赖项和集成测试）
- **Protocol Buffers 编译器** (`protoc`) 及 Go 插件：
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **golang-migrate** CLI：
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint**（用于代码检查）：
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **gofumpt**（用于格式化）：
  ```bash
  go install mvdan.cc/gofumpt@latest
  ```
- **k6**（可选，用于负载测试）：
  ```bash
  # macOS: brew install k6
  # Linux: 参见 https://k6.io/docs/get-started/installation/
  ```

## 项目结构

```
million-dollar-hunter/
├── api-gateway/                # 面向公众的 API 网关 (Go, Fiber)
│   ├── cmd/api-gateway/        # 主入口点
│   ├── internal/               # 私有应用程序代码
│   ├── tests/k6/               # 速率限制的负载测试
│   └── Makefile
├── services/
│   ├── auth-service/           # JWT 认证 (Go, gRPC)
│   │   ├── cmd/auth-service/
│   │   ├── internal/           # 服务逻辑、存储、处理器
│   │   ├── tests/              # 使用 Testcontainers 的集成测试
│   │   ├── db/migrations/      # PostgreSQL 迁移
│   │   └── .golangci.yml       # 代码检查器配置
│   ├── portfolio-service/      # CQRS 读模型 (Go, gRPC)
│   │   ├── cmd/server/
│   │   ├── internal/           # 服务、存储库、Kafka 消费者
│   │   └── migrations/         # 数据库迁移
│   ├── market-data-service/    # CoinGecko 集成 (Go, gRPC)
│   │   ├── cmd/market-data-service/
│   │   ├── internal/           # 缓存、客户端、处理器、存储库、工作器
│   │   └── docker-compose.yml  # 本地开发堆栈
│   └── ingestion-service/      # CQRS 写模型 (Go, HTTP)
│       ├── cmd/ingestion-service/
│       ├── internal/           # Alchemy/Moralis 客户端、速率限制
│       └── docker-compose.yml  # WireMock、Postgres、Redis、Kafka
├── frontend/                   # Next.js 15 App Router
│   ├── src/
│   │   ├── app/                # App Router 页面
│   │   ├── components/         # React 组件
│   │   ├── lib/                # API 客户端、工具
│   │   └── context/            # React 上下文（认证等）
│   ├── tsconfig.json           # TypeScript 配置（严格模式）
│   └── eslint.config.mjs       # ESLint 扁平配置
├── proto/                      # 共享 protobuf 定义
├── docs/                       # 架构、ADR、部署指南
├── ops/                        # Docker compose 编排文件
└── go.work                     # Go 工作区定义
```

## 环境配置

### API 网关

```bash
PORT=8080
JWT_SECRET=devsecret                    # 生产环境至少 32 字节
REDIS_URL=localhost:6379
AUTH_SERVICE_URL=http://localhost:9000
PORTFOLIO_SERVICE_URL=localhost:8081    # gRPC
MARKET_DATA_SERVICE_URL=localhost:50051 # gRPC
AUTH_VALIDATE_MODE=grpc                 # 或 "local"
AUTH_GRPC_ADDR=localhost:50051
FRONTEND_URL=http://localhost:3000      # CORS 配置
RATE_DEFAULT_RPS=10
RATE_DEFAULT_BURST=20
OPENAPI_PATH=../docs/openapi.yaml
```

**CORS 安全说明**: 使用凭证请求时，永远不要使用 `FRONTEND_URL=*`。始终指定确切的来源（多个来源用逗号分隔）。

### 认证服务

```bash
PORT=8081
GRPC_PORT=50051
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
JWT_ISSUER=million-hunter
JWT_AUDIENCE=million-hunter-client
JWT_SIGNING_KEY=devsecret               # 与 JWT_SECRET 相同
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168               # 7 天
ENABLE_MULTI_USER=false                 # MVP 模式
PASSWORD_MIN_LENGTH=12
LOCKOUT_AFTER_FAILS=3
LOCKOUT_WINDOW_MIN=15
```

### 投资组合服务

```bash
GRPC_ADDR=:50052
DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=portfolio-service
TOPIC_TRANSACTION_INGESTED=TransactionDataIngested
EXPORT_DIR=/tmp/exports
EXPORT_CLEANUP_TTL=1h
```

### 市场数据服务

```bash
GRPC_PORT=50051
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_TTL=60s
DATABASE_URL=postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable
COINGECKO_API_KEY=                      # 可选，提高速率限制
COINGECKO_BASE_URL=https://api.coingecko.com/api/v3
COINGECKO_RATE_LIMIT=50                 # 每分钟请求数
WORKER_ENABLED=true
WORKER_REFRESH_INTERVAL=30s
```

### 摄取服务

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

### 前端

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080  # 指向 API 网关
```

有关完整文档，请参阅每个服务目录中的 `.env.example` 文件。

## 构建和测试命令

### 仓库根目录（Go 工作区）

项目使用 Go 工作区（`go.work`），包括所有 Go 服务：

```bash
# 同步所有模块依赖
go work sync

# 在工作区中运行所有测试
go test ./...
```

### API 网关

```bash
cd api-gateway
make build              # 构建二进制文件
make test               # 运行测试
make validate-openapi   # 验证 OpenAPI 规范合规性
make k6                 # 运行负载测试
```

### 认证服务

```bash
cd services/auth-service
make check              # 格式化、代码检查和运行简短测试（推荐预提交）
make fmt                # 使用 gofumpt 格式化代码
make lint               # 运行 golangci-lint
make test               # 运行所有测试（集成测试需要 Docker）
make test-short         # 仅运行单元测试（跳过集成）
make build              # 构建二进制文件到 bin/auth-service
```

**注意**: 集成测试需要运行 Docker 守护进程（使用 Testcontainers 进行 PostgreSQL）。

### 投资组合服务

```bash
cd services/portfolio-service
make build              # 构建二进制文件到 bin/portfolio-service
make proto              # 从 proto 文件重新生成 gRPC 代码
make migrate-up         # 运行数据库迁移（需要 DATABASE_URL）
make migrate-down       # 回滚迁移
make run                # 本地运行服务
```

### 市场数据服务

```bash
cd services/market-data-service
make build              # 构建二进制文件
go test ./internal/... -v                      # 单元测试
docker-compose up -d postgres redis            # 启动依赖项
go test ./tests -v -run Integration            # 集成测试
go test ./tests -v -run Load                   # 负载测试
go test ./tests -bench=. -benchmem             # 基准测试
```

### 摄取服务

```bash
cd ingestion-service
make build              # 构建二进制文件
make test               # 运行测试（需要 docker-compose up）
make lint               # 运行 golangci-lint
make bench              # 运行性能基准测试（目标 ≥100 tx/s）
make up                 # 启动 docker-compose 依赖项
make down               # 停止并删除容器
```

### 前端 (Next.js)

```bash
cd frontend
npm run dev             # 启动开发服务器（端口 3000）
npm run build           # 生产构建
npm start               # 启动生产服务器
npm run lint            # 运行 ESLint
npm test                # 运行 Jest 测试
npm run test:watch      # 在监视模式下运行测试
npm run test:coverage   # 生成覆盖率报告
```

**注意**: 已启用 TypeScript 严格模式。路径别名 `@/*` 映射到 `./src/*`。

## 外部 API 集成

### CoinGecko（市场数据服务）

- **API**: `https://api.coingecko.com/api/v3/`
- **速率限制**: 50 请求/分钟（免费层），使用 API 密钥可更高
- **支持的链**: BSC、Solana、Ethereum、Polygon
- **缓存**: Redis，TTL 60 秒，目标缓存命中率 ≥80%

### Alchemy（摄取服务）

- **API**: 带分页的资产转移端点
- **用途**: 获取区块链交易（Ethereum、BSC、Polygon、Arbitrum、Optimism）
- **速率限制**: 使用 Redis 的令牌桶算法
- **功能**: ERC20、ERC721、ERC1155 转移的单一端点

### Moralis（摄取服务）

- **API**: 多链钱包余额
- **支持的链**: BSC、Solana、Ethereum
- **用途**: 备用提供商和 Solana 支持
- **速率限制**: 用于弹性的断路器模式

## Docker Compose 开发

对于包含所有依赖项和服务的集成本地开发：

```bash
# 启动完整堆栈（从仓库根目录）
docker-compose -f ops/docker-compose.yml up -d --build

# 查看日志
docker-compose -f ops/docker-compose.yml logs -f

# 停止服务
docker-compose -f ops/docker-compose.yml down

# 停止并删除卷
docker-compose -f ops/docker-compose.yml down -v
```

Docker Compose 设置包括：
- 所有后端服务（auth、portfolio、market-data、ingestion、api-gateway）
- 前端应用程序
- 基础设施（PostgreSQL、Redis、Kafka、Zookeeper）
- 健康检查和自动重启

配置通过 `ops/docker-compose.yml` 中的环境变量进行管理。

## 重要说明

### Go 工作区模式

此项目使用 Go 工作区模式（`go.work`）在 monorepo 中管理多个 Go 模块：

- 拉取更改后，始终从仓库根目录运行 `go work sync`
- 每个服务都有自己的 `go.mod` 文件
- 工作区模式在开发期间启用跨服务导入
- CI/CD 使用 `GOWORK=off` 独立构建每个服务

### 集成测试

集成测试需要 Docker 用于 Testcontainers：

- 测试自动启动 PostgreSQL 容器
- 使用 `make test-short` 或 `go test -short` 跳过集成测试
- 在运行完整测试套件之前确保 Docker 守护进程正在运行
- 如果服务已在运行，可能会发生端口冲突

### CORS 安全

API 网关设置 `Access-Control-Allow-Credentials: true`：

- **生产环境**: 使用特定来源（例如，`https://app.million-hunter.com`）
- **多个来源**: 逗号分隔的列表
- **开发环境**: `http://localhost:3000`
- **永远不要**: 使用 `FRONTEND_URL=*`（违反带凭证的 CORS 规范）

### CQRS 模式

系统实现 CQRS 以获得最佳性能：

- **写模型**（摄取服务）：针对高吞吐量数据摄取进行优化
- **读模型**（投资组合服务）：针对快速查询和聚合进行优化
- **事件流**: Kafka 连接写模型和读模型
- **原始数据存储**: JSONB 存储支持无需重新获取的重新处理

### 单用户 MVP 设计

当前实现专为单用户操作而设计：

- 硬编码的认证凭据（用户名：`aezi`，密码：`Aa@123456789`）
- JWT 架构已搭建但为 MVP 简化
- 可以通过设置 `ENABLE_MULTI_USER=true` 启用完整的多用户支持
- MVP 无 GDPR 合规负担

## 其他文档

有关更详细的信息，请参阅以下文档文件：

- **[AGENTS.md](AGENTS.md)** - 全面的开发者参考和 AI 代理简报
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - 部署说明和生产设置
- **[docs/TECHNICAL-DECISIONS.md](docs/TECHNICAL-DECISIONS.md)** - 架构决策和理由
- **[docs/testing-strategy.md](docs/testing-strategy.md)** - 测试方法和覆盖率目标
- **[TESTING-PHASE1-SUMMARY.md](TESTING-PHASE1-SUMMARY.md)** - 测试覆盖率摘要
- **[docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md](docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md)** - 产品需求文档
- **[docs/external-api-integrations.md](docs/external-api-integrations.md)** - 外部 API 集成详细信息

## 性能目标

### API 网关
- 吞吐量：≥100 请求/秒（单实例）
- p95 延迟：≤300ms（缓存请求）
- 速率限制准确性：<1% 错误率

### 市场数据服务
- 缓存命中率：≥80%（5 分钟窗口）
- p95 延迟：≤300ms（缓存），≤2s（缓存未命中 + CoinGecko）
- 吞吐量：≥100 请求/秒

### 摄取服务
- 交易吞吐量：≥100 tx/s（写模型）
- WireMock 延迟：<100ms（本地模拟）
- 外部 API 断路器：50% 失败阈值

### 投资组合服务
- 读模型延迟：<200ms p95
- Kafka 消费者延迟：正常负载下 <10s

### 前端
- 交互时间（TTI）：<3s（缓存）
- 首次内容绘制（FCP）：<1.5s
- 测试覆盖率：≥70% 语句/分支

## 贡献

### 预提交检查

在提交更改之前：

1. 运行 `make check`（auth-service）或 `make lint` + `make test`（其他服务）
2. 确保测试通过：`make test` 或 `go test ./...`
3. 前端：`npm run lint` 和 `npm test`
4. 验证没有未提交的密钥或敏感数据

### 提交约定

使用约定式提交以提高清晰度：

- `feat:` 新功能
- `fix:` 错误修复
- `test:` 添加或更新测试
- `refactor:` 代码重构，无功能更改
- `docs:` 文档更改
- `chore:` 维护任务（依赖项、CI 配置）

示例：`feat(auth): add refresh token rotation`

### 拉取请求要求

每个 PR 必须包括：

1. ✅ 所有测试通过（`make test` 或 `npm test`）
2. ✅ 代码检查通过（`make lint` 或 `npm run lint`）
3. ✅ 类型检查通过（TypeScript：`npm run build`）
4. ✅ 差异仅限于相关文件（避免无关更改）
5. ✅ 证明工件（测试或手动测试证据）
6. ✅ 一段描述：更改了什么、为什么以及任何注意事项
7. ✅ 覆盖率不下降（检查 CI 报告）

## 许可证

双重许可协议（个人使用 / 商业使用）

有关完整条款，请参阅 [LICENSE](LICENSE)。

## 项目状态

**状态**: 积极开发中  
**主要语言**: Go（后端）、TypeScript（前端）  
**部署**: Docker Compose（本地/暂存）、Kubernetes（生产 - 计划中）

---

*该平台代表了链上分析的综合方法，其中每个组件都经过精心架构，以平衡性能、可维护性和可扩展性。设计理念强调关注点分离、强大的错误处理和可观测性优先的开发实践。*
