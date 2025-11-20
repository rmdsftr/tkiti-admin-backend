package controllers

import (
	"admin-panel/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

func (uc *UserController) GetAllUsers(c *gin.Context) {
	var users []models.User

	if err := uc.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mendapatkan data user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mendapatkan data users",
		"data":    users,
	})
}

func (uc *UserController) CreateUsers(c *gin.Context) {
	var input struct {
		User_id string `json:"user_id" binding:"required"`
		Nama    string `json:"nama" binding:"required"`
		Lab     string `json:"lab" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user := models.User{
		UserID: input.User_id,
		Nama:   input.Nama,
		Lab:    input.Lab,
	}

	if err := uc.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menambahkan user",
		"data":    user,
	})
}
