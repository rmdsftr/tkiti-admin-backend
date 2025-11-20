package models

type ArticleTag struct {
	ArticleId string `gorm:"type:char(36);primaryKey"`
	TagId     int    `gorm:"primaryKey"`
}

func (ArticleTag) TableName() string {
	return "article_tag"
}
