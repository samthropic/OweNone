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
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/samfiallos/owenone/backend/internal/auth"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrForbidden       = errors.New("forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrSessionNotFound = errors.New("session not found")
	ErrFriendNotFound  = errors.New("friend not found")
)

const SessionLifetime = 30 * 24 * time.Hour

// DefaultCurrency is the preferred currency assigned to new accounts.
const DefaultCurrency = "USD"

var dummyHash string

func init() {
	var err error
	dummyHash, err = auth.HashPassword("dummy-timing-protection-do-not-use")
	if err != nil {
		panic(fmt.Sprintf("initialize dummy hash: %v", err))
	}
}

// CreateUserCommand carries the data needed to create a new user account.
type CreateUserCommand struct {
	ID                string
	Email             string
	DisplayName       string
	PasswordHash      string
	PreferredCurrency string
}

// UpdateUserCommand carries the data needed to update an existing user's profile.
// If PreferredCurrency is empty the store preserves the existing value.
// Payment app handles are always overwritten (empty clears a linked app).
type UpdateUserCommand struct {
	ID                string
	DisplayName       string
	Email             string
	PreferredCurrency string
	PaymentApps       domain.PaymentApps
}

// UpdateProfileInput is the validated input for the profile-update flow.
// If PreferredCurrency is empty the signed-in user's existing currency is kept.
// PaymentApps is applied only when UpdatePaymentApps is true (empty fields clear links).
type UpdateProfileInput struct {
	DisplayName        string
	Email              string
	PreferredCurrency  string
	PaymentApps        domain.PaymentApps
	UpdatePaymentApps  bool
}

// UserCredentials bundles a user record with its stored password hash.
type UserCredentials struct {
	User         domain.User
	PasswordHash string
}

// CreateSessionCommand carries the data needed to persist a session.
type CreateSessionCommand struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
}

// SignUpInput is the validated input for the sign-up flow.
type SignUpInput struct {
	Email       string
	DisplayName string
	Password    string
}

// Session is returned after a successful sign-up or log-in.
type Session struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expiresAt"`
	User      domain.User `json:"user"`
}

