import Link from "next/link";
import { categoryIcon, formatMoney, relativeTime, shortName } from "@/components/home/homeData";
import type { Dashboard } from "@/lib/api-types";

export function HomeActivityCard({ dashboard }: { dashboard: Dashboard }) {
  return (
    <section className="home-card">
      <div className="home-card-head">
        <div>
          <h3>Recent activity</h3>
          <p>Last 7 days</p>
        </div>
        <Link href="/activity" className="home-btn home-btn-small">View all</Link>
      </div>
      <div className="home-list">
        {dashboard.activity.map((item) => (
          <article key={item.id} className="home-row">
            <div className="home-icon">{categoryIcon(item.category, item.kind)}</div>
            <div className="home-row-copy">
              <p className="home-row-title">
                {item.actor.id === dashboard.user.id ? "You" : shortName(item.actor)} {item.kind === "expense" ? "added" : "settled"} &quot;{item.description}&quot;
              </p>
              <p className="home-row-sub">
                {item.kind === "expense" ? `${item.splitMethod === "equal" ? "Split equally" : "Exact split"}, ${item.peopleCount} people` : "Marked paid"}
                {item.groupName ? `, ${item.groupName}` : ""}
              </p>
            </div>
            <div className="home-row-right">
              <p
                className={`home-row-amount ${item.impact.amountMinor >= 0 ? "is-positive" : "is-negative"}`}
              >
                {formatMoney(item.impact)}
              </p>
              <p className="home-row-sub">{relativeTime(item.occurredAt, Date.parse(dashboard.generatedAt))}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
