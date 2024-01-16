package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/subroll/sqetest/internal/pkg/log"
	"go.uber.org/zap"
)

type (
	User struct {
		userSvc UserService
	}

	OTPRequest struct {
		UserID string `json:"user_id" validate:"required,uuid4"`
	}

	OTPResponse struct {
		UserID string `json:"user_id"`
		OTP    string `json:"otp"`
	}

	ValidateOTPRequest struct {
		UserID string `json:"user_id" validate:"required,uuid4"`
		OTP    string `json:"otp" validate:"required"`
		ReqID  string `json:"request_id" validate:"required"`
	}

	ValidateOTPResponse struct {
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}
)

func NewUser(deps Dependencies) *User {
	return &User{
		userSvc: deps.User,
	}
}

func (u *User) RequestOTP(c echo.Context) error {
	var otpReq OTPRequest
	if err := c.Bind(&otpReq); err != nil {
		log.Warn("fail to bind request", zap.Error(err))

		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	ctx := c.Request().Context()

	otp, err := u.userSvc.GenerateOTP(ctx, otpReq.UserID,
		c.Response().Header().Get(echo.HeaderXRequestID))
	if err != nil {
		log.Error("fail to generate otp", zap.Error(err))

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.JSON(http.StatusOK, OTPResponse{
		UserID: otpReq.UserID,
		OTP:    otp,
	})
}

func (u *User) ValidateOTP(c echo.Context) error {
	var validateOTPReq ValidateOTPRequest
	if err := c.Bind(&validateOTPReq); err != nil {
		log.Warn("fail to bind request", zap.Error(err))

		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	ctx := c.Request().Context()

	if err := u.userSvc.ValidateOTP(ctx, validateOTPReq.UserID, validateOTPReq.OTP,
		validateOTPReq.ReqID); err != nil {
		log.Error("fail to validate otp", zap.Error(err))

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.JSON(http.StatusOK, ValidateOTPResponse{
		UserID:  validateOTPReq.UserID,
		Message: "OTP validated successfully.",
	})
}