type Repository interface {
	Ping(context.Context) error
	LoadLedgerSnapshot(context.Context, string) (domain.LedgerSnapshot, error)
	CreateGroup(context.Context, CreateGroupCommand) error
	UpdateGroup(context.Context, UpdateGroupCommand) (domain.Group, error)
	DeleteGroup(context.Context, string, string) error
	AddFriend(context.Context, string, string) (domain.User, error)
	RemoveFriend(context.Context, string, string) error
	CreateExpense(context.Context, CreateExpenseCommand) (string, error)
	CreateSettlement(context.Context, CreateSettlementCommand) (string, error)
	CreateReminder(context.Context, CreateReminderCommand) (string, error)
	// Auth
	CreateUser(context.Context, CreateUserCommand) (domain.User, error)
	UserCredentialsByEmail(context.Context, string) (UserCredentials, error)
	UserCredentialsByID(context.Context, string) (UserCredentials, error)
	CreateSession(context.Context, CreateSessionCommand) error
	SessionUser(context.Context, string) (domain.User, error)
	DeleteSession(context.Context, string) error
	// Profile
	UpdateUser(context.Context, UpdateUserCommand) (domain.User, error)
	UpdateAvatarURL(context.Context, string, string) (domain.User, error)
	UpdatePassword(context.Context, string, string) error
	DeleteSessionsForUserExcept(context.Context, string, string) error
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

// ActivityFeed is the paginated response for GET /api/v1/activity.
type ActivityFeed struct {
	Activity []domain.Activity `json:"activity"`
	Total    int               `json:"total"`
	HasMore  bool              `json:"hasMore"`
}

// ActivityQuery carries the filter and pagination parameters for the activity feed.
// Kind filters by activity type: "expense", "settlement", or "" for all activities.
// GroupID filters to a specific group; must be a valid UUID when non-empty.
type ActivityQuery struct {
	Kind    string
	GroupID string
	Limit   int
	Offset  int
}

func (service *Service) Activity(ctx context.Context, userID string, query ActivityQuery) (ActivityFeed, error) {
	if err := validateUUID("user ID", userID); err != nil {
		return ActivityFeed{}, err
	}
	if query.Limit <= 0 {
		return ActivityFeed{}, invalid("limit must be a positive integer")
	}
	if query.Offset < 0 {
		return ActivityFeed{}, invalid("offset must be a non-negative integer")
	}
	if query.Kind != "" && query.Kind != "expense" && query.Kind != "settlement" {
		return ActivityFeed{}, invalid("kind must be expense or settlement")
	}
	if query.GroupID != "" {
		if err := validateUUID("group ID", query.GroupID); err != nil {
			return ActivityFeed{}, err
		}
	}
	if query.Limit > 200 {
		query.Limit = 200
	}

	snapshot, err := service.repository.LoadLedgerSnapshot(ctx, userID)
	if err != nil {
		return ActivityFeed{}, err
	}

	all, err := domain.BuildActivityFeed(snapshot, service.now())
	if err != nil {
		return ActivityFeed{}, err
	}

	// Filter by kind and groupId before paging so total and hasMore describe the filtered set.
	if query.Kind != "" {
		filtered := make([]domain.Activity, 0, len(all))
		for _, a := range all {
			if a.Kind == query.Kind {
				filtered = append(filtered, a)
			}
		}
		all = filtered
	}
	if query.GroupID != "" {
		filtered := make([]domain.Activity, 0, len(all))
		for _, a := range all {
			if a.GroupID == query.GroupID {
				filtered = append(filtered, a)
			}
		}
		all = filtered
	}

	total := len(all)
	start := query.Offset
	if start > total {
		start = total
	}
	end := start + query.Limit
	if end > total {
		end = total
	}

	activity := all[start:end]
	return ActivityFeed{
		Activity: activity,
		Total:    total,
		HasMore:  query.Offset+len(activity) < total,
	}, nil
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
	if utf8.RuneCountInString(input.Icon) > 16 {
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

type UpdateGroupInput struct {
	GroupID   string
	Name      string
	Icon      string
	MemberIDs []string
}

type UpdateGroupCommand struct {
	ActorID   string
	GroupID   string
	Name      string
	Icon      string
	MemberIDs []string
}

func (service *Service) UpdateGroup(ctx context.Context, actorID string, input UpdateGroupInput) (domain.Group, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return domain.Group{}, err
	}
	if err := validateUUID("group ID", input.GroupID); err != nil {
		return domain.Group{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	if len(input.Name) < 1 || len(input.Name) > 100 {
		return domain.Group{}, invalid("group name must contain between 1 and 100 characters")
	}
	input.Icon = strings.TrimSpace(input.Icon)
	if input.Icon == "" {
		return domain.Group{}, invalid("group icon must not be empty")
	}
	if utf8.RuneCountInString(input.Icon) > 16 {
		return domain.Group{}, invalid("group icon is too long")
	}
	input.MemberIDs = uniqueIDs(input.MemberIDs, actorID)
	for _, memberID := range input.MemberIDs {
		if err := validateUUID("member ID", memberID); err != nil {
			return domain.Group{}, err
		}
	}
	return service.repository.UpdateGroup(ctx, UpdateGroupCommand{
		ActorID: actorID, GroupID: input.GroupID, Name: input.Name, Icon: input.Icon, MemberIDs: input.MemberIDs,
	})
}

func (service *Service) DeleteGroup(ctx context.Context, actorID, groupID string) error {
	if err := validateUUID("actor ID", actorID); err != nil {
		return err
	}
	if err := validateUUID("group ID", groupID); err != nil {
		return err
	}
	return service.repository.DeleteGroup(ctx, actorID, groupID)
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

func (service *Service) RemoveFriend(ctx context.Context, actorID, email string) (domain.User, error) {
	if err := validateUUID("actor ID", actorID); err != nil {
		return domain.User{}, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return domain.User{}, invalid("email must be a valid address")
	}

	snapshot, err := service.repository.LoadLedgerSnapshot(ctx, actorID)
	if err != nil {
		return domain.User{}, err
	}
	dashboard, err := domain.BuildDashboard(snapshot, service.now())
	if err != nil {
		return domain.User{}, err
	}

	var target *domain.FriendBalance
	for i := range dashboard.Friends {
		if strings.EqualFold(dashboard.Friends[i].User.Email, email) {
			target = &dashboard.Friends[i]
			break
		}
	}
	if target == nil || !target.IsFriend {
		return domain.User{}, fmt.Errorf("%w: no friend with that email", ErrFriendNotFound)
	}
	if actorID == target.User.ID {
		return domain.User{}, invalid("you cannot remove yourself")
	}
	if target.Balance.AmountMinor != 0 {
		return domain.User{}, invalid(fmt.Sprintf("settle the outstanding balance with %s before removing them", target.User.DisplayName))
	}
	if err := service.repository.RemoveFriend(ctx, actorID, target.User.ID); err != nil {
		return domain.User{}, err
	}
	return target.User, nil
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
	if err := requireUSD(money.Currency); err != nil {
		return "", err
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
	PaymentMethod  string
	IdempotencyKey string
}

type CreateSettlementCommand struct {
	ID                 string
	GroupID            *string
	FromUserID         string
	ToUserID           string
	Money              domain.Money
	PaymentMethod      string
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
	if err := requireUSD(money.Currency); err != nil {
		return "", err
	}
	paymentMethod, err := normalizePaymentMethod(input.PaymentMethod)
	if err != nil {
		return "", err
	}
	if err := validateIdempotencyKey(input.IdempotencyKey); err != nil {
		return "", err
	}
	return service.repository.CreateSettlement(ctx, CreateSettlementCommand{
		ID: uuid.NewString(), GroupID: input.GroupID, FromUserID: actorID, ToUserID: input.ToUserID,
		Money: money, PaymentMethod: paymentMethod, IdempotencyKey: input.IdempotencyKey,
		RequestFingerprint: fingerprint(struct {
			GroupID       *string
			FromUserID    string
			ToUserID      string
			Money         domain.Money
			PaymentMethod string
		}{input.GroupID, actorID, input.ToUserID, money, paymentMethod}),
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

func (service *Service) SignUp(ctx context.Context, input SignUpInput) (Session, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	parsed, err := mail.ParseAddress(input.Email)
	if err != nil || parsed.Address != input.Email || len(input.Email) > 320 {
		return Session{}, invalid("email must be a valid address")
	}
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if len(input.DisplayName) < 1 || len(input.DisplayName) > 100 {
		return Session{}, invalid("display name must contain between 1 and 100 characters")
	}
	if err := validatePassword(input.Password); err != nil {
		return Session{}, err
	}
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return Session{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := service.repository.CreateUser(ctx, CreateUserCommand{
		ID: uuid.NewString(), Email: input.Email, DisplayName: input.DisplayName,
		PasswordHash: passwordHash, PreferredCurrency: DefaultCurrency,
	})
	if err != nil {
		return Session{}, err
	}
	return service.issueSession(ctx, user)
}

func (service *Service) LogIn(ctx context.Context, email, password string) (Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	creds, err := service.repository.UserCredentialsByEmail(ctx, email)
	if err != nil {
		// User not found: run dummy verification to resist timing oracle.
		auth.VerifyPassword(dummyHash, password)
		return Session{}, fmt.Errorf("%w: email or password is incorrect", ErrUnauthorized)
	}
	if creds.PasswordHash == "" || !auth.VerifyPassword(creds.PasswordHash, password) {
		return Session{}, fmt.Errorf("%w: email or password is incorrect", ErrUnauthorized)
	}
	return service.issueSession(ctx, creds.User)
}

func (service *Service) LogOut(ctx context.Context, token string) error {
	return service.repository.DeleteSession(ctx, auth.HashToken(token))
}

func (service *Service) Authenticate(ctx context.Context, token string) (domain.User, error) {
	user, err := service.repository.SessionUser(ctx, auth.HashToken(token))
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return domain.User{}, fmt.Errorf("%w: session not found or expired", ErrUnauthorized)
		}
		return domain.User{}, err
	}
	return user, nil
}

func (service *Service) issueSession(ctx context.Context, user domain.User) (Session, error) {
	token, err := auth.NewSessionToken()
	if err != nil {
		return Session{}, fmt.Errorf("generate session token: %w", err)
	}
	expiresAt := service.now().Add(SessionLifetime)
	if err := service.repository.CreateSession(ctx, CreateSessionCommand{
		TokenHash: auth.HashToken(token),
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return Session{}, err
	}
	return Session{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func (service *Service) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (domain.User, error) {
	if err := validateUUID("user ID", userID); err != nil {
		return domain.User{}, err
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	parsed, err := mail.ParseAddress(input.Email)
	if err != nil || parsed.Address != input.Email || len(input.Email) > 320 {
		return domain.User{}, invalid("email must be a valid address")
	}
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if len(input.DisplayName) < 1 || len(input.DisplayName) > 100 {
		return domain.User{}, invalid("display name must contain between 1 and 100 characters")
	}
	input.PreferredCurrency = strings.ToUpper(strings.TrimSpace(input.PreferredCurrency))
	if input.PreferredCurrency != "" {
		if err := requireUSD(input.PreferredCurrency); err != nil {
			return domain.User{}, err
		}
	}
	apps := input.PaymentApps
	if input.UpdatePaymentApps {
		normalized, err := normalizePaymentApps(input.PaymentApps)
		if err != nil {
			return domain.User{}, err
		}
		apps = normalized
	} else {
		creds, err := service.repository.UserCredentialsByID(ctx, userID)
		if err != nil {
			return domain.User{}, err
		}
		apps = creds.User.PaymentApps
	}
	return service.repository.UpdateUser(ctx, UpdateUserCommand{
		ID:                userID,
		DisplayName:       input.DisplayName,
		Email:             input.Email,
		PreferredCurrency: input.PreferredCurrency,
		PaymentApps:       apps,
	})
}

func (service *Service) SetAvatarURL(ctx context.Context, userID, avatarURL string) (domain.User, error) {
	if err := validateUUID("user ID", userID); err != nil {
		return domain.User{}, err
	}
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return domain.User{}, invalid("avatar URL is required")
	}
	if len(avatarURL) > 512 {
		return domain.User{}, invalid("avatar URL is too long")
	}
	return service.repository.UpdateAvatarURL(ctx, userID, avatarURL)
}

func (service *Service) ClearAvatarURL(ctx context.Context, userID string) (domain.User, error) {
	if err := validateUUID("user ID", userID); err != nil {
		return domain.User{}, err
	}
	return service.repository.UpdateAvatarURL(ctx, userID, "")
}

// ChangePassword verifies currentPassword, validates and hashes newPassword, persists
// the new hash, then revokes every session for the user except the caller's current one
// (identified by currentToken) so they stay logged in on the tab they are using.
func (service *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword, currentToken string) error {
	if err := validateUUID("user ID", userID); err != nil {
		return err
	}
	creds, err := service.repository.UserCredentialsByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	if creds.PasswordHash == "" || !auth.VerifyPassword(creds.PasswordHash, currentPassword) {
		return fmt.Errorf("%w: current password is incorrect", ErrUnauthorized)
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	if auth.VerifyPassword(creds.PasswordHash, newPassword) {
		return invalid("new password must be different from the current password")
	}
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := service.repository.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}
	return service.repository.DeleteSessionsForUserExcept(ctx, userID, auth.HashToken(currentToken))
}

func validatePassword(password string) error {
	runes := []rune(password)
	if len(runes) < 8 || len(runes) > 128 {
		return invalid("password must be between 8 and 128 characters")
	}
	if strings.TrimSpace(password) == "" {
		return invalid("password must not be all whitespace")
	}
	return nil
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

func requireUSD(currency string) error {
	if currency != DefaultCurrency {
		return invalid("only USD is supported")
	}
	return nil
}

func normalizePaymentMethod(value string) (string, error) {
	method := strings.TrimSpace(value)
	switch method {
	case "", "venmo", "paypal", "cashApp", "zelle", "other":
		return method, nil
	default:
		return "", invalid("paymentMethod must be venmo, paypal, cashApp, zelle, other, or empty")
	}
}

func normalizePaymentApps(apps domain.PaymentApps) (domain.PaymentApps, error) {
	venmo, err := normalizeHandle("Venmo", apps.Venmo, 64, "@")
	if err != nil {
		return domain.PaymentApps{}, err
	}
	paypal, err := normalizeHandle("PayPal", apps.PayPal, 64, "@")
	if err != nil {
		return domain.PaymentApps{}, err
	}
	cashapp, err := normalizeHandle("Cash App", apps.CashApp, 64, "$@")
	if err != nil {
		return domain.PaymentApps{}, err
	}
	zelle, err := normalizeZelle(apps.Zelle)
	if err != nil {
		return domain.PaymentApps{}, err
	}
	return domain.PaymentApps{
		Venmo:   venmo,
		PayPal:  paypal,
		CashApp: cashapp,
		Zelle:   zelle,
	}, nil
}

func normalizeHandle(label, value string, maxLen int, stripPrefix string) (string, error) {
	value = strings.TrimSpace(value)
	for _, prefix := range strings.Split(stripPrefix, "") {
		value = strings.TrimPrefix(value, prefix)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if utf8.RuneCountInString(value) > maxLen {
		return "", invalid(label + " handle is too long")
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return "", invalid(label + " handle can only use letters, numbers, dots, hyphens, and underscores")
	}
	return value, nil
}

func normalizeZelle(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if utf8.RuneCountInString(value) > 128 {
		return "", invalid("Zelle contact is too long")
	}
	// Accept email or a simple phone-like string (digits, spaces, +, -, parentheses).
	if parsed, err := mail.ParseAddress(value); err == nil && parsed.Address == value {
		return strings.ToLower(value), nil
	}
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r >= '0' && r <= '9', r == '+', r == '-', r == '(', r == ')', r == ' ':
			return r
		default:
			return -1
		}
	}, value)
	digits := 0
	for _, r := range cleaned {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	if digits >= 7 {
		return strings.TrimSpace(cleaned), nil
	}
	return "", invalid("Zelle contact must be an email or phone number")
}

func invalid(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}
