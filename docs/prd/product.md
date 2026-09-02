# PRD — Product

## 1. Overview

### 1.1 Problem Statement

The application is developed as a personal application to manage products and store orders.

Product is one of the core domains used to store information about products that can be sold through the application.

The system must provide capabilities to:

* Create a product.
* Retrieve a list of products.
* Filter products by name.
* Automatically sort products by name.
* Retrieve product details.
* Retrieve products using pagination.
* Update a product.
* Delete a product.
* Store an optional product photo.

Product photos are stored in MinIO, while the database only stores the photo object key/reference.

---

### 1.2 Background and Context

A product contains the following information:

* Unique identifier.
* Product name.
* Product price.
* Optional product photo.

Database:

* PostgreSQL.

Data access:

* GORM Query Builder.
* GORM must not be used as an ORM.

Product photos are stored in the MinIO bucket:

```text
supplyhub
```

under the folder:

```text
products/
```

The object key format is:

```text
products/<generated-file-name>
```

Example:

```text
products/550e8400-e29b-41d4-a716-446655440000.jpg
```

There is no UUID subfolder under `products`.

The database only stores the MinIO object key/reference.

---

# 2. Goals

## 2.1 Business Goals

### BG-01 — Manage Products

The system must provide functionality to manage products as master data.

### BG-02 — Store Product Information

The system must store:

* Name
* Price
* Photo

### BG-03 — Search Products

Users must be able to filter products by name.

### BG-04 — Automatic Product Sorting

The system must automatically sort products by name.

The client must not provide sorting parameters such as:

```text
sort
order
```

### BG-05 — Support Product Pagination

The system must provide a paginated product endpoint.

### BG-06 — Store Product Photos

The system must store product photos in MinIO:

```text
Bucket: supplyhub
Folder: products
```

### BG-07 — Soft Delete Products

Deleted products must use soft delete through the `deleted_at` field.

---

# 3. Use Cases

## 3.1 Create Product

**As a user,**

I want to create a product,

so that the product can be managed and used in transactions.

### Main Flow

1. User opens the product page.
2. User selects create product.
3. User enters the product name.
4. User enters the product price.
5. User may provide a product photo.
6. Client sends the create request.
7. Server validates the request.
8. Server creates a product UUID.
9. If a photo is provided, server stores the photo in MinIO.
10. Server stores the product information in PostgreSQL.
11. Server returns a success message.

### Success Response

```json
{
  "message": "Product created successfully"
}
```

The response must not contain `data`.

---

## 3.2 List Products

**As a user,**

I want to view a list of products,

so that I can see the available products.

### Main Flow

1. Client sends a list request.
2. Server retrieves active products.
3. Server applies the name filter if provided.
4. Server automatically sorts products by name.
5. Server returns the product list.

### Success Response

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

The list response must use the `products` wrapper.

The list endpoint does not return pagination metadata.

---

## 3.3 Product Detail

**As a user,**

I want to view product details,

so that I can see information about a specific product.

### Main Flow

1. Client sends the product UUID.
2. Server searches for the product.
3. Server verifies that the product has not been soft-deleted.
4. Server returns the product detail.

### Success Response

```json
{
  "message": "Product retrieved successfully",
  "data": {
    "product": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Coffee",
      "price": 25000,
      "photo": "products/550e8400-e29b-41d4-a716-446655440000.jpg"
    }
  }
}
```

The detail response must use the `product` wrapper.

---

## 3.4 List Products with Pagination

**As a user,**

I want to retrieve products using pagination,

so that the application does not need to retrieve all products at once.

### Main Flow

1. Client sends page and limit.
2. Client may provide a name filter.
3. Server applies the name filter.
4. Server automatically sorts products by name.
5. Server calculates pagination.
6. Server returns pagination metadata and products.

