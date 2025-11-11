# Million Dollar Hunter

منصة تتبع محفظة عملات رقمية شخصية مصممة لمراقبة رموز البلوك تشين وأنشطة المحافظ عبر سلاسل متعددة. مبنية على بنية معمارية للخدمات المصغرة باستخدام نمط CQRS لاستيعاب البيانات عالي الأداء والاستعلام.

> *"كل صياد عظيم يعلم أن الصبر والدقة يؤديان إلى الاكتشافات الأكثر قيمة."* — تم إنشاؤه بعناية فائقة لأنماط البنية المعمارية وتصميم الأنظمة.

## نظرة عامة

Million Dollar Hunter هو لوحة تحكم عملات رقمية على السلسلة لمستخدم واحد تمكن الأفراد من مراقبة واستعلام وتحليل رموز البلوك تشين وأنشطة المحافظ في الوقت الفعلي. توفر المنصة تحليلات عميقة للمحافظ والرموز عبر سلاسل BSC و Solana و Ethereum و Polygon، مع لوحات تحكم قابلة للتخصيص وقدرات تصدير بيانات شاملة.

يستخدم النظام بنية معمارية قائمة على الخدمات المصغرة التي تفصل عمليات الكتابة (استيعاب البيانات) عن عمليات القراءة (استعلامات المحفظة) باستخدام تدفق أحداث Kafka، مما يتيح التوسع والتحسين المستقل لكل مكون.

## البدء السريع

احصل على التطبيق يعمل في 5 خطوات:

### 1. بدء خدمات البنية التحتية

```bash
cd ops
docker-compose up -d
```

يبدأ هذا PostgreSQL و Redis و Kafka باستخدام Docker Compose.

### 2. تشغيل ترحيلات قاعدة البيانات

```bash
# خدمة المصادقة
cd services/auth-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
migrate -path db/migrations -database "$DATABASE_URL" up

# خدمة المحفظة
cd services/portfolio-service
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
make migrate-up

# كرر لـ market-data-service و ingestion-service حسب الحاجة
```

### 3. بدء خدمات الخلفية

افتح محطات منفصلة لكل خدمة:

```bash
# المحطة 1: خدمة المصادقة
cd services/auth-service && make run

# المحطة 2: خدمة المحفظة
cd services/portfolio-service && make run

# المحطة 3: خدمة بيانات السوق
cd services/market-data-service && go run cmd/market-data-service/main.go

# المحطة 4: خدمة الاستيعاب
cd ingestion-service && go run cmd/ingestion-service/main.go

# المحطة 5: بوابة API
cd api-gateway && make run
```

### 4. بدء الواجهة الأمامية

```bash
cd frontend
npm install
npm run dev
```

### 5. الوصول إلى التطبيق

- **الواجهة الأمامية**: http://localhost:3000
- **بوابة API**: http://localhost:8080
- **فحص الصحة**: `curl http://localhost:8080/healthz`

**أوراق اعتماد المصادقة MVP**:
- اسم المستخدم: `aezi`
- كلمة المرور: `Aa@123456789`

## نظرة عامة على البنية المعمارية

### الخدمات المصغرة مع نمط CQRS

تنفذ المنصة فصل مسؤولية الأمر والاستعلام (CQRS)، وفصل عمليات الكتابة (ingestion-service) عن عمليات القراءة (portfolio-service) باستخدام تدفق أحداث Kafka.

### المكونات الأساسية

1. **الواجهة الأمامية (Next.js 15)**
   - تطبيق React 19 حديث مع App Router
   - مكتبة مكونات Material-UI
   - TanStack Query لإدارة حالة الخادم
   - TypeScript مع وضع صارم مفعل

2. **بوابة API (Go/Fiber)**
   - نقطة دخول عامة واحدة لجميع طلبات العميل
   - برمجية وسيطة للمصادقة JWT
   - تحديد معدل مع Redis token bucket
   - توجيه الطلبات إلى خدمات gRPC الخلفية

3. **خدمة المصادقة (JWT/gRPC)**
   - واجهة مزدوجة HTTP/gRPC
   - توليد وتحقق من الرموز المميزة JWT
   - تدوير رمز التحديث
   - حماية قفل تسجيل الدخول (3 فشل في 15 دقيقة)

