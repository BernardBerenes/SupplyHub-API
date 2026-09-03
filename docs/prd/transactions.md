# PRD — Transaction

## 1. Overview

### 1.1 Problem Statement

The application is developed as a personal application to manage products and store orders.

Transaction is the core domain used to record orders placed against a store.

The system must provide capabilities to:

* Create a transaction.
* Retrieve transactions using pagination.
* Filter transactions by payment status, delivery status, and date.
* Automatically sort transactions by date.
* Update a transaction.
* Delete a transaction.
* Sync the store name captured inside a transaction with the current store name.

The Transaction domain does not provide a list API without pagination and does not provide a detail API.

### 1.2 Background and Context

A transaction captures a snapshot of the store it was created for, together with its payment and delivery progress.

Database:

* PostgreSQL.

Data access:

* GORM Query Builder.
* GORM must not be used as an ORM.

The Transaction domain follows:

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

The Store domain remains the source of truth for store master data. Transaction only holds a denormalized `store` snapshot (`id` and `name`) so that historical transactions keep displaying the store name as it was at the time of the transaction, unless explicitly synced. See [Store PRD](./stores.md).

---

# 2. Goals

## 2.1 Business Goals

### BG-01 — Manage Transactions

The system must provide functionality to create, update, delete, and list transactions.

### BG-02 — Store Transaction Information

The system must store the store snapshot, payment status, delivery status, and date required by the Transaction database schema.

### BG-03 — Filter Transactions

Users must be able to filter transactions by payment status, delivery status, and date.

### BG-04 — Automatic Transaction Sorting

The system must automatically sort transactions by date.

The client must not provide sorting parameters such as:

```text
sort
order
```

### BG-05 — Support Transaction Pagination

The system must provide a paginated transaction endpoint.

### BG-06 — Prevent Duplicate Pending Transactions

The system must prevent a store from having more than one pending transaction on the same date.

### BG-07 — Keep Store Snapshot in Sync

The system must allow the store name captured inside a transaction to be re-synced with the current store name, limited to transactions that are still pending.

### BG-08 — Soft Delete Transactions

Deleted transactions must use soft delete through the `deleted_at` field.

---

# 3. Use Cases

## 3.1 Create Transaction

**As a user,**

I want to create a transaction for a store,

so that an order against that store starts being tracked.

### Main Flow

1. User selects a store and a date.
2. Client sends the create request with `store_id` and `date`.
3. Server validates the request.
4. Server verifies the store exists and is active.
5. Server verifies there is no conflicting pending transaction for the same store and date.
6. Server creates the transaction with `payment_status = UNPAID` and `delivery_status = PENDING`.
7. Server stores a snapshot of the store as `{ "id": ..., "name": ... }`.
8. Server returns a success message.

### Success Response

```json
{
  "message": "Transaction created successfully"
}
```

The response must not contain `data`.

---

## 3.2 List Transactions with Pagination

**As a user,**

I want to retrieve transactions using pagination,

so that the application does not need to retrieve all transactions at once.

### Main Flow

1. Client sends page and limit.
2. Client may provide `payment_status`, `delivery_status`, and/or `date` filters.
3. Server applies the provided filters.
4. Server automatically sorts transactions by date.
5. Server calculates pagination.
6. Server returns pagination metadata and transactions.

### Success Response

```json
{
  "message": "Transactions retrieved successfully",
  "data": {
    "page": 1,
    "size": 10,
    "total_item": 10,
    "total_page": 1,
    "transactions": []
  }
}
```

### Pagination Definition

| Field          | Description                                         |
| -------------- | ---------------------------------------------------- |
| `page`         | Current page                                          |
| `size`         | Number of transactions returned on the current page   |
| `total_item`   | Total transactions matching the filter                |
| `total_page`   | Total number of pages                                  |
| `transactions` | Transactions returned on the current page              |

`size` represents the actual number of transactions returned on the current page, not the requested `limit`.

---

## 3.3 Update Transaction

**As a user,**

I want to update a transaction,

so that I can correct or progress its information.

### Main Flow

1. User selects a transaction.
2. User modifies transaction information.
3. Client sends the update request.
4. Server validates the request.
5. If `store_id` is provided, server verifies the transaction is still `PENDING`.
6. Server updates the transaction.
7. Server returns a success message.

### Success Response

```json
{
  "message": "Transaction updated successfully"
}
```

The response must not contain `data`.

---

## 3.4 Delete Transaction

**As a user,**

I want to delete a transaction,

so that it is no longer displayed as an active transaction.

### Main Flow

