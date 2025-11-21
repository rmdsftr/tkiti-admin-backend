package routes

import (
	"admin-panel/internal/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	admin := r.Group("/admin")
	{
		adminController := controllers.NewAdminController(db)
		admin.GET("/", adminController.GetAllAdmin)
		admin.POST("/", adminController.CreateAdmin)
		admin.PATCH("/:nim/:roleorstatus", adminController.UpdateRoleOrStatus)
		admin.DELETE("/:nim", adminController.DeleteAccount)
	}
}
