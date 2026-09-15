import { balanceState, formatMoney, initials, shortName } from "@/components/home/homeData";
import { AddFriendButton, FriendActionButton } from "@/components/home/HomeActionForms";
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
          return (
          <article key={friend.user.id} className="home-row">
            <div className={`home-avatar tone-${state.tone}`}>{initials(friend.user)}</div>
            <div className="home-row-copy">
              <p className="home-row-title">{shortName(friend.user)}</p>
              <p className="home-row-sub">{friend.groupNames.join(", ") || "No shared groups"}</p>
            </div>
            <div className="home-row-right">
              <p className="home-row-amount">{formatMoney(friend.balance)}</p>
              <p className="home-row-sub">{state.direction}</p>
              {friend.balance.amountMinor !== 0 ? (
                <FriendActionButton friend={friend} />
              ) : null}
            </div>
          </article>
          );
        })}
      </div>
    </section>
  );
}
