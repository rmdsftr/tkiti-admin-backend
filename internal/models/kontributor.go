package models

type Kontributor struct {
	Nim   string `gorm:"primaryKey;size:25;not null;index" json:"nim"`
	Admin Admin  `gorm:"foreignKey:Nim;references:Nim;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	RepositoriId string     `gorm:"primaryKey;type:char(36);not null;index" json:"repo_id"`
	Repositori   Repositori `gorm:"foreignKey:RepositoriId;references:RepositoriId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Kontributor) TableName() string {
	return "kontributor"
}
