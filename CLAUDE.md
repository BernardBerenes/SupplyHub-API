## 1. Project Requirements and Documentation

Before implementing, modifying, or refactoring any feature, Claude must read the relevant project documentation first.

The primary sources of truth are:

```
docs/prd/
docs/development-ruleset.md
```

### 1.1 PRD

Claude must read the relevant documents under:

```
docs/prd/
```

The PRD defines:

- Business requirements.
- Business flow.
- Functional requirements.
- User stories.
- Use cases.
- Feature scope.
- Expected application behavior.

The PRD is the source of truth for **what the application should do**.

Do not invent business requirements that are not defined in the PRD unless explicitly requested.

### 1.2 Development Ruleset

Claude must read:

```
docs/development-ruleset.md
```

The development ruleset defines:

- Project structure.
- Architecture.
- Coding conventions.
- Naming conventions.
- API conventions.
- Layer responsibilities.
- Middleware rules.
- Database access rules.
- Response standards.
- Security standards.
- Other development standards.

The development ruleset is the source of truth for **how the application should be implemented**.

### 1.3 Conflict Resolution

When implementing a feature, follow this priority:

```
Business Requirement
       ↓
docs/prd/
       ↓
docs/development-ruleset.md
       ↓
Existing Project Implementation
```

If the PRD and development ruleset do not provide enough information to make an implementation decision, ask for clarification rather than making assumptions that may affect business behavior or architecture.

---

## 2. Database and Data Access

The application uses **PostgreSQL** as the primary database.

All database access must be performed through the **Repository layer**.

The standard data access flow is:

```
Handler
   ↓
UseCase
   ↓
Repository
   ↓
GORM Query Builder
   ↓
PostgreSQL
```

### 2.1 GORM Usage

The project uses **GORM as a query builder and database abstraction, not as an ORM**.

GORM must only be used for:

- Building SQL queries.
- Executing SQL queries.
- Binding query parameters.
- Scanning query results.
- Managing database transactions when required.

The project must **not rely on GORM ORM features**.

Do not use ORM features such as:

- Model associations.
- `Preload`.
- `Association`.
- Automatic relationship loading.
- GORM lifecycle hooks.
- ORM-based relationship management.
- Implicit association handling.
- Domain models tightly coupled to GORM behavior.

The use of GORM must remain focused on explicit database queries.

**Exception — Schema Migration:** `AutoMigrate` is permitted as the schema migration mechanism (create table if missing, add missing columns if it already exists). It is run once at startup against explicit domain models and is not used for associations, `Preload`, or relationship loading. This is a deliberate, scoped exception to the query-builder-only rule above.

### 2.2 Query Builder

Repositories should use GORM's query builder APIs, including:

- `Table`
- `Select`
- `Where`
- `Joins`
- `Order`
- `Group`
- `Having`
- `Limit`
- `Offset`
- `Count`
- `Create`
- `Updates`
- `Delete`

Example:

```go
var products []Product

err := r.db.
    WithContext(ctx).
    Table("products").
    Select(
        "id",
        "name",
        "price",
        "created_at",
    ).
    Where("status = ?", "active").
    Order("created_at DESC").
    Find(&products).
    Error

if err != nil {
    return nil, err
}

return products, nil
```

### 2.3 Parameterized Queries

All user-provided or dynamic values must be passed using parameterized queries.

Do not concatenate values directly into SQL strings.

Incorrect:

```go
query := "SELECT * FROM products WHERE name = '" + name + "'"
```

Correct:

```go
err := r.db.
    Table("products").
    Where("name = ?", name).
    Find(&products).
    Error
```

This rule applies to all dynamic query values.

### 2.4 Repository Responsibility

The Repository layer is responsible for:

- Constructing database queries.
- Executing database queries.
- Mapping database results.
- Handling database-specific errors.
- Managing database persistence.
- Managing database transactions when required.

The Repository must not contain business logic.

For example, discount calculation must not be implemented inside the Repository.

Incorrect:

```
Repository
└── calculateDiscount()
```

Correct:

```
UseCase
└── calculateDiscount()
```

### 2.5 UseCase Restrictions

The UseCase layer contains business logic but must not directly access the database.

The UseCase must interact with the Repository through defined interfaces.

Incorrect:

```go
func (u *ProductUseCase) GetProduct(id int) error {
    var product Product

    return u.db.
        Table("products").
        Where("id = ?", id).
        First(&product).
        Error
}
```

