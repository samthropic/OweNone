package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/samfiallos/owenone/backend/internal/auth"
	"github.com/samfiallos/owenone/backend/internal/domain"
)

// fakeGroup holds in-memory group state for the fakeRepository.
type fakeGroup struct {
	id          string
	ownerID     string
	name        string
	icon        string
	members     []string
	hasExpenses bool
}

// fakeRepository is an in-memory Repository for service-level tests.
type fakeRepository struct {
	emailCreds      map[string]UserCredentials // keyed by lower(email)
	sessions        map[string]string          // tokenHash -> userID
	snapshot        domain.LedgerSnapshot
	removeFriendErr error
	removedActorID  string
	removedFriendID string
	groups          map[string]fakeGroup
	friendships     map[string]bool // "a|b" stored both ways; actor a is friends with b
}

var (
	errFakeConflict = errors.New("email already exists")
	errFakeNotFound = errors.New("not found")
)

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		emailCreds:  make(map[string]UserCredentials),
		sessions:    make(map[string]string),
		groups:      make(map[string]fakeGroup),
		friendships: make(map[string]bool),
	}
}

func (r *fakeRepository) makeFriends(a, b string) {
	r.friendships[a+"|"+b] = true
	r.friendships[b+"|"+a] = true
}

func (r *fakeRepository) areFriends(a, b string) bool {
	return r.friendships[a+"|"+b]
}

func (r *fakeRepository) Ping(context.Context) error { return nil }
func (r *fakeRepository) LoadLedgerSnapshot(_ context.Context, _ string) (domain.LedgerSnapshot, error) {
	return r.snapshot, nil
}
func (r *fakeRepository) CreateGroup(_ context.Context, cmd CreateGroupCommand) error {
	r.groups[cmd.ID] = fakeGroup{id: cmd.ID, ownerID: cmd.CreatorID, name: cmd.Name, icon: cmd.Icon, members: cmd.MemberIDs}
	return nil
}
func (r *fakeRepository) UpdateGroup(_ context.Context, cmd UpdateGroupCommand) (domain.Group, error) {
	g, ok := r.groups[cmd.GroupID]
	if !ok {
		return domain.Group{}, errFakeNotFound
	}
	if g.ownerID != cmd.ActorID {
		return domain.Group{}, fmt.Errorf("%w: only the group owner can edit this group", ErrForbidden)
	}
	// Build current member set.
	currentSet := make(map[string]struct{}, len(g.members))
	for _, id := range g.members {
		currentSet[id] = struct{}{}
	}
	// Build desired set, filtering out the actor (owner must not appear as regular member).
	desiredSet := make(map[string]struct{}, len(cmd.MemberIDs))
	for _, id := range cmd.MemberIDs {
		if id != cmd.ActorID {
			desiredSet[id] = struct{}{}
		}
	}
	// Friend-check only newly added members; existing members are grandfathered.
	for id := range desiredSet {
		if _, existing := currentSet[id]; !existing {
			if !r.areFriends(cmd.ActorID, id) {
				return domain.Group{}, fmt.Errorf("%w: group members must be existing friends", ErrForbidden)
			}
		}
	}
	newMembers := make([]string, 0, len(desiredSet))
	for id := range desiredSet {
		newMembers = append(newMembers, id)
	}
	g.name = cmd.Name
	g.icon = cmd.Icon
	g.members = newMembers
	r.groups[cmd.GroupID] = g
	members := make([]domain.User, len(newMembers))
	for i, id := range newMembers {
		members[i] = domain.User{ID: id}
	}
	return domain.Group{ID: g.id, Name: g.name, Icon: g.icon, Members: members, OwnerID: g.ownerID}, nil
}
func (r *fakeRepository) DeleteGroup(_ context.Context, actorID, groupID string) error {
	g, ok := r.groups[groupID]
	if !ok {
		return errFakeNotFound
	}
	if g.ownerID != actorID {
		return fmt.Errorf("%w: only the group owner can delete this group", ErrForbidden)
	}
	if g.hasExpenses {
		return fmt.Errorf("%w: this group has recorded expenses, so it cannot be deleted", ErrInvalidInput)
	}
	delete(r.groups, groupID)
	return nil
}
func (r *fakeRepository) AddFriend(_ context.Context, _, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (r *fakeRepository) RemoveFriend(_ context.Context, actorID, friendID string) error {
	r.removedActorID = actorID
	r.removedFriendID = friendID
	return r.removeFriendErr
}
func (r *fakeRepository) CreateExpense(_ context.Context, _ CreateExpenseCommand) (string, error) {
	return "", nil
}
func (r *fakeRepository) CreateSettlement(_ context.Context, _ CreateSettlementCommand) (string, error) {
	return "", nil
}
func (r *fakeRepository) CreateReminder(_ context.Context, _ CreateReminderCommand) (string, error) {
	return "", nil
}
func (r *fakeRepository) CreateUser(_ context.Context, cmd CreateUserCommand) (domain.User, error) {
	lower := strings.ToLower(cmd.Email)
	if _, exists := r.emailCreds[lower]; exists {
		return domain.User{}, errFakeConflict
	}
	user := domain.User{ID: cmd.ID, Email: cmd.Email, DisplayName: cmd.DisplayName, PreferredCurrency: cmd.PreferredCurrency}
	r.emailCreds[lower] = UserCredentials{User: user, PasswordHash: cmd.PasswordHash}
	return user, nil
}
func (r *fakeRepository) UserCredentialsByEmail(_ context.Context, email string) (UserCredentials, error) {
	creds, ok := r.emailCreds[strings.ToLower(email)]
	if !ok {
		return UserCredentials{}, errFakeNotFound
	}
	return creds, nil
}
func (r *fakeRepository) CreateSession(_ context.Context, cmd CreateSessionCommand) error {
	r.sessions[cmd.TokenHash] = cmd.UserID
	return nil
}
func (r *fakeRepository) SessionUser(_ context.Context, tokenHash string) (domain.User, error) {
	userID, ok := r.sessions[tokenHash]
	if !ok {
		return domain.User{}, ErrSessionNotFound
	}
	return domain.User{ID: userID}, nil
}
func (r *fakeRepository) DeleteSession(_ context.Context, _ string) error { return nil }

func (r *fakeRepository) UpdateUser(_ context.Context, cmd UpdateUserCommand) (domain.User, error) {
	var oldKey string
	var existing UserCredentials
	for key, creds := range r.emailCreds {
		if creds.User.ID == cmd.ID {
			oldKey = key
			existing = creds
			break
		}
	}
	if oldKey == "" {
		return domain.User{}, errFakeNotFound
	}
	newKey := strings.ToLower(cmd.Email)
	if newKey != oldKey {
		if _, exists := r.emailCreds[newKey]; exists {
			return domain.User{}, errFakeConflict
		}
	}
	currency := cmd.PreferredCurrency
	if currency == "" {
		currency = existing.User.PreferredCurrency
	}
	user := domain.User{
		ID: cmd.ID, Email: cmd.Email, DisplayName: cmd.DisplayName,
		PreferredCurrency: currency, PaymentApps: cmd.PaymentApps,
		AvatarURL: existing.User.AvatarURL,
	}
	delete(r.emailCreds, oldKey)
	r.emailCreds[newKey] = UserCredentials{User: user, PasswordHash: existing.PasswordHash}
	return user, nil
}

func (r *fakeRepository) UpdateAvatarURL(_ context.Context, userID, avatarURL string) (domain.User, error) {
	for key, creds := range r.emailCreds {
		if creds.User.ID == userID {
			user := creds.User
			user.AvatarURL = avatarURL
			r.emailCreds[key] = UserCredentials{User: user, PasswordHash: creds.PasswordHash}
			return user, nil
		}
	}
	return domain.User{}, errFakeNotFound
}

func (r *fakeRepository) UserCredentialsByID(_ context.Context, id string) (UserCredentials, error) {
	for _, creds := range r.emailCreds {
		if creds.User.ID == id {
			return creds, nil
		}
	}
	return UserCredentials{}, errFakeNotFound
}

func (r *fakeRepository) UpdatePassword(_ context.Context, userID, passwordHash string) error {
	for key, creds := range r.emailCreds {
		if creds.User.ID == userID {
			r.emailCreds[key] = UserCredentials{User: creds.User, PasswordHash: passwordHash}
			return nil
		}
	}
	return errFakeNotFound
}

func (r *fakeRepository) DeleteSessionsForUserExcept(_ context.Context, userID, keepTokenHash string) error {
	for tokenHash, uid := range r.sessions {
		if uid == userID && tokenHash != keepTokenHash {
			delete(r.sessions, tokenHash)
		}
	}
	return nil
}

// --- Split helper tests (pre-existing) ---

func TestEqualSplitsDistributesRemainderDeterministically(t *testing.T) {
	participants := []string{
		"30000000-0000-0000-0000-000000000003",
		"10000000-0000-0000-0000-000000000001",
		"20000000-0000-0000-0000-000000000002",
	}

	splits, err := equalSplits(participants, 1000)
	if err != nil {
		t.Fatalf("equalSplits() error = %v", err)
	}
	wantAmounts := []int64{334, 333, 333}
	for index, split := range splits {
		if split.AmountMinor != wantAmounts[index] {
			t.Errorf("split %d amount = %d, want %d", index, split.AmountMinor, wantAmounts[index])
		}
	}
	if splits[0].UserID != "10000000-0000-0000-0000-000000000001" {
		t.Errorf("remainder recipient = %s, want lowest UUID", splits[0].UserID)
	}
}

func TestExactSplitsRejectsWrongTotal(t *testing.T) {
	_, err := exactSplits(nil, 100)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("exactSplits() error = %v, want ErrInvalidInput", err)
	}
}

