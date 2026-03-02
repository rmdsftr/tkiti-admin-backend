package controllers

import (
	"admin-panel/internal/middleware"
	"admin-panel/internal/models"
	"admin-panel/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RepositoriController struct {
	DB             *gorm.DB
	StorageService *services.StorageService
}

func NewRepositoriController(db *gorm.DB, storageService *services.StorageService) *RepositoriController {
	return &RepositoriController{
		DB:             db,
		StorageService: storageService,
	}
}

func (rc *RepositoriController) GetKontributors(c *gin.Context) {
	type Result struct {
		Nim  string `json:"nim"`
		Nama string `json:"nama"`
	}

	var anggota []models.Admin

	if err := rc.DB.Find(&anggota).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var output []Result

	for _, a := range anggota {
		output = append(output, Result{
			Nim:  a.Nim,
			Nama: a.Nama,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"kontributors": output,
	})
}

func (rc *RepositoriController) SubmitNewRepositori(c *gin.Context) {
	judul := c.PostForm("judul_repo")
	deskripsi := c.PostForm("deskripsi")
	jenisRepo := c.PostForm("jenis_repo")

	nims := c.PostFormArray("nim")
	links := c.PostFormArray("links")

	if judul == "" || deskripsi == "" || jenisRepo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak lengkap"})
		return
	}

	file, err := c.FormFile("photo_url")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !middleware.IsValidImageType(file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipe gambar tidak valid"})
		return
	}

	if file.Size > 25*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran gambar terlalu besar"})
		return
	}

	photoURL, err := rc.StorageService.UploadFile(file, "repositori")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	repositori := models.Repositori{
		JudulRepo: judul,
		Deskripsi: deskripsi,
		PhotoUrl:  photoURL,
		JenisRepo: models.JenisRepoEnum(jenisRepo),
	}

	if err := rc.DB.Create(&repositori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(nims) > 0 {
		var contributors []models.Kontributor
		for _, nim := range nims {
			contributors = append(contributors, models.Kontributor{
				Nim:          nim,
				RepositoriId: repositori.RepositoriId,
			})
		}
		rc.DB.Create(&contributors)
	}

	if len(links) > 0 {
		var dokumentasi []models.Dokumentasi
		for _, l := range links {
			dokumentasi = append(dokumentasi, models.Dokumentasi{
				Link:         l,
				RepositoriId: repositori.RepositoriId,
			})
		}
		rc.DB.Create(&dokumentasi)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Repositori baru berhasil dibuat"})
}

type RepoRow struct {
	RepoId    string
	JudulRepo string
	Deskripsi string
	PhotoUrl  string
	JenisRepo string
	Link      string
	Nim       string
	Nama      string
}

type Names struct {
	Nim  string `json:"nim"`
	Nama string `json:"nama"`
}

type Data struct {
	RepoId       string   `json:"repo_id"`
	JudulRepo    string   `json:"judul_repo"`
	Deskripsi    string   `json:"deskripsi"`
	PhotoUrl     string   `json:"photo_url"`
	JenisRepo    string   `json:"jenis_repo"`
	Links        []string `json:"link"`
	Contributors []Names  `json:"contributors"`
}

func (rc *RepositoriController) GetAllRepositori(c *gin.Context) {
	var rows []RepoRow

	err := rc.DB.Table("repositori").
		Select(`
            repositori.repositori_id AS repo_id,
            repositori.judul_repo,
            repositori.deskripsi,
            repositori.photo_url,
            repositori.jenis_repo,
            dokumentasi.link,
            admin.nim,
            admin.nama
        `).
		Joins("LEFT JOIN kontributor ON kontributor.repositori_id = repositori.repositori_id").
		Joins("LEFT JOIN admin ON admin.nim = kontributor.nim").
		Joins("LEFT JOIN dokumentasi ON dokumentasi.repositori_id = repositori.repositori_id").
		Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	repoMap := make(map[string]*Data)

	for _, r := range rows {
		if _, exists := repoMap[r.RepoId]; !exists {
			repoMap[r.RepoId] = &Data{
				RepoId:       r.RepoId,
				JudulRepo:    r.JudulRepo,
				Deskripsi:    r.Deskripsi,
				PhotoUrl:     r.PhotoUrl,
				JenisRepo:    r.JenisRepo,
				Links:        []string{},
				Contributors: []Names{},
			}
		}

		repoData := repoMap[r.RepoId]

		if !contains(repoData.Links, r.Link) {
			repoData.Links = append(repoData.Links, r.Link)
		}

		if !hasContributor(repoData.Contributors, r.Nim) {
			repoData.Contributors = append(repoData.Contributors, Names{
				Nim:  r.Nim,
				Nama: r.Nama,
			})
		}
	}

	final := []Data{}
	for _, v := range repoMap {
		final = append(final, *v)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": final,
	})
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

func hasContributor(arr []Names, nim string) bool {
	for _, v := range arr {
		if v.Nim == nim {
			return true
		}
	}
	return false
}

func (rc *RepositoriController) GetAslabRepositori(c *gin.Context) {
	var rows []RepoRow
	nim := c.Param("nim")

	err := rc.DB.Table("repositori").
		Select(`
			repositori.repositori_id AS repo_id,
			repositori.judul_repo,
			repositori.deskripsi,
			repositori.photo_url,
			repositori.jenis_repo,
			dokumentasi.link,
			admin.nim,
			admin.nama
		`).
		Joins("LEFT JOIN kontributor k1 ON k1.repositori_id = repositori.repositori_id").
		Joins("LEFT JOIN dokumentasi ON dokumentasi.repositori_id = repositori.repositori_id").
		Joins("LEFT JOIN kontributor k2 ON k2.repositori_id = repositori.repositori_id").
		Joins("LEFT JOIN admin ON admin.nim = k2.nim").
		Where("k1.nim = ?", nim).
		Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	repoMap := make(map[string]*Data)

	for _, r := range rows {
		if _, exists := repoMap[r.RepoId]; !exists {
			repoMap[r.RepoId] = &Data{
				RepoId:       r.RepoId,
				JudulRepo:    r.JudulRepo,
				Deskripsi:    r.Deskripsi,
				PhotoUrl:     r.PhotoUrl,
				JenisRepo:    r.JenisRepo,
				Links:        []string{},
				Contributors: []Names{},
			}
		}

		repoData := repoMap[r.RepoId]

		if !contains(repoData.Links, r.Link) {
			repoData.Links = append(repoData.Links, r.Link)
		}

		if !hasContributor(repoData.Contributors, r.Nim) {
			repoData.Contributors = append(repoData.Contributors, Names{
				Nim:  r.Nim,
				Nama: r.Nama,
			})
		}
	}

	final := []Data{}
	for _, v := range repoMap {
		final = append(final, *v)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": final,
	})
}

