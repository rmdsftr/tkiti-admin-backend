package controllers

import (
	"admin-panel/internal/middleware"
	"admin-panel/internal/models"
	"admin-panel/internal/services"
	"admin-panel/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleController struct {
	DB             *gorm.DB
	StorageService *services.StorageService
}

func NewArticleController(db *gorm.DB, storageService *services.StorageService) *ArticleController {
	return &ArticleController{
		DB:             db,
		StorageService: storageService,
	}
}

func (ac *ArticleController) CreateArticle(c *gin.Context) {
	nim := c.Param("nim")
	status := c.Param("status")

	var anggota models.Admin

	if err := ac.DB.First(&anggota, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin tidak ditemukan"})
		return
	}

	if status != string(models.Published) && status != string(models.Draft) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status tidak valid (published/draft)",
		})
		return
	}

	var input struct {
		Judul   string `form:"judul" binding:"required"`
		Content string `form:"content" binding:"required"`
		Tags    string `form:"tags"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("photo_url")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Artikel harus memiliki foto"})
		return
	}

	if !middleware.IsValidImageType(file.Header.Get("Content-Type")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipe gambar tidak valid"})
		return
	}

	if file.Size > 25*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran gambar terlalu besar"})
		return
	}

	photourl, err := ac.StorageService.UploadFile(file, "artikel")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx := ac.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	artikel := models.Article{
		Judul:         input.Judul,
		Content:       input.Content,
		Nim:           nim,
		PhotoUrl:      photourl,
		StatusArticle: models.ArticleEnum(status),
	}

	if err := tx.Create(&artikel).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.Tags != "" {
		tagNames := strings.Split(input.Tags, ",")
		var tagIds []int
		seenTags := make(map[string]bool)

		for _, tagName := range tagNames {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			normalizedTag := utils.NormalizeTag(tagName)
			if seenTags[normalizedTag] {
				continue
			}
			seenTags[normalizedTag] = true

			var tag models.Tags

			err := tx.Where("tag = ?", normalizedTag).First(&tag).Error
			if err == gorm.ErrRecordNotFound {

				tag = models.Tags{
					Tag:          normalizedTag,
					TotalArticle: 1,
				}
				if err := tx.Create(&tag).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "gagal membuat tag",
						"details": err.Error(),
					})
					return
				}
			} else if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "gagal mencari tag",
					"details": err.Error(),
				})
				return
			} else {

				if err := tx.Model(&tag).Update("total_article", gorm.Expr("total_article + ?", 1)).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "gagal update tag",
						"details": err.Error(),
					})
					return
				}
			}

			tagIds = append(tagIds, tag.TagId)
		}

		if len(tagIds) > 0 {
			for _, tagId := range tagIds {
				articleTag := models.ArticleTag{
					ArticleId: artikel.ArticleId,
					TagId:     tagId,
				}
				if err := tx.Create(&articleTag).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "gagal menambahkan tags ke artikel",
						"details": err.Error(),
					})
					return
				}
			}
		}

		if err := tx.Preload("Tags").First(&artikel, "article_id = ?", artikel.ArticleId).Error; err != nil {

			c.JSON(http.StatusCreated, gin.H{
				"message": "Artikel baru berhasil ditambahkan (tanpa data tags)",
				"data":    artikel,
			})
			tx.Commit()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Artikel baru berhasil ditambahkan",
		"data":    artikel,
	})
}

func (ac *ArticleController) GetArtikelWithFilter(c *gin.Context) {
	status := c.Param("status")

	if status != string(models.Published) && status != string(models.Draft) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status tidak valid (published/draft)",
		})
		return
	}

	type ArticleResponse struct {
		ArticleId     string             `json:"article_id"`
		Judul         string             `json:"judul"`
		Content       string             `json:"content"`
		PhotoUrl      string             `json:"photo_url"`
		Views         int                `json:"views"`
		StatusArticle models.ArticleEnum `json:"status_article"`
		UpdatedAt     string             `json:"updated_at"`
		Nim           string             `json:"nim"`
		Nama          string             `json:"nama"`
		Tags          string             `json:"tag"`
	}
	var articles []models.Article

	err := ac.DB.
		Preload("Tags").
		Joins("JOIN admin ON admin.nim = article.nim").
		Where("status_article = ?", status).
		Order("updated_at DESC").
		Find(&articles).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var result []ArticleResponse

	for _, a := range articles {

		tagNames := ""
		for i, t := range a.Tags {
			if i > 0 {
				tagNames += ", "
			}
			tagNames += t.Tag
		}

		var admin models.Admin
		ac.DB.First(&admin, "nim = ?", a.Nim)

		result = append(result, ArticleResponse{
			ArticleId:     a.ArticleId,
			Judul:         a.Judul,
			Content:       a.Content,
			PhotoUrl:      a.PhotoUrl,
			Views:         a.Views,
			StatusArticle: a.StatusArticle,
			UpdatedAt:     a.UpdatedAt.Format("2006-01-02 15:04:05"),
			Nim:           a.Nim,
			Nama:          admin.Nama,
			Tags:          tagNames,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (ac *ArticleController) DeleteArticle(c *gin.Context) {
	id := c.Param("article_id")

	var article models.Article
	if err := ac.DB.First(&article, "article_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if article.PhotoUrl != "" {
		filePath := ac.StorageService.ExtractFilePathFromURL(article.PhotoUrl)
		if filePath != "" {
			if err := ac.StorageService.DeleteFile(filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
	}

	if err := ac.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus artikel",
	})
}

func (ac *ArticleController) GetArtikelByAslab(c *gin.Context) {
	status := c.Param("status")
	nim := c.Param("nim")

	if status != string(models.Published) && status != string(models.Draft) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status tidak valid (published/draft)",
		})
		return
	}

	var admin models.Admin
	if err := ac.DB.First(&admin, "nim = ?", nim).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "aslab tidak ditemukan",
		})
		return
	}

	var articles []models.Article
	if err := ac.DB.
		Preload("Tags").
		Where("status_article = ?", status).
		Where("nim = ?", nim).
		Order("updated_at DESC").
		Find(&articles).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	type ArticleResponse struct {
		ArticleId     string `json:"article_id"`
		Judul         string `json:"judul"`
		Content       string `json:"content"`
		PhotoUrl      string `json:"photo_url"`
		Views         int    `json:"views"`
		StatusArticle string `json:"status_article"`
		UpdatedAt     string `json:"updated_at"`
		Nim           string `json:"nim"`
		Nama          string `json:"nama"`
		Tags          string `json:"tag"`
	}

	var result []ArticleResponse

	for _, a := range articles {

		var tags string
		for i, t := range a.Tags {
			if i > 0 {
				tags += ", "
			}
			tags += t.Tag
		}

		result = append(result, ArticleResponse{
			ArticleId:     a.ArticleId,
			Judul:         a.Judul,
			Content:       a.Content,
			PhotoUrl:      a.PhotoUrl,
			Views:         a.Views,
			StatusArticle: string(a.StatusArticle),
			UpdatedAt:     a.UpdatedAt.Format("2006-01-02 15:04:05"),
			Nim:           nim,
			Nama:          admin.Nama,
			Tags:          tags,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (ac *ArticleController) GetArtikelById(c *gin.Context) {
	articleId := c.Param("article_id")

	if articleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "article_id diperlukan",
		})
		return
	}

	type ArticleResponse struct {
		ArticleId     string             `json:"article_id"`
		Judul         string             `json:"judul"`
		Content       string             `json:"content"`
		PhotoUrl      string             `json:"photo_url"`
		Views         int                `json:"views"`
		StatusArticle models.ArticleEnum `json:"status_article"`
		UpdatedAt     string             `json:"updated_at"`
		Nim           string             `json:"nim"`
		Nama          string             `json:"nama"`
		Tags          string             `json:"tag"`
	}

	var article models.Article

	err := ac.DB.
		Preload("Tags").
		Joins("JOIN admin ON admin.nim = article.nim").
		Where("article.article_id = ?", articleId).
		First(&article).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "artikel tidak ditemukan",
		})
		return
	}

	tagNames := ""
	for i, t := range article.Tags {
		if i > 0 {
			tagNames += ", "
		}
		tagNames += t.Tag
	}

	var admin models.Admin
	ac.DB.First(&admin, "nim = ?", article.Nim)

	response := ArticleResponse{
		ArticleId:     article.ArticleId,
		Judul:         article.Judul,
		Content:       article.Content,
		PhotoUrl:      article.PhotoUrl,
		Views:         article.Views,
		StatusArticle: article.StatusArticle,
		UpdatedAt:     article.UpdatedAt.Format("2006-01-02 15:04:05"),
		Nim:           article.Nim,
		Nama:          admin.Nama,
		Tags:          tagNames,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (ac *ArticleController) UpdateArticle(c *gin.Context) {
	articleId := c.Param("article_id")

	var artikel models.Article
	if err := ac.DB.Preload("Tags").First(&artikel, "article_id = ?", articleId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artikel tidak ditemukan"})
		return
	}

	var input struct {
		Judul   string `form:"judul" binding:"required"`
		Content string `form:"content" binding:"required"`
		Tags    string `form:"tags"`
		Status  string `form:"status"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := input.Status
	if status == "" {
		status = string(models.Published)
	}
	if status != string(models.Published) && status != string(models.Draft) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status tidak valid (published/draft)",
		})
		return
	}

	file, err := c.FormFile("photo_url")
	var newPhotoUrl string

	if err == nil {

		if !middleware.IsValidImageType(file.Header.Get("Content-Type")) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tipe gambar tidak valid"})
			return
		}

		if file.Size > 25*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran gambar terlalu besar"})
			return
		}

		newPhotoUrl, err = ac.StorageService.UploadFile(file, "artikel")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx := ac.DB.Begin()

	artikel.Judul = input.Judul
	artikel.Content = input.Content
	artikel.StatusArticle = models.ArticleEnum(status)

	if newPhotoUrl != "" {
		artikel.PhotoUrl = newPhotoUrl
	}

	if err := tx.Save(&artikel).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Where("article_id = ?", artikel.ArticleId).Delete(&models.ArticleTag{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus tag lama"})
		return
	}

	if input.Tags != "" {
		tagNames := strings.Split(input.Tags, ",")
		var tagIds []int
		seen := make(map[string]bool)

		for _, tagName := range tagNames {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			normalized := utils.NormalizeTag(tagName)
			if seen[normalized] {
				continue
			}
			seen[normalized] = true

			var tag models.Tags
			err := tx.Where("tag = ?", normalized).First(&tag).Error

			if err == gorm.ErrRecordNotFound {

				tag = models.Tags{Tag: normalized, TotalArticle: 1}
				if err := tx.Create(&tag).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tag"})
					return
				}
			} else if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari tag"})
				return
			} else {

				tx.Model(&tag).Update("total_article", gorm.Expr("total_article + 1"))
			}

			tagIds = append(tagIds, tag.TagId)
		}

		for _, tg := range tagIds {
			at := models.ArticleTag{
				ArticleId: artikel.ArticleId,
				TagId:     tg,
			}
			if err := tx.Create(&at).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan tag"})
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Artikel berhasil diperbarui",
		"data":    artikel,
	})
}