// --- Auth service tests ---

func TestSignUpRejectsShortPassword(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.SignUp(context.Background(), SignUpInput{
		Email: "test@example.com", DisplayName: "Test User", Password: "short",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("SignUp with short password = %v, want ErrInvalidInput", err)
	}
}

func TestSignUpRejectsAllWhitespacePassword(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.SignUp(context.Background(), SignUpInput{
		Email: "test@example.com", DisplayName: "Test User", Password: "        ",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("SignUp with whitespace password = %v, want ErrInvalidInput", err)
	}
}

func TestSignUpOnDuplicateEmailSurfacesConflict(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	if _, err := svc.SignUp(context.Background(), SignUpInput{
		Email: "dup@example.com", DisplayName: "First", Password: "password123",
	}); err != nil {
		t.Fatalf("first SignUp failed: %v", err)
	}
	_, err := svc.SignUp(context.Background(), SignUpInput{
		Email: "dup@example.com", DisplayName: "Second", Password: "password123",
	})
	if !errors.Is(err, errFakeConflict) {
		t.Errorf("duplicate SignUp = %v, want conflict error", err)
	}
}

func TestLogInWithWrongPasswordReturnsUnauthorized(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	repo.emailCreds["test@example.com"] = UserCredentials{
		User:         domain.User{ID: "10000000-0000-0000-0000-000000000001", Email: "test@example.com"},
		PasswordHash: hash,
	}

	_, err = svc.LogIn(context.Background(), "test@example.com", "wrong-password")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("LogIn with wrong password = %v, want ErrUnauthorized", err)
	}
}

func TestLogInWithUnknownEmailReturnsUnauthorized(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.LogIn(context.Background(), "nobody@example.com", "password123")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("LogIn with unknown email = %v, want ErrUnauthorized", err)
	}
}

