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

type InvoiceViewFormat struct {
	InvoiceID      string
	OrderID        string
	PaymentMethod  *string
	PaymentStatus  *string
	TableNumber    interface{}
	PaymentDue     interface{}
	PaymentDueDate time.Time
	OrderDetails   interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

// GET /invoices
func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := invoiceCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch invoices: " + err.Error(),
			})
			return
		}
		defer cursor.Close(ctx)

		var invoices []bson.M
		if err := cursor.All(ctx, &invoices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to decode invoices: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, invoices)
	}
}

// GET /invoices/:invoice_id
func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceID := c.Param("invoice_id")

		var invoice models.Invoice
		if err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceID}).Decode(&invoice); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"message": "invoice item not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invoice"})
			return
		}

		allOrderItems, err := ItemsByOrder(invoice.OrderID)
		if err != nil || len(allOrderItems) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch related order items"})
			return
		}

		invoiceView := InvoiceViewFormat{
			InvoiceID:      invoice.InvoiceID,
			OrderID:        invoice.OrderID,
			PaymentDueDate: invoice.PaymentDueDate,
			PaymentMethod:  invoice.PaymentMethod,
			PaymentStatus:  invoice.PaymentStatus,
			PaymentDue:     allOrderItems[0]["payment_due"],
			TableNumber:    allOrderItems[0]["table_number"],
			OrderDetails:   allOrderItems[0]["order_items"],
		}

		c.JSON(http.StatusOK, invoiceView)
	}
}

// POST /invoices
func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		var order models.Order

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request body"})
			return
		}

		if err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.OrderID}).Decode(&order); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "order not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while fetching the food item",
			})

			return
		}

		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}

		invoice.PaymentDueDate = time.Now().UTC().Add(24 * time.Hour)
		invoice.ID = primitive.NewObjectID()
		invoice.InvoiceID = invoice.ID.Hex()
		invoice.CreatedAt = time.Now().UTC()
		invoice.UpdatedAt = time.Now().UTC()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(ctx, invoice)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create invoice",
			})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// PATCH /invoices/:invoice_id
func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceID := c.Param("invoice_id")

		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request: " + err.Error(),
			})
			return
		}

		updateObj := bson.D{{Key: "updated_at", Value: time.Now().UTC()}}

		if invoice.OrderID != "" {
			var order models.Order
			if err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.OrderID}).Decode(&order); err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{"error": "order item not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while fetching the order item"})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "order_id", Value: invoice.OrderID})
		}

		if invoice.PaymentMethod != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: *invoice.PaymentMethod})
		}
		if invoice.PaymentStatus != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: *invoice.PaymentStatus})
		}
		if !invoice.PaymentDueDate.IsZero() {
			updateObj = append(updateObj, bson.E{Key: "payment_due_date", Value: invoice.PaymentDueDate})
		}

		filter := bson.M{"invoice_id": invoiceID}
		existing := invoiceCollection.FindOne(ctx, filter)
		if existing.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		} else if existing.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invoice"})
			return
		}

		_, err := invoiceCollection.UpdateOne(ctx, filter, bson.M{"$set": updateObj})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice: " + err.Error()})
			return
		}

		var updatedInvoice models.Invoice
		if err := invoiceCollection.FindOne(ctx, filter).Decode(&updatedInvoice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated invoice"})
			return
		}

		c.JSON(http.StatusOK, updatedInvoice)
	}
}
