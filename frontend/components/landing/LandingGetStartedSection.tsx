import Link from "next/link";

export function LandingGetStartedSection() {
  return (
    <section className="landing-shell cta-section" id="get-started">
      <div className="cta-panel">
        <p className="section-kicker">Get started</p>
        <h2 className="section-title">Ready to clear the backlog of IOUs?</h2>
        <p className="cta-body">
          Create an account, add the people you share money with, and let
          OweNone compress the graph into payments you can actually finish.
        </p>
        <div className="hero-actions">
          <Link href="/signup" className="btn btn-primary">
            Create free account
          </Link>
          <Link href="/login" className="btn btn-soft">
            I already have an account
          </Link>
        </div>
      </div>
    </section>
  );
}
