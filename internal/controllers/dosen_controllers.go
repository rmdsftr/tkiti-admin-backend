package controllers

import (
	"admin-panel/internal/middleware"
	"admin-panel/internal/models"
	"admin-panel/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DosenController struct {
	DB             *gorm.DB
	StorageService *services.StorageService
}

func NewDosenController(db *gorm.DB, storageService *services.StorageService) *DosenController {
	return &DosenController{
		DB:             db,
		StorageService: storageService,
	}
}

func (dc *DosenController) GetAllDosen(c *gin.Context) {
	var dosens []models.Dosen
	if err := dc.DB.Find(&dosens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data dosen"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dosens})
}

func (dc *DosenController) CreateDosen(c *gin.Context) {
	var input struct {
		NIP       string            `form:"nip" binding:"required"`
		NamaDosen string            `form:"nama_dosen" binding:"required"`
		Posisi    models.PosisiEnum `form:"posisi" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("foto")
	var photourl string
	if err == nil {
		if !middleware.IsValidImageType(file.Header.Get("Content-Type")) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipe file tidak valid"})
			return
		}

		if file.Size > 25*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran gambar terlalu besar"})
			return
		}

		photourl, err = dc.StorageService.UploadFile(file, "dosen")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	dosen := models.Dosen{
		NIP:       input.NIP,
		NamaDosen: input.NamaDosen,
		Foto:      photourl,
		Posisi:    input.Posisi,
	}

	if err := dc.DB.Create(&dosen).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gagal menyimpan dosen: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berhasil menambahkan dosen", "data": dosen})
}

func (dc *DosenController) DeleteDosen(c *gin.Context) {
	nip := c.Param("nip")
	var dosen models.Dosen
	if err := dc.DB.First(&dosen, "nip = ?", nip).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dosen tidak ditemukan"})
		return
	}

	if dosen.Foto != "" {
		filePath := dc.StorageService.ExtractFilePathFromURL(dosen.Foto)
		if filePath != "" {
			if err := dc.StorageService.DeleteFile(filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("gagal menghapus file: %v", err),
				})
				return
			}
		}
	}

	if err := dc.DB.Where("nip = ?", nip).Delete(&models.Dosen{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gagal menghapus dosen: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berhasil menghapus dosen"})
}
