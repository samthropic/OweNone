package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/samfiallos/owenone/backend/internal/app"
	"github.com/samfiallos/owenone/backend/internal/avatars"
	"github.com/samfiallos/owenone/backend/internal/domain"
	"github.com/samfiallos/owenone/backend/internal/store"
)

const maxRequestBodyBytes = 1 << 20

func (api *API) live(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *API) ready(response http.ResponseWriter, request *http.Request) {
	if err := api.service.Ping(request.Context()); err != nil {
		api.logger.ErrorContext(request.Context(), "readiness check failed", "error", err)
		writeError(response, http.StatusServiceUnavailable, "not_ready", "database is unavailable")
		return
	}
	writeJSON(response, http.StatusOK, map[string]string{"status": "ready"})
}

func (api *API) signUp(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		Password    string `json:"password"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	session, err := api.service.SignUp(request.Context(), app.SignUpInput{
		Email: body.Email, DisplayName: body.DisplayName, Password: body.Password,
	})
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(response, http.StatusConflict, "conflict", "an account with that email already exists")
			return
		}
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusCreated, session)
}

func (api *API) logIn(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	session, err := api.service.LogIn(request.Context(), body.Email, body.Password)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, session)
}

func (api *API) logOut(response http.ResponseWriter, request *http.Request) {
	if err := api.service.LogOut(request.Context(), tokenFromContext(request.Context())); err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"loggedOut": true})
}

func (api *API) session(response http.ResponseWriter, request *http.Request) {
	writeData(response, http.StatusOK, map[string]any{"user": userFromContext(request.Context())})
}

func (api *API) dashboard(response http.ResponseWriter, request *http.Request) {
	dashboard, err := api.service.Dashboard(request.Context(), userIDFromContext(request.Context()))
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, dashboard)
}

func (api *API) activityFeed(response http.ResponseWriter, request *http.Request) {
	limit, offset, ok := parsePaginationQuery(response, request)
	if !ok {
		return
	}
	kind := request.URL.Query().Get("kind")
	groupID := request.URL.Query().Get("groupId")
	feed, err := api.service.Activity(request.Context(), userIDFromContext(request.Context()), app.ActivityQuery{
		Kind:    kind,
		GroupID: groupID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, feed)
}

// parsePaginationQuery parses ?limit and ?offset from the request URL.
// It returns (limit, offset, true) on success. On parse failure it writes a 400
// and returns (0, 0, false). Defaults: limit=50, offset=0.
func parsePaginationQuery(response http.ResponseWriter, request *http.Request) (limit, offset int, ok bool) {
	query := request.URL.Query()

	limit = 50
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(response, http.StatusBadRequest, "invalid_request", "limit must be a positive integer")
			return 0, 0, false
		}
		limit = parsed
	}

	offset = 0
	if raw := query.Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(response, http.StatusBadRequest, "invalid_request", "offset must be a non-negative integer")
			return 0, 0, false
		}
		offset = parsed
	}

	return limit, offset, true
}

func (api *API) createGroup(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Name      string   `json:"name"`
		Icon      string   `json:"icon"`
		MemberIDs []string `json:"memberIds"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	id, err := api.service.CreateGroup(request.Context(), userIDFromContext(request.Context()), app.CreateGroupInput{Name: body.Name, Icon: body.Icon, MemberIDs: body.MemberIDs})
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusCreated, map[string]string{"id": id})
}

func (api *API) updateGroup(response http.ResponseWriter, request *http.Request) {
	var body struct {
		GroupID   string   `json:"groupId"`
		Name      string   `json:"name"`
		Icon      string   `json:"icon"`
		MemberIDs []string `json:"memberIds"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	group, err := api.service.UpdateGroup(request.Context(), userIDFromContext(request.Context()), app.UpdateGroupInput{
		GroupID: body.GroupID, Name: body.Name, Icon: body.Icon, MemberIDs: body.MemberIDs,
	})
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"group": group})
}

func (api *API) deleteGroup(response http.ResponseWriter, request *http.Request) {
	var body struct {
		GroupID string `json:"groupId"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := api.service.DeleteGroup(request.Context(), userIDFromContext(request.Context()), body.GroupID); err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"deleted": true})
}

