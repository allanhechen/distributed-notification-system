# Database Code Generation

## Status

Accepted

## Context

The application server depends heavily on communication with the database.
Writing SQL code by hand provides clarity and control, but manually
updating Go data structures and query methods is error-prone. Mismatches
in these data types can easily lead to runtime errors.

We want a solution that keeps the SQL a source of truth. This solution
must generate type-safe Go code, reduce boilerplate, and be easily
maintainable.

## Decision

ORMs offer higher-level abstractions, but they hide SQL behind their own
implementations. Our team prefers explicit SQL and minimal runtime
overhead.

We will use `sqlc` as our database code generation tool. sqlc reads our
database migrations and queries, and generates types, queries, parameters,
and result mappings.

Our project structure will follow this format (within the application
server):

-   `internal/db/migrations/` for the database migrations
-   `internal/db/` for the sqlc output
-   `internal/db/queries/` for the input SQL queries

## Consequences

-   Type-safe query code eliminates a class of runtime errors
-   SQL becomes the authoritative definition of database behavior
-   Developers must write and maintain .sql files for all queries
-   Generated code must be checked into version control
