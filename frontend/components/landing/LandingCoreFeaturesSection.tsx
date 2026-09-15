import { CORE_FEATURES } from "./landingData";

export function LandingCoreFeaturesSection() {
  return (
    <section className="landing-shell block" id="features">
      <div className="features-head">
        <div className="section-kicker">Core product</div>
        <h2 className="section-title">
          Built like infrastructure, surfaced like a luxury consumer product.
        </h2>
      </div>
      <div className="cards-4">
        {CORE_FEATURES.map((feature) => (
          <article key={feature.title} className="card">
            <div className="icon-chip">✦</div>
            <h3 className="card-title">{feature.title}</h3>
            <p className="card-text">{feature.description}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
