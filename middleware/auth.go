package middleware

import "github.com/gin-gonic/gin"

// ResetPasswordHandler function
func ResetPasswordHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.URL.Query().Get("token")
		c.Redirect(302, "cancavas://resetpassword?token="+token)
		return
	}
}
