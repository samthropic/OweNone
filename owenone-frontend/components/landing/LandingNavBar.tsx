import Link from "next/link";
import { LANDING_NAV_ITEMS } from "./landingData";

export function LandingNavBar() {
  return (
    <nav className="landing-nav" aria-label="Primary">
      <Link href="/" className="brand">
        <span className="brand-mark">∞</span>
        <span className="brand-copy">
          <strong>OweNone</strong>
          <small>Luxury infrastructure for shared money</small>
        </span>
      </Link>
      <div className="nav-links">
        {LANDING_NAV_ITEMS.map((item) => (
          <a key={item.href} href={item.href}>
            {item.label}
          </a>
        ))}
      </div>
      <a href="#waitlist" className="request-btn">
        Request access
      </a>
    </nav>
  );
}
