# PRD — Store

## 1. Overview

### 1.1 Problem Statement

The application is developed as a personal application to manage products and store orders.

Store is a core domain used to manage information about stores where products can be sold or orders can be associated.

The system must provide capabilities to:

* Create a store.
* Retrieve a list of stores.
* Filter stores by name.
* Automatically sort stores by name.
* Retrieve stores using pagination.
* Update a store.
* Delete a store.

The Store domain does not provide a detail API.

### 1.2 Background and Context

Store information is managed as master data.

Database:

* PostgreSQL.

Data access:

* GORM Query Builder.
* GORM must not be used as an ORM.

The Store domain follows:

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

---

# 2. Goals

## 2.1 Business Goals

### BG-01 — Manage Stores

The system must provide functionality to manage stores as master data.

### BG-02 — Store Store Information

The system must store the information required by the Store database schema.

### BG-03 — Search Stores

Users must be able to filter stores by name.

### BG-04 — Automatic Store Sorting

The system must automatically sort stores by name.

The client must not provide sorting parameters such as:

```text
sort
order
```

### BG-05 — Support Store Pagination

The system must provide a paginated store endpoint.

### BG-06 — Soft Delete Stores

Deleted stores must use soft delete through the `deleted_at` field.

---

# 3. Use Cases

## 3.1 Create Store

**As a user,**

I want to create a store,

so that the store can be managed and used when managing orders.

### Main Flow

1. User opens the store page.
2. User selects create store.
3. User enters the required store information.
4. Client sends the create request.
5. Server validates the request.
6. Server creates the store.
7. Server stores the information in PostgreSQL.
8. Server returns a success message.

### Success Response

```json
{
  "message": "Store created successfully"
}
```

The response must not contain `data`.

---

## 3.2 List Stores

**As a user,**

I want to view a list of stores,

so that I can see the available stores.

### Main Flow

1. Client sends a list request.
2. Server retrieves active stores.
3. Server applies the name filter if provided.
4. Server automatically sorts stores by name.
5. Server returns the store list.

### Success Response

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "stores": []
  }
}
```

The list response must use the `stores` wrapper.

The list endpoint does not return pagination metadata.

---

## 3.3 List Stores with Pagination

**As a user,**

I want to retrieve stores using pagination,

so that the application does not need to retrieve all stores at once.

### Main Flow

1. Client sends page and limit.
2. Client may provide a name filter.
3. Server applies the name filter.
4. Server automatically sorts stores by name.
5. Server calculates pagination.
6. Server returns pagination metadata and stores.

### Success Response

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "page": 1,
    "size": 10,
    "total_item": 10,
    "total_page": 1,
    "stores": []
  }
}
```

### Pagination Definition

| Field        | Description                                   |
| ------------ | --------------------------------------------- |
| `page`       | Current page                                  |
| `size`       | Number of stores returned on the current page |
| `total_item` | Total stores matching the filter              |
| `total_page` | Total number of pages                         |
| `stores`     | Stores returned on the current page           |

`size` represents the actual number of stores returned on the current page, not the requested `limit`.

---

## 3.4 Update Store

**As a user,**

I want to update a store,

so that I can modify store information.

### Main Flow

1. User selects a store.
2. User modifies store information.
3. Client sends the update request.
4. Server validates the request.
5. Server updates the store.
6. Server returns a success message.

### Success Response

```json
{
  "message": "Store updated successfully"
}
```

The response must not contain `data`.

---

## 3.5 Delete Store

**As a user,**

I want to delete a store,

so that the store is no longer displayed as an active store.

### Main Flow

1. User selects a store.
2. Client sends the delete request.
3. Server finds the store.
4. Server performs a soft delete.
5. Store is excluded from active store queries.
6. Server returns a success message.

### Success Response

```json
{
  "message": "Store deleted successfully"
}
```

The response must not contain `data`.

---

# 4. Requirements

## 4.1 Functional Requirements

### FR-01 — Create Store

Endpoint:

```http
POST /api/v1/stores
```

The request must contain all required fields defined by the Store database schema.

The server must validate all required fields before creating the store.

### FR-02 — Store Name

Store name is required.

The value must not be empty after trimming.

### FR-03 — List Stores

Endpoint:

```http
GET /api/v1/stores
```

Optional query parameter:

```text
name
```

Example:

```http
GET /api/v1/stores?name=main
```

The endpoint must return only active stores.

Response:

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "stores": []
  }
}
```

There must be no pagination metadata.

### FR-04 — Filter Stores by Name

The list and paginated list endpoints must support filtering by:

```text
name
```

Filtering must be performed at the database level.

### FR-05 — Automatic Store Sorting

Stores must automatically be sorted by name ascending.

Conceptually:

```sql
ORDER BY name ASC
```

The client must not send:

```text
sort
order
```

### FR-06 — Paginated Store List

Endpoint:

```http
POST /api/v1/stores/paginate
```

Request:

```json
{
  "page": 1,
  "limit": 10,
  "name": "main"
}
```

Available request fields:

| Field   | Required | Description             |
| ------- | -------- | ----------------------- |
| `page`  | No       | Current page            |
| `limit` | No       | Maximum stores per page |
| `name`  | No       | Store name filter       |

Success response:

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "page": 1,
    "size": 10,
    "total_item": 10,
    "total_page": 1,
    "stores": []
  }
}
```

