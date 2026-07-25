package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("forbidden")
)

type Repository interface {
	Ping(context.Context) error
	LoadLedgerSnapshot(context.Context, string) (domain.LedgerSnapshot, error)
	JoinWaitlist(context.Context, string) (bool, error)
	CreateGroup(context.Context, CreateGroupCommand) error
	AddFriend(context.Context, string, string) (domain.User, error)
	CreateExpense(context.Context, CreateExpenseCommand) (string, error)
	CreateSettlement(context.Context, CreateSettlementCommand) (string, error)
	CreateReminder(context.Context, CreateReminderCommand) (string, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (service *Service) Ping(ctx context.Context) error {
	return service.repository.Ping(ctx)
}

func (service *Service) Dashboard(ctx context.Context, userID string) (domain.Dashboard, error) {
	if err := validateUUID("user ID", userID); err != nil {
		return domain.Dashboard{}, err
	}
	snapshot, err := service.repository.LoadLedgerSnapshot(ctx, userID)
	if err != nil {
		return domain.Dashboard{}, err
	}
	return domain.BuildDashboard(snapshot, service.now())
}

func (service *Service) JoinWaitlist(ctx context.Context, email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 320 {
		return false, invalid("email must be a valid address")
	}
	return service.repository.JoinWaitlist(ctx, email)
}

type CreateGroupInput struct {
	Name      string
	Icon      string
	MemberIDs []string
}

type CreateGroupCommand struct {
	ID        string
	CreatorID string
	Name      string
	Icon      string
	MemberIDs []string
}

func (service *Service) CreateGroup(ctx context.Context, actorID string, input CreateGroupInput) (string, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return "", err
	}
	input.Name = strings.TrimSpace(input.Name)
	if len(input.Name) < 1 || len(input.Name) > 100 {
		return "", invalid("group name must contain between 1 and 100 characters")
	}
	input.Icon = strings.TrimSpace(input.Icon)
	if input.Icon == "" {
		input.Icon = "👥"
	}
	if len([]byte(input.Icon)) > 32 {
		return "", invalid("group icon is too long")
	}
	input.MemberIDs = uniqueIDs(input.MemberIDs, actorID)
	for _, memberID := range input.MemberIDs {
		if err := validateUUID("member ID", memberID); err != nil {
			return "", err
		}
	}

	command := CreateGroupCommand{ID: uuid.NewString(), CreatorID: actorID, Name: input.Name, Icon: input.Icon, MemberIDs: input.MemberIDs}
	if err := service.repository.CreateGroup(ctx, command); err != nil {
		return "", err
	}
	return command.ID, nil
}

func (service *Service) AddFriend(ctx context.Context, actorID, email string) (domain.User, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return domain.User{}, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return domain.User{}, invalid("email must be a valid address")
	}
	return service.repository.AddFriend(ctx, actorID, email)
}

type CreateExpenseInput struct {
	GroupID        string
	Description    string
	Category       string
	AmountMinor    int64
	Currency       string
	PaidByUserID   string
	SplitMethod    string
	ParticipantIDs []string
	ExactSplits    []domain.ExpenseSplit
	IdempotencyKey string
}

type CreateExpenseCommand struct {
	ID                 string
	ActorID            string
	GroupID            string
	Description        string
	Category           string
	Money              domain.Money
	PaidByUserID       string
	SplitMethod        string
	Splits             []domain.ExpenseSplit
	IdempotencyKey     string
	RequestFingerprint string
}

func (service *Service) CreateExpense(ctx context.Context, actorID string, input CreateExpenseInput) (string, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return "", err
	}
	if err := validateUUID("group ID", input.GroupID); err != nil {
		return "", err
	}
	if input.PaidByUserID == "" {
		input.PaidByUserID = actorID
	}
	if err := validateUUID("payer ID", input.PaidByUserID); err != nil {
		return "", err
	}
	input.Description = strings.TrimSpace(input.Description)
	if len(input.Description) < 1 || len(input.Description) > 200 {
		return "", invalid("expense description must contain between 1 and 200 characters")
	}
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	if input.Category == "" {
		input.Category = "general"
	}
	if len(input.Category) > 40 {
		return "", invalid("expense category is too long")
	}
	money, err := domain.NewMoney(input.AmountMinor, strings.ToUpper(strings.TrimSpace(input.Currency)))
	if err != nil || money.AmountMinor <= 0 {
		return "", invalid("expense amount must be positive and use a valid currency")
	}
	if err := validateIdempotencyKey(input.IdempotencyKey); err != nil {
		return "", err
	}

	var splits []domain.ExpenseSplit
	switch input.SplitMethod {
	case "equal":
		splits, err = equalSplits(input.ParticipantIDs, input.AmountMinor)
	case "exact":
		splits, err = exactSplits(input.ExactSplits, input.AmountMinor)
	default:
		err = invalid("split method must be equal or exact")
	}
	if err != nil {
		return "", err
	}

	return service.repository.CreateExpense(ctx, CreateExpenseCommand{
		ID: uuid.NewString(), ActorID: actorID, GroupID: input.GroupID,
		Description: input.Description, Category: input.Category, Money: money,
		PaidByUserID: input.PaidByUserID, SplitMethod: input.SplitMethod,
		Splits: splits, IdempotencyKey: input.IdempotencyKey,
		RequestFingerprint: fingerprint(struct {
			ActorID      string
			GroupID      string
			Description  string
			Category     string
			Money        domain.Money
			PaidByUserID string
			SplitMethod  string
			Splits       []domain.ExpenseSplit
		}{actorID, input.GroupID, input.Description, input.Category, money, input.PaidByUserID, input.SplitMethod, splits}),
	})
}