1. User selects a transaction.
2. Client sends the delete request.
3. Server finds the transaction.
4. Server performs a soft delete.
5. Transaction is excluded from active transaction queries.
6. Server returns a success message.

### Success Response

```json
{
  "message": "Transaction deleted successfully"
}
```

The response must not contain `data`.

---

## 3.5 Sync Store Names

**As a user,**

I want to sync the store name captured inside pending transactions,

so that a store rename is reflected on transactions that have not started delivery yet.

### Main Flow

1. Client sends the sync request.
2. Server retrieves every active transaction with `delivery_status = PENDING`.
3. For each transaction, server looks up the current name of the store referenced by the snapshot `id`.
4. Server updates the transaction's `store.name` when it differs from the current store name.
5. Server returns a success message.

### Success Response

```json
{
  "message": "Transactions synced successfully"
}
```

The response must not contain `data`.

---

# 4. Requirements

## 4.1 Functional Requirements

### FR-01 — Create Transaction

Endpoint:

```http
POST /api/v1/transactions
```

Request:

```json
{
  "store_id": 1,
  "date": "2026-09-03"
}
```

| Field      | Required | Description                          |
| ---------- | -------- | ------------------------------------ |
| `store_id` | Yes      | Identifier of the store being ordered from |
| `date`     | Yes      | Transaction date, format `YYYY-MM-DD` |

The client must not provide `payment_status` or `delivery_status` on create; these fields are set by the server.

---

### FR-02 — Store Reference Validation

`store_id` must reference an existing, active (non soft-deleted) store.

If the store does not exist or has been soft-deleted, the server must reject the request with:

```text
404 Not Found
```

```json
{
  "message": "Store not found"
}
```

---

### FR-03 — Transaction Date

`date`:

* Required.
* Format: `YYYY-MM-DD`.
* Free choice — the server does not restrict the value to today or any specific range (past or future dates are allowed).

---

### FR-04 — Default Status on Creation

When a transaction is created:

```text
payment_status  = UNPAID
delivery_status = PENDING
```

These defaults are set by the server and cannot be overridden on create.

---

### FR-05 — Store Snapshot

On creation, the server stores a snapshot of the referenced store in the `store` column:

```json
{
  "id": 1,
  "name": "Toko Surya"
}
```

The snapshot is captured at creation time and does not automatically change when the store's name changes. Keeping the snapshot up to date is handled by the Sync endpoint (FR-17).

---

### FR-06 — Duplicate Pending Transaction Prevention

Before creating a transaction, the server must check for an existing active transaction with the same `store_id` and the same `date`.

| Existing transaction (same store, same date) | New transaction | Result |
| --------------------------------------------- | ---------------- | ------ |
| `PENDING`                                      | `PENDING`         | Rejected |
| `ON_DELIVERY`                                  | `PENDING`         | Accepted |
| `DELIVERED`                                    | `PENDING`         | Accepted |
| No existing transaction for that store/date    | `PENDING`         | Accepted |

If a matching `PENDING` transaction already exists for the store and date, the server must reject the request with:

```text
409 Conflict
```

```json
{
  "message": "Store already has a pending transaction on this date"
}
```

This check only compares transactions that share the same `date` as the transaction being created. Transactions on other dates do not affect this check.

---

### FR-07 — List Transactions with Pagination

Endpoint:

```http
POST /api/v1/transactions/paginate
```

Request:

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

Available request fields:

| Field             | Required | Description                          |
| ----------------- | -------- | ------------------------------------ |
| `page`            | No       | Current page                          |
| `limit`           | No       | Maximum transactions per page (`10`, `25`, `50`, `100`) |
| `payment_status`  | No       | Filter by payment status               |
| `delivery_status` | No       | Filter by delivery status              |
| `date_from`       | No       | Filter transactions on or after this date |
| `date_to`         | No       | Filter transactions on or before this date |

Filtering by store name is intentionally not supported.

Success response:

```json
{
  "message": "Transactions retrieved successfully",
  "data": {
    "page": 1,
    "size": 10,
    "total_item": 10,
    "total_page": 1,
    "transactions": []
  }
}
```

---

### FR-08 — Filter by Payment Status

Optional filter:

```text
payment_status
```

Allowed values:

```text
PAID
UNPAID
```

Filtering must be performed at the database level.

---

### FR-09 — Filter by Delivery Status

Optional filter:

```text
delivery_status
```

Allowed values:

```text
PENDING
ON_DELIVERY
DELIVERED
```

Filtering must be performed at the database level.

---

### FR-10 — Filter by Date Range

Optional filters:

```text
date_from
date_to
```

Format: `YYYY-MM-DD` for both.

