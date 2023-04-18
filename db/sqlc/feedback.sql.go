// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: feedback.sql

package db

import (
	"context"
	"time"
)

const createFeedback = `-- name: CreateFeedback :one
INSERT INTO feedbacks (
  user_comment,
  homestay_commented,
  rating,
  commention,
  created_at
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING id, user_comment, homestay_commented, rating, commention, created_at
`

type CreateFeedbackParams struct {
	UserComment       string    `json:"user_comment"`
	HomestayCommented int64     `json:"homestay_commented"`
	Rating            string    `json:"rating"`
	Commention        string    `json:"commention"`
	CreatedAt         time.Time `json:"created_at"`
}

func (q *Queries) CreateFeedback(ctx context.Context, arg CreateFeedbackParams) (Feedback, error) {
	row := q.db.QueryRowContext(ctx, createFeedback,
		arg.UserComment,
		arg.HomestayCommented,
		arg.Rating,
		arg.Commention,
		arg.CreatedAt,
	)
	var i Feedback
	err := row.Scan(
		&i.ID,
		&i.UserComment,
		&i.HomestayCommented,
		&i.Rating,
		&i.Commention,
		&i.CreatedAt,
	)
	return i, err
}

const deleteFeedback = `-- name: DeleteFeedback :exec
DELETE FROM feedbacks WHERE id = $1
`

func (q *Queries) DeleteFeedback(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteFeedback, id)
	return err
}

const getFeedback = `-- name: GetFeedback :one
SELECT id, user_comment, homestay_commented, rating, commention, created_at FROM feedbacks
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetFeedback(ctx context.Context, id int64) (Feedback, error) {
	row := q.db.QueryRowContext(ctx, getFeedback, id)
	var i Feedback
	err := row.Scan(
		&i.ID,
		&i.UserComment,
		&i.HomestayCommented,
		&i.Rating,
		&i.Commention,
		&i.CreatedAt,
	)
	return i, err
}

const listFeedbacks = `-- name: ListFeedbacks :many
SELECT id, user_comment, homestay_commented, rating, commention, created_at FROM feedbacks
WHERE homestay_commented = $1
ORDER BY created_at
LIMIT $2
OFFSET $3
`

type ListFeedbacksParams struct {
	HomestayCommented int64 `json:"homestay_commented"`
	Limit             int32 `json:"limit"`
	Offset            int32 `json:"offset"`
}

func (q *Queries) ListFeedbacks(ctx context.Context, arg ListFeedbacksParams) ([]Feedback, error) {
	rows, err := q.db.QueryContext(ctx, listFeedbacks, arg.HomestayCommented, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Feedback{}
	for rows.Next() {
		var i Feedback
		if err := rows.Scan(
			&i.ID,
			&i.UserComment,
			&i.HomestayCommented,
			&i.Rating,
			&i.Commention,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateFeedback = `-- name: UpdateFeedback :one
UPDATE feedbacks
SET rating = $2, commention = $3
WHERE id = $1
RETURNING id, user_comment, homestay_commented, rating, commention, created_at
`

type UpdateFeedbackParams struct {
	ID         int64  `json:"id"`
	Rating     string `json:"rating"`
	Commention string `json:"commention"`
}

func (q *Queries) UpdateFeedback(ctx context.Context, arg UpdateFeedbackParams) (Feedback, error) {
	row := q.db.QueryRowContext(ctx, updateFeedback, arg.ID, arg.Rating, arg.Commention)
	var i Feedback
	err := row.Scan(
		&i.ID,
		&i.UserComment,
		&i.HomestayCommented,
		&i.Rating,
		&i.Commention,
		&i.CreatedAt,
	)
	return i, err
}