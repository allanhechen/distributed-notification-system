# Server Architecture

## Status

Accepted

## Context

The application server must be organized in a way that is extendable,
maintainable, and most importantly, testable. In the case that a bug
appears, we should know where it took place because the logic is contained
solely in one place.

Our application only has simple actions, all of which should be atomic.
In other words, we will not need to combine multiple actions together.

## Decision

We will structure our application with the repository pattern. We will
implement this pattern with 4 different layers:

1. Database interface layer
   a. We will write plain SQL and generate corresponding code with sqlc
   b. We currently only support CockroachdDB, but this enables us to
   migrate to other databases as well
2. Repository layer
   a. We will manage our transactions in this layer
   b. Repositories will be grouped by aggregates
   c. Only this layer will be tested with integration tests
3. Service layer
   a. We will have a thin service layer, our repository manages atomicity
   b. There will be no transactions in the service layer
4. Handler layer
   a. Handlers will implement methods on structs containing one or more
   services

## Consequences

-   Clear separation of concerns, each layer has its own responsibilities
-   Easy unit testing
-   Reduce coupling with the database
-   Predictable flow of data
