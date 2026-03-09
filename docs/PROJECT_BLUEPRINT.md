# PROJECT_BLUEPRINT.md

## 1. 项目定位

`trade-app-backend` 是一个面向二手交易场景的后端服务；课设版本的核心目标是先支持“用户发布闲置商品、浏览搜索、发起交易并完成成交”的最小闭环。

> 课设说明（1 周交付）：本仓库采用“**精简版 MVP**”路线，优先把最小交易闭环跑通并保证工程可运行、可测试、可部署；举报、聊天、图片/头像上传、通知、后台审核等能力统一放到后续阶段。

### 目标用户角色

- 买家
- 卖家
- （后续）管理员 / 审核员

### 核心业务目标

- 支持用户注册登录与鉴权（课设 MVP 默认 JWT Access Token）
- 支持商品发布、编辑、上下架、状态管理
- 支持基础搜索与分类过滤（收藏/复杂筛选后续再做）
- 支持最小交易闭环（下单 / 取消 / 完成；不接支付、不做物流）

---

## 2. MVP 边界

### MVP 必做

1. 用户认证（注册 / 登录 / 鉴权中间件）
2. 用户资料（仅基础资料：昵称/简介等；不做头像上传）
3. 分类（可选：先做扁平列表即可）
4. 商品发布与浏览（文字信息 + 价格 + 状态）
5. 基础搜索（按关键字/分类过滤；先不做复杂排序与筛选项模型）
6. 最小交易闭环（下单/取消/完成；不接支付、不做发货物流）
7. 基础工程能力（配置、日志、健康检查、迁移、基础测试、Docker 本地运行）

### MVP 非目标

以下内容可预留，但不建议第一阶段实现：

- 举报/审核/后台 admin
- 上传（头像/商品图片）与对象存储（MinIO/S3）
- 收藏、聊天、通知、评价
- 支付接入、物流、纠纷
- Redis（缓存/会话/限流/队列）
- Prometheus/OTel 等完整可观测性体系
- 微服务拆分
- 复杂推荐系统
- 搜索引擎集群（先用 PostgreSQL 搜索能力）
- 完整支付网关深度集成
- 实时 WebSocket 集群扩展优化
- 多租户
- 多语言国际化
- 复杂营销活动系统
- AI 审核 / 智能风控

---

## 3. 架构建议

### 推荐形态：模块化单体

对于一个从 0 到 1 的二手交易平台，推荐采用 **模块化单体（modular monolith）**，理由如下：

- 开发效率更高
- 仓库和部署形态简单
- 更容易统一认证、事务、一致性和日志
- 后期可按业务边界拆分
- 对 code agent 更友好，减少跨服务上下文切换

### 高层架构图（逻辑）

```text
Client
  -> HTTP API
    -> Transport Layer
      -> Application Layer
        -> Domain Layer
          -> Repository Interfaces
            -> PostgreSQL / (可选：Redis / 对象存储 / 第三方 Adapter)
```

### 基础原则

- API 层只负责协议适配
- 应用层负责用例编排和事务边界
- 领域层负责业务规则和状态流转
- 仓储层负责持久化与查询
- 外部系统通过 adapter 接入

---

## 4. 推荐技术栈

这是建议方案，不是绝对强制；如果调整，必须同步更新 `AGENTS.md` 和本文档。

### 4.1 服务端

- Go
- `chi` 作为路由器
- `net/http` 作为底层 HTTP 框架
- `go-playground/validator` 或同类库做输入校验
- OpenAPI 维护 API 契约

### 4.2 数据层

- PostgreSQL：主业务数据
- `pgx`：推荐直接使用 `pgxpool` 访问数据库（课设 MVP 先不强制引入 `sqlc`）
- （可选，后续）Redis：会话/缓存/限流/幂等键/队列基础
- `golang-migrate`：维护数据库迁移

### 4.3 文件与对象存储

- （后续）如需上传头像/商品图片，再引入 S3 兼容对象存储（本地可用 MinIO）

### 4.4 观测与运维

