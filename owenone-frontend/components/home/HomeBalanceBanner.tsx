import { formatMoney, shortName } from "@/components/home/homeData";
import { SettlementButton } from "@/components/home/HomeActionForms";
import type { Dashboard } from "@/lib/api-types";

export function HomeBalanceBanner({ dashboard }: { dashboard: Dashboard }) {
  const { summary } = dashboard;
  const nextAction = summary.nextAction;
  const owesNet = summary.netBalance.amountMinor < 0;

  return (
    <section className="home-balance-banner">
      <div>
        <p className="home-kicker">Your net balance</p>
        <p className={`home-net-amount ${owesNet ? "is-negative" : "is-positive"}`}>
          {formatMoney(summary.netBalance)}
        </p>
        <p className="home-muted">
          {owesNet ? "You owe" : "You are owed"} a net of{" "}
          <strong>{formatMoney({ ...summary.netBalance, amountMinor: Math.abs(summary.netBalance.amountMinor) }, false)}</strong>{" "}
          across {summary.activeGroups} active {summary.activeGroups === 1 ? "group" : "groups"}.
          {summary.paymentsSaved > 0 ? ` ${summary.paymentsSaved} payments can be skipped.` : ""}
        </p>
        <div className="home-chip-row">
          <span className="home-chip chip-rose">You owe {formatMoney(summary.youOweTotal, false)} total</span>
          <span className="home-chip chip-green">You are owed {formatMoney(summary.youAreOwedTotal, false)} total</span>
          <span className="home-chip chip-indigo">{summary.activeGroups} active groups</span>
        </div>
      </div>

      <div className="home-optimal-card">
        <p className="home-kicker">Optimal next action</p>
        {nextAction ? (
          <>
            <p className="home-optimal-title">Pay {shortName(nextAction.to)}</p>
            <p className="home-muted">{formatMoney(nextAction.money, false)} clears your net position.</p>
            <SettlementButton
              transfer={nextAction}
              label={`Pay ${formatMoney(nextAction.money, false)} now`}
              className="home-btn home-btn-dark"
            />
          </>
        ) : (
          <p className="home-optimal-title">Nothing to pay</p>
        )}
      </div>
    </section>
  );
}
