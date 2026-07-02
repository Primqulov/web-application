"use client";
import { useEffect, useState } from "react";
import { api, User } from "@/lib/api";
import { Modal } from "@/components/Modal";

export default function AdminUsers() {
  const [users, setUsers] = useState<User[]>([]);
  const [delId, setDelId] = useState("");
  async function load() { setUsers(await api.get<User[]>("/api/admin/users", { auth: "admin" } as any)); }
  useEffect(() => { load(); }, []);
  async function block(id: string, isBlocked: boolean) {
    await api.post(`/api/admin/users/${id}/block`, { isBlocked }, { auth: "admin" } as any);
    load();
  }
  async function del() {
    await api.delete(`/api/admin/users/${delId}`, { auth: "admin" } as any);
    setDelId("");
    load();
  }
  return (
    <div className="card p-4">
      <div className="-mx-4 px-4 overflow-x-auto scroll-y-auto">
        <table className="w-full min-w-[720px] text-sm">
          <thead><tr className="text-left text-[color:var(--text-muted)]"><th className="py-2">Ism</th><th>Telefon</th><th>Viloyat</th><th>Reyting</th><th>Holat</th><th></th></tr></thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} className="border-t" style={{ borderColor: "var(--border)" }}>
                <td className="py-2 whitespace-nowrap">{u.firstName} {u.lastName}</td>
                <td className="whitespace-nowrap">{u.phone}</td><td>{u.region}</td><td>{u.rating.toFixed(1)}</td>
                <td>{u.isBlocked ? <span className="text-danger">bloklangan</span> : "faol"}</td>
                <td>
                  <div className="flex flex-wrap gap-2 justify-end">
                    <button onClick={() => block(u.id, !u.isBlocked)} className="btn-secondary btn-sm">{u.isBlocked ? "Ochish" : "Bloklash"}</button>
                    <button onClick={() => setDelId(u.id)} className="btn-danger btn-sm">O'chirish</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* O'chirishni tasdiqlash — modal ko'rinishida */}
      <Modal open={!!delId} onClose={() => setDelId("")} title="Foydalanuvchini o'chirasizmi?" footer={
        <>
          <button onClick={() => setDelId("")} className="btn-secondary">Yo'q</button>
          <button onClick={del} className="btn-danger">Ha, o'chirish</button>
        </>
      }>
        <p className="text-sm muted">Foydalanuvchi o'chiriladi. Davom etasizmi?</p>
      </Modal>
    </div>
  );
}
