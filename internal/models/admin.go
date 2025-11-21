package models

type RoleEnum string

const (
	RoleAdmin  RoleEnum = "admin"
	RoleMember RoleEnum = "member"
)

type StatusEnum string

const (
	StatusAktif    StatusEnum = "aktif"
	StatusNonAktif StatusEnum = "nonaktif"
)

type Admin struct {
	Nim       string `gorm:"primaryKey;size:25" json:"nim"`
	Nama      string `gorm:"size:255;not null;index" json:"nama"`
	NoAslab   string `gorm:"size:25;not null;unique" json:"no_aslab"`
	Pword     string `gorm:"type:text;not null" json:"-"`
	Deskripsi string `gorm:"type:text" json:"deskripsi"`
	PhotoUrl  string `gorm:"type:text" json:"photo_url"`

	Role   RoleEnum   `gorm:"type:enum('admin','member');default:'member'" json:"role"`
	Status StatusEnum `gorm:"type:enum('aktif','nonaktif');default:'aktif'" json:"status"`
}

func (Admin) TableName() string {
	return "admin"
}
