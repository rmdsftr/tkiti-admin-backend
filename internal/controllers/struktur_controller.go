package controllers

import (
	"admin-panel/internal/models"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StrukturController struct {
	DB *gorm.DB
}

func NewStrukturController(db *gorm.DB) *StrukturController {
	return &StrukturController{
		DB: db,
	}
}

func (sc *StrukturController) GetCurrentStruktur(c *gin.Context) {
	periodeId := c.GetString("periode_id")

	type Anggota struct {
		Nim      string `json:"nim"`
		Nama     string `json:"nama"`
		Divisi   string `json:"divisi"`
		Jabatan  string `json:"jabatan"`
		Role     string `json:"role"`
		PhotoUrl string `json:"photo_url"`
	}

	var result []Anggota
	err := sc.DB.Table("pengurus").
		Select(`
			admin.nim,
			admin.nama,
			pengurus.divisi,
			pengurus.jabatan,
			admin.role,
			admin.photo_url
		`).
		Joins("JOIN admin ON admin.nim = pengurus.nim").
		Where("pengurus.periode_id = ?", periodeId).
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grouped := map[string][]Anggota{
		"inti":    {},
		"litbang": {},
		"rtk":     {},
		"pengpel": {},
	}

	for _, a := range result {
		div := strings.ToLower(a.Divisi)
		if _, ok := grouped[div]; ok {
			grouped[div] = append(grouped[div], a)
		}
	}

	sortAnggota := func(list []Anggota, divisi string) []Anggota {
		sort.SliceStable(list, func(i, j int) bool {
			a := strings.ToLower(list[i].Jabatan)
			b := strings.ToLower(list[j].Jabatan)

			if divisi == "inti" {

				if a == "kordas" && b != "kordas" {
					return true
				}
				if b == "kordas" && a != "kordas" {
					return false
				}
			} else {

				if a == "koordinator" && b != "koordinator" {
					return true
				}
				if b == "koordinator" && a != "koordinator" {
					return false
				}
			}

			return a < b
		})

		return list
	}

	grouped["inti"] = sortAnggota(grouped["inti"], "inti")
	grouped["litbang"] = sortAnggota(grouped["litbang"], "litbang")
	grouped["rtk"] = sortAnggota(grouped["rtk"], "rtk")
	grouped["pengpel"] = sortAnggota(grouped["pengpel"], "pengpel")

	// Fetch dosen data
	var dosenList []models.Dosen
	if err := sc.DB.Order("FIELD(posisi, 'kepala', 'anggota')").Find(&dosenList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ordered := gin.H{
		"dosen":   dosenList,
		"inti":    grouped["inti"],
		"litbang": grouped["litbang"],
		"rtk":     grouped["rtk"],
		"pengpel": grouped["pengpel"],
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Struktur berhasil diambil",
		"data":    ordered,
	})
}

func (sc *StrukturController) GetStrukturByPeriode(c *gin.Context) {
	periodeId := c.Param("periode_id")

	type Anggota struct {
		Nim      string `json:"nim"`
		Nama     string `json:"nama"`
		Divisi   string `json:"divisi"`
		Jabatan  string `json:"jabatan"`
		Role     string `json:"role"`
		PhotoUrl string `json:"photo_url"`
	}

	var result []Anggota
	err := sc.DB.Table("pengurus").
		Select(`
			admin.nim,
			admin.nama,
			pengurus.divisi,
			pengurus.jabatan,
			admin.role,
			admin.photo_url
		`).
		Joins("JOIN admin ON admin.nim = pengurus.nim").
		Where("pengurus.periode_id = ?", periodeId).
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	grouped := map[string][]Anggota{
		"inti":    {},
		"litbang": {},
		"rtk":     {},
		"pengpel": {},
	}

	for _, a := range result {
		div := strings.ToLower(a.Divisi)
		if _, ok := grouped[div]; ok {
			grouped[div] = append(grouped[div], a)
		}
	}

	sortAnggota := func(list []Anggota, divisi string) []Anggota {
		sort.SliceStable(list, func(i, j int) bool {
			a := strings.ToLower(list[i].Jabatan)
			b := strings.ToLower(list[j].Jabatan)

			if divisi == "inti" {

				if a == "kordas" && b != "kordas" {
					return true
				}
				if b == "kordas" && a != "kordas" {
					return false
				}
			} else {

				if a == "koordinator" && b != "koordinator" {
					return true
				}
				if b == "koordinator" && a != "koordinator" {
					return false
				}
			}

			return a < b
		})

		return list
	}

	grouped["inti"] = sortAnggota(grouped["inti"], "inti")
	grouped["litbang"] = sortAnggota(grouped["litbang"], "litbang")
	grouped["rtk"] = sortAnggota(grouped["rtk"], "rtk")
	grouped["pengpel"] = sortAnggota(grouped["pengpel"], "pengpel")

	// Fetch dosen data
	var dosenList []models.Dosen
	if err := sc.DB.Order("FIELD(posisi, 'kepala', 'anggota')").Find(&dosenList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ordered := gin.H{
		"dosen":   dosenList,
		"inti":    grouped["inti"],
		"litbang": grouped["litbang"],
		"rtk":     grouped["rtk"],
		"pengpel": grouped["pengpel"],
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Struktur berhasil diambil",
		"data":    ordered,
	})
}
