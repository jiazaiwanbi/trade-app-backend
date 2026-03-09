# AGENTS.md

## 项目身份

- 仓库名：`trade-app-backend`
- 项目类型：Go 语言二手交易平台后端
- 当前状态：空仓库，从 0 到 1 设计与实现
- 目标形态：**可演进的模块化单体（modular monolith）**
- 交付目标：提供清晰、稳定、可测试、可部署的后端基础设施与业务模块，而不是一次性“堆功能”

## 你在本仓库中的角色

你是本仓库的 code agent。你的职责不是随意写代码，而是**严格按蓝图分阶段交付**，保证每一步都可运行、可验证、可回滚。

在开始任何工作前，必须先阅读：

1. 本文件 `AGENTS.md`
2. `docs/PROJECT_BLUEPRINT.md`
3. `docs/CODE_AGENT_TASKS.md`

如果三者有冲突，优先级如下：

1. 当前任务 prompt
2. `AGENTS.md`
3. `docs/CODE_AGENT_TASKS.md`
4. `docs/PROJECT_BLUEPRINT.md`

---

## 总体交付原则

### 1. 只做当前阶段的事

- 严格按照 `docs/CODE_AGENT_TASKS.md` 的阶段顺序推进。
- 不要跨阶段提前实现大量未来功能。
- 可以为后续扩展预留接口，但不要把未来阶段整包做完。
- 除非当前阶段明确要求，否则不要私自引入微服务、事件总线大改造、复杂 CQRS、复杂 DDD 框架。

### 2. 优先保证“可运行”和“可验证”

每次提交都应该满足：

- 能编译
- 能通过基础测试
- 目录结构清晰
- 命名统一
- 配置明确
- 至少有一组可验证命令

### 3. 小步提交，避免大爆炸重构

每次变更应当：

- 聚焦单一目标
- 影响范围可控
- 不顺手“顺便重写半个项目”
- 对新增依赖给出理由
- 对目录或架构变更同步更新文档

### 4. 默认按“生产可用”标准思考

即便是 MVP，也要从一开始考虑：

- 身份认证与授权
- 错误码规范
- 审计与日志
- 幂等
- 配置管理
- 数据迁移
- 可观测性
- 测试与 CI
- 安全边界

---

## 默认技术决策

除非任务明确要求修改，否则默认采用以下技术选型。

### 基础栈

- Go：以 `go.mod` 为准（建议使用本机已安装的稳定版本）
- HTTP：`net/http` + `chi`
- 数据库：`PostgreSQL`
- SQL 访问：`pgx`（课设 MVP 不强制 `sqlc`）
- 数据迁移：`golang-migrate`
- 配置：环境变量 + 配置加载器
- 日志：结构化日志（JSON）
- （后续）缓存/会话/限流：`Redis`
- （后续）对象存储：S3 兼容（本地可用 MinIO）
- 文档：OpenAPI 作为 API 契约来源
- 容器化：`Dockerfile` + `docker-compose.yml`
- 质量检查：`go test`、格式化、迁移检查（`golangci-lint` 可后续补齐）

### 默认项目形态

本项目不是微服务优先，而是：

- **单仓库**
- **单进程后端 API**
- **内部按业务域分模块**
- **边界清晰**
- **未来可拆分，但当前不为拆而拆**

---

## 强制架构约束

### 1. 分层规则

默认分层如下：

- `transport`：HTTP handler、请求响应 DTO、路由装配
- `application`：用例编排、事务边界、权限检查、跨仓储调用
- `domain`：核心业务实体、领域规则、状态机、领域错误
- `repository/adapters`：数据库、缓存、对象存储、第三方适配
- `platform/shared`：配置、日志、数据库连接、通用工具、中间件

### 2. 禁止事项

- 禁止在 handler 中直接写业务逻辑
- 禁止在 handler 中直接操作数据库
- 禁止让 `domain` 依赖 HTTP、ORM、Redis、第三方 SDK
- 禁止随意使用全局变量保存业务状态
- 禁止把数据库表结构直接暴露给 API 响应
- 禁止一个模块直接“穿透”另一个模块的私有存储实现
- 禁止为了省事把所有逻辑堆到 `service.go` 一个文件里

### 3. 推荐依赖方向

推荐依赖方向：

