package server

import (
	"fmt"
	"lakelens/cmd/db"
	"lakelens/internal/consts"
	"lakelens/internal/handlers"
	iceberghdlr "lakelens/internal/handlers/iceberg"
	publichdlr "lakelens/internal/handlers/public"
	publicsrvc "lakelens/internal/services/public"
	"lakelens/internal/services"
	icebergserv "lakelens/internal/services/iceberg"
	"lakelens/internal/stash"
	"os"

	"github.com/gin-gonic/gin"
)


func InitHTTPServer() (error) {

	router := gin.Default()
	err := initRoutes(router)
	if err != nil {
		return err
	}

	go func ()  {
		err := router.Run(":" + os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
	} ()

	
	return nil
}

func initRoutes(router *gin.Engine) error {

	publicGrp := router.Group("/public")

	internalGroup := router.Group("/lens")
	internalGroup.Use()
	
	// TODO:
	queries := db.QueriesPool
	redis := db.RedisClient
	pool := db.Pool
	

	// < Stash
	stashService := stash.NewStashService(queries, redis, pool)
	// >

	// < Trial
	service := services.NewService(queries, redis, pool)
	handler := handlers.NewHandler(service)
	handler.RegisterRoutes(internalGroup)
	// >

	// < Public
	publicService := publicsrvc.NewPublicService(queries, redis, pool)
	publicHdlr := publichdlr.NewPublicHandler(publicService)
	publicHdlr.RegisterRoutes(publicGrp)
	// >


	// < Iceberg
	icebergService := icebergserv.NewIcebergService(queries, redis, pool, stashService)
	icebergHandler := iceberghdlr.NewIcebergHandler(icebergService)
	icebergGrp := router.Group("/" + consts.IcebergTable)
	icebergHandler.RegisterRoutes(icebergGrp)
	// >

	// < Manager
	managerService := services.NewManagerService(queries, redis, pool, stashService, icebergService)
	managerHandler := handlers.NewManagerHandler(managerService)
	managerGrp := router.Group("/manager")
	managerHandler.RegisterRoutes(managerGrp)
	// >

	return nil
}






