"use client";
import { useEffect, useState } from "react";
import { api, User } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { AvatarUploader } from "@/components/ui/ImageUpload";
import { useScript } from "@/lib/i18n";
import { T, useT } from "@/components/T";

export default function Settings() {
  const t = useT();
  const [me, setMe] = useState<User | null>(null);
  const script = useScript((s) => s.script);
  const setScript = useScript((s) => s.setScript);
  const [first, setFirst] = useState("");
  const [last, setLast] = useState("");
  const [avatarUrl, setAvatarUrl] = useState<string | undefined>(undefined);

  useEffect(() => {
    api.get<User>("/api/me").then((u) => {
      setMe(u);
      setFirst(u.firstName);
      setLast(u.lastName);
      setAvatarUrl(u.avatarUrl || undefined);
    });
  }, []);

  async function save() {
    await api.patch("/api/me", { firstName: first, lastName: last, langPref: script, avatarUrl: avatarUrl || "" });
  }

  return (
    <Shell title="Sozlamalar">
      <div className="card p-6 grid gap-4 max-w-xl">
        <h2 className="font-semibold heading"><T>Profil ma'lumotlari</T></h2>
        <AvatarUploader value={avatarUrl} name={`${first} ${last}`} onChange={setAvatarUrl} />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label className="block"><span className="text-sm"><T>Ism</T></span><input className="input mt-1" value={first} onChange={(e) => setFirst(e.target.value)} /></label>
          <label className="block"><span className="text-sm"><T>Familiya</T></span><input className="input mt-1" value={last} onChange={(e) => setLast(e.target.value)} /></label>
        </div>
        <div>
          <button onClick={save} className="btn-primary"><T>Saqlash</T></button>
        </div>
      </div>

      <div className="card p-6 max-w-xl">
        <h2 className="font-semibold mb-2"><T>Til sozlamalari</T></h2>
        <div className="flex gap-3">
          <label className="flex items-center gap-2"><input type="radio" checked={script === "latin"} onChange={() => setScript("latin")} /><T>O'zbekcha (Lotin)</T></label>
          <label className="flex items-center gap-2"><input type="radio" checked={script === "cyrillic"} onChange={() => setScript("cyrillic")} />Ўзбекча (Кирилл)</label>
        </div>
      </div>

      <div className="card p-6 max-w-xl border-danger">
        <h2 className="font-semibold text-danger mb-2"><T>Hisobni o'chirish</T></h2>
        <p className="text-sm text-[color:var(--text-muted)] mb-3"><T>Bu amalni qaytarib bo'lmaydi.</T></p>
        <button className="btn-danger"><T>O'chirish</T></button>
      </div>
    </Shell>
  );
}
