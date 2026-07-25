package domain

import "time"

type User struct {
	ID                string `json:"id"`
	Email             string `json:"email,omitempty"`
	DisplayName       string `json:"displayName"`
	PreferredCurrency string `json:"preferredCurrency,omitempty"`
}

type Group struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Members   []User `json:"members"`
	CreatedAt time.Time
}

type ExpenseSplit struct {
	UserID      string `json:"userId"`
	AmountMinor int64  `json:"amountMinor"`
}

type Expense struct {
	ID          string
	GroupID     string
	Description string
	Category    string
	PayerID     string
	CreatedByID string
	Money       Money
	SplitMethod string
	Splits      []ExpenseSplit
	CreatedAt   time.Time
}

type Settlement struct {
	ID         string
	GroupID    *string
	FromUserID string
	ToUserID   string
	Money      Money
	CreatedAt  time.Time
}

type LedgerSnapshot struct {
	CurrentUser User
	Friends     []User
	Groups      []Group
	Expenses    []Expense
	Settlements []Settlement
}

type Dashboard struct {
	User         User                 `json:"user"`
	GeneratedAt  time.Time            `json:"generatedAt"`
	Summary      DashboardSummary     `json:"summary"`
	Friends      []FriendBalance      `json:"friends"`
	Groups       []GroupBalance       `json:"groups"`
	Activity     []Activity           `json:"activity"`
	NetPositions []NamedPosition      `json:"netPositions"`
	Suggestion   SettlementSuggestion `json:"suggestion"`
}

type DashboardSummary struct {
	NetBalance      Money          `json:"netBalance"`
	YouOweTotal     Money          `json:"youOweTotal"`
	YouAreOwedTotal Money          `json:"youAreOwedTotal"`
	ActiveGroups    int            `json:"activeGroups"`
	PaymentsSaved   int            `json:"paymentsSaved"`
	NextAction      *NamedTransfer `json:"nextAction"`
}

type FriendBalance struct {
	User       User     `json:"user"`
	GroupNames []string `json:"groupNames"`
	Balance    Money    `json:"balance"`
}

type GroupBalance struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Icon    string `json:"icon"`
	Members []User `json:"members"`
	Balance Money  `json:"balance"`
}

type Activity struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Description string    `json:"description"`
	Category    string    `json:"category,omitempty"`
	GroupName   string    `json:"groupName,omitempty"`
	Actor       User      `json:"actor"`
	SplitMethod string    `json:"splitMethod,omitempty"`
	PeopleCount int       `json:"peopleCount,omitempty"`
	Impact      Money     `json:"impact"`
	OccurredAt  time.Time `json:"occurredAt"`
}

type NamedPosition struct {
	User    User  `json:"user"`
	Balance Money `json:"balance"`
}

type NamedTransfer struct {
	From  User  `json:"from"`
	To    User  `json:"to"`
	Money Money `json:"money"`
}

type SettlementSuggestion struct {
	OriginalPaymentCount int              `json:"originalPaymentCount"`
	ReducedPaymentCount  int              `json:"reducedPaymentCount"`
	OffsettingDebt       Money            `json:"offsettingDebt"`
	Groups               []GroupReference `json:"groups"`
	Transfers            []NamedTransfer  `json:"transfers"`
}

type GroupReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
