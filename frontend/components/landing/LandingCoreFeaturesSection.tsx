import { CORE_FEATURES } from "./landingData";

export function LandingCoreFeaturesSection() {
  return (
    <section className="landing-shell block" id="features">
      <div className="features-head">
        <p className="section-kicker">What it does</p>
        <h2 className="section-title">
          One net position per relationship — across every group you share.
        </h2>
      </div>
      <div className="feature-strip">
        {CORE_FEATURES.map((feature) => (
          <article key={feature.title}>
            <h3>{feature.title}</h3>
            <p>{feature.description}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
