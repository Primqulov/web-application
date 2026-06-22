"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function AdminDashboard() {
  const [stats, setStats] = useState<any>(null);
  useEffect(() => { api.get<any>("/api/admin/dashboard", { auth: "admin" } as any).then(setStats); }, []);
  return (
    <div className="grid sm:grid-cols-4 gap-3">
      <Card label="Foydalanuvchilar" value={stats?.users ?? "—"} />
      <Card label="E'lonlar" value={stats?.elons ?? "—"} />
      <Card label="Bajarilgan ishlar" value={stats?.completed ?? "—"} />
      <Card label="Aylanma (so'm)" value={(stats?.circulated ?? 0).toLocaleString("uz-UZ")} />
    </div>
  );
}
function Card({ label, value }: { label: string; value: any }) {
  return <div className="card p-5"><div className="text-sm text-[color:var(--text-muted)]">{label}</div><div className="text-2xl font-bold mt-1">{value}</div></div>;
}
