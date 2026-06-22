"use client";
import { useEffect, useState } from "react";
import { api, User } from "@/lib/api";

export default function AdminUsers() {
  const [users, setUsers] = useState<User[]>([]);
  async function load() { setUsers(await api.get<User[]>("/api/admin/users", { auth: "admin" } as any)); }
  useEffect(() => { load(); }, []);
  async function block(id: string, isBlocked: boolean) {
    await api.post(`/api/admin/users/${id}/block`, { isBlocked }, { auth: "admin" } as any);
    load();
  }
  async function del(id: string) {
    if (!confirm("O'chirilsinmi?")) return;
    await api.delete(`/api/admin/users/${id}`, { auth: "admin" } as any);
    load();
  }
  return (
    <div className="card p-4 overflow-x-auto">
      <table className="w-full text-sm">
        <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Ism</th><th>Telefon</th><th>Viloyat</th><th>Reyting</th><th>Holat</th><th></th></tr></thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id} className="border-t" style={{ borderColor: "var(--border)" }}>
              <td className="py-2">{u.firstName} {u.lastName}</td>
              <td>{u.phone}</td><td>{u.region}</td><td>{u.rating.toFixed(1)}</td>
              <td>{u.isBlocked ? <span className="text-danger">bloklangan</span> : "faol"}</td>
              <td className="text-right">
                <button onClick={() => block(u.id, !u.isBlocked)} className="btn-secondary mr-2">{u.isBlocked ? "Ochish" : "Bloklash"}</button>
                <button onClick={() => del(u.id)} className="btn-danger">O'chirish</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
