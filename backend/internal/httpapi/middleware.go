package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/samfiallos/owenone/backend/internal/domain"
)

type contextKey string

const (
	userIDContextKey contextKey = "userID"
	userContextKey   contextKey = "user"
	tokenContextKey  contextKey = "token"
)

func (api *API) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization")
		const bearerPrefix = "bearer "
		if len(authHeader) <= len(bearerPrefix) || !strings.EqualFold(authHeader[:len(bearerPrefix)], bearerPrefix) {
			writeError(response, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		token := authHeader[len(bearerPrefix):]
		if token == "" {
			writeError(response, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}

		user, err := api.service.Authenticate(request.Context(), token)
		if err != nil {
			writeError(response, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}

		ctx := context.WithValue(request.Context(), userIDContextKey, user.ID)
		ctx = context.WithValue(ctx, userContextKey, user)
		ctx = context.WithValue(ctx, tokenContextKey, token)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func userIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDContextKey).(string)
	return userID
}

func userFromContext(ctx context.Context) domain.User {
	user, _ := ctx.Value(userContextKey).(domain.User)
	return user
}

func tokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(tokenContextKey).(string)
	return token
}

func cors(frontendOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin != "" && origin == frontendOrigin {
			response.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
			response.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key, Authorization")
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
