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
  return (
    <div className="min-h-screen p-4">
      <div className="mx-auto max-w-7xl grid grid-cols-1 md:grid-cols-[240px_1fr] gap-4">
        <aside className="card p-4">
          <div className="heading font-bold mb-3">IB Admin</div>
          <nav className="grid gap-1">
            {nav.map((n) => (
              <Link key={n.href} href={n.href} className={`rounded-lg px-3 py-2 text-sm ${pathname === n.href ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}>{n.label}</Link>
            ))}
          </nav>
          <button onClick={() => { setAdminToken(null); router.replace("/admin/login"); }} className="mt-4 btn-secondary w-full">Chiqish</button>
        </aside>
        <main className="grid gap-4">{children}</main>
      </div>
    </div>
  );
}
