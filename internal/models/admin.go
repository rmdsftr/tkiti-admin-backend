package models

type Admin struct {
	Nim       string `gorm:"primaryKey;size:25" json:"nim"`
	Nama      string `gorm:"size:255;not null;index" json:"nama"`
	NoAslab   string `gorm:"size:25;not null;unique" json:"no_aslab"`
	Pword     string `gorm:"type:text;not null" json:"-"`
	Deskripsi string `gorm:"type:text" json:"deskripsi"`
	PhotoUrl  string `gorm:"type:text" json:"photo_url"`
}

func (Admin) TableName() string {
	return "admin"
}
