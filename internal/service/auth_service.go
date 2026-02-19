// internal/service/auth_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/teten-nugraha/goboot/pkg/logging"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/teten-nugraha/goboot/internal/domain"
	"github.com/teten-nugraha/goboot/internal/repository"
	"github.com/teten-nugraha/goboot/pkg/data"
	"github.com/teten-nugraha/goboot/pkg/security"
)

type AuthService struct {
	db     *gorm.DB
	users  *repository.UserRepo
	jwtCfg security.JWTConfig
	ttlMin int
	tx     data.TxManager
	log    *zap.Logger
}

func NewAuthService(db *gorm.DB, jwtCfg security.JWTConfig, ttlMin int, tx data.TxManager, logger *zap.Logger) *AuthService {
	return &AuthService{
		db:     db,
		users:  repository.NewUserRepo(db),
		jwtCfg: jwtCfg,
		ttlMin: ttlMin,
		tx:     tx,
		log:    logger.Named("auth_service"),
	}
}

func (s *AuthService) Register(ctx context.Context, email, name, password string) (*domain.User, error) {
	// logging
	ll := logging.FromCtx(ctx, s.log).With(
		zap.String("email", email),
	)
	defer logging.Track(ll, "register")

	if s.db == nil {
		ll.Error("db is not configured")
		return nil, errors.New("database not configured")
	}

	var created *domain.User
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		ll.Debug("checking existing user")
		exists, err := s.users.ExistsByEmail(ctx, email)
		if err != nil {
			ll.Error("existsByEmail failed", zap.Error(err))
			return err
		}
		if exists {
			ll.Error("email already registered")
			return fmt.Errorf("email already registered")
		}

		ll.Debug("hashing password")
		startHash := time.Now()
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			ll.Error("hashing password failed", zap.Error(err))
			return err
		}
		ll.Debug("hashing done", zap.Duration("elapsed", time.Since(startHash)))

		u := &domain.User{
			Email:        email,
			Name:         name,
			PasswordHash: string(hash),
		}
		if err := s.users.Create(ctx, u); err != nil {
			ll.Error("create user failed", zap.Error(err))
			return err
		}

		// Contoh: create profile/log audit di repo lain → tetap 1 transaksi
		// err = s.auditRepo.Log(ctx, ...)
		// if err != nil { return err }

		created = u
		ll.Info("user registered", zap.Uint("uid", u.ID))
		return nil
	}, &data.TxOptions{
		Propagation: data.PropagationRequired,
		ReadOnly:    false,
		// Isolation: sql.LevelReadCommitted,
		// Timeout:  5 * time.Second,
	})
	return created, err
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	if s.db == nil {
		return nil, errors.New("database not configured")
	}
	// Authenticate cukup non-tx (read-only), gunakan SUPPORTS
	var user *domain.User
	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		u, err := s.users.FindByEmail(ctx, email)
		if err != nil {
			return err
		}
		if u == nil {
			return fmt.Errorf("invalid credentials")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
			return fmt.Errorf("invalid credentials")
		}
		user = u
		return nil
	}, &data.TxOptions{
		Propagation: data.PropagationSupports, // pakai tx jika ada; kalau tidak, non-tx
		ReadOnly:    true,
	})
	return user, err
}

func (s *AuthService) IssueToken(u *domain.User) (security.TokenResult, error) {
	claims := map[string]any{
		"email": u.Email,
		"name":  u.Name,
		"uid":   strconv.FormatUint(uint64(u.ID), 10),
	}
	return security.GenerateJWT(s.jwtCfg, strconv.FormatUint(uint64(u.ID), 10), claims, s.ttlMin)
}
