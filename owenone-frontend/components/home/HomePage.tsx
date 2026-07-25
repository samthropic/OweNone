import { HomeActivityCard } from "@/components/home/HomeActivityCard";
import { HomeBalanceBanner } from "@/components/home/HomeBalanceBanner";
import { HomeFriendsCard } from "@/components/home/HomeFriendsCard";
import { HomeHeader } from "@/components/home/HomeHeader";
import { HomeRightColumn } from "@/components/home/HomeRightColumn";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { HomeSuggestionCard } from "@/components/home/HomeSuggestionCard";
import type { Dashboard } from "@/lib/api-types";

type HomePageProps = {
  dashboard: Dashboard;
};

export function HomePage({ dashboard }: HomePageProps) {
  return (
    <div className="home-root">
      <HomeSidebar user={dashboard.user} />

      <main className="home-main">
        <div className="home-main-inner">
          <HomeHeader dashboard={dashboard} />
          <HomeBalanceBanner dashboard={dashboard} />
          <HomeSuggestionCard dashboard={dashboard} />

          <div className="home-grid">
            <div className="home-left-column">
              <HomeFriendsCard dashboard={dashboard} />
              <HomeActivityCard dashboard={dashboard} />
            </div>
            <HomeRightColumn dashboard={dashboard} />
          </div>
        </div>
      </main>
    </div>
  );
}
