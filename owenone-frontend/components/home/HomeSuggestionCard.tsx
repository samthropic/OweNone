export function HomeSuggestionCard() {
  return (
    <section className="home-suggestion-card">
      <p className="home-kicker home-kicker-dark">Graph compression active</p>
      <h2>OweNone found a shortcut.</h2>
      <p>
        6 scattered IOUs across Holiday, Flatmates, and Dinners collapse into 1
        transfer.
      </p>
      <div className="home-suggestion-stats">
        <div>
          <strong>6 → 1</strong>
          <span>payments reduced</span>
        </div>
        <div>
          <strong>£22</strong>
          <span>offsetting debt cancelled</span>
        </div>
        <div>
          <strong>3</strong>
          <span>groups compressed</span>
        </div>
      </div>
      <button type="button" className="home-btn home-btn-primary">
        Accept & Pay
      </button>
    </section>
  );
}
