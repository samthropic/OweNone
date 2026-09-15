import type {
  LandingConnection,
  LandingFeature,
  LandingFlow,
  LandingHeroStat,
  LandingMiniStat,
  LandingNetworkNode,
  LandingNavItem,
  LandingProblemCard,
  LandingStep,
} from "./landingTypes";

export const LANDING_NAV_ITEMS: LandingNavItem[] = [
  { href: "#problem", label: "Problem" },
  { href: "#features", label: "Features" },
  { href: "#how", label: "How it works" },
  { href: "#waitlist", label: "Waitlist" },
];

export const HERO_STATS: LandingHeroStat[] = [
  { value: "67%", label: "fewer payment actions after graph compression" },
  { value: "Live", label: "network recalculation the moment balances change" },
  { value: "∞ groups", label: "one net position per relationship across contexts" },
];

export const PROBLEM_CARDS: LandingProblemCard[] = [
  {
    title: "Splitwise",
    body: "Useful inside a single group, but structurally blind once the same people owe each other across homes, trips, dinners, and time.",
  },
  {
    title: "Venmo",
    body: "Excellent at executing payments, but it leaves all routing, netting, and decision making to the user.",
  },
  {
    title: "OweNone",
    body: "A payment-graph coordinator layer that sees the whole social debt graph, compresses it in real time, and turns many transfers into the clean minimum.",
    highlighted: true,
  },
];

export const CORE_FEATURES: LandingFeature[] = [
  {
    title: "Global debt compression",
    description:
      "A graph-native reconciliation engine that reduces obligations across your entire network to the fewest mathematically valid transfers.",
  },
  {
    title: "Live recomputation",
    description:
      "Every new expense, settlement, and split triggers instant network updates, so your balances always reflect the latest optimal state.",
  },
  {
    title: "Cross-group netting",
    description:
      "Balances, friends, and recurring flows are collapsed across disconnected silos of living, travel, and social circles.",
  },
  {
    title: "Settlement intelligence",
    description:
      "The system identifies clean settlement paths that minimize total cash movement while preserving fairness.",
  },
];

export const NETWORK_NODES: LandingNetworkNode[] = [
  { id: "alex", name: "Alex", initials: "AL", balance: "+£18", x: 50, y: 18, tone: "green" },
  { id: "yasmin", name: "Yasmin", initials: "YA", balance: "-£12", x: 16, y: 58, tone: "rose" },
  { id: "jordan", name: "Jordan", initials: "JO", balance: "+£30", x: 84, y: 58, tone: "sky" },
  { id: "sam", name: "Sam", initials: "SA", balance: "-£18", x: 50, y: 86, tone: "amber" },
];

export const BEFORE_CONNECTIONS: LandingConnection[] = [
  { from: "yasmin", to: "alex", amount: "£12", color: "rose" },
  { from: "alex", to: "jordan", amount: "£30", color: "rose" },
  { from: "jordan", to: "sam", amount: "£18", color: "rose" },
];

export const AFTER_CONNECTIONS: LandingConnection[] = [
  { from: "yasmin", to: "sam", amount: "£18", color: "green" },
  { from: "alex", to: "jordan", amount: "£30", color: "green" },
];

export const BEFORE_FLOWS: LandingFlow[] = [
  { label: "Yasmin → Alex", amount: "£12" },
  { label: "Alex → Jordan", amount: "£30" },
  { label: "Jordan → Sam", amount: "£18" },
];

export const AFTER_FLOWS: LandingFlow[] = [
  { label: "Yasmin → Sam", amount: "£18" },
  { label: "Alex → Jordan", amount: "£30" },
];

export const MINI_STATS: LandingMiniStat[] = [
  { value: "3 → 2", label: "settlement actions reduced" },
  { value: "Instant", label: "network-wide graph recomputation" },
  { value: "1 layer", label: "for all shared money relationships" },
];

export const HOW_STEPS: LandingStep[] = [
  {
    number: "01",
    title: "Capture the network",
    description:
      "Every dinner, trip, rent split, and shared purchase becomes an edge in a real financial graph rather than an isolated expense row.",
  },
  {
    number: "02",
    title: "Net balances globally",
    description:
      "OweNone computes each person’s true position across groups and removes redundant flows before any payment is made.",
  },
  {
    number: "03",
    title: "Settle optimally",
    description:
      "Users see the minimum required transfers, not the mess that created them — one elegant action instead of many fragmented ones.",
  },
];

export const DIFFERENTIATORS: string[] = [
  "Global minimization instead of group-only simplification",
  "Continuous netting as relationships evolve",
  "Realtime state across every connected expense",
  "Designed to become the settlement layer, not just the ledger",
];
