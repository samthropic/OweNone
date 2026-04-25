import { HomeActivityCard } from "@/components/home/HomeActivityCard";
import { HomeBalanceBanner } from "@/components/home/HomeBalanceBanner";
import { HomeFriendsCard } from "@/components/home/HomeFriendsCard";
import { HomeHeader } from "@/components/home/HomeHeader";
import { HomeRightColumn } from "@/components/home/HomeRightColumn";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import { HomeSuggestionCard } from "@/components/home/HomeSuggestionCard";

export function HomePage() {
  return (
    <div className="home-root">
      <HomeSidebar />

      <main className="home-main">
        <div className="home-main-inner">
          <HomeHeader />
          <HomeBalanceBanner />
          <HomeSuggestionCard />

          <div className="home-grid">
            <div className="home-left-column">
              <HomeFriendsCard />
              <HomeActivityCard />
            </div>
            <HomeRightColumn />
          </div>
        </div>
      </main>
    </div>
  );
}
