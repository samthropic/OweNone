import { DIFFERENTIATORS } from "./landingData";

export function LandingWhySection() {
  return (
    <section className="block landing-shell">
      <div className="split-highlight">
        <div>
          <div className="section-kicker">Why it feels different</div>
          <h2 className="section-title">
            Sophisticated math.
            <br />
            Frictionless behavior.
          </h2>
          <p className="card-text why-copy">
            The best fintech products hide complexity instead of advertising it.
            OweNone absorbs the graph theory, routing logic, and
            recomputation overhead so users experience only calm, clarity, and
            fewer awkward asks.
          </p>
        </div>
        <div className="check-list">
          {DIFFERENTIATORS.map((item) => (
            <div key={item} className="check-item">
              <div className="check-badge">✓</div>
              <div>{item}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
