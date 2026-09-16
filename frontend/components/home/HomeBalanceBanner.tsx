import Link from "next/link";
import { formatMoney, shortName } from "@/components/home/homeData";
import type { Dashboard } from "@/lib/api-types";

export function HomeBalanceBanner({ dashboard }: { dashboard: Dashboard }) {
  const { summary, suggestion, user } = dashboard;
  const nextPayment = suggestion.transfers.find((t) => t.from.id === user.id);
  const waitingCount = suggestion.transfers.filter(
    (t) => t.to.id === user.id,
  ).length;
  const net = summary.netBalance.amountMinor;
  const owesNet = net < 0;
  const isSquare = net === 0;
  const youOwe = summary.youOweTotal.amountMinor > 0;

  let settleTitle = "You’re all settled";
  let settleBody = "No payments needed right now.";
  let settleCta = "View settle up";

  if (nextPayment) {
    settleTitle = `Pay ${shortName(nextPayment.to)}`;
    settleBody = `${formatMoney(nextPayment.money, false)} is the next transfer that clears your side.`;
    settleCta = "Go to Settle up";
  } else if (youOwe) {
    settleTitle = "You still have balances to clear";
    settleBody = `You owe ${formatMoney(summary.youOweTotal, false)} across friends. Settle up to record payments.`;
    settleCta = "Go to Settle up";
  } else if (waitingCount > 0) {
    settleTitle = "Waiting on others";
    settleBody = `${waitingCount} ${waitingCount === 1 ? "person still owes" : "people still owe"} you. Remind them from Settle up.`;
    settleCta = "Go to Settle up";
  }

  return (
    <section className="home-balance-banner">
      <div>
        <p className="home-kicker">Your net balance</p>
        <p
          className={`home-net-amount ${
            isSquare ? "" : owesNet ? "is-negative" : "is-positive"
          }`}
        >
          {formatMoney(summary.netBalance)}
        </p>
        <p className="home-muted">
          {isSquare
            ? "You’re even across your network."
            : owesNet
              ? "You owe"
              : "You are owed"}{" "}
          {!isSquare && (
            <>
              a net of{" "}
              <strong>
                {formatMoney(
                  {
                    ...summary.netBalance,
                    amountMinor: Math.abs(summary.netBalance.amountMinor),
                  },
                  false,
                )}
              </strong>{" "}
            </>
          )}
          across {summary.activeGroups} active{" "}
          {summary.activeGroups === 1 ? "group" : "groups"}.
        </p>
        <div className="home-chip-row">
          {summary.youOweTotal.amountMinor > 0 ? (
            <span className="home-chip chip-rose">
              You owe {formatMoney(summary.youOweTotal, false)}
            </span>
          ) : (
            <span className="home-chip chip-neutral">Nothing to pay</span>
          )}
          {summary.youAreOwedTotal.amountMinor > 0 ? (
            <span className="home-chip chip-green">
              Owed {formatMoney(summary.youAreOwedTotal, false)}
            </span>
          ) : (
            <span className="home-chip chip-neutral">Nothing owed to you</span>
          )}
        </div>
      </div>

      <div className="home-optimal-card">
        <p className="home-kicker">Settle up</p>
        <p className="home-optimal-title">{settleTitle}</p>
        <p className="home-muted">{settleBody}</p>
        <Link href="/settle-up" className="home-btn home-btn-dark">
          {settleCta}
        </Link>
      </div>
    </section>
  );
}
