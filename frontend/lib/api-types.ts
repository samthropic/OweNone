export type Money = {
  amountMinor: number;
  currency: string;
};

export type User = {
  id: string;
  email?: string;
  displayName: string;
  preferredCurrency?: string;
};

export type NamedTransfer = {
  from: User;
  to: User;
  money: Money;
};

export type Dashboard = {
  user: User;
  generatedAt: string;
  summary: {
    netBalance: Money;
    youOweTotal: Money;
    youAreOwedTotal: Money;
    activeGroups: number;
    paymentsSaved: number;
    nextAction: NamedTransfer | null;
  };
  friends: Array<{
    user: User;
    groupNames: string[];
    balance: Money;
  }>;
  groups: Array<{
    id: string;
    name: string;
    icon: string;
    members: User[];
    balance: Money;
  }>;
  activity: Array<{
    id: string;
    kind: "expense" | "settlement";
    description: string;
    category?: string;
    groupName?: string;
    actor: User;
    splitMethod?: string;
    peopleCount?: number;
    impact: Money;
    occurredAt: string;
  }>;
  netPositions: Array<{
    user: User;
    balance: Money;
  }>;
  suggestion: {
    originalPaymentCount: number;
    reducedPaymentCount: number;
    offsettingDebt: Money;
    groups: Array<{ id: string; name: string }>;
    transfers: NamedTransfer[];
  };
};

export type APIResponse<T> = { data: T };

export type APIError = {
  error?: {
    code?: string;
    message?: string;
  };
};