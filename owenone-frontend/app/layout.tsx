import type { Metadata } from "next";
import { DM_Sans, Syne } from "next/font/google";
import { Inter } from 'next/font/google';
import "./globals.css";

// Configure Inter
const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter', // This matches the CSS variable you just set
  display: 'swap',
});

const dmSans = DM_Sans({
  variable: "--font-dm-sans",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
});

const syne = Syne({
  variable: "--font-syne",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800"],
});

export const metadata: Metadata = {
  title: "OweNone — Untangle group IOUs",
  description:
    "The social graph debt compressor: fewer payments, live sync, explainable settlements for trips, flats, and dinners.",
};

// export default function RootLayout({
//   children,
// }: Readonly<{
//   children: React.ReactNode;
// }>) {
//   return (
//     <html lang="en" className={`${dmSans.variable} ${syne.variable} h-full antialiased`}>
//       <body className="min-h-full">{children}</body>
//     </html>
//   );
// }

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    // Apply the variable and Next.js antialiasing to the HTML/Body
    <html lang="en" className={`${inter.variable} antialiased`}>
      <body>
        {children}
      </body>
    </html>
  );
}