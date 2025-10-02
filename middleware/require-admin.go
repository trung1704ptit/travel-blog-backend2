package middleware

import (
	"net/http"

	"app/models"

	"github.com/gin-gonic/gin"
)

func RequireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUser := ctx.MustGet("currentUser").(models.User)

		if currentUser.Role != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "fail",
				"message": "Access denied. Admin role required.",
			})
			return
		}

		ctx.Next()
	}
}
