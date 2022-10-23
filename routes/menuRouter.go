package routes

import (
	"github.com/PranavMasekar/restaurant-management/controllers"
	"github.com/gin-gonic/gin"
)

func MenuRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/menus", controllers.GetMenues())
	incomingRoutes.GET("/menus/:menu_id", controllers.GetMenu())
	incomingRoutes.POST("/menus/:user_id", controllers.CreateMenu())
	incomingRoutes.PATCH("/menus/:user_id/:menu_id", controllers.UpdateMenu())
	incomingRoutes.DELETE("/menus/:user_id/:menu_id", controllers.DeleteMenu())
}
