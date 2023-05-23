package main

import (
	"database/sql"
	"log"

	"github.com/XuanHieuHo/homestay/api"
	db "github.com/XuanHieuHo/homestay/db/sqlc"
	"github.com/XuanHieuHo/homestay/util"
	_ "github.com/XuanHieuHo/homestay/docs"
)

// @title 		Service API
// @version      1.0
// @description  This is a sample server celler server.
// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config: ", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	store := db.NewStore(conn)
	runGinServer(config, store)
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
