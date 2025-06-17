package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "Dowlet_projects/ecommerce/api"
    "Dowlet_projects/ecommerce/config"
    _ "Dowlet_projects/ecommerce/docs"
    "Dowlet_projects/ecommerce/services"

    "github.com/gorilla/mux"
    "github.com/subosito/gotenv"
    httpSwagger "github.com/swaggo/http-swagger"
)

// @title E-commerce API
// @version 1.0
// @description API for an e-commerce platform with OTP-based authentication and product/market management.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication (e.g., "Bearer <token>")
func main() {
    // Load .env file
    if err := gotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
        log.Fatalf("Failed to load .env file: %v", err)
    }
    // // Debug: Log loaded environment variables
    // log.Printf("Loaded env: DB_USER=%s, DB_PASSWORD=%s, DB_NAME=%s, JWT_SECRET=%s",
    //     os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("JWT_SECRET"))

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database and Redis service
    dbService, err := services.NewDBService(cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.Redis)
    if err != nil {
        log.Fatalf("Failed to connect to database or Redis: %v", err)
    }
    defer dbService.Close()

    // Initialize router
    router := mux.NewRouter()

    // Serve static files with security headers
    fileServer := http.FileServer(http.Dir("uploads"))
    router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        fileServer.ServeHTTP(w, r)
    })))

    // Initialize handler with configuration and database service
    handler, err := api.NewHandler(dbService, cfg)
    if err != nil {
        log.Fatalf("Failed to create handler: %v", err)
    }
    handler.SetupRoutes(router)

    // Swagger UI route
    router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

    // Create HTTP server with timeouts
    srv := &http.Server{
        Addr:         cfg.ServerAddr,
        Handler:      router,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  30 * time.Second,
    }

    // Start server with graceful shutdown
    go func() {
        log.Printf("Starting server on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Handle graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, os.Kill)
    <-stop

    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server shutdown failed: %v", err)
    }
    log.Println("Server stopped gracefully")
}