### Success Response

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 10,
    "total_page": 5,
    "products": []
  }
}
```

### Pagination Definition

| Field        | Description                                     |
| ------------ | ----------------------------------------------- |
| `page`       | Current page                                    |
| `size`       | Number of products returned on the current page |
| `total_item` | Total number of products matching the filter    |
| `total_page` | Total number of pages                           |
| `products`   | Products returned on the current page           |

`size` represents the actual number of products returned on the current page, not the requested `limit`.

For example, if:

```text
limit = 10
```

and the current page contains only 3 products:

```json
{
  "page": 2,
  "size": 3,
  "total_item": 13,
  "total_page": 2,
  "products": []
}
```

---

## 3.5 Update Product

**As a user,**

I want to update a product,

so that I can modify product information.

### Main Flow

1. User selects a product.
2. User modifies product information.
3. User may provide a new photo.
4. Client sends the update request.
5. Server validates the request.
6. Server updates the product.
7. If a new photo is provided, server stores the new photo in MinIO.
8. Server updates the photo reference.
9. Server returns a success message.

### Success Response

```json
{
  "message": "Product updated successfully"
}
```

The response must not contain `data`.

If no new photo is provided, the existing photo must remain unchanged.

---

## 3.6 Delete Product

**As a user,**

I want to delete a product,

so that the product is no longer displayed as an active product.

### Main Flow

1. User selects a product.
2. Client sends the delete request.
3. Server finds the product.
4. Server performs a soft delete.
5. Product is excluded from active product queries.
6. Server returns a success message.

### Success Response

```json
{
  "message": "Product deleted successfully"
}
```

The response must not contain `data`.

---

# 4. Requirements

## 4.1 Functional Requirements

### FR-01 — Create Product

Endpoint:

```http
POST /api/v1/products
```

Content-Type:

```text
multipart/form-data
```

Fields:

| Field   | Required | Description   |
| ------- | -------- | ------------- |
| `name`  | Yes      | Product name  |
| `price` | Yes      | Product price |
| `photo` | No       | Product photo |

Example:

```text
name=Coffee
price=25000
photo=<file>
```

---

### FR-02 — Product Name

Product name:

* Required.
* Maximum 100 characters.
* Must not be empty after trimming.

---

### FR-03 — Product Price

Product price:

* Required.
* Integer.
* Must not be negative.

Database type:

```text
BIGINT
```

---

### FR-04 — Product Photo

Product photo is optional.

A product can be created without a photo.

When no photo is provided:

```text
photo = NULL
```

---

### FR-05 — Product Photo Storage

Product photos must be stored in MinIO using:

```text
Bucket: supplyhub
Folder: products
```

Object key format:

```text
products/<generated-file-name>
```

Example:

```text
products/550e8400-e29b-41d4-a716-446655440000.jpg
```

There must be no UUID directory:

```text
products/<uuid>/<file>
```

is not allowed.

The database stores only the MinIO object key/reference.

---

### FR-06 — List Products

Endpoint:

```http
GET /api/v1/products
```

Optional query parameter:

```text
name
```

Example:

```http
GET /api/v1/products?name=coffee
```

The endpoint must return only active products.

The response format is:

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "products": []
  }
}
```

There must be no pagination metadata.

---

### FR-07 — Filter Products by Name

The list and paginated list endpoints must support filtering by:

```text
name
```

Filtering must be performed at the database level.

Example:

```http
GET /api/v1/products?name=coffee
```

---

### FR-08 — Automatic Product Sorting

Products must automatically be sorted by name ascending.

Conceptually:

```sql
ORDER BY name ASC
```

The client must not send:

```text
sort
order
```

Sorting is fully controlled by the backend.

---

### FR-09 — Product Detail

Endpoint:

```http
GET /api/v1/products/:uuid
```

Example:

```http
GET /api/v1/products/550e8400-e29b-41d4-a716-446655440000
```

Success response:

```json
{
  "message": "Product retrieved successfully",
  "data": {
    "product": {}
  }
}
```

