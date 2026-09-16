import type { PaymentApps } from "@/lib/api-types";

export type PaymentAppLink = {
  id: "venmo" | "paypal" | "cashApp" | "zelle";
  label: string;
  handle: string;
  href: string | null;
  hint: string;
};

function dollars(amountMinor?: number) {
  if (amountMinor == null || !Number.isFinite(amountMinor)) return undefined;
  return (Math.abs(amountMinor) / 100).toFixed(2);
}

/** Build deep links / copy hints for a user's linked payment apps. */
export function paymentAppLinks(
  apps: PaymentApps | undefined,
  amountMinor?: number,
): PaymentAppLink[] {
  if (!apps) return [];
  const amount = dollars(amountMinor);
  const links: PaymentAppLink[] = [];

  if (apps.venmo) {
    const handle = apps.venmo.replace(/^@/, "");
    const params = new URLSearchParams({
      txn: "pay",
      audience: "private",
      recipients: handle,
    });
    if (amount) params.set("amount", amount);
    links.push({
      id: "venmo",
      label: "Venmo",
      handle: `@${handle}`,
      href: `https://venmo.com/${encodeURIComponent(handle)}?${params.toString()}`,
      hint: `Pay @${handle} on Venmo`,
    });
  }

  if (apps.paypal) {
    const handle = apps.paypal.replace(/^@/, "");
    const path = amount
      ? `${encodeURIComponent(handle)}/${amount}`
      : encodeURIComponent(handle);
    links.push({
      id: "paypal",
      label: "PayPal",
      handle,
      href: `https://www.paypal.me/${path}`,
      hint: `Pay ${handle} on PayPal.me`,
    });
  }

  if (apps.cashApp) {
    const tag = apps.cashApp.replace(/^\$/, "");
    const path = amount
      ? `$${encodeURIComponent(tag)}/${amount}`
      : `$${encodeURIComponent(tag)}`;
    links.push({
      id: "cashApp",
      label: "Cash App",
      handle: `$${tag}`,
      href: `https://cash.app/${path}`,
      hint: `Pay $${tag} on Cash App`,
    });
  }

  if (apps.zelle) {
    links.push({
      id: "zelle",
      label: "Zelle",
      handle: apps.zelle,
      href: null,
      hint: `Send via Zelle to ${apps.zelle}`,
    });
  }

  return links;
}

export function hasPaymentApps(apps: PaymentApps | undefined) {
  return paymentAppLinks(apps).length > 0;
}

export function paymentMethodLabel(id: PaymentAppLink["id"] | "other" | "") {
  switch (id) {
    case "venmo":
      return "Venmo";
    case "paypal":
      return "PayPal";
    case "cashApp":
      return "Cash App";
    case "zelle":
      return "Zelle";
    case "other":
      return "another method";
    default:
      return "payment";
  }
}
