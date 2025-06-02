package main

import (
	"fmt"
	"log"
	"net/http"

	"Dowlet_projects/ecommerce/api"
	_ "Dowlet_projects/ecommerce/docs" // Swagger docs
	"Dowlet_projects/ecommerce/services"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title E-commerce API
// @version 1.0
// @description API for an e-commerce platform with OTP-based authentication and product/market management.
// @host 192.168.55.42:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication (e.g., "Bearer <token>")
func main() {
	dbService, err := services.NewDBService("root", "", "ecommerce_db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbService.Close()

	router := mux.NewRouter()
	handler := api.NewHandler(dbService)
	handler.SetupRoutes(router)

	// Swagger UI route
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
