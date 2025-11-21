package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Kegiatan struct {
	KegiatanId string    `gorm:"type:char(36);primaryKey" json:"kegiatan_id"`
	Judul      string    `gorm:"type:varchar(255);index" json:"judul"`
	Deskripsi  string    `gorm:"type:text" json:"deskripsi"`
	PhotoUrl   string    `gorm:"type:text" json:"photo_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (k *Kegiatan) BeforeCreate(tx *gorm.DB) (err error) {
	k.KegiatanId = uuid.New().String()
	return
}

func (Kegiatan) TableName() string {
	return "kegiatan"
}
