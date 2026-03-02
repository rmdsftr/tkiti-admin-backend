package controllers

import (
	"admin-panel/internal/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PeriodeController struct {
	DB *gorm.DB
}

func NewPeriodeController(db *gorm.DB) *PeriodeController {
	return &PeriodeController{
		DB: db,
	}
}

func (pc *PeriodeController) CreatePeriode(c *gin.Context) {
	var periodeLama models.Periode

	err := pc.DB.Where("status_periode = ?", models.PeriodeAktif).First(&periodeLama).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil periode sebelumnya",
			})
			return
		}
	} else {
		periodeLama.StatusPeriode = models.PeriodeNonaktif
		if err := pc.DB.Save(&periodeLama).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menonaktifkan periode lama",
			})
			return
		}
	}

	var input struct {
		NamaPeriode string `json:"nama_periode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	periodeBaru := models.Periode{
		NamaPeriode:   input.NamaPeriode,
		StatusPeriode: models.PeriodeAktif,
	}

	if err := pc.DB.Create(&periodeBaru).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Periode baru gagal diaktivasi",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Periode baru berhasil diaktivasi",
		"data":    periodeBaru,
	})
}

func (pc *PeriodeController) GetAllPeriode(c *gin.Context) {
	type PeriodeResult struct {
		PeriodeId     string `json:"periode_id"`
		NamaPeriode   string `json:"nama_periode"`
		StatusPeriode string `json:"status_periode"`
	}

	var periodes []models.Periode

	if err := pc.DB.Find(&periodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	aktif := []PeriodeResult{}
	nonaktif := []PeriodeResult{}

	for _, p := range periodes {
		item := PeriodeResult{
			PeriodeId:     p.PeriodeId,
			NamaPeriode:   p.NamaPeriode,
			StatusPeriode: string(p.StatusPeriode),
		}

		if p.StatusPeriode == models.PeriodeAktif {
			aktif = append(aktif, item)
		} else {
			nonaktif = append(nonaktif, item)
		}
	}

	results := append(aktif, nonaktif...)

	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

