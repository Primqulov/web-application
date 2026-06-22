"use client";
import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api, Category, Elon } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { MultiImageUploader } from "@/components/ui/ImageUpload";
import { CheckCircle2, AlertTriangle, Plus } from "lucide-react";
import { T, useT } from "@/components/T";
import { fmtSum } from "@/lib/format";

type Form = {
  title: string;
  categoryId: string;
  description: string;
  locationUrl: string;
  locationText: string;
  region: string;
  district: string;
  workersNeeded: number;
  pricingType: "per_worker" | "total" | "";
  priceAmount: number | "";
  startDate: string;
  workTimeFrom: string;
  workTimeTo: string;
  contactPhone: string;
  images: string[];
};

export default function CreateElon() {
  const router = useRouter();
  const t = useT();
  const [form, setForm] = useState<Form>({
    title: "",
    categoryId: "",
    description: "",
    locationUrl: "",
    locationText: "",
    region: "",
    district: "",
    workersNeeded: 1,
    pricingType: "per_worker",
    priceAmount: "",
    startDate: "",
    workTimeFrom: "",
    workTimeTo: "",
    contactPhone: "",
    images: [],
  });
  const [state, setState] = useState<"idle" | "ok" | "err">("idle");
  const [newCat, setNewCat] = useState("");

  const { data: cats, refetch: refetchCats } = useQuery<Category[]>({
    queryKey: ["categories"],
    queryFn: () => api.get<Category[]>("/api/categories"),
  });

  const create = useMutation({
    mutationFn: async (publish: boolean) => {
      const body = { ...form, priceAmount: typeof form.priceAmount === "number" ? form.priceAmount : 0,
        pricingType: form.priceAmount === "" ? "negotiable" : form.pricingType };
      const created = await api.post<Elon>("/api/elons", body);
      if (publish) await api.post(`/api/elons/${created.id}/publish`);
      return created;
    },
    onSuccess: () => setState("ok"),
    onError: () => setState("err"),
  });

  const createCat = useMutation({
    mutationFn: (name: string) => api.post<Category>("/api/categories", { name }),
    onSuccess: (c) => { setForm((f) => ({ ...f, categoryId: c.id })); setNewCat(""); refetchCats(); },
  });

  const live = useMemo(() => {
    const n = form.workersNeeded;
    const p = typeof form.priceAmount === "number" ? form.priceAmount : 0;
    if (!p) return { per: 0, total: 0, neg: true };
    if (form.pricingType === "total") return { per: n ? Math.floor(p / n) : 0, total: p, neg: false };
    return { per: p, total: p * n, neg: false };
  }, [form.workersNeeded, form.priceAmount, form.pricingType]);

  useEffect(() => { if (cats?.length && !form.categoryId) setForm((f) => ({ ...f, categoryId: cats[0].id })); }, [cats]);

  if (state === "ok") return <SuccessPage onMine={() => router.push("/my-elons")} onHome={() => router.push("/dashboard")} />;
  if (state === "err") return <ErrorPage onRetry={() => setState("idle")} onSupport={() => router.push("/settings")} />;

  return (
    <Shell title="Yangi e'lon yaratish">
      <form
        onSubmit={(e) => { e.preventDefault(); create.mutate(true); }}
        className="card p-6 grid gap-4 max-w-3xl"
      >
        <h2 className="font-semibold text-lg"><T>E'lon ma'lumotlari</T></h2>

        <Field label="Ish nomi *">
          <input className="input" required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder={t("Masalan, Mebel tashish")} />
        </Field>

        <Field label="Turkum *">
          <div className="flex flex-wrap gap-2">
            {(cats || []).map((c) => (
              <button key={c.id} type="button"
                onClick={() => setForm({ ...form, categoryId: c.id })}
                className={`chip ${form.categoryId === c.id ? "chip-active" : ""}`}>
                {c.icon}<T>{c.name}</T>
              </button>
            ))}
          </div>
          <div className="mt-2 flex gap-2">
            <input className="input" value={newCat} onChange={(e) => setNewCat(e.target.value)} placeholder={t("Yangi turkum nomi")} />
            <button type="button" disabled={!newCat.trim()} onClick={() => createCat.mutate(newCat.trim())} className="btn-secondary gap-1">
              <Plus size={14} /><T>Qo'shish</T>
            </button>
          </div>
        </Field>

        <Field label="Batafsil ma'lumot *">
          <textarea className="input min-h-[100px]" required value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
        </Field>

        <div className="grid sm:grid-cols-2 gap-3">
          <Field label="Ish joyi (Lokatsiya URL)">
            <input className="input" placeholder={t("Xarita havolasi")} value={form.locationUrl} onChange={(e) => setForm({ ...form, locationUrl: e.target.value })} />
          </Field>
          <Field label="Manzil matni">
            <input className="input" value={form.locationText} onChange={(e) => setForm({ ...form, locationText: e.target.value })} />
          </Field>
        </div>

        <div className="grid sm:grid-cols-3 gap-3">
          <Field label="Viloyat">
            <input className="input" value={form.region} onChange={(e) => setForm({ ...form, region: e.target.value })} />
          </Field>
          <Field label="Tuman">
            <input className="input" value={form.district} onChange={(e) => setForm({ ...form, district: e.target.value })} />
          </Field>
          <Field label="Ishchilar soni *">
            <input type="number" min={1} className="input" value={form.workersNeeded} onChange={(e) => setForm({ ...form, workersNeeded: parseInt(e.target.value || "1", 10) })} />
          </Field>
        </div>

        <Field label="Taklif qilinayotgan narx">
          <div className="grid sm:grid-cols-[1fr_1fr_auto] gap-3 items-center">
            <select className="input" value={form.pricingType} onChange={(e) => setForm({ ...form, pricingType: e.target.value as any })}>
              <option value="per_worker">{t("Har bir ishchi uchun")}</option>
              <option value="total">{t("Umumiy summa")}</option>
            </select>
            <input
              type="number" min={0} className="input"
              placeholder={t("Bo'sh qoldirilsa Kelishiladi")}
              value={form.priceAmount}
              onChange={(e) => setForm({ ...form, priceAmount: e.target.value === "" ? "" : Number(e.target.value) })}
            />
            <span className="text-sm text-[color:var(--text-muted)]">so'm</span>
          </div>
          <div className="mt-2 text-sm text-[color:var(--text-muted)]">
            {live.neg ? <T>Agar narx kelishilgan holda bo'lsa, bo'sh qoldiring.</T> : <span><T>Kishi boshiga</T>: <b>{fmtSum(live.per)}</b> so'm • <T>Jami</T>: <b>{fmtSum(live.total)}</b> so'm</span>}
          </div>
        </Field>

        <div className="grid sm:grid-cols-3 gap-3">
          <Field label="Boshlanish sanasi *">
            <input required type="date" className="input" value={form.startDate} onChange={(e) => setForm({ ...form, startDate: e.target.value })} />
          </Field>
          <Field label="Ish vaqti (dan)">
            <input type="time" className="input" value={form.workTimeFrom} onChange={(e) => setForm({ ...form, workTimeFrom: e.target.value })} />
          </Field>
          <Field label="Ish vaqti (gacha)">
            <input type="time" className="input" value={form.workTimeTo} onChange={(e) => setForm({ ...form, workTimeTo: e.target.value })} />
          </Field>
        </div>

        <Field label="Aloqa telefon raqami *">
          <input className="input" required value={form.contactPhone} onChange={(e) => setForm({ ...form, contactPhone: e.target.value })} />
        </Field>

        <Field label="Rasmlar (ixtiyoriy)">
          <MultiImageUploader value={form.images} onChange={(images) => setForm({ ...form, images })} max={6} />
          <div className="mt-1 text-xs muted">JPG / PNG / WebP, har biri 8MB gacha. Maks 6 ta rasm.</div>
        </Field>

        <div className="flex justify-end gap-2">
          <button type="button" className="btn-secondary" onClick={() => router.back()}><T>Bekor qilish</T></button>
          <button type="button" className="btn-secondary" disabled={create.isPending} onClick={() => create.mutate(false)}><T>Qoralama saqlash</T></button>
          <button className="btn-primary" disabled={create.isPending}><T>E'lonni joylashtirish</T></button>
        </div>
      </form>
    </Shell>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-sm font-medium"><T>{label}</T></span>
      <div className="mt-1">{children}</div>
    </label>
  );
}

