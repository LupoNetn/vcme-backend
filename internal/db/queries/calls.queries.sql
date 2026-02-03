-- name: CreateCallLink :one
INSERT INTO calls (title,description,call_link,host_id)
VALUES ($1,$2,$3,$4) RETURNING *;

-- name: ListAllCalls :many
SELECT * FROM calls;

-- name: ListAllCallsByID :many
SELECT * FROM calls WHERE id = $1;

-- name: UpdateCall :one
UPDATE calls
SET
title = COALESCE(sql.narg('title'), title),
description = COALESCE(sql.narg('description'), title)
WHERE id = $1
RETURNING *;



