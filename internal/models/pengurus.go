package models

type DivisiEnum string

const (
	DivisiInti    DivisiEnum = "inti"
	DivisiLitbang DivisiEnum = "litbang"
	DivisiRTK     DivisiEnum = "rtk"
	DivisiPengpel DivisiEnum = "pengpel"
)

type JabatanEnum string

const (
	JabatanKordas JabatanEnum = "kordas"
	JabatanSekre  JabatanEnum = "sekretaris"
	JabatanBend   JabatanEnum = "bendahara"
	JabatanKoor   JabatanEnum = "koordinator"
	JabatanStaff  JabatanEnum = "anggota"
)

type Pengurus struct {
	PeriodeId string  `gorm:"type:char(36);primaryKey;" json:"periode_id"`
	Periode   Periode `gorm:"foreignKey:PeriodeId;references:PeriodeId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Nim   string `gorm:"primaryKey;size:25;" json:"nim"`
	Admin Admin  `gorm:"foreignKey:Nim;references:Nim;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Divisi  DivisiEnum  `gorm:"type:enum('inti','litbang','rtk','pengpel')" json:"divisi"`
	Jabatan JabatanEnum `gorm:"type:enum('kordas','sekretaris','bendahara','koordinator','anggota')" json:"jabatan"`
}

func (Pengurus) TableName() string {
	return "pengurus"
}
