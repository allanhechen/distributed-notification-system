-- name: GetRequestStatus :one
SELECT * FROM idempotent_requests
WHERE request_id = $1;