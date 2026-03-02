package models

type Dokumentasi struct {
	LinkId       int        `gorm:"primaryKey;autoIncrement" json:"link_id"`
	Link         string     `gorm:"type:text" json:"link"`
	RepositoriId string     `gorm:"type:char(36);not null;index" json:"repo_id"`
	Repositori   Repositori `gorm:"foreignKey:RepositoriId;references:RepositoriId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Dokumentasi) TableName() string {
	return "dokumentasi"
}
