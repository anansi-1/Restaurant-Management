package main

import (
	"log"
	"os"
	"restaurant-management/database"
	"restaurant-management/middleware"
	"restaurant-management/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client,"food")

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Could not load .env file")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	
	
	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	
	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)

}
