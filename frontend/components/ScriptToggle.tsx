"use client";
import { useScript } from "@/lib/i18n";

export function ScriptToggle({ className = "" }: { className?: string }) {
  const script = useScript((s) => s.script);
  const setScript = useScript((s) => s.setScript);
  return (
    <div className={`inline-flex items-center rounded-lg border overflow-hidden text-sm ${className}`} style={{ borderColor: "var(--border)" }}>
      <button
        onClick={() => setScript("latin")}
        className={`px-3 py-1.5 ${script === "latin" ? "bg-brand-navy text-white" : "text-[color:var(--text-muted)]"}`}
      >
        Lotin
      </button>
      <button
        onClick={() => setScript("cyrillic")}
        className={`px-3 py-1.5 ${script === "cyrillic" ? "bg-brand-navy text-white" : "text-[color:var(--text-muted)]"}`}
      >
        Кирилл
      </button>
    </div>
  );
}
