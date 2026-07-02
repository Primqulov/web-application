"use client";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api, Application } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { StatusBadge } from "@/components/StatusBadge";
import { Modal } from "@/components/Modal";
import { Search, Filter, Calendar } from "lucide-react";
import { T, useT } from "@/components/T";
import { fmtDate, fmtSumSom } from "@/lib/format";
import dayjs from "dayjs";

function shortReason(s: string, max = 50) {
  return s.length > max ? s.slice(0, max).trimEnd() + "…" : s;
}

export default function History() {
  const t = useT();
  const [q, setQ] = useState("");
  const [reasonView, setReasonView] = useState<Application | null>(null);
  const [status, setStatus] = useState<string>("");
  const [range, setRange] = useState<string>("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [rangeOpen, setRangeOpen] = useState(false);
  const { data } = useQuery<Application[]>({
    queryKey: ["history"],
    queryFn: () => api.get<Application[]>("/api/me/history"),
  });

  const list = useMemo(() => {
    let arr = data || [];
    if (q) arr = arr.filter((a) => a.elonTitle.toLowerCase().includes(q.toLowerCase()));
    if (status) arr = arr.filter((a) => a.status === status);
    if (from) arr = arr.filter((a) => dayjs(a.appliedAt).isAfter(dayjs(from).subtract(1, "day")));
    if (to) arr = arr.filter((a) => dayjs(a.appliedAt).isBefore(dayjs(to).add(1, "day")));
    if (range) {
      const cut = ({
        today: dayjs().startOf("day"),
        yesterday: dayjs().subtract(1, "day").startOf("day"),
        "7d": dayjs().subtract(7, "day"),
        "30d": dayjs().subtract(30, "day"),
        month: dayjs().startOf("month"),
      } as any)[range];
      if (cut) arr = arr.filter((a) => dayjs(a.appliedAt).isAfter(cut));
    }
    return arr;
  }, [data, q, status, from, to, range]);

  const grouped = useMemo(() => {
    const m: Record<string, Application[]> = {};
    for (const a of list) {
      const k = dayjs(a.appliedAt).format("D MMMM");
      (m[k] = m[k] || []).push(a);
    }
    return m;
  }, [list]);

  return (
    <Shell title="Ishlar tarixi">
      <div className="card p-4 flex flex-col sm:flex-row gap-2 sm:items-center">
        <div className="relative flex-1">
          <Search size={18} className="absolute left-3 top-2.5 text-[color:var(--text-muted)]" />
          <input className="input pl-9" placeholder={t("Qidiruv…")} value={q} onChange={(e) => setQ(e.target.value)} />
        </div>
        <div className="relative">
          <button onClick={() => setFiltersOpen((x) => !x)} className="btn-secondary gap-2"><Filter size={14} /><T>Filtr</T></button>
          {filtersOpen && (
            <div className="absolute right-0 mt-2 card p-4 w-72 z-10">
              <label className="block text-sm">
                <span className="text-[color:var(--text-muted)]"><T>Holat</T></span>
                <select className="input mt-1" value={status} onChange={(e) => setStatus(e.target.value)}>
                  <option value="">{t("Barchasi")}</option>
                  <option value="completed">{t("Bajarildi")}</option>
                  <option value="cancelled">{t("Bekor qilindi")}</option>
                  <option value="rejected">{t("Rad etildi")}</option>
                </select>
              </label>
              <div className="mt-3 flex gap-2">
                <button onClick={() => setFiltersOpen(false)} className="btn-primary flex-1"><T>Filtrlarni qo'llash</T></button>
                <button onClick={() => { setStatus(""); }} className="btn-secondary"><T>Tozalash</T></button>
              </div>
            </div>
          )}
        </div>
        <div className="relative">
          <button onClick={() => setRangeOpen((x) => !x)} className="btn-secondary gap-2"><Calendar size={14} /><T>Vaqt oralig'i</T></button>
          {rangeOpen && (
            <div className="absolute right-0 mt-2 card p-4 w-72 z-10 grid gap-2">
              {[
                ["today", "Bugun"], ["yesterday", "Kecha"], ["7d", "Oxirgi 7 kun"],
                ["30d", "Oxirgi 30 kun"], ["month", "Shu oy"],
              ].map(([v, l]) => (
                <button key={v} className={`chip text-left ${range === v ? "chip-active" : ""}`} onClick={() => setRange(v as string)}><T>{l as string}</T></button>
              ))}
              <div className="grid grid-cols-2 gap-2">
                <input type="date" className="input" value={from} onChange={(e) => setFrom(e.target.value)} />
                <input type="date" className="input" value={to} onChange={(e) => setTo(e.target.value)} />
              </div>
              <div className="flex gap-2">
                <button onClick={() => setRangeOpen(false)} className="btn-primary flex-1"><T>Filtrni qo'llash</T></button>
                <button onClick={() => { setRange(""); setFrom(""); setTo(""); }} className="btn-secondary"><T>Tozalash</T></button>
              </div>
            </div>
          )}
        </div>
      </div>

      {Object.keys(grouped).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)]"><T>Bo'sh.</T></div>}
      {Object.entries(grouped).map(([day, items]) => (
        <div key={day} className="grid gap-2">
          <div className="text-sm text-[color:var(--text-muted)]">{day}</div>
          {items.map((a) => (
            <div key={a.id} className="card p-4 flex items-center justify-between gap-3">
              <div>
                <div className={`font-medium ${a.status !== "completed" ? "line-through" : ""}`}><T>{a.elonTitle}</T></div>
                <div className="text-xs text-[color:var(--text-muted)]">{fmtDate(a.completedAt || a.decidedAt || a.appliedAt)}</div>
                {a.status === "cancelled" && (
                  <div className="text-xs text-danger mt-0.5">
                    <T>{a.cancelledBy === "worker" ? "Ishchi tomonidan bekor qilingan" : "Ish beruvchi tomonidan bekor qilingan"}</T>
                    {a.cancelReason && <> — <span className="text-[color:var(--text-muted)]"><T>{shortReason(a.cancelReason)}</T></span></>}
                    {a.cancelReason && a.cancelReason.length > 50 && (
                      <button onClick={() => setReasonView(a)} className="ml-1 text-tg-blue underline"><T>Batafsil</T></button>
                    )}
                  </div>
                )}
                {a.status === "rejected" && (
                  <div className="text-xs text-[color:var(--text-muted)] mt-0.5"><T>Ish beruvchi tomonidan qabul qilinmagan</T></div>
                )}
              </div>
              <div className="flex items-center gap-3">
                <span className={a.status !== "completed" ? "line-through text-[color:var(--text-muted)]" : "font-semibold text-accent-amber"}>
                  {fmtSumSom(a.amount, a.isNegotiable)}
                </span>
                <StatusBadge status={a.status} />
              </div>
            </div>
          ))}
        </div>
      ))}

      {/* Bekor qilish sababini batafsil o'qish */}
      <Modal open={!!reasonView} onClose={() => setReasonView(null)} title={t("Bekor qilish sababi")}>
        {reasonView && (
          <div className="grid gap-2">
            <p className="text-sm font-semibold"><T>{reasonView.elonTitle}</T></p>
            <p className="text-xs text-[color:var(--text-muted)]">
              <T>{reasonView.cancelledBy === "worker" ? "Ishchi tomonidan bekor qilingan" : "Ish beruvchi tomonidan bekor qilingan"}</T>
            </p>
            <p className="text-sm whitespace-pre-line"><T>{reasonView.cancelReason || ""}</T></p>
          </div>
        )}
      </Modal>
    </Shell>
  );
}
