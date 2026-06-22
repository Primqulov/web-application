"use client";
import { useState } from "react";
import Link from "next/link";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Pencil, Trash2, Plus, FileText, Send, Archive, ListChecks } from "lucide-react";
import { api, Elon } from "@/lib/api";
import { Shell, ShellSearch } from "@/components/Shell";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Tabs } from "@/components/ui/Tabs";
import { EmptyState } from "@/components/ui/EmptyState";
import { CardSkeleton } from "@/components/ui/Skeleton";
import { StatusBadge } from "@/components/StatusBadge";
import { T, useT } from "@/components/T";
import { fmtSumSom, fromNow } from "@/lib/format";

type MyElons = { drafts: Elon[]; active: Elon[]; archived: Elon[] };

export default function MyElons() {
  const t = useT();
  const qc = useQueryClient();
  const [tab, setTab] = useState<"drafts" | "active" | "archived">("active");
  const [q, setQ] = useState("");

  const { data, isLoading } = useQuery<MyElons>({
    queryKey: ["my-elons"],
    queryFn: () => api.get<MyElons>("/api/my/elons"),
  });

  const publish = useMutation({
    mutationFn: (id: string) => api.post(`/api/elons/${id}/publish`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["my-elons"] }),
  });
  const del = useMutation({
    mutationFn: (id: string) => api.delete(`/api/elons/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["my-elons"] }),
  });

  const counts = { drafts: data?.drafts.length ?? 0, active: data?.active.length ?? 0, archived: data?.archived.length ?? 0 };
  const list = (data?.[tab] || []).filter((e) => !q || e.title.toLowerCase().includes(q.toLowerCase()));
  const search = <ShellSearch value={q} onChange={setQ} placeholder={t("E'lon nomi bo'yicha qidirish…")} />;

  return (
    <Shell title="Mening e'lonlarim" search={search}>
      {/* Stat row */}
      <div className="grid sm:grid-cols-3 gap-3">
        <StatCard label="Faol e'lonlar" value={counts.active} icon={<ListChecks size={18} />} />
        <StatCard label="Qoralama" value={counts.drafts} icon={<FileText size={18} />} />
        <StatCard label="Arxivlangan" value={counts.archived} icon={<Archive size={18} />} />
      </div>

      {/* Tabs + CTA */}
      <div className="flex flex-col sm:flex-row gap-3 sm:items-center sm:justify-between">
        <Tabs
          value={tab}
          onChange={(v) => setTab(v as any)}
          items={[
            { value: "active",   label: t("Faol"),    count: counts.active },
            { value: "drafts",   label: t("Qoralama"), count: counts.drafts },
            { value: "archived", label: t("Arxiv"),    count: counts.archived },
          ]}
        />
        <Link href="/elon/create" className="btn btn-primary gap-2"><Plus size={16} /><T>Yangi e'lon</T></Link>
      </div>

      {/* List */}
      {isLoading ? (
        <div className="grid sm:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : list.length === 0 ? (
        <EmptyState
          icon={<ListChecks size={22} />}
          title={t("Hozircha e'lon yo'q")}
          body={
            tab === "drafts" ? t("Qoralama e'lonlaringiz shu yerda saqlanadi.")
            : tab === "archived" ? t("Yakunlangan yoki bekor qilingan e'lonlar shu yerda ko'rinadi.")
            : t("Birinchi e'loningizni yarating va arizalarni qabul qila boshlang.")
          }
          action={<Link href="/elon/create" className="btn btn-primary"><T>E'lon yaratish</T></Link>}
        />
      ) : (
        <div className="grid sm:grid-cols-2 gap-4">
          {list.map((e) => (
            <Card key={e.id} hover className="animate-fade-in">
              <div className="flex items-start gap-2">
                <Link href={`/elon/${e.id}`} className={`font-semibold heading flex-1 leading-tight line-clamp-2 hover:underline ${tab === "archived" ? "line-through opacity-70" : ""}`}>
                  <T>{e.title}</T>
                </Link>
                <div className="flex gap-1 shrink-0">
                  <Link href={`/elon/${e.id}/edit`} className="p-1.5 rounded-md hover:bg-[color:var(--bg-subtle)]" aria-label="Edit">
                    <Pencil size={14} className="muted" />
                  </Link>
                  <button
                    onClick={() => { if (confirm(t("Haqiqatan o'chirishni xohlaysizmi?"))) del.mutate(e.id); }}
                    className="p-1.5 rounded-md hover:bg-danger-bg text-danger" aria-label="Delete"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>

              <div className="mt-3 flex flex-wrap items-center gap-2">
                <StatusBadge status={e.status} />
                <span className="text-xs muted"><T>{e.categoryName}</T></span>
                {e.publishedAt && <span className="text-xs muted">· {fromNow(e.publishedAt)}</span>}
              </div>

              <div className="mt-4 pt-3 border-t flex items-end justify-between" style={{ borderColor: "var(--border)" }}>
                <div>
                  <div className="text-[11px] uppercase muted"><T>Narxi</T></div>
                  <div className="font-bold text-accent-amber">
                    {fmtSumSom(e.perWorkerAmount || e.priceAmount, e.pricingType === "negotiable")}
                  </div>
                </div>
                {tab === "drafts" && (
                  <Button size="sm" leftIcon={<Send size={13} />} onClick={() => publish.mutate(e.id)} loading={publish.isPending}>
                    <T>Joylashtirish</T>
                  </Button>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}
    </Shell>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <Card>
      <div className="flex items-center justify-between">
        <div>
          <div className="text-xs uppercase muted tracking-wide"><T>{label}</T></div>
          <div className="text-2xl font-bold heading mt-1">{value}</div>
        </div>
        <div className="h-10 w-10 grid place-items-center rounded-xl" style={{ background: "var(--brand-soft)", color: "var(--brand)" }}>
          {icon}
        </div>
      </div>
    </Card>
  );
}
