import { HOW_STEPS } from "./landingData";

export function LandingHowSection() {
  return (
    <section id="how" className="how-section block">
      <div className="landing-shell">
        <p className="section-kicker">How it works</p>
        <h2 className="section-title">
          Capture the network. Net globally. Settle the minimum.
        </h2>
        <ol className="how-list">
          {HOW_STEPS.map((step) => (
            <li key={step.number}>
              <h3>{step.title}</h3>
              <p>{step.description}</p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