type CreateSettlementInput struct {
	GroupID        *string
	ToUserID       string
	AmountMinor    int64
	Currency       string
	IdempotencyKey string
}

type CreateSettlementCommand struct {
	ID                 string
	GroupID            *string
	FromUserID         string
	ToUserID           string
	Money              domain.Money
	IdempotencyKey     string
	RequestFingerprint string
}

func (service *Service) CreateSettlement(ctx context.Context, actorID string, input CreateSettlementInput) (string, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return "", err
	}
	if err := validateUUID("recipient ID", input.ToUserID); err != nil {
		return "", err
	}
	if input.GroupID != nil {
		trimmed := strings.TrimSpace(*input.GroupID)
		if err := validateUUID("group ID", trimmed); err != nil {
			return "", err
		}
		input.GroupID = &trimmed
	}
	money, err := domain.NewMoney(input.AmountMinor, strings.ToUpper(strings.TrimSpace(input.Currency)))
	if err != nil || money.AmountMinor <= 0 {
		return "", invalid("settlement amount must be positive and use a valid currency")
	}
	if err := validateIdempotencyKey(input.IdempotencyKey); err != nil {
		return "", err
	}
	return service.repository.CreateSettlement(ctx, CreateSettlementCommand{
		ID: uuid.NewString(), GroupID: input.GroupID, FromUserID: actorID, ToUserID: input.ToUserID,
		Money: money, IdempotencyKey: input.IdempotencyKey,
		RequestFingerprint: fingerprint(struct {
			GroupID    *string
			FromUserID string
			ToUserID   string
			Money      domain.Money
		}{input.GroupID, actorID, input.ToUserID, money}),
	})
}

type CreateReminderCommand struct {
	ID          string
	SenderID    string
	RecipientID string
	Message     string
}

func (service *Service) CreateReminder(ctx context.Context, actorID, recipientID, message string) (string, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return "", err
	}
	if err := validateUUID("recipient ID", recipientID); err != nil {
		return "", err
	}
	message = strings.TrimSpace(message)
	if len(message) > 500 {
		return "", invalid("reminder message must be 500 characters or fewer")
	}
	command := CreateReminderCommand{ID: uuid.NewString(), SenderID: actorID, RecipientID: recipientID, Message: message}
	if _, err := service.repository.CreateReminder(ctx, command); err != nil {
		return "", err
	}
	return command.ID, nil
}

func equalSplits(participantIDs []string, amountMinor int64) ([]domain.ExpenseSplit, error) {
	participantIDs = slices.Clone(participantIDs)
	slices.Sort(participantIDs)
	participantIDs = slices.Compact(participantIDs)
	if len(participantIDs) == 0 || amountMinor < int64(len(participantIDs)) {
		return nil, invalid("equal split requires at least one participant and one minor unit per participant")
	}
	for _, userID := range participantIDs {
		if err := validateUUID("participant ID", userID); err != nil {
			return nil, err
		}
	}

	base := amountMinor / int64(len(participantIDs))
	remainder := amountMinor % int64(len(participantIDs))
	splits := make([]domain.ExpenseSplit, 0, len(participantIDs))
	for index, userID := range participantIDs {
		amount := base
		if int64(index) < remainder {
			amount++
		}
		splits = append(splits, domain.ExpenseSplit{UserID: userID, AmountMinor: amount})
	}
	return splits, nil
}

func exactSplits(input []domain.ExpenseSplit, amountMinor int64) ([]domain.ExpenseSplit, error) {
	seen := make(map[string]struct{}, len(input))
	var total int64
	for _, split := range input {
		if err := validateUUID("participant ID", split.UserID); err != nil {
			return nil, err
		}
		if split.AmountMinor <= 0 {
			return nil, invalid("exact split amounts must be positive")
		}
		if _, duplicate := seen[split.UserID]; duplicate {
			return nil, invalid("exact splits cannot contain duplicate participants")
		}
		seen[split.UserID] = struct{}{}
		total += split.AmountMinor
	}
	if len(input) == 0 || total != amountMinor {
		return nil, invalid(fmt.Sprintf("exact splits total %d, expected %d", total, amountMinor))
	}
	result := slices.Clone(input)
	slices.SortFunc(result, func(left, right domain.ExpenseSplit) int { return strings.Compare(left.UserID, right.UserID) })
	return result, nil
}

func validateUUID(label, value string) error {
	if _, err := uuid.Parse(value); err != nil {
		return invalid(label + " must be a valid UUID")
	}
	return nil
}

func validateIdempotencyKey(value string) error {
	if len(strings.TrimSpace(value)) < 8 || len(value) > 200 {
		return invalid("Idempotency-Key must contain between 8 and 200 characters")
	}
	return nil
}

func uniqueIDs(values []string, excluded string) []string {
	values = slices.Clone(values)
	slices.Sort(values)
	values = slices.Compact(values)
	return slices.DeleteFunc(values, func(value string) bool { return value == excluded })
}

func fingerprint(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("fingerprint command: %v", err))
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func invalid(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}
