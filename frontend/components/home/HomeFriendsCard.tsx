import Link from "next/link";
import { balanceState, formatMoney, shortName } from "@/components/home/homeData";
import { AddFriendButton } from "@/components/home/HomeActionForms";
import { UserAvatar } from "@/components/home/UserAvatar";
import { PaymentAppsLinks } from "@/components/settle/PaymentAppsLinks";
import type { Dashboard } from "@/lib/api-types";

export function HomeFriendsCard({ dashboard }: { dashboard: Dashboard }) {
  return (
    <section className="home-card">
      <div className="home-card-head">
        <div>
          <h3>Friends</h3>
          <p>Balances across all shared groups</p>
        </div>
        <AddFriendButton />
      </div>
      <div className="home-list">
        {dashboard.friends.map((friend) => {
          const state = balanceState(friend.balance.amountMinor);
          const owesThem = friend.balance.amountMinor < 0;
          const theyOwe = friend.balance.amountMinor > 0;
          return (
            <article key={friend.user.id} className="home-row">
              <UserAvatar user={friend.user} tone={state.tone} />
              <div className="home-row-copy">
                <p className="home-row-title">{shortName(friend.user)}</p>
                <p className="home-row-sub">
                  {friend.groupNames.join(", ") || "No shared groups"}
                </p>
                {owesThem ? (
                  <PaymentAppsLinks
                    apps={friend.user.paymentApps}
                    amountMinor={Math.abs(friend.balance.amountMinor)}
                    compact
                  />
                ) : null}
              </div>
              <div className="home-row-right">
                <p className="home-row-amount">{formatMoney(friend.balance)}</p>
                <p className="home-row-sub">{state.direction}</p>
                {owesThem ? (
                  <Link href="/settle-up" className="home-pill-btn">
                    Settle
                  </Link>
                ) : theyOwe ? (
                  <Link href="/settle-up" className="home-pill-btn">
                    Remind
                  </Link>
                ) : null}
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
