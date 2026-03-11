package main

import (
	"log"

	"SureMFService/config"
	"SureMFService/database/cloudsql"
	"SureMFService/database/firebase"
	"SureMFService/middleware"
	"SureMFService/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Init()

	port := config.AppConfig.Port
	log.Printf("Starting SureMFService on port %s", port)

	if err := firebase.Init(); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	defer firebase.Close()

	if err := cloudsql.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.AuditLogMiddleware())
	routes.Routes(router)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
