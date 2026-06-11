package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/slodkiadrianek/Go-API-template/common/middleware"
	config "github.com/slodkiadrianek/Go-API-template/config"
	authHandler "github.com/slodkiadrianek/Go-API-template/internal/auth/handler"
	authRepository "github.com/slodkiadrianek/Go-API-template/internal/auth/repository"
	authService "github.com/slodkiadrianek/Go-API-template/internal/auth/service"
	mailService "github.com/slodkiadrianek/Go-API-template/internal/mail"

	"github.com/slodkiadrianek/Go-API-template/common/log"
	"github.com/slodkiadrianek/Go-API-template/internal/server"
	userRepository "github.com/slodkiadrianek/Go-API-template/internal/user/repository"
)

func main() {
	loggerService := log.NewLogger("./logs", "2006-01-02", "15:04:05")
	defer func() {
		if closeErr := loggerService.Close(); closeErr != nil {
			fmt.Printf("failed to properly close file with logs:%s", closeErr.Error())
		}
	}()
	err := config.SetupEnvVariables("./.env")
	if err != nil {
		panic(err)
	}

	_, ok := os.LookupEnv("HOST_LINK")
	if !ok {
		err := errors.New("HOST_LINK variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "HOST_LINK",
		})
		panic(err)
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		err := errors.New("PORT variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "PORT",
		})
		panic(err)
	}

	dbLink, ok := os.LookupEnv("DB_LINK")
	if !ok {
		err := errors.New("DB_LINK variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "DB_LINK",
		})
		panic(err)
	}

	db, err := config.NewDB(dbLink, "postgres")
	if err != nil {
		loggerService.Error("Failed to connect to database", err)
		panic(err)
	}

	cacheConnectionLink, ok := os.LookupEnv("CACHE_LINK")
	if !ok {
		err := errors.New("CACHE_LINK variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "CACHE_LINK",
		})
		panic(err)
	}

	cacheService, err := config.NewCacheService(cacheConnectionLink)
	if err != nil {
		loggerService.Error("failed to connect to cache service", err.Error())
		panic(err)
	}

	accessTokenSecret, ok := os.LookupEnv("ACCESS_TOKEN_SECRET")
	if !ok {
		err := errors.New("ACCESS_TOKEN_SECRET variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "ACCESS_TOKEN_SECRET",
		})
		panic(err)
	}

	refreshTokenSecret, ok := os.LookupEnv("REFRESH_TOKEN_SECRET")
	if !ok {
		err := errors.New("REFRESH_TOKEN_SECRET variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "REFRESH_TOKEN_SECRET",
		})
		panic(err)
	}

	hostEmail, ok := os.LookupEnv("HOST_EMAIL")
	if !ok {
		err := errors.New("HOST_EMAIL variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "HOST_EMAIL",
		})
		panic(err)
	}

	passwordEmail, ok := os.LookupEnv("PASSWORD_EMAIL")
	if !ok {
		err := errors.New("PASSWORD_EMAIL variable has not been initialized")
		loggerService.Error(err.Error(), map[string]string{
			"variable": "PASSWORD_EMAIL",
		})
		panic(err)
	}

	authenticationMiddleware := middleware.NewAuthenticationMiddleware(accessTokenSecret, refreshTokenSecret, loggerService, cacheService)
	emailService := mailService.NewEmailService(hostEmail, passwordEmail, loggerService)
	userRepository := userRepository.NewUserRepository(loggerService, db.DBConnection)
	authRepository := authRepository.NewAuthRepository(loggerService, db.DBConnection)
	authService := authService.NewAuthService(loggerService, userRepository, authRepository, authenticationMiddleware, emailService)
	authController := authHandler.NewAuthHandler(loggerService, authService, authenticationMiddleware)

	dependenciesConfig := server.NewDependencyConfig(port, *authController)
	apiCtx, apiCtxCancel := context.WithCancel(context.Background())
	httpServer := server.NewServer(dependenciesConfig)
	go func() {
		loggerService.Info(fmt.Sprintf("server started at port:%s", port), nil)
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			loggerService.Error("Failed to start server", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	defer apiCtxCancel()

	if err := httpServer.Shutdown(apiCtx); err != nil {
		loggerService.Error("Server forced to shutdown:", err)
		err := loggerService.Close()
		if err != nil {
			panic(err)
		}
		err = db.Close()
		if err != nil {
			panic(err)
		}
		err = cacheService.Close()
		if err != nil {
			panic(err)
		}
	}

	loggerService.Info("server exited", nil)
}
