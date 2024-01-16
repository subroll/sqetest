package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("data not found")
	ErrOTPExist   = errors.New("there is still an active otp")
	ErrOTPExpired = errors.New("otp expired")
	ErrInvalidOTP = errors.New("invalid otp")
)

const (
	otpStatusUnused = iota
	otpStatusUsed
	otpStatusExpired
)

type (
	User struct {
		db      *sql.DB
		nowFunc func() time.Time
	}
)

func NewUser(deps Dependencies) *User {
	return &User{
		db:      deps.DB,
		nowFunc: deps.NowFunc,
	}
}

func (u *User) GetUserIDByUUID(ctx context.Context, uuid string) (uint64, error) {
	var id uint64
	if err := u.db.QueryRowContext(ctx, `SELECT id FROM users WHERE uuid = ?;`, uuid).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}

		return 0, err
	}

	return id, nil
}

func (u *User) StoreOTP(ctx context.Context, userID uint64, otp, requestID string) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		uid       uint64
		expiredAt time.Time
	)
	if err := tx.QueryRowContext(ctx, `SELECT id, expired_at FROM otps WHERE user_id = ? AND status = ? FOR UPDATE;`,
		userID, otpStatusUnused).Scan(&uid, &expiredAt); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	if uid > 0 && expiredAt.After(u.nowFunc()) {
		if err := tx.Commit(); err != nil {
			return err
		}

		return ErrOTPExist
	}

	if _, err := tx.ExecContext(ctx, `UPDATE otps SET status = ? WHERE id = ?;`, otpStatusExpired, uid); err != nil {
		return err
	}

	fiveMinFromNow := u.nowFunc().Add(5 * time.Minute)
	if _, err := tx.ExecContext(ctx, `INSERT INTO otps (user_id, otp, request_id, expired_at) VALUES (?, ?, ?, ?);`,
		userID, otp, requestID, fiveMinFromNow); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *User) UpdateOTPStatus(ctx context.Context, userID uint64, otp, requestID string) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		uid       uint64
		expiredAt time.Time
	)
	if err := tx.QueryRowContext(ctx, `SELECT id, expired_at FROM otps WHERE user_id = ? AND otp = ? AND status = ? FOR UPDATE;`,
		userID, otp, otpStatusUnused).Scan(&uid, &expiredAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidOTP
		}

		return err
	}

	if expiredAt.Before(u.nowFunc()) {
		if err := tx.Commit(); err != nil {
			return err
		}

		return ErrOTPExpired
	}

	if _, err := tx.ExecContext(ctx, `UPDATE otps SET status = ? WHERE id = ?;`, otpStatusUsed, uid); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
