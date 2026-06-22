"use client";
import * as React from "react";

interface Props {
  name?: string;
  src?: string;
  size?: "xs" | "sm" | "md" | "lg" | "xl";
  online?: boolean;
}

const sz = { xs: 24, sm: 32, md: 40, lg: 56, xl: 80 } as const;
const ts = { xs: 11, sm: 12, md: 14, lg: 18, xl: 24 } as const;

function initials(name?: string) {
  if (!name) return "?";
  const parts = name.trim().split(/\s+/);
  const a = parts[0]?.[0] || "";
  const b = parts[1]?.[0] || "";
  return (a + b).toUpperCase() || a.toUpperCase() || "?";
}

// Deterministic color based on name — gives users a consistent color identity.
function colorFor(name?: string) {
  if (!name) return { bg: "var(--brand)", fg: "#fff" };
  let h = 0;
  for (const c of name) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  const hue = h % 360;
  return { bg: `hsl(${hue}, 55%, 38%)`, fg: "#fff" };
}

export function Avatar({ name, src, size = "md", online }: Props) {
  const px = sz[size];
  const fs = ts[size];
  const { bg, fg } = colorFor(name);
  return (
    <div className="relative inline-flex shrink-0">
      {src ? (
        <img
          src={src} alt={name || ""}
          width={px} height={px}
          className="rounded-full object-cover"
          style={{ width: px, height: px }}
        />
      ) : (
        <div
          className="rounded-full grid place-items-center font-semibold tracking-wide"
          style={{ width: px, height: px, background: bg, color: fg, fontSize: fs }}
        >
          {initials(name)}
        </div>
      )}
      {online && (
        <span
          className="absolute bottom-0 right-0 ring-2 rounded-full"
          style={{
            width: Math.max(8, px * 0.22),
            height: Math.max(8, px * 0.22),
            background: "#16A34A",
            // @ts-ignore
            "--tw-ring-color": "var(--card)",
          }}
        />
      )}
    </div>
  );
}
