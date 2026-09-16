export type ParsedReceiptItem = {
  name: string;
  amountMinor: number;
};

export type ParsedReceipt = {
  items: ParsedReceiptItem[];
  /** Printed TOTAL when detected; items are reconciled to this amount. */
  receiptTotalMinor: number | null;
  subtotalMinor: number | null;
};

type MoneyLine = {
  kind: "item" | "subtotal" | "tax" | "total";
  name: string;
  amountMinor: number;
  priority: number; // for totals: higher wins
};

const TOTAL_LABEL =
  /^(?<label>grand\s*total|amount\s*due|balance\s*due|total\s*due|total)\b(?!\s*items)/i;

const SUBTOTAL_LABEL = /^(?<label>sub\s*total|subtotal)\b/i;

const TAX_LABEL = /^(?<label>sales\s*tax|tax|vat|gst|hst)\b/i;

const FOOTER_STOP =
  /\b(sub\s*total|subtotal|grand\s*total|amount\s*due|balance\s*due|total\s*due|\btotal\b|tax|vat|gst|hst|change|cash|credit|debit|visa|mastercard|amex|payment|tender|auth|approval)\b/i;

const SKIP_ITEM =
  /\b(change|cash|credit|debit|visa|mastercard|amex|tel\.?|phone|store\s*#|receipt|invoice|auth|approval|card\s*#|xxxxxx|www\.|http|savings|you\s*saved|member|cashier|lane|txn|trans(?:action)?|payment|tender|coupon|promo)\b/i;

const PRICE_AT_END =
  /^(?<name>.+?)\s+(?:\$|USD\s*)?(?<dollars>\d{1,5})\.(?<cents>\d{2})\s*[A-Za-z%*]?\s*$/;

const PRICE_AT_START =
  /^(?:\$|USD\s*)?(?<dollars>\d{1,5})\.(?<cents>\d{2})\s+(?<name>.+?)\s*$/;

/** Parse grocery-style receipt text into line items reconciled to the printed TOTAL. */
export function parseReceiptText(text: string): ParsedReceipt {
  const classified: MoneyLine[] = [];

  for (const rawLine of text.split(/\r?\n/)) {
    const line = rawLine.replace(/\s+/g, " ").trim();
    if (line.length < 3) continue;

    const priceMatch = line.match(PRICE_AT_END) ?? line.match(PRICE_AT_START);
    if (!priceMatch?.groups) continue;

    const amountMinor = moneyFromGroups(priceMatch.groups);
    if (amountMinor <= 0 || amountMinor > 1_000_000_00) continue;

    const labelSource =
      priceMatch.groups.name?.trim() ||
      line.replace(/(?:\$|USD\s*)?\d{1,5}\.\d{2}\s*[A-Za-z%*]?\s*$/i, "").trim();

    if (TOTAL_LABEL.test(labelSource) || TOTAL_LABEL.test(line)) {
      const priority = totalPriority(labelSource) || totalPriority(line);
      classified.push({
        kind: "total",
        name: cleanItemName(labelSource) || "Total",
        amountMinor,
        priority,
      });
      continue;
    }

    if (SUBTOTAL_LABEL.test(labelSource) || SUBTOTAL_LABEL.test(line)) {
      classified.push({
        kind: "subtotal",
        name: "Subtotal",
        amountMinor,
        priority: 0,
      });
      continue;
    }

    if (TAX_LABEL.test(labelSource) || TAX_LABEL.test(line)) {
      classified.push({
        kind: "tax",
        name: "Tax",
        amountMinor,
        priority: 0,
      });
      continue;
    }

    if (SKIP_ITEM.test(line) || FOOTER_STOP.test(labelSource)) continue;

    const name = cleanItemName(labelSource);
    if (!name || name.length < 2) continue;

    classified.push({ kind: "item", name, amountMinor, priority: 0 });
  }

  const subtotalMinor =
    [...classified].reverse().find((line) => line.kind === "subtotal")?.amountMinor ?? null;
  const taxMinor = classified
    .filter((line) => line.kind === "tax")
    .reduce((sum, line) => sum + line.amountMinor, 0);

  const totals = classified.filter((line) => line.kind === "total");
  let receiptTotalMinor = pickReceiptTotal(totals, subtotalMinor, taxMinor);

  // Items only: drop anything that merely duplicates a known total/subtotal amount
  // with a vague name, and ignore trailing noise after the first footer marker in source order.
  let items = classified
    .filter((line) => line.kind === "item")
    .filter((line) => {
      if (receiptTotalMinor != null && line.amountMinor === receiptTotalMinor) return false;
      if (subtotalMinor != null && line.amountMinor === subtotalMinor) return false;
      return true;
    });

  // Keep a prefix of items whose running sum stays near the receipt total.
  if (receiptTotalMinor != null) {
    items = trimItemsToTotal(items, receiptTotalMinor);
  }

  let itemsSum = items.reduce((sum, item) => sum + item.amountMinor, 0);

  if (receiptTotalMinor == null && subtotalMinor != null) {
    receiptTotalMinor = subtotalMinor + taxMinor;
  }

  if (receiptTotalMinor != null && receiptTotalMinor > itemsSum) {
    const gap = receiptTotalMinor - itemsSum;
    if (gap > 0 && gap <= Math.max(receiptTotalMinor, 1)) {
      items = [...items, { name: "Tax & fees", amountMinor: gap, kind: "item" as const, priority: 0 }];
      itemsSum = receiptTotalMinor;
    }
  }

  // Authoritative total is the printed one when we have it.
  if (receiptTotalMinor == null) {
    receiptTotalMinor = itemsSum > 0 ? itemsSum : null;
  }

  return {
    items: items.map(({ name, amountMinor }) => ({ name, amountMinor })),
    receiptTotalMinor,
    subtotalMinor,
  };
}

function totalPriority(label: string) {
  if (/grand\s*total/i.test(label)) return 40;
  if (/amount\s*due/i.test(label)) return 35;
  if (/balance\s*due/i.test(label)) return 30;
  if (/total\s*due/i.test(label)) return 25;
  if (/^total\b/i.test(label.trim()) || /\btotal\b/i.test(label)) return 10;
  return 0;
}

function pickReceiptTotal(
  totals: MoneyLine[],
  subtotalMinor: number | null,
  taxMinor: number,
): number | null {
  if (!totals.length) return null;

  const expected = subtotalMinor != null ? subtotalMinor + taxMinor : null;
  if (expected != null && expected > 0) {
    const exact = totals.find((line) => line.amountMinor === expected);
    if (exact) return exact.amountMinor;
    const near = totals
      .filter((line) => Math.abs(line.amountMinor - expected) <= 5)
      .sort(
        (left, right) =>
          Math.abs(left.amountMinor - expected) - Math.abs(right.amountMinor - expected),
      )[0];
    if (near) return near.amountMinor;
  }

  const usable = totals.filter(
    (line) => !/\b(points?|rewards?|savings?|bonus|available|credit\s*limit)\b/i.test(line.name),
  );
  const pool = usable.length ? usable : totals;

  const bestPriority = Math.max(...pool.map((line) => line.priority));
  const top = pool.filter((line) => line.priority === bestPriority);

  // When several TOTAL lines exist (common OCR noise), prefer the last
  // high-priority amount that isn't wildly larger than the subtotal.
  if (subtotalMinor != null && subtotalMinor > 0) {
    const plausible = top.filter(
      (line) => line.amountMinor >= subtotalMinor && line.amountMinor <= subtotalMinor * 1.4 + 500,
    );
    if (plausible.length) return plausible[plausible.length - 1].amountMinor;
  }

  return top[top.length - 1]?.amountMinor ?? null;
}

function trimItemsToTotal(items: MoneyLine[], totalMinor: number): MoneyLine[] {
  const kept: MoneyLine[] = [];
  let sum = 0;

  for (const item of items) {
    if (sum + item.amountMinor > totalMinor) {
      // Skip this line — often a misread payment/total/duplicate.
      continue;
    }
    kept.push(item);
    sum += item.amountMinor;
    if (sum === totalMinor) break;
  }

  return kept;
}

function moneyFromGroups(groups: Record<string, string>) {
  return Number(groups.dollars) * 100 + Number(groups.cents);
}

function cleanItemName(name: string) {
  return name
    .replace(/^[\d*\s.#-]+/, "")
    .replace(/\s+[A-Z]$/u, "")
    .replace(/\s{2,}/g, " ")
    .trim()
    .slice(0, 80);
}

/** Split amountMinor evenly across assigneeCount. Leftover cents go to the first
 *  indexes (at most one extra cent each), so the parts always sum to amountMinor. */
export function splitAmountAmong(amountMinor: number, assigneeCount: number): number[] {
  if (assigneeCount <= 0) return [];
  if (amountMinor < assigneeCount) {
    return Array.from({ length: assigneeCount }, (_, index) => (index < amountMinor ? 1 : 0));
  }
  const base = Math.floor(amountMinor / assigneeCount);
  const remainder = amountMinor % assigneeCount;
  return Array.from({ length: assigneeCount }, (_, index) => base + (index < remainder ? 1 : 0));
}

export function formatMinorAsDecimal(amountMinor: number) {
  return (amountMinor / 100).toFixed(2);
}

export function parseDecimalToMinor(text: string): number | null {
  const trimmed = text.trim();
  if (!/^\d{1,10}(?:\.\d{1,2})?$/.test(trimmed)) return null;
  const [whole, fraction = ""] = trimmed.split(".");
  const amountMinor = Number(whole) * 100 + Number(fraction.padEnd(2, "0"));
  if (!Number.isSafeInteger(amountMinor) || amountMinor <= 0) return null;
  return amountMinor;
}
