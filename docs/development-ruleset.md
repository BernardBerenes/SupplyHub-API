# Development Ruleset

## 1. Module Structure

### 1.1 Technology Stack

- Fiber (Backend Framework)
    - GORM (Query Builder)
    - Viper (Env reader using .env)
    - MinIO
    - Validator
- PostgreSQL (Primary Database)
- MinIO (Storage)

### 1.2 Architecture

The application must use a **Modular Monolith** architecture.

Each business domain must be organized as an independent module or bounded context.

Each module should have clear responsibilities and boundaries.

### 1.3 Project Structure

The project structure must follow this convention:

```
SupplyHub/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── database/
│   │   └── postgres.go
│   │
│   ├── storage/
│   │   └── minio.go
│   │
│   ├── middleware/
│   │   ├── rate_limiter.go
│   │   ├── auth.go
│   │   └── recovery.go
│   │
│   ├── products/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── usecase.go
│   │   └── handler.go
│   │
│   ├── transactions/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── usecase.go
│   │   └── handler.go
│   │
│   └── users/
│       ├── model.go
│       ├── repository.go
│       ├── usecase.go
│       └── handler.go
│
├── presenter/
│   └── response.go
│
├── .env
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

The `internal` directory contains application modules or bounded contexts.

The `presenter` package contains shared API response functionality used by all modules.

---

## 2. General Development Rules

### 2.1 Business Logic

Business logic must be implemented in the **usecase** layer.

Handlers/controllers must not contain business logic.

The standard flow is:

```
Handler
   ↓
UseCase
   ↓
Repository
   ↓
Database
```

---

### 2.2 Database Access

Database access must only be performed through the **repository** layer.

The handler/controller and usecase must not directly execute database queries.

The repository is responsible for:

- Executing database queries.
- Retrieving data.
- Creating data.
- Updating data.
- Deleting data.

Business rules and decisions must remain in the usecase layer.

---

### 2.3 Validation

Request validation must be performed at the **handler/controller** layer.

Validation must use a dedicated request model.

Example:

```
HTTP Request
     ↓
Request Model
     ↓
Validation
     ↓
UseCase
```

---

### 2.4 Controller Model Access

The handler/controller may directly use models that represent API requests and responses.

Examples:

```
CreateProductRequest
UpdateProductRequest
ProductResponse
```

The handler/controller must not contain business logic simply because it has access to these models.

---

# 3. Naming Convention

## 3.1 Database Table

Database table names must:

- Use plural form.
- Use `snake_case`.

Examples:

```
users
products
transactions
transaction_details
order_items
```

---

## 3.2 Database Field

Database field names must use `snake_case`.

Examples:

```
user_id
product_id
created_at
updated_at
total_amount
discount_percentage
```

---

## 3.3 Variable

Variable names must use `camelCase`.

Examples:

```go
userId
productName
totalAmount
transactionDetail
```

---

## 3.4 Function

Public functions must use **PascalCase**.

Private functions must use **camelCase**.

Examples:

```go
func CreateTransaction() {}
func GetProduct() {}

func validateRequest() {}
func calculateDiscount() {}
```

---

## 3.5 UseCase Naming

The file name stays a single word, lowercase: `usecase.go`.

The Go type/struct and any exported identifier must spell it as two words, PascalCase: `UseCase`.

Variables and struct fields must spell it as two words, camelCase: `useCase`.

Examples:

```go
// usecase.go

type UseCase struct {
}

func NewUseCase() *UseCase {
}
```

```go
type ProductUseCase struct {
}

productUseCase := NewProductUseCase()
```

Do not introduce a separate interface for the UseCase when there is only one implementation; export the struct directly.

---

# 4. Model Naming

Model/struct names must:

- Use singular form.
- Use PascalCase.

Examples:

```go
type Product struct {
}

type TransactionDetail struct {
}

type OrderItem struct {
}
```

Request models should use descriptive names based on their operation:

```go
type CreateProductRequest struct {
}

type UpdateProductRequest struct {
}

type LoginRequest struct {
}
```

Response models should use descriptive names based on the resource:

```go
type ProductResponse struct {
}

type TransactionResponse struct {
}

