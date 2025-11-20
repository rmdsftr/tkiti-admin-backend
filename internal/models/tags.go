package models

type Tags struct {
	TagId        int    `gorm:"primaryKey;autoIncrement" json:"tag_id"`
	Tag          string `gorm:"not null" json:"tag"`
	TotalArticle int    `json:"total_article"`

	Articles []Article `gorm:"many2many:article_tag"`
}

func (Tags) TableName() string {
	return "tags"
}
