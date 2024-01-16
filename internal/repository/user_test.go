package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func createDBMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	return db, mock
}

func TestNewUser(t *testing.T) {
	db, _ := createDBMock(t)
	nowFunc := func() time.Time {
		return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
	}

	exp := &User{
		db:      db,
		nowFunc: nowFunc,
	}

	got := NewUser(Dependencies{
		DB:      db,
		NowFunc: nowFunc,
	})

	assert.NotNil(t, exp, got)
	assert.NotNil(t, exp.nowFunc, got.nowFunc)
	assert.Equal(t, exp.db, got.db)
}

func TestUser_GetUserIDByUUID(t *testing.T) {
	t.Parallel()

	type arg struct {
		ctx  context.Context
		uuid string
	}

	type expectation struct {
		id  uint64
		err error
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, arg, expectation)
	}{
		{
			desc: "ErrorSQL",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectQuery(`SELECT id FROM users WHERE uuid = \?;`).
					WithArgs("fake-uuid").
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
					}, arg{
						ctx:  context.TODO(),
						uuid: "fake-uuid",
					}, expectation{
						id:  0,
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorNotFound",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectQuery(`SELECT id FROM users WHERE uuid = \?;`).
					WithArgs("fake-uuid").
					WillReturnError(sql.ErrNoRows)

				return &User{
						db: db,
					}, arg{
						ctx:  context.TODO(),
						uuid: "fake-uuid",
					}, expectation{
						id:  0,
						err: ErrNotFound,
					}
			},
		},
		{
			desc: "Success",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectQuery(`SELECT id FROM users WHERE uuid = \?;`).
					WithArgs("fake-uuid").
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).
							AddRow(1))

				return &User{
						db: db,
					}, arg{
						ctx:  context.TODO(),
						uuid: "fake-uuid",
					}, expectation{
						id:  1,
						err: nil,
					}
			},
		},
	}

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			u, a, e := tC.mockFn(t)

			got, err := u.GetUserIDByUUID(a.ctx, a.uuid)
			assert.Equal(t, got, e.id)
			assert.Equal(t, err, e.err)
		})
	}
}

func TestUser_StoreOTP(t *testing.T) {
	t.Parallel()

	type arg struct {
		ctx            context.Context
		userID         uint64
		otp, requestID string
	}

	type expectation struct {
		err error
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, arg, expectation)
	}{
		{
			desc: "ErrorStartTx",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin().
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorGetOTPData",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorCommitIfOTPAlreadyExist",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 2, 0, 0, time.Local)))

				mock.
					ExpectCommit().
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorOTPAlreadyExist",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 2, 0, 0, time.Local)))

				mock.
					ExpectCommit()

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: ErrOTPExist,
					}
			},
		},
		{
			desc: "ErrorExpiringExistingOTP",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusExpired, uint64(1)).
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorInsertingNewOTP",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusExpired, uint64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectExec(`INSERT INTO otps \(user_id, otp, request_id, expired_at\) VALUES \(\?, \?, \?, \?\);`).
					WithArgs(uint64(1), "xxxxx", "fake-request-id", time.Date(2024, time.January, 1, 0, 6, 0, 0, time.Local)).
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorCommittingNewOTP",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusExpired, uint64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectExec(`INSERT INTO otps \(user_id, otp, request_id, expired_at\) VALUES \(\?, \?, \?, \?\);`).
					WithArgs(uint64(1), "xxxxx", "fake-request-id", time.Date(2024, time.January, 1, 0, 6, 0, 0, time.Local)).
					WillReturnResult(sqlmock.NewResult(2, 1))

				mock.
					ExpectCommit().
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "SuccessStoringOTP",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusExpired, uint64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectExec(`INSERT INTO otps \(user_id, otp, request_id, expired_at\) VALUES \(\?, \?, \?, \?\);`).
					WithArgs(uint64(1), "xxxxx", "fake-request-id", time.Date(2024, time.January, 1, 0, 6, 0, 0, time.Local)).
					WillReturnResult(sqlmock.NewResult(2, 1))

				mock.
					ExpectCommit()

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{}
			},
		},
	}

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			u, a, e := tC.mockFn(t)

			err := u.StoreOTP(a.ctx, a.userID, a.otp, a.requestID)
			assert.Equal(t, err, e.err)
		})
	}
}

func TestUser_UpdateOTPStatus(t *testing.T) {
	t.Parallel()

	type arg struct {
		ctx            context.Context
		userID         uint64
		otp, requestID string
	}

	type expectation struct {
		err error
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, arg, expectation)
	}{
		{
			desc: "ErrorStartTx",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin().
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorOTPNotFound",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnError(sql.ErrNoRows)

				return &User{
						db: db,
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: ErrInvalidOTP,
					}
			},
		},
		{
			desc: "ErrorGetOTPData",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorCommitIfOTPExpired",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local)))

				mock.
					ExpectCommit().
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorOTPAlreadyExpired",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 0, 0, 0, time.Local)))

				mock.
					ExpectCommit()

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: ErrOTPExpired,
					}
			},
		},
		{
			desc: "ErrorUpdatingOTPStatus",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 2, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusUsed, uint64(1)).
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorCommittingOTPStatus",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 2, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusUsed, uint64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectCommit().
					WillReturnError(errors.New("fake error"))

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "SuccessUpdatingOTPStatus",
			mockFn: func(*testing.T) (*User, arg, expectation) {
				db, mock := createDBMock(t)

				mock.
					ExpectBegin()

				mock.
					ExpectQuery(`SELECT id, expired_at FROM otps WHERE user_id = \? AND otp = \? AND status = \? FOR UPDATE;`).
					WithArgs(uint64(1), "xxxxx", otpStatusUnused).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "expired_at"}).
							AddRow(uint64(1), time.Date(2024, time.January, 1, 0, 2, 0, 0, time.Local)))

				mock.
					ExpectExec(`UPDATE otps SET status = \? WHERE id = \?;`).
					WithArgs(otpStatusUsed, uint64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectCommit()

				return &User{
						db: db,
						nowFunc: func() time.Time {
							return time.Date(2024, time.January, 1, 0, 1, 0, 0, time.Local)
						},
					}, arg{
						ctx:       context.TODO(),
						userID:    1,
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectation{}
			},
		},
	}

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			u, a, e := tC.mockFn(t)

			err := u.UpdateOTPStatus(a.ctx, a.userID, a.otp, a.requestID)
			assert.Equal(t, err, e.err)
		})
	}
}
