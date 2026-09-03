# PRD — Transaction Detail

## 1. Overview

### 1.1 Problem Statement

Transaction Detail is a sub-domain of Transaction, used to record the line items (products) ordered within a transaction.

The system must provide capabilities to:

* Create a transaction detail under a specific transaction.
* Retrieve transaction details for a transaction using pagination.
* Update a transaction detail.
* Delete a transaction detail.
* Keep the product name captured inside a transaction detail in sync with the current product name, as part of the Transaction sync endpoint.

A transaction detail always belongs to exactly one transaction. There is no standalone (transaction-less) transaction detail, and there is no detail API.

### 1.2 Background and Context

A transaction detail captures a snapshot of the product it was created for, together with the quantity, unit, and price agreed for that line item.

Database:

* PostgreSQL.

Data access:

* GORM Query Builder.
* GORM must not be used as an ORM.

The Transaction Detail domain follows:

```text
Handler
   |
   v
Usecase
   |
   v
Repository
   |
   v
GORM Query Builder
   |
   v
PostgreSQL
```

The Product domain remains the source of truth for product master data. Transaction Detail only holds a denormalized `product` snapshot (`id` and `name`), the same pattern used by [Transaction](./transactions.md) for its `store` snapshot. See also the [Product PRD](./product.md).

---

# 2. Goals

## 2.1 Business Goals

### BG-01 — Manage Transaction Line Items

The system must provide functionality to create, update, delete, and list transaction details scoped to a transaction.

### BG-02 — Store Transaction Detail Information

The system must store the product snapshot, quantity, unit, and price required by the Transaction Detail database schema.

### BG-03 — Support Transaction Detail Pagination

The system must provide a paginated transaction detail endpoint, scoped to a single transaction.

### BG-04 — Keep Product Snapshot in Sync

The system must keep the product name captured inside a transaction detail in sync with the current product name, limited to details that belong to a `PENDING` transaction, as part of the existing Transaction sync endpoint.

### BG-05 — Soft Delete Transaction Details

Deleted transaction details must use soft delete through the `deleted_at` field.

---

# 3. Use Cases

## 3.1 Create Transaction Detail

**As a user,**

I want to add a product line item to a transaction,

so that the transaction records what was ordered.

### Main Flow

1. User selects a transaction and a product.
2. User enters quantity, unit, and price.
3. Client sends the create request with the transaction id in the path.
4. Server validates the request.
5. Server verifies the transaction exists.
6. Server verifies the product exists and is active.
7. Server creates the transaction detail, snapshotting the product as `{ "id": ..., "name": ... }`.
8. Server returns a success message.

### Success Response

```json
{
  "message": "Transaction detail created successfully"
}
```

The response must not contain `data`.

---

## 3.2 List Transaction Details with Pagination

**As a user,**

I want to retrieve the line items of a transaction using pagination,

so that the application does not need to retrieve all line items at once.

### Main Flow

1. Client sends page and limit for a specific transaction.
2. Server verifies the transaction exists.
3. Server calculates pagination.
4. Server returns pagination metadata and transaction details, ordered by creation order.

### Success Response

```json
{
  "message": "Transaction details retrieved successfully",
  "data": {
    "page": 1,
    "size": 10,
    "total_item": 10,
    "total_page": 1,
    "transaction_details": []
  }
}
```

---

## 3.3 Update Transaction Detail

**As a user,**

I want to update a transaction detail,

so that I can correct the product, quantity, unit, or price of a line item.

### Main Flow

1. User selects a transaction detail.
2. User modifies its information.
3. Client sends the update request.
4. Server validates the request.
5. Server updates the transaction detail.
6. Server returns a success message.

### Success Response

```json
{
  "message": "Transaction detail updated successfully"
}
```

The response must not contain `data`.

---

## 3.4 Delete Transaction Detail

**As a user,**

I want to delete a transaction detail,