Correct:

```
UseCase
   ↓
ProductRepository
   ↓
GORM Query Builder
   ↓
PostgreSQL
```

### 2.6 Handler Restrictions

Handlers must never access GORM or PostgreSQL directly.

Incorrect:

```
Handler
   ↓
GORM
   ↓
PostgreSQL
```

Correct:

```
Handler
   ↓
UseCase
   ↓
Repository
   ↓
GORM
   ↓
PostgreSQL
```

### 2.7 Domain Model and GORM

Domain models should not depend on GORM-specific behavior.

Avoid coupling business models to GORM features such as:

```go
gorm.Model
```

or GORM-specific relationship definitions when they are not required.

Database-specific concerns should remain within the Repository layer.

### 2.8 Database Naming Convention

Database naming must follow the project naming convention:

- Table names must use plural `snake_case`.
- Column names must use `snake_case`.

Examples:

```
products
transactions
transaction_details
order_items
```

Columns:

```
id
product_id
transaction_id
created_at
updated_at
total_amount
discount_percentage
```

### 2.9 Database Transactions

When multiple database operations must succeed or fail as a single unit, use a database transaction.

Transaction management must remain within the appropriate application/repository boundary and must not be implemented directly in the Handler.

Example:

```
UseCase
   ↓
Transaction Repository
   ↓
Database Transaction
   ├── INSERT transaction
   ├── INSERT transaction details
   └── UPDATE related data
```

If any required operation fails, the transaction must be rolled back.

### 2.10 Raw SQL

The default database implementation should use GORM's query builder.

Raw SQL may only be used when:

- The query cannot be reasonably expressed using the query builder.
- Explicit SQL is required for a specific database operation.
- There is a clear technical justification.

If raw SQL is necessary:

- Keep it inside the Repository layer.
- Use parameterized values.
- Do not expose it to the Handler or UseCase.
- Keep the implementation isolated.
- Document the reason when appropriate.

---

## 3. Implementation Rules

Before writing code, Claude must:

1. Read the relevant PRD under `docs/prd/`.
2. Read `docs/development-ruleset.md`.
3. Inspect the existing implementation related to the requested feature.
4. Identify the appropriate module/domain.
5. Follow the existing architecture and conventions.
6. Reuse existing shared components whenever applicable.
7. Avoid introducing unnecessary dependencies.
8. Do not introduce an ORM.
9. Use **GORM Query Builder** for database access.
10. Keep business logic in the UseCase layer.
11. Keep database access in the Repository layer.
12. Keep request validation in the Handler/Controller layer.
13. Use the shared Presenter for API responses.
14. Keep cross-cutting concerns such as authentication and rate limiting in Middleware.

Claude must not introduce a new architectural pattern when an existing project pattern already satisfies the requirement.

### 3.1 Minimal Changes

When modifying the project:

- Make the smallest reasonable change.
- Do not refactor unrelated code.
- Do not modify unrelated modules.
- Do not introduce unnecessary abstractions.
- Do not change existing API contracts without explicit requirements.
- Preserve existing behavior unless the PRD explicitly requires a change.

### 3.2 Before Completing a Task

Claude should verify:

```
PRD Requirement
      ↓
Implementation
      ↓
Development Ruleset
      ↓
Existing Project Conventions
      ↓
Tests / Validation
```

Before considering the task complete, verify that:

- The implementation satisfies the relevant PRD.
- The implementation follows `docs/development-ruleset.md`.
- Business logic is located in the UseCase layer.
- Database access is located in the Repository layer.
- GORM is used only as a query builder/database abstraction.
- No unnecessary ORM features have been introduced.
- API responses follow the shared Presenter.
- Existing tests continue to pass.
- New functionality has appropriate tests where applicable.

If a requirement cannot be implemented consistently with the existing architecture, explain the conflict before making a significant architectural change.

### 3.3 Code Style

- Do not add comments to code. Code must be self-explanatory through naming.
- Do not add logging (`log.Printf`, `log.Println`, etc.) in Handler, UseCase, or Repository code. Error paths must return the appropriate error/response without logging it.
  - Exception: fatal startup errors in `cmd/server/main.go` (e.g. failing to connect to the database or start the server) may use `log.Fatal`/`log.Fatalf`, since the process cannot continue and there is no request/response to return instead.