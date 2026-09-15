import { formatMoney } from "@/components/home/homeData";
import { SettlementButton } from "@/components/home/HomeActionForms";
import type { Dashboard } from "@/lib/api-types";

export function HomeSuggestionCard({ dashboard }: { dashboard: Dashboard }) {
  const { suggestion } = dashboard;
  const groupNames = suggestion.groups.map((group) => group.name);
  const nextPayment = suggestion.transfers.find((transfer) => transfer.from.id === dashboard.user.id);

  return (
    <section className="home-suggestion-card">
      <p className="home-kicker home-kicker-dark">Graph compression active</p>
      <h2>OweNone found a shortcut.</h2>
      <p>
        {suggestion.originalPaymentCount} scattered IOUs across {groupNames.join(", ") || "your network"} collapse into {suggestion.reducedPaymentCount} {suggestion.reducedPaymentCount === 1 ? "transfer" : "transfers"}.
      </p>
      <div className="home-suggestion-stats">
        <div>
          <strong>{suggestion.originalPaymentCount} → {suggestion.reducedPaymentCount}</strong>
          <span>payments reduced</span>
        </div>
        <div>
          <strong>{formatMoney(suggestion.offsettingDebt, false)}</strong>
          <span>offsetting debt cancelled</span>
        </div>
        <div>
          <strong>{suggestion.groups.length}</strong>
          <span>groups compressed</span>
        </div>
      </div>
      {nextPayment ? (
        <SettlementButton transfer={nextPayment} label="Accept & Pay" className="home-btn home-btn-primary" />
      ) : null}
    </section>
  );
}
