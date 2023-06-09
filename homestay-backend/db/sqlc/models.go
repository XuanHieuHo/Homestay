// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	BookingID       string    `json:"booking_id"`
	UserBooking     string    `json:"user_booking"`
	HomestayBooking int64     `json:"homestay_booking"`
	PromotionID     string    `json:"promotion_id"`
	Status          string    `json:"status"`
	BookingDate     time.Time `json:"booking_date"`
	CheckinDate     time.Time `json:"checkin_date"`
	CheckoutDate    time.Time `json:"checkout_date"`
	NumberOfGuest   int32     `json:"number_of_guest"`
	// must be positive
	ServiceFee float64 `json:"service_fee"`
	// must be positive
	Tax float64 `json:"tax"`
}

type Feedback struct {
	ID                int64     `json:"id"`
	UserComment       string    `json:"user_comment"`
	HomestayCommented int64     `json:"homestay_commented"`
	Rating            string    `json:"rating"`
	Commention        string    `json:"commention"`
	CreatedAt         time.Time `json:"created_at"`
}

type Homestay struct {
	ID          int64   `json:"id"`
	Description string  `json:"description"`
	Address     string  `json:"address"`
	NumberOfBed int32   `json:"number_of_bed"`
	Capacity    int32   `json:"capacity"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
	MainImage   string  `json:"main_image"`
	FirstImage  string  `json:"first_image"`
	SecondImage string  `json:"second_image"`
	ThirdImage  string  `json:"third_image"`
}

type Payment struct {
	ID        int64     `json:"id"`
	BookingID string    `json:"booking_id"`
	Amount    float64   `json:"amount"`
	PayDate   time.Time `json:"pay_date"`
	PayMethod string    `json:"pay_method"`
	Status    string    `json:"status"`
}

type Promotion struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DiscountPercent float64   `json:"discount_percent"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	Username                 string    `json:"username"`
	HashedPassword           string    `json:"hashed_password"`
	FullName                 string    `json:"full_name"`
	Email                    string    `json:"email"`
	Phone                    string    `json:"phone"`
	Role                     string    `json:"role"`
	IsBooking                bool      `json:"isBooking"`
	PasswordChangedAt        time.Time `json:"password_changed_at"`
	CreatedAt                time.Time `json:"created_at"`
	ResetPasswordToken       string    `json:"reset_password_token"`
	RspasswordTokenExpiredAt time.Time `json:"rspassword_token_expired_at"`
}
