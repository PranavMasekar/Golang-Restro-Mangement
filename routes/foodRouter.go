package routes

import (
	"github.com/PranavMasekar/restaurant-management/controllers"
	"github.com/gin-gonic/gin"
)

func FoodRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/foods", controllers.GetFoods())
	incomingRoutes.GET("/foods/:food_id", controllers.GetFood())
	incomingRoutes.POST("/foods/:user_id", controllers.CreateFood())
	incomingRoutes.PATCH("/foods/:user_id/:food_id", controllers.UpdateFood())
	incomingRoutes.DELETE("/foods/:user_id/:food_id", controllers.DeleteFood())
}