* `date_from` matches transactions with `date >= date_from`.
* `date_to` matches transactions with `date <= date_to`.
* Either may be provided alone (open-ended range) or both together (inclusive range).
* If both are provided and `date_to` is earlier than `date_from`, the server must reject the request as a validation error.

Filtering must be performed at the database level.

---

### FR-11 — Automatic Transaction Sorting

Transactions must automatically be sorted by date, most recent first.

Conceptually:

```sql
ORDER BY date DESC
```

The client must not send:

```text
sort
order
```

---

### FR-12 — Pagination Rules

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

Any other `limit` value must be rejected as a validation error.

Pagination calculation:

```text
offset = (page - 1) * limit
```

`total_page` is calculated based on the filtered `total_item`.

The `size` field represents the actual number of transactions returned on the current page, not the requested `limit`.

---

### FR-13 — Update Transaction

Endpoint:

```http
PATCH /api/v1/transactions/:uuid
```

Request (all fields optional):

```json
{
  "store_id": 2,
  "payment_status": "PAID",
  "delivery_status": "ON_DELIVERY",
  "date": "2026-09-05"
}
```

Only provided fields are updated. `store_id`, `payment_status`, `delivery_status`, and `date` can all be updated, subject to FR-14.

Success response:

```json
{
  "message": "Transaction updated successfully"
}
```

No transaction data is returned.

---

### FR-14 — Store Update Restricted to Pending Transactions

`store_id` may only be updated while the transaction's current `delivery_status` is `PENDING`.

If the request includes `store_id` and the transaction's current `delivery_status` is not `PENDING`, the server must reject the request with:

```text
409 Conflict
```

```json
{
  "message": "Store can only be changed while the transaction is pending"
}
```

When `store_id` is updated, the server must regenerate the `store` snapshot from the new store.

This restriction only applies to `store_id`. `payment_status`, `delivery_status`, and `date` can be updated regardless of the transaction's current status.

---

### FR-15 — Delete Transaction

Endpoint:

```http
DELETE /api/v1/transactions/:uuid
```

Delete must use soft delete:

```text
deleted_at = current_timestamp
```

Success response:

```json
{
  "message": "Transaction deleted successfully"
}
```

No transaction data is returned.

---

### FR-16 — Transaction Not Found

If a transaction UUID does not exist or the transaction has been soft-deleted:

```text
404 Not Found
```

```json
{
  "message": "Transaction not found"
}
```

---

### FR-17 — Sync Store Names

Endpoint:

```http
POST /api/v1/transactions/sync
```

The request body is empty.

The server must:

1. Retrieve every active (`deleted_at IS NULL`) transaction with `delivery_status = PENDING`.
2. For each transaction, look up the current name of the store referenced by `store.id`, regardless of that store's own soft-delete status.
3. Update `store.name` on the transaction when it differs from the current store name.
4. Also sync the `product.name` snapshot of every active transaction detail belonging to those same `PENDING` transactions — see [Transaction Detail PRD, FR-13](./transaction-details.md).

Transactions with `delivery_status` of `ON_DELIVERY` or `DELIVERED` must not be modified by this endpoint, and neither must their transaction details.

The sync is not scoped by store, date, or transaction — it evaluates all eligible transactions (and their details) in a single call.

Success response:

```json
{
  "message": "Transactions synced successfully"
}
```

---

# 5. Database Requirements

## 5.1 Transactions Table

```sql
CREATE TABLE transactions (
    id               UUID PRIMARY KEY,
    store            JSONB NOT NULL,
    payment_status   VARCHAR(10) NOT NULL DEFAULT 'UNPAID'
                     CHECK (payment_status IN ('PAID', 'UNPAID')),
    delivery_status  VARCHAR(15) NOT NULL DEFAULT 'PENDING'
                     CHECK (delivery_status IN ('PENDING', 'ON_DELIVERY', 'DELIVERED')),
    date             DATE NOT NULL,
    created_at       TIMESTAMP NULL,
    updated_at       TIMESTAMP NULL,
    deleted_at       TIMESTAMP NULL
);
```

### Field Definition

| Field             | Type          | Constraint | Description                                  |
| ----------------- | ------------- | ---------- | --------------------------------------------- |
| `id`              | UUID          | PK         | Transaction unique identifier                  |
| `store`           | JSONB         | NOT NULL   | Store snapshot: `{ "id": ..., "name": ... }`   |
| `payment_status`  | VARCHAR(10)   | NOT NULL   | `PAID` or `UNPAID`                             |
| `delivery_status` | VARCHAR(15)   | NOT NULL   | `PENDING`, `ON_DELIVERY`, or `DELIVERED`       |
| `date`            | DATE          | NOT NULL   | Transaction date                                |
| `created_at`      | TIMESTAMP     | NULL       | Creation timestamp                              |
| `updated_at`      | TIMESTAMP     | NULL       | Last update timestamp                           |
| `deleted_at`      | TIMESTAMP     | NULL       | Soft delete timestamp                           |

