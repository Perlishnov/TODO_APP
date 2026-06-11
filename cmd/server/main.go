package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Perlishnov/TODO_APP/wire"
    "github.com/gin-gonic/gin"
)

func main() {
    app, err := wire.InitApp()
    if err != nil {
        panic(err)
    }
    logger := app.Logger

    gin.SetMode(gin.ReleaseMode)
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(gin.Logger())

    api := router.Group("/api/v1")
    app.AuthController.RegisterRoutes(api)

    authHandler := app.AuthMiddleware.Authenticate()
    app.TaskController.RegisterRoutes(api, authHandler)

    srv := &http.Server{
        Addr:    ":" + os.Getenv("SERVER_PORT"),
        Handler: router,
    }

    go func() {
        logger.Infof("Server starting on port %s", os.Getenv("SERVER_PORT"))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("server failed: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        logger.Errorf("forced shutdown: %v", err)
    }
    logger.Info("Server exited")
}