If the product does not exist or has been soft-deleted:

```text
404 Not Found
```

Response:

```json
{
  "message": "Product not found"
}
```

---

### FR-10 — Paginated Product List

Endpoint:

```http
POST /api/v1/products/paginate
```

Request:

```json
{
  "page": 1,
  "limit": 10,
  "name": "coffee"
}
```

Available request fields:

| Field   | Required | Description                         |
| ------- | -------- | ----------------------------------- |
| `page`  | No       | Current page                        |
| `limit` | No       | Maximum number of products per page |
| `name`  | No       | Product name filter                 |

The client must not provide sorting parameters.

Success response:

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 10,
    "total_page": 5,
    "products": []
  }
}
```

---

## FR-11 — Pagination Rules

The pagination `limit` must only accept one of the following values:

```text
10
25
50
100
```

Allowed values:

| `limit` | Description           |
|--------:|-----------------------|
|    `10` | 10 products per page  |
|    `25` | 25 products per page  |
|    `50` | 50 products per page  |
|   `100` | 100 products per page |

Default values:

```text
page = 1
limit = 10
```

`page` must be greater than or equal to `1`.

`limit` must be one of:

```text
10, 25, 50, 100
```

Any other value must be rejected as a validation error.

Examples of invalid values:

```text
limit = 1
limit = 5
limit = 15
limit = 25
limit = 200
```

Pagination calculation:

```text
offset = (page - 1) * limit
```

`total_page` is calculated based on the filtered `total_item`.

The `size` field represents the actual number of products returned on the current page, not the requested `limit`.

For example, if:

```text
limit = 10
```

and the current page contains only 3 products:

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "page": 2,
    "size": 3,
    "total_item": 13,
    "total_page": 2,
    "products": []
  }
}
```

---

### FR-12 — Update Product

Endpoint:

```http
PATCH /api/v1/products/:uuid
```

Content-Type:

```text
multipart/form-data
```

Fields:

| Field   | Required | Description           |
| ------- | -------- | --------------------- |
| `name`  | No       | Updated product name  |
| `price` | No       | Updated product price |
| `photo` | No       | New product photo     |

Only provided fields are updated.

Success response:

```json
{
  "message": "Product updated successfully"
}
```

No product data is returned.

---

### FR-13 — Update Product Photo

When a new photo is uploaded:

1. Validate the photo.
2. Upload the new photo to MinIO.
3. Store the object under `products/`.
4. Update the product photo reference.
5. Remove the old photo after the new photo has been successfully stored and referenced.

The old photo must remain unchanged when no new photo is provided.

---

### FR-14 — Delete Product

Endpoint:

```http
DELETE /api/v1/products/:uuid
```

Delete must use soft delete:

```text
deleted_at = current_timestamp
```

Success response:

```json
{
  "message": "Product deleted successfully"
}
```

No product data is returned.

---

### FR-15 — Product Not Found

If the product UUID does not exist or the product has been soft-deleted:

```text
404 Not Found
```

Response:

```json
{
  "message": "Product not found"
}
```

---

# 5. Database Requirements

## 5.1 Products Table

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price BIGINT NOT NULL,
    photo TEXT NULL,
    created_at TIMESTAMP NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);