so that it is no longer part of the transaction.

### Main Flow

1. User selects a transaction detail.
2. Client sends the delete request.
3. Server finds the transaction detail.
4. Server performs a soft delete.
5. Server returns a success message.

### Success Response

```json
{
  "message": "Transaction detail deleted successfully"
}
```

The response must not contain `data`.

---

## 3.5 Sync Product Names

Transaction detail product names are synced as part of the Transaction sync endpoint (see [Transaction PRD, FR-17](./transactions.md)). There is no separate sync endpoint for this domain.

### Main Flow

1. Client calls `POST /api/v1/transactions/sync`.
2. For every active transaction with `delivery_status = PENDING`, the server also retrieves that transaction's active transaction details.
3. For each transaction detail, server looks up the current name of the product referenced by the snapshot `id`.
4. Server updates the transaction detail's `product.name` when it differs from the current product name.

---

# 4. Requirements

## 4.1 Functional Requirements

### FR-01 — Create Transaction Detail

Endpoint:

```http
POST /api/v1/transactions/:transaction_id/details
```

Request:

```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "quantity": 12,
  "unit": "DOZENS",
  "price": 25000
}
```

| Field        | Required | Description                                  |
| ------------ | -------- | --------------------------------------------- |
| `product_id` | Yes      | Identifier of the product being ordered        |
| `quantity`   | Yes      | Quantity ordered, must be greater than `0`     |
| `unit`       | Yes      | One of `PIECES`, `DOZENS`, `BOX`, `CARTON`     |
| `price`      | Yes      | Line item price, must be greater than `0`      |

The `transaction_id` in the path identifies the transaction this detail belongs to; it must not be provided in the request body.

---

### FR-02 — Transaction Reference Validation

`transaction_id` (from the path) must reference an existing, active (non soft-deleted) transaction.

If the transaction does not exist or has been soft-deleted:

```text
404 Not Found
```

```json
{
  "message": "Transaction not found"
}
```

---

### FR-03 — Product Reference Validation

`product_id` must reference an existing, active (non soft-deleted) product.

If the product does not exist or has been soft-deleted:

```text
404 Not Found
```

```json
{
  "message": "Product not found"
}
```

---

### FR-04 — Quantity

`quantity`:

* Required.
* Integer.
* Must be greater than `0`.

---

### FR-05 — Unit

`unit`:

* Required.
* Must be one of:

```text
PIECES
DOZENS
BOX
CARTON
```

Any other value must be rejected as a validation error.

---

### FR-06 — Price

`price`:

* Required.
* Integer.
* Must be greater than `0`.

---

### FR-07 — Product Snapshot

