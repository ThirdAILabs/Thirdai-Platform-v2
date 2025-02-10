package auth

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func clientIp(r *http.Request) string {
	// https://stackoverflow.com/questions/27234861/correct-way-of-getting-clients-ip-addresses-from-http-request
	if ip := r.Header.Get("X-Real-Ip"); len(ip) > 0 {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); len(ip) > 0 {
		return ip
	}
	if len(r.RemoteAddr) > 0 {
		return r.RemoteAddr
	}
	return "Unknown"
}

func protocol(r *http.Request) string {
	protocol := r.Header.Get("X-Forwarded-Proto")
	if len(protocol) > 0 {
		return protocol
	}
	return r.URL.Scheme
}

func pathParams(r *http.Request) []interface{} {
	params := make([]interface{}, 0)

	ctx := r.Context()
	if ctx == nil {
		return params
	}

	rctx := chi.RouteContext(ctx)
	for i := range rctx.URLParams.Keys {
		if rctx.URLParams.Keys[i] != "*" {
			params = append(params, slog.String(rctx.URLParams.Keys[i], rctx.URLParams.Values[i]))
		}
	}

	return params
}

func queryParams(r *http.Request) []interface{} {
	params := make([]interface{}, 0)
	for k, v := range r.URL.Query() {
		params = append(params, slog.String(k, strings.Join(v, ";")))
	}
	return params
}

type AuditLogger struct {
	logger *slog.Logger
}

func NewAuditLogger(stream io.Writer) AuditLogger {
	logger := slog.New(slog.NewJSONHandler(stream, nil))
	return AuditLogger{logger: logger}
}

func (log *AuditLogger) Middleware(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		user, err := UserFromContext(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.logger.Info("",
			"username", user.Username,
			"user_id", user.Id,
			"client_ip", clientIp(r),
			"protocol", protocol(r),
			"method", r.Method,
			"url", r.URL.Path,
			slog.Group("path_params", pathParams(r)...),
			slog.Group("query_params", queryParams(r)...),
		)

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}