4. **خدمة المحفظة (نموذج قراءة CQRS)**
   - خادم gRPC لاستعلامات المحفظة
   - مستهلك Kafka لأحداث المعاملات
   - يجمع أرصدة المحافظ من سجل المعاملات
   - يوفر ملخصات المحفظة ووظيفة التصدير

5. **خدمة بيانات السوق (تكامل CoinGecko)**
   - بيانات أسعار الرموز في الوقت الفعلي من CoinGecko API
   - تخزين مؤقت Redis مع TTL 60 ثانية
   - عامل خلفي لتحديث الأسعار
   - واجهة gRPC لاستعلامات الأسعار

6. **خدمة الاستيعاب (نموذج كتابة CQRS)**
   - يجلب بيانات البلوك تشين من واجهات برمجة تطبيقات Alchemy/Moralis
   - ينشر أحداث المعاملات إلى Kafka
   - يتعامل مع تتبع المحافظ متعددة السلاسل
   - نمط قاطع الدائرة لمرونة API

### تدفق البيانات

1. **خدمة الاستيعاب** تجلب معاملات البلوك تشين من واجهات برمجة تطبيقات Alchemy/Moralis
2. تنشر أحداث `TransactionDataIngested` إلى Kafka
3. **خدمة المحفظة** تستهلك الأحداث وتحدّث نماذج القراءة
4. **خدمة بيانات السوق** تثري المحافظ بأسعار الرموز في الوقت الفعلي من CoinGecko
5. **بوابة API** توجه الطلبات المصادق عليها وتفرض حدود المعدل
6. **خدمة المصادقة** تصدر وتحقق من الرموز المميزة JWT عبر gRPC

## المكدس التقني

### الخلفية
- **اللغة**: Go 1.21+ (وضع workspace)
- **قواعد البيانات**: PostgreSQL 15 (الرئيسية)، Redis (التخزين المؤقت، تحديد المعدل)
- **الرسائل**: Apache Kafka (تدفق الأحداث)
- **واجهات برمجة التطبيقات**: gRPC (بين الخدمات)، REST (البوابة العامة)
- **القابلية للمراقبة**: OpenTelemetry، Prometheus، تسجيل JSON منظم (zerolog)

### الواجهة الأمامية
- **الإطار**: Next.js 15 مع App Router
- **مكتبة UI**: React 19، TypeScript، Material-UI
- **إدارة الحالة**: TanStack Query (حالة الخادم)، React Context (حالة UI)
- **الرسوم البيانية**: TradingView Lightweight Charts

### واجهات برمجة التطبيقات الخارجية
- **CoinGecko**: بيانات السوق وأسعار الرموز
- **Alchemy**: بيانات معاملات البلوك تشين (Ethereum، BSC، Polygon)
- **Moralis**: أرصدة المحافظ متعددة السلاسل (Solana، احتياطي)

## المتطلبات الأساسية

قبل بدء التطوير، تأكد من تثبيت الأدوات التالية:

- **Go 1.21+** (دعم وضع workspace)
- **Node.js 20+** و npm
- **Docker & Docker Compose** (للتبعيات واختبارات التكامل)
- **مترجم Protocol Buffers** (`protoc`) مع إضافات Go:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- **golang-migrate** CLI:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint** (للتحقق من الكود):
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **gofumpt** (للتنسيق):
  ```bash
  go install mvdan.cc/gofumpt@latest
  ```
- **k6** (اختياري، لاختبارات التحميل):
  ```bash
  # macOS: brew install k6
  # Linux: راجع https://k6.io/docs/get-started/installation/
  ```

## هيكل المشروع

