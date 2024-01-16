package service

import (
	"context"
)

type Dependencies struct {
	User UserRepository

	RandNumberGenerator func(uint8) (string, error)
}

type UserRepository interface {
	GetUserIDByUUID(ctx context.Context, uuid string) (uint64, error)
	StoreOTP(ctx context.Context, userID uint64, otp, requestID string) error
	UpdateOTPStatus(ctx context.Context, userID uint64, otp, requestID string) error
}
