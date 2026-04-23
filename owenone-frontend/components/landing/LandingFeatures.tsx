import { FEATURE_CARDS } from "./content";

function FeatureIcon({ variant }: { variant: "blue" | "teal" | "purple" }) {
  const cls = `feat-icon icon-${variant}`;
  if (variant === "blue") {
    return (
      <div className={cls} aria-hidden>
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
          <path
            d="M12 3L4 8v8l8 5 8-5V8l-8-5z"
            stroke="#60A5FA"
            strokeWidth="1.5"
            strokeLinejoin="round"
          />
          <path d="M8 12h8M12 8v8" stroke="#60A5FA" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      </div>
    );
  }
  if (variant === "teal") {
    return (
      <div className={cls} aria-hidden>
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
          <path
            d="M4 12a8 8 0 1116 0"
            stroke="#2DD4BF"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
          <path
            d="M12 8v4l3 2"
            stroke="#2DD4BF"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </div>
    );
  }
  return (
    <div className={cls} aria-hidden>
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
        <path
          d="M7 8h10M7 12h6M7 16h10"
          stroke="#A78BFA"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        <circle cx="17" cy="8" r="2" fill="#A78BFA" opacity="0.35" />
      </svg>
    </div>
  );
}

export function LandingFeatures() {
  return (
    <section className="features" id="features" aria-labelledby="features-heading">
      <div className="container">
        <p className="features-eyebrow">How it works</p>
        <h2 id="features-heading">Three ideas. Infinite clarity.</h2>
        <p className="features-sub">
          OweNone turns messy group IOUs into a single, explainable settlement plan—without losing
          nuance or trust.
        </p>
        <div className="features-grid">
          {FEATURE_CARDS.map((card) => (
            <article key={card.id} className="feat-card">
              <FeatureIcon variant={card.iconVariant} />
              <h3 className="feat-title">{card.title}</h3>
              <p className="feat-desc">{card.description}</p>
              <div className="feat-stat">
                <span className={`feat-stat-num num-${card.statTone}`}>{card.statValue}</span>
                <span className="feat-stat-label">{card.statLabel}</span>
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
