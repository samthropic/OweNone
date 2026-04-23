import "./landing.css";
import { LandingCoreFeaturesSection } from "@/components/landing/LandingCoreFeaturesSection";
import { LandingHeroSection } from "@/components/landing/LandingHeroSection";
import { LandingHowSection } from "@/components/landing/LandingHowSection";
import { LandingNavBar } from "@/components/landing/LandingNavBar";
import { LandingProblemSection } from "@/components/landing/LandingProblemSection";
import { LandingWaitlistSection } from "@/components/landing/LandingWaitlistSection";
import { LandingWhySection } from "@/components/landing/LandingWhySection";

export function LandingPage() {
  return (
    <div className="landing-root">
      <header className="topbar">
        <div className="landing-shell">
        <LandingNavBar />
        </div>
      </header>
      <main>
        <div className="landing-shell">
        <LandingHeroSection />
        </div>
        <LandingProblemSection />
        <LandingCoreFeaturesSection />
        <LandingHowSection />
        <LandingWhySection />
        <LandingWaitlistSection />
      </main>
    </div>
  );
}
