package api

import (
	"database/sql"
	"log"
	"net/http"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/util"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

// Image input is base64
type createHomestayRequest struct {
	Description string  `json:"description" binding:"required"`
	Address     string  `json:"address" binding:"required"`
	NumberOfBed int32   `json:"number_of_bed" binding:"required"`
	Capacity    int32   `json:"capacity" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	MainImage   string  `json:"main_image" binding:"required"`
	FirstImage  string  `json:"first_image" binding:"required"`
	SecondImage string  `json:"second_image" binding:"required"`
	ThirdImage  string  `json:"third_image" binding:"required"`
}
// @Summary Admin Create Homestay
// @ID createHomestay
// @Produce json
// @Accept json
// @Tags Admin
// @Param data body createHomestayRequest true "createHomestayRequest data"
// @Security bearerAuth
// @Success 200 {object} db.Homestay
// @Failure 400 {string} error
// @Failure 500 {string} error
// @Router /api/admin/homestays [post]
func (server *Server) createHomestay(ctx *gin.Context) {
	// create connection to cloudinary
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("Cannot load config: ", err)
	}
	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var req createHomestayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	params := uploader.UploadParams{
		Folder:         "homestay",
		Format:         "jpg",
		Transformation: "f_auto,fl_lossy,q_auto:eco,dpr_auto,w_auto",
	}

	mainImg, err := cld.Upload.Upload(ctx, req.MainImage, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Upload other images
	firstImg, err := cld.Upload.Upload(ctx, req.FirstImage, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	secondImg, err := cld.Upload.Upload(ctx, req.SecondImage, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	thirdImg, err := cld.Upload.Upload(ctx, req.ThirdImage, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateHomestayParams{
		Description: req.Description,
		Address:     req.Address,
		NumberOfBed: req.NumberOfBed,
		Capacity:    req.Capacity,
		Price:       req.Price,
		Status:      "available",
		MainImage:   mainImg.SecureURL,
		FirstImage:  firstImg.SecureURL,
		SecondImage: secondImg.SecureURL,
		ThirdImage:  thirdImg.SecureURL,
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

type getHomestayByIDResponse struct {
	Homestay []struct {
		db.Homestay          `json:"homestay"`
		listFeedbackResponse `json:"list_of_feedbacks"`
	} `json:"homestays"`
}
// @Summary Get Homestay By ID
// @ID getHomestayByID
// @Produce json
// @Accept json
// @Tags Started
// @Param id path string true "ID"
// @Param data query listHomestayRequest true "listHomestayRequest data"
// @Success 200 {object} getHomestayByIDResponse
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/homestays/{id} [get]
func (server *Server) getHomestayByID(ctx *gin.Context) {
	var req getHomestayRequest
	var reqList listHomestayRequest
	var result getHomestayByIDResponse
	var resultFeedbacks listFeedbackResponse
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindQuery(&reqList); err != nil {
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
	feedbacks, err := server.store.ListFeedbacks(ctx, db.ListFeedbacksParams{
		HomestayCommented: homestay.ID,
		Limit:             reqList.PageSize,
		Offset:            (reqList.PageID - 1) * reqList.PageSize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	for _, feedback := range feedbacks {
		user, err := server.store.GetUser(ctx, feedback.UserComment)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		userResult := newUserResponse(user)
		resultFeedbacks.Feedbacks = append(resultFeedbacks.Feedbacks, struct {
			db.Feedback `json:"feedback"`
			User        userResponse `json:"commentor"`
		}{feedback, userResult})
	}
	result.Homestay = append(result.Homestay, struct {
		db.Homestay          `json:"homestay"`
		listFeedbackResponse `json:"list_of_feedbacks"`
	}{homestay, resultFeedbacks})

	ctx.JSON(http.StatusOK, result)
}

type listHomestayRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

type listHomestayResponse struct {
	Homestays []struct {
		db.Homestay          `json:"homestay"`
		listFeedbackResponse `json:"list_of_feedbacks"`
	} `json:"homestays"`
}

// @Summary Get List Homestay
// @ID listHomestay
// @Produce json
// @Accept json
// @Param data query listHomestayRequest true "listHomestayRequest data"
// @Tags Started
// @Success 200 {object} listHomestayResponse
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 500 {string} error
// @Router /api/homestays [get]
func (server *Server) listHomestay(ctx *gin.Context) {
	var req listHomestayRequest
	var result listHomestayResponse
	var resultFeedbacks listFeedbackResponse
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

	for _, homestay := range homestays {
		feedbacks, err := server.store.ListFeedbacks(ctx, db.ListFeedbacksParams{
			HomestayCommented: homestay.ID,
			Limit:             req.PageSize,
			Offset:            (req.PageID - 1) * req.PageSize,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		for _, feedback := range feedbacks {
			user, err := server.store.GetUser(ctx, feedback.UserComment)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
			userResult := newUserResponse(user)
			resultFeedbacks.Feedbacks = append(resultFeedbacks.Feedbacks, struct {
				db.Feedback `json:"feedback"`
				User        userResponse `json:"commentor"`
			}{feedback, userResult})
		}

		result.Homestays = append(result.Homestays, struct {
			db.Homestay          `json:"homestay"`
			listFeedbackResponse `json:"list_of_feedbacks"`
		}{homestay, resultFeedbacks})
	}
	ctx.JSON(http.StatusOK, result)
}

type updateHomestayStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
// @Summary Admin Update Homestay Status
// @ID updateHomestayStatus
// @Produce json
// @Accept json
// @Param data body updateHomestayStatusRequest true "updateHomestayStatusRequest data"
// @Param id path string true "ID"
// @Security bearerAuth
// @Tags Admin
// @Success 200 {object} db.Homestay
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/admin/homestayStatus/{id} [put]
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
// @Summary Admin Update Homestay Info
// @ID updateHomestayInfo
// @Produce json
// @Accept json
// @Param data body createHomestayRequest true "createHomestayRequest data"
// @Param id path string true "ID"
// @Security bearerAuth
// @Tags Admin
// @Success 200 {object} db.Homestay
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/admin/homestayStatus/{id} [put]
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
// @Summary Admin Delete Homestay
// @ID deleteHomestay
// @Produce json
// @Accept json
// @Tags Admin
// @Param id path string true "ID"
// @Security bearerAuth
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 500 {string} error
// @Router /api/admin/homestays/{id} [delete]
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
