# SupplyHub API

A personal API for managing products, stores, and transactions.

- Framework: [Fiber](https://gofiber.io/) (Go)
- Database: PostgreSQL (via GORM query builder)
- Storage: MinIO (product photos)
- Auth: JWT (Bearer token)

## Setup

```bash
cp .env.example .env
# fill in DB_USER, DB_PASSWORD, JWT_SECRET, MINIO_ACCESS_KEY, MINIO_SECRET_KEY

go run ./cmd/server
```

The server listens on `PORT` (`8000` in `.env.example`; falls back to `3000` if unset). All endpoints below are prefixed with:

```
/api/v1
```

## Authentication

Every endpoint except `POST /api/v1/login` requires:

```
Authorization: Bearer <access_token>
```

Missing or invalid tokens return:

```json
// 401 Unauthorized
{
  "message": "Unauthorized"
}
```

## Response shape

Success responses:

```json
{
  "message": "...",
  "data": { }        // omitted entirely for create/update/delete/sync
}
```

Error responses:

```json
{
  "message": "...",
  "errors": [         // present only for field-level validation errors
    { "field": "name", "message": "name is required" }
  ]
}
```

`500 Internal Server Error` always returns `{ "message": "Internal server error" }` with no further detail.

---

## Table of Contents

- [Auth](#auth)
- [Products](#products)
- [Stores](#stores)
- [Transactions](#transactions)
- [Transaction Details](#transaction-details)

---

## Auth

### Login

```
POST /api/v1/login
```

No auth required.

**Request**

```json
{
  "username": "admin",
  "password": "secret"
}
```

**Success — 200**

```json
{
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"username","message":"username is required"}]}` (also for missing `password`) |
| 401 | `{"message":"Invalid username or password"}` |
| 500 | `{"message":"Internal server error"}` |

---

## Products

All endpoints below require `Authorization: Bearer <access_token>`.

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/api/v1/products` | Create product |
| GET | `/api/v1/products` | List products |
| GET | `/api/v1/products/:uuid` | Product detail |
| POST | `/api/v1/products/paginate` | List products with pagination |
| PATCH | `/api/v1/products/:uuid` | Update product |
| DELETE | `/api/v1/products/:uuid` | Delete product |

### Create Product

```
POST /api/v1/products
Content-Type: multipart/form-data
```

**Request fields**

| Field | Required | Notes |
| --- | --- | --- |
| `name` | Yes | Max 100 chars |
| `price` | Yes | Integer, not negative |
| `photo` | No | Image file (jpeg/png), max 5 MB |

**Success — 200**

```json
{
  "message": "Product created successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"name","message":"name is required and must be at most 100 characters"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"price","message":"price is required and must not be negative"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"photo","message":"photo is not a valid image"}]}` (also: `unsupported photo format: gif`, `photo exceeds maximum size of 5242880 bytes`) |
| 401 | `{"message":"Unauthorized"}` |
| 500 | `{"message":"Internal server error"}` |

### List Products

```
GET /api/v1/products?name=coffee
```

`name` query param is optional (partial match).

**Success — 200**

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "products": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Coffee",
        "price": 25000,
        "photo": "products/550e8400-e29b-41d4-a716-446655440000.jpg"
      }
    ]
  }
}
```

`photo` is `null` if the product has no photo.

### Product Detail

```
GET /api/v1/products/:uuid
```

**Success — 200**

```json
{
  "message": "Product retrieved successfully",
  "data": {
    "product": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Coffee",
      "price": 25000,
      "photo": null
    }
  }
}
```

**Errors**

| Status | Body |
| --- | --- |
| 404 | `{"message":"Product not found"}` |

### List Products (Paginated)

```
POST /api/v1/products/paginate
```

**Request**

```json
{
  "page": 1,
  "limit": 10,
  "name": "coffee"
}
```

All fields optional. `page` defaults to `1`, `limit` defaults to `10`. `limit` must be one of `10`, `25`, `50`, `100`.

**Success — 200**

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "page": 1,
    "size": 1,
    "total_item": 1,
    "total_page": 1,
    "products": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Coffee",
        "price": 25000,
        "photo": null
      }
    ]
  }
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"limit","message":"limit is oneof"}]}` |

### Update Product

```
PATCH /api/v1/products/:uuid
Content-Type: multipart/form-data
```

All fields optional; only provided fields are updated.

| Field | Notes |
| --- | --- |
| `name` | Max 100 chars |
| `price` | Not negative |
| `photo` | Replaces existing photo; old photo is deleted from MinIO |

**Success — 200**

```json
{
  "message": "Product updated successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"price","message":"price must not be negative"}]}` (also `name`/`photo` variants, same wording style as create) |
| 404 | `{"message":"Product not found"}` |

### Delete Product

```
DELETE /api/v1/products/:uuid
```

Soft delete.

**Success — 200**

```json
{
  "message": "Product deleted successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 404 | `{"message":"Product not found"}` |

---

## Stores

All endpoints below require `Authorization: Bearer <access_token>`. There is no store detail endpoint.

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/api/v1/stores` | Create store |
| GET | `/api/v1/stores` | List stores |
| POST | `/api/v1/stores/paginate` | List stores with pagination |
| PATCH | `/api/v1/stores/:uuid` | Update store |
| DELETE | `/api/v1/stores/:uuid` | Delete store |

### Create Store

```
POST /api/v1/stores
```

**Request**

```json
{
  "name": "Toko Surya"
}
```

**Success — 200**

```json
{
  "message": "Store created successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"name","message":"name is required"}]}` |

### List Stores

```
GET /api/v1/stores?name=surya
```

`name` query param is optional (partial match).

**Success — 200**

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "stores": [
      { "id": 1, "name": "Toko Surya" }
    ]
  }
}
```

### List Stores (Paginated)

```
POST /api/v1/stores/paginate
```

**Request**

```json
{
  "page": 1,
  "limit": 10,
  "name": "surya"
}
```

All fields optional. `limit` must be one of `10`, `25`, `50`, `100`.

**Success — 200**

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "page": 1,
    "size": 1,
    "total_item": 1,
    "total_page": 1,
    "stores": [
      { "id": 1, "name": "Toko Surya" }
    ]
  }
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"limit","message":"limit is oneof"}]}` |

