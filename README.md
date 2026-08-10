# Gocart

Gocart is a backend **REST API** for an e-commerce platform built with **Go**, **Gin**, and **GORM**. It features **JWT authentication**, **role-based authorization**, **product** and **category management**, **shopping cart** and **checkout workflows**, **order** and **payment** processing, a **merchant dashboard** for product and fulfillment management, and **MinIO integration** for product image storage.

## Key Features

### Authentication and Users

- Register a new **customer**.
- Log in with **email or username** and **password**.
- Receive a signed **JWT access token**.
- Access the authenticated **profile endpoint**.
- Default role assignment is **customer**.

### Catalog Management

- Browse **products** publicly.
- Filter products by **category**, **price range**, **stock status**, and **search term**.
- Sort products by **id**, **name**, **price**, **created_at**, or **stock**.
- Browse **categories** publicly.
- **Admins** can create, update, and delete categories.

### Cart and Checkout

- Create and retrieve a **cart** automatically for authenticated users.
- **Add**, **update**, **remove**, and **clear** cart items.
- Enforce **stock checks** while modifying the cart.
- **Checkout** converts the cart into an order and creates a pending payment; stock is deducted when the payment is processed.
- **Checkout** accepts an optional `idempotency_key` to prevent duplicate orders.
- **Cart totals** and **item counts** are recalculated after cart mutations.
- **Signed payment webhooks** confirm or fail payments; the webhook is the source of truth that finalizes the order, so replaying a delivery is a safe no-op.

### Orders

- List the current user’s **orders**.
- Fetch **order details** by id.
- **Cancel** an order; pending-payment orders cancel directly, confirmed orders also restore product stock.

### Merchants

- Register a **merchant profile** from an authenticated customer account.
- Manage **products** (create, update, delete) scoped to the merchant.
- Fulfill **orders** by advancing them to **shipped** and **delivered**.
- View a **dashboard** with product, order, and revenue statistics.

### Image Uploads

- Upload one or more **product images** with product create and update requests.
- Store image objects in **MinIO** under a product-scoped path.

## Tech Stack

- **Go 1.25**
- **Gin** for **HTTP routing** and **middleware**
- **GORM** for **database access** and **schema migration**
- **PostgreSQL** as the primary database
- **JWT** for authentication
- **bcrypt** for password hashing
- **MinIO** for image storage
- **Zerolog** for structured logging
- **Docker** and **Docker Compose**

## Architecture

Gocart follows a layered backend structure:

```mermaid
flowchart TD
    Client[Client / Frontend / API consumer] --> Router[Gin router]
    Router --> Handlers[HTTP handlers]
    Handlers --> Services[Services]
    Services --> Repositories[Database repositories]
    Services --> Storage[MinIO storage]
    Repositories --> Postgres[(PostgreSQL)]
    Storage --> MinIO[(MinIO)]
```

### Layer Responsibilities

- Handlers parse requests and return HTTP responses.
- Services contain business logic such as validation, cart totals, stock checks, and checkout flow.
- Repositories encapsulate database queries and persistence.
- Storage handles image upload and object deletion in MinIO.
- Middleware enforces authentication and role-based access control.

## Requirements

- **Go 1.25** or newer
- **PostgreSQL 17** or compatible
- **MinIO**
- **Docker** and **Docker Compose** for containerized setup

## Configuration

The application loads environment variables from a local `.env` file unless `GO_MODE=release` is set. The minimum runtime configuration currently used by the codebase is:

| Variable | Required | Notes |
| --- | --- | --- |
| `SERVER_PORT` | Yes | Port number without the leading colon. The app listens on `:SERVER_PORT`. |
| `ENV` | Yes | Use `production` to switch Gin to release mode. |
| `DB_HOST` | Yes | PostgreSQL host. |
| `DB_PORT` | Yes | PostgreSQL port. |
| `DB_USER` | Yes | Database user. |
| `DB_PASSWORD` | Yes | Database password. |
| `DB_NAME` | Yes | Database name. |
| `DB_SSL_MODE` | Yes | Passed into the PostgreSQL DSN. |
| `JWT_SECRET` | Yes | Signing key for JWT tokens. |
| `JWT_EXPIRY` | Yes | Duration string such as `24h` or `168h`. |
| `WEBHOOK_SECRET` | Yes | Shared secret used to verify payment webhook signatures (HMAC-SHA256). |
| `MINIO_ENDPOINT` | Yes | MinIO endpoint, for example `localhost:9000`. |
| `MINIO_ACCESS_KEY` | Yes | MinIO access key. |
| `MINIO_SECRET_KEY` | Yes | MinIO secret key. |
| `MINIO_BUCKET` | Yes | Bucket name used for product images. |
| `MINIO_USE_SSL` | No | Defaults to `false`. |
| `UPLOAD_DIR` | No | Defaults to `./uploads`. |
| `MAX_UPLOAD_SIZE` | Yes | Parsed as an integer. |
| `TOKEN_DURATION_MINUTES` | No | Defaults to `60`. |
| `LOGIN_RATE_LIMIT` | No | Max login attempts per IP within the window. Defaults to `5`. |
| `LOGIN_RATE_WINDOW` | No | Login rate-limit window in seconds. Defaults to `60`. |
| `SEED_ADMIN_EMAIL` | No | If set, seeds an admin user on startup. |
| `SEED_ADMIN_PASSWORD` | No | If set, seeds an admin user on startup. |

