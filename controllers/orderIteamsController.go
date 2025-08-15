package controllers

import (
	"context"
	"log"
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

type OrderItemPack struct {
	TableID    *string
	OrderItems []models.OrderItem
}


var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

// GET /orderItems
func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := orderItemCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error fetching orderitems",
			})
		}
		defer cursor.Close(ctx)

		var orderItems []models.OrderItem

		if err = cursor.All(ctx, &orderItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "error decoding orderitems",
			})
			return
		}

		c.JSON(http.StatusOK, orderItems)

	}
}

// GET /orderItems-order/:order_id
func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		orderID := c.Param("order_id")
		allOrderItems, err := ItemsByOrder(orderID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order ID"})
			return
		}
		c.JSON(http.StatusOK, allOrderItems)

	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {

	c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}

	lookupStage := bson.D{{
		Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "food"},
			{Key: "localField", Value: "food_id"},
			{Key: "foreignField", Value: "food_id"},
			{Key: "as", Value: "food"},
		},
	}}

	unwindStage := bson.D{
		{
			Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$food"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			},
		},
	}

	lookupOrderStage := bson.D{{
		Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "order"},
			{Key: "localField", Value: "order_id"},
			{Key: "foreignField", Value: "order_id"},
			{Key: "as", Value: "order"},
		}},
	}

	unwindOrderStage := bson.D{{
		Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$order"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		},
	}}

	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "table"},
		{Key: "localField", Value: "order.table_id"},
		{Key: "foreignField", Value: "table_id"},
		{Key: "as", Value: "table"},
	}}}
	unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$table"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "id", Value: 0},
		{Key: "amount", Value: "$food.price"},
		{Key: "total_count", Value: 1},
		{Key: "food_name", Value: "$food.name"},
		{Key: "food_image", Value: "$food.food_image"},
		{Key: "table_number", Value: "$table.table_number"},
		{Key: "table_id", Value: "$table.table_id"},
		{Key: "order_id", Value: "$order.order_id"},
		{Key: "price", Value: "$food.price"},
		{Key: "quantity", Value: 1},
	}}}

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "order_id", Value: "$order_id"},
			{Key: "table_id", Value: "$table_id"},
			{Key: "table_number", Value: "$table_number"},
		}},
		{Key: "payment_due", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
		{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
		{Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
	}}}

	projectStage2 := bson.D{{Key: "$project", Value: bson.D{
		{Key: "id", Value: 0},
		{Key: "payment_due", Value: 1},
		{Key: `total_count`, Value: 1},
		{Key: "table_number", Value: "$_id.table_number"},
		{Key: "order_items", Value: 1},
	}}}

	result, err := orderItemCollection.Aggregate(c, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	})

	if err = result.All(c, &OrderItems); err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	if err = result.All(c, &OrderItems); err != nil {
		panic(err)
	}

	defer cancel()

	return OrderItems, err

}

// GET /orderItems/:orderItem_id
func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		OrderItemID := c.Param("orderItem_id")
		var OrderItem models.OrderItem
		err := orderItemCollection.FindOne(ctx, bson.M{"orderItem_id": OrderItemID}).Decode(&OrderItem)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "orderItem not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the menu item",
			})
			return
		}

		c.JSON(http.StatusOK, OrderItem)

	}
}

// POST /orderItems
func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var orderItemPack OrderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.OrderDate, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		orderItemsToBeInserted := []interface{}{}
		order.TableID = orderItemPack.TableID
		order_id, _ := OrderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.OrderItems {
			orderItem.OrderID = order_id

			validationErr := validate.Struct(orderItem)

			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}
			orderItem.ID = primitive.NewObjectID()
			orderItem.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.OrderItemID = orderItem.ID.Hex()
			var num = toFixed(*orderItem.UnitPrice, 2)
			orderItem.UnitPrice = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)

		if err != nil {
			log.Fatal(err)
		}
		defer cancel()

		c.JSON(http.StatusOK, insertedOrderItems)
	}
}
// PATCH /orderItems/:orderItem_id
func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItem models.OrderItem

		orderItemId := c.Param("order_item_id")

		filter := bson.M{"order_item_id": orderItemId}

		var updateObj primitive.D

		if orderItem.UnitPrice != nil {
			updateObj = append(updateObj, bson.E{Key: "unit_price", Value: orderItem.UnitPrice})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{Key: "quantity", Value: *orderItem.Quantity})
		}

		if orderItem.FoodID != nil {
			updateObj = append(updateObj, bson.E{Key: "food_id", Value: *orderItem.FoodID})
		}

		orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: orderItem.UpdatedAt})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Order item update failed"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()

		c.JSON(http.StatusOK, result)
	}
}