func TestSignUpAndLogInHappyPath(t *testing.T) {
	svc := NewService(newFakeRepository())

	session, err := svc.SignUp(context.Background(), SignUpInput{
		Email: "happy@example.com", DisplayName: "Happy User", Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if session.Token == "" {
		t.Error("SignUp returned empty token")
	}
	if session.User.Email != "happy@example.com" {
		t.Errorf("SignUp user email = %q, want happy@example.com", session.User.Email)
	}
	if session.User.PreferredCurrency != "USD" {
		t.Errorf("SignUp preferred currency = %q, want USD (default)", session.User.PreferredCurrency)
	}

	session2, err := svc.LogIn(context.Background(), "happy@example.com", "password123")
	if err != nil {
		t.Fatalf("LogIn: %v", err)
	}
	if session2.Token == "" {
		t.Error("LogIn returned empty token")
	}
}

// --- RemoveFriend tests ---

const (
	removeFriendActorID     = "10000000-0000-0000-0000-000000000001"
	removeFriendFriendID    = "20000000-0000-0000-0000-000000000002"
	removeFriendGroupOnlyID = "30000000-0000-0000-0000-000000000003"
)

func snapshotWithFriend() domain.LedgerSnapshot {
	return domain.LedgerSnapshot{
		CurrentUser: domain.User{ID: removeFriendActorID, PreferredCurrency: "GBP"},
		Friends:     []domain.User{{ID: removeFriendFriendID, Email: "friend@example.com", DisplayName: "Friend"}},
	}
}

func TestRemoveFriendSucceedsAndCallsRepository(t *testing.T) {
	repo := newFakeRepository()
	repo.snapshot = snapshotWithFriend()
	svc := NewService(repo)

	user, err := svc.RemoveFriend(context.Background(), removeFriendActorID, "friend@example.com")
	if err != nil {
		t.Fatalf("RemoveFriend() error = %v", err)
	}
	if user.ID != removeFriendFriendID {
		t.Errorf("removed user ID = %q, want %q", user.ID, removeFriendFriendID)
	}
	if repo.removedActorID != removeFriendActorID {
		t.Errorf("repository called with actor %q, want %q", repo.removedActorID, removeFriendActorID)
	}
	if repo.removedFriendID != removeFriendFriendID {
		t.Errorf("repository called with friend %q, want %q", repo.removedFriendID, removeFriendFriendID)
	}
}

func TestRemoveFriendBlockedByOutstandingBalance(t *testing.T) {
	repo := newFakeRepository()
	repo.snapshot = domain.LedgerSnapshot{
		CurrentUser: domain.User{ID: removeFriendActorID, PreferredCurrency: "GBP"},
		Friends:     []domain.User{{ID: removeFriendFriendID, Email: "friend@example.com", DisplayName: "Friend"}},
		Expenses: []domain.Expense{{
			ID:          "e1",
			GroupID:     "g1",
			PayerID:     removeFriendActorID,
			CreatedByID: removeFriendActorID,
			Money:       domain.Money{AmountMinor: 500, Currency: "GBP"},
			SplitMethod: "exact",
			CreatedAt:   time.Now(),
			Splits:      []domain.ExpenseSplit{{UserID: removeFriendFriendID, AmountMinor: 500}},
		}},
	}
	svc := NewService(repo)

	_, err := svc.RemoveFriend(context.Background(), removeFriendActorID, "friend@example.com")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("RemoveFriend with outstanding balance = %v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), "settle the outstanding balance") {
		t.Errorf("error message = %q, want settle outstanding balance message", err.Error())
	}
}

func TestRemoveFriendUnknownEmailReturnsNotFound(t *testing.T) {
	repo := newFakeRepository()
	repo.snapshot = snapshotWithFriend()
	svc := NewService(repo)

	_, err := svc.RemoveFriend(context.Background(), removeFriendActorID, "nobody@example.com")
	if !errors.Is(err, ErrFriendNotFound) {
		t.Errorf("RemoveFriend unknown email = %v, want ErrFriendNotFound", err)
	}
}

func TestRemoveFriendGroupOnlyContactReturnsNotFound(t *testing.T) {
	repo := newFakeRepository()
	repo.snapshot = domain.LedgerSnapshot{
		CurrentUser: domain.User{ID: removeFriendActorID, PreferredCurrency: "GBP"},
		Friends:     []domain.User{},
		Groups: []domain.Group{{
			ID:   "g1",
			Name: "Group",
			Members: []domain.User{
				{ID: removeFriendActorID, DisplayName: "Actor"},
				{ID: removeFriendGroupOnlyID, Email: "grouponly@example.com", DisplayName: "GroupOnly"},
			},
		}},
	}
	svc := NewService(repo)

	_, err := svc.RemoveFriend(context.Background(), removeFriendActorID, "grouponly@example.com")
	if !errors.Is(err, ErrFriendNotFound) {
		t.Errorf("RemoveFriend group-only contact = %v, want ErrFriendNotFound", err)
	}
}

func TestRemoveFriendInvalidEmailReturnsInvalidInput(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.RemoveFriend(context.Background(), removeFriendActorID, "not-an-email")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("RemoveFriend invalid email = %v, want ErrInvalidInput", err)
	}
}

// --- UpdateProfile tests ---

const profileUserID = "50000000-0000-0000-0000-000000000005"

func seedProfileUser(repo *fakeRepository) {
	repo.emailCreds["alice@example.com"] = UserCredentials{
		User:         domain.User{ID: profileUserID, Email: "alice@example.com", DisplayName: "Alice", PreferredCurrency: "GBP"},
		PasswordHash: "placeholder",
	}
}

func TestUpdateProfileSucceedsAndNormalisesInput(t *testing.T) {
	repo := newFakeRepository()
	seedProfileUser(repo)
	svc := NewService(repo)

	user, err := svc.UpdateProfile(context.Background(), profileUserID, UpdateProfileInput{
		DisplayName:       "  Alice Updated  ",
		Email:             "  ALICE@Example.COM  ",
		PreferredCurrency: "usd",
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if user.DisplayName != "Alice Updated" {
		t.Errorf("display name = %q, want trimmed value", user.DisplayName)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("email = %q, want lowercased value", user.Email)
	}
	if user.PreferredCurrency != "USD" {
		t.Errorf("preferred currency = %q, want uppercased value", user.PreferredCurrency)
	}
}

func TestUpdateProfileKeepsExistingCurrencyWhenEmpty(t *testing.T) {
	repo := newFakeRepository()
	seedProfileUser(repo)
	svc := NewService(repo)

	user, err := svc.UpdateProfile(context.Background(), profileUserID, UpdateProfileInput{
		DisplayName:       "Alice",
		Email:             "alice@example.com",
		PreferredCurrency: "",
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if user.PreferredCurrency != "GBP" {
		t.Errorf("preferred currency = %q, want existing GBP", user.PreferredCurrency)
	}
}

func TestUpdateProfileRejectsInvalidEmail(t *testing.T) {
	repo := newFakeRepository()
	seedProfileUser(repo)
	svc := NewService(repo)

	_, err := svc.UpdateProfile(context.Background(), profileUserID, UpdateProfileInput{
		DisplayName: "Alice", Email: "not-an-email", PreferredCurrency: "USD",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("invalid email = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateProfileRejectsEmptyDisplayName(t *testing.T) {
	repo := newFakeRepository()
	seedProfileUser(repo)
	svc := NewService(repo)

	_, err := svc.UpdateProfile(context.Background(), profileUserID, UpdateProfileInput{
		DisplayName: "   ", Email: "alice@example.com", PreferredCurrency: "USD",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("empty display name = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateProfileRejectsBadCurrencyCode(t *testing.T) {
	repo := newFakeRepository()
	seedProfileUser(repo)
	svc := NewService(repo)

	_, err := svc.UpdateProfile(context.Background(), profileUserID, UpdateProfileInput{
		DisplayName: "Alice", Email: "alice@example.com", PreferredCurrency: "GBP",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("non-USD currency = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateProfileDuplicateEmailSurfacesConflict(t *testing.T) {
	repo := newFakeRepository()
	seedProfileUser(repo)
	repo.emailCreds["bob@example.com"] = UserCredentials{
		User: domain.User{ID: "60000000-0000-0000-0000-000000000006", Email: "bob@example.com", DisplayName: "Bob", PreferredCurrency: "GBP"},
	}
	svc := NewService(repo)

	_, err := svc.UpdateProfile(context.Background(), profileUserID, UpdateProfileInput{
		DisplayName: "Alice", Email: "bob@example.com", PreferredCurrency: "USD",
	})
	if !errors.Is(err, errFakeConflict) {
		t.Errorf("duplicate email = %v, want errFakeConflict", err)
	}
}

// --- ChangePassword tests ---

func TestChangePasswordSucceedsAndUpdatesHash(t *testing.T) {
	repo := newFakeRepository()
	oldHash, err := auth.HashPassword("old-password-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	repo.emailCreds["pw@example.com"] = UserCredentials{
		User:         domain.User{ID: profileUserID, Email: "pw@example.com", DisplayName: "PW"},
		PasswordHash: oldHash,
	}
	keepToken := "raw-keep-token"
	keepHash := auth.HashToken(keepToken)
	otherHash := auth.HashToken("other-raw-token")
	repo.sessions[keepHash] = profileUserID
	repo.sessions[otherHash] = profileUserID
	svc := NewService(repo)

	if err := svc.ChangePassword(context.Background(), profileUserID, "old-password-123", "new-password-456", keepToken); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}

	creds, _ := repo.UserCredentialsByID(context.Background(), profileUserID)
	if !auth.VerifyPassword(creds.PasswordHash, "new-password-456") {
		t.Error("new password hash does not verify for the new password")
	}
	if auth.VerifyPassword(creds.PasswordHash, "old-password-123") {
		t.Error("new password hash verifies for the old password (should not)")
	}
	if _, ok := repo.sessions[keepHash]; !ok {
		t.Error("current session was revoked but should be kept")
	}
	if _, ok := repo.sessions[otherHash]; ok {
		t.Error("other session was not revoked after password change")
	}
}

func TestChangePasswordWrongCurrentPasswordReturnsUnauthorized(t *testing.T) {
	repo := newFakeRepository()
	hash, _ := auth.HashPassword("correct-pass-123")
	repo.emailCreds["pw@example.com"] = UserCredentials{
		User:         domain.User{ID: profileUserID, Email: "pw@example.com"},
		PasswordHash: hash,
	}
	svc := NewService(repo)

	err := svc.ChangePassword(context.Background(), profileUserID, "wrong-password", "new-password-456", "token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("wrong current password = %v, want ErrUnauthorized", err)
	}
}

func TestChangePasswordTooShortNewPasswordReturnsInvalidInput(t *testing.T) {
	repo := newFakeRepository()
	hash, _ := auth.HashPassword("correct-pass-123")
	repo.emailCreds["pw@example.com"] = UserCredentials{
		User:         domain.User{ID: profileUserID, Email: "pw@example.com"},
		PasswordHash: hash,
	}
	svc := NewService(repo)

	err := svc.ChangePassword(context.Background(), profileUserID, "correct-pass-123", "short", "token")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("short new password = %v, want ErrInvalidInput", err)
	}
}

func TestChangePasswordSamePasswordReturnsInvalidInput(t *testing.T) {
	repo := newFakeRepository()
	hash, _ := auth.HashPassword("same-password-123")
	repo.emailCreds["pw@example.com"] = UserCredentials{
		User:         domain.User{ID: profileUserID, Email: "pw@example.com"},
		PasswordHash: hash,
	}
	svc := NewService(repo)

	err := svc.ChangePassword(context.Background(), profileUserID, "same-password-123", "same-password-123", "token")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("same password = %v, want ErrInvalidInput", err)
	}
}

// --- Group management tests ---

const (
	groupOwnerID  = "10000000-0000-0000-0000-000000000001"
	groupMemberID = "20000000-0000-0000-0000-000000000002"
	testGroupID   = "30000000-0000-0000-0000-000000000003"
)

func TestUpdateGroupSucceedsAndNormalisesInput(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{id: testGroupID, ownerID: groupOwnerID, name: "Old", icon: "🏠"}
	svc := NewService(repo)

	group, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID: testGroupID, Name: "  New Name  ", Icon: "🚀", MemberIDs: []string{},
	})
	if err != nil {
		t.Fatalf("UpdateGroup() error = %v", err)
	}
	if group.Name != "New Name" {
		t.Errorf("group.Name = %q, want %q", group.Name, "New Name")
	}
	if group.Icon != "🚀" {
		t.Errorf("group.Icon = %q, want %q", group.Icon, "🚀")
	}
}

func TestUpdateGroupRejectsBlankName(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID: testGroupID, Name: "   ", Icon: "🏠", MemberIDs: nil,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("blank name = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateGroupRejectsOverLongName(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID: testGroupID, Name: strings.Repeat("a", 101), Icon: "🏠", MemberIDs: nil,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("over-long name = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateGroupByNonOwnerReturnsForbidden(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️"}
	svc := NewService(repo)

	_, err := svc.UpdateGroup(context.Background(), groupMemberID, UpdateGroupInput{
		GroupID: testGroupID, Name: "Trip", Icon: "✈️", MemberIDs: nil,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("non-owner update = %v, want ErrForbidden", err)
	}
}

func TestUpdateGroupUnknownGroupReturnsError(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID: testGroupID, Name: "Trip", Icon: "✈️", MemberIDs: nil,
	})
	if err == nil {
		t.Error("unknown group = nil, want error")
	}
}

func TestDeleteGroupSucceeds(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️"}
	svc := NewService(repo)

	if err := svc.DeleteGroup(context.Background(), groupOwnerID, testGroupID); err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}
	if _, exists := repo.groups[testGroupID]; exists {
		t.Error("group still present after deletion")
	}
}

func TestDeleteGroupWithExpensesReturnsInvalidInput(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️", hasExpenses: true}
	svc := NewService(repo)

	err := svc.DeleteGroup(context.Background(), groupOwnerID, testGroupID)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("delete with expenses = %v, want ErrInvalidInput", err)
	}
}

func TestDeleteGroupByNonOwnerReturnsForbidden(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️"}
	svc := NewService(repo)

	err := svc.DeleteGroup(context.Background(), groupMemberID, testGroupID)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("non-owner delete = %v, want ErrForbidden", err)
	}
}

func TestDeleteGroupUnknownGroupReturnsError(t *testing.T) {
	svc := NewService(newFakeRepository())
	err := svc.DeleteGroup(context.Background(), groupOwnerID, testGroupID)
	if err == nil {
		t.Error("unknown group = nil, want error")
	}
}

const newMemberID = "40000000-0000-0000-0000-000000000004"

func TestUpdateGroupKeepsExistingNonFriendMember(t *testing.T) {
	// Regression: an existing member who was unfriended after joining must not block edits.
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{
		id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️",
		members: []string{groupMemberID}, // member is already in the group
	}
	// groupOwnerID and groupMemberID are NOT friends (no entry in repo.friendships)
	svc := NewService(repo)

	_, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID:   testGroupID,
		Name:      "Trip",
		Icon:      "✈️",
		MemberIDs: []string{groupMemberID}, // keep the existing non-friend member
	})
	if err != nil {
		t.Fatalf("UpdateGroup keeping existing non-friend = %v, want nil", err)
	}
}

func TestUpdateGroupRejectsAddingNewNonFriendMember(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{
		id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️",
		members: []string{}, // no existing members
	}
	// newMemberID is NOT a friend of groupOwnerID
	svc := NewService(repo)

	_, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID:   testGroupID,
		Name:      "Trip",
		Icon:      "✈️",
		MemberIDs: []string{newMemberID},
	})
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("UpdateGroup adding new non-friend = %v, want ErrForbidden", err)
	}
}

func TestUpdateGroupIgnoresOwnerIDInMemberIds(t *testing.T) {
	repo := newFakeRepository()
	repo.groups[testGroupID] = fakeGroup{
		id: testGroupID, ownerID: groupOwnerID, name: "Trip", icon: "✈️",
		members: []string{},
	}
	svc := NewService(repo)

	// Including the owner's own ID must be silently dropped, not error or duplicate.
	_, err := svc.UpdateGroup(context.Background(), groupOwnerID, UpdateGroupInput{
		GroupID:   testGroupID,
		Name:      "Trip",
		Icon:      "✈️",
		MemberIDs: []string{groupOwnerID},
	})
	if err != nil {
		t.Fatalf("UpdateGroup with owner ID in memberIds = %v, want nil", err)
	}
}

// --- Activity feed service tests ---

const activityTestUserID = "10000000-0000-0000-0000-000000000001"
const activityTestFriendID = "20000000-0000-0000-0000-000000000002"

func activitySnapshot(now time.Time, count int) domain.LedgerSnapshot {
	currentUser := domain.User{ID: activityTestUserID, PreferredCurrency: "GBP"}
	friend := domain.User{ID: activityTestFriendID}
	group := domain.Group{ID: "g1", Name: "Group", Members: []domain.User{currentUser, friend}}
	expenses := make([]domain.Expense, count)
	for i := range expenses {
		expenses[i] = domain.Expense{
			ID: fmt.Sprintf("expense-%d", i), GroupID: group.ID,
			Description: fmt.Sprintf("Expense %d", i),
			PayerID:     activityTestFriendID, CreatedByID: activityTestFriendID,
			Money:       domain.Money{AmountMinor: 100, Currency: "GBP"},
			SplitMethod: "exact",
			CreatedAt:   now.Add(-time.Duration(i+1) * time.Hour),
			Splits:      []domain.ExpenseSplit{{UserID: activityTestUserID, AmountMinor: 100}},
		}
	}
	return domain.LedgerSnapshot{CurrentUser: currentUser, Groups: []domain.Group{group}, Expenses: expenses}
}

func TestActivityServiceDefaultPaginationWorks(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activitySnapshot(now, 3)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 3 {
		t.Errorf("Total = %d, want 3", feed.Total)
	}
	if len(feed.Activity) != 3 {
		t.Errorf("len(Activity) = %d, want 3", len(feed.Activity))
	}
	if feed.Activity == nil {
		t.Error("Activity is nil, want []")
	}
	if feed.HasMore {
		t.Error("HasMore = true, want false for complete page")
	}
}

func TestActivityServiceLimitAboveMaxIsClamped(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activitySnapshot(now, 5)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 300, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() with limit 300 error = %v; expected clamp to 200, not rejection", err)
	}
	// Clamped to 200; 5 items exist so all 5 are returned.
	if feed.Total != 5 {
		t.Errorf("Total = %d, want 5", feed.Total)
	}
}

func TestActivityServiceRejectsNegativeLimit(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: -1, Offset: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("negative limit = %v, want ErrInvalidInput", err)
	}
}

