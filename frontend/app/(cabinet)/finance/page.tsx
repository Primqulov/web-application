"use client";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api, FinanceSummary } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { T } from "@/components/T";
import { fmtSum, fmtSumSom } from "@/lib/format";
import { StatusBadge } from "@/components/StatusBadge";

export default function Finance() {
  const [role, setRole] = useState<"worker" | "employer">("worker");
  const { data } = useQuery<FinanceSummary>({
    queryKey: ["finance"],
    queryFn: () => api.get<FinanceSummary>("/api/me/finance"),
  });

  const entries = useMemo(() => (data?.entries || []).filter((e) => e.role === role), [data, role]);
  const sum = role === "worker" ? data?.earnedSum ?? 0 : data?.spentSum ?? 0;

  return (
    <Shell title="Moliya">
      <div className="card p-2 flex gap-2">
        <button onClick={() => setRole("worker")} className={`flex-1 rounded-lg px-3 py-2 text-sm ${role === "worker" ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}><T>Ishchi sifatida</T></button>
        <button onClick={() => setRole("employer")} className={`flex-1 rounded-lg px-3 py-2 text-sm ${role === "employer" ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}><T>Ish beruvchi sifatida</T></button>
      </div>
      <div className="grid sm:grid-cols-3 gap-3">
        <Card label={role === "worker" ? "Jami ishlangan" : "Jami sarflangan"} value={`${fmtSum(sum)} so'm`} />
        <Card label="Kelishilgan ishlar" value={`${data?.negotiableCount ?? 0} ta`} />
        <Card label="Bekor qilingan" value={`${data?.cancelledCount ?? 0} ta`} />
      </div>
      <div className="grid gap-2">
        {entries.length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)]"><T>Bo'sh.</T></div>}
        {entries.map((e) => (
          <div key={e.id} className="card p-4 flex items-center justify-between gap-3">
            <div>
              <div className="font-medium"><T>{e.elonTitle}</T></div>
              <div className="text-xs text-[color:var(--text-muted)]">{new Date(e.createdAt).toLocaleDateString("uz-UZ")}</div>
            </div>
            <div className="flex items-center gap-3">
              <span className={e.status === "cancelled" ? "line-through text-[color:var(--text-muted)]" : "font-semibold text-accent-amber"}>
                {fmtSumSom(e.amount, e.isNegotiable)}
              </span>
              <StatusBadge status={e.status === "cancelled" ? "cancelled" : "completed"} />
            </div>
          </div>
        ))}
      </div>
    </Shell>
  );
}

function Card({ label, value }: { label: string; value: string }) {
  return (
    <div className="card p-5">
      <div className="text-sm text-[color:var(--text-muted)]"><T>{label}</T></div>
      <div className="text-xl font-bold mt-1">{value}</div>
    </div>
  );
}
