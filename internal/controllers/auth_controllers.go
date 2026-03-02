package controllers

import (
	"admin-panel/internal/config"
	"admin-panel/internal/models"
	"admin-panel/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewAuthController(db *gorm.DB, cfg *config.Config) *AuthController {
	return &AuthController{
		DB:     db,
		Config: cfg,
	}
}

func (ac *AuthController) Login(c *gin.Context) {
	var input struct {
		Nim   string `json:"nim" binding:"required"`
		Pword string `json:"pword" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.Admin
	if err := ac.DB.First(&user, "nim = ?", input.Nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if user.Status != models.StatusAktif {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Akun anda sudah dinonaktifkan",
		})
		return
	}

	if !utils.CheckPassword(input.Pword, user.Pword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Password tidak sesuai",
		})
		return
	}

	token, err := utils.GenerateTokenPair(ac.Config, user.Nim, user.NoAslab, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	utils.SetTokenCookies(c, ac.Config, token)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"data": gin.H{
			"nim":      user.Nim,
			"nama":     user.Nama,
			"no_aslab": user.NoAslab,
			"role":     user.Role,
			"photo_url" : user.PhotoUrl,
		},
	})
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	refreshToken, err := utils.GetTokenFromCookie(c, "refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Refresh token not found",
		})
		return
	}

	claims, err := utils.ValidateToken(refreshToken, ac.Config.JWT.RefreshSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid refresh token",
		})
		return
	}

	token, err := utils.GenerateTokenPair(ac.Config, claims.Nim, claims.NoAslab, claims.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to generate tokens",
		})
		return
	}

	utils.SetTokenCookies(c, ac.Config, token)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Token berhasil diperbarui",
	})
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	nim := c.Param("nim")

	var input struct {
		PasswordLama    string `json:"pword_lama" binding:"required"`
		PasswordBaru    string `json:"pword_baru" binding:"required,min=6"`
		PasswordConfirm string `json:"pword_confirm" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.Admin
	if err := ac.DB.First(&user, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	if !utils.CheckPassword(input.PasswordLama, user.Pword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password lama tidak sesuai",
		})
		return
	}

	if input.PasswordBaru != input.PasswordConfirm {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password baru tidak cocok dengan konfirmasi",
		})
		return
	}

	hashedPassword, err := utils.HashPassword(input.PasswordBaru)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal meng-hash password",
		})
		return
	}

	if err := ac.DB.Model(&user).Update("pword", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memperbarui password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password berhasil diperbarui",
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	utils.ClearTokenCookie(c, ac.Config)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout successful",
	})
}
