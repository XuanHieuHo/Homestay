-- name: CreatePayment :one
INSERT INTO payments (
  booking_id,
  amount,
  pay_date,
  pay_method,
  status
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments
WHERE id = $1 LIMIT 1;

-- name: GetPaymentByBookingID :one
SELECT * FROM payments
WHERE booking_id = $1 LIMIT 1;

-- name: ListPayments :many
SELECT * FROM payments
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListPaymentsUnpaid :many
SELECT * FROM payments
WHERE status = $3
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: TotalIncome :one
SELECT CAST(SUM(amount) AS FLOAT) AS TotalIncome FROM payments
WHERE (
pay_date BETWEEN $1 AND $2
AND status = $3);

-- name: UpdatePayment :one
UPDATE payments
SET pay_date = $2, pay_method = $3, status = $4
WHERE id = $1
RETURNING *;

-- name: DeletePayment :exec
DELETE FROM payments WHERE id = $1;