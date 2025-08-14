package controllers

import (
	"context"
	"net/http"
	"restaurant-management/database"
	"restaurant-management/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var orderCollection = database.OpenCollection(database.Client, "order")

// GET  /orders
func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := orderCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching order items"})
			return
		}
		defer cursor.Close(ctx)

		var orders []bson.M

		err = cursor.All(ctx, &orders)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error decoding order items"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"error": ""})

	}
}

// GET  /orders/:order_id
func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderID := c.Param("order_id")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required"})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&order)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Order item not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the order item",
			})
			return
		}

		c.JSON(http.StatusOK, order)

	}
}

// POST /orders
func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var order models.Order
		var table models.Table
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		if err := validate.Struct(order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.TableID}).Decode(&table)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "table not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the table",
			})
			return
		}

		order.ID = primitive.NewObjectID()
		order.OrderID = order.ID.Hex()
		order.CreatedAt = time.Now().UTC()
		order.UpdatedAt = time.Now().UTC()

		result, err := orderCollection.InsertOne(ctx, order)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error inserting order",
			})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}
                 
// PATCH orders/:order_id"
func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var order models.Order
		var table models.Table

		orderID := c.Param("order_id")

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request" + err.Error(),
			})
			return
		}

		updateObj := bson.D{{Key: "updated_at", Value: time.Now().UTC()}}

		if !order.OrderDate.IsZero() {
			updateObj = append(updateObj, bson.E{Key: "order_date", Value: order.OrderDate})
		}

		if order.TableID != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": *order.TableID}).Decode(&table)

			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "table not found",
					})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error occurred while fetching the table",
				})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.TableID})

		}

		result, err := orderCollection.UpdateOne(
			ctx,
			bson.M{"order_id": orderID},
			bson.D{{Key: "$set", Value: updateObj}})
		if err != nil {

			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error":"Failed to update order" + err.Error()},
			)
			return
		}     

		c.JSON(http.StatusOK,result)
	}
}


