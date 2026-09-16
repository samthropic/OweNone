package httpapi

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/avatars"
	"github.com/samfiallos/owenone/backend/internal/domain"
	"github.com/samfiallos/owenone/backend/internal/store"
)

const testUserID = "10000000-0000-0000-0000-000000000001"

type stubService struct {
	dashboard         domain.Dashboard
	authenticateErr   error
	logInErr          error
	signUpErr         error
	removeFriendErr   error
	removedFriendUser domain.User
	updateProfileErr  error
	updateProfileUser domain.User
	changePasswordErr error
	updateGroupErr    error
	updateGroupResult domain.Group
	deleteGroupErr    error
	activityFeed      app.ActivityFeed
	activityErr       error
}

func (*stubService) Ping(context.Context) error { return nil }
func (s *stubService) Dashboard(_ context.Context, _ string) (domain.Dashboard, error) {
	return s.dashboard, nil
}
func (s *stubService) SignUp(_ context.Context, _ app.SignUpInput) (app.Session, error) {
	if s.signUpErr != nil {
		return app.Session{}, s.signUpErr
	}
	return app.Session{Token: "test-token", User: domain.User{ID: testUserID, DisplayName: "New"}}, nil
}
func (s *stubService) LogIn(_ context.Context, _, _ string) (app.Session, error) {
	if s.logInErr != nil {
		return app.Session{}, s.logInErr
	}
	return app.Session{Token: "test-token", User: domain.User{ID: testUserID, DisplayName: "Sarah"}}, nil
}
func (*stubService) LogOut(_ context.Context, _ string) error { return nil }
func (s *stubService) Authenticate(_ context.Context, _ string) (domain.User, error) {
	if s.authenticateErr != nil {
		return domain.User{}, s.authenticateErr
	}
	return domain.User{ID: testUserID, DisplayName: "Sarah"}, nil
}
func (*stubService) CreateGroup(context.Context, string, app.CreateGroupInput) (string, error) {
	return "group", nil
}
func (s *stubService) UpdateGroup(_ context.Context, _ string, _ app.UpdateGroupInput) (domain.Group, error) {
	if s.updateGroupErr != nil {
		return domain.Group{}, s.updateGroupErr
	}
	return s.updateGroupResult, nil
}
func (s *stubService) DeleteGroup(_ context.Context, _, _ string) error {
	return s.deleteGroupErr
}
func (*stubService) AddFriend(context.Context, string, string) (domain.User, error) {
	return domain.User{}, nil
}
func (s *stubService) RemoveFriend(_ context.Context, _, _ string) (domain.User, error) {
	if s.removeFriendErr != nil {
		return domain.User{}, s.removeFriendErr
	}
	return s.removedFriendUser, nil
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
func (s *stubService) UpdateProfile(_ context.Context, _ string, _ app.UpdateProfileInput) (domain.User, error) {
	if s.updateProfileErr != nil {
		return domain.User{}, s.updateProfileErr
	}
	return s.updateProfileUser, nil
}
func (s *stubService) SetAvatarURL(_ context.Context, _, url string) (domain.User, error) {
	user := s.updateProfileUser
	user.AvatarURL = url
	return user, nil
}
func (s *stubService) ClearAvatarURL(_ context.Context, _ string) (domain.User, error) {
	user := s.updateProfileUser
	user.AvatarURL = ""
	return user, nil
}
func (s *stubService) ChangePassword(_ context.Context, _, _, _, _ string) error {
	return s.changePasswordErr
}
func (s *stubService) Activity(_ context.Context, _ string, _ app.ActivityQuery) (app.ActivityFeed, error) {
	if s.activityErr != nil {
		return app.ActivityFeed{Activity: []domain.Activity{}}, s.activityErr
	}
	feed := s.activityFeed
	if feed.Activity == nil {
		feed.Activity = []domain.Activity{}
	}
	return feed, nil
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
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"displayName":"Sarah"`) {
		t.Errorf("body = %s, want dashboard data", response.Body.String())
	}
}

func TestSignUpCreatesSession(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup",
		strings.NewReader(`{"email":"new@example.com","displayName":"New User","password":"password123"}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if !strings.Contains(response.Body.String(), `"token"`) {
		t.Errorf("body = %s, want session token", response.Body.String())
	}
}

func TestSignUpRejectsUnknownFields(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup",
		strings.NewReader(`{"email":"a@b.com","displayName":"A","password":"pass1234","extra":true}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestLogInReturnsSession(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"sarah@example.com","password":"password123"}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"token"`) {
		t.Errorf("body = %s, want session token", response.Body.String())
	}
}

func TestLogInBadCredentialsReturns401(t *testing.T) {
	service := &stubService{logInErr: app.ErrUnauthorized}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"a@b.com","password":"wrongpass"}`))
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

func TestLogOutSucceeds(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"loggedOut":true`) {
		t.Errorf("body = %s, want loggedOut true", response.Body.String())
	}
}

