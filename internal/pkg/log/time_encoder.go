package log

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// RFC3339NanoEncoder serializes a time.Time to an RFC3339Nano-formatted string
// with nanosecond precision.
func RFC3339NanoEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339Nano))
}
