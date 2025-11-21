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

type KegiatanController struct {
	DB             *gorm.DB
	StorageService *services.StorageService
}

func NewKegiatanController(db *gorm.DB, storageService *services.StorageService) *KegiatanController {
	return &KegiatanController{
		DB:             db,
		StorageService: storageService,
	}
}

func (kc *KegiatanController) CreateKegiatan(c *gin.Context) {
	var input struct {
		Judul     string `form:"judul" binding:"required"`
		Deskripsi string `form:"deskripsi" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("photo_url")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file foto wajib diupload"})
		return
	}

	if !middleware.IsValidImageType(file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipe file tidak valid"})
		return
	}

	if file.Size > 25*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran gambar terlalu besar"})
		return
	}

	photourl, err := kc.StorageService.UploadFile(file, "kegiatan")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	event := models.Kegiatan{
		Judul:     input.Judul,
		Deskripsi: input.Deskripsi,
		PhotoUrl:  photourl,
	}

	if err := kc.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan kegiatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berhasil", "data": event})
}

func (kc *KegiatanController) GetAllKegiatan(c *gin.Context) {
	var events []models.Kegiatan

	if err := kc.DB.Find(&events).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
	})
}

func (kc *KegiatanController) DeleteKegiatan(c *gin.Context) {
	id := c.Param("kegiatan_id")
	var event models.Kegiatan
	if err := kc.DB.First(&event, "kegiatan_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "kegiatan tidak ditemukan"})
		return
	}

	if event.PhotoUrl != "" {
		filePath := kc.StorageService.ExtractFilePathFromURL(event.PhotoUrl)
		if filePath != "" {
			if err := kc.StorageService.DeleteFile(filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("gagal menghapus file: %v", err),
				})
				return
			}
		}
	}

	if err := kc.DB.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus kegiatan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "berhasil menghapus kegiatan"})
}