function SuccessPage({ onMine, onHome }: { onMine: () => void; onHome: () => void }) {
  return (
    <Shell title="E'lon joylashtirildi">
      <div className="card p-10 text-center">
        <CheckCircle2 size={56} className="mx-auto text-success" />
        <h2 className="text-xl font-bold mt-3"><T>E'loningiz muvaffaqiyatli joylashtirildi!</T></h2>
        <p className="text-[color:var(--text-muted)] mt-1"><T>Endi ishchilar arizalarini kuting.</T></p>
        <div className="mt-5 flex justify-center gap-2">
          <button onClick={onMine} className="btn-primary"><T>Mening e'lonlarimga o'tish</T></button>
          <button onClick={onHome} className="btn-secondary"><T>Asosiy sahifaga qaytish</T></button>
        </div>
      </div>
    </Shell>
  );
}

function ErrorPage({ onRetry, onSupport }: { onRetry: () => void; onSupport: () => void }) {
  return (
    <Shell title="Xatolik">
      <div className="card p-10 text-center">
        <AlertTriangle size={56} className="mx-auto text-danger" />
        <h2 className="text-xl font-bold mt-3"><T>E'lonni joylashtirishda xatolik yuz berdi</T></h2>
        <p className="text-[color:var(--text-muted)] mt-1"><T>Iltimos, qayta urinib ko'ring.</T></p>
        <div className="mt-5 flex justify-center gap-2">
          <button onClick={onRetry} className="btn-primary"><T>Qayta urinib ko'rish</T></button>
          <button onClick={onSupport} className="btn-secondary"><T>Yordamga murojaat qilish</T></button>
        </div>
      </div>
    </Shell>
  );
}
