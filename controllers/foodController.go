package controllers

// incomingRoutes.GET("/foods",controller.GetFoods())
// incomingRoutes.GET("/foods/:food_id",controller.GetFood())
// incomingRoutes.POST("/foods",controller.CreateFood())
// incomingRoutes.PATCH("/foods/:food_id",controller.UpdateFood())

import (
	"github.com/gin-gonic/gin"
)


func GetFoods() gin.HandlerFunc{

	return func(ctx *gin.Context) {
		
	}
}

func GetFood() gin.HandlerFunc{

	return func(ctx *gin.Context) {

	}
}
func CreateFood() gin.HandlerFunc{

	return func(ctx *gin.Context) {

	}
}
func UpdateFood() gin.HandlerFunc{

	return func(ctx *gin.Context) {

	}
}