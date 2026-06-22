"use client";
import { useEffect, useState } from "react";
import { api, Category } from "@/lib/api";

export default function AdminCategories() {
  const [cats, setCats] = useState<Category[]>([]);
  async function load() { setCats(await api.get<Category[]>("/api/admin/categories", { auth: "admin" } as any)); }
  useEffect(() => { load(); }, []);
  async function toggle(c: Category) {
    await api.patch(`/api/admin/categories/${c.id}/active`, { isActive: !c.isActive }, { auth: "admin" } as any);
    load();
  }
  return (
    <div className="card p-4">
      <table className="w-full text-sm">
        <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Nomi</th><th>Slug</th><th>Foydalanish</th><th>Holat</th><th></th></tr></thead>
        <tbody>
          {cats.map((c) => (
            <tr key={c.id} className="border-t" style={{ borderColor: "var(--border)" }}>
              <td className="py-2">{c.name}</td><td>{c.slug}</td><td>{c.usageCount}</td>
              <td>{c.isActive ? "Faol" : "O'chirilgan"}</td>
              <td className="text-right"><button onClick={() => toggle(c)} className="btn-secondary">{c.isActive ? "O'chirish" : "Yoqish"}</button></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
