"use client";
import { useMemo, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, Application, Notification, Elon } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { StatusBadge } from "@/components/StatusBadge";
import { SlotProgress } from "@/components/ui/SlotProgress";
import { Modal } from "@/components/Modal";
import { Phone, MapPin, ChevronDown, ExternalLink } from "lucide-react";
import { T, useT } from "@/components/T";
import Link from "next/link";

// Bekor qilish sababini ro'yxatda qisqa ko'rsatish uchun.
function shortReason(s: string, max = 60) {
  return s.length > max ? s.slice(0, max).trimEnd() + "…" : s;
}

export default function Process() {
  const [tab, setTab] = useState<"worker" | "employer">("worker");
  const [cancelId, setCancelId] = useState("");
  const [cancelReason, setCancelReason] = useState("");
  const [openElons, setOpenElons] = useState<Record<string, boolean>>({});
  const [reasonView, setReasonView] = useState<Application | null>(null);
  const [errMsg, setErrMsg] = useState("");
  const t = useT();
  const qc = useQueryClient();

  const { data: mine } = useQuery<Application[]>({
    queryKey: ["my-applications"],
    queryFn: () => api.get<Application[]>("/api/my/applications"),
  });
  const { data: received } = useQuery<Record<string, Application[]>>({
    queryKey: ["my-elons-applications"],
    queryFn: () => api.get<Record<string, Application[]>>("/api/my/elons/applications"),
  });
  const { data: notifs } = useQuery<Notification[]>({
    queryKey: ["notifications"],
    queryFn: () => api.get<Notification[]>("/api/notifications"),
  });
  // Har bir e'lonning to'lish holatini (acceptedCount / workersNeeded) ko'rsatish
  // uchun ish beruvchining e'lonlarini olamiz.
  const { data: myElons } = useQuery<{ active: Elon[]; archived: Elon[] }>({
    queryKey: ["my-elons"],
    queryFn: () => api.get<{ active: Elon[]; archived: Elon[] }>("/api/my/elons"),
  });
  const elonById = useMemo(() => {
    const m: Record<string, Elon> = {};
    [...(myElons?.active || []), ...(myElons?.archived || [])].forEach((e) => { m[e.id] = e; });
    return m;
  }, [myElons]);

  // Qizil nuqtalar uchun: ariza bilan bog'liq o'qilmagan bildirishnomalardagi
  // ariza id'lari. Karta/tab shu id'lar bilan solishtiriladi.
  const unreadAppIds = new Set(
    (notifs || []).filter((n) => !n.isRead && n.relatedEntity?.type === "application").map((n) => n.relatedEntity!.id)
  );
  const seen = useMutation({
    mutationFn: (ids: string[]) => api.post("/api/notifications/read", { relatedIds: ids }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notifications"] }),
  });
  // Foydalanuvchi arizani ko'rgach/harakat qilgach shu arizaning nuqtasini tozalaymiz.
  function markSeen(...ids: string[]) {
    const fresh = ids.filter((id) => unreadAppIds.has(id));
    if (fresh.length) seen.mutate(fresh);
  }

  const workerDot = (mine || []).some((a) => unreadAppIds.has(a.id));
  const employerDot = Object.values(received || {}).some((apps) => apps.some((a) => unreadAppIds.has(a.id)));

  function refreshLists() {
    qc.invalidateQueries({ queryKey: ["my-elons-applications"] });
    qc.invalidateQueries({ queryKey: ["my-applications"] });
    qc.invalidateQueries({ queryKey: ["my-elons"] });
  }
  const accept = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/accept`),
    onSuccess: refreshLists,
    onError: (e: any) => setErrMsg(e?.message || "Xatolik yuz berdi"),
  });
  const reject = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/reject`),
    onSuccess: refreshLists,
  });
  const cancel = useMutation({
    mutationFn: () => api.post(`/api/applications/${cancelId}/cancel`, { reason: cancelReason.trim() }),
    onSuccess: () => { setCancelId(""); setCancelReason(""); refreshLists(); },
  });
  const done = useMutation({
    mutationFn: (id: string) => api.post(`/api/applications/${id}/confirm-done`),
    onSuccess: refreshLists,
  });

  return (
    <Shell title="Jarayonlar">
      <div className="card p-2 flex gap-2">
        <button onClick={() => setTab("worker")} className={`flex-1 rounded-lg px-3 py-2 text-sm ${tab === "worker" ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}>
          <span className="inline-flex items-center gap-1.5"><T>Ishlardagi jarayon</T>{workerDot && <span className="h-2 w-2 rounded-full bg-danger" />}</span>
        </button>
        <button onClick={() => setTab("employer")} className={`flex-1 rounded-lg px-3 py-2 text-sm ${tab === "employer" ? "bg-brand-navy text-white" : "hover:bg-black/5"}`}>
          <span className="inline-flex items-center gap-1.5"><T>E'lonlardagi jarayon</T>{employerDot && <span className="h-2 w-2 rounded-full bg-danger" />}</span>
        </button>
      </div>

      {tab === "worker" && (
        <div className="grid sm:grid-cols-2 gap-4">
          {(mine || []).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)] sm:col-span-2"><T>Sizda ariza yo'q.</T></div>}
          {(mine || []).map((a) => (
            <div key={a.id} className="card p-4">
              <div className="flex items-center justify-between">
                <Link href={`/elon/${a.elonId}`} onClick={() => markSeen(a.id)} className="font-semibold inline-flex items-center gap-2">
                  {unreadAppIds.has(a.id) && <span className="h-2.5 w-2.5 rounded-full bg-danger shrink-0" />}
                  <T>{a.elonTitle}</T>
                </Link>
                <StatusBadge status={a.status} />
              </div>
              <div className="text-xs text-[color:var(--text-muted)] mt-1 flex flex-wrap gap-x-3 gap-y-0.5">
                {a.ownerName && <span><T>Ish beruvchi</T>: <span className="font-medium text-[color:var(--text)]">{a.ownerName}</span></span>}
                <span><T>Ariza</T>: {a.peopleCount || 1} <T>kishi</T></span>
              </div>
              {a.status === "pending" && <p className="text-sm text-[color:var(--text-muted)] mt-2"><T>Ariza ko'rib chiqilmoqda…</T></p>}
              {a.status === "accepted" && (
                <div className="mt-2 grid gap-2">
                  <p className="text-sm text-success"><T>Ish beruvchi tomonidan qabul qilindi</T></p>
                  <div className="flex flex-wrap gap-2">
                    <a href={`tel:${a.workerPhone}`} className="btn-secondary gap-2"><Phone size={14} /><T>Qo'ng'iroq qilish</T></a>
                    <Link href={`/elon/${a.elonId}`} className="btn-secondary gap-2"><MapPin size={14} /><T>Manzilni ko'rish</T></Link>
                    <button onClick={() => done.mutate(a.id)} className="btn-primary"><T>Bajarildi</T></button>
                    <button onClick={() => { setCancelReason(""); setCancelId(a.id); }} className="btn-danger"><T>Bekor qilish</T></button>
                  </div>
                </div>
              )}
              {a.status === "cancelled" && a.cancelReason && (
                <div className="mt-2 text-sm">
                  <span className="text-[color:var(--text-muted)]">
                    <T>{a.cancelledBy === "worker" ? "Ishchi bekor qildi" : "Ish beruvchi bekor qildi"}</T> — <T>{shortReason(a.cancelReason)}</T>
                  </span>
                  {a.cancelReason.length > 60 && (
                    <button onClick={() => setReasonView(a)} className="ml-1.5 text-tg-blue underline text-xs"><T>Batafsil</T></button>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {tab === "employer" && (
        <div className="grid gap-4">
          {Object.keys(received || {}).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)]"><T>Hozircha arizalar yo'q.</T></div>}
          {Object.entries(received || {}).map(([elonId, apps]) => {
            const open = !!openElons[elonId];
            return (
            <div key={elonId} className="card p-4">
              {/* Sarlavha bosilganda sahifa ochilmaydi — pastidan arizachilar ro'yxati ochiladi. */}
              <button
                type="button"
                onClick={() => {
                  setOpenElons((s) => ({ ...s, [elonId]: !s[elonId] }));
                  if (!open) markSeen(...apps.map((a) => a.id));
                }}
                className="w-full flex items-center justify-between gap-2 text-left"
              >
                <span className="font-semibold inline-flex items-center gap-2 min-w-0">
                  {apps.some((a) => unreadAppIds.has(a.id)) && <span className="h-2.5 w-2.5 rounded-full bg-danger shrink-0" />}
                  <span className="truncate"><T>{apps[0]?.elonTitle || "E'lon"}</T></span>
                </span>
                <span className="flex items-center gap-2 shrink-0">
                  <span className="badge-amber">{apps.length} <T>ta ariza</T></span>
                  <ChevronDown size={18} className={`transition-transform ${open ? "rotate-180" : ""}`} />
                </span>
              </button>
              {elonById[elonId] && (
                <div className="mt-3">
                  <SlotProgress accepted={elonById[elonId].acceptedCount} needed={elonById[elonId].workersNeeded} />
                </div>
              )}
              {open && (
              <div className="grid gap-2 mt-3">
                <Link href={`/elon/${elonId}`} className="text-xs text-tg-blue inline-flex items-center gap-1 w-fit">
                  <ExternalLink size={12} /><T>E'lonni ochish</T>
                </Link>
                {apps.map((a) => (
                  <div key={a.id} className="flex flex-wrap items-center gap-2 border-t pt-2" style={{ borderColor: "var(--border)" }}>
                    <div className="grid h-8 w-8 place-items-center rounded-full bg-brand-navy text-white text-xs uppercase">{(a.workerName?.trim()?.[0]) || a.workerPhone?.slice(-2) || "?"}</div>
                    <div className="mr-auto min-w-0">
                      <div className="font-medium text-sm inline-flex items-center gap-1.5">
                        {unreadAppIds.has(a.id) && <span className="h-2 w-2 rounded-full bg-danger shrink-0" />}
                        <span className="truncate">{a.workerName?.trim() || a.workerPhone}</span>
                      </div>
                      <div className="text-xs text-[color:var(--text-muted)]">{a.workerPhone}</div>
                      {a.status === "cancelled" && a.cancelReason && (
                        <div className="text-xs text-[color:var(--text-muted)] mt-0.5">
                          <T>{a.cancelledBy === "worker" ? "Ishchi bekor qildi" : "Siz bekor qildingiz"}</T> — <T>{shortReason(a.cancelReason, 40)}</T>
                          {a.cancelReason.length > 40 && (
                            <button onClick={() => setReasonView(a)} className="ml-1 text-tg-blue underline"><T>Batafsil</T></button>
                          )}
                        </div>
                      )}
                    </div>
                    <span className="badge-amber">{a.peopleCount || 1} <T>kishi</T></span>
                    <StatusBadge status={a.status} />
                    <a href={`tel:${a.workerPhone}`} className="btn-secondary gap-1"><Phone size={12} /><T>Qo'ng'iroq</T></a>
                    {a.status === "pending" && <>
                      <button onClick={() => { markSeen(a.id); accept.mutate(a.id); }} className="btn-primary"><T>Qabul qilish</T></button>
                      <button onClick={() => { markSeen(a.id); reject.mutate(a.id); }} className="btn-danger">×</button>
                    </>}
                    {a.status === "accepted" && <>
                      <button onClick={() => { markSeen(a.id); done.mutate(a.id); }} className="btn-primary"><T>Bajarildi</T></button>
                      <button onClick={() => { markSeen(a.id); setCancelReason(""); setCancelId(a.id); }} className="btn-danger"><T>Bekor qilish</T></button>
                    </>}
                  </div>
                ))}
              </div>
              )}
            </div>
            );
          })}
        </div>
      )}

      <Modal open={!!cancelId} onClose={() => setCancelId("")} title={t("Ishni bekor qilasizmi?")} footer={
        <>
          <button onClick={() => setCancelId("")} className="btn-secondary"><T>Yo'q</T></button>
          <button onClick={() => cancel.mutate()} disabled={cancel.isPending || !cancelReason.trim()} className="btn-danger disabled:opacity-50"><T>Ha, bekor qilish</T></button>
        </>
      }>
        <p className="text-sm muted mb-3"><T>Ushbu ishni bekor qilasiz. Keyinroq qayta ariza topshirishingiz mumkin.</T></p>
        <label className="block">
          <span className="text-sm font-medium"><T>BEKOR QILISH SABABI</T> <span className="text-danger">*</span></span>
          <textarea className="input mt-1" rows={3} value={cancelReason} onChange={(ev) => setCancelReason(ev.target.value)} placeholder={t("Masalan: rejalarim o'zgardi")} />
          {!cancelReason.trim() && <span className="text-xs text-danger mt-1 block"><T>Sababni yozmasangiz bekor qila olmaysiz.</T></span>}
        </label>
      </Modal>

      {/* Xato/ogohlantirish — modal ko'rinishida */}
      <Modal open={!!errMsg} onClose={() => setErrMsg("")} title={t("Ogohlantirish")} footer={
        <button onClick={() => setErrMsg("")} className="btn-primary"><T>Tushunarli</T></button>
      }>
        <p className="text-sm"><T>{errMsg}</T></p>
      </Modal>

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
