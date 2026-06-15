# touchgal-donation

A donation site with a Nuxt 4 frontend, Go Fiber v3 API, PostgreSQL storage, and Docker Compose local orchestration. The public site lets supporters create donation orders, view recent public donations, and check payment status. The configurable admin console manages tiers, payment methods, donations, and site settings.

## Local start

```bash
cp .env.example .env
# Edit .env before first start:
# - set INITIAL_ADMIN_PASSWORD to a non-default password
# - set SESSION_SECRET and CSRF_SECRET to independent random values of at least 32 bytes
# - optionally change INITIAL_ADMIN_BASE_PATH

docker compose up --build
```

Equivalent Make targets are available:

```bash
make dev    # docker compose up --build
make up     # docker compose up -d --build
make down
make logs
make test   # API tests and web typecheck
make build  # docker compose build
```

## Access URLs

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080/api/v1`
- Admin console: the `INITIAL_ADMIN_BASE_PATH` value in `.env`, for example `http://localhost:3000/support-console-9c2e`

The browser-facing API base is configured by `NUXT_PUBLIC_API_BASE`; Nuxt SSR inside Docker uses `NUXT_API_INTERNAL_BASE` to reach the API service over the Compose network.

## Default admin behavior

On startup, if the database has no administrator account, the API creates the first admin from `INITIAL_ADMIN_USERNAME` and `INITIAL_ADMIN_PASSWORD`. That first account must change its password after login.

If the database has no configured admin path, the API initializes it from `INITIAL_ADMIN_BASE_PATH`. The admin path is configurable but is not the only security control; admin APIs still require the `admin_session` HttpOnly cookie and `X-CSRF-Token` on unsafe methods.

Production startup must fail when default secrets or the default initial admin password are still configured.

## Payment modes

Amounts are stored and exchanged as integer cents, and public donation records never expose donor email addresses. Public payment status is read-only; the frontend must not mutate donation status.

Supported payment provider types:

- `mock_qr`: development-only QR flow for local testing.
- `static_qr`: displays a configured QR image and is suitable for manual confirmation by an admin.
- `redirect_url`: redirects supporters to an existing payment page using a URL template.
- `wechat_native`, `alipay_f2f`, `stripe_checkout`: real payment integrations that require provider credentials, signature verification, idempotent webhooks, and transaction-safe status updates before production use.

## Production security checklist

Before deploying with `APP_ENV=production`:

- Replace `SESSION_SECRET`, `CSRF_SECRET`, and `INITIAL_ADMIN_PASSWORD`; never use the example defaults.
- Use a private PostgreSQL password and a production `DATABASE_URL` with appropriate TLS settings.
- Set `APP_PUBLIC_URL`, `FRONTEND_ORIGIN`, `NUXT_PUBLIC_API_BASE`, and `NUXT_API_INTERNAL_BASE` to the production HTTPS origins or private service URL as appropriate.
- Keep timestamps in UTC at the API/database boundary and format them for display only at the edge.
- Keep all money values in integer cents; never use floating point for donation amounts.
- Keep donor email addresses out of public APIs, logs, and exports intended for public display.
- Use HTTPS so `admin_session` can be sent as a Secure, HttpOnly, SameSite=Lax cookie.
- Change `INITIAL_ADMIN_BASE_PATH`; do not rely on obscurity as the only admin protection.
- Enable only configured payment methods and verify all real provider webhooks before updating order status.
- Store uploads in a non-executable location and back up both PostgreSQL data and uploaded payment QR assets.
- Do not log passwords, session tokens, CSRF tokens, or payment provider secrets.
