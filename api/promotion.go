package api

import (
	"database/sql"
	"net/http"
	"time"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createPromotionRequest struct {
	Title           string  `json:"title" binding:"required"`
	Description     string  `json:"description" binding:"required"`
	DiscountPercent float64 `json:"discount_percent" binding:"required"`
	EndDate         int64   `json:"end_date" binding:"required"`
}

type promotionResponse struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DiscountPercent float64   `json:"discount_percent"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}

func newPromotionResponse(promotion db.Promotion) promotionResponse {
	return promotionResponse{
		ID:              promotion.ID,
		Title:           promotion.Title,
		Description:     promotion.Description,
		DiscountPercent: promotion.DiscountPercent,
		StartDate:       promotion.StartDate,
		EndDate:         promotion.EndDate,
	}
}

func (server *Server) createPromotion(ctx *gin.Context) {
	var req createPromotionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	endDate := time.Now().Add(time.Duration(req.EndDate) * 24 * time.Hour)

	arg := db.CreatePromotionParams{
		Title:           req.Title,
		Description:     req.Description,
		DiscountPercent: req.DiscountPercent,
		EndDate:         endDate,
	}

	promotion, err := server.store.CreatePromotion(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newPromotionResponse(promotion)
	ctx.JSON(http.StatusOK, rsp)
}

type getPromotionRequest struct {
	Title string `uri:"title" binding:"required,alphanum"`
}

func (server *Server) getPromotionByTitle(ctx *gin.Context) {
	var req getPromotionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	promotion, err := server.store.GetPromotion(ctx, req.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newPromotionResponse(promotion)
	ctx.JSON(http.StatusOK, rsp)
}

type listPromotionRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listPromotion(ctx *gin.Context) {
	var req listPromotionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListPromotionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	promotions, err := server.store.ListPromotions(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, promotions)
}

type updatePromotionRequest struct {
	Description     string  `json:"description" binding:"required"`
	DiscountPercent float64 `json:"discount_percent" binding:"required"`
	EndDate         int64   `json:"end_date" binding:"required"`
}

func (server *Server) updatePromotion(ctx *gin.Context) {
	var reqPromotion getPromotionRequest
	var reqUpdate updatePromotionRequest

	if err := ctx.ShouldBindUri(&reqPromotion); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	promotion, err := server.store.GetPromotion(ctx, reqPromotion.Title)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	endDate := time.Now().Add(time.Duration(reqUpdate.EndDate) * 24 * time.Hour)

	arg := db.UpdatePromotionParams{
		ID:              promotion.ID,
		Description:     reqUpdate.Description,
		DiscountPercent: reqUpdate.DiscountPercent,
		EndDate:         endDate,
	}

	promotion, err = server.store.UpdatePromotion(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, promotion)
}

type deletePromotionRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deletePromotion(ctx *gin.Context) {
	var req deletePromotionRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeletePromotion(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Delete Promotion Successfully")
}
