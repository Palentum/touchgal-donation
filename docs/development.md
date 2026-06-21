# touchgal-donation 项目开发指南

本项目是捐赠站点：Nuxt 4 前端、Go Fiber v3 API、PostgreSQL/GORM 数据层，通过 Docker Compose 本地编排。公共站点负责创建捐赠订单、展示公开捐赠和查询支付状态；管理后台负责档位、支付方式、订单、导出和站点设置。

## 技术栈与入口

- 前端：`apps/web`，Nuxt 4 + Vue 3 + TypeScript strict + Tailwind CSS 4。
- 后端：`services/api`，Go 1.25 + Fiber v3 + GORM + PostgreSQL。
- 编排：`docker-compose.yml` 启动 `postgres`、`api`、`web`；上传文件挂载在 `uploads_data`。
- 常用命令：`make dev` 本地构建启动，`make test` 运行 API 测试和前端 typecheck，`make build` 构建镜像。

## Context7 文档基线

2026-06-21 使用 Context7 查询并采用以下库文档作为约束来源：

- Nuxt：`/nuxt/nuxt`。`runtimeConfig.public` 才暴露给客户端；私有运行时配置只在服务端可用。通用渲染数据读取优先使用 `useFetch`/`useAsyncData`，它们能避免 SSR/客户端重复请求；表单提交等客户端事件可用 `$fetch`。
- Fiber：`/gofiber/docs`。Fiber v3 推荐集中 ErrorHandler 返回清洗后的 JSON；中间件顺序影响安全行为。若后续引入 Fiber session，必须先注册 session，再注册依赖 session 的 CSRF。
- GORM：`/go-gorm/gorm`。`db.Transaction(func(tx *gorm.DB) error { ... })` 内返回错误会自动回滚；函数正常返回会提交。GORM 支持 AutoMigrate，但本项目使用嵌入式 SQL 迁移，不用 AutoMigrate 管理生产 schema。

## 后端约定

- `cmd/server/main.go` 只做启动编排：加载配置、创建上传目录、连接数据库、执行迁移、seed 管理员、组装服务和路由。
- 配置集中在 `internal/config`。生产环境必须拒绝默认 `SESSION_SECRET`、`CSRF_SECRET`、`INITIAL_ADMIN_PASSWORD`。
- 路由集中在 `internal/http/router.go`。现有顺序是 recover、request id、logger、安全响应头、CORS、上传静态文件、`/api/v1` 路由。
- API 成功响应统一为 `{ data, request_id }`；错误响应统一为 `{ error: { code, message, details? }, request_id }`。
- 业务错误使用 `internal/service/errors.go` 的 `service.Error`，由 Fiber `ErrorHandler` 转换为公开错误。不要把内部错误、密钥、token、支付凭证写入公开响应或日志。
- 管理接口挂在 `/api/v1/admin`，必须同时经过 `AuthRequired` 和 `CSRFRequired`；登录接口只限速，不要求已有会话。
- 数据访问优先通过 repository/service 分层。需要多表一致性时使用 GORM transaction，并从事务函数返回错误触发回滚。
- 数据库 schema 通过 `internal/db/migrations/*.sql` 和 `schema_migrations` 维护；新增 schema 改动应添加显式 SQL 迁移并考虑回滚脚本。
- 时间在 API/数据库边界保持 UTC；展示层按站点时区格式化。
- 金额一律使用整数分，禁止用浮点数表示捐赠金额。

## 前端约定

- API 类型集中在 `apps/web/types/api.ts`，后端响应变化必须同步更新类型。
- 公共接口通过 `apps/web/composables/useApi.ts` 访问；管理接口通过 `apps/web/composables/useAdminApi.ts` 访问。组件不要绕过这些封装直接拼 API 基础地址。
- `useApi` 根据 `import.meta.server` 在 SSR 时使用 `NUXT_API_INTERNAL_BASE`，客户端使用 `NUXT_PUBLIC_API_BASE`。
- `useAdminApi` 负责 unsafe method 的 CSRF 头、管理会话凭据和错误 toast；新增管理写接口应复用这一路径。
- 页面放在 `apps/web/pages`，布局放在 `apps/web/layouts`，可复用 UI 组件放在 `apps/web/components/ui`，领域组件按 `donate`、`admin` 分组。
- 保持 TypeScript strict；新增可空字段必须在模板和脚本里显式处理 pending/error/empty 状态。

## 业务与安全不变量

- 公开捐赠记录不能暴露 donor email。
- 前端只能创建订单和查询状态，不能直接改写支付状态。
- `mock_qr` 只用于本地开发；`static_qr` 用于人工确认；真实支付 provider 上线前必须补齐凭证、签名验证、幂等 webhook 和事务安全状态更新。
- 管理路径可配置，但不能作为唯一安全控制；真正的控制是 `admin_session` HttpOnly cookie 与 unsafe method 的 `X-CSRF-Token`。
- 上传文件应放在非可执行目录；支付二维码资产和 PostgreSQL 数据都需要备份。

## 验证

- 后端逻辑改动：在 `services/api` 运行 `go test ./...`。
- 前端类型或组件改动：在 `apps/web` 运行 `npm run typecheck`。
- 横跨前后端或配置改动：运行 `make test`；涉及 Compose、Dockerfile 或环境变量时再运行相应构建/启动验证。
- 文档或 skill 改动：读取生成文件，确认路径、frontmatter、链接和项目事实一致。

## 文档与 skill 同步规则

项目级 skill 位于 `.omp/skills/touchgal-donation/SKILL.md`。每次修改代码时，如果改动会改变架构、目录、接口契约、环境变量、数据模型、迁移方式、安全不变量、支付流程、验证命令或开发工作流，必须在同一变更中同步更新本文件和该 skill。只改局部实现且现有说明仍准确时，不需要为了形式修改文档或 skill。
