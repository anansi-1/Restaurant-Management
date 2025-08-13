package controllers

import (
	"context"
	"net/http"
	"restaurant-management/database"
	"restaurant-management/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

var validate = validator.New()

// GET /foods
func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		if q := c.Query("startIndex"); q != "" {
			if si, err := strconv.Atoi(q); err == nil && si >= 0 {
				startIndex = si
			}
		}

		matchStage := bson.D{
			{Key: "$match", Value: bson.D{}},
		}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total_count", Value: bson.D{
					{Key: "$sum", Value: 1},
				}},
				{Key: "data", Value: bson.D{
					{Key: "$push", Value: "$$ROOT"},
				}},
			}},
		}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "food_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}

		var results []bson.M
		cursor, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error fetching food items",
			})
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error decoding food items",
			})
			return
		}

		if len(results) > 0 {
			c.JSON(http.StatusOK, results[0])
		} else {
			c.JSON(http.StatusOK, gin.H{
				"total_count": 0,
				"food_items":  []interface{}{},
			})
		}
	}
}

// GET /foods/:food_id
func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		foodID := c.Param("food_id")

		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Food item not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the food item",
			})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

// POST /foods
func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var food models.Food
		var menu models.Menu

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "menu not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the food item",
			})
			return
		}

		food.ID = primitive.NewObjectID()
		food.CreatedAt = time.Now().UTC()
		food.UpdatedAt = time.Now().UTC()
		food.FoodID = food.ID.Hex()

		num := toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error inserting food item",
			})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// PATCH /foods/:food_id
func UpdateFood() gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		var food models.Food

		foodID := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request" + err.Error(),
			})
			return
		}

		updateObj := bson.D{{Key: "updated_at", Value: time.Now().UTC()}}

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: *food.Name})
		}
		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: *food.Price})
		}
		if food.FoodImage != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: *food.FoodImage})
		}
		if food.MenuID != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": *food.MenuID}).Decode(&menu)

			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{
						"error": "Menu item not found",
					})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error occurred while fetching the menu item",
				})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: *food.MenuID})
		}

		upsert := true
		opts := options.Update().SetUpsert(upsert)

		filter := bson.M{"food_id": foodID}

		result, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{{Key: "$set", Value: updateObj}},
			opts,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update food item: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

}
