"use client";
import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api, Category, Elon, User } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { MultiImageUploader } from "@/components/ui/ImageUpload";
import { MapPicker, LatLng } from "@/components/ui/MapPicker";
import { CheckCircle2, AlertTriangle } from "lucide-react";
import { T, useT } from "@/components/T";
import { fmtSum, fmtThousands, onlyDigits, fmtPhone, phoneDigits } from "@/lib/format";

type Form = {
  title: string;
  categoryId: string;
  description: string;
  loc: LatLng | null;
  workersNeeded: number | "";
  pricingType: "per_worker" | "total";
  priceAmount: number | "";
  startDate: string;
  workTimeFrom: string;
  contactPhone: string;
  images: string[];
};

function pad(n: number) {
  return String(n).padStart(2, "0");
}
function ymd(d: Date) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}
function hm(d: Date) {
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export default function CreateElon() {
  const router = useRouter();
  const t = useT();

  // Eng erta joylashtirish vaqti — hozirgi vaqtdan 1 soat keyin (sahifa
  // ochilgan payt asos). Sana esa faqat 3 kun ichida: bugun, erta yoki indin.
  const startRef = useRef(new Date());
  const minMoment = new Date(startRef.current.getTime() + 60 * 60 * 1000);
  const MIN_DATE = ymd(minMoment);
  const MAX_DATE = ymd(new Date(startRef.current.getTime() + 2 * 86400000));
  const minH = minMoment.getHours();
  const minM = minMoment.getMinutes();

  const [form, setForm] = useState<Form>({
    title: "",
    categoryId: "",
    description: "",
    loc: null,
    workersNeeded: 1,
    pricingType: "per_worker",
    priceAmount: "",
    startDate: MIN_DATE,
    workTimeFrom: hm(minMoment),
    contactPhone: "+998 ",
    images: [],
  });
  const [state, setState] = useState<"idle" | "ok" | "err">("idle");
  const [errMsg, setErrMsg] = useState("");

  const { data: cats } = useQuery<Category[]>({
    queryKey: ["categories"],
    queryFn: () => api.get<Category[]>("/api/categories"),
  });

  // Aloqa raqamini foydalanuvchi profilidan oldindan to'ldirish.
  useEffect(() => {
    api.get<User>("/api/me").then((u) => {
      if (u.phone) setForm((f) => ({ ...f, contactPhone: fmtPhone(u.phone) }));
    }).catch(() => {});
  }, []);

  const create = useMutation({
    mutationFn: async () => {
      const workers = typeof form.workersNeeded === "number" && form.workersNeeded > 0 ? form.workersNeeded : 1;
      const price = typeof form.priceAmount === "number" ? form.priceAmount : 0;
      const body = {
        title: form.title,
        categoryId: form.categoryId,
        description: form.description,
        lat: form.loc?.lat || 0,
        lng: form.loc?.lng || 0,
        workersNeeded: workers,
        pricingType: price === 0 ? "negotiable" : form.pricingType,
        priceAmount: price,
        startDate: form.startDate,
        workTimeFrom: form.workTimeFrom,
        contactPhone: phoneDigits(form.contactPhone) ? fmtPhone(form.contactPhone) : "",
        images: form.images,
      };
      // E'lon darhol chop etiladi (qoralama bosqichi yo'q).
      return api.post<Elon>("/api/elons", body);
    },
    onSuccess: () => setState("ok"),
    onError: (e: any) => { setErrMsg(e?.message || ""); setState("err"); },
  });

  const live = useMemo(() => {
    const n = typeof form.workersNeeded === "number" ? form.workersNeeded : 0;
    const p = typeof form.priceAmount === "number" ? form.priceAmount : 0;
    if (!p) return { per: 0, total: 0, neg: true };
    if (form.pricingType === "total") return { per: n ? Math.floor(p / n) : 0, total: p, neg: false };
    return { per: p, total: p * n, neg: false };
  }, [form.workersNeeded, form.priceAmount, form.pricingType]);

  useEffect(() => { if (cats?.length && !form.categoryId) setForm((f) => ({ ...f, categoryId: cats[0].id })); }, [cats]);

  // Tanlangan vaqtni eng erta ruxsat etilgan (hozir+1soat) dan oldin bo'lmasligini ta'minlaydi.
  function clampTime(dateStr: string, hh: number, mm: number): string {
    if (dateStr === MIN_DATE && (hh < minH || (hh === minH && mm < minM))) {
      hh = minH;
      mm = minM;
    }
    return `${pad(hh)}:${pad(mm)}`;
  }
  const [curH, curM] = (form.workTimeFrom || hm(minMoment)).split(":").map((x) => parseInt(x, 10) || 0);

  function submit() {
    if (!form.loc) { setErrMsg(t("Iltimos, ish joyini xaritadan belgilang.")); setState("err"); return; }
    const chosen = new Date(`${form.startDate}T${form.workTimeFrom || "00:00"}:00`);
    if (isNaN(chosen.getTime()) || chosen.getTime() < minMoment.getTime()) {
      setErrMsg(t("Ish boshlanish vaqti hozirgi vaqtdan kamida 1 soat keyin bo'lishi kerak."));
      setState("err"); return;
    }
    create.mutate();
  }

  if (state === "ok") return <SuccessPage onMine={() => router.push("/my-elons")} onHome={() => router.push("/dashboard")} />;
  if (state === "err") return <ErrorPage msg={errMsg} onRetry={() => setState("idle")} onSupport={() => router.push("/feedback")} />;

  return (
    <Shell title="Yangi e'lon yaratish">
      <form
        onSubmit={(e) => { e.preventDefault(); submit(); }}
        className="card p-6 grid gap-4 max-w-3xl"
      >
        <h2 className="font-semibold text-lg"><T>E'lon ma'lumotlari</T></h2>

        <Field label="Ish nomi *">
          <input className="input" required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder={t("Masalan, Mebel tashish")} />
        </Field>

        <Field label="Kategoriya *">
          <div className="flex flex-wrap gap-2">
            {(cats || []).map((c) => (
              <button key={c.id} type="button"
                onClick={() => setForm({ ...form, categoryId: c.id })}
                className={`chip ${form.categoryId === c.id ? "chip-active" : ""}`}>
                {c.icon}<T>{c.name}</T>
              </button>
            ))}
          </div>
          <div className="mt-1 text-xs muted"><T>Kategoriyalardan birini tanlang.</T></div>
        </Field>

        <Field label="Batafsil ma'lumot *">
          <textarea className="input min-h-[100px]" required value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
        </Field>

        {/* Ish joyi — xaritadan tanlanadi. Viloyat/tuman avtomatik aniqlanadi. */}
        <Field label="Ish joyi (xaritadan belgilang) *">
          <MapPicker value={form.loc} onChange={(loc) => setForm((f) => ({ ...f, loc }))} />
        </Field>

        <Field label="Ishchilar soni *">
          <input
            type="number" min={1} inputMode="numeric" className="input max-w-[200px]"
            value={form.workersNeeded}
            onChange={(e) => {
              const v = e.target.value;
              setForm({ ...form, workersNeeded: v === "" ? "" : Math.max(1, parseInt(v, 10) || 1) });
            }}
            onBlur={() => { if (form.workersNeeded === "" ) setForm((f) => ({ ...f, workersNeeded: 1 })); }}
            placeholder={t("Masalan, 3")}
          />
        </Field>

        <Field label="Taklif qilinayotgan narx">
          <div className="grid sm:grid-cols-[1fr_1fr_auto] gap-3 items-center">
            <select className="input" value={form.pricingType} onChange={(e) => setForm({ ...form, pricingType: e.target.value as any })}>
              <option value="per_worker">{t("Har bir ishchi uchun")}</option>
              <option value="total">{t("Umumiy summa")}</option>
            </select>
            <input
              type="text" inputMode="numeric" className="input"
              placeholder={t("Masalan, 150 000")}
              value={form.priceAmount === "" ? "" : fmtThousands(String(form.priceAmount))}
              onChange={(e) => {
                const digits = onlyDigits(e.target.value);
                setForm({ ...form, priceAmount: digits === "" ? "" : Number(digits) });
              }}
            />
            <span className="text-sm text-[color:var(--text-muted)]">so'm</span>
          </div>
          <div className="mt-2 text-sm text-[color:var(--text-muted)]">
            {live.neg ? <T>Agar narx kelishilgan holda bo'lsa, bo'sh qoldiring.</T> : <span><T>Kishi boshiga</T>: <b>{fmtSum(live.per)}</b> so'm • <T>Jami</T>: <b>{fmtSum(live.total)}</b> so'm</span>}
          </div>
        </Field>

        <div className="grid sm:grid-cols-2 gap-3">
          <Field label="Boshlanish sanasi *">
            <input required type="date" min={MIN_DATE} max={MAX_DATE} className="input" value={form.startDate}
              onChange={(e) => { const d = e.target.value; setForm({ ...form, startDate: d, workTimeFrom: clampTime(d, curH, curM) }); }} />
            <div className="mt-1 text-xs muted"><T>Bugundan 3 kun ichida</T></div>
          </Field>
          <Field label="Boshlanish vaqti (24 soat) *">
            <div className="flex items-center gap-2">
              <select className="input" value={curH}
                onChange={(e) => setForm({ ...form, workTimeFrom: clampTime(form.startDate, parseInt(e.target.value, 10), curM) })}>
                {Array.from({ length: 24 }).map((_, h) => (
                  <option key={h} value={h} disabled={form.startDate === MIN_DATE && h < minH}>{pad(h)}</option>
                ))}
              </select>
              <span className="font-semibold">:</span>
              <select className="input" value={curM}
                onChange={(e) => setForm({ ...form, workTimeFrom: clampTime(form.startDate, curH, parseInt(e.target.value, 10)) })}>
                {Array.from({ length: 60 }).map((_, m) => (
                  <option key={m} value={m} disabled={form.startDate === MIN_DATE && curH === minH && m < minM}>{pad(m)}</option>
                ))}
              </select>
            </div>
            <div className="mt-1 text-xs muted"><T>Kamida hozirgi vaqtdan 1 soat keyin (24 soatlik)</T></div>
          </Field>
        </div>

        <Field label="Aloqa telefon raqami *">
          <input
            className="input max-w-[260px]"
            required
            inputMode="numeric"
            value={form.contactPhone}
            onChange={(e) => setForm({ ...form, contactPhone: fmtPhone(e.target.value) })}
            placeholder="+998 90 020 25 35"
          />
        </Field>

        <Field label="Rasmlar (ixtiyoriy)">
          <MultiImageUploader value={form.images} onChange={(images) => setForm({ ...form, images })} max={6} />
          <div className="mt-1 text-xs muted">JPG / PNG / WebP, har biri 8MB gacha. Maks 6 ta rasm.</div>
        </Field>

        <div className="flex flex-wrap justify-end gap-2">
          <button type="button" className="btn-secondary" onClick={() => router.back()}><T>Bekor qilish</T></button>
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

function ErrorPage({ msg, onRetry, onSupport }: { msg?: string; onRetry: () => void; onSupport: () => void }) {
  return (
    <Shell title="Xatolik">
      <div className="card p-10 text-center">
        <AlertTriangle size={56} className="mx-auto text-danger" />
        <h2 className="text-xl font-bold mt-3"><T>E'lonni joylashtirishda xatolik yuz berdi</T></h2>
        <p className="text-[color:var(--text-muted)] mt-1">{msg ? msg : <T>Iltimos, qayta urinib ko'ring.</T>}</p>
        <div className="mt-5 flex justify-center gap-2">
          <button onClick={onRetry} className="btn-primary"><T>Qayta urinib ko'rish</T></button>
          <button onClick={onSupport} className="btn-secondary"><T>Yordamga murojaat qilish</T></button>
        </div>
      </div>
    </Shell>
  );
}
