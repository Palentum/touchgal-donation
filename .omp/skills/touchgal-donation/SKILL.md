---
name: touchgal-donation
description: Use for development work in the touchgal-donation repo: Nuxt 4 frontend, Go Fiber v3 API, GORM/PostgreSQL, Docker Compose, donation/payment/admin invariants, and docs/skill maintenance.
globs:
  - "apps/web/**"
  - "services/api/**"
  - "docker-compose.yml"
  - "Makefile"
  - "README.md"
  - "docs/**"
alwaysApply: true
---

# touchgal-donation skill

## 必读

- 项目文档：`docs/development.md`。
- README：`README.md`。
- 修改代码时，如果改动会改变架构、目录、接口契约、环境变量、数据模型、迁移方式、安全不变量、支付流程、验证命令或开发工作流，必须同步更新 `docs/development.md` 和本 skill。
- 只改局部实现且现有文档仍准确时，不要为了形式修改文档或 skill。

## Context7 基线

使用 Context7 查询当前框架文档后再处理框架/API 相关问题：

- Nuxt：`/nuxt/nuxt`。`runtimeConfig.public` 暴露给客户端；私有 runtime config 只在服务端。SSR 数据读取优先 `useFetch`/`useAsyncData`，客户端事件提交可用 `$fetch`。
- Fiber：`/gofiber/docs`。保持集中 ErrorHandler 和安全中间件顺序；如引入 Fiber session，session 必须先于依赖它的 CSRF。
- GORM：`/go-gorm/gorm`。多表一致性使用 `db.Transaction`，事务函数返回错误即回滚；本项目使用显式 SQL 迁移，不用 AutoMigrate 接管 schema。

## 项目地图

- `services/api/cmd/server/main.go`：启动编排。
- `services/api/internal/config`：环境变量和生产安全校验。
- `services/api/internal/db`：PostgreSQL 连接、嵌入式 SQL 迁移。
- `services/api/internal/http/router.go`：Fiber app、中间件和 `/api/v1` 路由。
- `services/api/internal/http/handler`：HTTP handler。
- `services/api/internal/service`：业务逻辑和公开业务错误。
- `services/api/internal/repository`：GORM 数据访问。
- `services/api/internal/model`：数据库模型和 GORM hook。
- `services/api/internal/service/payment`：支付 provider 抽象与实现。
- `apps/web/types/api.ts`：前后端 API 契约类型。
- `apps/web/composables/useApi.ts`：公共接口封装。
- `apps/web/composables/useAdminApi.ts`：管理接口、CSRF、凭据和 toast 封装。
- `apps/web/pages`、`layouts`、`components`：Nuxt 页面、布局和组件。

## 后端规则

- API 成功响应统一 `{ data, request_id }`；错误响应统一 `{ error: { code, message, details? }, request_id }`。
- 业务错误用 `service.Error` 系列构造函数，由 `HandleError` 转换成公开响应。
- 不向公开响应或日志泄露密码、session、CSRF token、支付密钥、donor email。
- 管理路由必须同时经过 `AuthRequired` 和 `CSRFRequired`；登录路由只限速。
- 金额始终用整数分；时间在 API/数据库边界保持 UTC。
- schema 变更添加 `internal/db/migrations/*.sql`，不要改用 GORM AutoMigrate。
- 涉及多条写入、支付状态、审计或跨表一致性时使用事务。

## 前端规则

- 组件通过 `useApi` / `useAdminApi` 访问后端，不直接散落 `$fetch` 基础地址。
- 后端响应字段变化必须同步 `apps/web/types/api.ts` 和调用点。
- SSR 需要的读取逻辑使用 Nuxt 数据获取能力；管理写操作复用 `useAdminApi` 注入 CSRF 和 credentials。
- 模板要处理 pending/error/empty 状态；保持 TypeScript strict。
- 公开页面不能展示 donor email；感谢页只查询状态，不改写支付状态。

## 业务不变量

- 公开捐赠记录不含 donor email。
- 支付状态只能由后端受控流程或管理后台更新。
- `mock_qr` 只用于开发；真实 provider 上线前必须有凭证、签名验证、幂等 webhook 和事务安全状态更新。
- 管理路径不是安全边界；`admin_session` HttpOnly cookie 与 `X-CSRF-Token` 才是管理写操作控制点。
- 生产环境必须替换默认 secret、管理员初始密码、数据库口令和公开 URL。

## 验证

- 后端改动：`cd services/api && go test ./...`。
- 前端改动：`cd apps/web && npm run typecheck`。
- 跨端、配置或合并验证：`make test`。
- 只改文档/skill：读取变更后的文件，确认 frontmatter、路径和项目事实一致。
