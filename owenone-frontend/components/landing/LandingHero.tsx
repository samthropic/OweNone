import { TRUST_AVATARS } from "./content";

export function LandingHero() {
  return (
    <section className="hero" id="how-it-works" aria-labelledby="hero-heading">
      <div className="container">
        <div className="hero-eyebrow">
          <span className="eyebrow-dot" aria-hidden />
          <span className="eyebrow-text">Introducing the debt graph compressor</span>
        </div>
        <h1 id="hero-heading">
          Drowning in Group <em>IOUs?</em>
          <br />
          OweNone Untangles the Mess.
        </h1>
        <p className="hero-sub">
          Stop making five payments across three apps. The first{" "}
          <strong>social graph debt compressor</strong> uses smart math to find your single
          optimal path. Instant sync, zero friction.
        </p>
        <div className="hero-actions">
          <button type="button" className="btn-primary">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden>
              <path
                d="M8 1.5L14.5 8L8 14.5M1.5 8H14.5"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
              />
            </svg>
            Start for Free
          </button>
          <button type="button" className="btn-ghost">
            <svg width="15" height="15" viewBox="0 0 15 15" fill="none" aria-hidden>
              <circle cx="7.5" cy="7.5" r="6.5" stroke="currentColor" strokeWidth="1.3" />
              <path d="M6 5.5L9.5 7.5L6 9.5V5.5Z" fill="currentColor" />
            </svg>
            Watch demo
          </button>
        </div>
        <div className="trust-bar">
          <div className="trust-avatars" aria-hidden>
            {TRUST_AVATARS.map((a) => (
              <div key={a.initials} className="av" style={{ background: a.gradient }}>
                {a.initials}
              </div>
            ))}
          </div>
          <span className="trust-text">
            <strong>2,400+ people</strong> on the waitlist
          </span>
        </div>
      </div>
    </section>
  );
}
