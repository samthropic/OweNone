import Link from "next/link";
import Image from "next/image";
import { UserAvatar } from "@/components/home/UserAvatar";
import type { AuthUser } from "@/lib/api-types";
import { LANDING_NAV_ITEMS } from "./landingData";

export function LandingNavBar({ user }: { user: AuthUser | null }) {
  return (
    <nav className="landing-nav" aria-label="Primary">
      <Link href="/" className="brand">
        <Image
          src="/owenone-mark.png"
          alt=""
          width={256}
          height={259}
          className="brand-mark"
          preload={true}
        />
        <div className="brand-text">
          <p className="brand-name">OweNone</p>
          <p className="brand-tag">Untangle group IOUs</p>
        </div>
      </Link>
      <div className="nav-links">
        {LANDING_NAV_ITEMS.map((item) => (
          <a key={item.href} href={item.href}>
            {item.label}
          </a>
        ))}
      </div>
      <div className="nav-actions">
        {user ? (
          <Link
            href="/"
            className="nav-user"
            aria-label={`Open your dashboard, signed in as ${user.displayName}`}
          >
            <UserAvatar user={user} className="nav-user-avatar" />
          </Link>
        ) : (
          <>
            <Link href="/login" className="nav-login">
              Log in
            </Link>
            <Link href="/signup" className="btn btn-primary">
              Sign up
            </Link>
          </>
        )}
      </div>
    </nav>
  );
}