type LoginResponse struct {
}
```

---

# 5. Constants and Enumerations

Constants and enumerations must use **UPPER_SNAKE_CASE**.

Examples:

```go
const ORDER_STATUS_PENDING = "pending"
const ORDER_STATUS_PACKING = "packing"
const ORDER_STATUS_ON_DELIVERY = "on_delivery"
const ORDER_STATUS_DELIVERED = "delivered"
```

---

# 6. API Endpoint Rules

## 6.1 API Prefix

Every API endpoint must start with:

```
/api/{api_version}
```

Examples:

```
/api/v1/products
/api/v1/transactions
/api/v1/users
```

---

## 6.2 Pagination Endpoint

Endpoints that retrieve data using pagination must use the `POST` method.

Example:

```
POST /api/v1/products/list-paginate
```

---

## 6.3 Endpoint Naming

Endpoint names must use `kebab-case`.

If an endpoint contains multiple words, the words must be separated using `-`.

Examples:

```
list-paginate
transaction-details
order-items
```

---

# 7. CRUD Endpoint Convention

For domains that provide CRUD functionality, the following endpoint convention must be used:

| Operation | Method | Endpoint |
| --- | --- | --- |
| List | GET | `/api/v1/products/list` |
| List Paginate | POST | `/api/v1/products/list-paginate` |
| Create | POST | `/api/v1/products/create` |
| Update | PATCH | `/api/v1/products/update/:id` |
| Delete | DELETE | `/api/v1/products/delete/:id` |

Example:

```
GET    /api/v1/products/list
POST   /api/v1/products/list-paginate
POST   /api/v1/products/create
PATCH  /api/v1/products/update/:id
DELETE /api/v1/products/delete/:id
```

---

# 8. API Response Standard

All API responses must follow a standardized response structure.

Response models and helper functions for API responses must be implemented in a **single shared presenter package**.

Individual domains must not create their own implementation of the common API response wrapper.

---

# 9. Shared Presenter

## 9.1 Single Shared Presenter

The application must have **one shared presenter** responsible for the common API response structure.

The presenter must be reusable across all modules/domains.

Example:

```
/library-api/
├── internal/
│   ├── products/
│   ├── transactions/
│   └── users/
│
└── presenter/
    └── response.go
```

The `presenter` package is a cross-domain component and must not belong to a specific business domain.

---

## 9.2 Presenter Responsibilities

The shared presenter is responsible for:

- Standard success response.
- Standard paginated response.
- Standard error response.
- Validation error formatting.
- Pagination metadata.
- Generic response list mapping.

The presenter must provide reusable generic types and helper functions.

---

## 9.3 Success Response

The standard success response must follow:

```json
{
  "message": "Product retrieved successfully",
  "data": {}
}
```

The common response wrapper should be represented by a generic response model.

Example:

```go
type Success[T any] struct {
    Message          string            `json:"message"`
    Data             T                 `json:"data,omitempty"`
    PaginateMetadata *PaginateMetadata `json:"metadata,omitempty"`
}
```

The `data` field must contain the domain-specific response model.

For example:

```go
type ProductResponse struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Price int64  `json:"price"`
}
```

The presenter wraps the domain response:

```
ProductResponse
      ↓
Success[ProductResponse]
      ↓
HTTP Response
```

---

## 9.4 Success Response Without Data

Endpoints that do not return response data must not include the `data` field.

This applies to operations such as:

- Create
- Update
- Delete

Example:

```json
{
  "message": "Product created successfully"
}
```

The shared presenter must provide a helper for this type of response.

---

## 9.5 Paginated Response

Paginated responses must use the shared `Success` response structure with standardized pagination metadata.

Example:

```json
{
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Laptop",
      "price": 15000000
    }
  ],
  "metadata": {
    "page": 1,
    "size": 10,
    "total": 25,
    "total_page": 3
  }
}
```

Pagination metadata must use the following structure:

```go
type PaginateMetadata struct {
    Page      int   `json:"page"`
    Size      int   `json:"size"`
    Total     int64 `json:"total"`
    TotalPage int   `json:"total_page"`
}
```

---

## 9.6 Error Response

The shared presenter must provide a standardized error response.

Validation error:

```json
{
  "message": "Invalid request",
  "errors": [
    {
      "field": "username",
      "message": "username is required"
    }
  ]
}
```

The error model must be shared across all domains.

Example:

```go
type Error struct {
    Message string      `json:"message,omitempty"`
    Errors  []ErrorItem `json:"errors,omitempty"`
}

