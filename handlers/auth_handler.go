package handlers

import (
	"fmt"
	"net/http"
	"net/mail"

	"github.com/yimm/rfid-api/helpers"
	"github.com/yimm/rfid-api/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	LoginID  string `json:"login_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  UserInfo `json:"user"`
	Token string   `json:"token"`
}

type UserInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Username/Email atau password salah.")
		return
	}

	var user *models.User
	var err error

	if isValidEmail(req.LoginID) {
		user, err = models.GetUserByEmail(req.LoginID)
	} else {
		user, err = models.GetUserByUsername(req.LoginID)
	}

	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Username/Email atau password salah.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Username/Email atau password salah.")
		return
	}

	if !user.IsActive {
		helpers.ErrorResponse(c, http.StatusForbidden, "Akun Anda tidak aktif. Hubungi administrator.")
		return
	}

	models.DeleteUserTokens(user.ID)

	tokenID, plainToken, err := models.CreateToken(user.ID, "mobile-app", nil)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat membuat token.")
		return
	}

	role := models.GetUserRoleSimple(user.ID)

	token := fmt.Sprintf("%d|%s", tokenID, plainToken)

	helpers.SuccessResponse(c, http.StatusOK, "Login berhasil.", LoginResponse{
		User: UserInfo{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
			Role:     role,
		},
		Token: token,
	})
}

func Logout(c *gin.Context) {
	tokenID, exists := c.Get("token_id")
	if !exists {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Unauthenticated.")
		return
	}

	if err := models.DeleteToken(tokenID.(int64)); err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat logout.")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Logout berhasil.", nil)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}