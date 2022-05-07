package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PranavMasekar/restaurant-management/database"
	"github.com/PranavMasekar/restaurant-management/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	order_id         string
	Payment_status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoices items"})
		}

		var allInvoices []bson.M

		if err = result.All(c, &allInvoices); err != nil {
			log.Fatal(err)
		}

		ctx.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		invoiceId := ctx.Param("invoice_id")
		var invoice models.Invoice
		err := invoiceCollection.FindOne(c, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "errror occured while fetching Invoice item"})
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_id)
		invoiceView.order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date

		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}

		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = *&invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		ctx.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice
		err := ctx.BindJSON(&invoice)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		var order models.Order
		err = orderCollection.FindOne(c, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		defer cancel()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Order Not found"})
			return
		}
		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(c, invoice)
		if insertErr != nil {
			msg := fmt.Sprintf("invoice item was not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		var invoice models.Invoice
		invoiceid := ctx.Param("invoice_id")

		err := ctx.BindJSON(&invoice)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		filter := bson.M{"invoice_id": invoiceid}
		var updateObj primitive.D

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_method})
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_status})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", invoice.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		// Setting default payment status as PENDING
		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		result, err := invoiceCollection.UpdateOne(
			c,
			filter,
			bson.D{

				{"$set", updateObj},
			},
			&opt,
		)
		if err != nil {
			msg := fmt.Sprintf("invoice item update failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		ctx.JSON(http.StatusOK, result)
	}
}
