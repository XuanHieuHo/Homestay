-- name: CreateBooking :one
INSERT INTO bookings (
  booking_id,
  user_booking,
  homestay_booking,
  promotion_id,
  status,
  booking_date,
  checkin_date,
  checkout_date,
  number_of_guest,
  service_fee,
  tax
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings
WHERE booking_id = $1 LIMIT 1;

-- name: GetBookingByHomestayAndTime :many
SELECT * FROM bookings
WHERE homestay_booking = $1 
AND (
  (checkin_date >= $2 AND checkout_date <= $3)
  OR (checkin_date <= $2 AND checkout_date >= $2)
  OR (checkin_date >= $2 AND checkin_date <= $3 AND checkout_date >= $3)
  OR (checkin_date = $3)
  OR (checkout_date = $2)
);


-- name: ListBookingByUser :many
SELECT * FROM bookings
WHERE user_booking = $1
ORDER BY booking_id
LIMIT $2
OFFSET $3;

-- name: ListBookingByHomestay :many
SELECT * FROM bookings
WHERE homestay_booking = $1
ORDER BY booking_id
LIMIT $2
OFFSET $3;

-- name: UpdateBooking :one
UPDATE bookings
SET status = $2, checkout_date = $3, checkin_date = $4
WHERE booking_id = $1
RETURNING *;

-- name: FinishBooking :one
UPDATE bookings
SET status = $2
WHERE booking_id = $1
RETURNING *;

-- name: DeleteBooking :exec
DELETE FROM bookings WHERE booking_id = $1;