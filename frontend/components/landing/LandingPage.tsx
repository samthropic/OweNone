import { LandingCoreFeaturesSection } from "@/components/landing/LandingCoreFeaturesSection";
import { LandingHeroSection } from "@/components/landing/LandingHeroSection";
import { LandingHowSection } from "@/components/landing/LandingHowSection";
import { LandingNavBar } from "@/components/landing/LandingNavBar";
import { LandingProblemSection } from "@/components/landing/LandingProblemSection";
import { LandingGetStartedSection } from "@/components/landing/LandingGetStartedSection";
import { LandingWhySection } from "@/components/landing/LandingWhySection";
import type { AuthUser } from "@/lib/api-types";

export function LandingPage({ user }: { user: AuthUser | null }) {
  return (
    <div className="landing-root">
      <header className="topbar">
        <div className="landing-shell">
          <LandingNavBar user={user} />
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
        <LandingGetStartedSection />
      </main>
    </div>
  );
}
