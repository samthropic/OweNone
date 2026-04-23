import type {
  FeatureCard,
  FooterLink,
  NavLink,
  SecurityItem,
  TrustAvatar,
} from "./types";

export const NAV_LINKS: NavLink[] = [
  { href: "#features", label: "Features" },
  { href: "#how-it-works", label: "How It Works" },
  { href: "#security", label: "Security" },
  { href: "#about", label: "About Us" },
];

export const TRUST_AVATARS: TrustAvatar[] = [
  { initials: "SJ", gradient: "linear-gradient(135deg,#3B82F6,#8B5CF6)" },
  { initials: "MK", gradient: "linear-gradient(135deg,#14B8A6,#10B981)" },
  { initials: "JK", gradient: "linear-gradient(135deg,#F59E0B,#EF4444)" },
  { initials: "AL", gradient: "linear-gradient(135deg,#8B5CF6,#EC4899)" },
  { initials: "RP", gradient: "linear-gradient(135deg,#06B6D4,#3B82F6)" },
];

export const FEATURE_CARDS: FeatureCard[] = [
  {
    id: "compress",
    title: "Compress the graph",
    description:
      "We net out circular debts and surface the minimum set of settlements so your crew moves money once—not in loops.",
    statValue: "−60%",
    statLabel: "avg. payments vs. naive split",
    statTone: "blue",
    iconVariant: "blue",
  },
  {
    id: "live",
    title: "Live, shared truth",
    description:
      "Edits sync instantly across the group. Everyone sees the same balances, suggestions, and payment plan.",
    statValue: "<300ms",
    statLabel: "typical sync latency",
    statTone: "teal",
    iconVariant: "teal",
  },
  {
    id: "explain",
    title: "Explainable math",
    description:
      "Every suggestion is traceable: see who owed whom, what got netted, and why the next step is optimal.",
    statValue: "100%",
    statLabel: "ledger-backed steps",
    statTone: "purple",
    iconVariant: "purple",
  },
];

export const SECURITY_ITEMS: SecurityItem[] = [
  {
    id: "encrypt",
    title: "Encryption in transit & at rest",
    description:
      "TLS everywhere, hardened storage primitives, and least-privilege access for operational tooling.",
  },
  {
    id: "privacy",
    title: "Privacy by design",
    description:
      "Share only what the group needs to settle. Sensitive metadata stays minimized and auditable.",
  },
  {
    id: "sessions",
    title: "Session controls",
    description:
      "Device-aware sessions, easy sign-out everywhere, and optional step-up for high-risk actions.",
  },
];

export const CTA_PERKS: string[] = [
  "Early access pricing for founding groups",
  "Priority support while we polish the beta",
  "Founding member badge on your profile",
];

export const FOOTER_LINKS: FooterLink[] = [
  { href: "#", label: "Privacy" },
  { href: "#", label: "Terms" },
  { href: "#", label: "Contact" },
];
