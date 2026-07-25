package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

const testUserID = "10000000-0000-0000-0000-000000000001"

type stubService struct {
	dashboard     domain.Dashboard
	waitlistEmail string
}

func (service *stubService) Ping(context.Context) error { return nil }
func (service *stubService) Dashboard(_ context.Context, _ string) (domain.Dashboard, error) {
	return service.dashboard, nil
}
func (service *stubService) JoinWaitlist(_ context.Context, email string) (bool, error) {
	service.waitlistEmail = email
	return true, nil
}
func (*stubService) CreateGroup(context.Context, string, app.CreateGroupInput) (string, error) {
	return "group", nil
}
func (*stubService) AddFriend(context.Context, string, string) (domain.User, error) {
	return domain.User{}, nil
}
func (*stubService) CreateExpense(context.Context, string, app.CreateExpenseInput) (string, error) {
	return "expense", nil
}
func (*stubService) CreateSettlement(context.Context, string, app.CreateSettlementInput) (string, error) {
	return "settlement", nil
}
func (*stubService) CreateReminder(context.Context, string, string, string) (string, error) {
	return "reminder", nil
}

func TestDashboardRequiresUser(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

func TestDashboardReturnsDataEnvelope(t *testing.T) {
	service := &stubService{dashboard: domain.Dashboard{User: domain.User{ID: testUserID, DisplayName: "Sarah"}}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	request.Header.Set("X-User-ID", testUserID)
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"displayName":"Sarah"`) {
		t.Errorf("body = %s, want dashboard data", response.Body.String())
	}
}

func TestWaitlistRejectsUnknownJSONFields(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/waitlist", strings.NewReader(`{"email":"sarah@example.com","admin":true}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), "unknown field") {
		t.Errorf("body = %s, want unknown field error", response.Body.String())
	}
}

func TestCORSAllowsOnlyConfiguredOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/dashboard", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("allow origin = %q, want configured origin", got)
	}
}

func testHandler(service Service) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(service, logger, "http://localhost:3000")
}
