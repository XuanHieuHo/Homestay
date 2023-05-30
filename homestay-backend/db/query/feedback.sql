-- name: CreateFeedback :one
INSERT INTO feedbacks (
  user_comment,
  homestay_commented,
  rating,
  commention,
  created_at
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetFeedback :one
SELECT * FROM feedbacks
WHERE id = $1 LIMIT 1;

-- name: ListFeedbacks :many
SELECT * FROM feedbacks
WHERE homestay_commented = $1
ORDER BY created_at
LIMIT $2
OFFSET $3;

-- name: UpdateFeedback :one
UPDATE feedbacks
SET rating = $2, commention = $3
WHERE id = $1
RETURNING *;

-- name: DeleteFeedback :exec
DELETE FROM feedbacks WHERE id = $1;