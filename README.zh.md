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

## 项目状态

**状态**: 积极开发中  
**主要语言**: Go（后端）、TypeScript（前端）  
**部署**: Docker Compose（本地/暂存）、Kubernetes（生产 - 计划中）

---

*该平台代表了链上分析的综合方法，其中每个组件都经过精心架构，以平衡性能、可维护性和可扩展性。设计理念强调关注点分离、强大的错误处理和可观测性优先的开发实践。*

