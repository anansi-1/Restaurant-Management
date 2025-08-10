package controllers

import (
	"github.com/gin-gonic/gin"
)


	// incomingRoutes.GET("/invoices",controller.GetInvoices())
	// incomingRoutes.GET("/invoices/:invoice_id",controller.GetInvoice())
	// incomingRoutes.POST("/invoices",controller.CreateInvoice())
	// incomingRoutes.PATCH("/invoices/:invoice_id",controller.UpdateInvoice())

func GetInvoices() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}