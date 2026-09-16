import Link from "next/link";
import { balanceState, formatMoney, shortName } from "@/components/home/homeData";
import { NewGroupButton } from "@/components/home/HomeActionForms";
import type { Dashboard, User } from "@/lib/api-types";

function memberSummary(members: User[], currentUserID: string) {
  const names = members.filter((member) => member.id !== currentUserID).map(shortName);
  if (names.length <= 2) return names.join(", ");
  return `${names.slice(0, 2).join(", ")} + ${names.length - 2}`;
}

export function HomeGroupsCard({ dashboard }: { dashboard: Dashboard }) {
  return (
    <section className="home-card">
      <div className="home-card-head">
        <div>
          <h3>Your groups</h3>
          <p>{dashboard.summary.activeGroups} active</p>
        </div>
        <Link href="/groups" className="home-btn home-btn-small">
          View all
        </Link>
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
                <p className="home-row-sub">
                  {memberSummary(group.members, dashboard.user.id)}
                </p>
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
  );
}