func (api *API) addFriend(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	friend, err := api.service.AddFriend(request.Context(), userIDFromContext(request.Context()), body.Email)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusCreated, friend)
}

func (api *API) removeFriend(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	user, err := api.service.RemoveFriend(request.Context(), userIDFromContext(request.Context()), body.Email)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"removed": true, "user": user})
}

func (api *API) createExpense(response http.ResponseWriter, request *http.Request) {
	var body struct {
		GroupID      string `json:"groupId"`
		Description  string `json:"description"`
		Category     string `json:"category"`
		AmountMinor  int64  `json:"amountMinor"`
		Currency     string `json:"currency"`
		PaidByUserID string `json:"paidByUserId"`
		Split        struct {
			Method         string                `json:"method"`
			ParticipantIDs []string              `json:"participantIds"`
			Shares         []domain.ExpenseSplit `json:"shares"`
		} `json:"split"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	id, err := api.service.CreateExpense(request.Context(), userIDFromContext(request.Context()), app.CreateExpenseInput{
		GroupID: body.GroupID, Description: body.Description, Category: body.Category,
		AmountMinor: body.AmountMinor, Currency: body.Currency, PaidByUserID: body.PaidByUserID,
		SplitMethod: body.Split.Method, ParticipantIDs: body.Split.ParticipantIDs,
		ExactSplits: body.Split.Shares, IdempotencyKey: request.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusCreated, map[string]string{"id": id})
}

func (api *API) createSettlement(response http.ResponseWriter, request *http.Request) {
	var body struct {
		GroupID       *string `json:"groupId"`
		ToUserID      string  `json:"toUserId"`
		AmountMinor   int64   `json:"amountMinor"`
		Currency      string  `json:"currency"`
		PaymentMethod string  `json:"paymentMethod"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	id, err := api.service.CreateSettlement(request.Context(), userIDFromContext(request.Context()), app.CreateSettlementInput{
		GroupID: body.GroupID, ToUserID: body.ToUserID, AmountMinor: body.AmountMinor,
		Currency: body.Currency, PaymentMethod: body.PaymentMethod,
		IdempotencyKey: request.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusCreated, map[string]string{"id": id})
}

func (api *API) createReminder(response http.ResponseWriter, request *http.Request) {
	var body struct {
		RecipientID string `json:"recipientId"`
		Message     string `json:"message"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	id, err := api.service.CreateReminder(request.Context(), userIDFromContext(request.Context()), body.RecipientID, body.Message)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusCreated, map[string]string{"id": id})
}

func (api *API) updateProfile(response http.ResponseWriter, request *http.Request) {
	var body struct {
		DisplayName       string `json:"displayName"`
		Email             string `json:"email"`
		PreferredCurrency string `json:"preferredCurrency"`
		PaymentApps       *struct {
			Venmo   string `json:"venmo"`
			PayPal  string `json:"paypal"`
			CashApp string `json:"cashApp"`
			Zelle   string `json:"zelle"`
		} `json:"paymentApps"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	input := app.UpdateProfileInput{
		DisplayName: body.DisplayName, Email: body.Email, PreferredCurrency: body.PreferredCurrency,
	}
	if body.PaymentApps != nil {
		input.UpdatePaymentApps = true
		input.PaymentApps = domain.PaymentApps{
			Venmo: body.PaymentApps.Venmo, PayPal: body.PaymentApps.PayPal,
			CashApp: body.PaymentApps.CashApp, Zelle: body.PaymentApps.Zelle,
		}
	}
	user, err := api.service.UpdateProfile(request.Context(), userIDFromContext(request.Context()), input)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(response, http.StatusConflict, "conflict", "that email is already in use")
			return
		}
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"user": user})
}

func (api *API) changePassword(response http.ResponseWriter, request *http.Request) {
	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	err := api.service.ChangePassword(
		request.Context(),
		userIDFromContext(request.Context()),
		body.CurrentPassword,
		body.NewPassword,
		tokenFromContext(request.Context()),
	)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"updated": true})
}