### Update Store

```
PATCH /api/v1/stores/:uuid
```

**Request**

```json
{
  "name": "Toko Surya Baru"
}
```

**Success — 200**

```json
{
  "message": "Store updated successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"name","message":"name must not be empty"}]}` |
| 404 | `{"message":"Store not found"}` |

### Delete Store

```
DELETE /api/v1/stores/:uuid
```

Soft delete.

**Success — 200**

```json
{
  "message": "Store deleted successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 404 | `{"message":"Store not found"}` |

---

## Transactions

All endpoints below require `Authorization: Bearer <access_token>`. There is no plain list or detail endpoint.

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/api/v1/transactions` | Create transaction |
| POST | `/api/v1/transactions/paginate` | List transactions with pagination |
| PATCH | `/api/v1/transactions/:uuid` | Update transaction |
| DELETE | `/api/v1/transactions/:uuid` | Delete transaction |
| POST | `/api/v1/transactions/sync` | Sync store/product name snapshots |

### Create Transaction

```
POST /api/v1/transactions
```

**Request**

```json
{
  "store_id": 1,
  "date": "2026-09-03"
}
```

`payment_status` and `delivery_status` are not accepted here — the server always sets them to `UNPAID` and `PENDING`.

**Success — 200**

```json
{
  "message": "Transaction created successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"store_id","message":"store_id is required"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"date","message":"date must be in YYYY-MM-DD format"}]}` |
| 404 | `{"message":"Store not found"}` |
| 409 | `{"message":"Store already has a pending transaction on this date"}` |

### List Transactions (Paginated)

```
POST /api/v1/transactions/paginate
```

**Request**

```json
{
  "page": 1,
  "limit": 10,
  "payment_status": "UNPAID",
  "delivery_status": "PENDING",
  "date_from": "2026-09-01",
  "date_to": "2026-09-30"
}
```

All fields optional. `limit` must be one of `10`, `25`, `50`, `100`. `payment_status` must be `PAID`/`UNPAID`; `delivery_status` must be `PENDING`/`ON_DELIVERY`/`DELIVERED`. Filtering by store name is not supported. Results are always sorted by `date` descending.

**Success — 200**

```json
{
  "message": "Transactions retrieved successfully",
  "data": {
    "page": 1,
    "size": 1,
    "total_item": 1,
    "total_page": 1,
    "transactions": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "store": { "id": 1, "name": "Toko Surya" },
        "payment_status": "UNPAID",
        "delivery_status": "PENDING",
        "date": "2026-09-03"
      }
    ]
  }
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"limit","message":"limit is oneof"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"payment_status","message":"payment_status is oneof"}]}` (delivery_status uses the same "is oneof" wording) |
| 400 | `{"message":"Invalid request","errors":[{"field":"date_from","message":"date_from is datetime"}]}` (date_to uses the same "is datetime" wording) |
| 400 | `{"message":"Invalid request","errors":[{"field":"date_to","message":"date_to must not be before date_from"}]}` |

### Update Transaction

```
PATCH /api/v1/transactions/:uuid
```

**Request** (all fields optional)

```json
{
  "store_id": 2,
  "payment_status": "PAID",
  "delivery_status": "ON_DELIVERY",
  "date": "2026-09-05"
}
```

