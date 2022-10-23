package routes

import (
	"github.com/PranavMasekar/restaurant-management/controllers"
	"github.com/gin-gonic/gin"
)

func OrderRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/orders", controllers.GetOrders())
	incomingRoutes.GET("/orders/:order_id", controllers.GetOrder())
	incomingRoutes.POST("/orders/:user_id", controllers.CreateOrder())
	incomingRoutes.PATCH("/orders/:user_id/:order_id", controllers.UpdateOrder())
	incomingRoutes.DELETE("/orders/:user_id/:order_id", controllers.DeleteOrder())
}