`payment_status` and `delivery_status` are implemented as `VARCHAR` with a `CHECK` constraint rather than a native PostgreSQL `ENUM` type, to remain compatible with GORM `AutoMigrate`.

---

## 5.2 Database Index

Indexes may be created to support filtering and sorting:

```sql
CREATE INDEX idx_transactions_date ON transactions (date);
CREATE INDEX idx_transactions_payment_status ON transactions (payment_status);
CREATE INDEX idx_transactions_delivery_status ON transactions (delivery_status);
```

The duplicate-pending check (FR-06) additionally benefits from an index covering `store` lookup by id together with `date` and `delivery_status`; since `store` is JSONB, this may use an expression index on `(store->>'id')` if lookup performance becomes a concern.

Active transaction queries must filter:

```sql
WHERE deleted_at IS NULL
```

Paginated query:

```sql
SELECT *
FROM transactions
WHERE deleted_at IS NULL
ORDER BY date DESC
LIMIT ?
OFFSET ?;
```

---

## 5.3 Soft Delete

Active transaction queries must filter:

```sql
WHERE deleted_at IS NULL
```

Delete operations must update `deleted_at` instead of permanently removing the record.

---

# 6. API Summary

| Method   | Endpoint                        | Description               |
| -------- | -------------------------------- | -------------------------- |
| `POST`   | `/api/v1/transactions`           | Create transaction          |
| `POST`   | `/api/v1/transactions/paginate`  | List transactions with pagination |
| `PATCH`  | `/api/v1/transactions/:uuid`     | Update transaction           |
| `DELETE` | `/api/v1/transactions/:uuid`     | Delete transaction           |
| `POST`   | `/api/v1/transactions/sync`      | Sync store names             |

There is intentionally no plain list endpoint and no detail endpoint:

```http
GET /api/v1/transactions
GET /api/v1/transactions/:uuid
```

---

# 7. API Response Structure

## 7.1 Create

```json
{
  "message": "Transaction created successfully"
}
```

## 7.2 List Pagination

```json
{
  "message": "Transactions retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 12,
    "total_page": 6,
    "transactions": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "store": {
          "id": 1,
          "name": "Toko Surya"
        },
        "payment_status": "UNPAID",
        "delivery_status": "PENDING",
        "date": "2026-09-03"
      }
    ]
  }
}
```

## 7.3 Update

```json
{
  "message": "Transaction updated successfully"
}
```

## 7.4 Delete

```json
{
  "message": "Transaction deleted successfully"
}
```

## 7.5 Sync

```json
{
  "message": "Transactions synced successfully"
}
```

---

# 8. Non-Functional Requirements

## 8.1 Security

### NFR-01 — Authentication

All Transaction endpoints require a valid JWT access token.

```text
Authorization: Bearer <access_token>
```

Authentication follows the Authentication PRD.

### NFR-02 — Input Validation

All client inputs must be validated before processing.

### NFR-03 — Parameterized Queries

Dynamic values must be passed as query parameters.

The application must not concatenate user-provided values directly into SQL.

## 8.2 Performance

### NFR-04 — Database-side Filtering

Filtering must be performed by PostgreSQL.

### NFR-05 — Database-side Sorting

Sorting must be performed by PostgreSQL.

### NFR-06 — Database-side Pagination

Pagination must be performed using:

```text
LIMIT
OFFSET
```

The application must not retrieve all records and paginate them in memory.

## 8.3 Maintainability

### NFR-07 — Layered Architecture

Transaction domain must follow:

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

### NFR-08 — Repository Responsibility

Repository is responsible for:

* Transaction queries.
* Transaction creation.
* Transaction updates.
* Transaction soft deletion.
* Transaction pagination.
* Bulk fetch of `PENDING` transactions and store name lookups for the sync operation.

Repository must not handle HTTP concerns.

### NFR-09 — Usecase Responsibility

Usecase is responsible for Transaction business logic, including:

* Default status assignment on create.
* Duplicate pending transaction validation (FR-06).
* Store update restriction while not pending (FR-14).
* Orchestrating the sync operation (FR-17).

Usecase must not directly access GORM or PostgreSQL.

### NFR-10 — Handler Responsibility

Handler is responsible for:

* HTTP request parsing.
* Request validation.
* Calling usecase.
* Mapping responses.

