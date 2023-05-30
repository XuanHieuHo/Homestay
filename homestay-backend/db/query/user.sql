-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  email,
  phone,
  role,
  "isBooking",
  password_changed_at,
  created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByResetPassToken :one
SELECT * FROM users
WHERE reset_password_token = $1 LIMIT 1;

-- name: UpdateResetPasswordToken :one
UPDATE users
SET reset_password_token = $2, rspassword_token_expired_at = $3
WHERE username = $1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET full_name = $2, email = $3, phone = $4
WHERE username = $1
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET "isBooking" = $2
WHERE username = $1
RETURNING *;

-- name: ChangeUserPassword :one
UPDATE users
SET hashed_password = $2, password_changed_at = $3
WHERE username = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE username = $1;