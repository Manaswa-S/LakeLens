package server

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"main.go/cmd/db"
	"main.go/internal/handlers"
	"main.go/internal/services"
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
	queries := db.QueriesPool
	redis := db.RedisClient
	pool := db.Pool

	services := services.NewService(queries, redis, pool)
	handlers := handlers.NewHandler(services)
	handlers.RegisterRoutes(internalGroup)



	return nil
}






