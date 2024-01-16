package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	mockrepo "github.com/subroll/sqetest/internal/mocks/service"
)

func TestUser_GenerateOTP(t *testing.T) {
	t.Parallel()

	type arg struct {
		ctx               context.Context
		userID, requestID string
	}

	type expectaion struct {
		otp string
		err error
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, arg, expectaion)
	}{
		{
			desc: "ErrorGetUserIDByUUID",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
				})

				userRepo.On("GetUserIDByUUID", context.TODO(), "fake-uuid").Return(uint64(0), errors.New("fake error"))

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						requestID: "fake-request-id",
					}, expectaion{
						otp: "",
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorGeneratingOTP",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
					RandNumberGenerator: func(uint8) (string, error) {
						return "", errors.New("fake error")
					},
				})

				userRepo.On("GetUserIDByUUID", context.TODO(), "fake-uuid").Return(uint64(1), nil)

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						requestID: "fake-request-id",
					}, expectaion{
						otp: "",
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorStoringOTP",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
					RandNumberGenerator: func(uint8) (string, error) {
						return "xxxxx", nil
					},
				})

				userRepo.On("GetUserIDByUUID", context.TODO(), "fake-uuid").Return(uint64(1), nil)
				userRepo.On("StoreOTP", context.TODO(), uint64(1), "xxxxx", "fake-request-id").Return(errors.New("fake error"))

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						requestID: "fake-request-id",
					}, expectaion{
						otp: "",
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "SuccessGenerateOTP",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
					RandNumberGenerator: func(uint8) (string, error) {
						return "xxxxx", nil
					},
				})

				userRepo.On("GetUserIDByUUID", context.TODO(), "fake-uuid").Return(uint64(1), nil)
				userRepo.On("StoreOTP", context.TODO(), uint64(1), "xxxxx", "fake-request-id").Return(nil)

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						requestID: "fake-request-id",
					}, expectaion{
						otp: "xxxxx",
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

			got, err := u.GenerateOTP(a.ctx, a.userID, a.requestID)
			assert.Equal(t, e.otp, got)
			assert.Equal(t, e.err, err)
		})
	}
}

func TestUser_ValidateOTP(t *testing.T) {
	t.Parallel()

	type arg struct {
		ctx       context.Context
		userID    string
		otp       string
		requestID string
	}

	type expectaion struct {
		err error
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, arg, expectaion)
	}{
		{
			desc: "ErrorGetUserIDByUUID",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
				})

				userRepo.On("GetUserIDByUUID", context.TODO(), "fake-uuid").Return(uint64(0), errors.New("fake error"))

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectaion{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "ErrorValidatingOTP",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
				})

				userRepo.
					On("GetUserIDByUUID", context.TODO(), "fake-uuid").
					Return(uint64(1), nil)

				userRepo.
					On("UpdateOTPStatus", context.TODO(), uint64(1), "xxxxx", "fake-request-id").
					Return(errors.New("fake error"))

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectaion{
						err: errors.New("fake error"),
					}
			},
		},
		{
			desc: "SuccessGenerateOTP",
			mockFn: func(*testing.T) (*User, arg, expectaion) {
				userRepo := mockrepo.NewUserRepository(t)
				user := NewUser(Dependencies{
					User: userRepo,
				})

				userRepo.
					On("GetUserIDByUUID", context.TODO(), "fake-uuid").
					Return(uint64(1), nil)

				userRepo.
					On("UpdateOTPStatus", context.TODO(), uint64(1), "xxxxx", "fake-request-id").
					Return(nil)

				return user, arg{
						ctx:       context.TODO(),
						userID:    "fake-uuid",
						otp:       "xxxxx",
						requestID: "fake-request-id",
					}, expectaion{
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

			err := u.ValidateOTP(a.ctx, a.userID, a.otp, a.requestID)
			assert.Equal(t, e.err, err)
		})
	}
}
