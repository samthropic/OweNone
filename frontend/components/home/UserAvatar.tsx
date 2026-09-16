import { initials } from "@/components/home/homeData";
import type { User } from "@/lib/api-types";

type AvatarUser = Pick<User, "displayName" | "avatarUrl">;

const API_ORIGIN =
  process.env.NEXT_PUBLIC_OWENONE_API_URL ??
  process.env.OWENONE_API_URL ??
  "http://localhost:8080";

function resolveAvatarUrl(path: string) {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }
  return new URL(path, API_ORIGIN).toString();
}

export function UserAvatar({
  user,
  className = "home-avatar",
  tone,
  size = "md",
}: {
  user: AvatarUser;
  className?: string;
  tone?: string;
  size?: "md" | "lg";
}) {
  const classes = [
    className,
    tone ? `tone-${tone}` : null,
    size === "lg" ? "is-lg" : null,
    user.avatarUrl ? "has-photo" : null,
  ]
    .filter(Boolean)
    .join(" ");

  if (user.avatarUrl) {
    return (
      // eslint-disable-next-line @next/next/no-img-element -- remote avatar URLs from the API
      <img
        src={resolveAvatarUrl(user.avatarUrl)}
        alt=""
        className={classes}
        width={size === "lg" ? 52 : 38}
        height={size === "lg" ? 52 : 38}
      />
    );
  }

  return (
    <div className={classes} aria-hidden="true">
      {initials(user)}
    </div>
  );
}
