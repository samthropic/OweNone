package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/samfiallos/owenone/backend/internal/app"
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

func (api *API) dashboard(response http.ResponseWriter, request *http.Request) {
	dashboard, err := api.service.Dashboard(request.Context(), userIDFromContext(request.Context()))
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	writeData(response, http.StatusOK, dashboard)
}

func (api *API) joinWaitlist(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	created, err := api.service.JoinWaitlist(request.Context(), body.Email)
	if err != nil {
		api.writeServiceError(response, request, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeData(response, status, map[string]bool{"joined": true})
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
		GroupID     *string `json:"groupId"`
		ToUserID    string  `json:"toUserId"`
		AmountMinor int64   `json:"amountMinor"`
		Currency    string  `json:"currency"`
	}
	if err := decodeJSON(response, request, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	id, err := api.service.CreateSettlement(request.Context(), userIDFromContext(request.Context()), app.CreateSettlementInput{
		GroupID: body.GroupID, ToUserID: body.ToUserID, AmountMinor: body.AmountMinor,
		Currency: body.Currency, IdempotencyKey: request.Header.Get("Idempotency-Key"),
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

func (api *API) writeServiceError(response http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, app.ErrInvalidInput):
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, app.ErrForbidden):
		writeError(response, http.StatusForbidden, "forbidden", "you do not have access to this resource")
	case errors.Is(err, store.ErrNotFound):
		writeError(response, http.StatusNotFound, "not_found", "the requested resource was not found")
	case errors.Is(err, store.ErrConflict):
		writeError(response, http.StatusConflict, "conflict", "the request conflicts with existing data")
	default:
		api.logger.ErrorContext(request.Context(), "request failed", "error", err)
		writeError(response, http.StatusInternalServerError, "internal_error", "an internal error occurred")
	}
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
