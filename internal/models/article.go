package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Article struct {
	ArticleId string    `gorm:"type:char(36);primaryKey" json:"article_id"`
	Judul     string    `gorm:"size:255;index" json:"judul"`
	PhotoUrl  string    `gorm:"type:text" json:"photo_url"`
	Content   string    `gorm:"type:text" json:"content"`
	Views     int       `json:"views"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Nim       string    `gorm:"size:25" json:"nim"`

	Tags []Tags `gorm:"many2many:article_tag" json:"tags,omitempty"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) (err error) {
	a.ArticleId = uuid.New().String()
	return
}

func (Article) TableName() string {
	return "article"
}
