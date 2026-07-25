import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

export const metadata: Metadata = {
  title: "OweNone — Untangle group IOUs",
  description:
    "The social graph debt compressor: fewer payments, live sync, explainable settlements for trips, flats, and dinners.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    // Some browser extensions inject attributes before React hydrates.
    // This prevents noisy hydration warnings for those external mutations.
    <html
      lang="en"
      className={`${inter.variable} antialiased`}
      suppressHydrationWarning
    >
      <body suppressHydrationWarning>{children}</body>
    </html>
  );
}