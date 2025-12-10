# Idempotency

As mentioned in
[the idempotency ADR](../decisions/0005-server-request-idempotency.md), we
must handle requests in an idempotent way to address the limitations
within our system model. This implementation requires the client to
generate unique UUIDs to be used with every new request.

requestIds must be unique even between users.

## The Happy Path

The happy path for our implementation only considers the scenario where
everything goes right with our systems, no requests are dropped, and no
nodes crash.

1. The client sends a request to an endpoint with a fresh X-REQUEST-ID
2. The request passes through the API gateway, reaching one of the
   application servers
3. The request is logged in the database table `request_idempotency_keys`
   with an initial lock of 120 seconds
   a. We DO NOT store a hash of the initial request, bad client
   implementations might receive undefined behavior and that is
   acceptable. To help counteract client errors, we send back a
   `X-Cache-Status: Idempotency-Hit` header to signifiy
   b. This log records an "in progress" status for the request
   c. Any other requests with this request_id received during this time
   receives a response with a 425 Too Early status
4. The request is wrapped within a context with a 100-second expiry window
   a. This allows the server to have some flexibility for poor network
   communication
   b. If the commit message reaches the database late, the relevant row
   may have already been garbage collected (this is fine, the client
   simply retries the request)
   c. There are no other consequences of a request arriving late to the
   database, we have serializable transactions
5. The handler handles the request, writing any side effects to a
   transaction outbox table
   a. The `request_idempotency_keys` table is also updated in the same
   transaction
   b. Successful requests (and persistent user error requests) are cached
   in the idempotency table for a period of 24 hours
6. The handler sends a response back to the client

## Various Edge Cases

If at any point (other than #1) the client crashes, the server continues
processing the result without this knowledge. If the server succeeds, any
side effects are persisted as if the client did not crash.

### Possible server crashes

-   Crash at #2
    a. The server behaves as if it never received the request

-   Crash at #3, #4, before transaction commit in #5
    a. The lock expires, the row may be overwritten by subsequent retries
    b. Any requests received during this time receive a request with the
    425 Too Early status

-   Context timeout at #4
    a. Same as above

-   Crash after transaction commit in #5, #6
    a. Request considered succeeded, return cached result to client