### Example `.env`

```env
SERVER_PORT=8080
ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gocart
DB_SSL_MODE=disable

JWT_SECRET=change-me-in-production
JWT_EXPIRY=24h

WEBHOOK_SECRET=change-me-webhook-secret

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=gocart
MINIO_USE_SSL=false

MAX_UPLOAD_SIZE=10485760
TOKEN_DURATION_MINUTES=60

# Optional; defaults to 5 attempts / 60 seconds per IP on /auth/login
LOGIN_RATE_LIMIT=5
LOGIN_RATE_WINDOW=60

# Optional; if unset, no admin is seeded
SEED_ADMIN_EMAIL=admin@gocart.com
SEED_ADMIN_PASSWORD=change-me
```

## Local Development

### 1. Start PostgreSQL and MinIO

The repository includes a `docker-compose.yml` file that starts PostgreSQL, MinIO, and the API container.

### 2. Run the API locally

If you prefer to run the API directly on your machine:

```bash
go run ./cmd/api
```

The server starts on the configured port and auto-migrates these tables:

- **users**
- **categories**
- **products**
- **product_images**
- **merchants**
- **carts**
- **cart_items**
- **orders**
- **order_items**
- **payments**

### 3. Build a local binary

```bash
go build -o gocart ./cmd/api
```

## Docker

### Build the image

```bash
docker build -t gocart .
```

### Run the stack

```bash
docker compose up --build
```

The compose file currently starts:

- `app` on port `8080`
- `postgres` on port `5432`
- `minio` on ports `9000` and `9001`

## Admin Seeding

On startup, the app seeds an admin user if `SEED_ADMIN_EMAIL` and `SEED_ADMIN_PASSWORD` are set and the account does not already exist. If either is unset, seeding is skipped entirely.

Example:

- Email: `admin@gocart.com`
- Password: `change-me`

Change the password before any production use.

## API Overview

### Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### Public Catalog

- `GET /api/v1/products`
- `GET /api/v1/products/:id`
- `GET /api/v1/categories`
- `GET /api/v1/categories/:id`

### Authenticated User Routes

All routes below require `Authorization: Bearer <token>`.

- `GET /api/v1/users/me`
- `PUT /api/v1/users/me/password` — change password (requires `current_password`, `new_password`, `confirm_password`)
- `GET /api/v1/admin/dashboard` — admin-only platform metrics (users, merchants, products, orders, revenue, orders by status)
- `GET /api/v1/cart`
- `POST /api/v1/cart/items`
- `PUT /api/v1/cart/items/:itemID`
- `DELETE /api/v1/cart/items/:itemID`
- `DELETE /api/v1/cart`
- `POST /api/v1/orders/checkout`
- `GET /api/v1/orders`
- `GET /api/v1/orders/:id`
- `PUT /api/v1/orders/:id/cancel`
- `POST /api/v1/payments/:reference/process`
- `GET /api/v1/payments/:reference`
- `POST /api/v1/merchants/register`

### Webhooks

Public callback endpoint — authenticated by the `X-Webhook-Signature` header (HMAC-SHA256 of the raw body using `WEBHOOK_SECRET`), not by JWT.

- `POST /api/v1/webhooks/payments`

### Dev-only (skipped when `ENV=production`)

- `POST /api/v1/dev/simulate-payment` — mimics a payment provider callback by building a signed webhook event.

### Merchant Routes

Merchant routes require a valid JWT for an account with a merchant profile.

- `GET /api/v1/merchants/me`
- `PUT /api/v1/merchants/me`
- `GET /api/v1/merchants/dashboard`
- `GET /api/v1/merchants/products`
- `GET /api/v1/merchants/products/:id`
- `POST /api/v1/merchants/products`
- `PUT /api/v1/merchants/products/:id`
- `DELETE /api/v1/merchants/products/:id`
- `GET /api/v1/merchants/orders`
- `GET /api/v1/merchants/orders/:id`
- `PATCH /api/v1/merchants/orders/:id/status`

