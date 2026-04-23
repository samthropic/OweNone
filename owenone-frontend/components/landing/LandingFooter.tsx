import { FOOTER_LINKS } from "./content";

export function LandingFooter() {
  const year = new Date().getFullYear();
  return (
    <footer className="landing-footer">
      <div className="footer-inner">
        <p className="footer-copy">© {year} OweNone. All rights reserved.</p>
        <nav className="footer-links" aria-label="Footer">
          {FOOTER_LINKS.map((link) => (
            <a key={link.label} href={link.href}>
              {link.label}
            </a>
          ))}
        </nav>
      </div>
    </footer>
  );
}
