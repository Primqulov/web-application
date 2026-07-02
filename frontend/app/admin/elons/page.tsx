"use client";
import { useEffect, useState } from "react";
import { api, Elon } from "@/lib/api";
import { Modal } from "@/components/Modal";

export default function AdminElons() {
  const [elons, setElons] = useState<Elon[]>([]);
  const [delId, setDelId] = useState("");
  async function load() { setElons(await api.get<Elon[]>("/api/admin/elons", { auth: "admin" } as any)); }
  useEffect(() => { load(); }, []);
  async function del() {
    await api.delete(`/api/admin/elons/${delId}`, { auth: "admin" } as any);
    setDelId("");
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
                <td className="text-right"><button onClick={() => setDelId(e.id)} className="btn-danger btn-sm">O'chirish</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* O'chirishni tasdiqlash — modal ko'rinishida */}
      <Modal open={!!delId} onClose={() => setDelId("")} title="E'lonni o'chirasizmi?" footer={
        <>
          <button onClick={() => setDelId("")} className="btn-secondary">Yo'q</button>
          <button onClick={del} className="btn-danger">Ha, o'chirish</button>
        </>
      }>
        <p className="text-sm muted">E'lon o'chiriladi. Davom etasizmi?</p>
      </Modal>
    </div>
  );
}
