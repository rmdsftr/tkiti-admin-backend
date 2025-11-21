package main

import (
	"admin-panel/internal/config"
	"admin-panel/internal/database"
	"admin-panel/internal/middleware"
	"admin-panel/internal/routes"
	"admin-panel/internal/services"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config         *config.Config
	DB             *gorm.DB
	Router         *gin.Engine
	StorageService *services.StorageService
}

func newApp(cfg *config.Config) *App {
	app := &App{
		Config: cfg,
	}

	app.initDatabase()
	app.initStorageService()
	app.initRouter()

	return app
}

func (app *App) initDatabase() {
	db, err := database.Connect(app.Config)
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Gagal melakukan migrasi database:", err)
	}

	app.DB = db
	log.Println("Database connected successfully")
}

func (app *App) initStorageService() {
	storageService, err := services.NewStorageService(app.Config)
	if err != nil {
		log.Fatal("Gagal menginisialisasi storage service:", err)
	}

	if storageService == nil {
		log.Fatal("Storage service is nil after initialization")
	}

	app.StorageService = storageService
	log.Println("Storage service initialized successfully")
}

func (app *App) initRouter() {
	r := gin.Default()
	r.Use(middleware.CORS())
	routes.SetupRoutes(r, app.DB, app.StorageService)
	app.Router = r
}

func main() {
	cfg := config.LoadConfig()
	app := newApp(cfg)
	log.Printf("Starting server on %s", cfg.ServerPort)
	app.Router.Run(cfg.ServerPort)
}
