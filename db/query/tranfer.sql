-- name: CreateTranfer :one
INSERT INTO tranfers (
  from_account_id, to_account_id, ammount
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetTranfer :one
SELECT * FROM tranfers
WHERE id = $1 LIMIT 1;

-- name: ListTranfers :many
SELECT * FROM tranfers
ORDER BY id
LIMIT $1
OFFSET $2
;

-- name: UpdateTranfer :one
UPDATE tranfers 
SET ammount = $2
WHERE id = $1
RETURNING *;

-- name: DeleteTranfer :exec
DELETE FROM tranfers WHERE id = $1;
