"use client";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { useQuery, useMutation } from "@tanstack/react-query";
import {
  Info, Users, Calendar, MapPin, FileText, Wallet, ShieldCheck,
  Phone, Send, Share2, UserRound,
} from "lucide-react";
import { api, Elon, getAccess } from "@/lib/api";
import { Modal } from "@/components/Modal";
import { StatusBadge } from "@/components/StatusBadge";
import { Shell } from "@/components/Shell";
import { ScriptToggle } from "@/components/ScriptToggle";
import { ThemeToggle } from "@/components/ThemeToggle";
import { T, useT } from "@/components/T";
import { fmtSumSom } from "@/lib/format";
import dayjs from "dayjs";

export default function ElonDetails() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const t = useT();
  const [open, setOpen] = useState(false);
  const [phone, setPhone] = useState("");
  const [status, setStatus] = useState<"none" | "pending" | "accepted">("none");
  const [me, setMe] = useState<any>(null);
  const [authed, setAuthed] = useState(false);

  useEffect(() => {
    const has = !!getAccess();
    setAuthed(has);
    if (has) {
      api.get<any>("/api/me").then((u) => { setMe(u); setPhone(u.phone || ""); }).catch(() => {});
      api.get<any[]>("/api/my/applications").then((apps) => {
        const mine = apps.find((a) => a.elonId === id);
        if (mine) setStatus(mine.status === "accepted" ? "accepted" : mine.status === "pending" ? "pending" : "none");
      }).catch(() => {});
    }
  }, [id]);

  const { data: e } = useQuery<Elon>({
    queryKey: ["elon", id],
    queryFn: () => api.get<Elon>(`/api/elons/${id}`, { auth: "none" } as any),
    enabled: !!id,
  });

  const apply = useMutation({
    mutationFn: () => api.post(`/api/elons/${id}/apply`, { phone }),
    onSuccess: () => { setOpen(false); setStatus("pending"); },
  });

  if (!e) {
    return (
      <div className="min-h-screen grid place-items-center">
        <div className="muted text-sm"><T>Yuklanmoqda…</T></div>
      </div>
    );
  }
  const isOwner = me && me.id === e.ownerId;
  const dateLine = e.startDate
    ? `${dayjs(e.startDate).format("D-MMM")}${e.workTimeFrom ? `, ${e.workTimeFrom}` : ""}${e.workTimeTo ? ` - ${e.workTimeTo}` : ""}`
    : "—";
  const mapImg = e.locationUrl
    ? `https://staticmap.openstreetmap.de/staticmap.php?center=41.2995,69.2401&zoom=13&size=300x200&markers=41.2995,69.2401,red-circle`
    : "";

  /* ── inner content (shared between auth/anon) ── */
  const content = (
    <div className="grid grid-cols-1 lg:grid-cols-[1fr_360px] gap-4">
      {/* ── Main column ─────────────── */}
      <div className="grid gap-4">
        {/* Ish ma'lumotlari */}
        <section className="card p-5 relative overflow-hidden">
          <span className="absolute left-0 top-0 bottom-0 w-1 bg-brand-navy rounded-l-2xl" />
          <h2 className="font-semibold heading flex items-center gap-2 mb-4">
            <Info size={18} /><T>Ish ma'lumotlari</T>
          </h2>
          <div className="grid sm:grid-cols-2 gap-4">
            <KV icon={<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="3" /><polygon points="12,2 22,20 2,20" /></svg>} label="KATEGORIYA" value={e.categoryName} />
            <KV icon={<Users size={18} />} label="ISHCHILAR SONI" value={`${e.workersNeeded} kishi`} />
            <KV icon={<Calendar size={18} />} label="SANA VA VAQT" value={dateLine} />
          </div>
        </section>

        {/* Manzil */}
        <section className="card p-5">
          <h2 className="font-semibold heading flex items-center gap-2 mb-4">
            <MapPin size={18} /><T>Manzil</T>
          </h2>
          <div className="grid sm:grid-cols-[180px_1fr] gap-4">
            {e.locationUrl ? (
              <a href={e.locationUrl} target="_blank" rel="noreferrer" className="block rounded-xl overflow-hidden border h-[120px] bg-[color:var(--bg)]" style={{ borderColor: "var(--border)" }}>
                {mapImg && (<img src={mapImg} alt="map" className="w-full h-full object-cover" />)}
                {!mapImg && (<div className="w-full h-full grid place-items-center muted text-sm"><T>Xaritani ochish</T></div>)}
              </a>
            ) : (
              <div className="rounded-xl border h-[120px] grid place-items-center muted text-sm" style={{ borderColor: "var(--border)" }}>
                <MapPin size={24} />
              </div>
            )}
            <div className="flex flex-col justify-center">
              <div className="font-semibold"><T>{e.locationText || e.region || "Manzil ko'rsatilmagan"}</T></div>
              {e.district && <div className="text-sm muted"><T>{e.district}</T></div>}
            </div>
          </div>
        </section>

        {/* Batafsil ma'lumot */}
        <section className="card p-5">
          <h2 className="font-semibold heading flex items-center gap-2 mb-3">
            <FileText size={18} /><T>Batafsil ma'lumot</T>
          </h2>
          <p className="text-sm leading-relaxed whitespace-pre-line muted">
            <T>{e.description}</T>
          </p>
        </section>
      </div>

      {/* ── Right column ────────────── */}
      <aside className="grid gap-4 content-start">
        {/* To'lov ma'lumotlari */}
        <section
          className="card p-5 relative overflow-hidden"
          style={{ background: "linear-gradient(180deg, rgba(232,146,10,0.06) 0%, var(--card) 60%)" }}
        >
          <h3 className="font-semibold heading flex items-center gap-2 mb-3">
            <Wallet size={18} className="text-accent-amber" /><T>To'lov ma'lumotlari</T>
          </h3>
          <div className="space-y-2 text-sm">
            <Row label="Kishi boshiga" value={fmtSumSom(e.perWorkerAmount, e.pricingType === "negotiable")} />
            <Row label="Ishchilar soni" value={`${e.workersNeeded} ta`} />
          </div>
          <div className="mt-4 rounded-xl p-3" style={{ background: "rgba(15,31,86,0.05)" }}>
            <div className="text-[11px] uppercase tracking-wider muted"><T>JAMI SUMMA</T>:</div>
            <div className="font-extrabold text-xl heading mt-1">
              {fmtSumSom(e.priceAmount, e.pricingType === "negotiable")}
            </div>
          </div>
        </section>

        {/* Owner card */}
        <section className="card p-5 text-center">
          <div className="mx-auto h-16 w-16 rounded-full grid place-items-center" style={{ background: "rgba(34,158,217,0.12)" }}>
            <UserRound size={28} className="text-tg-blue" />
          </div>
          <div className="mt-3 font-semibold heading">{e.ownerName || "Foydalanuvchi"}</div>
          <div className="mt-1 text-xs inline-flex items-center gap-1 muted">
            <ShieldCheck size={12} className="text-success" /><T>Tasdiqlangan buyurtmachi</T>
          </div>
          {e.contactPhone && (
            <a
              href={`tel:${e.contactPhone}`}
              className="mt-3 inline-flex w-full items-center justify-center gap-2 rounded-lg px-4 py-2.5 text-sm font-medium border"
              style={{ borderColor: "var(--border)", background: "rgba(127,127,127,0.05)" }}
            >
              <Phone size={14} className="muted" />{e.contactPhone}
            </a>
          )}
        </section>

        {/* Actions */}
        {!isOwner && (
          <>
            {status === "none" && (
              <button onClick={() => authed ? setOpen(true) : router.push("/login")} className="btn-primary w-full py-3 gap-2">
                <Send size={16} /><T>Ariza topshirish</T>
              </button>
            )}
            {status === "pending" && (
              <button disabled className="w-full rounded-lg py-3 bg-pending-bg text-pending font-medium"><T>Kutilmoqda</T></button>
            )}
            {status === "accepted" && (
              <button disabled className="w-full rounded-lg py-3 bg-success-bg text-success font-medium"><T>Ish qabul qilindi</T></button>
            )}
          </>
        )}
        <button
          onClick={() => navigator.share?.({ url: location.href, title: e.title }).catch(() => {})}
          className="w-full rounded-lg py-3 bg-accent-amber text-white font-medium inline-flex items-center justify-center gap-2 hover:opacity-90"
        >
          <Share2 size={16} /><T>Ulashish</T>
        </button>
      </aside>

      <Modal open={open} onClose={() => setOpen(false)} title={t("Ariza topshirishni tasdiqlaysizmi?")} footer={
        <>
          <button onClick={() => setOpen(false)} className="btn-secondary"><T>Bekor qilish</T></button>
          <button onClick={() => apply.mutate()} disabled={apply.isPending} className="btn-primary"><T>Tasdiqlash</T></button>
        </>
      }>
        <p className="text-sm muted mb-3"><T>{e.title}</T> — {fmtSumSom(e.perWorkerAmount, e.pricingType === "negotiable")} / <T>kishi boshiga</T></p>
        <label className="block">
          <span className="text-sm font-medium"><T>TELEFON RAQAMINGIZ</T></span>
          <input className="input mt-1" value={phone} onChange={(ev) => setPhone(ev.target.value)} placeholder="+998 ..." />
        </label>
      </Modal>
    </div>
  );

  /* ── Layouts: cabinet (auth) or public (anon) ── */
  if (authed) {
    return (
      <Shell title={e.title}>
        <div className="card p-4 flex items-center justify-between animate-fade-in">
          <h1 className="text-xl sm:text-2xl font-bold heading leading-tight"><T>{e.title}</T></h1>
          <StatusBadge status={e.status} />
        </div>
        {content}
      </Shell>
    );
  }

  // Public (anonymous) view
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b" style={{ borderColor: "var(--border)", background: "var(--card)" }}>
        <div className="mx-auto max-w-7xl flex items-center justify-between px-4 py-3">
          <Link href="/" className="heading font-extrabold text-xl">Ishchi Bormi</Link>
          <div className="flex items-center gap-3">
            <ScriptToggle />
            <ThemeToggle />
            <Link href="/login" className="btn-primary"><T>Kirish</T></Link>
          </div>
        </div>
      </header>
      <main className="flex-1 mx-auto max-w-7xl w-full px-4 py-6 grid gap-4">
        <div className="card p-4 flex items-center justify-between">
          <h1 className="text-xl sm:text-2xl font-bold heading"><T>{e.title}</T></h1>
          <StatusBadge status={e.status} />
        </div>
        {content}
      </main>
    </div>
  );
}

function KV({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-start gap-3">
      <div className="shrink-0 grid h-9 w-9 place-items-center rounded-full muted" style={{ background: "rgba(127,127,127,0.08)" }}>
        {icon}
      </div>
      <div className="min-w-0">
        <div className="text-[11px] uppercase tracking-wider muted"><T>{label}</T></div>
        <div className="font-semibold"><T>{value}</T></div>
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between border-b py-2" style={{ borderColor: "var(--border)" }}>
      <span className="muted"><T>{label}</T>:</span>
      <span className="font-semibold">{value}</span>
    </div>
  );
}

