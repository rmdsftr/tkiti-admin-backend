package middleware

import (
	"admin-panel/internal/models"
	"errors"
	"sync"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PeriodeMiddleware struct {
	DB *gorm.DB

	cacheMutex sync.RWMutex
	cacheID    string
	cacheTime  time.Time
	ttl        time.Duration
}

func NewPeriodeMiddleware(db *gorm.DB) *PeriodeMiddleware {
	return &PeriodeMiddleware{
		DB:  db,
		ttl: 10 * time.Second,
	}
}

func (pm *PeriodeMiddleware) GetPeriodeIdActive() gin.HandlerFunc {
	return func(c *gin.Context) {

		pm.cacheMutex.RLock()
		if time.Since(pm.cacheTime) < pm.ttl && pm.cacheID != "" {
			c.Set("periode_id", pm.cacheID)
			pm.cacheMutex.RUnlock()
			c.Next()
			return
		}
		pm.cacheMutex.RUnlock()

		var periode models.Periode
		err := pm.DB.Where("status_periode = ?", models.PeriodeAktif).First(&periode).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mendapatkan periode aktif",
			})
			c.Abort()
			return
		}

		periodeId := periode.PeriodeId

		pm.cacheMutex.Lock()
		pm.cacheID = periodeId
		pm.cacheTime = time.Now()
		pm.cacheMutex.Unlock()

		c.Set("periode_id", periodeId)
		c.Next()
	}
}
