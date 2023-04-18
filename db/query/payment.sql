-- name: CreatePayment :one
INSERT INTO payments (
  id,
  booking_id,
  amount,
  pay_date,
  pay_method,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments
WHERE id = $1 LIMIT 1;

-- name: ListPayments :many
SELECT * FROM payments
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdatePayment :one
UPDATE payments
SET pay_date = $2, pay_method = $3, status = $4
WHERE id = $1
RETURNING *;

-- name: DeletePayment :exec
DELETE FROM payments WHERE id = $1;