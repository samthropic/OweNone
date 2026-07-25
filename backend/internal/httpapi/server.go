package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

type Service interface {
	Ping(context.Context) error
	Dashboard(context.Context, string) (domain.Dashboard, error)
	JoinWaitlist(context.Context, string) (bool, error)
	CreateGroup(context.Context, string, app.CreateGroupInput) (string, error)
	AddFriend(context.Context, string, string) (domain.User, error)
	CreateExpense(context.Context, string, app.CreateExpenseInput) (string, error)
	CreateSettlement(context.Context, string, app.CreateSettlementInput) (string, error)
	CreateReminder(context.Context, string, string, string) (string, error)
}

type API struct {
	service Service
	logger  *slog.Logger
}

func New(service Service, logger *slog.Logger, frontendOrigin string) http.Handler {
	api := &API{service: service, logger: logger}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", api.live)
	mux.HandleFunc("GET /health/ready", api.ready)
	mux.HandleFunc("POST /api/v1/waitlist", api.joinWaitlist)
	mux.Handle("GET /api/v1/dashboard", requireUser(http.HandlerFunc(api.dashboard)))
	mux.Handle("POST /api/v1/groups", requireUser(http.HandlerFunc(api.createGroup)))
	mux.Handle("POST /api/v1/friends", requireUser(http.HandlerFunc(api.addFriend)))
	mux.Handle("POST /api/v1/expenses", requireUser(http.HandlerFunc(api.createExpense)))
	mux.Handle("POST /api/v1/settlements", requireUser(http.HandlerFunc(api.createSettlement)))
	mux.Handle("POST /api/v1/reminders", requireUser(http.HandlerFunc(api.createReminder)))

	return recoverPanic(logger, requestLog(logger, cors(frontendOrigin, securityHeaders(mux))))
}