On creation, the server stores a snapshot of the referenced product in the `product` column:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Coffee"
}
```

The snapshot is captured at creation time and does not automatically change when the product's name changes. Keeping the snapshot up to date is handled by the Transaction sync endpoint (see 3.5).

---

### FR-08 — List Transaction Details with Pagination

Endpoint:

```http
POST /api/v1/transactions/:transaction_id/details/paginate
```

Request:

```json
{
  "page": 1,
  "limit": 10
}
```

| Field   | Required | Description                              |
| ------- | -------- | ------------------------------------------ |
| `page`  | No       | Current page                                |
| `limit` | No       | Maximum transaction details per page (`10`, `25`, `50`, `100`) |

`transaction_id` in the path must reference an existing, active transaction; otherwise `404 Not Found` with `"Transaction not found"`.

There is no filter beyond scoping to `transaction_id`. Results are ordered by creation order (oldest first).

Success response:

```json
{
  "message": "Transaction details retrieved successfully",
  "data": {
    "page": 1,
    "size": 10,
    "total_item": 10,
    "total_page": 1,
    "transaction_details": []
  }
}
```

---

### FR-09 — Pagination Rules

The pagination `limit` must only accept:

```text
10
25
50
100
```

Default values:

```text
page = 1
limit = 10
```

`page` must be greater than or equal to `1`.

Pagination calculation:

```text
offset = (page - 1) * limit
```

The `size` field represents the actual number of transaction details returned on the current page, not the requested `limit`.

---

### FR-10 — Update Transaction Detail

Endpoint:

```http
PATCH /api/v1/transactions/:transaction_id/details/:uuid
```

Request (all fields optional):

```json
{
  "product_id": "660e8400-e29b-41d4-a716-446655440000",
  "quantity": 24,
  "unit": "BOX",
  "price": 30000
}
```

Only provided fields are updated. There is no restriction based on the parent transaction's `delivery_status` — a detail may be updated regardless of the transaction's current status.

If `quantity` is provided, it must be greater than `0`. If `price` is provided, it must be greater than `0`. If `unit` is provided, it must be one of `PIECES`, `DOZENS`, `BOX`, `CARTON`. If `product_id` is provided, it must reference an existing, active product, and the server regenerates the `product` snapshot.

Success response:

```json
{
  "message": "Transaction detail updated successfully"
}
```

No transaction detail data is returned.

---

### FR-11 — Delete Transaction Detail

Endpoint:

```http
DELETE /api/v1/transactions/:transaction_id/details/:uuid
```

Delete must use soft delete:

```text
deleted_at = current_timestamp
```

There is no restriction based on the parent transaction's `delivery_status`.

Success response:

```json
{
  "message": "Transaction detail deleted successfully"
}
```

No transaction detail data is returned.

---

### FR-12 — Transaction Detail Not Found

If a transaction detail UUID does not exist, has been soft-deleted, or does not belong to the `transaction_id` in the path:

```text
404 Not Found
```

```json
{
  "message": "Transaction detail not found"
}
```

---

### FR-13 — Product Name Sync

The existing `POST /api/v1/transactions/sync` endpoint (see [Transaction PRD, FR-17](./transactions.md)) must, for every active transaction with `delivery_status = PENDING`, also sync the `product.name` of that transaction's active transaction details with the current name of the referenced product.

Transaction details belonging to `ON_DELIVERY` or `DELIVERED` transactions must not be modified.

There is no separate sync endpoint for this domain.

---

# 5. Database Requirements

## 5.1 Transaction Details Table

```sql
CREATE TABLE transaction_details (
    id             UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    product        JSONB NOT NULL,
    quantity       BIGINT NOT NULL,
    unit           VARCHAR(10) NOT NULL
                   CHECK (unit IN ('PIECES', 'DOZENS', 'BOX', 'CARTON')),
    price          BIGINT NOT NULL,
    created_at     TIMESTAMP NULL,
    updated_at     TIMESTAMP NULL,
    deleted_at     TIMESTAMP NULL
);
```

### Field Definition

| Field            | Type         | Constraint | Description                                     |
| ---------------- | ------------ | ---------- | ------------------------------------------------ |
| `id`             | UUID         | PK         | Transaction detail unique identifier              |
| `transaction_id` | UUID         | NOT NULL   | Identifier of the parent transaction               |
| `product`        | JSONB        | NOT NULL   | Product snapshot: `{ "id": ..., "name": ... }`     |
| `quantity`       | BIGINT       | NOT NULL   | Quantity ordered, greater than `0`                  |
| `unit`           | VARCHAR(10)  | NOT NULL   | `PIECES`, `DOZENS`, `BOX`, or `CARTON`             |
| `price`          | BIGINT       | NOT NULL   | Line item price, greater than `0`                   |
| `created_at`     | TIMESTAMP    | NULL       | Creation timestamp                                  |
| `updated_at`     | TIMESTAMP    | NULL       | Last update timestamp                               |
| `deleted_at`     | TIMESTAMP    | NULL       | Soft delete timestamp                               |

`unit` is implemented as `VARCHAR` with a `CHECK` constraint rather than a native PostgreSQL `ENUM` type, to remain compatible with GORM `AutoMigrate`, consistent with the Transaction domain's `payment_status`/`delivery_status` columns.

There is intentionally no foreign key constraint enforced at the database level beyond application-level validation, consistent with how the rest of this codebase uses GORM strictly as a query builder; existence of `transaction_id` is validated by the Usecase layer.

---

## 5.2 Database Index

```sql
CREATE INDEX idx_transaction_details_transaction_id ON transaction_details (transaction_id);
```

Active transaction detail queries must filter:

```sql
WHERE deleted_at IS NULL
```

Paginated query, scoped to a transaction:

```sql
SELECT *
FROM transaction_details
WHERE transaction_id = ?
  AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT ?