```
million-dollar-hunter/
├── api-gateway/                # بوابة API العامة (Go, Fiber)
│   ├── cmd/api-gateway/        # نقطة الدخول الرئيسية
│   ├── internal/               # كود التطبيق الخاص
│   ├── tests/k6/               # اختبارات التحميل لتحديد المعدل
│   └── Makefile
├── services/
│   ├── auth-service/           # المصادقة JWT (Go, gRPC)
│   │   ├── cmd/auth-service/
│   │   ├── internal/           # منطق الخدمة، التخزين، المعالجات
│   │   ├── tests/              # اختبارات التكامل مع Testcontainers
│   │   ├── db/migrations/      # ترحيلات PostgreSQL
│   │   └── .golangci.yml       # إعدادات المحلل
│   ├── portfolio-service/      # نموذج قراءة CQRS (Go, gRPC)
│   │   ├── cmd/server/
│   │   ├── internal/           # الخدمة، المستودع، مستهلك Kafka
│   │   └── migrations/         # ترحيلات قاعدة البيانات
│   ├── market-data-service/    # تكامل CoinGecko (Go, gRPC)
│   │   ├── cmd/market-data-service/
│   │   ├── internal/           # التخزين المؤقت، العميل، المعالج، المستودع، العامل
│   │   └── docker-compose.yml  # مكدس التطوير المحلي
│   └── ingestion-service/      # نموذج كتابة CQRS (Go, HTTP)
│       ├── cmd/ingestion-service/
│       ├── internal/           # عملاء Alchemy/Moralis، تحديد المعدل
│       └── docker-compose.yml  # WireMock، Postgres، Redis، Kafka
├── frontend/                   # Next.js 15 App Router
│   ├── src/
│   │   ├── app/                # صفحات App Router
│   │   ├── components/         # مكونات React
│   │   ├── lib/                # عميل API، أدوات
│   │   └── context/            # سياق React (المصادقة، إلخ)
│   ├── tsconfig.json           # إعدادات TypeScript (وضع صارم)
│   └── eslint.config.mjs       # إعدادات ESLint المسطحة
├── proto/                      # تعريفات protobuf المشتركة
├── docs/                       # البنية المعمارية، ADRs، أدلة النشر
├── ops/                        # ملفات تنسيق Docker compose
└── go.work                     # تعريف workspace Go
```

## تكوين البيئة

### بوابة API

```bash
PORT=8080
JWT_SECRET=devsecret                    # الحد الأدنى 32 بايت في الإنتاج
REDIS_URL=localhost:6379
AUTH_SERVICE_URL=http://localhost:9000
PORTFOLIO_SERVICE_URL=localhost:8081    # gRPC
MARKET_DATA_SERVICE_URL=localhost:50051 # gRPC
AUTH_VALIDATE_MODE=grpc                 # أو "local"
AUTH_GRPC_ADDR=localhost:50051
FRONTEND_URL=http://localhost:3000      # تكوين CORS
RATE_DEFAULT_RPS=10
RATE_DEFAULT_BURST=20
OPENAPI_PATH=../docs/openapi.yaml
```

**ملاحظة أمان CORS**: لا تستخدم أبدًا `FRONTEND_URL=*` مع الطلبات ذات الأوراق الاعتماد. حدد دائمًا المصادر الدقيقة (مفصولة بفواصل لمصادر متعددة).

### خدمة المصادقة

```bash
PORT=8081
GRPC_PORT=50051
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth?sslmode=disable
JWT_ISSUER=million-hunter
JWT_AUDIENCE=million-hunter-client
JWT_SIGNING_KEY=devsecret               # نفس JWT_SECRET
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168               # 7 أيام
ENABLE_MULTI_USER=false                 # وضع MVP
PASSWORD_MIN_LENGTH=12
LOCKOUT_AFTER_FAILS=3
LOCKOUT_WINDOW_MIN=15
```

### خدمة المحفظة

```bash
GRPC_ADDR=:50052
DATABASE_URL=postgres://postgres:postgres@localhost:5432/portfolio?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=portfolio-service
TOPIC_TRANSACTION_INGESTED=TransactionDataIngested
EXPORT_DIR=/tmp/exports
EXPORT_CLEANUP_TTL=1h
```

### خدمة بيانات السوق

```bash
GRPC_PORT=50051
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_TTL=60s
DATABASE_URL=postgres://postgres:postgres@localhost:5432/market_data?sslmode=disable
COINGECKO_API_KEY=                      # اختياري، يزيد حدود المعدل
COINGECKO_BASE_URL=https://api.coingecko.com/api/v3
COINGECKO_RATE_LIMIT=50                 # طلبات في الدقيقة
WORKER_ENABLED=true
WORKER_REFRESH_INTERVAL=30s
```

### خدمة الاستيعاب

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

