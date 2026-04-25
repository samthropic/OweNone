import { friends } from "@/components/home/homeData";

export function HomeFriendsCard() {
  return (
    <section className="home-card">
      <div className="home-card-head">
        <div>
          <h3>Friends</h3>
          <p>Balances across all shared groups</p>
        </div>
        <button type="button" className="home-btn home-btn-small">
          Add friend
        </button>
      </div>
      <div className="home-list">
        {friends.map((friend) => (
          <article key={friend.id} className="home-row">
            <div className={`home-avatar tone-${friend.tone}`}>{friend.initials}</div>
            <div className="home-row-copy">
              <p className="home-row-title">{friend.name}</p>
              <p className="home-row-sub">{friend.context}</p>
            </div>
            <div className="home-row-right">
              <p className="home-row-amount">{friend.amount}</p>
              <p className="home-row-sub">{friend.direction}</p>
              {friend.cta ? (
                <button type="button" className="home-pill-btn">
                  {friend.cta}
                </button>
              ) : null}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
