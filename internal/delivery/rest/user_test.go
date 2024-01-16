package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mocksvc "github.com/subroll/sqetest/internal/mocks/delivery/rest"
)

func TestUser_RequestOTP(t *testing.T) {
	t.Parallel()

	type expectaion struct {
		httpStatus int
		response   string
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion)
	}{
		{
			desc: "ErrorBindingRequest",
			mockFn: func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion) {
				user := NewUser(Dependencies{})

				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/otp/request", strings.NewReader(" "))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				return user, c, rec, expectaion{
					httpStatus: http.StatusBadRequest,
					response:   "Bad Request",
				}
			},
		},
		{
			desc: "ErrorGenerateOTP",
			mockFn: func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion) {
				userSvc := mocksvc.NewUserService(t)
				user := NewUser(Dependencies{
					User: userSvc,
				})

				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/otp/request", strings.NewReader(`{"user_id":"fake-uuid"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				rec.Header().Set(echo.HeaderXRequestID, "fake-request-id")
				c := e.NewContext(req, rec)
				ctx := c.Request().Context()

				userSvc.On("GenerateOTP", ctx, "fake-uuid", "fake-request-id").Return("", errors.New("fake error"))

				return user, c, rec, expectaion{
					httpStatus: http.StatusInternalServerError,
					response:   "Internal Server Error",
				}
			},
		},
		{
			desc: "SuccessGenerateOTP",
			mockFn: func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion) {
				userSvc := mocksvc.NewUserService(t)
				user := NewUser(Dependencies{
					User: userSvc,
				})

				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/otp/request", strings.NewReader(`{"user_id":"fake-uuid"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				rec.Header().Set(echo.HeaderXRequestID, "fake-request-id")
				c := e.NewContext(req, rec)
				ctx := c.Request().Context()

				userSvc.On("GenerateOTP", ctx, "fake-uuid", "fake-request-id").Return("12345", nil)

				return user, c, rec, expectaion{
					httpStatus: http.StatusOK,
					response: `{"user_id":"fake-uuid","otp":"12345"}
`,
				}
			},
		},
	}

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			u, c, rec, exp := tC.mockFn(t)
			err := u.RequestOTP(c)
			if err != nil {
				echoError := err.(*echo.HTTPError)

				assert.Equal(t, fmt.Sprintf("code=%d, message=%v", exp.httpStatus, exp.response), echoError.Error())
			} else {
				assert.Equal(t, exp.httpStatus, rec.Code)
				assert.Equal(t, exp.response, rec.Body.String())
			}
		})
	}
}

func TestUser_ValidateOTP(t *testing.T) {
	t.Parallel()

	type expectaion struct {
		httpStatus int
		response   string
	}

	testCases := []struct {
		desc   string
		mockFn func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion)
	}{
		{
			desc: "ErrorBindingRequest",
			mockFn: func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion) {
				user := NewUser(Dependencies{})

				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/otp/validate", strings.NewReader(" "))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				return user, c, rec, expectaion{
					httpStatus: http.StatusBadRequest,
					response:   "Bad Request",
				}
			},
		},
		{
			desc: "ErrorValidateOTP",
			mockFn: func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion) {
				userSvc := mocksvc.NewUserService(t)
				user := NewUser(Dependencies{
					User: userSvc,
				})

				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/otp/validate", strings.NewReader(`{"user_id":"fake-uuid","otp":"12345","request_id":"fake-request-id"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				ctx := c.Request().Context()

				userSvc.On("ValidateOTP", ctx, "fake-uuid", "12345", "fake-request-id").Return(errors.New("fake error"))

				return user, c, rec, expectaion{
					httpStatus: http.StatusInternalServerError,
					response:   "Internal Server Error",
				}
			},
		},
		{
			desc: "SuccessValidateOTP",
			mockFn: func(*testing.T) (*User, echo.Context, *httptest.ResponseRecorder, expectaion) {
				userSvc := mocksvc.NewUserService(t)
				user := NewUser(Dependencies{
					User: userSvc,
				})

				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/otp/validate", strings.NewReader(`{"user_id":"fake-uuid","otp":"12345","request_id":"fake-request-id"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				rec.Header().Set(echo.HeaderXRequestID, "fake-request-id")
				c := e.NewContext(req, rec)
				ctx := c.Request().Context()

				userSvc.On("ValidateOTP", ctx, "fake-uuid", "12345", "fake-request-id").Return(nil)

				return user, c, rec, expectaion{
					httpStatus: http.StatusOK,
					response: `{"user_id":"fake-uuid","message":"OTP validated successfully."}
`,
				}
			},
		},
	}

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			u, c, rec, exp := tC.mockFn(t)
			err := u.ValidateOTP(c)
			if err != nil {
				echoError := err.(*echo.HTTPError)

				assert.Equal(t, fmt.Sprintf("code=%d, message=%v", exp.httpStatus, exp.response), echoError.Error())
			} else {
				assert.Equal(t, exp.httpStatus, rec.Code)
				assert.Equal(t, exp.response, rec.Body.String())
			}
		})
	}
}