`transport -> application -> domain`
`application -> repository interfaces`
`adapters -> repository implementations`

领域层不反向依赖基础设施层。

### 4. 接口设计规则

- 接口定义在“使用方”一侧，而不是实现方一侧
- 先定义最小接口，再实现
- 不要预先抽象一堆永远不会有第二实现的接口
- 对外部系统（对象存储、短信、邮件、支付）使用 adapter 接口隔离

---

## 业务域拆分规则

默认业务模块如下：

- `auth`：注册、登录、鉴权中间件（课设 MVP 默认 JWT Access Token；不强制 refresh token）
- `user`：用户基础资料（昵称/简介等；不做头像上传）
- `category`：分类列表（可选；先扁平，后续再树形）
- `listing`：商品发布、编辑、上下架、状态流转
- `order`：下单、取消、完成（不接支付、不做物流）

后续模块（课设不要求，本周不做）：

- `media`：头像/商品图片上传、对象存储
- `favorite`：收藏
- `chat`：私信聊天
- `payment`：支付与回调
- `review`：评价
- `report`：举报
- `notification`：站内通知
- `admin`：后台审核/封禁/审计

模块之间可以协作，但必须通过清晰的应用层编排完成，不要互相“偷用内部实现”。

---

## 编码规范

### 命名与文件组织

- 包名短小、语义明确、全小写
- 文件名体现职责，不要 `utils.go`、`common.go` 滥用
- 按功能拆文件，不要单文件过大
- 导出对象必须有明确对外语义
- DTO / Entity / Model 不混用命名

### 函数与方法

- 函数尽量短小、单一职责
- 对复杂逻辑拆出私有辅助函数
- 错误必须显式返回，不吞错
- 除启动阶段不可恢复错误外，避免 `panic`
- 任何 IO 密集逻辑都应接受 `context.Context`

### 错误处理

- 区分领域错误、参数错误、权限错误、资源不存在、冲突错误、内部错误
- HTTP 层统一映射错误码与返回结构
- 错误信息对外不要泄露内部实现细节
- 日志里要保留足够排查信息

### 配置与密钥

- 所有配置来自环境变量或配置文件
- 仓库中只保留 `.env.example`
- 禁止把真实密钥、数据库地址、Token、证书提交到仓库
- 配置结构体要有默认值、校验逻辑和文档说明

---

## API 与数据约束

### API 设计

- 路径前缀统一为 `/api/v1`
- 统一 JSON 响应格式
- 列表接口使用一致的分页协议
- 过滤、排序、关键词搜索参数命名统一
- 变更接口优先使用明确语义，例如：
  - `POST /listings`
  - `PATCH /listings/{id}`
  - `POST /listings/{id}/publish`
  - `POST /orders/{id}/cancel`

### 数据库设计

- 所有表必须有清晰主键策略
- 关键表必须带 `created_at`、`updated_at`
- 需要软删除的表显式设计 `deleted_at`
- 迁移必须可回滚
- 索引要围绕实际查询路径设计
- 订单、支付、消息等关键表必须考虑幂等与审计

### 状态机

涉及状态流转的模块必须显式建模：

- `listing`：draft / pending_review / published / reserved / sold / archived / rejected
- `order`：pending_payment / paid / awaiting_shipment / shipped / received / completed / cancelled / disputed / refunded
- `report`：pending / reviewing / resolved / rejected

不要把状态流转规则散落在多个 handler 中。

---

## 安全规则

### 身份认证

- 课设 MVP：优先使用 JWT Access Token（足够支撑鉴权与资源归属校验）
- （后续）需要“刷新 Token/多端会话/真正登出”时，再引入 Refresh Token 与会话存储（Refresh Token 推荐哈希后入库）
- 敏感接口做鉴权与资源归属校验
- 登录、注册、刷新、发送验证码类接口必须考虑限流

### 输入安全

- 所有外部输入都要校验
- 文件上传必须校验类型、大小、数量、来源
- 严格限制排序字段和过滤字段，避免拼接 SQL 注入风险
- 不信任客户端上传的价格、用户 ID、角色字段

### 审计与风控

- 后台管理动作要有审计日志
- 举报、封禁、状态修改要记录操作者与原因
- 订单创建与支付回调要做幂等
- 对批量接口、聊天接口、上传接口考虑频率限制

---

