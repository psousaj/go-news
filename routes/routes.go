package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine) {
	r.GET("/hello", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
            c.String(http.StatusBadRequest, "Name is required")
            return
        }

		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
            "name":    c.Query("name"),
		})
	})
}