type ErrorItem struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

---

## 9.7 Validation Error Formatting

Validation errors must be converted into the shared `ErrorItem` format using the presenter.

The validation error formatting logic must not be duplicated inside individual modules.

Example:

```go
errors := presenter.FormatValidationError(err)

return presenter.ErrorResponse(
    ctx,
    fiber.StatusBadRequest,
    "Invalid request",
    errors,
)
```

---

## 9.8 Internal Error

Internal errors must return a generic response:

```json
{
  "message": "Internal server error"
}
```

Internal implementation details must not be exposed to the client.

Detailed error information must be logged to the application's error log.

The shared presenter may be used to return the standardized response, while error logging remains the responsibility of the appropriate application layer.

---

## 9.9 Response Mapping

The shared presenter must provide generic mapping helpers for converting model/domain data into API response models.

Example:

```go
func MapToResponseList[T any, R any](
    items []T,
    mapper func(T) R,
) []R
```

For paginated data:

```go
func MapToResponseListPaginate[T any, R any](
    items []T,
    total int64,
    page int,
    limit int,
    mapper func(T) R,
) ([]R, *PaginateMetadata)
```

`limit` is the requested page size, used to calculate `total_page`. The returned metadata's `Size` is the actual number of items on the page (`len(items)`), which may be smaller than `limit` on the last page.

These helpers must be reusable across all domains.

---

## 9.10 Domain Response vs Shared Response

The distinction between domain response models and the shared presenter must be maintained.

### Domain

Defines **what data is returned**.

Example:

```go
type ProductResponse struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Price int64  `json:"price"`
}
```

### Shared Presenter

Defines **how the API response is structured**.

Example:

```go
type Success[T any] struct {
    Message string `json:"message"`
    Data    T      `json:"data"`
}
```

Therefore:

```
Product Module
      │
      │ ProductResponse
      ▼
Shared Presenter
      │
      │ Success[ProductResponse]
      ▼
HTTP Response
```

The domain owns the response data model.

The shared presenter owns the common API response wrapper.

---

# 10. Layer Responsibility

The standard application flow must follow:

```
HTTP Request
     │
     ▼
Handler / Controller
     │
     │ Request Model
     │ Validation
     ▼
UseCase
     │
     │ Business Logic
     ▼
Repository
     │
     │ Database Access
     ▼
Database
```

### Handler / Controller

Responsible for:

- Receiving HTTP requests.
- Binding request data.
- Using request models.
- Request validation.
- Calling the usecase.
- Returning responses through the shared presenter.

### UseCase

Responsible for:

- Business logic.
- Business rules.
- Orchestrating application operations.
- Calling repositories.

### Repository

Responsible for:

- Database access.
- Query execution.
- Data persistence.
- Data retrieval.

### Model

Responsible for:

- Domain models.
- Request models.
- Domain-specific response models.

### Presenter

Responsible for:

- Common API response structure.
- Success response.
- Error response.
- Pagination metadata.
- Validation error formatting.
- Generic response mapping.

---

# 11. Module Dependency Rules

Modules should depend on abstractions and shared components rather than directly depending on another module's internal implementation.

A module must not directly access another module's:

- Repository implementation.
- Handler.
- Internal business logic.
- Internal data structures that are not part of its public contract.

Cross-module communication should be performed through clearly defined interfaces or application-level contracts.

Example:

```
products
    │
    ▼
products usecase
    │
    ▼
defined contract
    │
    ▼
other module
```

The purpose of this rule is to preserve the bounded context boundaries within the modular monolith.

---

# 12. Rules Summary

