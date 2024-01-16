package rest

import (
	"context"
)

type Dependencies struct {
	User UserService
}

type UserService interface {
	GenerateOTP(ctx context.Context, userUUID, requestID string) (string, error)
	ValidateOTP(ctx context.Context, userID, otp, requestID string) error
}
