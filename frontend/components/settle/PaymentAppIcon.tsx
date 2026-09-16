import type { PaymentAppLink } from "@/lib/payment-apps";

type Props = {
  id: PaymentAppLink["id"];
  className?: string;
};

/** Simplified brand marks for payment-app chips. */
export function PaymentAppIcon({ id, className = "pay-app-icon" }: Props) {
  switch (id) {
    case "venmo":
      return (
        <svg
          className={className}
          viewBox="0 0 24 24"
          aria-hidden="true"
          focusable="false"
        >
          <rect width="24" height="24" rx="6" fill="#008CFF" />
          <path
            d="M17.2 5.2c.45.74.68 1.52.68 2.46 0 3.06-2.62 7.04-4.74 9.84H7.9L6.3 5.9l4.08-.4 1.02 8.2c.94-1.54 2.1-3.96 2.1-5.6 0-.9-.16-1.52-.4-2.02l3.1-.88Z"
            fill="#fff"
          />
        </svg>
      );
    case "paypal":
      return (
        <svg
          className={className}
          viewBox="0 0 24 24"
          aria-hidden="true"
          focusable="false"
        >
          <rect width="24" height="24" rx="6" fill="#003087" />
          <path
            d="M9.1 17.6 9.8 13.4h2.7c2.2 0 3.4-1.1 3.7-2.9.4-2.1-.8-3.3-2.9-3.3H9.6L8.2 17.6h.9Zm1.7-9.2h1.8c1.1 0 1.7.5 1.5 1.5-.2 1-.9 1.5-2 1.5h-1.6l.3-3Z"
            fill="#fff"
          />
          <path
            d="M10.4 18.8 11.1 14.6h2.7c2.2 0 3.4-1.1 3.7-2.9.2-1.1 0-2-.5-2.6.9.3 1.5 1 1.3 2.3-.4 2.1-1.9 3.5-4.1 3.5h-1.8l-.7 4h-.3Z"
            fill="#009CDE"
          />
        </svg>
      );
    case "cashApp":
      return (
        <svg
          className={className}
          viewBox="0 0 24 24"
          aria-hidden="true"
          focusable="false"
        >
          <rect width="24" height="24" rx="6" fill="#00D632" />
          <path
            d="M13.4 6.4c.7.2 1.3.6 1.7 1.1l-1.4 1.2c-.2-.3-.6-.5-1-.6-.7-.2-1.4.1-1.6.7-.2.5.1 1 .8 1.3l.9.3c1.6.6 2.3 1.6 2.1 3-.2 1.5-1.4 2.5-3.1 2.7v1.5h-1.5v-1.5c-.8-.1-1.6-.4-2.2-.9l1.3-1.3c.4.3.9.5 1.5.6.8.2 1.5-.1 1.7-.7.2-.5-.1-1-.9-1.3l-.8-.3c-1.5-.5-2.2-1.6-2-3 .2-1.5 1.4-2.5 3.1-2.7V5h1.5v1.4Z"
            fill="#fff"
          />
        </svg>
      );
    case "zelle":
      return (
        <svg
          className={className}
          viewBox="0 0 24 24"
          aria-hidden="true"
          focusable="false"
        >
          <rect width="24" height="24" rx="6" fill="#6D1ED4" />
          <path
            d="M7.2 7.2h9.1v2.1l-5.6 7.2H16.8v2.3H7.2v-2.1l5.6-7.2H7.2V7.2Z"
            fill="#fff"
          />
        </svg>
      );
    default:
      return null;
  }
}
