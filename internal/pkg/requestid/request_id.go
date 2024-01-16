package requestid

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

type reqIDKey string

var key = reqIDKey("request_id")

func InjectToCtx(resHeaders http.Header, ctx context.Context) context.Context {
	id := resHeaders.Get(echo.HeaderXRequestID)

	return context.WithValue(ctx, key, id)
}

func ExtractFromCtx(ctx context.Context) string {
	return ctx.Value(key).(string)
}