func TestActivityServiceRejectsZeroLimit(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 0, Offset: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("zero limit = %v, want ErrInvalidInput", err)
	}
}

func TestActivityServiceRejectsNegativeOffset(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 10, Offset: -1})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("negative offset = %v, want ErrInvalidInput", err)
	}
}

func TestActivityServicePagingFlipsHasMore(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activitySnapshot(now, 5)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	page1, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() page1 error = %v", err)
	}
	if len(page1.Activity) != 2 {
		t.Errorf("page1 len = %d, want 2", len(page1.Activity))
	}
	if !page1.HasMore {
		t.Error("page1 HasMore = false, want true")
	}

	// Last page: offset 4, 1 item remaining.
	last, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("Activity() last page error = %v", err)
	}
	if len(last.Activity) != 1 {
		t.Errorf("last page len = %d, want 1", len(last.Activity))
	}
	if last.HasMore {
		t.Error("last page HasMore = true, want false")
	}
	if last.Total != 5 {
		t.Errorf("last page Total = %d, want 5", last.Total)
	}
}

func TestActivityServiceOffsetPastEndReturnsEmptyWithCorrectTotal(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activitySnapshot(now, 2)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 50, Offset: 100})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 2 {
		t.Errorf("Total = %d, want 2", feed.Total)
	}
	if len(feed.Activity) != 0 {
		t.Errorf("len(Activity) = %d, want 0", len(feed.Activity))
	}
	if feed.Activity == nil {
		t.Error("Activity is nil, want []")
	}
	if feed.HasMore {
		t.Error("HasMore = true, want false")
	}
}

