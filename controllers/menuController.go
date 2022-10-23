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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenues() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		// Get all Menus
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, bson.M{"error": "error while listening menu items"})
			return
		}
		var allMenus []bson.M
		// Convert menus to slice of menus
		if err = result.All(c, &allMenus); err != nil {
			log.Fatal(err)
		}
		// Return all menus
		ctx.JSON(http.StatusOK, allMenus)
	}

}
func GetMenu() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		menuId := ctx.Param("menu_id")
		var menu models.Menu
		// Find the Menu Item
		err := foodCollection.FindOne(c, bson.M{"menu_id": menuId}).Decode(&menu)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			ctx.JSON(http.StatusOK, menu)
		}
	}
}
func CreateMenu() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var menu models.Menu
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

		err = ctx.BindJSON(&menu)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := validate.Struct(menu)
		if validationError != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()

		result, insertError := menuCollection.InsertOne(c, menu)
		if insertError != nil {
			msg := fmt.Sprintf("Menu item was not Created")
			ctx.JSON(http.StatusInternalServerError, gin.H{"err": msg})
			return
		}
		defer cancel()
		// response
		ctx.JSON(http.StatusOK, result)

	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func UpdateMenu() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
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

		err = ctx.BindJSON(&menu)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menuId := ctx.Param("menu_id")
		filter := bson.M{"menu_id": menuId}

		var updateObj primitive.D

		if menu.Start_Date != nil && menu.End_Date != nil {
			if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
				msg := "kindly retype the time"
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
			}
			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
			}
			if menu.Category != "" {
				updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
			}

			menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{Key: "crated_at", Value: menu.Created_at})

			upsert := true

			opt := options.UpdateOptions{
				Upsert: &upsert,
			}
			result, err := menuCollection.UpdateOne(
				c,
				filter,
				bson.D{
					{Key: "$set", Value: updateObj},
				},
				&opt,
			)

			if err != nil {
				msg := "menu updation failed"
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			}
			defer cancel()
			ctx.JSON(http.StatusOK, result)
		}

	}
}

func DeleteMenu() gin.HandlerFunc {
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

		menuId := ctx.Param("menu_id")

		res, err := menuCollection.DeleteOne(c, bson.M{"menu_id": menuId})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			ctx.JSON(http.StatusOK, res)
		}
	}
}
