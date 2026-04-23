export type NavLink = {
  href: string;
  label: string;
};

export type TrustAvatar = {
  initials: string;
  gradient: string;
};

export type FeatureCard = {
  id: string;
  title: string;
  description: string;
  statValue: string;
  statLabel: string;
  statTone: "blue" | "teal" | "purple";
  iconVariant: "blue" | "teal" | "purple";
};

export type SecurityItem = {
  id: string;
  title: string;
  description: string;
};

export type FooterLink = {
  href: string;
  label: string;
};
