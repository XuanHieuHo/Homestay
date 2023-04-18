package api

import (
	"fmt"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/token"
	"github.com/XuanHieuHo/homestay/util"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	api := router.Group("/api")

	// login and register
	api.POST("/login", server.loginUser)
	api.POST("/register", server.createUser)

	// -----------------------------------user--------------------------------
	authRoutes := api.Group("/").Use(authMiddleware(server.tokenMaker))
	//user
	authRoutes.GET("/users/:username", server.getUserByUsername)
	authRoutes.PUT("/users/:username", server.updateUser)
	//promotion
	authRoutes.GET("/promotions/:title", server.getPromotionByTitle)
	authRoutes.GET("/promotions/", server.listPromotion)
	// homestay
	authRoutes.GET("/homestays/:id", server.getHomestayByID)
	authRoutes.GET("/homestays/", server.listHomestay)
	// feedback
	authRoutes.POST("/users/:user_comment/feedbacks/:homestay_commented", server.createFeedback)
	authRoutes.GET("/homestays/:id/feedbacks", server.listFeedbackByID)
	authRoutes.PUT("/users/:user_comment/feedbacks/:homestay_commented/:id", server.updateFeedback)
	authRoutes.DELETE("/users/:user_comment/feedbacks/:homestay_commented/:id", server.deleteFeedback)

	// -----------------------------------admin--------------------------------
	authAdminRoutes := api.Group("/admin").Use(authAdminMiddleware(server.tokenMaker, server.store))
	// user
	authAdminRoutes.GET("/users", server.listUser)
	authAdminRoutes.GET("/users/:username", server.adminGetUserByUsername)
	authAdminRoutes.PUT("/users/:username", server.adminUpdateUser)
	authAdminRoutes.DELETE("/users/:username", server.deleteUser)
	// promotion
	authAdminRoutes.POST("/promotions", server.createPromotion)
	authAdminRoutes.GET("/promotions/:title", server.getPromotionByTitle)
	authAdminRoutes.GET("/promotions/", server.listPromotion)
	authAdminRoutes.PUT("/promotions/:title", server.updatePromotion)
	authAdminRoutes.DELETE("/promotions/:id", server.deletePromotion)
	// homestay
	authAdminRoutes.POST("/homestays", server.createHomestay)
	authAdminRoutes.GET("/homestays/:id", server.getHomestayByID)
	authAdminRoutes.GET("/homestays/", server.listHomestay)
	authAdminRoutes.PUT("/homestayStatus/:id", server.updateHomestayStatus)
	authAdminRoutes.PUT("/homestays/:id", server.updateHomestayInfo)
	authAdminRoutes.DELETE("/homestays/:id", server.deleteHomestay)
	// feedback
	authAdminRoutes.GET("/homestays/:id/feedbacks", server.listFeedbackByID)
	authAdminRoutes.DELETE("/users/:user_comment/feedbacks/:homestay_commented/:id", server.adminDeleteFeedback)

	server.router = router
}

// Start runs thes HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
