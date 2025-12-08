# Server Request Idempotency

## Status

Accepted

## Context

The system model allows for unreliable communication and nodes that may
crash at any time. This system must account for these behaviors to achieve
the MVP.

This ADR only concerns itself with the behaviors of the server, the
consumers will be addressed in a separate ADR.

## Decision

We will implement end-to-end idempotency checks consistent throughout the
application servers. This will be accomplished by generating an
idempotency key on the client, sent with the `X-Idempotency-Key` header
on every request.

The server reasonably assumes that a client will generate a new
idempotency key for every request. The server will store all idempotency
keys for a period of 24 hours in a durable, crash-safe storage layer
persistent across server restarts, but clients utilizing the same
idempotency key for multiple requests accept undefined behavior.

On initial request, the server checks for the existence of the associated
idempotency key. There are two situations that may arise if the key
exists:

1. The original request has already completed

-   The server sends back a cached version of the initial response
-   No additional data is changed by the server

2. The original request is currently in progress

-   The server responds with a `409 conflict` status
-   No additional data is changed by the server

If the idempotency key does not exist, the server stores the idempotency
key and begins processing of the request. This information is initially
stored with a short TTL (~5 minutes) with an in progress status. Two
situations may follow:

1. The server fulfills the request

-   The TTL for the idempotency key is updated to 24 hours
-   The status of the request is updated to "completed"

2. The server cannot fulfill the request

-   The server responds with a 5XX status
-   All side effects produced by the request do not persist

The server might be unable to fulfill the request for any reason,
including the following:

-   The server crashes
-   The server cannot communicate with the database
-   The server runs out of memory
-   The server does not fulfill the request in time
-   The server is preempted for a long period of time

If a server crashes while a request is in progress and does not mark the
idempotency key as completed, the in-progress record will expire and the
client may safely retry. The retry will cause the server to reprocess the
request from the beginning. All requests with the same idempotency key
are guaranteed to be idempotent within 24 hours.

## Consequences

-   Clients must ensure key uniqueness per operation
-   Reliability improves under retries
-   Server must maintain durable idempotency state
-   Server resource usage increases due to 24-hour retention
-   System enforces linearizable semantics per idempotency key
