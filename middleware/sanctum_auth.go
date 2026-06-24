package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/yimm/rfid-api/helpers"
	"github.com/yimm/rfid-api/models"

	"github.com/gin-gonic/gin"
)

func SanctumAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		token := parts[1]
		tokenParts := strings.Split(token, "|")
		if len(tokenParts) != 2 {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		tokenID, err := strconv.ParseInt(tokenParts[0], 10, 64)
		if err != nil {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		plainToken := tokenParts[1]

		sanctumToken, err := models.ValidateToken(tokenID, plainToken)
		if err != nil {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		if sanctumToken.TokenableType != "App\\Models\\User" {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		user, err := models.GetUserByID(sanctumToken.TokenableID)
		if err != nil {
			helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}

		role := models.GetUserRoleSimple(user.ID)
		user.Role = role

		c.Set("user", user)
		c.Set("token_id", tokenID)
		c.Set("token", sanctumToken)

		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				helpers.ErrorResponse(c, http.StatusInternalServerError, "Internal server error.")
				c.Abort()
			}
		}()
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
