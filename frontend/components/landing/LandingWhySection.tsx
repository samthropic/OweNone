import { DIFFERENTIATORS } from "./landingData";

export function LandingWhySection() {
  return (
    <section className="block landing-shell" id="why">
      <div className="why-panel">
        <div>
          <p className="section-kicker">Why it feels different</p>
          <h2 className="section-title">
            The math stays underwater.
            <br />
            You just pay less often.
          </h2>
          <p className="why-copy">
            OweNone recomputes the whole debt graph when balances change, then
            shows the shortest set of transfers — not every IOU that created
            them.
          </p>
        </div>
        <ul className="why-list">
          {DIFFERENTIATORS.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
    </section>
  );
}
