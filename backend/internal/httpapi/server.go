package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/avatars"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

type Service interface {
	Ping(context.Context) error
	Dashboard(context.Context, string) (domain.Dashboard, error)
	Activity(context.Context, string, app.ActivityQuery) (app.ActivityFeed, error)
	CreateGroup(context.Context, string, app.CreateGroupInput) (string, error)
	UpdateGroup(context.Context, string, app.UpdateGroupInput) (domain.Group, error)
	DeleteGroup(context.Context, string, string) error
	AddFriend(context.Context, string, string) (domain.User, error)
	RemoveFriend(context.Context, string, string) (domain.User, error)
	CreateExpense(context.Context, string, app.CreateExpenseInput) (string, error)
	CreateSettlement(context.Context, string, app.CreateSettlementInput) (string, error)
	CreateReminder(context.Context, string, string, string) (string, error)
	SignUp(context.Context, app.SignUpInput) (app.Session, error)
	LogIn(context.Context, string, string) (app.Session, error)
	LogOut(context.Context, string) error
	Authenticate(context.Context, string) (domain.User, error)
	UpdateProfile(context.Context, string, app.UpdateProfileInput) (domain.User, error)
	SetAvatarURL(context.Context, string, string) (domain.User, error)
	ClearAvatarURL(context.Context, string) (domain.User, error)
	ChangePassword(context.Context, string, string, string, string) error
}

type API struct {
	service Service
	logger  *slog.Logger
	avatars *avatars.Storage
}

func New(service Service, logger *slog.Logger, frontendOrigin string, avatarStorage *avatars.Storage) http.Handler {
	api := &API{service: service, logger: logger, avatars: avatarStorage}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", api.live)
	mux.HandleFunc("GET /health/ready", api.ready)
	mux.HandleFunc("POST /api/v1/auth/signup", api.signUp)
	mux.HandleFunc("POST /api/v1/auth/login", api.logIn)
	mux.Handle("POST /api/v1/auth/logout", api.requireUser(http.HandlerFunc(api.logOut)))
	mux.Handle("GET /api/v1/auth/session", api.requireUser(http.HandlerFunc(api.session)))
	mux.Handle("GET /api/v1/dashboard", api.requireUser(http.HandlerFunc(api.dashboard)))
	mux.Handle("GET /api/v1/activity", api.requireUser(http.HandlerFunc(api.activityFeed)))
	mux.Handle("POST /api/v1/groups", api.requireUser(http.HandlerFunc(api.createGroup)))
	mux.Handle("POST /api/v1/groups/update", api.requireUser(http.HandlerFunc(api.updateGroup)))
	mux.Handle("POST /api/v1/groups/delete", api.requireUser(http.HandlerFunc(api.deleteGroup)))
	mux.Handle("POST /api/v1/friends", api.requireUser(http.HandlerFunc(api.addFriend)))
	mux.Handle("POST /api/v1/friends/remove", api.requireUser(http.HandlerFunc(api.removeFriend)))
	mux.Handle("POST /api/v1/expenses", api.requireUser(http.HandlerFunc(api.createExpense)))
	mux.Handle("POST /api/v1/settlements", api.requireUser(http.HandlerFunc(api.createSettlement)))
	mux.Handle("POST /api/v1/reminders", api.requireUser(http.HandlerFunc(api.createReminder)))
	mux.Handle("POST /api/v1/profile", api.requireUser(http.HandlerFunc(api.updateProfile)))
	mux.Handle("POST /api/v1/profile/password", api.requireUser(http.HandlerFunc(api.changePassword)))
	mux.Handle("POST /api/v1/profile/avatar", api.requireUser(http.HandlerFunc(api.uploadAvatar)))
	mux.Handle("DELETE /api/v1/profile/avatar", api.requireUser(http.HandlerFunc(api.deleteAvatar)))
	mux.HandleFunc("GET /api/v1/avatars/{userID}", api.serveAvatar)

	return recoverPanic(logger, requestLog(logger, cors(frontendOrigin, securityHeaders(mux))))
}
