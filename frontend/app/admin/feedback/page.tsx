"use client";
import { useEffect, useState } from "react";
import { api, Feedback } from "@/lib/api";

export default function AdminFeedback() {
  const [items, setItems] = useState<Feedback[]>([]);
  const [filter, setFilter] = useState<"all" | "suggestion" | "complaint">("all");

  async function load() {
    setItems(await api.get<Feedback[]>("/api/admin/feedback", { auth: "admin" } as any));
  }
  useEffect(() => { load(); }, []);

  async function resolve(id: string) {
    await api.patch(`/api/admin/feedback/${id}/resolve`, {}, { auth: "admin" } as any);
    load();
  }

  const shown = items.filter((f) => filter === "all" || f.type === filter);

  return (
    <div className="card p-4">
      <div className="flex gap-2 mb-3">
        {(["all", "suggestion", "complaint"] as const).map((f) => (
          <button key={f} onClick={() => setFilter(f)}
            className={`chip ${filter === f ? "chip-active" : ""}`}>
            {f === "all" ? "Barchasi" : f === "suggestion" ? "Takliflar" : "Shikoyatlar"}
          </button>
        ))}
      </div>
      <div className="-mx-4 px-4 overflow-x-auto scroll-y-auto">
        <table className="w-full min-w-[760px] text-sm">
          <thead>
            <tr className="text-left text-[color:var(--text-muted)]">
              <th className="py-2">Turi</th><th>Foydalanuvchi</th><th>Mavzu</th><th>Xabar</th><th>Holat</th><th>Sana</th><th></th>
            </tr>
          </thead>
          <tbody>
            {shown.map((f) => (
              <tr key={f.id} className="border-t align-top" style={{ borderColor: "var(--border)" }}>
                <td className="py-2 whitespace-nowrap">{f.type === "complaint" ? "Shikoyat" : "Taklif"}</td>
                <td className="whitespace-nowrap">{f.userName || "—"}<div className="text-xs muted">{f.userPhone}</div></td>
                <td>{f.subject || "—"}</td>
                <td className="max-w-[320px] break-words">{f.message}</td>
                <td className="whitespace-nowrap">{f.status === "resolved" ? "Hal qilindi" : "Ochiq"}</td>
                <td className="whitespace-nowrap">{new Date(f.createdAt).toLocaleString("uz-UZ")}</td>
                <td className="text-right">
                  {f.status !== "resolved" && (
                    <button onClick={() => resolve(f.id)} className="btn-primary btn-sm">Hal qilish</button>
                  )}
                </td>
              </tr>
            ))}
            {shown.length === 0 && (
              <tr><td colSpan={7} className="py-6 text-center muted">Murojaatlar yo'q.</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
