export type LandingNavItem = {
  href: string;
  label: string;
};

export type LandingHeroStat = {
  value: string;
  label: string;
};

export type LandingProblemCard = {
  title: string;
  body: string;
  highlighted?: boolean;
};

export type LandingFeature = {
  title: string;
  description: string;
};

export type NetworkNodeTone = "green" | "rose" | "sky" | "amber";

export type LandingNetworkNode = {
  id: string;
  name: string;
  initials: string;
  balance: string;
  x: number;
  y: number;
  tone: NetworkNodeTone;
};

export type LandingConnection = {
  from: string;
  to: string;
  amount: string;
  color: "rose" | "green";
};

export type LandingFlow = {
  label: string;
  amount: string;
};

export type LandingMiniStat = {
  value: string;
  label: string;
};

export type LandingStep = {
  number: string;
  title: string;
  description: string;
};