func (rc *RepositoriController) DeleteRepositori(c *gin.Context) {
	repoId := c.Param("repo_id")

	var repo models.Repositori
	if err := rc.DB.Delete(&repo, "repositori_id=?", repoId).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Repositori berhasil dihapus",
	})
}

func (rc *RepositoriController) CountRepositoriAslab(c *gin.Context) {
	nim := c.Param("nim")

	var result struct {
		JumlahArtikel  int `json:"jumlah_artikel"`
		JumlahPrestasi int `json:"jumlah_prestasi"`
		JumlahIlmiah   int `json:"jumlah_ilmiah"`
		JumlahProyek   int `json:"jumlah_proyek"`
	}

	var (
		articleCount  int64
		prestasiCount int64
		ilmiahCount   int64
		proyekCount   int64
	)

	rc.DB.Table("article").
		Where("nim = ?", nim).
		Count(&articleCount)

	rc.DB.Table("repositori").
		Joins("JOIN kontributor ON kontributor.repositori_id = repositori.repositori_id").
		Where("kontributor.nim = ?", nim).
		Where("repositori.jenis_repo = ?", "prestasi").
		Count(&prestasiCount)

	rc.DB.Table("repositori").
		Joins("JOIN kontributor ON kontributor.repositori_id = repositori.repositori_id").
		Where("kontributor.nim = ?", nim).
		Where("repositori.jenis_repo = ?", "ilmiah").
		Count(&ilmiahCount)

	rc.DB.Table("repositori").
		Joins("JOIN kontributor ON kontributor.repositori_id = repositori.repositori_id").
		Where("kontributor.nim = ?", nim).
		Where("repositori.jenis_repo = ?", "proyek").
		Count(&proyekCount)

	result.JumlahArtikel = int(articleCount)
	result.JumlahPrestasi = int(prestasiCount)
	result.JumlahIlmiah = int(ilmiahCount)
	result.JumlahProyek = int(proyekCount)

	c.JSON(200, result)
}

