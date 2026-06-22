"use client";
import { useScript, tr } from "@/lib/i18n";

export function T({ children }: { children: string }) {
  const script = useScript((s) => s.script);
  return <>{tr(children, script)}</>;
}

/** Hook helper for plain strings inside attributes. */
export function useT() {
  const script = useScript((s) => s.script);
  return (s: string) => tr(s, script);
}
