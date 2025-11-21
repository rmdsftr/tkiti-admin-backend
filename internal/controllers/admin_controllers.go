package controllers

import (
	"admin-panel/internal/models"
	"admin-panel/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB *gorm.DB
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{DB: db}
}

func (ac *AdminController) GetAllAdmin(c *gin.Context) {
	type AdminWithCount struct {
		Nim        string `json:"nim"`
		Nama       string `json:"nama"`
		Role       string `json:"role"`
		Status     string `json:"status"`
		JmlArtikel int64  `json:"jumlah_artikel"`
	}

	var result []AdminWithCount

	err := ac.DB.Table("admin").
		Select("admin.nim, admin.nama, admin.role, admin.status, COUNT(article.article_id) AS jml_artikel").
		Joins("LEFT JOIN article ON article.nim = admin.nim").
		Group("admin.nim, admin.nama, admin.role, admin.status").
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil semua data admin",
		"data":    result,
	})
}

func (ac *AdminController) CreateAdmin(c *gin.Context) {
	var input struct {
		Nim       string `json:"nim" binding:"required"`
		Nama      string `json:"nama" binding:"required"`
		NoAslab   string `json:"no_aslab" binding:"required"`
		Pword     string `json:"pword" binding:"required"`
		Deskripsi string `json:"deskripsi"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var existing models.Admin
	if err := ac.DB.First(&existing, "nim = ?", input.Nim).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Admin dengan NIM tersebut sudah ada",
		})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Pword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal melakukan hashing password",
		})
		return
	}

	admin := models.Admin{
		Nim:       input.Nim,
		Nama:      utils.ToTitleCase(input.Nama),
		NoAslab:   strings.ToUpper(input.NoAslab),
		Pword:     hashedPassword,
		Deskripsi: input.Deskripsi,
	}

	if err := ac.DB.Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat data admin",
		"data":    admin,
	})
}

func (ac *AdminController) UpdateRoleOrStatus(c *gin.Context) {
	nim := c.Param("nim")
	val := c.Param("roleorstatus")

	var admin models.Admin

	if err := ac.DB.First(&admin, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Admin tidak ditemukan",
		})
		return
	}

	if val == string(models.RoleAdmin) || val == string(models.RoleMember) {
		ac.DB.Model(&admin).Update("role", models.RoleEnum(val))

		c.JSON(http.StatusOK, gin.H{
			"message": "Role berhasil diperbarui",
			"data":    admin,
		})
		return
	}

	if val == string(models.StatusAktif) || val == string(models.StatusNonAktif) {
		ac.DB.Model(&admin).Update("status", models.StatusEnum(val))

		c.JSON(http.StatusOK, gin.H{
			"message": "Status berhasil diperbarui",
			"data":    admin,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Value tidak valid, gunakan: admin/member atau aktif/nonaktif",
	})
}

func (ac *AdminController) DeleteAccount(c *gin.Context) {
	nim := c.Param("nim")
	var admin models.Admin

	if err := ac.DB.First(&admin, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	if err := ac.DB.Delete(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus akun user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus akun user",
	})
}
