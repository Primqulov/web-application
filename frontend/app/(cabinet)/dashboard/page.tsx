"use client";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { Star, MapPin, Clock, Share2, ArrowUpRight, Briefcase } from "lucide-react";
import { api, Category, Elon } from "@/lib/api";
import { Shell, ShellSearch } from "@/components/Shell";
import { EmptyState } from "@/components/ui/EmptyState";
import { CardSkeleton } from "@/components/ui/Skeleton";
import { Select } from "@/components/ui/Input";
import { Avatar } from "@/components/ui/Avatar";
import { T, useT } from "@/components/T";
import { fmtSumSom, fromNow } from "@/lib/format";

export default function Dashboard() {
  const t = useT();
  const [q, setQ] = useState("");
  const [cat, setCat] = useState<string>("");
  const [sort, setSort] = useState<string>("time");

  const { data: cats } = useQuery<Category[]>({
    queryKey: ["categories"],
    queryFn: () => api.get<Category[]>("/api/categories"),
  });
  const { data, isLoading } = useQuery<{ items: Elon[] }>({
    queryKey: ["feed", q, cat, sort],
    queryFn: () => {
      const p = new URLSearchParams();
      if (q) p.set("q", q);
      if (cat) p.set("categoryId", cat);
      if (sort) p.set("sort", sort);
      return api.get<{ items: Elon[] }>(`/api/elons?${p.toString()}`);
    },
  });

  const items = data?.items || [];
  const search = <ShellSearch value={q} onChange={setQ} placeholder={t("Xizmatlar yoki ishchilarni qidiring…")} />;

  return (
    <Shell title="Bosh sahifa" search={search}>
      {/* Categories */}
      <div className="card p-3 flex items-center gap-2 overflow-x-auto scroll-y-auto">
        <Chip active={cat === ""} onClick={() => setCat("")}><T>Barchasi</T></Chip>
        {(cats || []).slice(0, 12).map((c) => (
          <Chip key={c.id} active={cat === c.id} onClick={() => setCat(c.id)}>
            {c.icon && <span>{c.icon}</span>}<T>{c.name}</T>
          </Chip>
        ))}
      </div>

      {/* Header strip */}
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold heading"><T>Ish e'lonlari</T></h2>
          <p className="text-xs muted mt-0.5">
            <T>Topildi</T>: <b className="heading">{items.length}</b> <T>ta e'lon</T>
          </p>
        </div>
        <div className="w-44">
          <Select value={sort} onChange={(e) => setSort(e.target.value)}>
            <option value="time">{t("Eng yangilari")}</option>
            <option value="price">{t("Yuqori narx")}</option>
            <option value="rating">{t("Yuqori reyting")}</option>
          </Select>
        </div>
      </div>

      {/* Grid */}
      {isLoading ? (
        <div className="grid sm:grid-cols-2 xl:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : items.length === 0 ? (
        <EmptyState
          icon={<Briefcase size={22} />}
          title={t("Hozircha e'lonlar yo'q")}
          body={t("Filtrlarni o'zgartirib qaytadan urinib ko'ring yoki o'zingiz birinchi e'lonni joylashtiring.")}
          action={<Link href="/elon/create" className="btn btn-primary"><T>E'lon yaratish</T></Link>}
        />
      ) : (
        <div className="grid sm:grid-cols-2 xl:grid-cols-3 gap-4">
          {items.map((e) => <JobCard key={e.id} e={e} />)}
        </div>
      )}
    </Shell>
  );
}

function Chip({ active, onClick, children }: { active?: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button onClick={onClick} className={`chip shrink-0 ${active ? "chip-active" : ""}`}>{children}</button>
  );
}

function JobCard({ e }: { e: Elon }) {
  const neg = e.pricingType === "negotiable";
  return (
    <Link href={`/elon/${e.id}`} className="card p-5 block transition hover:-translate-y-0.5 hover:shadow-pop animate-fade-in">
      <div className="flex items-start gap-3">
        <Avatar name={e.ownerName} size="md" />
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <h3 className="font-semibold heading leading-tight line-clamp-1"><T>{e.title}</T></h3>
            <button
              onClick={(ev) => { ev.preventDefault(); navigator.share?.({ url: location.origin + `/elon/${e.id}`, title: e.title }).catch(() => {}); }}
              className="muted hover:text-accent-amber transition"
            >
              <Share2 size={16} />
            </button>
          </div>
          <div className="mt-0.5 flex items-center gap-2 text-xs muted">
            <span className="truncate">{e.ownerName || "—"}</span>
            <span>·</span>
            <span className="inline-flex items-center gap-1">
              <Star size={11} className="text-accent-amber" fill="currentColor" />
              {(e.ownerRating ?? 0).toFixed(1)}
            </span>
          </div>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap gap-2 text-xs muted">
        <span className="inline-flex items-center gap-1 surface px-2 py-1">
          <MapPin size={12} />{e.locationText || e.region || "—"}
        </span>
        {e.publishedAt && (
          <span className="inline-flex items-center gap-1 surface px-2 py-1">
            <Clock size={12} />{fromNow(e.publishedAt)}
          </span>
        )}
        <span className="badge-neutral"><T>{e.categoryName}</T></span>
      </div>

      <div className="mt-4 pt-3 border-t flex items-end justify-between" style={{ borderColor: "var(--border)" }}>
        <div>
          <div className="text-[11px] uppercase tracking-wide muted"><T>Narxi</T></div>
          <div className="text-lg font-bold text-accent-amber leading-none mt-0.5">
            {fmtSumSom(e.perWorkerAmount || e.priceAmount, neg)}
          </div>
        </div>
        <span className="inline-flex items-center gap-1 text-xs font-medium heading group">
          <T>Ko'rish</T><ArrowUpRight size={14} className="transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
        </span>
      </div>
    </Link>
  );
}
