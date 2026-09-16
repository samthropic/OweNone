import { dashboardDate } from "@/components/home/homeData";
import { HeaderActions } from "@/components/home/HomeActionForms";
import type { Dashboard } from "@/lib/api-types";

export function HomeHeader({ dashboard }: { dashboard: Dashboard }) {
  return (
    <header className="home-header">
      <div>
        <h1>Overview</h1>
        <p>{dashboardDate(dashboard)}</p>
      </div>
      <div className="home-header-actions">
        <HeaderActions dashboard={dashboard} />
      </div>
    </header>
  );
}
