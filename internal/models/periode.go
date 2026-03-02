package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusPeriodeEnum string

const (
	PeriodeAktif    StatusPeriodeEnum = "aktif"
	PeriodeNonaktif StatusPeriodeEnum = "nonaktif"
)

type Periode struct {
	PeriodeId     string            `gorm:"type:char(36);primaryKey;" json:"periode_id"`
	NamaPeriode   string            `gorm:"index;not null" json:"nama_periode"`
	StatusPeriode StatusPeriodeEnum `gorm:"type:enum('aktif','nonaktif')" json:"status_periode"`
}

func (p *Periode) BeforeCreate(tx *gorm.DB) (err error) {
	p.PeriodeId = uuid.New().String()
	return
}

func (Periode) TableName() string {
	return "periode"
}
