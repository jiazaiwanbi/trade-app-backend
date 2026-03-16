# trade-app-backend（课设精简版）

`trade-app-backend` 是一个面向二手交易场景的 Go 后端服务，当前采用模块化单体架构，覆盖从用户注册登录、商品发布，到最小订单闭环的核心能力。

## 当前能力

- 用户注册、登录、JWT Access Token 鉴权
- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`
- 分类列表
- 商品创建、列表、详情、编辑、我的商品
- 订单创建、支付模拟、发货模拟、收货确认、取消、我的订单
- 健康检查和数据库就绪检查

## 快速开始

### 1. 准备环境变量

复制环境变量模板：

```bash
cp .env.example .env
```

应用启动时会自动尝试加载项目根目录下的 `.env` 文件。

至少确认以下变量存在：

- `JWT_SECRET`
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`

### 2. 启动 PostgreSQL

```bash
docker compose up -d
```

### 3. 执行数据库迁移

```bash
go run ./cmd/migrate up
go run ./cmd/migrate version
```

预期输出：

```text
version=5 dirty=false
```

### 4. 启动 API

```bash
go run ./cmd/api
```

### 5. 验证健康检查

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/readyz
```

## Docker 方式运行

构建镜像：

```bash
docker build -t trade-app-backend .
```

运行容器时显式注入环境变量：

```bash
docker run --rm -p 8080:8080 --env-file .env trade-app-backend
```

## 常用命令

```bash
make fmt
make test
make run
make migrate-up
make migrate-down
make migrate-version
```

## 核心接口

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`
- `GET /api/v1/categories`
- `GET /api/v1/listings`
- `GET /api/v1/listings/{id}`
- `POST /api/v1/listings`
- `PATCH /api/v1/listings/{id}`
- `GET /api/v1/users/me/listings`
- `POST /api/v1/orders`
- `POST /api/v1/orders/{id}/cancel`
- `POST /api/v1/orders/{id}/pay`
- `POST /api/v1/orders/{id}/ship`
- `POST /api/v1/orders/{id}/receive`
- `GET /api/v1/users/me/orders`

OpenAPI 文档见：`api/openapi/openapi.yaml`

## 最小联调示例

### 1. 注册卖家和买家

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "seller@example.com",
  "password": "secret123",
  "nickname": "seller"
}
```

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "buyer@example.com",
  "password": "secret123",
  "nickname": "buyer"
}
```

### 2. 卖家创建商品

请求头：

```text
Authorization: Bearer <seller_token>
```

```http
POST /api/v1/listings
Content-Type: application/json

{
  "category_id": 1,
  "title": "ThinkPad X1 Carbon",
  "description": "Lightly used business laptop",
  "price_cents": 450000,
  "image_urls": [
    "https://example.com/image-1.jpg",
    "https://example.com/image-2.jpg"
  ],
  "publish": true
}
```

### 3. 买家下单

请求头：

```text
Authorization: Bearer <buyer_token>
```

```http
POST /api/v1/orders
Content-Type: application/json

{
  "listing_id": 1
}
```

### 4. 推进订单状态

- 买家或卖家可取消：`POST /api/v1/orders/{id}/cancel`
- 买家支付模拟：`POST /api/v1/orders/{id}/pay`
- 卖家发货模拟：`POST /api/v1/orders/{id}/ship`
- 买家确认收货：`POST /api/v1/orders/{id}/receive`

## 项目结构

```text
cmd/
  api/                # API 启动入口
  migrate/            # 数据库迁移入口
api/openapi/          # OpenAPI 契约
db/migrations/        # SQL 迁移文件
internal/
  application/        # 用例编排
  domain/             # 领域模型和规则
  platform/           # 配置、日志、数据库、HTTP 基础设施
  repository/         # PostgreSQL 仓储实现
  shared/             # 通用响应和辅助组件
  transport/          # HTTP handler、router、DTO
```

## 注意事项

- `go run ./cmd/api` 会校验 `JWT_SECRET`，未配置时会启动失败。
- `go run ./cmd/migrate up` 需要先确保 PostgreSQL 已启动。
- 当前仓库里部分文件曾出现过 BOM 或错误转码问题；如果再次出现乱码，优先检查编辑器保存编码是否为 UTF-8。
