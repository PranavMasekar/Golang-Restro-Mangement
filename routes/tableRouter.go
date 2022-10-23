package routes

import (
	"github.com/PranavMasekar/restaurant-management/controllers"
	"github.com/gin-gonic/gin"
)

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/tables", controllers.GetTables())
	incomingRoutes.GET("/tables/:table_id", controllers.GetTable())
	incomingRoutes.POST("/tables/:user_id", controllers.CreateTable())
	incomingRoutes.POST("/tables/:user_id/:table_id", controllers.UpdateTable())
	incomingRoutes.DELETE("/tables/:user_id/:table_id", controllers.DeleteTable())
}
