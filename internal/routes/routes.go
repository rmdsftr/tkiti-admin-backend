package routes

import (
	"admin-panel/internal/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	users := r.Group("/users")
	{
		userController := controllers.NewUserController(db)
		users.GET("/", userController.GetAllUsers)
		users.POST("/", userController.CreateUsers)
	}
}
