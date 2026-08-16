# Coupons Service

API for creating, applying, reading, and deactivating coupons.

## Run

API: `http://localhost:8080`  
Health: `GET /health/live`, `GET /health/ready`
Swagger: `http://localhost:8080/docs/`

Run with pgAdmin:

```bash
make up-tools
```

pgAdmin: `http://localhost:5050`

## API

```text
POST /coupons
GET  /coupons/{code}
POST /coupons/{code}/apply
POST /coupons/{code}/deactivate
```

Money uses integer minor units: `10000 USD` means `$100.00`. Percentage discounts round down to the nearest minor unit. A coupon can be applied only once to the same invoice.

OpenAPI: [`api/openapi.yaml`](api/openapi.yaml)

## Tests

```bash
make check
```

Integration tests require PostgreSQL:

```bash
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/coupons?sslmode=disable' make test-integration
```