// --- Activity kind-filter tests ---

// activityMixedSnapshot builds a snapshot with numExpenses expenses and numSettlements settlements,
// all newest-first relative to the provided base time.
func activityMixedSnapshot(now time.Time, numExpenses, numSettlements int) domain.LedgerSnapshot {
	currentUser := domain.User{ID: activityTestUserID, PreferredCurrency: "GBP"}
	friend := domain.User{ID: activityTestFriendID}
	group := domain.Group{ID: "g1", Name: "Group", Members: []domain.User{currentUser, friend}}

	expenses := make([]domain.Expense, numExpenses)
	for i := range expenses {
		expenses[i] = domain.Expense{
			ID: fmt.Sprintf("expense-%d", i), GroupID: group.ID,
			Description: fmt.Sprintf("Expense %d", i),
			PayerID:     activityTestFriendID, CreatedByID: activityTestFriendID,
			Money:       domain.Money{AmountMinor: 100, Currency: "GBP"},
			SplitMethod: "exact",
			CreatedAt:   now.Add(-time.Duration(i+1) * time.Hour),
			Splits:      []domain.ExpenseSplit{{UserID: activityTestUserID, AmountMinor: 100}},
		}
	}

	// Settlements are offset so they interleave with expenses in time order.
	settlements := make([]domain.Settlement, numSettlements)
	for i := range settlements {
		settlements[i] = domain.Settlement{
			ID:         fmt.Sprintf("settlement-%d", i),
			GroupID:    &group.ID,
			FromUserID: activityTestUserID,
			ToUserID:   activityTestFriendID,
			Money:      domain.Money{AmountMinor: 50, Currency: "GBP"},
			CreatedAt:  now.Add(-time.Duration(i)*time.Hour - 30*time.Minute),
		}
	}

	return domain.LedgerSnapshot{
		CurrentUser: currentUser,
		Groups:      []domain.Group{group},
		Expenses:    expenses,
		Settlements: settlements,
	}
}

