package response

import "github.com/gin-gonic/gin"

// Success sends a standardized JSON success response.
func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
	})
}

// Paginated sends a standardized paginated JSON response.
func Paginated(c *gin.Context, statusCode int, data any, pagination any) {
	c.JSON(statusCode, gin.H{
		"success":    true,
		"data":       data,
		"pagination": pagination,
	})
}
