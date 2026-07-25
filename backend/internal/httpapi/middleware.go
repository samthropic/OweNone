package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		userID := request.Header.Get("X-User-ID")
		if userID == "" {
			writeError(response, http.StatusUnauthorized, "unauthorized", "X-User-ID is required")
			return
		}
		ctx := context.WithValue(request.Context(), userIDContextKey, userID)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func userIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDContextKey).(string)
	return userID
}

func cors(frontendOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin != "" && origin == frontendOrigin {
			response.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
			response.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key, X-User-ID")
			response.Header().Set("Access-Control-Max-Age", "600")
			response.Header().Add("Vary", "Origin")
		}
		if request.Method == http.MethodOptions {
			if origin != frontendOrigin {
				writeError(response, http.StatusForbidden, "origin_not_allowed", "origin is not allowed")
				return
			}
			response.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(response, request)
	})
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(response, request)
		logger.InfoContext(request.Context(), "request completed",
			"method", request.Method,
			"path", request.URL.Path,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

func recoverPanic(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(request.Context(), "request panic", "error", recovered, "stack", string(debug.Stack()))
				writeError(response, http.StatusInternalServerError, "internal_error", "an internal error occurred")
			}
		}()
		next.ServeHTTP(response, request)
	})
}
