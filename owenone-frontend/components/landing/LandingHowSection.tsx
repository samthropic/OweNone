import { HOW_STEPS } from "./landingData";

export function LandingHowSection() {
  return (
    <section id="how" className="dark-section block">
      <div className="landing-shell">
        <div className="section-kicker">How it works</div>
        <h2 className="section-title">
          One graph beneath every split.
          <br />
          One optimal state above it.
        </h2>
        <div className="cards-3">
          {HOW_STEPS.map((step) => (
            <article key={step.number} className="card">
              <div className="dark-number">{step.number}</div>
              <h3 className="card-title">{step.title}</h3>
              <p className="dark-copy">{step.description}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
