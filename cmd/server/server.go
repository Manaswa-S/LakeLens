package server

import (
	"fmt"
	"lakelens/cmd/db"
	"lakelens/internal/consts"
	iceberghdlr "lakelens/internal/handlers/iceberg"
	managerhdlr "lakelens/internal/handlers/manager"
	publichdlr "lakelens/internal/handlers/public"
	icebergserv "lakelens/internal/services/iceberg"
	managersrvc "lakelens/internal/services/manager"
	publicsrvc "lakelens/internal/services/public"
	"lakelens/internal/stash"
	"os"

	"github.com/gin-gonic/gin"
)

func InitHTTPServer() error {

	router := gin.Default()
	err := initRoutes(router)
	if err != nil {
		return err
	}

	go func() {
		err := router.Run(":" + os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	return nil
}

func initRoutes(router *gin.Engine) error {

	publicGrp := router.Group("/public")

	lensGrp := router.Group("/lens")
	// TODO: add middlewares
	lensGrp.Use()

	// TODO:
	queries := db.QueriesPool
	redis := db.RedisClient
	pool := db.Pool

	// < Stash
	stashService := stash.NewStashService(queries, redis, pool)
	// >

	// < Public
	publicService := publicsrvc.NewPublicService(queries, redis, pool)
	publicHdlr := publichdlr.NewPublicHandler(publicService)
	publicHdlr.RegisterRoutes(publicGrp)
	// >

	// < Iceberg
	icebergService := icebergserv.NewIcebergService(queries, redis, pool, stashService)
	icebergHandler := iceberghdlr.NewIcebergHandler(icebergService)
	icebergGrp := lensGrp.Group("/" + consts.IcebergTable)
	icebergHandler.RegisterRoutes(icebergGrp)
	// >

	// < Manager
	managerService := managersrvc.NewManagerService(queries, redis, pool, stashService, icebergService)
	managerHandler := managerhdlr.NewManagerHandler(managerService)
	managerGrp := lensGrp.Group("/manager")
	managerHandler.RegisterRoutes(managerGrp)
	// >

	return nil
}
