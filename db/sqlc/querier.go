// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"context"
)

type Querier interface {
	CreateBooking(ctx context.Context, arg CreateBookingParams) (Booking, error)
	CreateFeedback(ctx context.Context, arg CreateFeedbackParams) (Feedback, error)
	CreateHomestay(ctx context.Context, arg CreateHomestayParams) (Homestay, error)
	CreatePayment(ctx context.Context, arg CreatePaymentParams) (Payment, error)
	CreatePromotion(ctx context.Context, arg CreatePromotionParams) (Promotion, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeleteBooking(ctx context.Context, id int64) error
	DeleteFeedback(ctx context.Context, id int64) error
	DeleteHomestay(ctx context.Context, id int64) error
	DeletePayment(ctx context.Context, id int64) error
	DeletePromotion(ctx context.Context, id int64) error
	DeleteUser(ctx context.Context, username string) error
	GetBooking(ctx context.Context, id int64) (Booking, error)
	GetFeedback(ctx context.Context, id int64) (Feedback, error)
	GetHomestay(ctx context.Context, id int64) (Homestay, error)
	GetPayment(ctx context.Context, id int64) (Payment, error)
	GetPromotion(ctx context.Context, title string) (Promotion, error)
	GetUser(ctx context.Context, username string) (User, error)
	ListBookingByHomestay(ctx context.Context, arg ListBookingByHomestayParams) ([]Booking, error)
	ListBookingByUser(ctx context.Context, arg ListBookingByUserParams) ([]Booking, error)
	ListFeedbacks(ctx context.Context, arg ListFeedbacksParams) ([]Feedback, error)
	ListHomestays(ctx context.Context, arg ListHomestaysParams) ([]Homestay, error)
	ListPayments(ctx context.Context, arg ListPaymentsParams) ([]Payment, error)
	ListPromotions(ctx context.Context, arg ListPromotionsParams) ([]Promotion, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
	UpdateBooking(ctx context.Context, arg UpdateBookingParams) (Booking, error)
	UpdateFeedback(ctx context.Context, arg UpdateFeedbackParams) (Feedback, error)
	UpdateHomestayInfo(ctx context.Context, arg UpdateHomestayInfoParams) (Homestay, error)
	UpdateHomestayStatus(ctx context.Context, arg UpdateHomestayStatusParams) (Homestay, error)
	UpdatePayment(ctx context.Context, arg UpdatePaymentParams) (Payment, error)
	UpdatePromotion(ctx context.Context, arg UpdatePromotionParams) (Promotion, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
}

var _ Querier = (*Queries)(nil)
