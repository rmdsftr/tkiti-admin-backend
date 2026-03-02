package controllers

import (
	"admin-panel/internal/middleware"
	"admin-panel/internal/models"
	"admin-panel/internal/services"
	"admin-panel/pkg/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB             *gorm.DB
	StorageService *services.StorageService
}

func NewAdminController(db *gorm.DB, storageService *services.StorageService) *AdminController {
	return &AdminController{
		DB:             db,
		StorageService: storageService,
	}
}

func (ac *AdminController) GetAllAdmin(c *gin.Context) {
	type AdminWithCount struct {
		Nim        string `json:"nim"`
		Nama       string `json:"nama"`
		Divisi     string `json:"divisi"`
		Jabatan    string `json:"jabatan"`
		JmlArtikel int64  `json:"jumlah_artikel"`
		Role       string `json:"role"`
		Status     string `json:"status"`
		PhotoUrl   string `json:"photo_url"`
	}

	periodeId := c.GetString("periode_id")

	if periodeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak ada periode aktif saat ini",
		})
		return
	}

	var result []AdminWithCount

	err := ac.DB.Table("pengurus").
		Select(`
        admin.nim,
        admin.nama,
        pengurus.divisi,
        pengurus.jabatan,
        admin.role,
        admin.status,
        admin.photo_url,
        COUNT(article.article_id) AS jml_artikel
    `).
		Joins("JOIN admin ON admin.nim = pengurus.nim").
		Joins("LEFT JOIN article ON article.nim = admin.nim").
		Where("pengurus.periode_id = ?", periodeId).
		Where("article.status_article=?", "published").
		Group("admin.nim, admin.nama, pengurus.divisi, pengurus.jabatan, admin.role, admin.status, admin.photo_url").
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil admin pada periode aktif",
		"data":    result,
	})
}

