package routes

import (
	"admin-panel/internal/controllers"
	"admin-panel/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, storageService *services.StorageService) {
	admin := r.Group("/admin")
	{
		adminController := controllers.NewAdminController(db)
		admin.GET("", adminController.GetAllAdmin)
		admin.POST("", adminController.CreateAdmin)
		admin.PATCH("/:nim/:roleorstatus", adminController.UpdateRoleOrStatus)
		admin.DELETE("/:nim", adminController.DeleteAccount)
	}

	kegiatan := r.Group("/kegiatan")
	{
		kegiatanController := controllers.NewKegiatanController(db, storageService)
		kegiatan.POST("", kegiatanController.CreateKegiatan)
		kegiatan.GET("", kegiatanController.GetAllKegiatan)
		kegiatan.DELETE("/:kegiatan_id", kegiatanController.DeleteKegiatan)
	}
}
