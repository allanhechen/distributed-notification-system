-- name: GetRequestStatus :one
SELECT * FROM idempotent_requests
WHERE request_id = $1;

-- name: CreateRequestStatus :exec
INSERT INTO idempotent_requests(
    request_id,
    user_id,
    request_status_id,
    expires_at
) VALUES ($1, $2, $3, $4);
