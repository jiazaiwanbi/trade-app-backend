# trade-app-backend（课设精简版）

本仓库实现一个二手交易平台后端的课设精简版 MVP，当前已经具备从注册登录到商品发布、下单、完成订单的最小交易闭环。

## 当前交付范围

已完成阶段：
- 阶段 1：脚手架与本地可运行
- 阶段 2：配置、日志、错误和中间件
- 阶段 3：PostgreSQL 与迁移体系
- 阶段 4：认证与用户资料
- 阶段 5：商品与分类
- 阶段 6：最小订单闭环
- 阶段 7：OpenAPI、README、Docker 收尾

当前实现的业务能力：
- 用户注册、登录、JWT Access Token
- `GET/PATCH /api/v1/users/me`
- 分类列表
- 商品创建、列表、详情、编辑、我的商品
- 订单创建、取消、完成、我的订单

## 从零启动

### 1. 准备环境变量

复制样例文件：

```powershell
Copy-Item .env.example .env
```

至少确认这些变量可用：
- `JWT_SECRET`
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`

### 2. 启动 PostgreSQL

```powershell
docker compose up -d
```

### 3. 执行数据库迁移

```powershell
go run ./cmd/migrate up
go run ./cmd/migrate version
```

预期版本：`version=4 dirty=false`

### 4. 启动 API

```powershell
go run ./cmd/api
```

### 5. 验证健康检查

```powershell
Invoke-WebRequest http://127.0.0.1:8080/healthz
Invoke-WebRequest http://127.0.0.1:8080/readyz
```

## Docker 启动方式

构建镜像：

```powershell
docker build -t trade-app-backend .
```

运行容器时请自行注入环境变量，例如：

```powershell
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
- `POST /api/v1/listings`
- `GET /api/v1/listings`
- `GET /api/v1/listings/{id}`
- `PATCH /api/v1/listings/{id}`
- `GET /api/v1/users/me/listings`
- `POST /api/v1/orders`
- `POST /api/v1/orders/{id}/cancel`
- `POST /api/v1/orders/{id}/complete`
- `GET /api/v1/users/me/orders`

完整契约见：`api/openapi/openapi.yaml`

## 最小联调示例

### 1. 注册卖家与买家

```json
POST /api/v1/auth/register
{
  "email": "seller@example.com",
  "password": "secret123",
  "nickname": "seller"
}
```

```json
POST /api/v1/auth/register
{
  "email": "buyer@example.com",
  "password": "secret123",
  "nickname": "buyer"
}
```

### 2. 卖家创建商品

请求头：`Authorization: Bearer <seller_token>`

```json
POST /api/v1/listings
{
  "category_id": 1,
  "title": "ThinkPad X1 Carbon",
  "description": "Lightly used business laptop",
  "price_cents": 450000,
  "publish": true
}
```

### 3. 买家下单

请求头：`Authorization: Bearer <buyer_token>`

```json
POST /api/v1/orders
{
  "listing_id": 1
}
```

### 4. 买家完成订单

请求头：`Authorization: Bearer <buyer_token>`

```text
POST /api/v1/orders/1/complete
```

### 5. 查询我的订单

请求头：`Authorization: Bearer <buyer_token>`

```text
GET /api/v1/users/me/orders?page=1&page_size=20
```

## 文档与验证说明

- OpenAPI 已建立：`api/openapi/openapi.yaml`
- `openapi lint` / `openapi validate`：当前仓库未引入对应工具，阶段 7 按要求明确跳过
- 所有接口文档仅覆盖当前已经实现的能力，没有额外声明未来阶段接口
