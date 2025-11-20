package main

import (
	"admin-panel/internal/config"
	"admin-panel/internal/database"
	"admin-panel/internal/routes"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router *gin.Engine
}

func newApp(cfg *config.Config) *App {
	app := &App{
		Config: cfg,
	}

	app.initDatabase()
	app.initRouter()

	return app
}

func (app *App) initDatabase() {
	db, err := database.Connect(app.Config)
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database")
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("gagal melakukan migrasi database")
	}

	app.DB = db
}

func (app *App) initRouter() {
	r := gin.Default()
	routes.SetupRoutes(r, app.DB)
	app.Router = r
}

func main() {
	cfg := config.LoadConfig()
	app := newApp(cfg)
	app.Router.Run(cfg.ServerPort)
}
