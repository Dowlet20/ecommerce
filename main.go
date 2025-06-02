package main

import (
	"log"
	"net/http"

	"github.com/Dowlet-projects/ecommerce/api"
	"github.com/Dowlet-projects/ecommerce/services"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	dbService, err := services.NewDBService("your_mysql_user", "your_mysql_password", "ecommerce_db")
	if err != nil {
		log.Fatal(err)
	}
	defer dbService.Close()

	// Initialize router
	router := mux.NewRouter()

	// Initialize API handlers
	apiHandler := api.NewHandler(dbService)

	// Setup routes
	apiHandler.SetupRoutes(router)

	// Serve uploaded files
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Start server
	log.Fatal(http.ListenAndServe(":8080", router))
}