### Admin Routes

Admin routes require both a valid JWT and the `admin` role. Admin manages categories only (products are merchant-managed).

- `POST /api/v1/admin/categories`
- `PUT /api/v1/admin/categories/:id`
- `DELETE /api/v1/admin/categories/:id`

## Endpoint Details

### Authentication

#### Register

`POST /api/v1/auth/register`

Example request:

```json
{
  "email": "customer@example.com",
  "username": "customer1",
  "password": "secret123",
  "confirm_password": "secret123",
  "first_name": "Chris",
  "last_name": "Taylor"
}
```

#### Login

`POST /api/v1/auth/login`

Example request:

```json
{
  "username_or_email": "customer@example.com",
  "password": "secret123"
}
```

### Products

`GET /api/v1/products` supports these query parameters:

- `page` default `1`
- `page_size` default `10`, max `100`
- `sort` default `created_at`
- `order` `asc` or `desc`, default `desc`
- `category_id`
- `min_price`
- `max_price`
- `search`
- `in_stock=true`

The list **response** is paginated and includes the **data set**, **total item count**, **current page**, **page size**, and **total pages**.

Product create and update use **multipart form data**. Send product fields as form values and attach images under the `images` field.

## Error Handling

The API returns JSON error payloads and uses standard HTTP status codes for common failures:

- `400 Bad Request` for invalid input
- `401 Unauthorized` for missing or invalid authentication
- `403 Forbidden` for insufficient role permissions
- `404 Not Found` when a resource does not exist
- `409 Conflict` for stock-related cart conflicts
- `500 Internal Server Error` for unexpected failures

## Security Notes

- Passwords are hashed with bcrypt before being stored.
- JWTs are signed with HS256.
- Admin access is enforced by role middleware.
- Product and cart flows check inventory before accepting changes.

## Useful cURL Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email":"customer@example.com",
    "username":"customer1",
    "password":"secret123",
    "confirm_password":"secret123",
    "first_name":"Chris",
    "last_name":"Taylor"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "username_or_email":"customer@example.com",
    "password":"secret123"
  }'
```

### List Products

```bash
curl 'http://localhost:8080/api/v1/products?page=1&page_size=10&sort=price&order=asc&search=laptop'
```

### Add to Cart

```bash
curl -X POST http://localhost:8080/api/v1/cart/items \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

### Checkout

```bash
curl -X POST http://localhost:8080/api/v1/orders/checkout \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "shipping_address": "123 Broad Street, Lagos",
    "idempotency_key": "optional-unique-key"
  }'
```

## API Documentation (Swagger)

When the server is running, interactive API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

The docs are generated from annotations in `cmd/api/main.go`. Regenerate them with:

```bash
make docs
```

## Development Commands

A `Makefile` provides common tasks:

| Command | Description |
| --- | --- |
| `make run` | Run the API locally with `go run ./cmd/api`. |
| `make build` | Build a binary to `bin/api`. |
| `make test` | Run all Go tests. |
| `make lint` | Run `go vet ./...`. |
| `make docs` | Regenerate the Swagger documentation into `docs/`. |

## Testing

The project includes unit tests across the handler, middleware, service, mapper, and repository layers. Run them with:

```bash
make test
# or
go test ./...
```

Repository tests use an in-memory **SQLite** database (via `gorm.io/driver/sqlite`). Because that driver uses the CGO-backed `mattn/go-sqlite3`, the repository tests require **CGO** to be enabled (`CGO_ENABLED=1`). If you build or test with `CGO_ENABLED=0`, exclude the repository package or re-enable CGO.

An importable endpoint-to-endpoint **Postman collection** covering the full flow (register → catalog → admin → merchant → cart → checkout → webhook → orders → fulfillment) is included at `gocart-e2e.postman_collection.json`. Set `baseUrl`, `webhookSecret`, and the admin credentials in the collection variables, then run Login to populate the auth tokens.

## Project Structure

```
cmd/api/             Application entrypoint (config loading, DB setup, router wiring)
docs/                Generated Swagger documentation
internal/
  config/            Environment configuration loading
  dto/               Request/response data transfer objects
  errors/            Centralized error codes and AppError type
  handlers/          HTTP handlers (parse requests, write responses)
  logger/            Structured logging (zerolog)
  mapper/            Model <-> DTO conversions
  middleware/        Auth and role-based access control
  models/            GORM models
  query/             Product filter query parameters
  repositories/      Database persistence and the UnitOfWork transaction wrapper
  routes/            Route registration
  seed/              Admin user seeding
  services/          Business logic
  storage/           MinIO image storage
```
