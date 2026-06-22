"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

interface Report {
  id: string; reporterId: string; targetType: string; targetId: string;
  reason: string; description?: string; status: string; createdAt: string;
}

export default function AdminReports() {
  const [items, setItems] = useState<Report[]>([]);
  async function load() { setItems(await api.get<Report[]>("/api/admin/reports", { auth: "admin" } as any)); }
  useEffect(() => { load(); }, []);
  async function resolve(id: string, status: string) {
    await api.patch(`/api/admin/reports/${id}/resolve`, { status }, { auth: "admin" } as any);
    load();
  }
  return (
    <div className="card p-4">
      <table className="w-full text-sm">
        <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Sabab</th><th>Maqsad</th><th>Holat</th><th>Sana</th><th></th></tr></thead>
        <tbody>
          {items.map((r) => (
            <tr key={r.id} className="border-t" style={{ borderColor: "var(--border)" }}>
              <td className="py-2">{r.reason}</td>
              <td>{r.targetType}:{r.targetId.slice(-6)}</td>
              <td>{r.status}</td>
              <td>{new Date(r.createdAt).toLocaleString("uz-UZ")}</td>
              <td className="text-right">
                {r.status === "open" && <>
                  <button onClick={() => resolve(r.id, "resolved")} className="btn-primary mr-2">Hal qilish</button>
                  <button onClick={() => resolve(r.id, "dismissed")} className="btn-secondary">Rad etish</button>
                </>}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
