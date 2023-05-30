-- name: CreateHomestay :one
INSERT INTO homestays (
  description,
  address,
  number_of_bed,
  capacity,
  price,
  status,
  main_image,
  first_image,
  second_image,
  third_image
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetHomestay :one
SELECT * FROM homestays
WHERE id = $1 LIMIT 1;

-- name: ListHomestays :many
SELECT * FROM homestays
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateHomestayStatus :one
UPDATE homestays
SET status = $2
WHERE id = $1
RETURNING *;

-- name: UpdateHomestayInfo :one
UPDATE homestays
SET description = $2, address = $3, number_of_bed = $4, capacity = $5, price = $6, main_image = $7, first_image = $8, second_image = $9, third_image = $10
WHERE id = $1
RETURNING *;

-- name: DeleteHomestay :exec
DELETE FROM homestays WHERE id = $1;