```

### Field Definition

| Field        | Type         | Constraint | Description               |
| ------------ | ------------ | ---------- | ------------------------- |
| `id`         | UUID         | PK         | Product unique identifier |
| `name`       | VARCHAR(100) | NOT NULL   | Product name              |
| `price`      | BIGINT       | NOT NULL   | Product price             |
| `photo`      | TEXT         | NULL       | MinIO object key          |
| `created_at` | TIMESTAMP    | NULL       | Creation timestamp        |
| `updated_at` | TIMESTAMP    | NULL       | Last update timestamp     |
| `deleted_at` | TIMESTAMP    | NULL       | Soft delete timestamp     |

---

## 5.2 Database Index

An index may be created for product name:

```sql
CREATE INDEX idx_products_name
ON products (name);
```

Active product queries must filter:

```sql
WHERE deleted_at IS NULL
```

List query:

```sql
SELECT *
FROM products
WHERE deleted_at IS NULL
ORDER BY name ASC;
```

Paginated query:

```sql
SELECT *
FROM products
WHERE deleted_at IS NULL
ORDER BY name ASC
LIMIT ?
OFFSET ?;
```

---

# 6. API Summary

| Method   | Endpoint                    | Description                   |
| -------- | --------------------------- | ----------------------------- |
| `POST`   | `/api/v1/products`          | Create product                |
| `GET`    | `/api/v1/products`          | List products                 |
| `POST`   | `/api/v1/products/paginate` | List products with pagination |
| `GET`    | `/api/v1/products/:uuid`    | Product detail                |
| `PATCH`  | `/api/v1/products/:uuid`    | Update product                |
| `DELETE` | `/api/v1/products/:uuid`    | Delete product                |

---

# 7. API Response Structure

## 7.1 Create

```json
{
  "message": "Product created successfully"
}
```

---

## 7.2 List

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

---

## 7.3 Detail

```json
{
  "message": "Product retrieved successfully",
  "data": {
    "product": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Coffee",
      "price": 25000,
      "photo": "products/550e8400-e29b-41d4-a716-446655440000.jpg"
    }
  }
}
```

---

## 7.4 List Pagination

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 10,
    "total_page": 5,
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

---

## 7.5 Update

```json
{
  "message": "Product updated successfully"
}
```

---

## 7.6 Delete

```json
{
  "message": "Product deleted successfully"
}
```

---

# 8. Non-Functional Requirements

## 8.1 Security

### NFR-01 — Authentication

All product endpoints require a valid JWT access token.

```text
Authorization: Bearer <access_token>
```

Authentication follows the Authentication PRD.

### NFR-02 — Photo Validation

The server must validate product photos before storing them in MinIO.

Validation must include:

* Valid image file.
* Allowed image format.
* Maximum file size.
* File content validation.

The server must not rely solely on the MIME type provided by the client.

### NFR-03 — File Name Generation

The original filename must not be used directly as the MinIO object key.

The server must generate the filename.

Example:

```text
products/<generated-file-name>.jpg
```

### NFR-04 — Input Validation

All client inputs must be validated before processing.

---

## 8.2 Performance

### NFR-05 — Database-side Filtering

Filtering must be performed by PostgreSQL.

### NFR-06 — Database-side Sorting

Sorting must be performed by PostgreSQL.

### NFR-07 — Database-side Pagination

Pagination must be performed using:

```text
LIMIT
OFFSET
```

The application must not retrieve all records and paginate them in memory.

### NFR-08 — Photo Storage

Product photo binary data must not be stored in PostgreSQL.

Only the MinIO object key is stored in the database.

---

## 8.3 Maintainability

### NFR-09 — Layered Architecture

Product domain must follow:

```text
Handler
   │
   ▼
Usecase
   │
   ▼
Repository
   │
   ▼
GORM Query Builder
   │
   ▼
PostgreSQL
```

For photo storage:

```text
Handler
   │
   ▼
Usecase
   │
   ▼
Storage
   │
   ▼
