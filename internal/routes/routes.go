package routes

import (
	"admin-panel/internal/config"
	"admin-panel/internal/controllers"
	"admin-panel/internal/middleware"
	"admin-panel/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, storageService *services.StorageService, cfg *config.Config) {
	pm := middleware.NewPeriodeMiddleware(db)

	auth := r.Group("/auth")
	{
		authController := controllers.NewAuthController(db, cfg)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
		auth.POST("/logout", authController.Logout)
		auth.POST("/change-pw/:nim", authController.ChangePassword)
	}

	admin := r.Group("/admin")
	admin.Use(pm.GetPeriodeIdActive())
	{
		adminController := controllers.NewAdminController(db, storageService)
		admin.GET("", adminController.GetAllAdmin)
		admin.POST("", adminController.CreateAdmin)
		admin.PATCH("/:nim/:roleorstatus", adminController.UpdateRoleOrStatus)
		admin.DELETE("/:nim", adminController.DeleteAccount)
		admin.DELETE("/:nim/pengurus", adminController.DeletePengurus)
		admin.GET("/periode/:periode_id", adminController.GetAdminbyPeriode)
		admin.PATCH("/change-photo/:nim", adminController.ChangePhoto)
		admin.PATCH("/delete-photo/:nim", adminController.DeletePhoto)
		admin.GET("/photo/:nim", adminController.GetUserPhotoProfile)
		admin.GET("/profile-data/:nim", adminController.GetUserDataProfile)
		admin.POST("/old", adminController.CreateAdminFromOld)
		admin.GET("/list-out-periode", adminController.GetAslabNotInPeriode)
	}

	kegiatan := r.Group("/kegiatan")
	{
		kegiatanController := controllers.NewKegiatanController(db, storageService)
		kegiatan.POST("", kegiatanController.CreateKegiatan)
		kegiatan.GET("", kegiatanController.GetAllKegiatan)
		kegiatan.DELETE("/:kegiatan_id", kegiatanController.DeleteKegiatan)
	}

	artikel := r.Group("/artikel")
	{
		artikelController := controllers.NewArticleController(db, storageService)
		artikel.POST("/:status/:nim", artikelController.CreateArticle)
		artikel.GET("/:status", artikelController.GetArtikelWithFilter)
		artikel.DELETE("/:article_id", artikelController.DeleteArticle)
		artikel.GET("/aslab/:status/:nim", artikelController.GetArtikelByAslab)
		artikel.GET("/edit/:article_id", artikelController.GetArtikelById)
		artikel.PUT("/update/:article_id", artikelController.UpdateArticle)
	}

	periode := r.Group("/periode")
	{
		periodeController := controllers.NewPeriodeController(db)
		periode.POST("/new", periodeController.CreatePeriode)
		periode.GET("", periodeController.GetAllPeriode)
	}

	struktur := r.Group("/struktur")
	struktur.Use(pm.GetPeriodeIdActive())
	{
		strukturController := controllers.NewStrukturController(db)
		struktur.GET("", strukturController.GetCurrentStruktur)
		struktur.GET("/periode/:periode_id", strukturController.GetStrukturByPeriode)
	}

	repositori := r.Group("/repositori")
	{
		repositoriController := controllers.NewRepositoriController(db, storageService)
		repositori.GET("/kontributor", repositoriController.GetKontributors)
		repositori.POST("", repositoriController.SubmitNewRepositori)
		repositori.GET("", repositoriController.GetAllRepositori)
		repositori.GET("/periode/:periode_id", repositoriController.GetAllRepositoriByPeriode)
		repositori.GET("/aslab/:nim", repositoriController.GetAslabRepositori)
		repositori.DELETE("/:repo_id", repositoriController.DeleteRepositori)
		repositori.GET("/jumlah/:nim", repositoriController.CountRepositoriAslab)
		repositori.GET("/count", repositoriController.CountAllRepositori)
		repositori.GET("/filter/:filter", repositoriController.AllRepositoriOverview)
	}
	dosen := r.Group("/dosen")
	{
		dosenController := controllers.NewDosenController(db, storageService)
		dosen.GET("", dosenController.GetAllDosen)
		dosen.POST("", dosenController.CreateDosen)
		dosen.DELETE("/:nip", dosenController.DeleteDosen)
	}
}
