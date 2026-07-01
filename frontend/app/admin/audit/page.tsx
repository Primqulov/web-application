"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

interface Audit { id: string; adminId: string; action: string; target?: string; detail?: string; createdAt: string; }

export default function AdminAudit() {
  const [items, setItems] = useState<Audit[]>([]);
  useEffect(() => { api.get<Audit[]>("/api/admin/audit", { auth: "admin" } as any).then(setItems); }, []);
  return (
    <div className="card p-4">
      <div className="-mx-4 px-4 overflow-x-auto scroll-y-auto">
        <table className="w-full min-w-[680px] text-sm">
          <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Vaqt</th><th>Admin</th><th>Amal</th><th>Maqsad</th><th>Tafsilot</th></tr></thead>
          <tbody>
            {items.map((a) => (
              <tr key={a.id} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="py-2 whitespace-nowrap">{new Date(a.createdAt).toLocaleString("uz-UZ")}</td>
                <td>{a.adminId.slice(-6)}</td><td>{a.action}</td><td>{a.target}</td><td>{a.detail}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
