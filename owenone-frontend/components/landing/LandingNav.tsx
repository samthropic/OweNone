import Link from "next/link";
import { NAV_LINKS } from "./content";

function LogoMark() {
  return (
    <svg viewBox="0 0 20 20" fill="none" aria-hidden>
      <circle cx="5" cy="5" r="2.5" fill="white" opacity="0.9" />
      <circle cx="15" cy="5" r="2.5" fill="white" opacity="0.9" />
      <circle cx="5" cy="15" r="2.5" fill="white" opacity="0.9" />
      <circle cx="15" cy="15" r="2.5" fill="white" opacity="0.9" />
      <line
        x1="5"
        y1="5"
        x2="15"
        y2="15"
        stroke="white"
        strokeWidth="1.5"
        opacity="0.5"
      />
      <line
        x1="15"
        y1="5"
        x2="5"
        y2="15"
        stroke="white"
        strokeWidth="1.5"
        opacity="0.5"
      />
      <circle cx="10" cy="10" r="2" fill="white" />
    </svg>
  );
}

export function LandingNav() {
  return (
    <nav className="landing-nav" aria-label="Primary">
      <div className="nav-inner">
        <Link className="logo" href="/">
          <div className="logo-mark">
            <LogoMark />
          </div>
          <span className="logo-name">
            Owe<span>None</span>
          </span>
        </Link>
        <div className="nav-links">
          {NAV_LINKS.map((link) => (
            <a key={link.href} href={link.href}>
              {link.label}
            </a>
          ))}
        </div>
        <div className="nav-right">
          <button type="button" className="avatar-pill" aria-haspopup="menu">
            <span className="avatar">SJ</span>
            <span className="avatar-name">Sarah J.</span>
            <svg
              width="12"
              height="12"
              viewBox="0 0 12 12"
              fill="none"
              aria-hidden
              style={{ color: "#475569", marginLeft: 2 }}
            >
              <path
                d="M3 4.5L6 7.5L9 4.5"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
              />
            </svg>
          </button>
          <button type="button" className="btn-cta">
            Get Early Access
          </button>
        </div>
      </div>
    </nav>
  );
}