- 结构化日志
- 健康检查 / 就绪检查
- Docker 化部署
- （可选，后续）Prometheus 指标 / OpenTelemetry
- （可选，后续）CI（GitHub Actions）

---

## 5. 模块边界设计

### 5.1 auth

职责：

- 注册
- 登录
- （可选）登出
- 鉴权中间件（保护需要登录的接口）

关键设计点：

- 课设 MVP 优先采用 **JWT Access Token**（无需 Redis/会话表即可跑通）
- 密码必须做安全哈希（bcrypt/argon2 二选一即可）
- （后续）如需要“真正登出/多端会话/刷新 Token”，再引入 Refresh Token + 会话存储

### 5.2 user

职责：

- 我的资料
- 昵称 / 简介（不做头像上传；如需头像仅保留 `avatar_url` 字段占位）
- （可选）用户状态（正常/禁用），用于后续扩展

### 5.3 category

职责：

- 分类列表（课设 MVP 可先做扁平结构；需要树再扩展）

### 5.4 listing

职责：

- 商品创建、编辑、删除、上下架
- 商品状态流转
- 价格、成色、描述（标签/复杂筛选可后续加）
- 归属权校验
- 保留/售出语义

### 5.5 order

职责（课设 MVP）：

- 下单（买家对某个商品发起交易）
- 取消订单（买家/卖家在允许状态下取消）
- 完成订单（双方确认成交后完成）
- 与商品状态联动（下单可将商品置为 reserved，完成后置为 sold）

设计建议：

- 默认一个订单对应一件商品（简化库存语义）
- 不接支付、不做发货/物流、不做纠纷（都属于后续扩展）

### 5.6 后续模块（不在课设 1 周 MVP 范围）

- `media`：头像/商品图片上传与对象存储
- `favorite`：收藏
- `chat`：私信聊天
- `payment`：支付与回调
- `review`：评价
- `report`：举报
- `notification`：站内通知
- `admin`：后台审核/封禁/审计

---

## 6. 目录结构建议

```text
cmd/
  api/                    # 主服务启动入口
  migrate/                # 迁移工具入口

configs/                  # 配置模板、环境配置
api/openapi/              # OpenAPI 契约
db/
  migrations/             # up/down SQL
  query/                  # sqlc 查询文件
  sqlc/                   # sqlc 配置
internal/
  platform/               # config/logger/db/redis/http server/bootstrap
  shared/                 # 通用 error、response、pagination、id、clock 等
  transport/
    http/
      middleware/
      handler/
      dto/
      router/
  application/
    auth/
    user/
    category/
    listing/
    media/
    favorite/
    chat/
    order/
    payment/
    review/
    report/
    notification/
    admin/
  domain/
    auth/
    user/
    category/
    listing/
    media/
    favorite/
    chat/
    order/
    payment/
    review/
    report/
    notification/
    admin/
  repository/
    postgres/
    redis/
    s3/
  jobs/                   # 异步任务消费者/生产者（后续）
scripts/
deployments/
docs/
```

### 目录规则

- `application` 放用例编排，不放协议细节
- `domain` 放核心规则，不放数据库 schema
- `transport/http/dto` 专门承载请求响应对象
- `repository/postgres` 按模块组织 DAO 与查询映射
- `shared` 只放真正跨模块共用且稳定的能力

---

## 7. API 设计建议

### 7.1 URL 风格

统一使用：

```text
/api/v1/...
```

### 7.2 资源设计示例

#### 认证

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- （可选）`POST /api/v1/auth/logout`

#### 用户

- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`

#### 分类

- `GET /api/v1/categories`
- （可选）`GET /api/v1/categories/tree`

#### 商品

- `POST /api/v1/listings`
- `GET /api/v1/listings/{id}`
- `PATCH /api/v1/listings/{id}`
- （可选）`POST /api/v1/listings/{id}/publish`
- （可选）`POST /api/v1/listings/{id}/archive`
- `GET /api/v1/listings`
- `GET /api/v1/users/me/listings`

#### 订单

- `POST /api/v1/orders`
- `GET /api/v1/orders/{id}`
- `GET /api/v1/users/me/orders`
- `POST /api/v1/orders/{id}/cancel`
- `POST /api/v1/orders/{id}/complete`

### 7.3 响应格式建议

成功响应建议统一：

```json
{
  "code": "OK",
  "message": "success",
  "data": {}
}
```

失败响应建议统一：

```json
{
  "code": "VALIDATION_ERROR",
  "message": "invalid request",
  "details": [
    {
      "field": "price",
      "reason": "must be greater than 0"
    }
  ]
}
```

分页响应建议统一带：

- `page`
- `page_size`
- `total`
- `items`

---

## 8. 数据模型建议

以下是建议的 MVP 数据表，不要求一次性全部实现，但应围绕它规划迁移与模块边界。

### 8.1 用户与认证

- `users`
- （可选）`user_profiles`（如果你不想把所有字段塞进 `users`）

### 8.2 分类与商品

- `categories`
- `listings`
- （可选）更复杂的筛选项/标签/图片等都放到后续

### 8.3 订单

- `orders`

---

## 9. 核心状态机建议

### 9.1 商品状态

```text
draft
 -> published
 -> reserved
 -> sold
 -> archived
```

### 9.2 订单状态

```text
created
 -> cancelled
 -> completed
```

状态流转建议在 `domain` 中集中建模，不要散落在 handler。

---

## 10. 关键工程约束

### 10.1 事务边界

以下场景应优先在应用层建立事务边界：

- 创建订单并锁定商品（listing -> reserved）
- 完成订单并更新商品状态（listing -> sold）

### 10.2 幂等

以下动作必须支持幂等或重复请求保护：

- 创建订单
- （可选）取消/完成订单（视客户端重试策略决定）

### 10.3 权限

至少区分：

- 未登录用户
- 已登录普通用户
- 商品所有者
- 买家 / 卖家
- 管理员

---

## 11. 可观测性与运维建议

### 基础能力

- `GET /healthz`
- `GET /readyz`
- 请求日志
- 错误日志
- DB 连接池指标
- HTTP latency / status code 指标
- trace id 透传

### 部署建议

环境分层：

- local
- dev
- staging
- prod

配置必须以环境变量为主，避免多处散落配置源。

---

## 12. 测试策略

### 单元测试

重点覆盖：

- 状态机
- 权限检查
- 参数合法性
- 订单流程
- 商品与订单的边界条件（如重复下单、状态流转非法等）

### 集成测试

重点覆盖：

- repository 查询
- migration 执行
- handler 与 DB 联动
- auth 中间件
- 关键交易流程

### 端到端测试（后续）

- 注册 -> 登录 -> 发布商品 -> 搜索 -> 下单 -> 取消/完成

---

## 13. 发布与演进路线

### 课设（1 周）推荐路线

1. 脚手架：可启动 API + Docker 本地 Postgres + `/healthz`/`/readyz`
2. 平台层：配置/日志/统一错误与响应/基础中间件
3. 数据层：迁移体系 + 连接池 + 最小 repository 样板
4. 业务闭环：auth + user + listing (+ category 可选) + order（最小状态机）
5. 收尾：OpenAPI 草案 + 基础测试 + README 命令说明

> 以上路线与 `docs/CODE_AGENT_TASKS.md` 的精简阶段保持一致。

---

## 14. 给 code agent 的总体执行方式

建议每次只投喂一个阶段任务，格式如下：

```text
请严格阅读 AGENTS.md、docs/PROJECT_BLUEPRINT.md、docs/CODE_AGENT_TASKS.md，
现在只完成第 X 阶段，不要提前实现后续阶段。
要求：
1. 输出执行计划
2. 只修改与当前阶段相关的文件
3. 完成后运行必要校验命令
4. 输出修改文件、运行命令、验收结果、剩余风险
```

---

## 15. 成功标准

当项目进入“可对外联调”的状态时，应至少具备：

- 清晰的目录结构
- 完整的基础设施脚手架
- 稳定的认证与权限控制
- 商品发布与搜索闭环
- 订单主流程闭环
- OpenAPI 文档
- 基础测试（CI 可选，后续补齐）
- Docker 化部署能力
- 必要的日志、健康检查、错误处理规范
