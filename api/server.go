package api

import (
	"fmt"

	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/token"
	"github.com/XuanHieuHo/homestay/util"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config util.Config
	store db.Store
	tokenMaker token.Maker
	router *gin.Engine
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

	// login and register
	router.POST("/login", server.loginUser)
	router.POST("/register", server.createUser)


	// user
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.GET("/users/:username", server.getUserByUsername)
	authRoutes.PUT("/users/:username", server.updateUser)


	// admin
	authAdminRoutes := router.Group("/admin").Use(authAdminMiddleware(server.tokenMaker, server.store))
	authAdminRoutes.GET("/users", server.listUser)
	authAdminRoutes.GET("/users/:username", server.adminGetUserByUsername)
	authAdminRoutes.PUT("/users/:username", server.adminUpdateUser)
	authAdminRoutes.DELETE("/users/:username", server.deleteUser)

	server.router = router
}

// Start runs thes HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}