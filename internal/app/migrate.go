// internal/app/migrate.go
package app

import (
	"github.com/teten-nugraha/goboot/internal/domain"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	if db != nil {
		db.AutoMigrate(&domain.User{})
	}
}