func (api *API) uploadAvatar(response http.ResponseWriter, request *http.Request) {
	if api.avatars == nil {
		writeError(response, http.StatusServiceUnavailable, "not_ready", "avatar storage is unavailable")
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, avatars.MaxBytes+1<<20)
	if err := request.ParseMultipartForm(avatars.MaxBytes + 1<<20); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "image must be 2 MB or smaller")
		return
	}
	file, _, err := request.FormFile("avatar")
	if err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "avatar file is required")
		return
	}
	defer file.Close()

	data, err := avatars.ReadLimited(file)
	if err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", "could not read image")
		return
	}
	if len(data) > avatars.MaxBytes {
		writeError(response, http.StatusBadRequest, "invalid_request", "image must be 2 MB or smaller")
		return
	}

	userID := userIDFromContext(request.Context())
	publicURL, err := api.avatars.Save(userID, data)
	if err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	user, err := api.service.SetAvatarURL(request.Context(), userID, publicURL)
	if err != nil {
		_ = api.avatars.Remove(userID)
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"user": user})
}

func (api *API) deleteAvatar(response http.ResponseWriter, request *http.Request) {
	userID := userIDFromContext(request.Context())
	if api.avatars != nil {
		if err := api.avatars.Remove(userID); err != nil {
			api.logger.ErrorContext(request.Context(), "remove avatar file", "error", err)
			writeError(response, http.StatusInternalServerError, "internal_error", "an internal error occurred")
			return
		}
	}
	user, err := api.service.ClearAvatarURL(request.Context(), userID)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"user": user})
}

func (api *API) serveAvatar(response http.ResponseWriter, request *http.Request) {
	if api.avatars == nil {
		http.NotFound(response, request)
		return
	}
	userID := request.PathValue("userID")
	if _, err := uuid.Parse(userID); err != nil {
		http.NotFound(response, request)
		return
	}
	file, contentType, err := api.avatars.Open(userID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(response, request)
			return
		}
		api.logger.ErrorContext(request.Context(), "open avatar", "error", err)
		http.Error(response, "internal error", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		api.logger.ErrorContext(request.Context(), "stat avatar", "error", err)
		http.Error(response, "internal error", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", contentType)
	response.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(response, request, info.Name(), info.ModTime(), file)
}

func (api *API) writeServiceError(response http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, app.ErrUnauthorized):
		writeError(response, http.StatusUnauthorized, "unauthorized", clientMessage(err, app.ErrUnauthorized, "authentication required"))
	case errors.Is(err, app.ErrInvalidInput):
		writeError(response, http.StatusBadRequest, "invalid_request", clientMessage(err, app.ErrInvalidInput, "the request is invalid"))
	case errors.Is(err, app.ErrForbidden):
		writeError(response, http.StatusForbidden, "forbidden", "you do not have access to this resource")
	case errors.Is(err, app.ErrFriendNotFound):
		writeError(response, http.StatusNotFound, "not_found", "no friend with that email")
	case errors.Is(err, store.ErrNotFound):
		writeError(response, http.StatusNotFound, "not_found", "the requested resource was not found")
	case errors.Is(err, store.ErrConflict):
		writeError(response, http.StatusConflict, "conflict", "the request conflicts with existing data")
	default:
		api.logger.ErrorContext(request.Context(), "request failed", "error", err)
		writeError(response, http.StatusInternalServerError, "internal_error", "an internal error occurred")
	}
}

// clientMessage strips the sentinel prefix from a wrapped error so API consumers
// receive only the human-readable part.
func clientMessage(err, sentinel error, fallback string) string {
	message := strings.TrimPrefix(err.Error(), sentinel.Error()+": ")
	if message == "" || message == sentinel.Error() {
		return fallback
	}
	return message
}

func decodeJSON(response http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON object")
	}
	return nil
}

func writeData(response http.ResponseWriter, status int, data any) {
	writeJSON(response, status, map[string]any{"data": data})
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(body); err != nil {
		http.Error(response, "failed to encode response", http.StatusInternalServerError)
	}
}