MinIO
```

### NFR-10 — Repository Responsibility

Repository is responsible for:

* Product queries.
* Product detail queries.
* Product creation.
* Product updates.
* Product soft deletion.
* Product pagination.
* Database transactions where required.

Repository must not handle HTTP concerns or MinIO business logic.

### NFR-11 — Usecase Responsibility

Usecase is responsible for product business logic and coordination between repository and storage.

### NFR-12 — Handler Responsibility

Handler is responsible for:

* HTTP request parsing.
* Request validation.
* Multipart form handling.
* Calling usecase.
* Mapping responses.

Handler must not access the database directly.

---

# 9. Out of Scope

The following features are out of scope for the Product MVP:

* Product category.
* Product brand.
* Product SKU.
* Product barcode.
* Product stock.
* Product variants.
* Product unit conversion.
* Multiple product photos.
* Product gallery.
* Product supplier.
* Product purchase price.
* Product discount.
* Bulk product import.
* Bulk product delete.
* Product restore.
* Permanent deletion.
* Product audit history.
* Product rating/review.
* Advanced filtering.
* Full-text search.

---

# 10. Acceptance Criteria

### AC-01 — Create Product

Given valid name and price,

when the user creates a product,

then the product is created and the response only contains a success message.

---

### AC-02 — Create Product Without Photo

Given valid name and price without a photo,

when the user creates a product,

then the product is successfully created with:

```text
photo = NULL
```

---

### AC-03 — Create Product With Photo

Given a valid product photo,

when the user creates a product,

then the photo is stored under:

```text
supplyhub/products/
```

and the database stores its object key.

There must be no UUID subfolder.

---

### AC-04 — List Products

Given active products exist,

when the client calls:

```http
GET /api/v1/products
```

then the response must have:

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "products": []
  }
}
```

---

### AC-05 — Filter Products

Given products exist with different names,

when the client provides:

```text
name
```

then the backend filters products based on name.

---

### AC-06 — Automatic Sorting

Given multiple active products,

when the client requests the product list,

then products are automatically sorted by name ascending.

The client does not provide `sort` or `order`.

---

### AC-07 — Product Detail

Given an active product,

when the client calls:

```http
GET /api/v1/products/:uuid
```

then the response must have:

```json
{
  "message": "Product retrieved successfully",
  "data": {
    "product": {}
  }
}
```

---

### AC-08 — Product Detail Not Found

Given the product does not exist or has been soft-deleted,

when the client requests the product detail,

then the server returns:

```text
404 Not Found
```

---

### AC-09 — Pagination

Given multiple active products,

when the client requests:

```http
POST /api/v1/products/paginate
```

then the response must have:

```json
{
  "message": "Products retrieved successfully",
  "data": {
    "page": 1,
    "size": 2,
    "total_item": 10,
    "total_page": 5,
    "products": []
  }
}
```

---

### AC-10 — Pagination Size

Given `limit = 10`,

when only 2 products are returned on the current page,

then:

```text
size = 2
```

`size` must represent the actual number of products returned.

---

### AC-11 — Pagination Filter

Given a name filter is provided,

when the client requests paginated products,

then filtering must be applied before pagination and `total_item` must represent the number of products matching the filter.

---

### AC-12 — Update Product

Given an existing active product,

when the client sends valid update data,

then the product is updated and the response only contains:

```json
{
  "message": "Product updated successfully"
}
```

---

### AC-13 — Update Without Photo

Given a product has an existing photo,

when the client updates the product without providing a new photo,

then the existing photo remains unchanged.

---

### AC-14 — Update With Photo

Given a product has an existing photo,

when the client provides a new photo,

then the new photo is stored under:

```text
supplyhub/products/
```

and the product's photo reference is updated.

---

### AC-15 — Delete Product

Given an active product,

when the client calls:

```http
DELETE /api/v1/products/:uuid
```

then the product is soft-deleted and the response only contains:

```json
{
  "message": "Product deleted successfully"
}
```

---

### AC-16 — Deleted Product Not Listed

Given a product has been soft-deleted,

when the client requests the product list or pagination,

then the deleted product must not appear in the response.

---

### AC-17 — Invalid Photo

Given an invalid or unsupported photo,

when the client uploads the photo,

then the server rejects the request with a validation error.

---

### AC-18 — Unauthorized Request

Given the product endpoint requires authentication,

when the client does not provide a valid JWT access token,

then the server returns:

```text
401 Unauthorized
```
