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
// @Router /api/users/:username/bookings/:homestay_booking [post]
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
// @Router /api/users/:username/bookings/:homestay_booking/:booking_id/cancel [put]
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
// @Router /api/users/:username/bookings/:homestay_booking/:booking_id/checkout [put]
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
