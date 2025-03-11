package server

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"main.go/internal/handlers"
	"main.go/internal/services"
	sqlc "main.go/internal/sqlc/generate"
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

	_ = router.Group("/public")

	internalGroup := router.Group("/internal")
	internalGroup.Use()
	
	// TODO:
	var queries *sqlc.Queries
	var redis *redis.Client

	services := services.NewService(queries, redis)
	handlers := handlers.NewHandler(services)
	handlers.RegisterRoutes(internalGroup)



	return nil
}






