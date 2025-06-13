package server

import (
	"errors"
	"fmt"
	"lakelens/cmd/db"
	"lakelens/cmd/errpipe"
	"lakelens/internal/auth"
	configs "lakelens/internal/config"
	"lakelens/internal/consts"
	iceberghdlr "lakelens/internal/handlers/iceberg"
	managerhdlr "lakelens/internal/handlers/manager"
	publichdlr "lakelens/internal/handlers/public"
	"lakelens/internal/middlewares"
	"lakelens/internal/notifications/mailer"
	icebergserv "lakelens/internal/services/iceberg"
	managersrvc "lakelens/internal/services/manager"
	publicsrvc "lakelens/internal/services/public"
	"lakelens/internal/stash"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func InitHTTPServer(ds *db.DataStore) error {

	router := gin.Default()

	// router.Use(func(c *gin.Context) {
	// 	fmt.Println("Request Origin:", c.Request.Header.Get("Origin"))
	// 	c.Next()
	// })

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	err := initRoutes(router, ds)
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

func initRoutes(router *gin.Engine, ds *db.DataStore) error {

	// < Init Request logger middleware directly on router.
	requestsLoggerFile, err := os.OpenFile(configs.Paths.RequestLoggerFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	router.Use(middlewares.Logger(requestsLoggerFile))
	// >

	// < Init the error handler.

	errPro, err := getErrProcessor()
	if err != nil {
		return err
	}
	errPro.ProcessErrors()

	// >

	// TODO:
	pool := ds.PgPool
	queries := ds.Queries
	redis := ds.Redis

	// < JWTAuth
	jwtKey, exists := os.LookupEnv("TOKEN_SIGNING_KEY")
	if !exists {
		return fmt.Errorf("jwt token signing key not found in env")
	}

	refreshAtIss := "lakelens-refreshat"
	refreshAtSub := "service:auth-account-refresh"

	accAuthIss := "lakelens-accauth"
	accAuthSub := "service:public-account-auth"

	authService := auth.NewAuthService(&auth.AuthServCreds{
		SigningKey: jwtKey,
		ATTTL:      900,
		RTTTL:      604800,

		RefreshATIssuer: refreshAtIss,
		RefrestATSub:    refreshAtSub,

		AccAuthIssuer: accAuthIss,
		AccAuthSub:    accAuthSub,
	}, redis)
	// >

	// < Middlewares

	authMid := middlewares.NewAuthMiddleware(map[string]bool{
		accAuthIss:   true,
		refreshAtIss: true,
	}, redis, authService)

	// >

	publicGrp := router.Group("/public")

	lensGrp := router.Group("/lens")
	lensGrp.Use(authMid.Authenticator())

	// < Stash
	stashService := stash.NewStashService(queries, redis, pool)
	// >

	// < Google OAuth2
	goConf, err := getGOAuthConf()
	if err != nil {
		return err
	}
	// >

	// < Mailer
	mailer, err := getMailer()
	if err != nil {
		return err
	}
	mailer.RetryErrored()
	// >

	// < Public
	publicService := publicsrvc.NewPublicService(queries, redis, pool, goConf, authService, mailer)
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

func getGOAuthConf() (*oauth2.Config, error) {

	clientID, exists := os.LookupEnv("GOOGLE_OAUTH_CLIENT_ID")
	if !exists {
		return nil, errors.New("google oAuth client id not found in env")
	}
	clientSec, exists := os.LookupEnv("GOOGLE_OAUTH_CLIENT_SECRET")
	if !exists {
		return nil, errors.New("google oAuth client secret not found in env")
	}
	fendBase, exists := os.LookupEnv("FRONTEND_URL")
	if !exists {
		return nil, errors.New("frontend url not found in env")
	}

	redirectURI := fendBase + "/api/account/oauth/google/callback"

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSec,
		RedirectURL:  redirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return conf, nil
}

func getMailer() (*mailer.EmailService, error) {

	uname, exists := os.LookupEnv("Google_SMTP_Uname")
	if !exists {
		return nil, errors.New("google smtp username not found in env")
	}
	pass, exists := os.LookupEnv("Google_SMTP_Pass")
	if !exists {
		return nil, errors.New("google smtp pass not found in env")
	}
	host, exists := os.LookupEnv("Google_SMTP_Host")
	if !exists {
		return nil, errors.New("google smtp host not found in env")
	}
	hostAddr, exists := os.LookupEnv("Google_SMTP_HostAddr")
	if !exists {
		return nil, errors.New("google smtp host address not found in env")
	}
	from, exists := os.LookupEnv("Google_SMTP_From")
	if !exists {
		return nil, errors.New("google smtp from not found in env")
	}

	cfg := mailer.SMTPConfig{
		Username: uname,
		Password: pass,
		Host:     host,
		HostAddr: hostAddr,
		From:     from,
	}
	gmail := mailer.NewGmailMailer(cfg)

	var m mailer.Mailer = gmail
	emailServ := mailer.NewEmailService(m)

	return emailServ, nil
}

func getErrProcessor() (*errpipe.ErrorProcessor, error) {

	telegramBotToken, exists := os.LookupEnv("Telegram_Bot_Token")
	if !exists || telegramBotToken == "" {
		return nil, errors.New("Telegram_Bot_Token not found or is empty string")
	}

	telegramChatID, exists := os.LookupEnv("Telegram_Chat_ID")
	if !exists || telegramChatID == "" {
		return nil, errors.New("Telegram_Chat_ID not found or is empty string")
	}

	teleHandler, err := errpipe.NewTelegramErrorHandler(telegramBotToken, telegramChatID)
	if err != nil {
		return nil, err
	}

	var ep errpipe.ErrorHandler = teleHandler
	errProcessor := errpipe.NewErrorProcessor(ep)

	return errProcessor, nil

}
