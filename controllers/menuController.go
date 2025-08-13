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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

// GET /menus
func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := menuCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error fetching menu items",
			})
			return
		}

		defer cursor.Close(ctx)

		var menus []bson.M

		if err = cursor.All(ctx, &menus); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error decoding menu items",
			})
			return
		}

		c.JSON(http.StatusOK, menus)
	}
}

// GET /menus/:menu_id
func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		menuID := c.Param("menu_id")

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuID}).Decode(&menu)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "menu not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the menu item",
			})
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

// POST /menus
func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		menu.ID = primitive.NewObjectID()
		menu.MenuID = menu.ID.Hex()
		menu.CreatedAt = time.Now().UTC()
		menu.UpdatedAt = time.Now().UTC()

		result, insertErr := menuCollection.InsertOne(ctx, menu)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error inserting menu",
			})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// PATCH /menus/:menu_id
func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		menuID := c.Param("menu_id")
		filter := bson.M{"menu_id": menuID}

		if menu.StartDate != nil && menu.EndDate != nil {
			if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time span, please re-enter the dates"})
				return
			}
		}

		updateObj := bson.D{
			{Key: "updated_at", Value: time.Now().UTC()},
		}

		if menu.StartDate != nil {
			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.StartDate})
		}
		if menu.EndDate != nil {
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.EndDate})
		}
		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
		}
		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
		}

		upsert := true
		opts := options.Update().SetUpsert(upsert)

		result, err := menuCollection.UpdateOne(
			ctx,
			filter,
			bson.D{{Key: "$set", Value: updateObj}},
			opts,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
