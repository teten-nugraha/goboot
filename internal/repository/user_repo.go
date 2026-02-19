// internal/repository/user_repo.go
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/teten-nugraha/goboot/internal/domain"
	"github.com/teten-nugraha/goboot/pkg/data"
)

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	return data.TxDB(ctx, r.db).WithContext(ctx).Create(u).Error
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := data.TxDB(ctx, r.db).WithContext(ctx).
		Model(&domain.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	db := data.TxDB(ctx, r.db).WithContext(ctx)
	if err := db.Where("email = ?", email).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