| Area | Standard |
| --- | --- |
| Architecture | Modular Monolith |
| Module | Business domain / bounded context |
| Business Logic | UseCase |
| Database Access | Repository only |
| Request Validation | Handler / Controller |
| Validation Model | Request Model |
| Controller Model Access | Request / Response models |
| Database Table | Plural + `snake_case` |
| Database Field | `snake_case` |
| Variable | `camelCase` |
| Public Function | `PascalCase` |
| Private Function | `camelCase` |
| Model / Struct | Singular + `PascalCase` |
| Constant / Enumeration | `UPPER_SNAKE_CASE` |
| API Prefix | `/api/{api_version}` |
| Pagination | `POST` |
| Multi-word Endpoint | `kebab-case` |
| CRUD List | `GET` |
| CRUD List Paginate | `POST` |
| CRUD Create | `POST` |
| CRUD Update | `PATCH` |
| CRUD Delete | `DELETE` |
| API Response | Shared presenter |
| Success Response | Shared presenter |
| Error Response | Shared presenter |
| Validation Formatting | Shared presenter |
| Pagination Metadata | Shared presenter |
| Response Mapping | Shared presenter |
| Internal Error | Generic response + error logging |

# 13. Rate Limiting

## 13.1 Request Limit

The application must implement rate limiting to limit the number of requests made by a client.

The default rate limit is:

> **100 requests per 5 minutes per client.**
> 

A client must not be allowed to exceed 100 requests within a rolling 5-minute window.

```
Limit  : 100 requests
Window : 5 minutes
Scope  : Per client
```

---

## 13.2 Rate Limiting Scope

Rate limiting must be applied at the **application/API level** and must not be implemented separately inside individual business domains.

For example, the following endpoints must use the same rate-limiting mechanism:

```
/api/v1/products
/api/v1/transactions
/api/v1/users
/api/v1/orders
```

The rate limiter should be implemented as middleware or another centralized mechanism.

Example:

```
HTTP Request
     │
     ▼
Rate Limiter
     │
     ├── Limit exceeded → 429 Too Many Requests
     │
     ▼
Handler / Controller
     │
     ▼
UseCase
     │
     ▼
Repository
     │
     ▼
Database
```

---

## 13.3 Rate Limit Response

When a client exceeds the configured limit, the API must return HTTP status:

```
429 Too Many Requests
```

The response must use the application's shared error presenter.

Example:

```json
{
  "message": "Too many requests"
}
```

The response must not expose internal rate-limiter implementation details.

---

## 13.4 Rate Limiting Configuration

The rate limit values must not be hardcoded throughout the application.

The configuration should be centralized.

Example:

```
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW_MINUTES=5
```

The application should use these configuration values when initializing the rate limiter.

---

## 13.5 Rate Limiter Implementation

The rate limiter implementation must be centralized and reusable.

Individual handlers or modules must not implement their own rate-limiting logic.

Not allowed:

```
internal/
├── products/
│   └── handler.go       ← rate limiting
├── transactions/
│   └── handler.go       ← rate limiting
└── users/
    └── handler.go       ← rate limiting
```

Preferred:

```
middleware/
└── rate_limiter.go
```

or another centralized infrastructure package.

---

## 13.6 Client Identification

The rate limiter must identify requests based on a consistent client identifier.

The initial implementation should use the client's **IP address** as the rate-limiting key.

Example:

```
Client IP
    │
    ▼
Rate Limiter
    │
    └── 100 requests / 5 minutes
```

If the application later requires authenticated-user-based rate limiting, the identification strategy may be extended to use the authenticated user identity.

---

## 13.7 Rate Limit Enforcement

The rate limit must be checked before the request reaches the application handler.

Example:

```
Request
   │
   ▼
Rate Limiter
   │
   ├── Request 1  → Allow
   ├── Request 2  → Allow
   ├── ...
   ├── Request 100 → Allow
   │
   └── Request 101 → Reject
                      │
                      ▼
                  HTTP 429
```

Once the client exceeds the configured limit, subsequent requests within the applicable rate-limit window must be rejected.

---

---

## 13.8 Configuration Changes

The default configuration is:

```
Maximum Requests : 100
Time Window      : 5 minutes
```

Any future changes to the rate limit must be made through centralized configuration rather than modifying individual endpoints.

---

## 13.9 Rate Limiting Rule Summary

| Configuration | Standard |
| --- | --- |
| Maximum Requests | 100 |
| Time Window | 5 minutes |
| Initial Client Identifier | IP address |
| Scope | Application/API |
| Implementation | Centralized middleware |
| Exceeded Limit | HTTP 429 |
| Error Response | Shared presenter |
| Configuration | Centralized |
| Per-Domain Implementation | Not allowed |