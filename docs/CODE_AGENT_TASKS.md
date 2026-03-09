# CODE_AGENT_TASKS.md

本文档用于把整个项目拆成 **可逐步投喂给 code agent 的阶段任务**。

课设目标（截至 2026-03-09 起算 1 周内）：
- 优先完成“能跑起来 + 最小交易闭环”
- 明确不做：举报、后台审核、收藏、聊天、通知、支付、头像/商品图片上传、Redis、对象存储、复杂可观测性

---

## 总规则

每次你把任务交给 code agent 时，都附上这段固定要求：

```text
请先阅读 AGENTS.md、docs/PROJECT_BLUEPRINT.md、docs/CODE_AGENT_TASKS.md。
现在只完成当前阶段任务，不要提前实现后续阶段。
完成后请输出：
1. 执行计划
2. 修改文件列表
3. 运行命令
4. 验收结果
5. 剩余风险和下一步建议
```

---

# 阶段 1：脚手架与本地可运行

## 目标

让项目可以在本地启动（包含 Docker 依赖），并有最小健康检查。

## 任务

- 初始化目录结构（`cmd/api`、`internal/`、`db/migrations` 等）
- `docker-compose.yml`：仅包含 PostgreSQL（Redis/MinIO 后续再加）
- 最小 HTTP server：`/healthz`、`/readyz`
- `.env.example`、Makefile（至少 `fmt`/`test`/`run`/`migrate-up`）

## 验收标准

- `go run ./cmd/api` 可启动并访问健康检查
- `docker compose up -d` 可拉起 PostgreSQL

## 投喂 prompt

```text
请完成阶段 1：脚手架与本地可运行（课设精简版）。
要求：
1. 只做可启动空壳与本地 Postgres 依赖
2. 提供 /healthz、/readyz
3. 输出启动与验证命令
```

---

# 阶段 2：平台层（配置/日志/错误/中间件）

## 目标

后续业务模块都在统一的响应/错误/中间件约束下实现。

## 任务

- 配置加载（环境变量 + 校验 + 默认值）
- JSON 结构化日志 + request id
- 中间件：recovery、request id、logger、timeout（CORS 可选）
- 统一错误模型与统一 JSON 响应封装

## 验收标准

- `/healthz`、`/readyz` 走统一路由和中间件链
- handler 不直接 `panic`/不直接打印日志

## 投喂 prompt

```text
请完成阶段 2：平台层（配置/日志/错误/中间件）。
要求：
1. 让现有健康检查走统一响应/错误处理
2. 不要开始业务表与业务模块实现
```

---

# 阶段 3：PostgreSQL + 迁移体系

## 目标

能稳定管理数据库 schema，并为业务模块提供可复用的数据访问基础。

## 任务

- `pgxpool` 连接池
- `golang-migrate`：up/down + `cmd/migrate`（或 Makefile 命令）
- readiness 检查能判断 DB 连接是否可用

## 验收标准

- 迁移可执行（up/down 至少各跑一次）
- API 启动时能正确报告 DB readiness

## 投喂 prompt

```text
请完成阶段 3：PostgreSQL + 迁移体系（精简版）。
要求：
1. 只接 PostgreSQL（不接 Redis/sqlc）
2. 提供可运行的迁移命令
```

---

# 阶段 4：认证与用户资料（auth + user）

## 目标

实现最小登录态与“我的资料”接口，作为后续 listing/order 的权限基础。

## 任务

- 表：`users`（可选 `user_profiles`）
- 注册/登录：密码安全哈希（bcrypt/argon2 二选一即可）
- JWT Access Token（鉴权中间件）
- `GET/PATCH /api/v1/users/me`

## 验收标准

- 登录后可访问 `GET /users/me`；未登录返回 401
- 密码不明文入库

## 投喂 prompt

```text
请完成阶段 4：认证与用户资料（JWT Access Token）。
要求：
1. 不做 refresh token/会话表/Redis
2. 输出 OpenAPI 片段或接口说明
3. 补基础测试（至少覆盖注册/登录/鉴权失败）
```

---

# 阶段 5：商品（listing）+ 分类（category 可选）

## 目标

用户可发布商品、浏览列表、查看详情、编辑、下架/售出（状态流转）。

## 任务

- 表：`listings`、（可选）`categories`
- 接口：`POST/GET/PATCH /listings`、`GET /listings/{id}`、`GET /users/me/listings`
- 基础搜索：关键字 + 分类过滤 + 分页（先不做复杂排序/筛选模型）
- 权限：只有卖家可编辑自己的商品

## 验收标准

- 未登录不能发布/编辑；已登录可发布并在列表中查到
- 商品状态流转集中在 domain

## 投喂 prompt

```text
请完成阶段 5：商品发布与浏览（精简版）。
要求：
1. 不做图片/上传/收藏/聊天
2. 补主要路径测试（创建/查询/越权编辑失败）
```

---

# 阶段 6：订单（最小交易闭环）

## 目标

买家可对某商品下单，卖家/买家可取消，双方完成成交。

## 任务

- 表：`orders`
- 状态机：`created -> cancelled/completed`
- 下单时锁定商品（listing -> reserved），完成后置为 sold
- 权限：买家仅能操作自己的订单；卖家仅能操作自己商品相关订单

## 验收标准

- 并发/重复下单不会产生多笔“有效”订单（至少保证 DB 层唯一约束或事务一致性）
- 订单与商品状态联动正确

## 投喂 prompt

```text
请完成阶段 6：订单（最小闭环）。
要求：
1. 不接支付、不做物流
2. 补表驱动测试覆盖状态流转
```

---

# 阶段 7：收尾（OpenAPI/README/Docker）

## 目标

让验收者能按文档把服务跑起来并调用核心接口。

## 任务

- `api/openapi/`：最小 OpenAPI 文档（覆盖已实现接口）
- README：本地启动、环境变量、常用命令
- Dockerfile：能构建并启动 API

## 验收标准

- `docker compose up -d` + `go run ./cmd/api` 或 `docker compose up --build` 能启动服务
- OpenAPI 与实现接口一致（若校验工具未引入，则明确说明“未建立，跳过”）

## 投喂 prompt

```text
请完成阶段 7：收尾（OpenAPI/README/Docker）。
要求：
1. 只补当前已实现能力的文档与部署，不要加新业务模块
2. 输出“验收者从零启动”的步骤
```

---

# 推荐投喂顺序（精简版）

1. 阶段 1
2. 阶段 2
3. 阶段 3
4. 阶段 4
5. 阶段 5
6. 阶段 6
7. 阶段 7

---

# 建议你怎么使用这份任务框架

最稳妥的方式是：

1. 每次只投喂一个阶段 prompt 给 code agent
2. 要求它输出：
   - 改了哪些文件
   - 跑了哪些命令
   - 哪些通过 / 哪些没通过
   - 下一步建议
3. 你做人类 reviewer，只验收“当前阶段是否达标”