`store_id` may only be changed while `delivery_status` is still `PENDING`.

**Success — 200**

```json
{
  "message": "Transaction updated successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"store_id","message":"store_id must be valid"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"payment_status","message":"payment_status must be PAID or UNPAID"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"delivery_status","message":"delivery_status must be PENDING, ON_DELIVERY, or DELIVERED"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"date","message":"date must be in YYYY-MM-DD format"}]}` |
| 404 | `{"message":"Transaction not found"}` |
| 404 | `{"message":"Store not found"}` (when `store_id` references a missing/deleted store) |
| 409 | `{"message":"Store can only be changed while the transaction is pending"}` |

### Delete Transaction

```
DELETE /api/v1/transactions/:uuid
```

Soft delete.

**Success — 200**

```json
{
  "message": "Transaction deleted successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 404 | `{"message":"Transaction not found"}` |

### Sync Store/Product Names

```
POST /api/v1/transactions/sync
```

No request body. For every `PENDING` transaction, re-syncs the store name snapshot (`store.name`) and the product name snapshot (`product.name`) of its transaction details, in case the underlying store/product was renamed. `ON_DELIVERY`/`DELIVERED` transactions and their details are left untouched.

**Success — 200**

```json
{
  "message": "Transactions synced successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 500 | `{"message":"Internal server error"}` |

---

## Transaction Details

Line items of a transaction. All endpoints require `Authorization: Bearer <access_token>` and are nested under a transaction. There is no detail (single-item) endpoint.

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/api/v1/transactions/:transaction_id/details` | Create transaction detail |
| POST | `/api/v1/transactions/:transaction_id/details/paginate` | List transaction details with pagination |
| PATCH | `/api/v1/transactions/:transaction_id/details/:uuid` | Update transaction detail |
| DELETE | `/api/v1/transactions/:transaction_id/details/:uuid` | Delete transaction detail |

### Create Transaction Detail

```
POST /api/v1/transactions/:transaction_id/details
```

**Request**

```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "quantity": 12,
  "unit": "DOZENS",
  "price": 25000
}
```

`unit` must be one of `PIECES`, `DOZENS`, `BOX`, `CARTON`. `quantity` and `price` must be greater than `0`. There is no restriction based on the parent transaction's `delivery_status`.

**Success — 200**

```json
{
  "message": "Transaction detail created successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"product_id","message":"product_id is required"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"quantity","message":"quantity must be greater than 0"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"price","message":"price must be greater than 0"}]}` |
| 400 | `{"message":"Invalid request","errors":[{"field":"unit","message":"unit must be PIECES, DOZENS, BOX, or CARTON"}]}` |
| 404 | `{"message":"Transaction not found"}` |
| 404 | `{"message":"Product not found"}` |

### List Transaction Details (Paginated)

```
POST /api/v1/transactions/:transaction_id/details/paginate
```

**Request**

```json
{
  "page": 1,
  "limit": 10
}
```

All fields optional. `limit` must be one of `10`, `25`, `50`, `100`. Results are ordered by creation order (oldest first).

**Success — 200**

```json
{
  "message": "Transaction details retrieved successfully",
  "data": {
    "page": 1,
    "size": 1,
    "total_item": 1,
    "total_page": 1,
    "transaction_details": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "transaction_id": "660e8400-e29b-41d4-a716-446655440000",
        "product": { "id": "550e8400-e29b-41d4-a716-446655440000", "name": "Coffee" },
        "quantity": 12,
        "unit": "DOZENS",
        "price": 25000
      }
    ]
  }
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | `{"message":"Invalid request","errors":[{"field":"limit","message":"limit is oneof"}]}` |
| 404 | `{"message":"Transaction not found"}` |

### Update Transaction Detail

```
PATCH /api/v1/transactions/:transaction_id/details/:uuid
```

**Request** (all fields optional)

```json
{
  "product_id": "660e8400-e29b-41d4-a716-446655440000",
  "quantity": 24,
  "unit": "BOX",
  "price": 30000
}
```

**Success — 200**

```json
{
  "message": "Transaction detail updated successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 400 | Same validation errors as create (`quantity`, `price`, `unit`, `product_id`) |
| 404 | `{"message":"Transaction detail not found"}` (also returned if `:uuid` exists but belongs to a different `transaction_id`) |
| 404 | `{"message":"Product not found"}` (when `product_id` references a missing/deleted product) |

### Delete Transaction Detail

```
DELETE /api/v1/transactions/:transaction_id/details/:uuid
```

Soft delete.

**Success — 200**

```json
{
  "message": "Transaction detail deleted successfully"
}
```

**Errors**

| Status | Body |
| --- | --- |
| 404 | `{"message":"Transaction detail not found"}` |
