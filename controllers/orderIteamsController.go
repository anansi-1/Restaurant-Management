package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// incomingRoutes.GET("/orderItems",controller.GetOrderItems())
// incomingRoutes.GET("/orderItems/:orderItem_id",controller.GetOrderItem())
// incomingRoutes.GET("/orderItems-order/:order_id",controller.GetOrderItemsByOrder())
// incomingRoutes.POST("/orderItems",controller.CreateOrderItem())
// incomingRoutes.PATCH("/orderItems/:orderItem_id",controller.UpdateOrderItem())

func GetOrderItems() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M,error){
	
}

func GetOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}



func UpdateOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}