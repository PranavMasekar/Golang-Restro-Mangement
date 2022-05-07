package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PranavMasekar/restaurant-management/database"
	"github.com/PranavMasekar/restaurant-management/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(ctx.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(ctx.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		projectStage := bson.D{{"$project", bson.D{{"_id", 0}, {"total_count", 1}, {"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}

		result, err := userCollection.Aggregate(c, mongo.Pipeline{matchStage, projectStage})

		defer cancel()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}

		var allUsers []bson.M

		if err = result.All(c, &allUsers); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allUsers)
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		userId := ctx.Param("user_id")
		var user models.User
		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

}