func TestActivityKindFilterReturnsOnlyExpenses(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMixedSnapshot(now, 3, 2)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Kind: "expense", Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 3 {
		t.Errorf("Total = %d, want 3 (only expenses)", feed.Total)
	}
	if len(feed.Activity) != 3 {
		t.Errorf("len(Activity) = %d, want 3", len(feed.Activity))
	}
	for _, a := range feed.Activity {
		if a.Kind != "expense" {
			t.Errorf("got kind %q, want expense", a.Kind)
		}
	}
}

func TestActivityKindFilterReturnsOnlySettlements(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMixedSnapshot(now, 3, 2)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Kind: "settlement", Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 2 {
		t.Errorf("Total = %d, want 2 (only settlements)", feed.Total)
	}
	if len(feed.Activity) != 2 {
		t.Errorf("len(Activity) = %d, want 2", len(feed.Activity))
	}
	for _, a := range feed.Activity {
		if a.Kind != "settlement" {
			t.Errorf("got kind %q, want settlement", a.Kind)
		}
	}
}

func TestActivityUnknownKindReturnsInvalidInput(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Kind: "banana", Limit: 10, Offset: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("unknown kind = %v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), "kind must be expense or settlement") {
		t.Errorf("error message = %q, want kind validation message", err.Error())
	}
}

