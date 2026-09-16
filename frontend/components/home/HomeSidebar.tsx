import Image from "next/image";
import Link from "next/link";
import { UserAvatar } from "@/components/home/UserAvatar";
import type { User } from "@/lib/api-types";

export type SidebarTab = "Overview" | "Friends" | "Groups" | "Activity" | "Settle Up" | "Profile";

const ROUTES: Partial<Record<SidebarTab, string>> = {
  Overview: "/",
  Friends: "/friends",
  Groups: "/groups",
  Activity: "/activity",
  "Settle Up": "/settle-up",
};

const MENU_ITEMS: SidebarTab[] = ["Overview", "Friends", "Groups", "Activity", "Settle Up"];

export function HomeSidebar({ user, active = "Overview" }: { user: User; active?: SidebarTab }) {
  return (
    <aside className="home-sidebar">
      <Link href="/" className="home-brand">
        <Image
          src="/owenone-mark.png"
          alt=""
          width={256}
          height={259}
          className="home-brand-mark"
          priority
        />
        <div>
          <p className="home-brand-name">OweNone</p>
          <p className="home-brand-tag">Untangle group IOUs</p>
        </div>
      </Link>

      <p className="home-sidebar-label">Menu</p>
      <nav className="home-nav">
        {MENU_ITEMS.map((item) => {
          const isActive = item === active;
          const href = ROUTES[item];
          const className = `home-nav-item${isActive ? " is-active" : ""}`;

          if (href !== undefined) {
            return (
              <Link
                key={item}
                href={href}
                className={className}
                aria-current={isActive ? "page" : undefined}
              >
                {item}
              </Link>
            );
          }

          return (
            <button
              key={item}
              type="button"
              className={className}
              disabled
              title={`${item} is not available yet`}
            >
              {item}
            </button>
          );
        })}
      </nav>

      <div className="home-sidebar-footer">
        <Link
          href="/profile"
          className={`home-sidebar-user${active === "Profile" ? " is-active" : ""}`}
          aria-current={active === "Profile" ? "page" : undefined}
          aria-label="Open your profile"
        >
          <UserAvatar user={user} tone="indigo" />
          <div>
            <p className="home-user-name">{user.displayName}</p>
            <p className="home-user-email">{user.email}</p>
          </div>
        </Link>
      </div>
    </aside>
  );
}
