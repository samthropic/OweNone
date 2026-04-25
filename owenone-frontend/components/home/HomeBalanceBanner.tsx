export function HomeBalanceBanner() {
  return (
    <section className="home-balance-banner">
      <div>
        <p className="home-kicker">Your net balance</p>
        <p className="home-net-amount">-£12.00</p>
        <p className="home-muted">
          You owe a net of <strong>£12</strong> across 3 active groups. One payment can
          clear everything.
        </p>
        <div className="home-chip-row">
          <span className="home-chip chip-rose">You owe £34 total</span>
          <span className="home-chip chip-green">You are owed £22 total</span>
          <span className="home-chip chip-indigo">3 active groups</span>
        </div>
      </div>

      <div className="home-optimal-card">
        <p className="home-kicker">Optimal next action</p>
        <p className="home-optimal-title">Pay Jordan K.</p>
        <p className="home-muted">£12.00 clears all groups.</p>
        <button type="button" className="home-btn home-btn-dark">
          Pay £12.00 now
        </button>
      </div>
    </section>
  );
}
