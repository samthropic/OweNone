"use client";

import { useState } from "react";
import { PaymentAppIcon } from "@/components/settle/PaymentAppIcon";
import {
  hasPaymentApps,
  paymentAppLinks,
  type PaymentAppLink,
} from "@/lib/payment-apps";
import type { PaymentApps } from "@/lib/api-types";

type Props = {
  apps?: PaymentApps;
  amountMinor?: number;
  /** Compact chips for settle rows; default expands labels. */
  compact?: boolean;
  /**
   * When set, choosing an app opens/copies the handle then calls back so the
   * settle form can prompt “Did you send…?” — apps do not auto-record payments.
   */
  onChoose?: (link: PaymentAppLink) => void;
  selectedId?: PaymentAppLink["id"] | null;
};

export function PaymentAppsLinks({
  apps,
  amountMinor,
  compact = false,
  onChoose,
  selectedId = null,
}: Props) {
  const links = paymentAppLinks(apps, amountMinor);
  const [copied, setCopied] = useState<string | null>(null);

  if (!hasPaymentApps(apps)) {
    if (compact || onChoose) return null;
    return (
      <p className="pay-apps pay-apps-empty">
        No payment apps linked yet.
      </p>
    );
  }

  async function copyHandle(id: string, value: string) {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(id);
      window.setTimeout(() => setCopied(null), 1600);
    } catch {
      // Clipboard can fail in insecure contexts; ignore.
    }
  }

  function choose(link: PaymentAppLink) {
    if (link.href) {
      window.open(link.href, "_blank", "noopener,noreferrer");
    } else {
      void copyHandle(link.id, link.handle);
    }
    onChoose?.(link);
  }

  return (
    <ul
      className={`pay-apps${compact ? " is-compact" : ""}`}
      aria-label={onChoose ? "Pay with" : "Payment apps"}
    >
      {links.map((link) => {
        const isSelected = selectedId === link.id;
        if (onChoose) {
          return (
            <li key={link.id} className={`pay-app pay-app-${link.id}`}>
              <button
                type="button"
                className={`pay-app-link${isSelected ? " is-selected" : ""}`}
                onClick={() => choose(link)}
                title={
                  link.href
                    ? `Open ${link.label}, then confirm in OweNone`
                    : `Copy ${link.label} details, then confirm in OweNone`
                }
              >
                <PaymentAppIcon id={link.id} />
                <span className="pay-app-label">{link.label}</span>
                <span className="pay-app-handle">
                  {copied === link.id && !link.href ? "Copied" : link.handle}
                </span>
              </button>
            </li>
          );
        }

        return (
          <li key={link.id} className={`pay-app pay-app-${link.id}`}>
            {link.href ? (
              <a
                href={link.href}
                target="_blank"
                rel="noopener noreferrer"
                className="pay-app-link"
                title={link.hint}
              >
                <PaymentAppIcon id={link.id} />
                <span className="pay-app-label">{link.label}</span>
                <span className="pay-app-handle">{link.handle}</span>
              </a>
            ) : (
              <button
                type="button"
                className="pay-app-link pay-app-copy"
                onClick={() => copyHandle(link.id, link.handle)}
                title={link.hint}
              >
                <PaymentAppIcon id={link.id} />
                <span className="pay-app-label">{link.label}</span>
                <span className="pay-app-handle">
                  {copied === link.id ? "Copied" : link.handle}
                </span>
              </button>
            )}
          </li>
        );
      })}
    </ul>
  );
}
