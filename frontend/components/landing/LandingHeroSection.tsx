"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  AFTER_CONNECTIONS,
  AFTER_FLOWS,
  BEFORE_CONNECTIONS,
  BEFORE_FLOWS,
  NETWORK_NODES,
} from "./landingData";
import { LandingNetwork } from "./LandingNetwork";

export function LandingHeroSection() {
  const [compressed, setCompressed] = useState(false);

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (media.matches) {
      setCompressed(true);
      return;
    }

    const timer = window.setTimeout(() => setCompressed(true), 1300);
    return () => window.clearTimeout(timer);
  }, []);

  const connections = compressed ? AFTER_CONNECTIONS : BEFORE_CONNECTIONS;
  const flows = compressed ? AFTER_FLOWS : BEFORE_FLOWS;

  return (
    <section className="hero">
      <div className="hero-copy">
        <h1 className="hero-title">
          Drowning in group IOUs?{" "}
          <em>Let the tide go out.</em>
        </h1>
        <p className="lead">
          Shared dinners, trips, and rent pile up like high water. OweNone nets
          the whole harbor and leaves only the payments that still matter.
        </p>
        <div className="hero-actions">
          <Link href="/signup" className="btn btn-primary">
            Create your account
          </Link>
        </div>
      </div>

      <aside className="tide-card" aria-live="polite">
        <header className="tide-card-head">
          <div>
            <p className="tide-kicker">Today&apos;s clearing</p>
            <p className="tide-status">
              {compressed ? "Low tide · 2 payments" : "High tide · 3 payments"}
            </p>
          </div>
          <button
            type="button"
            className="btn btn-soft stage-toggle"
            onClick={() => setCompressed((value) => !value)}
            aria-pressed={compressed}
          >
            {compressed ? "Show high tide" : "Clear the tide"}
          </button>
        </header>

        <LandingNetwork
          nodes={NETWORK_NODES}
          connections={connections}
          variant={compressed ? "compressed" : "default"}
        />

        <ul className="tide-chips">
          {flows.map((flow) => (
            <li key={flow.label}>
              <span>{flow.label}</span>
              <strong>{flow.amount}</strong>
            </li>
          ))}
        </ul>
      </aside>
    </section>
  );
}
