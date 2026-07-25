import { balanceState, formatMoney, shortName } from "@/components/home/homeData";
import { NewGroupButton } from "@/components/home/HomeActionForms";
import type { Dashboard, Money, User } from "@/lib/api-types";

function memberSummary(members: User[], currentUserID: string) {
  const names = members.filter((member) => member.id !== currentUserID).map(shortName);
  if (names.length <= 2) return names.join(", ");
  return `${names.slice(0, 2).join(", ")} + ${names.length - 2}`;
}

function perspectiveMoney(money: Money): Money {
  return { ...money, amountMinor: -money.amountMinor };
}

export function HomeRightColumn({ dashboard }: { dashboard: Dashboard }) {
  const stats = [
    { label: "Total owed", value: formatMoney(dashboard.summary.youOweTotal, false), tone: "is-negative" },
    { label: "You are owed", value: formatMoney(dashboard.summary.youAreOwedTotal, false), tone: "is-positive" },
    { label: "Payments saved", value: String(dashboard.summary.paymentsSaved), tone: "is-indigo" },
    { label: "Net balance", value: formatMoney(dashboard.summary.netBalance), tone: dashboard.summary.netBalance.amountMinor < 0 ? "is-negative" : "is-positive" },
  ];

  return (
    <div className="home-right-column">
      <section className="home-card">
        <div className="home-card-head">
          <div>
            <h3>Your groups</h3>
            <p>{dashboard.summary.activeGroups} active</p>
          </div>
          <NewGroupButton dashboard={dashboard} />
        </div>
        <div className="home-list">
          {dashboard.groups.map((group) => {
            const state = balanceState(group.balance.amountMinor);
            return (
            <article key={group.id} className="home-row">
              <div className="home-icon">{group.icon}</div>
              <div className="home-row-copy">
                <p className="home-row-title">{group.name}</p>
                <p className="home-row-sub">{memberSummary(group.members, dashboard.user.id)}</p>
              </div>
              <div className="home-row-right">
                <p
                  className={`home-row-amount ${group.balance.amountMinor > 0 ? "is-positive" : ""} ${group.balance.amountMinor < 0 ? "is-negative" : ""}`}
                >
                  {formatMoney(group.balance)}
                </p>
                <p className="home-row-sub">{state.direction}</p>
              </div>
            </article>
            );
          })}
        </div>
      </section>

      <section className="home-card">
        <div className="home-card-head">
          <div>
            <h3>Net positions</h3>
            <p>After compression</p>
          </div>
        </div>
        <div className="home-list compact">
          {dashboard.netPositions.map((position) => {
            const money = perspectiveMoney(position.balance);
            return (
            <article key={position.user.id} className="home-compact-row">
              <p className="home-row-title">{shortName(position.user)}</p>
              <p className={`home-row-amount ${money.amountMinor >= 0 ? "is-positive" : "is-negative"}`}>{formatMoney(money)}</p>
            </article>
            );
          })}
        </div>
      </section>

      <section className="home-stat-grid">
        {stats.map((stat) => (
          <article key={stat.label} className="home-stat-card">
            <p>{stat.label}</p>
            <strong className={stat.tone}>{stat.value}</strong>
          </article>
        ))}
      </section>
    </div>
  );
}
