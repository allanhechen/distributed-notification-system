# System Model and Timing Assumptions

## Status

Accepted

## Context

This ADR focuses on formalizing the system model and timing assumptions
that this application generally assumes. This application will only
consider situations deemed impossible by the system model once the MVP has
been achieved.

## Decision

This application generally assumes a partially synchronous model, meaning
that there is generally the following:

-   Reasonable network delay
-   Reasonable process pauses
-   Reasonable clock error
-   The network may drop, delay, or reorder messages, but not corrupt them

However, the system also accepts that any of the above can be violated at
any time.

Additionally, nodes are assumed to behave with crash-recovery faults,
meaning the following are possible:

-   Nodes can crash at any moment
-   Crashed nodes may recover after some unknown time
-   Secondary storage is kept during crashes
-   Volatile memory is lost during crashes

## Consequences

-   The system guarantees correctness only under the failure and timing
    assumptions described above
-   Any behavior outside this model is considered out of scope for the MVP
-   Future revisions may expand the failure model or adopt stronger fault
    -tolerance mechanisms
-   The application must tolerate any behaviors permitted by the partially
    synchronous and crash-recovery models
