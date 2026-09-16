import { HomeActivityCard } from "@/components/home/HomeActivityCard";
import { HomeBalanceBanner } from "@/components/home/HomeBalanceBanner";
import { HomeFriendsCard } from "@/components/home/HomeFriendsCard";
import { HomeGroupsCard } from "@/components/home/HomeGroupsCard";
import { HomeHeader } from "@/components/home/HomeHeader";
import { HomeSidebar } from "@/components/home/HomeSidebar";
import type { Dashboard } from "@/lib/api-types";

type HomePageProps = {
  dashboard: Dashboard;
};

export function HomePage({ dashboard }: HomePageProps) {
  return (
    <div className="home-root">
      <HomeSidebar user={dashboard.user} active="Overview" />

      <main className="home-main">
        <div className="home-main-inner">
          <HomeHeader dashboard={dashboard} />
          <HomeBalanceBanner dashboard={dashboard} />

          <div className="home-grid">
            <HomeFriendsCard dashboard={dashboard} />
            <HomeGroupsCard dashboard={dashboard} />
          </div>

          <div className="home-activity-wide">
            <HomeActivityCard dashboard={dashboard} />
          </div>
        </div>
      </main>
    </div>
  );
}