### FR-07 — Pagination Rules

The pagination `limit` must only accept:

```text
10
20
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

The `size` field represents the actual number of stores returned on the current page, not the requested `limit`.

### FR-08 — Update Store

Endpoint:

```http
PATCH /api/v1/stores/:uuid
```

Only provided fields are updated.

Success response:

```json
{
  "message": "Store updated successfully"
}
```

No store data is returned.

### FR-09 — Delete Store

Endpoint:

```http
DELETE /api/v1/stores/:uuid
```

Delete must use soft delete:

```text
deleted_at = current_timestamp
```

Success response:

```json
{
  "message": "Store deleted successfully"
}
```

No store data is returned.

### FR-10 — Store Not Found

If a store UUID does not exist or the store has been soft-deleted:

```text
404 Not Found
```

Response:

```json
{
  "message": "Store not found"
}
```

---

# 5. Database Requirements

## 5.1 Stores Table

The Stores table must follow the Store schema defined for the application.

The implementation must not introduce fields that are not defined by the approved Store schema.

### 5.2 Database Naming

Database table names must use plural snake_case.

Example:

```text
stores
```

Database fields must use snake_case.

### 5.3 Soft Delete

Active store queries must filter:

```sql
WHERE deleted_at IS NULL
```

Delete operations must update `deleted_at` instead of permanently removing the record.

---

# 6. API Summary

| Method   | Endpoint                  | Description                 |
| -------- | ------------------------- | --------------------------- |
| `POST`   | `/api/v1/stores`          | Create store                |
| `GET`    | `/api/v1/stores`          | List stores                 |
| `POST`   | `/api/v1/stores/paginate` | List stores with pagination |
| `PATCH`  | `/api/v1/stores/:uuid`    | Update store                |
| `DELETE` | `/api/v1/stores/:uuid`    | Delete store                |

There is intentionally no detail endpoint:

```http
GET /api/v1/stores/:uuid
```

---

# 7. API Response Structure

## 7.1 Create

```json
{
  "message": "Store created successfully"
}
```

## 7.2 List

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "stores": []
  }
}
```

## 7.3 List Pagination

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 12,
    "total_page": 2,
    "stores": []
  }
}
```

## 7.4 Update

```json
{
  "message": "Store updated successfully"
}
```

## 7.5 Delete

```json
{
  "message": "Store deleted successfully"
}
```

---

# 8. Non-Functional Requirements

## 8.1 Security

### NFR-01 — Authentication

All Store endpoints require a valid JWT access token.

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

Store domain must follow:

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

* Store queries.
* Store creation.
* Store updates.
* Store soft deletion.
* Store pagination.
* Database transactions where required.

Repository must not handle HTTP concerns.

### NFR-09 — Usecase Responsibility

Usecase is responsible for Store business logic.

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

The following features are out of scope for the Store MVP:

* Store detail API.
* Store categories.
* Store branches hierarchy.
* Store operating hours.
* Store location/map integration.
* Store inventory.
* Store-specific pricing.
* Store-specific discounts.
* Store-specific users/roles.
* Bulk store import.
* Bulk store delete.
* Store restore.
* Permanent deletion.
* Store audit history.
* Advanced filtering.
* Full-text search.

---

# 10. Acceptance Criteria

### AC-01 — Create Store

Given valid store information, when the user creates a store, then the store is created and the response only contains a success message.

### AC-02 — List Stores

Given active stores exist, when the client calls:

```http
GET /api/v1/stores
```

then the response contains:

```json
{
  "message": "Stores retrieved successfully",
  "data": {
    "stores": []
  }
}
```

### AC-03 — Filter Stores

Given stores exist with different names, when the client provides `name`, then the backend filters stores based on name.

### AC-04 — Automatic Sorting

Given multiple active stores, when the client requests the store list, then stores are automatically sorted by name ascending.

The client does not provide `sort` or `order`.

### AC-05 — No Detail API

The Store domain must not expose:

```http
GET /api/v1/stores/:uuid
```

### AC-06 — Pagination

Given multiple active stores, when the client requests:

```http
POST /api/v1/stores/paginate
```

then the response contains:

```text
page
size
total_item
total_page
stores
```

### AC-07 — Pagination Limit

Given a pagination request, when `limit` is provided, then only these values are accepted:

```text
10
20
50
100
```

Any other value must be rejected.

### AC-08 — Pagination Filter

Given a name filter is provided, when the client requests paginated stores, then filtering must be applied before pagination and `total_item` must represent the number of stores matching the filter.

### AC-09 — Update Store

Given an existing active store, when the client sends valid update data, then the store is updated and the response only contains:

```json
{
  "message": "Store updated successfully"
}
```

### AC-10 — Delete Store

Given an active store, when the client calls:

```http
DELETE /api/v1/stores/:uuid
```

then the store is soft-deleted and the response only contains:

```json
{
  "message": "Store deleted successfully"
}
```

### AC-11 — Deleted Store Not Listed

Given a store has been soft-deleted, when the client requests the store list or pagination, then the deleted store must not appear in the response.

### AC-12 — Unauthorized Request

Given the Store endpoint requires authentication, when the client does not provide a valid JWT access token, then the server returns:

```text
401 Unauthorized
```