Handler must not access the database directly.

---

# 9. Out of Scope

The following features are out of scope for the Transaction MVP:

* Transaction line items (`transaction_details`: product, quantity, unit, price).
* Transaction detail API.
* Plain (non-paginated) transaction list API.
* Filtering transactions by store name.
* Re-validating the duplicate-pending rule on update.
* Delivery status transition validation (e.g. preventing `DELIVERED` from reverting to `PENDING`).
* Scoping the sync endpoint to a single transaction or store.
* Transaction restore.
* Permanent deletion.
* Transaction audit history.
* Transaction cancellation/reason tracking.
* Notifications on status change.
* Bulk transaction import.
* Bulk transaction delete.

---

# 10. Acceptance Criteria

### AC-01 — Create Transaction

Given a valid `store_id` and `date`, when the user creates a transaction, then the transaction is created with `payment_status = UNPAID`, `delivery_status = PENDING`, and the response only contains a success message.

### AC-02 — Store Snapshot on Create

Given a store with name "Toko Surya" and id `1`, when a transaction is created for that store, then the transaction's `store` field is stored as `{ "id": 1, "name": "Toko Surya" }`.

### AC-03 — Reject Unknown Store

Given a `store_id` that does not exist or has been soft-deleted, when the user creates a transaction, then the server returns `404 Not Found` with `"Store not found"`.

### AC-04 — Reject Duplicate Pending Transaction

Given a store already has a `PENDING` transaction on `2026-09-03`, when the user creates another transaction for the same store and date, then the server returns `409 Conflict`.

### AC-05 — Allow New Pending After Delivery Started

Given a store has an `ON_DELIVERY` transaction on `2026-09-03`, when the user creates a new transaction for the same store and the same date, then the transaction is created successfully as `PENDING`.

### AC-06 — Different Dates Do Not Conflict

Given a store has a `PENDING` transaction on `2026-09-03`, when the user creates a transaction for the same store on `2026-09-04`, then the transaction is created successfully.

### AC-07 — List Pagination

Given multiple active transactions, when the client calls:

```http
POST /api/v1/transactions/paginate
```

then the response contains:

```text
page
size
total_item
total_page
transactions
```

### AC-08 — Filter by Payment Status, Delivery Status, and Date Range

Given transactions exist with different payment statuses, delivery statuses, and dates, when the client provides `payment_status`, `delivery_status`, and/or `date_from`/`date_to`, then the backend filters transactions accordingly. Filtering by store name is not supported.

### AC-08a — Reject Invalid Date Range

Given `date_from` is later than `date_to`, when the client requests the paginated transaction list, then the server rejects the request as a validation error.

### AC-09 — Automatic Sorting

Given multiple active transactions, when the client requests the paginated transaction list, then transactions are automatically sorted by date, most recent first. The client does not provide `sort` or `order`.

### AC-10 — Pagination Limit

Given a pagination request, when `limit` is provided, then only `10`, `20`, `50`, or `100` are accepted. Any other value is rejected.

### AC-11 — Update Transaction

Given an existing active transaction, when the client sends valid update data, then the transaction is updated and the response only contains a success message.

### AC-12 — Reject Store Change When Not Pending

Given a transaction with `delivery_status = ON_DELIVERY` or `DELIVERED`, when the client attempts to update `store_id`, then the server returns `409 Conflict` and the store is not changed.

### AC-13 — Allow Store Change When Pending

Given a transaction with `delivery_status = PENDING`, when the client updates `store_id`, then the transaction's `store` snapshot is regenerated from the new store.

### AC-14 — Delete Transaction

Given an active transaction, when the client calls:

```http
DELETE /api/v1/transactions/:uuid
```

then the transaction is soft-deleted and the response only contains a success message.

### AC-15 — Deleted Transaction Not Listed

Given a transaction has been soft-deleted, when the client requests the paginated transaction list, then the deleted transaction must not appear in the response.

### AC-16 — Sync Updates Only Pending Transactions

Given a store's name changes from "Toko Surya" to "Toko Surya Baru", when the sync endpoint is called, then every active `PENDING` transaction referencing that store has its `store.name` updated to "Toko Surya Baru", while `ON_DELIVERY` and `DELIVERED` transactions referencing the same store remain unchanged.

### AC-17 — Sync Runs Across All Transactions

Given multiple stores have pending transactions, when the sync endpoint is called once, then all eligible pending transactions across all stores are synced in that single call.

### AC-18 — Unauthorized Request

Given the Transaction endpoint requires authentication, when the client does not provide a valid JWT access token, then the server returns `401 Unauthorized`.
