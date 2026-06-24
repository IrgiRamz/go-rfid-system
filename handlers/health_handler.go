package handlers

import (
	"net/http"

	"github.com/yimm/rfid-api/helpers"
	"github.com/yimm/rfid-api/models"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	helpers.HealthResponse(c, http.StatusOK, "Go API is running")
}

func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
		return nil, false
	}
	return user.(*models.User), true
}