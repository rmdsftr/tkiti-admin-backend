package models

type ArticleTag struct {
	ArticleId string  `gorm:"type:char(36);primaryKey"`
	Article   Article `gorm:"foreignKey:ArticleId;references:ArticleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	TagId int  `gorm:"primaryKey"`
	Tags  Tags `gorm:"foreignKey:TagId;references:TagId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ArticleTag) TableName() string {
	return "article_tag"
}