## 测试规则

### 基础要求

- 核心业务规则必须有单元测试
- handler 至少覆盖成功与主要失败路径
- repository 层关键查询建议配 integration test
- 复杂状态机要有表驱动测试
- 修复 bug 时优先补测试，再修实现

### 不同层的测试重点

- `domain`：业务规则、状态流转、边界条件
- `application`：用例编排、事务、一致性、权限检查
- `transport`：参数绑定、鉴权、返回码、响应结构
- `repository`：查询正确性、索引依赖、事务行为

---

## 可观测性规则

必须逐步建设以下能力：

- 结构化日志
- request id / trace id
- 健康检查
- readiness / liveness
- Prometheus 指标
- 慢查询与错误日志
- 可选 OpenTelemetry 埋点
- 非生产环境可启用 pprof

---

## 交付工作流

每次接到一个阶段任务时，按如下流程执行：

1. 阅读当前阶段说明与验收条件
2. 输出简短执行计划（不超过 8 条）
3. 列出将要修改的文件或新增目录
4. 只实现当前阶段范围内内容
5. 完成后运行对应验证命令
6. 输出：
   - 变更摘要
   - 修改文件列表
   - 执行过的命令
   - 验收结果
   - 剩余风险 / 后续建议

---

## 质量门禁

当仓库具备对应能力后，每次改动默认都应执行：

```bash
go fmt ./...
go test ./...
golangci-lint run   # 若当前阶段尚未引入，则跳过
```

如果当前阶段已引入下列工具，也应同步执行：

```bash
sqlc generate       # 若未引入，则跳过
openapi lint        # 若未引入，则跳过
openapi validate    # 若未引入，则跳过
migrate up
migrate down
```

说明：

- 若某命令尚未在当前阶段建立，则在结果中明确说明“未建立，跳过”
- 不要伪造通过结果
- 不要声称运行了未运行的命令

---

## 目录建议

当仓库初始化完成后，优先采用如下目录结构（可微调，但不要偏离）：

```text
trade-app-backend/
├── AGENTS.md
├── Makefile
├── README.md
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── cmd/
│   ├── api/
│   └── migrate/
├── configs/
├── api/
│   └── openapi/
├── db/
│   ├── migrations/
│   ├── query/             # （可选）使用 sqlc 时再加入
│   └── sqlc/              # （可选）使用 sqlc 时再加入
├── internal/
│   ├── platform/
│   ├── shared/
│   ├── transport/
│   ├── application/
│   ├── domain/
│   ├── repository/
│   └── jobs/
├── pkg/                # 仅放通用、稳定、可复用组件；非必须
├── scripts/
├── deployments/
└── docs/
    ├── PROJECT_BLUEPRINT.md
    └── CODE_AGENT_TASKS.md
```

---

## 文档维护规则

以下场景必须同步更新文档：

- 新增或删除核心模块
- API 契约变化
- 目录结构重大变化
- 技术选型变化
- 新增基础设施依赖
- 认证方式变化
- 订单 / 商品 / 举报状态机变化

至少更新以下之一：

- `README.md`
- `docs/PROJECT_BLUEPRINT.md`
- `api/openapi/*`
- 当前阶段说明中的验收标准

---

## Definition of Done

一个阶段完成，至少满足：

- 当前阶段范围内的代码或脚手架已经落地
- 命名、结构、依赖符合本文件约束
- 有明确的验证命令
- 对外接口或契约有文档
- 没有明显的占位垃圾代码
- 没有未使用的大量依赖
- 没有把未来阶段的复杂实现硬塞进当前阶段
- 输出了下一阶段建议，但未擅自继续实现

---

## 遇到不确定时的决策顺序

如果需求没有写死，按以下优先级决策：

1. 可维护性
2. 一致性
3. 易测试
4. 安全性
5. 性能
6. 开发速度

若需要在“快速实现”和“长期可维护”之间取舍，优先选择后者，但避免过度设计。

---

## 最后要求

- 你不是在做 demo，而是在搭一个未来可持续迭代的后端项目。
- 你不是一次性生成器，而是阶段性交付的工程助手。
- 任何实现都要能解释“为什么这样做”。
- 任何重构都要以当前阶段目标为边界。
- 永远优先遵守本文件和 `docs/CODE_AGENT_TASKS.md`。
