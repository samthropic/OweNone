import "./landing.css";
import { BackgroundAtmosphere } from "./BackgroundAtmosphere";
import { DebtGraphPreview } from "./DebtGraphPreview";
import { LandingBottomGrid } from "./LandingBottomGrid";
import { LandingFeatures } from "./LandingFeatures";
import { LandingFooter } from "./LandingFooter";
import { LandingHero } from "./LandingHero";
import { LandingNav } from "./LandingNav";

export function LandingPage() {
  return (
    <div className="landing-root">
      <BackgroundAtmosphere />
      <LandingNav />
      <main>
        <LandingHero />
        <DebtGraphPreview />
        <LandingFeatures />
        <LandingBottomGrid />
      </main>
      <LandingFooter />
    </div>
  );
}
