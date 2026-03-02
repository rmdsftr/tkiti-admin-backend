package models

type PosisiEnum string

const (
	PosisiKepala  PosisiEnum = "kepala"
	PosisiAnggota PosisiEnum = "anggota"
)

type Dosen struct {
	NIP       string     `gorm:"column:nip;primaryKey" json:"nip"`
	NamaDosen string     `gorm:"size:255;not null;index" json:"nama_dosen"`
	Foto      string     `gorm:"type:text" json:"foto"`
	Posisi    PosisiEnum `gorm:"type:enum('kepala','anggota');default:'anggota'" json:"posisi"`
}

func (Dosen) TableName() string {
	return "dosen"
}