OFFSET ?;
```

---

## 5.3 Soft Delete

Delete operations must update `deleted_at` instead of permanently removing the record.

---

# 6. API Summary

| Method   | Endpoint                                             | Description                    |
| -------- | ------------------------------------------------------ | -------------------------------- |
| `POST`   | `/api/v1/transactions/:transaction_id/details`          | Create transaction detail          |
| `POST`   | `/api/v1/transactions/:transaction_id/details/paginate`  | List transaction details with pagination |
| `PATCH`  | `/api/v1/transactions/:transaction_id/details/:uuid`      | Update transaction detail            |
| `DELETE` | `/api/v1/transactions/:transaction_id/details/:uuid`      | Delete transaction detail            |

Product name sync is not a separate endpoint — it is part of `POST /api/v1/transactions/sync` (see [Transaction PRD](./transactions.md)).

There is intentionally no detail (single-item) API:

```http
GET /api/v1/transactions/:transaction_id/details/:uuid
```

---

# 7. API Response Structure

## 7.1 Create

```json
{
  "message": "Transaction detail created successfully"
}
```

## 7.2 List Pagination

```json
{
  "message": "Transaction details retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 2,
    "total_page": 1,
    "transaction_details": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
        "product": {
          "id": "660e8400-e29b-41d4-a716-446655440000",
          "name": "Coffee"
        },
        "quantity": 12,
        "unit": "DOZENS",
        "price": 25000
      }
    ]
  }
}
```

## 7.3 Update

```json
{
  "message": "Transaction detail updated successfully"
}
```

## 7.4 Delete

```json
{
  "message": "Transaction detail deleted successfully"
}
```

---

# 8. Non-Functional Requirements

## 8.1 Security

### NFR-01 — Authentication

All Transaction Detail endpoints require a valid JWT access token.

```text
Authorization: Bearer <access_token>
```

Authentication follows the Authentication PRD.

### NFR-02 — Input Validation

All client inputs must be validated before processing.

### NFR-03 — Parameterized Queries

Dynamic values must be passed as query parameters.

## 8.2 Performance

### NFR-04 — Database-side Pagination

Pagination must be performed using:

```text
LIMIT
OFFSET
```

## 8.3 Maintainability

### NFR-05 — Layered Architecture

Transaction Detail domain must follow:

```text
Handler
   |
   v
Usecase
   |
   v
Repository
   |
   v
GORM Query Builder
   |
   v
