package service

import (
	"context"
)

const (
	otpLength = 5
)

type (
	User struct {
		userRepo     UserRepository
		otpGenerator func(uint8) (string, error)
	}
)

func NewUser(deps Dependencies) *User {
	return &User{
		userRepo:     deps.User,
		otpGenerator: deps.RandNumberGenerator,
	}
}

func (u *User) GenerateOTP(ctx context.Context, userUUID, requestID string) (string, error) {
	userID, err := u.userRepo.GetUserIDByUUID(ctx, userUUID)
	if err != nil {
		return "", err
	}

	otp, err := u.otpGenerator(otpLength)
	if err != nil {
		return "", err
	}

	if err := u.userRepo.StoreOTP(ctx, userID, otp, requestID); err != nil {
		return "", err
	}

	return otp, nil
}

func (u *User) ValidateOTP(ctx context.Context, userUUID, otp, requestID string) error {
	userID, err := u.userRepo.GetUserIDByUUID(ctx, userUUID)
	if err != nil {
		return err
	}

	return u.userRepo.UpdateOTPStatus(ctx, userID, otp, requestID)
}
