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

type createFeedbackRequest struct {
	Rating     string `json:"rating" binding:"required"`
	Commention string `json:"commention" binding:"required"`
}

type getUserAndHomestay struct {
	UserComment       string `uri:"username" binding:"required,alphanum"`
	HomestayCommented int64  `uri:"homestay_commented" binding:"required,min=1"`
}
// @Summary User Create Feedback
// @ID createFeedback
// @Produce json
// @Accept json
// @Tags User
// @Param data body createFeedbackRequest true "createFeedbackRequest data"
// @Param username path string true "UserComment"
// @Param homestay_commented path string true "HomestayCommented"
// @Security bearerAuth
// @Success 200 {object} db.Feedback
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/feedbacks/{homestay_commented} [post]
func (server *Server) createFeedback(ctx *gin.Context) {
	var reqCreate createFeedbackRequest
	var reqGet getUserAndHomestay

	if err := ctx.ShouldBindUri(&reqGet); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqCreate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqGet.UserComment)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, reqGet.HomestayCommented)
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

	arg := db.CreateFeedbackParams{
		UserComment:       user.Username,
		HomestayCommented: homestay.ID,
		Rating:            reqCreate.Rating,
		Commention:        reqCreate.Commention,
		CreatedAt:         time.Now(),
	}

	feedback, err := server.store.CreateFeedback(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, feedback)
}

type listFeedbackRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=20"`
}
type listFeedbackResponse struct {
	Feedbacks []struct {
		db.Feedback `json:"feedback"`
		User userResponse `json:"commentor"`
	} `json:"feedbacks"`
}
// @Summary User Get List Feedback About Homestay
// @ID listFeedbackByID
// @Produce json
// @Accept json
// @Tags User
// @Param data query listFeedbackRequest true "listFeedbackRequest data"
// @Param id path string true "ID"
// @Security bearerAuth
// @Success 200 {array} listFeedbackResponse
// @Failure 400 {string} error
// @Failure 500 {string} error
// @Router /api/homestays/{id}/feedbacks [get]
func (server *Server) listFeedbackByID(ctx *gin.Context) {
	var reqHomestay getHomestayRequest
	var req listFeedbackRequest
	var result listFeedbackResponse
	if err := ctx.ShouldBindUri(&reqHomestay); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListFeedbacksParams{
		HomestayCommented: reqHomestay.ID,
		Limit:             req.PageSize,
		Offset:            (req.PageID - 1) * req.PageSize,
	}

	feedbacks, err := server.store.ListFeedbacks(ctx, arg)
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
		result.Feedbacks = append(result.Feedbacks, struct {
			db.Feedback `json:"feedback"`
			User userResponse `json:"commentor"`
		}{feedback, userResult})
	}
	ctx.JSON(http.StatusOK, result)
}

type updateFeedbackRequest struct {
	ID                int64  `uri:"id" binding:"required,min=1"`
	UserComment       string `uri:"username" binding:"required,alphanum"`
	HomestayCommented int64  `uri:"homestay_commented" binding:"required,min=1"`
}
// @Summary User Update Feedback
// @ID updateFeedback
// @Produce json
// @Accept json
// @Tags User
// @Param data body createFeedbackRequest true "createFeedbackRequest data"
// @Param id path string true "ID"
// @Param username path string true "UserComment"
// @Param homestay_commented path string true "HomestayCommented"
// @Security bearerAuth
// @Success 200 {object} db.Feedback
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/feedbacks/{homestay_commented}/{id} [put]
func (server *Server) updateFeedback(ctx *gin.Context) {
	var reqCreate createFeedbackRequest
	var reqGet updateFeedbackRequest

	if err := ctx.ShouldBindUri(&reqGet); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqCreate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqGet.UserComment)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, reqGet.HomestayCommented)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	feedback, err := server.store.GetFeedback(ctx, reqGet.ID)
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

	if feedback.UserComment != user.Username {
		err := errors.New("this feedback doesn't belong to authenticated user ")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if feedback.HomestayCommented != homestay.ID {
		err := errors.New("this feedback doesn't belong to authenticated homestay ")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateFeedbackParams{
		ID:         feedback.ID,
		Rating:     reqCreate.Rating,
		Commention: reqCreate.Commention,
	}

	feedbackUpdate, err := server.store.UpdateFeedback(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, feedbackUpdate)
}
// @Summary User Delete Feedback
// @ID deleteFeedback
// @Produce json
// @Accept json
// @Tags User
// @Param id path string true "ID"
// @Param username path string true "UserComment"
// @Param homestay_commented path string true "HomestayCommented"
// @Security bearerAuth
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/feedbacks/{homestay_commented}/{id} [delete]
func (server *Server) deleteFeedback(ctx *gin.Context) {
	var reqGet updateFeedbackRequest

	if err := ctx.ShouldBindUri(&reqGet); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqGet.UserComment)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, reqGet.HomestayCommented)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	feedback, err := server.store.GetFeedback(ctx, reqGet.ID)
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

	if feedback.UserComment != user.Username {
		err := errors.New("this feedback doesn't belong to authenticated user ")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if feedback.HomestayCommented != homestay.ID {
		err := errors.New("this feedback doesn't belong to authenticated homestay ")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.store.DeleteFeedback(ctx, feedback.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Delete Feedback Successfully")
}
// @Summary Admin Delete Feedback
// @ID adminDeleteFeedback
// @Produce json
// @Accept json
// @Tags Admin
// @Param id path string true "ID"
// @Param username path string true "UserComment"
// @Param homestay_commented path string true "HomestayCommented"
// @Security bearerAuth
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/admin/users/{username}/feedbacks/{homestay_commented}/{id} [delete]
func (server *Server) adminDeleteFeedback(ctx *gin.Context) {
	var reqGet updateFeedbackRequest

	if err := ctx.ShouldBindUri(&reqGet); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqGet.UserComment)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	homestay, err := server.store.GetHomestay(ctx, reqGet.HomestayCommented)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	feedback, err := server.store.GetFeedback(ctx, reqGet.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if feedback.UserComment != user.Username {
		err := errors.New("this feedback doesn't belong to authenticated user ")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if feedback.HomestayCommented != homestay.ID {
		err := errors.New("this feedback doesn't belong to authenticated homestay ")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err = server.store.DeleteFeedback(ctx, feedback.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Delete Feedback Successfully")

}
