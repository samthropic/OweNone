import { CTA_PERKS, SECURITY_ITEMS } from "./content";

function SecurityGlyph() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M12 3l7 4v5c0 5-3.5 9-7 10-3.5-1-7-5-7-10V7l7-4z"
        stroke="#60A5FA"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
      <path
        d="M9 12l2 2 4-4"
        stroke="#60A5FA"
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function LandingBottomGrid() {
  return (
    <section className="container" id="security" aria-labelledby="security-heading">
      <div className="bottom-section">
        <div className="security-panel" id="about">
          <h2 className="security-title" id="security-heading">
            Security you can reason about
          </h2>
          <p className="security-sub">
            Money stories are sensitive. We treat your graph like production infrastructure—not a
            growth hack.
          </p>
          <div className="security-items">
            {SECURITY_ITEMS.map((item) => (
              <div key={item.id} className="sec-item">
                <div className="sec-icon">
                  <SecurityGlyph />
                </div>
                <div>
                  <p className="sec-text-title">{item.title}</p>
                  <p className="sec-text-desc">{item.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
        <div className="cta-panel">
          <p className="cta-tagline">Get early access</p>
          <h2 className="cta-heading">Ship calmer group trips—starting today.</h2>
          <p className="cta-sub">
            Join the waitlist and we&apos;ll onboard your first circle when your spot opens.
          </p>
          <ul className="cta-perks" aria-label="Included with early access">
            {CTA_PERKS.map((perk) => (
              <li key={perk} className="perk">
                <span className="perk-check" aria-hidden>
                  <svg width="10" height="8" viewBox="0 0 10 8" fill="none">
                    <path
                      d="M1 4l2.5 2.5L9 1"
                      stroke="#34D399"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </span>
                {perk}
              </li>
            ))}
          </ul>
          <button type="button" className="btn-signup">
            Reserve my spot
            <svg width="18" height="18" viewBox="0 0 16 16" fill="none" aria-hidden>
              <path
                d="M8 1.5L14.5 8L8 14.5M1.5 8H14.5"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
              />
            </svg>
          </button>
          <p className="cta-disclaimer">No spam. Unsubscribe any time.</p>
        </div>
      </div>
    </section>
  );
}
