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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")
var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)

func GetOrders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items"})
		}
		var allOrders []bson.M
		if err = result.All(c, &allOrders); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allOrders)
	}
}
func GetOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		orderId := ctx.Param("order_id")

		var order models.Order

		err := orderCollection.FindOne(c, bson.M{"order_id": orderId}).Decode(&order)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "errror occured while fetching Order"})
		}
		var orderResponse models.OrderResponse

		orderResponse.ID = order.ID
		orderResponse.Order_id = order.Order_id

		var totalAmount float64
		for _, foodId := range order.Food_Items {
			var food models.Food
			err := foodCollection.FindOne(c, bson.M{"food_id": foodId}).Decode(&food)
			defer cancel()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "errror occured while fetching food item"})
				return
			}
			totalAmount += *food.Price
			orderResponse.Food_Items = append(orderResponse.Food_Items, food)
		}
		orderResponse.Amount = totalAmount
		ctx.JSON(http.StatusOK, orderResponse)
	}
}
func CreateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Checking request is from Manager
		userId := ctx.Param("user_id")
		var user models.User
		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if *user.Type != "MANAGER" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Not Authorized"})
			return
		}

		var order models.Order
		var table models.Table

		err = ctx.BindJSON(&order)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		validationError := validate.Struct(order)

		if validationError != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}

		if order.Table_id != nil {
			err := tableCollection.FindOne(c, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Table was not found")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()

		result, insertError := orderCollection.InsertOne(c, order)

		if insertError != nil {
			msg := fmt.Sprintf("Food item was not inserted")
			ctx.JSON(http.StatusInternalServerError, gin.H{"err": msg})
			return
		}
		defer cancel()
		ctx.JSON(http.StatusOK, result)

	}
}
func UpdateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var table models.Table
		var order models.Order
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Checking request is from Manager
		userId := ctx.Param("user_id")
		var user models.User
		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if *user.Type != "MANAGER" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Not Authorized"})
			return
		}

		var updateObj primitive.D

		orderId := ctx.Param("order_id")

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.Table_id != nil {
			err := orderCollection.FindOne(ctx, bson.M{"tabled_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Table was not found")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu", Value: order.Table_id})
		}

		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})

		upsert := true
		filter := bson.M{"order_id": orderId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			c,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			msg := fmt.Sprint("food item update failed")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Checking request is from Manager
		userId := ctx.Param("user_id")
		var user models.User
		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if *user.Type != "MANAGER" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Not Authorized"})
			return
		}

		orderId := ctx.Param("order_id")

		res, err := orderCollection.DeleteOne(c, bson.M{"order_id": orderId})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			ctx.JSON(http.StatusOK, res)
		}
	}
}

func OrderItemOrderCreator(order models.Order) string {
	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(c, order)
	defer cancel()
	return order.Order_id
}