func TestSessionReturnsUser(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"user"`) {
		t.Errorf("body = %s, want user field", response.Body.String())
	}
}

func TestInvalidBearerReturns401(t *testing.T) {
	service := &stubService{authenticateErr: app.ErrUnauthorized}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	request.Header.Set("Authorization", "Bearer bad-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
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

func TestRemoveFriendReturns200OnSuccess(t *testing.T) {
	service := &stubService{removedFriendUser: domain.User{ID: "20000000-0000-0000-0000-000000000002", DisplayName: "Friend"}}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/friends/remove",
		strings.NewReader(`{"email":"friend@example.com"}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"removed":true`) {
		t.Errorf("body = %s, want removed:true", response.Body.String())
	}
}

func TestRemoveFriendReturns400ForOutstandingBalance(t *testing.T) {
	service := &stubService{removeFriendErr: fmt.Errorf("%w: settle the outstanding balance with Friend before removing them", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/friends/remove",
		strings.NewReader(`{"email":"friend@example.com"}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestRemoveFriendReturns404ForUnknownEmail(t *testing.T) {
	service := &stubService{removeFriendErr: fmt.Errorf("%w: no friend with that email", app.ErrFriendNotFound)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/friends/remove",
		strings.NewReader(`{"email":"nobody@example.com"}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if !strings.Contains(response.Body.String(), `"code":"not_found"`) {
		t.Errorf("body = %s, want not_found code", response.Body.String())
	}
}

func TestRemoveFriendReturns401WithoutBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/friends/remove",
		strings.NewReader(`{"email":"friend@example.com"}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

func testHandler(service Service) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	storage, err := avatars.NewStorage(tTempAvatarDir())
	if err != nil {
		panic(err)
	}
	return New(service, logger, "http://localhost:3000", storage)
}

var testAvatarDir string

func tTempAvatarDir() string {
	if testAvatarDir != "" {
		return testAvatarDir
	}
	dir, err := os.MkdirTemp("", "owenone-avatars-")
	if err != nil {
		panic(err)
	}
	testAvatarDir = dir
	return dir
}

// --- Profile handler tests ---

func TestUpdateProfileReturns200OnSuccess(t *testing.T) {
	service := &stubService{updateProfileUser: domain.User{ID: testUserID, DisplayName: "Updated"}}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile",
		strings.NewReader(`{"displayName":"Updated","email":"alice@example.com","preferredCurrency":"GBP"}`))
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"user"`) {
		t.Errorf("body = %s, want user field", response.Body.String())
	}
}

func TestUpdateProfileReturns400OnValidationFailure(t *testing.T) {
	service := &stubService{updateProfileErr: fmt.Errorf("%w: email must be a valid address", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile",
		strings.NewReader(`{"displayName":"Alice","email":"bad","preferredCurrency":"GBP"}`))
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestUpdateProfileReturns409OnDuplicateEmail(t *testing.T) {
	service := &stubService{updateProfileErr: store.ErrConflict}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile",
		strings.NewReader(`{"displayName":"Alice","email":"existing@example.com","preferredCurrency":"GBP"}`))
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	if !strings.Contains(response.Body.String(), "that email is already in use") {
		t.Errorf("body = %s, want 'that email is already in use' message", response.Body.String())
	}
}

func TestUpdateProfileReturns401WithoutBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile",
		strings.NewReader(`{"displayName":"Alice","email":"alice@example.com","preferredCurrency":"GBP"}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

// --- ChangePassword handler tests ---

func TestChangePasswordReturns200OnSuccess(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile/password",
		strings.NewReader(`{"currentPassword":"old-pass-123","newPassword":"new-pass-456"}`))
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"updated":true`) {
		t.Errorf("body = %s, want updated:true", response.Body.String())
	}
}

func TestChangePasswordReturns401OnWrongCurrentPassword(t *testing.T) {
	service := &stubService{changePasswordErr: fmt.Errorf("%w: current password is incorrect", app.ErrUnauthorized)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile/password",
		strings.NewReader(`{"currentPassword":"wrong","newPassword":"new-pass-456"}`))
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), "current password is incorrect") {
		t.Errorf("body = %s, want 'current password is incorrect' message", response.Body.String())
	}
}

func TestChangePasswordReturns400OnWeakNewPassword(t *testing.T) {
	service := &stubService{changePasswordErr: fmt.Errorf("%w: password must be between 8 and 128 characters", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile/password",
		strings.NewReader(`{"currentPassword":"old-pass-123","newPassword":"short"}`))
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestChangePasswordReturns401WithoutBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/profile/password",
		strings.NewReader(`{"currentPassword":"old-pass-123","newPassword":"new-pass-456"}`))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

// --- UpdateGroup handler tests ---

const (
	hTestGroupID    = "30000000-0000-0000-0000-000000000003"
	updateGroupBody = `{"groupId":"30000000-0000-0000-0000-000000000003","name":"Trip","icon":"✈️","memberIds":[]}`
)

func TestUpdateGroupReturns200OnSuccess(t *testing.T) {
	service := &stubService{updateGroupResult: domain.Group{ID: hTestGroupID, Name: "Trip", Icon: "✈️"}}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/update", strings.NewReader(updateGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"group"`) {
		t.Errorf("body = %s, want group field", response.Body.String())
	}
}

func TestUpdateGroupReturns400OnValidationError(t *testing.T) {
	service := &stubService{updateGroupErr: fmt.Errorf("%w: group name must contain between 1 and 100 characters", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/update", strings.NewReader(updateGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestUpdateGroupReturns403ForNonOwner(t *testing.T) {
	service := &stubService{updateGroupErr: fmt.Errorf("%w: only the group owner can edit this group", app.ErrForbidden)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/update", strings.NewReader(updateGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestUpdateGroupReturns404ForUnknownGroup(t *testing.T) {
	service := &stubService{updateGroupErr: fmt.Errorf("not found: %w", store.ErrNotFound)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/update", strings.NewReader(updateGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestUpdateGroupReturns401WithoutBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/update", strings.NewReader(updateGroupBody))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

// --- DeleteGroup handler tests ---

const deleteGroupBody = `{"groupId":"30000000-0000-0000-0000-000000000003"}`

func TestDeleteGroupReturns200OnSuccess(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/delete", strings.NewReader(deleteGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"deleted":true`) {
		t.Errorf("body = %s, want deleted:true", response.Body.String())
	}
}

func TestDeleteGroupReturns400WhenGroupHasExpenses(t *testing.T) {
	service := &stubService{deleteGroupErr: fmt.Errorf("%w: this group has recorded expenses, so it cannot be deleted", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/delete", strings.NewReader(deleteGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestDeleteGroupReturns403ForNonOwner(t *testing.T) {
	service := &stubService{deleteGroupErr: fmt.Errorf("%w: only the group owner can delete this group", app.ErrForbidden)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/delete", strings.NewReader(deleteGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestDeleteGroupReturns404ForUnknownGroup(t *testing.T) {
	service := &stubService{deleteGroupErr: fmt.Errorf("not found: %w", store.ErrNotFound)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/delete", strings.NewReader(deleteGroupBody))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestDeleteGroupReturns401WithoutBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/groups/delete", strings.NewReader(deleteGroupBody))
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

// --- Activity feed handler tests ---

func TestActivityFeedRequiresUser(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

func TestActivityFeedReturns200WithEnvelope(t *testing.T) {
	service := &stubService{activityFeed: app.ActivityFeed{Activity: []domain.Activity{}, Total: 0, HasMore: false}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, field := range []string{`"activity"`, `"total"`, `"hasMore"`} {
		if !strings.Contains(body, field) {
			t.Errorf("body missing %s: %s", field, body)
		}
	}
}

func TestActivityFeedEncodesEmptyActivityAsArray(t *testing.T) {
	service := &stubService{activityFeed: app.ActivityFeed{Activity: []domain.Activity{}, Total: 0, HasMore: false}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if strings.Contains(body, `"activity":null`) {
		t.Errorf("activity serialized as null, want []: %s", body)
	}
	if !strings.Contains(body, `"activity":[]`) {
		t.Errorf("body = %s, want \"activity\":[]", body)
	}
}

func TestActivityFeedInvalidLimitReturns400(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?limit=abc", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestActivityFeedInvalidOffsetReturns400(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?offset=xyz", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestActivityFeedKindSettlementReturns200WithEnvelope(t *testing.T) {
	service := &stubService{activityFeed: app.ActivityFeed{Activity: []domain.Activity{}, Total: 0, HasMore: false}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?kind=settlement", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, field := range []string{`"activity"`, `"total"`, `"hasMore"`} {
		if !strings.Contains(body, field) {
			t.Errorf("body missing %s: %s", field, body)
		}
	}
	if !strings.Contains(body, `"activity":[]`) {
		t.Errorf("empty-filter body = %s, want \"activity\":[]", body)
	}
}

func TestActivityFeedUnknownKindReturns400(t *testing.T) {
	service := &stubService{activityErr: fmt.Errorf("%w: kind must be expense or settlement", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?kind=banana", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestActivityFeedUnfilteredStillBehavesAsBefore(t *testing.T) {
	feed := app.ActivityFeed{
		Activity: []domain.Activity{{ID: "a1", Kind: "expense", Description: "Lunch", Actor: domain.User{ID: testUserID, DisplayName: "Sarah"}}},
		Total:    1,
		HasMore:  false,
	}
	service := &stubService{activityFeed: feed}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"total":1`) {
		t.Errorf("body = %s, want total:1", body)
	}
	if !strings.Contains(body, `"hasMore":false`) {
		t.Errorf("body = %s, want hasMore:false", body)
	}
}

func TestActivityFeedKindRequiresBearer(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?kind=settlement", nil)
	response := httptest.NewRecorder()
	testHandler(&stubService{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"unauthorized"`) {
		t.Errorf("body = %s, want unauthorized code", response.Body.String())
	}
}

func TestActivityFeedGroupIDFilterReturns200(t *testing.T) {
	service := &stubService{activityFeed: app.ActivityFeed{Activity: []domain.Activity{}, Total: 0, HasMore: false}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?groupId=10000000-0000-0000-0000-000000000001", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, field := range []string{`"activity"`, `"total"`, `"hasMore"`} {
		if !strings.Contains(body, field) {
			t.Errorf("body missing %s: %s", field, body)
		}
	}
}

func TestActivityFeedGroupIDMalformedReturns400(t *testing.T) {
	service := &stubService{activityErr: fmt.Errorf("%w: group ID must be a valid UUID", app.ErrInvalidInput)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?groupId=not-a-uuid", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Errorf("body = %s, want invalid_request code", response.Body.String())
	}
}

func TestActivityFeedGroupIDAndKindReturns200(t *testing.T) {
	service := &stubService{activityFeed: app.ActivityFeed{Activity: []domain.Activity{}, Total: 0, HasMore: false}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/activity?groupId=10000000-0000-0000-0000-000000000001&kind=expense", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	testHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, field := range []string{`"activity"`, `"total"`, `"hasMore"`} {
		if !strings.Contains(body, field) {
			t.Errorf("body missing %s: %s", field, body)
		}
	}
}
