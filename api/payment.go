package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/token"
	"github.com/gin-gonic/gin"
)

type getPaymentByBookingIDRequest struct {
	Username  string `uri:"username" binding:"required,alphanum"`
	BookingID string `uri:"booking_id" binding:"required,alphanum"`
}

func (server *Server) userGetPaymentByBookingID(ctx *gin.Context) {
	var req getPaymentByBookingIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payment, err := server.store.GetPaymentByBookingID(ctx, req.BookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
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
	ctx.JSON(http.StatusOK, payment)
}

func (server *Server) adminGetPaymentByBookingID(ctx *gin.Context) {
	var req getPaymentByBookingIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payment, err := server.store.GetPaymentByBookingID(ctx, req.BookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, payment)
}

type listPaymentRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) adminListPaymentUnpaid(ctx *gin.Context) {
	var req listPaymentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListPaymentsUnpaidParams{
		Limit:  req.PageID,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	payments, err := server.store.ListPaymentsUnpaid(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, payments)
}

func (server *Server) userListPaymentUnpaid(ctx *gin.Context) {
	var req listPaymentRequest
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

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if user.Username != authPayload.Username {
		err := errors.New("user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.ListPaymentsUnpaidParams{
		Limit:  req.PageID,
		Offset: (req.PageID - 1) * req.PageSize,
		Status: "unpaid",
	}

	payments, err := server.store.ListPaymentsUnpaid(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, payments)
}

type totalIncomeMonthlyRequest struct {
	Month int `json:"month" binding:"required,min=1,max=12"`
	Year  int `json:"year" binding:"required,min=2023"`
}

func (server *Server) getTotalIncomeMonthly(ctx *gin.Context) {
	var req totalIncomeMonthlyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	startMonth := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.Local)
	endMonth := time.Date(req.Year, time.Month(req.Month)+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Hour * 24)

	arg := db.TotalIncomeParams{
		PayDate:   startMonth,
		PayDate_2: endMonth,
		Status:    "paid",
	}

	totalIncome, err := server.store.TotalIncome(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, totalIncome)
}

type totalIncomeYearlyRequest struct {
	Year int `json:"year" binding:"required,min=2023"`
}

func (server *Server) getTotalIncomeYearly(ctx *gin.Context) {
	var req totalIncomeYearlyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	startYear := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.Local)
	endYear := time.Date(req.Year, 12, 31, 0, 0, 0, 0, time.Local)

	arg := db.TotalIncomeParams{
		PayDate:   startYear,
		PayDate_2: endYear,
		Status:    "paid",
	}

	totalIncome, err := server.store.TotalIncome(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, totalIncome)
}
