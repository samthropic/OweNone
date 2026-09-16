import type { NextConfig } from "next";
import path from "node:path";

const apiOrigin =
  process.env.OWENONE_API_URL ??
  (process.env.VERCEL_URL
    ? `https://${process.env.VERCEL_URL}`
    : "http://localhost:8080");

const nextConfig: NextConfig = {
  // Vercel multi-service rewrites break `/_next/image`; serve public assets directly.
  images: {
    unoptimized: true,
  },
  turbopack: {
    root: path.resolve(__dirname),
  },
  async rewrites() {
    return [
      {
        source: "/api/v1/avatars/:userId",
        destination: `${apiOrigin}/api/v1/avatars/:userId`,
      },
    ];
  },
};

export default nextConfig;
