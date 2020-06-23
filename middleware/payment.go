package middleware

import (
	"github.com/eaemenkkstudios/cancanvas-backend/repository"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"github.com/gin-gonic/gin"
)

var orderRepository = repository.NewOrderRepository()

// ResultHandler function
func ResultHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.URL.Query().Get("token")
		payerID := c.Request.URL.Query().Get("PayerID")
		paypalClient := service.NewPayPalService()
		status, err := paypalClient.CompleteOrder(token)
		if err == nil {
			orderRepository.UpdateOrder(token, status, &payerID)
		}
		c.Redirect(302, "cancavas://order?id="+token)
		return
	}
}

// CancelHandler function
func CancelHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.URL.Query().Get("token")
		paypalClient := service.NewPayPalService()
		status, err := paypalClient.GetOrderStatus(token)
		if err == nil && status != "COMPLETED" {
			orderRepository.DeleteOrder(token)
		}
		c.Redirect(302, "cancavas://order?id="+token)
		return
	}
}
