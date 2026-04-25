export type Tone = "rose" | "green" | "sky" | "amber" | "neutral";

export type Friend = {
  id: string;
  name: string;
  initials: string;
  context: string;
  amount: string;
  direction: string;
  tone: Tone;
  cta?: string;
};

export type ActivityItem = {
  id: string;
  icon: string;
  title: string;
  meta: string;
  amount: string;
  amountTone: "positive" | "negative";
  time: string;
};

export type GroupItem = {
  id: string;
  icon: string;
  name: string;
  members: string;
  balance: string;
  tone: "positive" | "negative" | "neutral";
  status: string;
};

export const friends: Friend[] = [
  {
    id: "jordan",
    name: "Jordan K.",
    initials: "JK",
    context: "Holiday '24, Flatmates, Dinners",
    amount: "-£12.00",
    direction: "You owe",
    tone: "rose",
    cta: "Pay",
  },
  {
    id: "mike",
    name: "Mike K.",
    initials: "MK",
    context: "Flatmates, Groceries",
    amount: "+£18.50",
    direction: "Owes you",
    tone: "green",
    cta: "Remind",
  },
  {
    id: "alex",
    name: "Alex L.",
    initials: "AL",
    context: "Holiday '24, Dinners",
    amount: "+£9.00",
    direction: "Owes you",
    tone: "sky",
    cta: "Remind",
  },
  {
    id: "sam",
    name: "Sam R.",
    initials: "SR",
    context: "Flatmates",
    amount: "-£8.00",
    direction: "You owe",
    tone: "amber",
    cta: "Pay",
  },
  {
    id: "priya",
    name: "Priya R.",
    initials: "PR",
    context: "Dinners, all settled",
    amount: "£0.00",
    direction: "Settled",
    tone: "neutral",
  },
];

export const activity: ActivityItem[] = [
  {
    id: "pizza",
    icon: "🍕",
    title: "Jordan added \"Pizza night - Soho\"",
    meta: "Split equally, 4 people, Holiday '24",
    amount: "-£14.25",
    amountTone: "negative",
    time: "2h ago",
  },
  {
    id: "rent",
    icon: "💸",
    title: "Mike settled \"Monthly rent\"",
    meta: "Marked paid, Flatmates",
    amount: "+£18.50",
    amountTone: "positive",
    time: "Yesterday",
  },
  {
    id: "train",
    icon: "✈",
    title: "You added \"Train tickets - Edinburgh\"",
    meta: "Split equally, 3 people, Holiday '24",
    amount: "-£42.00",
    amountTone: "negative",
    time: "2 days ago",
  },
];

export const groups: GroupItem[] = [
  {
    id: "holiday",
    icon: "✈",
    name: "Holiday '24",
    members: "Jordan, Alex + 2",
    balance: "-£18",
    tone: "negative",
    status: "You owe",
  },
  {
    id: "flatmates",
    icon: "🏠",
    name: "Flatmates",
    members: "Mike, Sam + 1",
    balance: "+£10.50",
    tone: "positive",
    status: "Owed to you",
  },
  {
    id: "dinners",
    icon: "🍽",
    name: "Dinners",
    members: "Alex, Priya, Jordan",
    balance: "£0",
    tone: "neutral",
    status: "Settled",
  },
];
