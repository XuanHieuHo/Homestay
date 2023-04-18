package api

import (
	"database/sql"
	"net/http"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createHomestayRequest struct {
	Description string `json:"description" binding:"required"`
	Address     string `json:"address" binding:"required"`
	NumberOfBed int32  `json:"number_of_bed" binding:"required"`
	Capacity    int32  `json:"capacity" binding:"required"`
	Price       string `json:"price" binding:"required"`
	MainImage   string `json:"main_image" binding:"required"`
	FirstImage  string `json:"first_image" binding:"required"`
	SecondImage string `json:"second_image" binding:"required"`
	ThirdImage  string `json:"third_image" binding:"required"`
}

func (server *Server) createHomestay(ctx *gin.Context) {
	var req createHomestayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateHomestayParams{
		Description: req.Description,
		Address:     req.Address,
		NumberOfBed: req.NumberOfBed,
		Capacity:    req.Capacity,
		Price:       req.Price,
		Status:      "available",
		MainImage:   req.MainImage,
		FirstImage:  req.FirstImage,
		SecondImage: req.SecondImage,
		ThirdImage:  req.ThirdImage,
	}

	homestay, err := server.store.CreateHomestay(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, homestay)
}

type getHomestayRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getHomestayByID(ctx *gin.Context) {
	var req getHomestayRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, homestay)
}

type listHomestayRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listHomestay(ctx *gin.Context) {
	var req listHomestayRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListHomestaysParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	homestays, err := server.store.ListHomestays(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, homestays)
}

type updateHomestayStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (server *Server) updateHomestayStatus(ctx *gin.Context) {
	var reqHomestay getHomestayRequest
	var reqUpdate updateHomestayStatusRequest

	if err := ctx.ShouldBindUri(&reqHomestay); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, reqHomestay.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.UpdateHomestayStatusParams{
		ID:     homestay.ID,
		Status: reqUpdate.Status,
	}

	homestay, err = server.store.UpdateHomestayStatus(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, homestay)
}

func (server *Server) updateHomestayInfo(ctx *gin.Context) {
	var reqHomestay getHomestayRequest
	var reqUpdate createHomestayRequest

	if err := ctx.ShouldBindUri(&reqHomestay); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, reqHomestay.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.UpdateHomestayInfoParams{
		ID:          homestay.ID,
		Description: reqUpdate.Description,
		Address:     reqUpdate.Address,
		NumberOfBed: reqUpdate.NumberOfBed,
		Capacity:    reqUpdate.Capacity,
		Price:       reqUpdate.Price,
		MainImage:   reqUpdate.MainImage,
		FirstImage:  reqUpdate.FirstImage,
		SecondImage: reqUpdate.SecondImage,
		ThirdImage:  reqUpdate.ThirdImage,
	}

	homestay, err = server.store.UpdateHomestayInfo(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, homestay)
}

func (server *Server) deleteHomestay(ctx *gin.Context) {
	var req getHomestayRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteHomestay(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Delete Homestay Successfully")
}
