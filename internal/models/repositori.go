package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JenisRepoEnum string

const (
	JenisPrestasi  JenisRepoEnum = "prestasi"
	JenisPublikasi JenisRepoEnum = "ilmiah"
	JenisProyek    JenisRepoEnum = "proyek"
)

type Repositori struct {
	RepositoriId string        `gorm:"type:char(36);primaryKey" json:"repo_id"`
	JudulRepo    string        `gorm:"size:255;not null" json:"judul_repo"`
	Deskripsi    string        `gorm:"type:text" json:"deskripsi"`
	PhotoUrl     string        `gorm:"type:text" json:"photo_url"`
	JenisRepo    JenisRepoEnum `gorm:"type:enum('prestasi','ilmiah','proyek');index" json:"jenis_repo"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	Kontributor []Kontributor `gorm:"foreignKey:RepositoriId;references:RepositoriId" json:"kontributor,omitempty"`
	Dokumentasi []Dokumentasi `gorm:"foreignKey:RepositoriId;references:RepositoriId" json:"dokumentasi,omitempty"`
}

func (r *Repositori) BeforeCreate(tx *gorm.DB) (err error) {
	r.RepositoriId = uuid.New().String()
	return
}

func (Repositori) TableName() string {
	return "repositori"
}
