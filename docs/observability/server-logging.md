# Server Logging

## Goals

All servers independently report the statuses of all external services. A
server sends a log message during startup and shutdown, and additional log
messages for each request processed.

## Implementation

Server startup and shutdown logging messages are located within main.go
and are fairly trivial. The request logging is implemented using
middleware in api.go. This middleware will also put a logger into the
request's context, which will output all required information when writing
logs.

Any additional information relevant to external services accesses will be
the responsibility of the request handlers. This information will also be
placed into a canonical log line by the middleware at the end of the
request.
