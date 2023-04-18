-- name: CreateBooking :one
INSERT INTO bookings (
  id,
  user_booking,
  homestay_booking,
  promotion_id,
  payment_id,
  status,
  booking_date,
  checkin_date,
  checkout_date,
  number_of_guest,
  service_fee,
  tax
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings
WHERE id = $1 LIMIT 1;

-- name: ListBookingByUser :many
SELECT * FROM bookings
WHERE user_booking = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: ListBookingByHomestay :many
SELECT * FROM bookings
WHERE homestay_booking = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpdateBooking :one
UPDATE bookings
SET status = $2, checkout_date = $3
WHERE id = $1
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings WHERE id = $1;