func TestActivityEmptyKindReturnsAll(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMixedSnapshot(now, 3, 2)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 5 {
		t.Errorf("Total = %d, want 5 (3 expenses + 2 settlements)", feed.Total)
	}
	if len(feed.Activity) != 5 {
		t.Errorf("len(Activity) = %d, want 5", len(feed.Activity))
	}
}

func TestActivityKindFilterTotalAndHasMoreReflectFilteredSet(t *testing.T) {
	// Dataset: 3 expenses, 2 settlements. Filtering by settlement with limit=1
	// must report total=2 and paginate through exactly 2 items with no gaps or duplicates.
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMixedSnapshot(now, 3, 2)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	page1, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Kind: "settlement", Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("page1 error = %v", err)
	}
	if page1.Total != 2 {
		t.Errorf("page1 Total = %d, want 2 (filtered set)", page1.Total)
	}
	if len(page1.Activity) != 1 {
		t.Fatalf("page1 len = %d, want 1", len(page1.Activity))
	}
	if !page1.HasMore {
		t.Error("page1 HasMore = false, want true")
	}
	if page1.Activity[0].Kind != "settlement" {
		t.Errorf("page1 item kind = %q, want settlement", page1.Activity[0].Kind)
	}

	page2, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Kind: "settlement", Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("page2 error = %v", err)
	}
	if page2.Total != 2 {
		t.Errorf("page2 Total = %d, want 2 (filtered set)", page2.Total)
	}
	if len(page2.Activity) != 1 {
		t.Fatalf("page2 len = %d, want 1", len(page2.Activity))
	}
	if page2.HasMore {
		t.Error("page2 HasMore = true, want false")
	}
	if page2.Activity[0].Kind != "settlement" {
		t.Errorf("page2 item kind = %q, want settlement", page2.Activity[0].Kind)
	}

	// No duplicates, exactly 2 unique IDs.
	if page1.Activity[0].ID == page2.Activity[0].ID {
		t.Errorf("duplicate ID %q across pages", page1.Activity[0].ID)
	}
}

