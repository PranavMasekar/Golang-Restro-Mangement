package routes

import (
	"github.com/PranavMasekar/restaurant-management/controllers"
	"github.com/gin-gonic/gin"
)

func InvoiceRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/invoices", controllers.GetInvoices())
	incomingRoutes.GET("/invoices/:invoice_id", controllers.GetInvoice())
	incomingRoutes.POST("/invoices/:user_id", controllers.CreateInvoice())
	incomingRoutes.PATCH("/invoices/:user_id/:invoice_id", controllers.UpdateInvoice())
	incomingRoutes.DELETE("/invoices/:user_id/:invoice_id", controllers.DeleteInvoice())
}
