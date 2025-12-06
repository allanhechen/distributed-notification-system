# Choosing a Message Queue

## Status

Accepted

## Context

We need to choose a message queue to handle messages going from the server
to the workers. This message queue needs to support the following
features:

-   High request throughput
-   Track request statuses (the server will send and forget)
-   Handle an arbitrary amount of workers
-   Ordering does not matter
-   Handle different queues, one for each type of notification
-   Support fan out to multiple queues depending on the type of message

## Decision

We will implement RabbitMQ, as it is the most popular AMQP service, and
also supports our goals of running in Docker.

## Consequences

-   Messages must be logged by the server and the workers
    -   The message queue itself is ephemeral
-   Developers must run an instance of RabbitMQ locally
-   Workers may receive duplicate requests
    -   Need to handle deduplication in the workers
-   Must configure retry/DLQ with RabbitMQ
-   Writers must manually ACK messages after completion

-   Communication protocol between workers and RabbitMQ yet to be decided
