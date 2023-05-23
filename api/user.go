package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/mail"
	"github.com/XuanHieuHo/homestay/token"
	"github.com/XuanHieuHo/homestay/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// swagger:parameters createUserRequest
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,e164"`
}
// swagger:response userResponse
type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:  user.Username,
		FullName:  user.FullName,
		Email:     user.Email,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
	}
}
// @Summary Create user
// @ID createUser
// @Produce json
// @Accept json
// @Param data body createUserRequest true "createUserRequest data"
// @Tags Started
// @Success 200 {object} userResponse
// @Failure 400 {string} error
// @Failure 403 {string} error
// @Failure 500 {string} error
// @Router /api/register [post]
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
		Phone:          req.Phone,
		CreatedAt:      time.Now(),
	}

	user, err := server.store.CreateUser(ctx, arg)
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
	rsp := newUserResponse(user)
	ctx.JSON(http.StatusOK, rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

// @Summary Login user
// @ID loginUser
// @Produce json
// @Accept json
// @Param data body loginUserRequest true "loginUserRequest data"
// @Tags Started
// @Success 200 {object} loginUserResponse
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 403 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/login [post]
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
}

type forgotPasswordRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Email    string `json:"email" binding:"required,email"`
}
// @Summary Send Reset Password Token
// @ID sendResetPasswordToken
// @Produce json
// @Accept json
// @Param data body forgotPasswordRequest true "forgotPasswordRequest data"
// @Tags Started
// @Success 200 {string} succesfully
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/forgotpassword [post]
func (server *Server) sendResetPasswordToken(ctx *gin.Context) {
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal("Cannot load config: ", err)
	}
	var req forgotPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	userEmail, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if user.Username != userEmail.Username {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	resetToken := util.RandomResetPasswordToken()

	_, err = server.store.UpdateResetPasswordToken(ctx, db.UpdateResetPasswordTokenParams{
		Username:                 user.Username,
		ResetPasswordToken:       resetToken,
		RspasswordTokenExpiredAt: time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	sender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "Homestay verification code"
	content := fmt.Sprintf(`
		<html>
	  	<head>
	    <title>Mã xác nhận đổi mật khẩu</title>
	    <style>
	      * {
	        box-sizing: border-box;
	        margin: 0;
	        padding: 0;
	      }

	      body {
	        background-color: #f8f8f8;
	        font-family: sans-serif;
	        line-height: 1.5;
	        color: #333;
	      }

	      .container {
	        max-width: 800px;
	        margin: 0 auto;
	        padding: 20px;
	        background-color: #fff;
	        border-radius: 5px;
	        box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.1);
	      }

	      h1 {
	        font-size: 24px;
	        font-weight: bold;
	        margin-bottom: 20px;
	      }
	      p {
	        font-size: 18px;
	        margin-bottom: 20px;
	      }
	    </style>
	  </head>
	  <body>
	    <div class="container">
	      <h1>Giới thiệu Homestay</h1>
	      <p>
	        Chào mừng đến với Homestay của chúng tôi! Chúng tôi cung cấp các dịch
	        vụ lưu trú tại nhà ấm cúng và tiện nghi để bạn có một kỳ nghỉ tuyệt vời.
	        Với các phòng ngủ được trang bị đầy đủ tiện nghi, chúng tôi hy vọng sẽ
	        mang đến cho bạn một trải nghiệm nghỉ dưỡng tuyệt vời.
	      </p>
	      <p>
	        Đây là mã xác nhận của bạn: %s
	      </p>
	    </div>
	  </body>
	</html>
		`, resetToken)

	to := []string{user.Email}

	err = sender.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, "reset password link has been sent to your email")

}

type resetPasswordRequest struct {
	Username       string `json:"username" binding:"required,alphanum"`
	OTPCode        string `json:"otpcode" binding:"required"`
	FirstPassword  string `json:"first_password" binding:"required,min=6"`
	SecondPassword string `json:"second_password" binding:"required,min=6"`
}
// @Summary Reset Password
// @ID resetPassword
// @Produce json
// @Accept json
// @Param data body resetPasswordRequest true "resetPasswordRequest data"
// @Tags Started
// @Success 200 {string} succesfully
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/resetpassword [post]
func (server *Server) resetPassword(ctx *gin.Context) {
	var req resetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if user.ResetPasswordToken != req.OTPCode {
		err := errors.New("opt code is wrong")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if time.Now().After(user.RspasswordTokenExpiredAt) {
		err := errors.New("opt code has expired")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.FirstPassword != req.SecondPassword {
		err := errors.New("two password don't match")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.FirstPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.ChangeUserPasswordParams{
		Username:          user.Username,
		HashedPassword:    hashedPassword,
		PasswordChangedAt: time.Now(),
	}

	_, err = server.store.ChangeUserPassword(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "password has been changed successfully")
}

type getUserRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}
// @Summary Get User By Username
// @ID getUserByUsername
// @Produce json
// @Accept json
// @Tags User
// @Security bearerAuth
// @Param username path string true "Username"
// @Success 200 {object} userResponse
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username} [get]
func (server *Server) getUserByUsername(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
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

	rsp := userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		Phone:             user.Phone,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}
// @Summary Admin Get User By Username
// @ID adminGetUserByUsername
// @Produce json
// @Accept json
// @Tags Admin
// @Security bearerAuth
// @Param username path string true "Username"
// @Success 200 {object} userResponse
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/admin/users/{username} [get]
func (server *Server) adminGetUserByUsername(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		Phone:             user.Phone,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}

type listUserRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}
// @Summary Get List User
// @ID listUser
// @Produce json
// @Accept json
// @Param data query listUserRequest true "listUserRequest data"
// @Tags Admin
// @Security bearerAuth
// @Success 200 {array} db.User
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/admin/users/ [get]
func (server *Server) listUser(ctx *gin.Context) {
	var req listUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, users)
}

type updateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,e164"`
}
// @Summary Update User
// @ID updateUser
// @Produce json
// @Accept json
// @Param data body updateUserRequest true "updateUserRequest data"
// @Param username path string true "Username"
// @Tags User
// @Security bearerAuth
// @Success 200 {object} db.User
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username} [put]
func (server *Server) updateUser(ctx *gin.Context) {
	var reqUsername getUserRequest
	var reqUpdate updateUserRequest

	if err := ctx.ShouldBindUri(&reqUsername); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqUsername.Username)
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

	arg := db.UpdateUserParams{
		Username: reqUsername.Username,
		FullName: reqUpdate.FullName,
		Email:    reqUpdate.Email,
		Phone:    reqUpdate.Phone,
	}

	user, err = server.store.UpdateUser(ctx, arg)
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

	ctx.JSON(http.StatusOK, user)
}
// @Summary Admin Update User
// @ID adminUpdateUser
// @Produce json
// @Accept json
// @Param data body updateUserRequest true "updateUserRequest data"
// @Param username path string true "Username"
// @Tags Admin
// @Security bearerAuth
// @Success 200 {object} db.User
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/admin/users/{username} [put]
func (server *Server) adminUpdateUser(ctx *gin.Context) {
	var reqUsername getUserRequest
	var reqUpdate updateUserRequest

	if err := ctx.ShouldBindUri(&reqUsername); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqUpdate); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateUserParams{
		Username: reqUsername.Username,
		FullName: reqUpdate.FullName,
		Email:    reqUpdate.Email,
		Phone:    reqUpdate.Phone,
	}

	user, err := server.store.UpdateUser(ctx, arg)
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

	ctx.JSON(http.StatusOK, user)
}
// @Summary Admin Delete User
// @ID deleteUser
// @Produce json
// @Accept json
// @Tags Admin
// @Param username path string true "Username"
// @Security bearerAuth
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 500 {string} error
// @Router /api/admin/users/{username} [delete]
func (server *Server) deleteUser(ctx *gin.Context) {
	var req getUserRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteUser(ctx, req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, "Delete User Successfully")
}

type checkPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}
// @Summary User Check Password
// @ID checkPassword
// @Produce json
// @Accept json
// @Param data body checkPasswordRequest true "checkPasswordRequest data"
// @Param username path string true "Username"
// @Tags User
// @Security bearerAuth
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/check [put]
func (server *Server) checkPassword(ctx *gin.Context) {
	var reqUsername getUserRequest
	var reqPassword checkPasswordRequest

	if err := ctx.ShouldBindJSON(&reqPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindUri(&reqUsername); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqUsername.Username)
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

	err = util.CheckPassword(reqPassword.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, "the password is correct")
}

type newPasswordRequest struct {
	OriginalPassword string `json:"original_password" binding:"required,min=6"`
	FirstPassword    string `json:"first_password" binding:"required,min=6"`
	SecondPassword   string `json:"second_password" binding:"required,min=6"`
}
// @Summary User Change Password
// @ID changePassword
// @Produce json
// @Accept json
// @Param data body newPasswordRequest true "newPasswordRequest data"
// @Param username path string true "Username"
// @Tags User
// @Security bearerAuth
// @Success 200 {string} successfully
// @Failure 400 {string} error
// @Failure 401 {string} error
// @Failure 403 {string} error
// @Failure 404 {string} error
// @Failure 500 {string} error
// @Router /api/users/{username}/change [put]
func (server *Server) changePassword(ctx *gin.Context) {
	var reqUsername getUserRequest
	var reqPassword newPasswordRequest

	if err := ctx.ShouldBindJSON(&reqPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindUri(&reqUsername); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, reqUsername.Username)
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

	err = util.CheckPassword(reqPassword.OriginalPassword, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if reqPassword.FirstPassword != reqPassword.SecondPassword {
		err := errors.New("two password don't match")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(reqPassword.FirstPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.ChangeUserPasswordParams{
		Username:          reqUsername.Username,
		HashedPassword:    hashedPassword,
		PasswordChangedAt: time.Now(),
	}

	_, err = server.store.ChangeUserPassword(ctx, arg)
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

	ctx.JSON(http.StatusOK, "password has been changed successfully")
}
