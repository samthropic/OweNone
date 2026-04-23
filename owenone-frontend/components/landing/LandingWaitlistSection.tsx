export function LandingWaitlistSection() {
  return (
    <section className="landing-shell cta-section" id="waitlist">
      <div className="cta-box">
        <div className="cta-copy">
          <div className="section-kicker private-beta">Private beta</div>
          <h2 className="section-title">
            The next payment app won&apos;t just move money.
            <br />
            It will think first.
          </h2>
          <p className="card-text waitlist-copy">
            OweNone is for households, friend groups, travellers, and
            operators who want shared finances to feel coordinated instead of
            chaotic. Join the waitlist for first access.
          </p>
          <form className="input-row">
            <input
              className="input"
              type="email"
              placeholder="Enter your email"
              aria-label="Email address"
              required
            />
            <button type="submit" className="btn btn-dark waitlist-btn">
              Join waitlist
            </button>
          </form>
        </div>
      </div>
    </section>
  );
}
