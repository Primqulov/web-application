"use client";
import { useEffect, useState } from "react";
import { api, Elon } from "@/lib/api";

export default function AdminElons() {
  const [elons, setElons] = useState<Elon[]>([]);
  async function load() { setElons(await api.get<Elon[]>("/api/admin/elons", { auth: "admin" } as any)); }
  useEffect(() => { load(); }, []);
  async function del(id: string) {
    if (!confirm("O'chirilsinmi?")) return;
    await api.delete(`/api/admin/elons/${id}`, { auth: "admin" } as any);
    load();
  }
  return (
    <div className="card p-4">
      <div className="-mx-4 px-4 overflow-x-auto scroll-y-auto">
        <table className="w-full min-w-[680px] text-sm">
          <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Sarlavha</th><th>Kategoriya</th><th>Holat</th><th>Ishchilar</th><th>Narx</th><th></th></tr></thead>
          <tbody>
            {elons.map((e) => (
              <tr key={e.id} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="py-2">{e.title}</td><td>{e.categoryName}</td><td>{e.status}</td><td>{e.workersNeeded}</td>
                <td className="whitespace-nowrap">{e.priceAmount.toLocaleString("uz-UZ")}</td>
                <td className="text-right"><button onClick={() => del(e.id)} className="btn-danger btn-sm">O'chirish</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
