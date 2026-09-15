import type { Dashboard, Money, User } from "@/lib/api-types";

export type Tone = "rose" | "green" | "sky" | "amber" | "neutral";

export function formatMoney(money: Money, signed = true) {
  const formatter = new Intl.NumberFormat("en-GB", {
    style: "currency",
    currency: money.currency,
  });
  const absolute = formatter.format(Math.abs(money.amountMinor) / 100);
  if (!signed || money.amountMinor === 0) return absolute;
  return `${money.amountMinor > 0 ? "+" : "-"}${absolute}`;
}

export function shortName(user: User) {
  const parts = user.displayName.trim().split(/\s+/);
  return parts.length > 1
    ? `${parts[0]} ${parts.at(-1)?.slice(0, 1)}.`
    : parts[0];
}

export function initials(user: User) {
  return user.displayName
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}

export function relativeTime(value: string, now = Date.now()) {
  const elapsedSeconds = Math.max(0, Math.floor((now - Date.parse(value)) / 1000));
  if (elapsedSeconds < 60) return "Just now";
  if (elapsedSeconds < 3600) return `${Math.floor(elapsedSeconds / 60)}m ago`;
  if (elapsedSeconds < 86400) return `${Math.floor(elapsedSeconds / 3600)}h ago`;
  const days = Math.floor(elapsedSeconds / 86400);
  return days === 1 ? "Yesterday" : `${days} days ago`;
}

export function balanceState(amountMinor: number) {
  if (amountMinor > 0) return { direction: "Owes you", tone: "green" as const };
  if (amountMinor < 0) return { direction: "You owe", tone: "rose" as const };
  return { direction: "Settled", tone: "neutral" as const };
}

export function categoryIcon(category = "general", kind = "expense") {
  if (kind === "settlement") return "✓";
  return ({ food: "🍕", travel: "✈️", home: "🏠" } as Record<string, string>)[category] ?? "🧾";
}

export function dashboardDate(dashboard: Dashboard) {
  return new Intl.DateTimeFormat("en-GB", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(dashboard.generatedAt));
}
