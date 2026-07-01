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
      <div className="-mx-4 px-4 overflow-x-auto scroll-y-auto">
        <table className="w-full min-w-[680px] text-sm">
          <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Sabab</th><th>Maqsad</th><th>Holat</th><th>Sana</th><th></th></tr></thead>
          <tbody>
            {items.map((r) => (
              <tr key={r.id} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="py-2">{r.reason}</td>
                <td className="whitespace-nowrap">{r.targetType}:{r.targetId.slice(-6)}</td>
                <td>{r.status}</td>
                <td className="whitespace-nowrap">{new Date(r.createdAt).toLocaleString("uz-UZ")}</td>
                <td>
                  {r.status === "open" && (
                    <div className="flex flex-wrap gap-2 justify-end">
                      <button onClick={() => resolve(r.id, "resolved")} className="btn-primary btn-sm">Hal qilish</button>
                      <button onClick={() => resolve(r.id, "dismissed")} className="btn-secondary btn-sm">Rad etish</button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
