package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func PanicRecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errMsg := panicMessage(err)
				logx.Errorw("panic recovered",
					logx.Field("time", time.Now().Format(time.RFC3339)),
					logx.Field("error", errMsg),
					logx.Field("stack", string(debug.Stack())),
					logx.Field("url", r.URL.String()),
					logx.Field("method", r.Method),
					logx.Field("headers", formatHeaders(r.Header)),
				)
				writeJSON(w, CodeServerError, "系统异常，请稍后重试", nil)
			}
		}()

		next(w, r)
	}
}

func panicMessage(err interface{}) string {
	switch value := err.(type) {
	case string:
		return value
	case error:
		return value.Error()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func formatHeaders(headers http.Header) string {
	var buffer bytes.Buffer
	for key, values := range headers {
		for _, value := range values {
			buffer.WriteString(fmt.Sprintf("%s: %s; ", key, value))
		}
	}
	return buffer.String()
}
