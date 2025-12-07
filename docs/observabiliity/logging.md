# Logging

## The goal behind logging

We want events to be traceable from outside of the application. This will
enable the following:

-   Following a request through microservices
-   Easier debugging
-   Analytics tools for performance monitoring
-   Early warning systems

## What is an event

We define events to be any significant step in the processing of a
request. The events we are currently concerned with include:

-   The beginning of a request
-   The end of a request
-   Database queries
-   Handoff between the application and the message queue
-   Connection statuses between services (ex. DB and MQ)
-   Server status updates

## Log Levels

1. Info

    - "Business as usual"
    - Significant and noteworthy business events
    - Login, transactions, order number

2. Warn

    - Abnormal situations that may indicate future problems
    - "Payment processing is taking longer than usual"

3. Error

    - Unrecoverable errors that affect a specific operation
    - "Database connection failed"

4. Fatal
    - Unrecoverable issues that affect the entire program
    - "System out of memory"

## Log Output

In the future, we will collect and ship logs to a central log handling
service. To support this, logs will include information to determine their
origin, along with the commit hash. We will use structured logs with
consistent identification fields. These fields include:

-   `requestId` (from the JWT, if available)
-   `userId` (from the JWT, if available)
-   Instance identification
-   Commit hash

The service administrator can configure log output into two outputs by
setting an environment variable:

1. `development` mode:

    - Outputs to stdout
    - Has color formatting
    - Does not include the identification fields
    - Logs are not persistent
    - Output stored in key=value pairs

2. `production` mode:

    - Outputs to "./logs" from the working directory of the service
    - Logs has the identification fields
    - Logs are persistent
    - Logs stored in JSON format