func TestActivityKindFilterMatchingNothingReturnsEmptyNonNil(t *testing.T) {
	// Snapshot with only expenses; filtering by settlement must return total=0, hasMore=false, activity=[].
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activitySnapshot(now, 3) // expenses only
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{Kind: "settlement", Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 0 {
		t.Errorf("Total = %d, want 0", feed.Total)
	}
	if feed.HasMore {
		t.Error("HasMore = true, want false")
	}
	if feed.Activity == nil {
		t.Error("Activity is nil, want non-nil empty slice")
	}
	if len(feed.Activity) != 0 {
		t.Errorf("len(Activity) = %d, want 0", len(feed.Activity))
	}
}

// --- Activity groupId-filter tests ---

const (
	activityGroupAID = "aaaaaaaa-0000-0000-0000-000000000001"
	activityGroupBID = "bbbbbbbb-0000-0000-0000-000000000002"
)

// activityMultiGroupSnapshot builds a snapshot with 2 expenses in group A, 1 expense in group B,
// and 1 settlement scoped to group A. Used for groupId-filter tests.
func activityMultiGroupSnapshot(now time.Time) domain.LedgerSnapshot {
	currentUser := domain.User{ID: activityTestUserID, PreferredCurrency: "GBP"}
	friend := domain.User{ID: activityTestFriendID}
	groupA := domain.Group{ID: activityGroupAID, Name: "Group A", Members: []domain.User{currentUser, friend}}
	groupB := domain.Group{ID: activityGroupBID, Name: "Group B", Members: []domain.User{currentUser, friend}}

	expenses := []domain.Expense{
		{
			ID: "expense-a1", GroupID: groupA.ID, Description: "Group A Expense 1",
			PayerID: activityTestFriendID, CreatedByID: activityTestFriendID,
			Money: domain.Money{AmountMinor: 100, Currency: "GBP"}, SplitMethod: "exact",
			CreatedAt: now.Add(-1 * time.Hour),
			Splits:    []domain.ExpenseSplit{{UserID: activityTestUserID, AmountMinor: 100}},
		},
		{
			ID: "expense-a2", GroupID: groupA.ID, Description: "Group A Expense 2",
			PayerID: activityTestFriendID, CreatedByID: activityTestFriendID,
			Money: domain.Money{AmountMinor: 200, Currency: "GBP"}, SplitMethod: "exact",
			CreatedAt: now.Add(-2 * time.Hour),
			Splits:    []domain.ExpenseSplit{{UserID: activityTestUserID, AmountMinor: 200}},
		},
		{
			ID: "expense-b1", GroupID: groupB.ID, Description: "Group B Expense",
			PayerID: activityTestFriendID, CreatedByID: activityTestFriendID,
			Money: domain.Money{AmountMinor: 300, Currency: "GBP"}, SplitMethod: "exact",
			CreatedAt: now.Add(-3 * time.Hour),
			Splits:    []domain.ExpenseSplit{{UserID: activityTestUserID, AmountMinor: 300}},
		},
	}

	groupAID := groupA.ID
	settlements := []domain.Settlement{{
		ID: "settlement-a1", GroupID: &groupAID,
		FromUserID: activityTestUserID, ToUserID: activityTestFriendID,
		Money:     domain.Money{AmountMinor: 50, Currency: "GBP"},
		CreatedAt: now.Add(-4 * time.Hour),
	}}

	return domain.LedgerSnapshot{
		CurrentUser: currentUser,
		Groups:      []domain.Group{groupA, groupB},
		Expenses:    expenses,
		Settlements: settlements,
	}
}

func TestActivityGroupIDFilterReturnsOnlyGroupItems(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMultiGroupSnapshot(now)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{GroupID: activityGroupAID, Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	// Group A has 2 expenses + 1 settlement = 3 items.
	if feed.Total != 3 {
		t.Errorf("Total = %d, want 3 (2 expenses + 1 settlement in group A)", feed.Total)
	}
	if len(feed.Activity) != 3 {
		t.Errorf("len(Activity) = %d, want 3", len(feed.Activity))
	}
	for _, a := range feed.Activity {
		if a.GroupID != activityGroupAID {
			t.Errorf("activity %q has GroupID %q, want %q", a.ID, a.GroupID, activityGroupAID)
		}
	}
}

func TestActivityGroupIDFilterCombinedWithKind(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMultiGroupSnapshot(now)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{
		GroupID: activityGroupAID, Kind: "expense", Limit: 50, Offset: 0,
	})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	// Group A with kind=expense yields only the 2 expenses.
	if feed.Total != 2 {
		t.Errorf("Total = %d, want 2 (group A expenses only)", feed.Total)
	}
	for _, a := range feed.Activity {
		if a.Kind != "expense" {
			t.Errorf("activity %q has kind %q, want expense", a.ID, a.Kind)
		}
		if a.GroupID != activityGroupAID {
			t.Errorf("activity %q has GroupID %q, want %q", a.ID, a.GroupID, activityGroupAID)
		}
	}
}

func TestActivityGroupIDMalformedReturnsInvalidInput(t *testing.T) {
	svc := NewService(newFakeRepository())
	_, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{GroupID: "not-a-uuid", Limit: 10, Offset: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("malformed groupId = %v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), "group ID") {
		t.Errorf("error message = %q, want mention of group ID", err.Error())
	}
}

func TestActivityGroupIDUnknownUUIDReturnsEmptyNonNil(t *testing.T) {
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMultiGroupSnapshot(now)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{
		GroupID: "99999999-9999-9999-9999-999999999999", Limit: 50, Offset: 0,
	})
	if err != nil {
		t.Fatalf("Activity() error = %v", err)
	}
	if feed.Total != 0 {
		t.Errorf("Total = %d, want 0", feed.Total)
	}
	if feed.HasMore {
		t.Error("HasMore = true, want false")
	}
	if feed.Activity == nil {
		t.Error("Activity is nil, want non-nil empty slice")
	}
	if len(feed.Activity) != 0 {
		t.Errorf("len(Activity) = %d, want 0", len(feed.Activity))
	}
}

func TestActivityGroupIDFilterTotalAndHasMoreReflectFilteredSet(t *testing.T) {
	// Group A has 3 items (2 expenses + 1 settlement). Page through with limit=1 and
	// assert: total is always 3, hasMore is true on pages 0-1 and false on page 2,
	// no duplicate IDs, no gaps.
	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	repo.snapshot = activityMultiGroupSnapshot(now)
	svc := &Service{repository: repo, now: func() time.Time { return now }}

	seenIDs := make(map[string]bool)
	for page := 0; page < 3; page++ {
		feed, err := svc.Activity(context.Background(), activityTestUserID, ActivityQuery{
			GroupID: activityGroupAID, Limit: 1, Offset: page,
		})
		if err != nil {
			t.Fatalf("page %d error = %v", page, err)
		}
		if feed.Total != 3 {
			t.Errorf("page %d Total = %d, want 3 (filtered set)", page, feed.Total)
		}
		if len(feed.Activity) != 1 {
			t.Fatalf("page %d len(Activity) = %d, want 1", page, len(feed.Activity))
		}
		id := feed.Activity[0].ID
		if seenIDs[id] {
			t.Errorf("duplicate ID %q on page %d", id, page)
		}
		seenIDs[id] = true
		wantHasMore := page < 2
		if feed.HasMore != wantHasMore {
			t.Errorf("page %d HasMore = %v, want %v", page, feed.HasMore, wantHasMore)
		}
	}
	if len(seenIDs) != 3 {
		t.Errorf("got %d unique IDs across pages, want 3", len(seenIDs))
	}
}
