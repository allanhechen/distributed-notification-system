# Choosing a Database

## Status

Accepted

## Context

We need to choose a database to store the main application data. This
database must meet the following requirements:

-   Partitioned, as the tables will be large
-   Support distributed transactions, because we need linearizability
-   Minimal manual repartitioning
-   Does not need to support complex access patterns, our queries are simple
-   Be hostable locally
-   Database must handle more reads than writes

## Decision

We will implement CockroachDB for our database layer, since it meets
all the requirements above while also implementing a familiar interfaces.

## Consequences

-   Will not have support for all PostgreSQL features

    -   Increment primary key must be replaced with UUIDs

-   Backup strategy yet to be decided
