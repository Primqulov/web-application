"use client";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { getAdminToken, setAdminToken } from "@/lib/api";

const nav = [
  { href: "/admin", label: "Dashboard" },
  { href: "/admin/users", label: "Foydalanuvchilar" },
  { href: "/admin/elons", label: "E'lonlar" },
  { href: "/admin/categories", label: "Turkumlar" },
  { href: "/admin/reports", label: "Shikoyatlar" },
  { href: "/admin/feedback", label: "Taklif va shikoyatlar" },
  { href: "/admin/notifications", label: "Bildirishnomalar" },
  { href: "/admin/audit", label: "Audit log" },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  useEffect(() => {
    if (pathname !== "/admin/login" && !getAdminToken()) router.replace("/admin/login");
  }, [pathname, router]);
  if (pathname === "/admin/login") return <>{children}</>;
  const logout = () => { setAdminToken(null); router.replace("/admin/login"); };
  return (
    <div className="min-h-screen p-3 sm:p-4">
      <div className="mx-auto max-w-7xl grid grid-cols-1 md:grid-cols-[240px_1fr] gap-4">
        {/* On mobile the sidebar collapses to a sticky top bar with a
            horizontally scrollable pill nav; on md+ it's a vertical sidebar. */}
        <aside className="card p-3 md:p-4 flex flex-col md:sticky md:top-4 md:h-[calc(100vh-2rem)]">
          <div className="flex items-center justify-between gap-2 mb-3">
            <div className="heading font-bold">IB Admin</div>
            <button onClick={logout} className="btn-secondary btn-sm md:hidden">Chiqish</button>
          </div>
          <nav className="flex md:flex-col gap-1 overflow-x-auto md:overflow-y-auto scroll-y-auto -mx-1 px-1 md:flex-1">
            {nav.map((n) => (
              <Link
                key={n.href}
                href={n.href}
                className={`shrink-0 whitespace-nowrap rounded-lg px-3 py-2 text-sm ${pathname === n.href ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}
              >
                {n.label}
              </Link>
            ))}
          </nav>
          <button onClick={logout} className="mt-4 btn-secondary w-full hidden md:block">Chiqish</button>
        </aside>
        <main className="grid gap-4 min-w-0">{children}</main>
      </div>
    </div>
  );
}
