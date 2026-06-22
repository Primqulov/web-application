"use client";
import { ReactNode } from "react";

/**
 * Cabinet layout is intentionally minimal — each cabinet page renders the
 * <Shell title=... /> with its own search/topbar config to keep flexibility
 * per-page (e.g. dashboard wants a search in the topbar, others don't).
 */
export default function CabinetLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
