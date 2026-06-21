# touchgal-donation

一个捐赠站点，包含 Nuxt 4 前端、Go Fiber v3 API、PostgreSQL 存储，并使用 Docker Compose 进行本地编排。公开站点支持用户创建捐赠订单、查看近期公开捐赠、查询支付状态；可配置的管理后台用于管理捐赠档位、支付方式、捐赠记录和站点设置。

## 项目文档

- 开发指南：`docs/development.md`
- 项目 skill：`.omp/skills/touchgal-donation/SKILL.md`

## 本地启动

```bash
cp .env.example .env
# 首次启动前编辑 .env：
# - 将 INITIAL_ADMIN_PASSWORD 改为非默认密码
# - 将 SESSION_SECRET 和 CSRF_SECRET 改为相互独立、长度不少于 32 字节的随机值
# - 可按需修改 INITIAL_ADMIN_BASE_PATH

docker compose up --build
```

也可以使用等价的 Make 目标：

```bash
make dev    # docker compose up --build
make up     # docker compose up -d --build
make down
make logs
make test   # API 测试和前端类型检查
make build  # docker compose build
```

## 访问地址

- 前端：`http://localhost:3000`
- API：`http://localhost:8080/api/v1`
- 管理后台：`.env` 中的 `INITIAL_ADMIN_BASE_PATH`，例如 `http://localhost:3000/support-console-9c2e`

浏览器访问 API 的基础地址由 `NUXT_PUBLIC_API_BASE` 配置；Docker 内部的 Nuxt SSR 通过 `NUXT_API_INTERNAL_BASE` 在 Compose 网络中访问 API 服务。

## 默认管理后台行为

启动时，如果数据库中没有管理员账号，API 会使用 `INITIAL_ADMIN_USERNAME` 和 `INITIAL_ADMIN_PASSWORD` 创建第一个管理员。该账号登录后必须修改密码。

如果数据库中没有配置管理后台路径，API 会使用 `INITIAL_ADMIN_BASE_PATH` 初始化。管理后台路径可配置，但它不是唯一安全控制；管理 API 仍然要求 `admin_session` HttpOnly cookie，并且 unsafe method 需要 `X-CSRF-Token`。

生产环境启动时，如果仍使用默认密钥或默认初始管理员密码，服务必须启动失败。

## 支付模式

金额以整数分存储和传输，公开捐赠记录绝不暴露捐赠者邮箱。公开支付状态接口只读；前端不得直接改写捐赠状态。

支持的支付 provider 类型：

- `mock_qr`：仅用于本地测试的开发 QR 流程。
- `static_qr`：展示已配置的二维码图片，适合由管理员手动确认付款。
- `redirect_url`：使用 URL 模板将捐赠者跳转到现有支付页面。
- `wechat_native`、`alipay_f2f`、`stripe_checkout`：真实支付集成；生产使用前必须补齐 provider 凭证、签名验证、幂等 webhook 和事务安全的状态更新。

## 生产安全检查清单

使用 `APP_ENV=production` 部署前：

- 替换 `SESSION_SECRET`、`CSRF_SECRET` 和 `INITIAL_ADMIN_PASSWORD`，不要使用示例默认值。
- 使用私有 PostgreSQL 密码，并为生产 `DATABASE_URL` 配置合适的 TLS 设置。
- 按需将 `APP_PUBLIC_URL`、`FRONTEND_ORIGIN`、`NUXT_PUBLIC_API_BASE` 和 `NUXT_API_INTERNAL_BASE` 配置为生产 HTTPS 源或私有服务地址。
- 在 API/数据库边界保持 UTC 时间戳，只在展示边缘做格式化。
- 所有金额保持整数分表示，不要用浮点数表示捐赠金额。
- 不要在公开 API、日志或面向公开展示的导出中包含捐赠者邮箱。
- 使用 HTTPS，确保 `admin_session` 能以 Secure、HttpOnly、SameSite=Lax cookie 发送。
- 修改 `INITIAL_ADMIN_BASE_PATH`，不要把路径隐蔽性当作唯一管理后台保护。
- 只启用已配置的支付方式；更新订单状态前必须验证真实 provider 的 webhook。
- 将上传文件存放在不可执行位置，并备份 PostgreSQL 数据和上传的支付二维码资产。
- 不要记录密码、session token、CSRF token 或支付 provider 密钥。
