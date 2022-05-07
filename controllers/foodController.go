package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/PranavMasekar/restaurant-management/database"
	"github.com/PranavMasekar/restaurant-management/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(ctx.Query("startIndex"))
		matchStage := bson.D{{"$match", bson.D{{}}}}
		// The members which are added in group stage are inserted in data field
		groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					// 0 means dont send this feild to frontEnd and 1 means send to frontend
					{"_id", 0},
					{"total_count", 1},
					// $slice => specifies the number of elements to return to array fields by using startIndex and End index
					{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		result, err := foodCollection.Aggregate(c, mongo.Pipeline{matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing food items"})
		}
		var allFoods []bson.M
		if err = result.All(c, &allFoods); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allFoods[0])
	}
}

func GetFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		foodId := ctx.Param("food_id")
		var food models.Food
		// Find the Food Item
		err := foodCollection.FindOne(c, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "errror occured while fetching food item"})
		}

		ctx.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var food models.Food
		var menu models.Menu
		// Get the request body into struct Food
		err := ctx.BindJSON(&food)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate food struct
		validationError := validate.Struct(food)
		if validationError != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}
		// Check whether menu exits or not in DB
		err = menuCollection.FindOne(c, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()

		if err != nil {
			msg := fmt.Sprintf("Menu not found")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		// Set created and updated values
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// Create New Id for food
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()

		var num = toFixed(*food.Price, 2)
		food.Price = &num
		// Insert into DB
		result, insertError := foodCollection.InsertOne(c, food)

		if insertError != nil {
			msg := fmt.Sprintf("Food item was not inserted")
			ctx.JSON(http.StatusInternalServerError, gin.H{"err": msg})
			return
		}
		defer cancel()
		// response
		ctx.JSON(http.StatusOK, result)

	}
}

func UpdateFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		foodId := ctx.Param("food_id")

		if err := ctx.BindJSON(&food); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{"name", food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{"price", food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{"food_image", food.Food_image})
		}

		if food.Menu_id != nil {
			// Get Menu
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Menu was not found")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"menu", food.Price})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", food.Updated_at})

		upsert := true
		filter := bson.M{"food_id": foodId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			c,
			filter,
			bson.D{
				{"$set", updateObj},
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

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