func (ac *AdminController) CreateAdmin(c *gin.Context) {
	var input struct {
		Nim       string             `json:"nim" binding:"required"`
		Nama      string             `json:"nama" binding:"required"`
		NoAslab   string             `json:"no_aslab" binding:"required"`
		Pword     string             `json:"pword" binding:"required"`
		Divisi    models.DivisiEnum  `json:"divisi" binding:"required,oneof=inti litbang rtk pengpel"`
		Jabatan   models.JabatanEnum `json:"jabatan" binding:"required,oneof=kordas sekretaris bendahara koordinator anggota"`
		Deskripsi string             `json:"deskripsi"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	periodeId := c.GetString("periode_id")
	if periodeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Periode ID tidak ditemukan",
		})
		return
	}

	var existing models.Admin
	if err := ac.DB.First(&existing, "nim = ?", input.Nim).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Admin dengan NIM tersebut sudah ada",
		})
		return
	}

	uniqueRules := map[string]struct {
		divisi  models.DivisiEnum
		jabatan models.JabatanEnum
		message string
	}{
		"inti-kordas":         {"inti", "kordas", "Koordinator asisten pada periode ini sudah ada"},
		"litbang-koordinator": {"litbang", "koordinator", "Koordinator divisi litbang pada periode ini sudah ada"},
		"rtk-koordinator":     {"rtk", "koordinator", "Koordinator divisi RTK pada periode ini sudah ada"},
		"pengpel-koordinator": {"pengpel", "koordinator", "Koordinator divisi pengpel pada periode ini sudah ada"},
	}

	key := string(input.Divisi) + "-" + string(input.Jabatan)
	if rule, exists := uniqueRules[key]; exists {
		var count int64
		err := ac.DB.Table("pengurus").
			Where("divisi = ? AND jabatan = ? AND periode_id = ?", rule.divisi, rule.jabatan, periodeId).
			Count(&count).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memeriksa data pengurus",
			})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": rule.message,
			})
			return
		}
	}

	hashedPassword, err := utils.HashPassword(input.Pword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal melakukan hashing password",
		})
		return
	}

	tx := ac.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	admin := models.Admin{
		Nim:       input.Nim,
		Nama:      utils.ToTitleCase(input.Nama),
		NoAslab:   strings.ToUpper(input.NoAslab),
		Pword:     hashedPassword,
		Deskripsi: input.Deskripsi,
	}

	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal membuat admin: " + err.Error(),
		})
		return
	}

	pengurus := models.Pengurus{
		Nim:       input.Nim,
		PeriodeId: periodeId,
		Divisi:    input.Divisi,
		Jabatan:   input.Jabatan,
	}

	if key := string(input.Divisi) + "-" + string(input.Jabatan); uniqueRules[key].divisi != "" {
		var count int64
		if err := tx.Table("pengurus").
			Where("divisi = ? AND jabatan = ? AND periode_id = ?", input.Divisi, input.Jabatan, periodeId).
			Count(&count).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memeriksa data pengurus",
			})
			return
		}

		if count > 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": uniqueRules[key].message,
			})
			return
		}
	}

	if err := tx.Create(&pengurus).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal membuat pengurus: " + err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil membuat data pengurus",
		"data":    pengurus,
	})
}

func (ac *AdminController) UpdateRoleOrStatus(c *gin.Context) {
	nim := c.Param("nim")
	val := c.Param("roleorstatus")

	var admin models.Admin

	if err := ac.DB.First(&admin, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Admin tidak ditemukan",
		})
		return
	}

	if val == string(models.RoleAdmin) || val == string(models.RoleMember) {
		ac.DB.Model(&admin).Update("role", models.RoleEnum(val))

		c.JSON(http.StatusOK, gin.H{
			"message": "Role berhasil diperbarui",
			"data":    admin,
		})
		return
	}

	if val == string(models.StatusAktif) || val == string(models.StatusNonAktif) {
		ac.DB.Model(&admin).Update("status", models.StatusEnum(val))

		c.JSON(http.StatusOK, gin.H{
			"message": "Status berhasil diperbarui",
			"data":    admin,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Value tidak valid, gunakan: admin/member atau aktif/nonaktif",
	})
}

func (ac *AdminController) DeleteAccount(c *gin.Context) {
	nim := c.Param("nim")
	var admin models.Admin

	if err := ac.DB.First(&admin, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	if err := ac.DB.Delete(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus akun user",
		})
		return
	}

	var pengurus models.Pengurus
	if err := ac.DB.Delete(&pengurus, "nim=?", nim).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus akun user dari tabel pengurus",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus akun user",
	})
}

func (ac *AdminController) DeletePengurus(c *gin.Context) {
	nim := c.Param("nim")
	periodeId := c.GetString("periode_id")

	var admin models.Admin
	if err := ac.DB.First(&admin, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	if err := ac.DB.Delete(&models.Pengurus{}, "nim = ? AND periode_id = ?", nim, periodeId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus akun user dari tabel pengurus",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus user dari pengurus",
	})
}

func (ac *AdminController) GetAdminbyPeriode(c *gin.Context) {
	periodeId := c.Param("periode_id")
	periodeIdCurrent := c.GetString("periode_id")

	type AdminWithCount struct {
		Nim        string `json:"nim"`
		Nama       string `json:"nama"`
		Divisi     string `json:"divisi"`
		Jabatan    string `json:"jabatan"`
		JmlArtikel int64  `json:"jumlah_artikel"`
		Role       string `json:"role"`
		Status     string `json:"status"`
		PhotoUrl   string `json:"photo_url"`
	}

	if periodeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak ada periode aktif saat ini",
		})
		return
	}

	var result []AdminWithCount

	err := ac.DB.Table("pengurus").
		Select(`
		admin.nim,
		admin.nama,
		pengurus.divisi,
		pengurus.jabatan,
		admin.role,
		admin.status,
		admin.photo_url,
		COUNT(article.article_id) AS jml_artikel
	`).
		Joins("JOIN admin ON admin.nim = pengurus.nim").
		Joins(`
		LEFT JOIN article 
		ON article.nim = admin.nim 
		AND article.status_article = 'published'
	`).
		Where("pengurus.periode_id = ?", periodeId).
		Group(`
		admin.nim, admin.nama, pengurus.divisi, 
		pengurus.jabatan, admin.role, admin.status, admin.photo_url
	`).
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var isCurrent bool
	if periodeId == periodeIdCurrent {
		isCurrent = true
	} else {
		isCurrent = false
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       result,
		"is_current": isCurrent,
	})
}

func (ac *AdminController) ChangePhoto(c *gin.Context) {
	nim := c.Param("nim")

	var user models.Admin
	if err := ac.DB.First(&user, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	file, err := c.FormFile("photo_url")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file foto wajib diupload"})
		return
	}

	if !middleware.IsValidImageType(file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipe file tidak valid"})
		return
	}

	if file.Size > 25*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran gambar terlalu besar"})
		return
	}

	photoURL, err := ac.StorageService.UploadFile(file, "admin-photo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengupload gambar: " + err.Error(),
		})
		return
	}

	if user.PhotoUrl != "" {
		filePath := ac.StorageService.ExtractFilePathFromURL(user.PhotoUrl)
		if filePath != "" {
			if err := ac.StorageService.DeleteFile(filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("gagal menghapus file: %v", err),
				})
				return
			}
		}
	}

	if err := ac.DB.Model(&user).Update("photo_url", photoURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal menyimpan foto ke database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Foto profil berhasil diperbarui",
		"photo_url": photoURL,
	})
}

func (ac *AdminController) DeletePhoto(c *gin.Context) {
	nim := c.Param("nim")

	var user models.Admin
	if err := ac.DB.First(&user, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User tidak ditemukan",
		})
		return
	}

	if user.PhotoUrl != "" {
		key := ac.StorageService.ExtractFilePathFromURL(user.PhotoUrl)
		if key != "" {
			if err := ac.StorageService.DeleteFile(key); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal menghapus file foto",
				})
				return
			}
		}
	}

	if err := ac.DB.Model(&user).
		Update("photo_url", "").
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memperbarui data user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto berhasil dihapus",
	})
}

func (ac *AdminController) GetUserPhotoProfile(c *gin.Context) {
	nim := c.Param("nim")

	var user models.Admin
	if err := ac.DB.First(&user, "nim=?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"photo_url": user.PhotoUrl,
	})
}

func (ac *AdminController) GetUserDataProfile(c *gin.Context) {
	nim := c.Param("nim")

	var user models.Admin
	if err := ac.DB.First(&user, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "admin tidak ditemukan",
		})
		return
	}

	var pengurus []models.Pengurus
	if err := ac.DB.
		Preload("Periode").
		Where("nim = ?", nim).
		Find(&pengurus).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	type PeriodeInfo struct {
		PeriodeId   string `json:"periode_id"`
		NamaPeriode string `json:"nama_periode"`
		Divisi      string `json:"divisi"`
		Jabatan     string `json:"jabatan"`
	}

	var periodeList []PeriodeInfo
	for _, p := range pengurus {
		periodeList = append(periodeList, PeriodeInfo{
			PeriodeId:   p.Periode.PeriodeId,
			NamaPeriode: p.Periode.NamaPeriode,
			Divisi:      string(p.Divisi),
			Jabatan:     string(p.Jabatan),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"nim":       user.Nim,
		"nama":      user.Nama,
		"no_aslab":  user.NoAslab,
		"role":      user.Role,
		"photo_url": user.PhotoUrl,
		"pengurus":  periodeList,
	})
}

func (ac *AdminController) CreateAdminFromOld(c *gin.Context) {
	var input struct {
		Nim     string             `json:"nim" binding:"required"`
		Divisi  models.DivisiEnum  `json:"divisi" binding:"required,oneof=inti litbang rtk pengpel"`
		Jabatan models.JabatanEnum `json:"jabatan" binding:"omitempty,oneof=kordas sekretaris bendahara koordinator anggota"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	periodeId := c.GetString("periode_id")

	uniqueRules := map[string]struct {
		divisi  models.DivisiEnum
		jabatan models.JabatanEnum
		message string
	}{
		"inti-kordas":         {"inti", "kordas", "Koordinator asisten pada periode ini sudah ada"},
		"litbang-koordinator": {"litbang", "koordinator", "Koordinator divisi litbang pada periode ini sudah ada"},
		"rtk-koordinator":     {"rtk", "koordinator", "Koordinator divisi RTK pada periode ini sudah ada"},
		"pengpel-koordinator": {"pengpel", "koordinator", "Koordinator divisi pengpel pada periode ini sudah ada"},
	}

	key := string(input.Divisi) + "-" + string(input.Jabatan)
	if rule, exists := uniqueRules[key]; exists {
		var count int64
		err := ac.DB.Table("pengurus").
			Where("divisi = ? AND jabatan = ? AND periode_id = ?", rule.divisi, rule.jabatan, periodeId).
			Count(&count).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memeriksa data pengurus",
			})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": rule.message,
			})
			return
		}
	}

	submit := models.Pengurus{
		PeriodeId: periodeId,
		Nim:       input.Nim,
		Divisi:    input.Divisi,
		Jabatan:   input.Jabatan,
	}

	if err := ac.DB.Create(&submit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengurus baru berhasil ditambahkan",
	})
}

func (ac *AdminController) GetAslabNotInPeriode(c *gin.Context) {
	type Result struct {
		Nim  string `json:"nim"`
		Nama string `json:"nama"`
	}

	periodeId := c.GetString("periode_id")

	var anggota []Result

	err := ac.DB.
		Table("admin").
		Select("admin.nim, admin.nama").
		Joins("LEFT JOIN pengurus ON pengurus.nim = admin.nim AND pengurus.periode_id = ?", periodeId).
		Where("pengurus.nim IS NULL").
		Scan(&anggota).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admins": anggota,
	})
}
