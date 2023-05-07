package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XuanHieuHo/homestay/util"
	"github.com/lib/pq"
)

type Store interface {
	Querier
	BookingTx(ctx context.Context, arg BookingTxParams) (BookingTxResult, error)
	CancelBookingTx(ctx context.Context, arg CancelBookingParams) (string, error)
	CheckoutTx(ctx context.Context, arg CheckoutParams) (string, error)
}

type SQLStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type BookingTxParams struct {
	UserBooking     string    `json:"user_booking"`
	HomestayBooking int64     `json:"homestay_booking"`
	PromotionID     string    `json:"promotion_id"`
	CheckinDate     time.Time `json:"checkin_date"`
	NumberOfDay     int32     `json:"number_of_day"`
	NumberOfGuest   int32     `json:"number_of_guest"`
}

type BookingTxResult struct {
	Booking         Booking  `json:"booking"`
	UserBooking     User     `json:"user_booking"`
	HomestayBooking Homestay `json:"homestay_booking"`
	Payment         Payment  `json:"payment"`
}

// Transaction booking homestay
// to - do : kiểm tra khoảng thời gian đó, đã có ai đặt phòng hay chưa !!!!!!!
func (store *SQLStore) BookingTx(ctx context.Context, arg BookingTxParams) (BookingTxResult, error) {
	var result BookingTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// if homestay has still available => create Booking
		result.HomestayBooking, err = q.GetHomestay(ctx, arg.HomestayBooking)
		if err != nil {
			return err
		}

		if result.HomestayBooking.Status != "available" {
			err = fmt.Errorf("homestay can't be booked")
			return err
		}

		user, err := q.GetUser(ctx, arg.UserBooking)
		if err != nil {
			return err
		}
		result.UserBooking = User{
			Username:  user.Username,
			FullName:  user.FullName,
			Email:     user.Email,
			Phone:     user.Phone,
			IsBooking: user.IsBooking,
		}

		if result.UserBooking.IsBooking {
			err = fmt.Errorf("user has already booked")
			return err
		}

		var discount float64
		if arg.PromotionID != "" {
			promotion, err := q.GetPromotion(ctx, arg.PromotionID)
			if err != nil {
				if err == sql.ErrNoRows {
					err = fmt.Errorf("promotion code doesn't exist")
					return err
				} else {
					return err
				}
			}

			if time.Now().After(promotion.EndDate) {
				err = fmt.Errorf("promotion code has expired")
				return err
			}
			discount = promotion.DiscountPercent / 100
		} else {
			discount = 0
		}

		bookingID := util.RandomBookingCode()

		bookings, err := q.GetBookingByHomestayAndTime(ctx, GetBookingByHomestayAndTimeParams{
			HomestayBooking: arg.HomestayBooking,
			CheckinDate:     arg.CheckinDate,
			CheckoutDate:    arg.CheckinDate.AddDate(0, 0, int(arg.NumberOfDay)),
		})

		if err != nil {
			return err
		}

		if len(bookings) > 0 {
			return fmt.Errorf("this homestay has been booked in this time 1") 
		}

		result.Booking, err = q.CreateBooking(ctx, CreateBookingParams{
			BookingID:       bookingID,
			UserBooking:     arg.UserBooking,
			HomestayBooking: arg.HomestayBooking,
			PromotionID:     arg.PromotionID,
			Status:          "validated",
			BookingDate:     time.Now(),
			CheckinDate:     arg.CheckinDate,
			CheckoutDate:    arg.CheckinDate.AddDate(0, 0, int(arg.NumberOfDay)),
			NumberOfGuest:   arg.NumberOfGuest,
			ServiceFee:      float64(15 * arg.NumberOfGuest * arg.NumberOfDay),
			Tax:             0.1,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code.Name() {
				case "foreign_key_violation", "unique_violation":
					err = fmt.Errorf("this homestay has been booked in this time 2")
					return err
				}
			}
			return err
		}

		var amount float64
		if result.Booking.NumberOfGuest > result.HomestayBooking.Capacity {
			amount = (1 + result.Booking.Tax) * (float64(result.Booking.NumberOfGuest-result.HomestayBooking.Capacity)*10 + (result.HomestayBooking.Price * float64(arg.NumberOfDay)) + result.Booking.ServiceFee)
		} else {
			amount = (1 + result.Booking.Tax) * ((result.HomestayBooking.Price * float64(arg.NumberOfDay)) + result.Booking.ServiceFee)
		}

		totalAmount := (1 - discount) * amount

		result.Payment, err = q.CreatePayment(ctx, CreatePaymentParams{
			BookingID: result.Booking.BookingID,
			Amount:    totalAmount,
			Status:    "unpaid",
		})
		if err != nil {
			return err
		}

		_, err = q.UpdateUserStatus(ctx, UpdateUserStatusParams{
			Username:  result.UserBooking.Username,
			IsBooking: true,
		})
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

type CancelBookingParams struct {
	BookingID       string `json:"booking_id"`
	UserBooking     string `json:"user_booking"`
	HomestayBooking int64  `json:"homestay_booking"`
}

// Transaction cancel booking
func (store *SQLStore) CancelBookingTx(ctx context.Context, arg CancelBookingParams) (string, error) {

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		homestayBooking, err := q.GetHomestay(ctx, arg.HomestayBooking)
		if err != nil {
			err = fmt.Errorf("1")
			return err
		}
		if homestayBooking.Status != "available" {
			err = fmt.Errorf("homestay can't be booked")
			return err
		}

		userBooking, err := q.GetUser(ctx, arg.UserBooking)
		if err != nil {
			err = fmt.Errorf("2")
			return err
		}
		if !userBooking.IsBooking {
			err = fmt.Errorf("user hasn't booked any homestay")
			return err
		}

		booking, err := q.GetBooking(ctx, arg.BookingID)
		if err != nil {
			err = fmt.Errorf("3")
			return err
		}
		if booking.Status != "validated" {
			err = fmt.Errorf("booking isn't validated")
			return err
		}

		payment, err := q.GetPaymentByBookingID(ctx, arg.BookingID)
		if err != nil {
			err = fmt.Errorf("4")
			return err
		}
		if payment.Status != "unpaid" {
			err = fmt.Errorf("payment is illegal")
			return err
		}

		_, err = q.UpdateBooking(ctx, UpdateBookingParams{
			BookingID:    arg.BookingID,
			Status:       "cancel",
			CheckoutDate: time.Date(1, 1, 1, 0, 0, 0, 0, time.Local),
			CheckinDate:  time.Date(1, 1, 1, 0, 0, 0, 0, time.Local),
		})
		if err != nil {
			return err
		}

		_, err = q.UpdateUserStatus(ctx, UpdateUserStatusParams{
			Username:  arg.UserBooking,
			IsBooking: false,
		})
		if err != nil {
			err = fmt.Errorf("6")
			return err
		}

		_, err = q.UpdatePayment(ctx, UpdatePaymentParams{
			ID:     payment.ID,
			Status: "invalidated",
		})
		if err != nil {
			err = fmt.Errorf("7")
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return "Cancelling booking successfully", err
}

type CheckoutParams struct {
	BookingID       string `json:"booking_id"`
	UserBooking     string `json:"user_booking"`
	HomestayBooking int64  `json:"homestay_booking"`
}

// transaction checkout and pay
func (store *SQLStore) CheckoutTx(ctx context.Context, arg CheckoutParams) (string, error) {

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		homestayBooking, err := q.GetHomestay(ctx, arg.HomestayBooking)
		if err != nil {
			return err
		}
		if homestayBooking.Status != "available" {
			err = fmt.Errorf("homestay can't booked")
			return err
		}

		userBooking, err := q.GetUser(ctx, arg.UserBooking)
		if err != nil {
			return err
		}
		if !userBooking.IsBooking {
			err = fmt.Errorf("user hasn't booked any homestay")
			return err
		}

		booking, err := q.GetBooking(ctx, arg.BookingID)
		if err != nil {
			return err
		}
		if booking.Status != "validated" {
			err = fmt.Errorf("booking isn't validated")
			return err
		}

		payment, err := q.GetPaymentByBookingID(ctx, arg.BookingID)
		if err != nil {
			return err
		}
		if payment.Status != "unpaid" {
			err = fmt.Errorf("payment is illegal")
			return err
		}

		_, err = q.FinishBooking(ctx, FinishBookingParams{
			BookingID: arg.BookingID,
			Status:    "completed",
		})
		if err != nil {
			return err
		}

		_, err = q.UpdateUserStatus(ctx, UpdateUserStatusParams{
			Username:  arg.UserBooking,
			IsBooking: false,
		})
		if err != nil {
			return err
		}

		_, err = q.UpdatePayment(ctx, UpdatePaymentParams{
			ID:        payment.ID,
			PayDate:   time.Now(),
			Status:    "paid",
			PayMethod: "cash",
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return "Checkout and pay the bill successfully", err
}
