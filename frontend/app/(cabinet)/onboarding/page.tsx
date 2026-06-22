"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, User } from "@/lib/api";
import { CheckCircle2 } from "lucide-react";
import { Shell } from "@/components/Shell";
import { AvatarUploader } from "@/components/ui/ImageUpload";
import { T, useT } from "@/components/T";

const REGIONS = ["Toshkent", "Samarqand", "Buxoro", "Farg'ona", "Namangan", "Andijon", "Qashqadaryo", "Surxondaryo", "Xorazm", "Navoiy", "Jizzax", "Sirdaryo"];

export default function Onboarding() {
  const router = useRouter();
  const t = useT();
  const [me, setMe] = useState<User | null>(null);
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [region, setRegion] = useState("");
  const [district, setDistrict] = useState("");
  const [avatarUrl, setAvatarUrl] = useState<string | undefined>(undefined);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api.get<User>("/api/me").then((u) => {
      setMe(u);
      setFirstName(u.firstName || "");
      setLastName(u.lastName || "");
      setRegion(u.region || "");
      setDistrict(u.district || "");
      setAvatarUrl(u.avatarUrl || undefined);
    });
  }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api.patch("/api/me", { firstName, lastName, region, district, avatarUrl: avatarUrl || "" });
      router.replace("/dashboard");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Shell title="Profilni to'ldirish">
      <form onSubmit={submit} className="card p-6 grid gap-4 max-w-2xl">
        <AvatarUploader value={avatarUrl} name={`${firstName} ${lastName}`} onChange={setAvatarUrl} />
        <label className="block">
          <span className="text-sm font-medium"><T>To'liq ism</T></span>
          <div className="grid grid-cols-2 gap-2 mt-1">
            <input className="input" value={firstName} onChange={(e) => setFirstName(e.target.value)} placeholder={t("Ism")} required />
            <input className="input" value={lastName} onChange={(e) => setLastName(e.target.value)} placeholder={t("Familiya")} />
          </div>
        </label>
        <label className="block">
          <span className="text-sm font-medium"><T>Telefon raqami</T></span>
          <div className="mt-1 flex items-center gap-2">
            <input className="input bg-black/5" value={me?.phone || ""} disabled />
            <span className="badge-success"><CheckCircle2 size={12} /> <T>Tasdiqlangan</T></span>
          </div>
        </label>
        <div className="grid grid-cols-2 gap-3">
          <label className="block">
            <span className="text-sm font-medium"><T>Viloyat</T></span>
            <select className="input mt-1" value={region} onChange={(e) => setRegion(e.target.value)} required>
              <option value="">{t("Tanlang")}</option>
              {REGIONS.map((r) => <option key={r} value={r}>{r}</option>)}
            </select>
          </label>
          <label className="block">
            <span className="text-sm font-medium"><T>Tuman</T></span>
            <input className="input mt-1" value={district} onChange={(e) => setDistrict(e.target.value)} placeholder={t("Tuman")} />
          </label>
        </div>
        <div className="flex justify-end gap-2 mt-2">
          <button type="button" className="btn-secondary" onClick={() => router.back()}><T>Bekor qilish</T></button>
          <button disabled={saving} className="btn-primary disabled:opacity-50"><T>Saqlash</T></button>
        </div>
      </form>
    </Shell>
  );
}
