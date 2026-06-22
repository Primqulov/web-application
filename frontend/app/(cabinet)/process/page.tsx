"use client";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, Application } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { StatusBadge } from "@/components/StatusBadge";
import { Phone, MapPin } from "lucide-react";
import { T } from "@/components/T";
import Link from "next/link";

export default function Process() {
  const [tab, setTab] = useState<"worker" | "employer">("worker");
  const qc = useQueryClient();

  const { data: mine } = useQuery<Application[]>({
    queryKey: ["my-applications"],
    queryFn: () => api.get<Application[]>("/api/my/applications"),
  });
  const { data: received } = useQuery<Record<string, Application[]>>({
    queryKey: ["my-elons-applications"],
    queryFn: () => api.get<Record<string, Application[]>>("/api/my/elons/applications"),
  });

  const accept = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/accept`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["my-elons-applications"] }),
  });
  const reject = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/reject`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["my-elons-applications"] }),
  });
  const cancel = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/cancel`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["my-applications"] }),
  });
  const done = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/confirm-done`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["my-applications"] });
      qc.invalidateQueries({ queryKey: ["my-elons-applications"] });
    },
  });

  return (
    <Shell title="Jarayonlar">
      <div className="card p-2 flex gap-2">
        <button onClick={() => setTab("worker")} className={`flex-1 rounded-lg px-3 py-2 text-sm ${tab === "worker" ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}><T>Ishlardagi jarayon</T></button>
        <button onClick={() => setTab("employer")} className={`flex-1 rounded-lg px-3 py-2 text-sm ${tab === "employer" ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}><T>E'lonlardagi jarayon</T></button>
      </div>

      {tab === "worker" && (
        <div className="grid sm:grid-cols-2 gap-4">
          {(mine || []).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)] sm:col-span-2"><T>Sizda ariza yo'q.</T></div>}
          {(mine || []).map((a) => (
            <div key={a.id} className="card p-4">
              <div className="flex items-center justify-between"><Link href={`/elon/${a.elonId}`} className="font-semibold"><T>{a.elonTitle}</T></Link><StatusBadge status={a.status} /></div>
              {a.status === "pending" && <p className="text-sm text-[color:var(--text-muted)] mt-2"><T>Ariza ko'rib chiqilmoqda…</T></p>}
              {a.status === "accepted" && (
                <div className="mt-2 grid gap-2">
                  <p className="text-sm text-success"><T>Ish beruvchi tomonidan qabul qilindi</T></p>
                  <div className="flex gap-2">
                    <a href={`tel:${a.workerPhone}`} className="btn-secondary gap-2"><Phone size={14} /><T>Qo'ng'iroq qilish</T></a>
                    <Link href={`/elon/${a.elonId}`} className="btn-secondary gap-2"><MapPin size={14} /><T>Manzilni ko'rish</T></Link>
                    <button onClick={() => done.mutate(a.id)} className="btn-primary"><T>Bajarildi</T></button>
                    <button onClick={() => cancel.mutate(a.id)} className="btn-danger"><T>Bekor qilish</T></button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {tab === "employer" && (
        <div className="grid gap-4">
          {Object.keys(received || {}).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)]"><T>Hozircha arizalar yo'q.</T></div>}
          {Object.entries(received || {}).map(([elonId, apps]) => (
            <div key={elonId} className="card p-4">
              <div className="flex items-center justify-between mb-3">
                <Link href={`/elon/${elonId}`} className="font-semibold"><T>{apps[0]?.elonTitle || "E'lon"}</T></Link>
                <span className="badge-amber">{apps.length} <T>ta ariza</T></span>
              </div>
              <div className="grid gap-2">
                {apps.map((a) => (
                  <div key={a.id} className="flex flex-wrap items-center gap-2 border-t pt-2" style={{ borderColor: "var(--border)" }}>
                    <div className="grid h-8 w-8 place-items-center rounded-full bg-brand-navy text-white text-xs">{a.workerPhone?.slice(-2) || "?"}</div>
                    <div className="mr-auto">
                      <div className="font-medium text-sm">{a.workerPhone}</div>
                      <div className="text-xs text-[color:var(--text-muted)]"><T>{a.status}</T></div>
                    </div>
                    <StatusBadge status={a.status} />
                    <a href={`tel:${a.workerPhone}`} className="btn-secondary gap-1"><Phone size={12} /><T>Qo'ng'iroq</T></a>
                    {a.status === "pending" && <>
                      <button onClick={() => accept.mutate(a.id)} className="btn-primary"><T>Qabul qilish</T></button>
                      <button onClick={() => reject.mutate(a.id)} className="btn-danger">×</button>
                    </>}
                    {a.status === "accepted" && (
                      <button onClick={() => done.mutate(a.id)} className="btn-primary"><T>Bajarildi</T></button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </Shell>
  );
}
