import type { Dashboard, Money, User } from "@/lib/api-types";
import { currencyLocale } from "@/lib/currencies";

export type Tone = "rose" | "green" | "sky" | "amber" | "neutral";

// Cache Intl.NumberFormat instances keyed by "locale|currency" — formatMoney is
// called for every row across the dashboard, activity, groups and settle screens,
// so constructing a new formatter on every render is measurably wasteful.
const formatterCache = new Map<string, Intl.NumberFormat>();

function getFormatter(locale: string, currency: string): Intl.NumberFormat {
  const key = `${locale}|${currency}`;
  let fmt = formatterCache.get(key);
  if (!fmt) {
    fmt = new Intl.NumberFormat(locale, { style: "currency", currency });
    formatterCache.set(key, fmt);
  }
  return fmt;
}

export function formatMoney(money: Money, signed = true) {
  const locale = currencyLocale(money.currency);
  let absolute: string;
  try {
    // Intl.NumberFormat throws RangeError for unknown currency codes. money.currency
    // comes from the API, so guard against malformed values to prevent a full page
    // crash — degrade to a plain number with the raw code appended instead.
    absolute = getFormatter(locale, money.currency).format(Math.abs(money.amountMinor) / 100);
  } catch {
    absolute = `${(Math.abs(money.amountMinor) / 100).toFixed(2)} ${money.currency}`;
  }
  if (!signed || money.amountMinor === 0) return absolute;
  return `${money.amountMinor > 0 ? "+" : "-"}${absolute}`;
}

export function shortName(user: User) {
  const parts = user.displayName.trim().split(/\s+/);
  return parts.length > 1
    ? `${parts[0]} ${parts.at(-1)?.slice(0, 1)}.`
    : parts[0];
}

export function initials(user: Pick<User, "displayName">) {
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
  if (kind === "settlement") return "PAY";
  return ({ food: "FOOD", travel: "TRIP", home: "HOME" } as Record<string, string>)[category] ?? "EXP";
}

export function dashboardDate(dashboard: Dashboard) {
  return new Intl.DateTimeFormat("en-GB", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(dashboard.generatedAt));
}