func (rc *RepositoriController) CountAllRepositori(c *gin.Context) {
	var result struct {
		Total int `json:"total"`
	}

	var hitung int64

	if err := rc.DB.Table("repositori").Count(&hitung).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result.Total = int(hitung)

	c.JSON(http.StatusOK, result)
}

func (rc *RepositoriController) AllRepositoriOverview(c *gin.Context) {
	filter := c.Param("filter")

	var result []struct {
		JudulRepo string `json:"judul_repo"`
		JenisRepo string `json:"jenis_repo"`
	}

	db := rc.DB.Table("repositori").Select("judul_repo, jenis_repo")

	switch filter {
	case "latest":
		db = db.Order("created_at DESC")
	case "oldest":
		db = db.Order("created_at ASC")
	default:

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "filter harus 'latest' atau 'oldest'",
		})
		return
	}

	if err := db.Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}


func (rc *RepositoriController) GetAllRepositoriByPeriode(c *gin.Context) {
	periodeId := c.Param("periode_id")
	var rows []RepoRow

	err := rc.DB.Table("repositori").
		Select(`
            repositori.repositori_id AS repo_id,
            repositori.judul_repo,
            repositori.deskripsi,
            repositori.photo_url,
            repositori.jenis_repo,
            dokumentasi.link,
            admin.nim,
            admin.nama
        `).
		Joins("LEFT JOIN kontributor ON kontributor.repositori_id = repositori.repositori_id").
		Joins("LEFT JOIN admin ON admin.nim = kontributor.nim").
		Joins("LEFT JOIN dokumentasi ON dokumentasi.repositori_id = repositori.repositori_id").
		Joins("LEFT JOIN pengurus ON pengurus.nim = admin.nim").
		Joins("LEFT JOIN periode ON periode.periode_id = pengurus.periode_id").
		Where("periode.periode_id = ?", periodeId).
		Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	repoMap := make(map[string]*Data)

	for _, r := range rows {
		if _, exists := repoMap[r.RepoId]; !exists {
			repoMap[r.RepoId] = &Data{
				RepoId:       r.RepoId,
				JudulRepo:    r.JudulRepo,
				Deskripsi:    r.Deskripsi,
				PhotoUrl:     r.PhotoUrl,
				JenisRepo:    r.JenisRepo,
				Links:        []string{},
				Contributors: []Names{},
			}
		}

		repoData := repoMap[r.RepoId]

		if !contains(repoData.Links, r.Link) {
			repoData.Links = append(repoData.Links, r.Link)
		}

		if !hasContributor(repoData.Contributors, r.Nim) {
			repoData.Contributors = append(repoData.Contributors, Names{
				Nim:  r.Nim,
				Nama: r.Nama,
			})
		}
	}

	final := []Data{}
	for _, v := range repoMap {
		final = append(final, *v)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": final,
	})
}