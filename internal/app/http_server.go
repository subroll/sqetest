package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/subroll/sqetest/internal/delivery/rest"
	"github.com/subroll/sqetest/internal/pkg/config"
	"github.com/subroll/sqetest/internal/pkg/log"
	"github.com/subroll/sqetest/internal/pkg/stringutil"
	"github.com/subroll/sqetest/internal/repository"
	"github.com/subroll/sqetest/internal/service"
	"go.uber.org/zap"
)

type (
	reqValidator struct {
		v *validator.Validate
	}

	HTTPServer struct {
		server *echo.Echo
		v      *validator.Validate
		db     *sql.DB

		pingHandler echo.HandlerFunc
		userHandler *rest.User

		userSvc *service.User

		userRepo *repository.User
	}
)

func (rv *reqValidator) Validate(i interface{}) error {
	if err := rv.v.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (hs *HTTPServer) Start() error {
	err := hs.server.Start(viper.GetString(config.HTTPPort))
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (hs *HTTPServer) Stop(ctx context.Context) error {
	if err := hs.server.Shutdown(ctx); err != nil {
		return err
	}

	if err := hs.db.Close(); err != nil {
		return err
	}

	return nil
}

func (hs *HTTPServer) route() {
	hs.server.Use(middleware.RequestID())
	hs.server.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			var logFn func(string, ...zap.Field)

			switch {
			case v.Status >= http.StatusBadRequest && v.Status < http.StatusInternalServerError:
				logFn = log.Warn
			case v.Status >= http.StatusInternalServerError:
				logFn = log.Error
			default:
				logFn = log.Info
			}

			logFn("access log",
				zap.String("latency", v.Latency.String()),
				zap.String("protocol", v.Protocol),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("host", v.Host),
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.String("uri_path", v.URIPath),
				zap.String("route_path", v.RoutePath),
				zap.String("request_id", v.RequestID),
				zap.String("referer", v.Referer),
				zap.String("user_agent", v.UserAgent),
				zap.Int("status", v.Status),
				zap.Error(v.Error),
				zap.String("content_length", v.ContentLength),
				zap.Int64("response_size", v.ResponseSize))

			return nil
		},
		LogLatency:       true,
		LogProtocol:      true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRequestID:     true,
		LogReferer:       true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
	}))
	hs.server.Use(middleware.Recover())

	hs.server.GET("/ping", hs.pingHandler)
	hs.server.POST("/otp/request", hs.userHandler.RequestOTP)
	hs.server.POST("/otp/validate", hs.userHandler.ValidateOTP)
}

func (hs *HTTPServer) makeHandler() {
	hs.pingHandler = rest.Ping

	deps := rest.Dependencies{
		User: hs.userSvc,
	}

	hs.userHandler = rest.NewUser(deps)
}

func (hs *HTTPServer) makeService() {
	deps := service.Dependencies{
		User:                hs.userRepo,
		RandNumberGenerator: stringutil.RandomNumbers,
	}

	hs.userSvc = service.NewUser(deps)
}

func (hs *HTTPServer) makeRepository() {
	deps := repository.Dependencies{
		DB:      hs.db,
		NowFunc: time.Now,
	}

	hs.userRepo = repository.NewUser(deps)
}

func NewHTTPServer() (*HTTPServer, error) {
	ctx := context.TODO()
	if err := config.Load(); err != nil {
		return nil, err
	}

	v := validator.New()
	e := echo.New()
	e.HideBanner = true
	e.Validator = &reqValidator{v: v}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local",
		viper.GetString(config.DBUsername),
		viper.GetString(config.DBPassword),
		viper.GetString(config.DBAddress),
		viper.GetString(config.DBName))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(10 * time.Second)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	hs := &HTTPServer{
		server: e,
		v:      v,
		db:     db,
	}

	hs.makeRepository()
	hs.makeService()
	hs.makeHandler()
	hs.route()

	return hs, nil
}
