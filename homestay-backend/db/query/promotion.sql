-- name: CreatePromotion :one
INSERT INTO promotions (
  title,
  description,
  discount_percent,
  start_date,
  end_date
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPromotion :one
SELECT * FROM promotions
WHERE title = $1 LIMIT 1;

-- name: ListPromotions :many
SELECT * FROM promotions
ORDER BY title
LIMIT $1
OFFSET $2;

-- name: UpdatePromotion :one
UPDATE promotions
SET description = $2, discount_percent = $3, end_date = $4
WHERE id = $1
RETURNING *;

-- name: DeletePromotion :exec
DELETE FROM promotions WHERE id = $1;