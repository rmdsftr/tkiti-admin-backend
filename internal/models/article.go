package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArticleEnum string

const (
	Published ArticleEnum = "published"
	Draft     ArticleEnum = "draft"
)

type Article struct {
	ArticleId     string      `gorm:"type:char(36);primaryKey" json:"article_id"`
	Judul         string      `gorm:"size:255;index" json:"judul"`
	PhotoUrl      string      `gorm:"type:text" json:"photo_url"`
	Content       string      `gorm:"type:text" json:"content"`
	StatusArticle ArticleEnum `gorm:"type:enum('published','draft');" json:"status_article"`
	Views         int         `json:"views"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	Nim   string `gorm:"size:25" json:"nim"`
	Admin Admin  `gorm:"foreignKey:Nim;references:Nim;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Tags []Tags `gorm:"many2many:article_tag;joinForeignKey:ArticleId;joinReferences:TagId"`
}

func (a *Article) BeforeCreate(tx *gorm.DB) (err error) {
	a.ArticleId = uuid.New().String()
	return
}

func (Article) TableName() string {
	return "article"
}
