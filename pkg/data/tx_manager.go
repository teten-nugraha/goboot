// pkg/data/tx_manager.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Propagation int

const (
	PropagationRequired     Propagation = iota // REQUIRED: pakai tx yang ada, kalau belum ada → buat baru
	PropagationRequiresNew                     // REQUIRES_NEW: selalu buat tx baru (suspend tx lama)
	PropagationSupports                        // SUPPORTS: pakai tx jika ada; kalau tidak ada → non-tx
	PropagationMandatory                       // MANDATORY: harus ada tx, kalau tidak → error
	PropagationNotSupported                    // NOT_SUPPORTED: non-tx walaupun ada tx
	PropagationNever                           // NEVER: error kalau ada tx aktif
	PropagationNested                          // NESTED: pakai savepoint jika ada tx; kalau tidak ada → buat tx baru
)

type TxOptions struct {
	Propagation Propagation
	ReadOnly    bool
	Isolation   sql.IsolationLevel // sql.LevelReadCommitted, dll
	Timeout     time.Duration      // jika > 0, akan membuat context.WithTimeout
	Name        string             // optional, untuk log/savepoint name
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error, opts *TxOptions) error
	Current(ctx context.Context) *gorm.DB
}

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) TxManager {
	return &txManager{db: db}
}

// ---- Context wiring untuk *gorm.DB (tx-aware) ----

type txCtxKey struct{}

func injectTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// TxFromCtx mengembalikan *gorm.DB yang “aktif” (transaksi) jika ada.
func TxFromCtx(ctx context.Context) *gorm.DB {
	if v := ctx.Value(txCtxKey{}); v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx
		}
	}
	return nil
}

// TxDB mengembalikan DB yang tepat: jika ada tx di ctx → pakai tx; jika tidak → gunakan defaultDB.
func TxDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx := TxFromCtx(ctx); tx != nil {
		return tx
	}
	return defaultDB
}

func (m *txManager) Current(ctx context.Context) *gorm.DB {
	return TxFromCtx(ctx)
}

// ---- Implementasi WithTx ----

func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context) error, opts *TxOptions) (err error) {
	if m.db == nil {
		return errors.New("TxManager: underlying DB is nil (check DSN/config)")
	}
	if opts == nil {
		opts = &TxOptions{Propagation: PropagationRequired}
	}
	existing := TxFromCtx(ctx)

	switch opts.Propagation {
	case PropagationNotSupported:
		// Jalankan tanpa transaksi, suspend kalau ada tx aktif
		return fn(injectTx(ctx, nil))
	case PropagationNever:
		if existing != nil {
			return errors.New("TxManager: propagation=NEVER but a transaction exists")
		}
		return fn(ctx)
	case PropagationMandatory:
		if existing == nil {
			return errors.New("TxManager: propagation=MANDATORY but no transaction exists")
		}
		// Tetap pakai existing
		return fn(ctx)
	case PropagationSupports:
		// Pakai existing jika ada; kalau tidak ada non-tx
		return fn(ctx)
	case PropagationRequired:
		if existing != nil {
			// Pakai tx yang ada
			return fn(ctx)
		}
		// Buat transaksi baru
		return m.runNewTx(ctx, fn, opts)
	case PropagationRequiresNew:
		// Selalu buat tx baru (suspend existing)
		return m.runNewTx(injectTx(ctx, nil), fn, opts)
	case PropagationNested:
		if existing == nil {
			// Tidak ada tx → sama dengan REQUIRED (buat tx baru)
			return m.runNewTx(ctx, fn, opts)
		}
		// Ada tx → gunakan savepoint
		return m.runWithSavepoint(ctx, fn, opts)
	default:
		return fmt.Errorf("TxManager: unknown propagation %v", opts.Propagation)
	}
}

func (m *txManager) runNewTx(ctx context.Context, fn func(ctx context.Context) error, opts *TxOptions) (err error) {
	// Timeout context
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Begin tx
	txOpts := &sql.TxOptions{}
	if opts.Isolation != 0 {
		txOpts.Isolation = opts.Isolation
	}
	// readOnly di SQL-level:
	txOpts.ReadOnly = opts.ReadOnly

	tx := m.db.Begin(txOpts)
	if tx.Error != nil {
		return tx.Error
	}

	// Untuk driver yang perlu statement khusus read-only (opsional, misal Postgres),
	// kita bisa tambah:
	// if opts.ReadOnly {
	//     tx.Exec("SET TRANSACTION READ ONLY")
	// }

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()

	// Pastikan semua query di bawah ini memakai ctx & tx
	tx = tx.WithContext(ctx)
	ctx = injectTx(ctx, tx)

	err = fn(ctx)
	return err
}

func (m *txManager) runWithSavepoint(ctx context.Context, fn func(ctx context.Context) error, opts *TxOptions) (err error) {
	existing := TxFromCtx(ctx)
	if existing == nil {
		return errors.New("TxManager: NESTED requires an existing transaction")
	}
	// Nama savepoint
	name := opts.Name
	if name == "" {
		name = "tx_nested"
	}
	if err := existing.SavePoint(name).Error; err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = existing.RollbackTo(name).Error
			panic(p)
		} else if err != nil {
			_ = existing.RollbackTo(name).Error
		}
		// Jika sukses, tidak perlu commit di sini; commit dilakukan oleh tx terluar.
	}()

	// Jalankan fungsi di dalam tx yang sama (existing)
	return fn(ctx)
}