### الواجهة الأمامية

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080  # يشير إلى بوابة API
```

راجع ملفات `.env.example` في كل دليل خدمة للحصول على وثائق كاملة.

## أوامر البناء والاختبار

### جذر المستودع (Go Workspace)

يستخدم المشروع workspace Go (`go.work`) يتضمن جميع الخدمات Go:

```bash
# مزامنة جميع تبعيات الوحدات
go work sync

# تشغيل جميع الاختبارات عبر workspace
go test ./...
```

### بوابة API

```bash
cd api-gateway
make build              # بناء ثنائي
make test               # تشغيل الاختبارات
make validate-openapi   # التحقق من امتثال مواصفات OpenAPI
make k6                 # تشغيل اختبارات التحميل
```

### خدمة المصادقة

```bash
cd services/auth-service
make check              # التنسيق والتحقق وتشغيل اختبارات قصيرة (موصى به قبل الالتزام)
make fmt                # تنسيق الكود باستخدام gofumpt
make lint               # تشغيل golangci-lint
make test               # تشغيل جميع الاختبارات (يتطلب Docker لاختبارات التكامل)
make test-short         # تشغيل اختبارات الوحدة فقط (تخطي التكامل)
make build              # بناء ثنائي إلى bin/auth-service
```

**ملاحظة**: تتطلب اختبارات التكامل تشغيل daemon Docker (يستخدم Testcontainers لـ PostgreSQL).

### خدمة المحفظة

```bash
cd services/portfolio-service
make build              # بناء ثنائي إلى bin/portfolio-service
make proto              # إعادة إنشاء كود gRPC من ملفات proto
make migrate-up         # تشغيل ترحيلات قاعدة البيانات (يتطلب DATABASE_URL)
make migrate-down       # التراجع عن الترحيلات
make run                # تشغيل الخدمة محليًا
```

### خدمة بيانات السوق

```bash
cd services/market-data-service
make build              # بناء ثنائي
go test ./internal/... -v                      # اختبارات الوحدة
docker-compose up -d postgres redis            # بدء التبعيات
go test ./tests -v -run Integration            # اختبارات التكامل
go test ./tests -v -run Load                   # اختبارات التحميل
go test ./tests -bench=. -benchmem             # معايير الأداء
```

### خدمة الاستيعاب

```bash
cd ingestion-service
make build              # بناء ثنائي
make test               # تشغيل الاختبارات (يتطلب docker-compose up)
make lint               # تشغيل golangci-lint
make bench              # تشغيل معايير الأداء (هدف ≥100 tx/s)
make up                 # بدء تبعيات docker-compose
make down               # إيقاف وإزالة الحاويات
```

### الواجهة الأمامية (Next.js)

```bash
cd frontend
npm run dev             # بدء خادم التطوير (المنفذ 3000)
npm run build           # بناء الإنتاج
npm start               # بدء خادم الإنتاج
npm run lint            # تشغيل ESLint
npm test                # تشغيل اختبارات Jest
npm run test:watch      # تشغيل الاختبارات في وضع المراقبة
npm run test:coverage   # إنشاء تقرير التغطية
```

**ملاحظة**: وضع TypeScript الصارم مفعل. اسم المسار `@/*` ينتقل إلى `./src/*`.

## تكاملات واجهات برمجة التطبيقات الخارجية

### CoinGecko (خدمة بيانات السوق)

- **API**: `https://api.coingecko.com/api/v3/`
- **حد المعدل**: 50 طلب/دقيقة (المستوى المجاني)، أعلى مع مفتاح API
- **السلاسل المدعومة**: BSC، Solana، Ethereum، Polygon
- **التخزين المؤقت**: Redis مع TTL 60 ثانية، استهداف معدل ضربات التخزين المؤقت ≥80%

### Alchemy (خدمة الاستيعاب)

- **API**: نقطة نهاية تحويلات الأصول مع التصفح
- **الغرض**: جلب معاملات البلوك تشين (Ethereum، BSC، Polygon، Arbitrum، Optimism)
- **تحديد المعدل**: خوارزمية token bucket مع Redis
- **الميزات**: نقطة نهاية واحدة لتحويلات ERC20 و ERC721 و ERC1155

### Moralis (خدمة الاستيعاب)

- **API**: أرصدة المحافظ متعددة السلاسل
- **السلاسل المدعومة**: BSC، Solana، Ethereum
- **الغرض**: مزود احتياطي ودعم Solana
- **تحديد المعدل**: نمط قاطع الدائرة للمرونة

## التطوير باستخدام Docker Compose

للتنمية المحلية المتكاملة مع جميع التبعيات والخدمات:

```bash
# بدء المكدس الكامل (من جذر المستودع)
docker-compose -f ops/docker-compose.yml up -d --build

# عرض السجلات
docker-compose -f ops/docker-compose.yml logs -f

# إيقاف الخدمات
docker-compose -f ops/docker-compose.yml down

# إيقاف وإزالة الأحجام
docker-compose -f ops/docker-compose.yml down -v
```

يتضمن إعداد Docker Compose:
- جميع الخدمات الخلفية (auth، portfolio، market-data، ingestion، api-gateway)
- تطبيق الواجهة الأمامية
- البنية التحتية (PostgreSQL، Redis، Kafka، Zookeeper)
- فحوصات الصحة وإعادة التشغيل التلقائي

يتم إدارة التكوين عبر متغيرات البيئة في `ops/docker-compose.yml`.

## ملاحظات مهمة

### وضع Go Workspace

يستخدم هذا المشروع وضع workspace Go (`go.work`) لإدارة وحدات Go متعددة في monorepo:

- قم دائمًا بتشغيل `go work sync` من جذر المستودع بعد سحب التغييرات
- لكل خدمة ملف `go.mod` خاص بها
- يتيح وضع workspace الاستيراد بين الخدمات أثناء التطوير
- يبني CI/CD كل خدمة بشكل مستقل مع `GOWORK=off`

### اختبارات التكامل

تتطلب اختبارات التكامل Docker لـ Testcontainers:

- تبدأ الاختبارات تلقائيًا حاويات PostgreSQL
- استخدم `make test-short` أو `go test -short` لتخطي اختبارات التكامل
- تأكد من تشغيل daemon Docker قبل تشغيل مجموعة الاختبارات الكاملة
- قد تحدث تعارضات المنافذ إذا كانت الخدمات قيد التشغيل بالفعل

### أمان CORS

تحدد بوابة API `Access-Control-Allow-Credentials: true`:

- **الإنتاج**: استخدم مصدرًا محددًا (مثل `https://app.million-hunter.com`)
- **مصادر متعددة**: قائمة مفصولة بفواصل
- **التطوير**: `http://localhost:3000`
- **أبدًا**: لا تستخدم `FRONTEND_URL=*` (ينتهك مواصفات CORS مع الأوراق الاعتماد)

### نمط CQRS

ينفذ النظام CQRS للحصول على أداء مثالي:

- **نموذج الكتابة** (خدمة الاستيعاب): محسّن لاستيعاب البيانات عالي الإنتاجية
- **نموذج القراءة** (خدمة المحفظة): محسّن للاستعلامات والتجميعات السريعة
- **تدفق الأحداث**: Kafka يربط نماذج الكتابة والقراءة
- **تخزين البيانات الأولية**: يتيح تخزين JSONB إعادة المعالجة دون إعادة الجلب

### تصميم MVP لمستخدم واحد

تم تصميم التنفيذ الحالي لعملية مستخدم واحد:

- أوراق اعتماد المصادقة المبرمجة (اسم المستخدم: `aezi`، كلمة المرور: `Aa@123456789`)
- تم إعداد بنية JWT ولكن تم تبسيطها لـ MVP
- يمكن تمكين دعم متعدد المستخدمين الكامل عن طريق تعيين `ENABLE_MULTI_USER=true`
- لا يوجد عبء امتثال GDPR لـ MVP

## وثائق إضافية

لمزيد من المعلومات التفصيلية، راجع ملفات الوثائق هذه:

- **[AGENTS.md](AGENTS.md)** - مرجع شامل للمطورين وإحاطة وكيل AI
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - تعليمات النشر وإعداد الإنتاج
- **[docs/TECHNICAL-DECISIONS.md](docs/TECHNICAL-DECISIONS.md)** - قرارات البنية المعمارية والمبررات
- **[docs/testing-strategy.md](docs/testing-strategy.md)** - نهج الاختبار وأهداف التغطية
- **[TESTING-PHASE1-SUMMARY.md](TESTING-PHASE1-SUMMARY.md)** - ملخص تغطية الاختبار
- **[docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md](docs/PRD-Million-Dollar-Hunter-Crypto-Dashboard.md)** - وثيقة متطلبات المنتج
- **[docs/external-api-integrations.md](docs/external-api-integrations.md)** - تفاصيل تكامل واجهات برمجة التطبيقات الخارجية

## أهداف الأداء

### بوابة API
- الإنتاجية: ≥100 طلب/ثانية (مثيل واحد)
- زمن الاستجابة p95: ≤300ms (الطلبات المخزنة مؤقتًا)
- دقة حد المعدل: <1% معدل خطأ

### خدمة بيانات السوق
- معدل ضربات التخزين المؤقت: ≥80% (نوافذ 5 دقائق)
- زمن الاستجابة p95: ≤300ms (مخزن مؤقتًا)، ≤2s (فشل التخزين المؤقت + CoinGecko)
- الإنتاجية: ≥100 طلب/ثانية

### خدمة الاستيعاب
- إنتاجية المعاملات: ≥100 tx/s (نموذج الكتابة)
- زمن استجابة WireMock: <100ms (محاكيات محلية)
- قاطع دائرة واجهة برمجة التطبيقات الخارجية: عتبة فشل 50%

### خدمة المحفظة
- زمن استجابة نموذج القراءة: <200ms p95
- تأخر مستهلك Kafka: <10s تحت الحمل العادي

### الواجهة الأمامية
- الوقت حتى التفاعل (TTI): <3s (مخزن مؤقتًا)
- أول رسم للمحتوى (FCP): <1.5s
- تغطية الاختبار: ≥70% عبارات/فروع

## المساهمة

### فحوصات ما قبل الالتزام

قبل الالتزام بالتغييرات:

1. قم بتشغيل `make check` (auth-service) أو `make lint` + `make test` (خدمات أخرى)
2. تأكد من نجاح الاختبارات: `make test` أو `go test ./...`
3. الواجهة الأمامية: `npm run lint` و `npm test`
4. تحقق من عدم وجود أسرار أو بيانات حساسة غير ملتزم بها

### اتفاقيات الالتزام

استخدم الالتزامات التقليدية للوضوح:

- `feat:` ميزة جديدة
- `fix:` إصلاح خطأ
- `test:` إضافة أو تحديث الاختبارات
- `refactor:` إعادة هيكلة الكود دون تغييرات وظيفية
- `docs:` تغييرات الوثائق
- `chore:` مهام الصيانة (التبعيات، تكوين CI)

مثال: `feat(auth): add refresh token rotation`

### متطلبات Pull Request

يجب أن يتضمن كل PR:

1. ✅ نجاح جميع الاختبارات (`make test` أو `npm test`)
2. ✅ نجاح التحقق من الكود (`make lint` أو `npm run lint`)
3. ✅ نجاح فحص الأنواع (TypeScript: `npm run build`)
4. ✅ Diff محصور في الملفات ذات الصلة (تجنب التغييرات غير ذات الصلة)
5. ✅ دليل (اختبارات أو دليل اختبار يدوي)
6. ✅ وصف فقرة واحدة: ما الذي تغير، ولماذا، وأي محاذير
7. ✅ لا انخفاض في التغطية (تحقق من تقارير CI)

## الترخيص

اتفاقية الترخيص المزدوج (الاستخدام الشخصي / الاستخدام التجاري)

راجع [LICENSE](LICENSE) للحصول على الشروط الكاملة.

## حالة المشروع

**الحالة**: التطوير النشط  
**اللغة الأساسية**: Go (الخلفية)، TypeScript (الواجهة الأمامية)  
**النشر**: Docker Compose (محلي/المرحلة)، Kubernetes (الإنتاج - مخطط)

---

*تمثل هذه المنصة نهجًا شاملاً لتحليل السلسلة، حيث تم تصميم كل مكون بعناية لتحقيق التوازن بين الأداء والقابلية للصيانة وقابلية التوسع. تؤكد فلسفة التصميم على فصل الاهتمامات ومعالجة الأخطاء القوية وممارسات التطوير التي تركز على القابلية للمراقبة أولاً.*
