package database

import (
	"admin-panel/internal/config"
	"admin-panel/internal/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal terkoneksi ke database")
	}

	DB = db
	return db, nil
}

func Migrate(db *gorm.DB) error {

	db.DisableForeignKeyConstraintWhenMigrating = true

	err := db.AutoMigrate(
		&models.Admin{},
		&models.Tags{},
		&models.Periode{},
		&models.Kegiatan{},
		&models.Repositori{},
		&models.Article{},
		&models.Pengurus{},
		&models.ArticleTag{},
		&models.Kontributor{},
		&models.Dokumentasi{},
	)
	if err != nil {
		return fmt.Errorf("gagal melakukan migrasi database: %w", err)
	}

	db.Exec(`
		ALTER TABLE article 
		ADD CONSTRAINT fk_article_admin 
		FOREIGN KEY (nim) REFERENCES admin(nim) 
		ON UPDATE CASCADE ON DELETE SET NULL
	`)

	db.Exec(`
		ALTER TABLE pengurus 
		ADD CONSTRAINT fk_pengurus_admin 
		FOREIGN KEY (nim) REFERENCES admin(nim) 
		ON UPDATE CASCADE ON DELETE CASCADE
	`)

	db.Exec(`
		ALTER TABLE pengurus 
		ADD CONSTRAINT fk_pengurus_periode 
		FOREIGN KEY (periode_id) REFERENCES periode(periode_id) 
		ON UPDATE CASCADE ON DELETE CASCADE
	`)

	db.Exec(`
		ALTER TABLE article_tag 
		ADD CONSTRAINT fk_article_tag_article 
		FOREIGN KEY (article_id) REFERENCES article(article_id) 
		ON UPDATE CASCADE ON DELETE CASCADE
	`)

	db.Exec(`
		ALTER TABLE article_tag 
		ADD CONSTRAINT fk_article_tag_tags 
		FOREIGN KEY (tag_id) REFERENCES tags(tag_id) 
		ON UPDATE CASCADE ON DELETE CASCADE
	`)

	db.Exec(`
		ALTER TABLE kontributor 
		ADD CONSTRAINT fk_kontributor_admin 
		FOREIGN KEY (nim) REFERENCES admin(nim) 
		ON UPDATE CASCADE ON DELETE RESTRICT
	`)

	db.Exec(`
		ALTER TABLE kontributor 
		ADD CONSTRAINT fk_kontributor_repositori 
		FOREIGN KEY (repo_id) REFERENCES repositori(repositori_id) 
		ON UPDATE CASCADE ON DELETE CASCADE
	`)

	db.Exec(`
		ALTER TABLE dokumentasi 
		ADD CONSTRAINT fk_dokumentasi_repositori 
		FOREIGN KEY (repo_id) REFERENCES repositori(repositori_id) 
		ON UPDATE CASCADE ON DELETE CASCADE
	`)

	return nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
