"use client";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api, Category, Elon } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { T, useT } from "@/components/T";

export default function EditElon() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const t = useT();
  const [e, setE] = useState<Elon | null>(null);
  const [cats, setCats] = useState<Category[]>([]);
  useEffect(() => {
    api.get<Elon>(`/api/elons/${id}`).then(setE);
    api.get<Category[]>("/api/categories").then(setCats);
  }, [id]);
  if (!e) return <div className="p-6">Yuklanmoqda…</div>;
  async function save(ev: React.FormEvent) {
    ev.preventDefault();
    await api.patch(`/api/elons/${e!.id}`, {
      title: e!.title, categoryId: e!.categoryId, description: e!.description,
      locationUrl: e!.locationUrl, locationText: e!.locationText, region: e!.region, district: e!.district,
      workersNeeded: e!.workersNeeded, pricingType: e!.pricingType,
      priceAmount: e!.pricingType === "total" ? e!.priceAmount : e!.perWorkerAmount,
      startDate: e!.startDate, workTimeFrom: e!.workTimeFrom, workTimeTo: e!.workTimeTo,
      contactPhone: e!.contactPhone,
    });
    router.push("/my-elons");
  }
  return (
    <Shell title="E'lonni tahrirlash">
      <form onSubmit={save} className="card p-6 grid gap-4 max-w-3xl">
        <input className="input" value={e.title} onChange={(ev) => setE({ ...e, title: ev.target.value })} placeholder={t("Sarlavha")} required />
        <select className="input" value={e.categoryId} onChange={(ev) => setE({ ...e, categoryId: ev.target.value })}>
          {cats.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select>
        <textarea className="input min-h-[100px]" value={e.description} onChange={(ev) => setE({ ...e, description: ev.target.value })} required />
        <div className="grid grid-cols-2 gap-3">
          <input className="input" type="number" min={1} value={e.workersNeeded} onChange={(ev) => setE({ ...e, workersNeeded: parseInt(ev.target.value || "1", 10) })} />
          <input className="input" value={e.contactPhone || ""} onChange={(ev) => setE({ ...e, contactPhone: ev.target.value })} placeholder="+998..." />
        </div>
        <div className="flex justify-end gap-2">
          <button type="button" onClick={() => router.back()} className="btn-secondary"><T>Bekor qilish</T></button>
          <button className="btn-primary"><T>Saqlash</T></button>
        </div>
      </form>
    </Shell>
  );
}
