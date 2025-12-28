# RabbitMQ Entities

## Status

Accepted

## Context

It is difficult to change the shape of RabbitMQ entities after they have
been created. To ensure minimal friction going forward, we must agree upon
their shape beforehand.

We must support the following features:

-   Dedicated queues for each notification type (ex. IOS, Android, Email)
-   Support for dedicated queues to receive all notifications for logging
-   Messages shall persist upon node restart
-   The message queue shall be highly available
-   Failed messages must be aggregated for further analysis

## Decision

We will send our notifications through a topic exchange. This allows for
individual queues to be bound to certain targets while also facilitating
a logging queue. We will use quorum queues to support high-availability,
while denoting persistence in the message, queue, and exchange levels. We
will add a DLQ to aggregate and handle failed messages, although this
action will be completed at a future date.

## Consequences

-   We must run multiple instances of RabbitMQ to achieve quorum queues
-   We accept the overhead associated with running quorum queues
-   We do not care about message ordering
