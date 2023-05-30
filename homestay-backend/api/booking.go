package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/token"
	"github.com/gin-gonic/gin"
)

type createBookingRequest struct {
	PromotionID   string `json:"promotion_id" binding:"required"`
	CheckinDate   string `json:"checkin_date" binding:"required"`
	NumberOfDay   int32  `json:"number_of_day" binding:"required"`
	NumberOfGuest int32  `json:"number_of_guest" binding:"required,min=1"`
}

type createUserAndHomestay struct {
	UserBooking     string `uri:"username" binding:"required,alphanum"`
	HomestayBooking int64  `uri:"homestay_booking" binding:"required,min=1"`
}

// @Summary User Create Booking
// @ID createBooking
// @Produce json
// @Accept json
// @Param data body createBookingRequest true "createBookingRequest data"
// @Param username path string true "UserBooking"
// @Param homestay_booking path string true "HomestayBooking"
// @Security bearerAuth
// @Tags User
// @Success 200 {object} db.BookingTxResult
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/bookings/{homestay_booking} [post]
func (server *Server) createBooking(ctx *gin.Context) {
	var reqUserHomestay createUserAndHomestay
	var req createBookingRequest

	if err := ctx.ShouldBindUri(&reqUserHomestay); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqUserHomestay.UserBooking)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if user.Username != authPayload.Username {
		err := errors.New("user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.BookingTxParams{
		UserBooking:     reqUserHomestay.UserBooking,
		HomestayBooking: reqUserHomestay.HomestayBooking,
		PromotionID:     req.PromotionID,
		CheckinDate:     req.CheckinDate,
		NumberOfDay:     req.NumberOfDay,
		NumberOfGuest:   req.NumberOfGuest,
	}

	result, err := server.store.BookingTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, result)
}

type cancelBookingRequest struct {
	BookingID       string `uri:"booking_id" binding:"required,alphanum"`
	UserBooking     string `uri:"username" binding:"required,alphanum"`
	HomestayBooking int64  `uri:"homestay_booking" binding:"required,min=1"`
}

// @Summary User Cancel Booking
// @ID cancelBooking
// @Produce json
// @Accept json
// @Param username path string true "UserBooking"
// @Param homestay_booking path string true "HomestayBooking"
// @Param booking_id path string true "BookingID"
// @Security bearerAuth
// @Tags User
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/bookings/{homestay_booking}/{booking_id}/cancel [put]
func (server *Server) cancelBooking(ctx *gin.Context) {
	var req cancelBookingRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.UserBooking)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if user.Username != authPayload.Username {
		err := errors.New("user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CancelBookingParams{
		BookingID:       req.BookingID,
		UserBooking:     req.UserBooking,
		HomestayBooking: req.HomestayBooking,
	}

	rsp, err := server.store.CancelBookingTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	ctx.JSON(http.StatusOK, rsp)
}

type checkoutRequest struct {
	BookingID       string `uri:"booking_id" binding:"required,alphanum"`
	UserBooking     string `uri:"username" binding:"required,alphanum"`
	HomestayBooking int64  `uri:"homestay_booking" binding:"required,min=1"`
}

// @Summary Admin CheckOut Booking
// @ID checkoutBooking
// @Produce json
// @Accept json
// @Param username path string true "UserBooking"
// @Param homestay_booking path string true "HomestayBooking"
// @Param booking_id path string true "BookingID"
// @Security bearerAuth
// @Tags Admin
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 500 {string} error
// @Router /api/admin/users/{username}/bookings/{homestay_booking}/{booking_id}/checkout [put]
func (server *Server) checkoutBooking(ctx *gin.Context) {
	var req checkoutRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CheckoutParams{
		BookingID:       req.BookingID,
		UserBooking:     req.UserBooking,
		HomestayBooking: req.HomestayBooking,
	}

	rsp, err := server.store.CheckoutTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	ctx.JSON(http.StatusOK, rsp)
}

type listBookingRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

type listBookingResponse struct {
	ListBooking []struct {
		Booking         db.Booking       `json:"booking"`
		UserBooking     userResponse     `json:"user_booking"`
		HomestayBooking db.Homestay      `json:"homestay_booking"`
		Payment         db.Payment       `json:"payment"`
		DetailPayment   db.DetailPayment `json:"detail_payment"`
	} `json:"list_booking"`
}

// @Summary User Get List Booking
// @ID userGetListBooking
// @Produce json
// @Accept json
// @Tags User
// @Param username path string true "Username"
// @Param data query listBookingRequest true "listBookingRequest data"
// @Security bearerAuth
// @Success 200 {object} listBookingResponse
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/list_booking/ [get]
func (server *Server) userGetListBooking(ctx *gin.Context) {
	var result listBookingResponse
	var req listBookingRequest
	var reqUser getUserRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindUri(&reqUser); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqUser.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	userResult := newUserResponse(user)

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if user.Username != authPayload.Username {
		err := errors.New("user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.ListBookingByUserParams{
		UserBooking: user.Username,
		Limit:       req.PageSize,
		Offset:      (req.PageID - 1) * req.PageSize,
	}

	bookings, err := server.store.ListBookingByUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	for _, booking := range bookings {
		payment, err := server.store.GetPaymentByBookingID(ctx, booking.BookingID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		promotion, err := server.store.GetPromotion(ctx, booking.PromotionID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		homestay, err := server.store.GetHomestay(ctx, booking.HomestayBooking)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		duration := booking.CheckoutDate.Sub(booking.CheckinDate)
		numofday := int32(duration.Hours() / 24)
		homestayfee := float64(numofday) * homestay.Price

		var surchargeCapacity float64
		var amount float64
		if booking.NumberOfGuest > homestay.Capacity {
			surchargeCapacity = float64(booking.NumberOfGuest-homestay.Capacity) * 10
			amount = (surchargeCapacity + homestayfee + booking.ServiceFee)
		} else {
			surchargeCapacity = 0
			amount = homestayfee + booking.ServiceFee
		}

		tax := booking.Tax * amount

		detail := db.DetailPayment{
			UserBooking:       booking.UserBooking,
			HomestayBooking:   booking.HomestayBooking,
			CheckinDate:       booking.CheckinDate.String(),
			Discount:          promotion.DiscountPercent,
			NumberOfDay:       numofday,
			NumberOfGuest:     booking.NumberOfGuest,
			Tax:               tax,
			ServiceFee:        booking.ServiceFee,
			SurchargeCapacity: surchargeCapacity,
			HomestayFee:       homestayfee,
			TotalAmount:       payment.Amount,
		}

		result.ListBooking = append(result.ListBooking, struct {
			Booking         db.Booking       `json:"booking"`
			UserBooking     userResponse     `json:"user_booking"`
			HomestayBooking db.Homestay      `json:"homestay_booking"`
			Payment         db.Payment       `json:"payment"`
			DetailPayment   db.DetailPayment `json:"detail_payment"`
		}{booking, userResult, homestay, payment, detail})
	}
	ctx.JSON(http.StatusOK, result)
}
