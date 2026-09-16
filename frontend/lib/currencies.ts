// OweNone is USD-only. Amounts are stored as integer cents (minor units).
export type SupportedCurrency = {
  code: string;
  label: string;
  locale: string;
};

export const DEFAULT_CURRENCY = "USD";

export const SUPPORTED_CURRENCIES: SupportedCurrency[] = [
  { code: "USD", label: "US dollar", locale: "en-US" },
];

export function currencyLocale(_code?: string) {
  return "en-US";
}

export function isSupportedCurrency(code: string) {
  return code.trim().toUpperCase() === DEFAULT_CURRENCY;
}
