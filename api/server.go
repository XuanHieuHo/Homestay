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
	authUserRoutes := api.Group("/").Use(authMiddleware(server.tokenMaker))
	//user
	authUserRoutes.GET("/users/:username", server.getUserByUsername)
	authUserRoutes.PUT("/users/:username", server.updateUser)
	authUserRoutes.PUT("/users/:username/check", server.checkPassword)
	authUserRoutes.PUT("/users/:username/change", server.changePassword)
	//promotion
	authUserRoutes.GET("/promotions/:title", server.getPromotionByTitle)
	authUserRoutes.GET("/promotions/", server.listPromotion)
	// homestay
	authUserRoutes.GET("/homestays/:id", server.getHomestayByID)
	authUserRoutes.GET("/homestays/", server.listHomestay)
	// feedback
	authUserRoutes.POST("/users/:username/feedbacks/:homestay_commented", server.createFeedback)
	authUserRoutes.GET("/homestays/:id/feedbacks", server.listFeedbackByID)
	authUserRoutes.PUT("/users/:username/feedbacks/:homestay_commented/:id", server.updateFeedback)
	authUserRoutes.DELETE("/users/:username/feedbacks/:homestay_commented/:id", server.deleteFeedback)
	// booking
	authUserRoutes.POST("/users/:username/bookings/:homestay_booking", server.createBooking)
	authUserRoutes.PUT("/users/:username/bookings/:homestay_booking/:booking_id/cancel", server.cancelBooking)

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
	authAdminRoutes.DELETE("/users/:username/feedbacks/:homestay_commented/:id", server.adminDeleteFeedback)
	// booking
	authAdminRoutes.PUT("/users/:username/bookings/:homestay_booking/:booking_id/checkout", server.checkoutBooking)

	// -----------------------------------staff--------------------------------
	authStaffRoutes := api.Group("/staff").Use(authStaffMiddleware(server.tokenMaker, server.store))
	// booking
	authStaffRoutes.PUT("/users/:username/bookings/:homestay_booking/:booking_id/checkout", server.checkoutBooking)
	server.router = router
}

// Start runs thes HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
