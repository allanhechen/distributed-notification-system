# Server Database Structure

## Status

Accepted

## Context

We need a way to access the database within the application server. This
method must be scalable, and low coupling is desired between the database
interface and the rest of the application.

## Decision

The Database component acts as a central Facade, grouping related database
operations into logical domains. These groups of requests hold a private
reference to a shared database pool, in addition to implementing multiple
database request handlers. These request handlers perform the duties of
querying the database, and can be extended in many ways.

The root database type is passed with dependency injection to the API
request handlers.

## Consequences

-   Developers need to write slightly more boilerplate code for each
    database handler
-   Low coupling is achieved with the database
-   Database requests are grouped and easily found
-   Dependency injection allows for easy unit tests