PostgreSQL
```

### NFR-06 — Module Boundaries

The Transaction Detail Usecase must not directly access the Transaction or Product module's Repository. Cross-module lookups (verifying a transaction exists, verifying a product exists, reading pending transaction ids for sync) must go through each module's Usecase, via a narrow interface owned by the Transaction Detail module.

### NFR-07 — Repository Responsibility

Repository is responsible for:

* Transaction detail queries, scoped by `transaction_id`.
* Transaction detail creation.
* Transaction detail updates.
* Transaction detail soft deletion.
* Transaction detail pagination.

### NFR-08 — Usecase Responsibility

Usecase is responsible for Transaction Detail business logic, including:

* Transaction existence validation.
* Product existence validation and snapshot generation.
* Orchestrating the product-name sync for a given set of transactions.

### NFR-09 — Handler Responsibility

Handler is responsible for:

* HTTP request parsing, including path parameters (`transaction_id`, `uuid`).
* Request validation (`quantity > 0`, `price > 0`, `unit` enum).
* Calling usecase.
* Mapping responses.

---

# 9. Out of Scope

The following features are out of scope for the Transaction Detail MVP:

* Transaction detail detail API (single-item retrieval).
* Restricting create/update/delete based on the parent transaction's `delivery_status`.
* A standalone sync endpoint for transaction details (it is part of the Transaction sync endpoint).
* Filtering or sorting transaction details beyond scoping to `transaction_id`.
* Recalculating or validating transaction totals against transaction details.
* Transaction detail restore.
* Permanent deletion.
* Transaction detail audit history.
* Bulk transaction detail import.
* Bulk transaction detail delete.

---

# 10. Acceptance Criteria

### AC-01 — Create Transaction Detail

Given a valid `product_id`, `quantity`, `unit`, and `price`, when the user creates a transaction detail under an existing transaction, then the detail is created and the response only contains a success message.

### AC-02 — Product Snapshot on Create

Given a product with name "Coffee", when a transaction detail is created for that product, then the detail's `product` field is stored as `{ "id": ..., "name": "Coffee" }`.

### AC-03 — Reject Unknown Transaction

Given a `transaction_id` in the path that does not exist or has been soft-deleted, when the user creates a transaction detail, then the server returns `404 Not Found` with `"Transaction not found"`.

### AC-04 — Reject Unknown Product

Given a `product_id` that does not exist or has been soft-deleted, when the user creates a transaction detail, then the server returns `404 Not Found` with `"Product not found"`.

### AC-05 — Reject Non-Positive Quantity or Price

Given `quantity <= 0` or `price <= 0`, when the user creates or updates a transaction detail, then the server rejects the request as a validation error.

### AC-06 — Reject Invalid Unit

Given a `unit` value other than `PIECES`, `DOZENS`, `BOX`, or `CARTON`, when the user creates or updates a transaction detail, then the server rejects the request as a validation error.

### AC-07 — List Pagination Scoped to Transaction

Given a transaction with multiple transaction details, when the client calls:

```http
POST /api/v1/transactions/:transaction_id/details/paginate
```

then the response contains only details belonging to that transaction, along with `page`, `size`, `total_item`, `total_page`, and `transaction_details`.

### AC-08 — Update Transaction Detail

Given an existing active transaction detail, when the client sends valid update data, then the detail is updated and the response only contains a success message.

### AC-09 — Update Allowed Regardless of Delivery Status

Given a transaction detail belonging to a transaction with `delivery_status = ON_DELIVERY` or `DELIVERED`, when the client updates the detail, then the update succeeds.

### AC-10 — Delete Transaction Detail

Given an active transaction detail, when the client calls:

```http
DELETE /api/v1/transactions/:transaction_id/details/:uuid
```

then the detail is soft-deleted and the response only contains a success message.

### AC-11 — Detail Not Found When Transaction Mismatched

Given a transaction detail that exists but belongs to a different transaction than the one in the path, when the client updates or deletes it, then the server returns `404 Not Found` with `"Transaction detail not found"`.

### AC-12 — Sync Updates Only Details of Pending Transactions

Given a product's name changes from "Coffee" to "Premium Coffee", when `POST /api/v1/transactions/sync` is called, then every active transaction detail referencing that product, belonging to a `PENDING` transaction, has its `product.name` updated to "Premium Coffee", while details belonging to `ON_DELIVERY` or `DELIVERED` transactions remain unchanged.

### AC-13 — Unauthorized Request

Given the Transaction Detail endpoint requires authentication, when the client does not provide a valid JWT access token, then the server returns `401 Unauthorized`.
