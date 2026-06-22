import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./features/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50:  "#EEF1FA",
          100: "#D4DCF1",
          200: "#A8B7E1",
          300: "#7B91D0",
          400: "#4F6CC0",
          500: "#2347B0",
          600: "#1B3990",
          700: "#16245C", // main
          800: "#0F1F56",
          900: "#0A1542",
          navy: "#0F1F56",
          navy700: "#16245C",
        },
        accent: {
          50:  "#FEF6E7",
          100: "#FDE9C0",
          200: "#FBD589",
          300: "#F8BE52",
          400: "#F1A926",
          500: "#E8920A",
          600: "#C97A08",
          amber: "#E8920A",
          amberBg: "#FEF6E7",
        },
        tg: { blue: "#229ED9", darkBlue: "#1A85B5" },
        success: { DEFAULT: "#16A34A", bg: "#DCFCE7" },
        pending: { DEFAULT: "#B45309", bg: "#FEF3C7" },
        danger:  { DEFAULT: "#DC2626", bg: "#FEE2E2" },
        info:    { DEFAULT: "#2563EB", bg: "#DBEAFE" },
      },
      borderRadius: {
        xs: "0.25rem", sm: "0.375rem", DEFAULT: "0.5rem",
        md: "0.625rem", lg: "0.75rem", xl: "1rem", "2xl": "1.25rem", "3xl": "1.5rem",
      },
      boxShadow: {
        sm:   "0 1px 2px rgba(15,23,42,0.04)",
        card: "0 1px 2px rgba(15,23,42,0.04), 0 0 0 1px rgba(15,23,42,0.03)",
        pop:  "0 8px 24px -8px rgba(15,23,42,0.18), 0 2px 6px rgba(15,23,42,0.06)",
        ring: "0 0 0 4px rgba(15,31,86,0.12)",
      },
      transitionTimingFunction: {
        spring: "cubic-bezier(0.34, 1.56, 0.64, 1)",
      },
      keyframes: {
        "fade-in":   { "0%": { opacity: "0" }, "100%": { opacity: "1" } },
        "slide-up":  { "0%": { opacity: "0", transform: "translateY(8px)" }, "100%": { opacity: "1", transform: "translateY(0)" } },
        "scale-in":  { "0%": { opacity: "0", transform: "scale(0.96)" }, "100%": { opacity: "1", transform: "scale(1)" } },
        "shimmer":   { "0%": { backgroundPosition: "-200% 0" }, "100%": { backgroundPosition: "200% 0" } },
      },
      animation: {
        "fade-in":  "fade-in 200ms ease-out",
        "slide-up": "slide-up 240ms ease-out",
        "scale-in": "scale-in 180ms ease-out",
        "shimmer":  "shimmer 1.5s linear infinite",
      },
      fontFamily: {
        sans: [
          "Inter", "ui-sans-serif", "system-ui", "-apple-system", "Segoe UI",
          "Roboto", "Helvetica Neue", "Arial", "sans-serif",
        ],
      },
      fontSize: {
        "2xs": ["0.6875rem", { lineHeight: "1rem" }],
      },
    },
  },
  plugins: [],
};
export default config;
