"use client";
import { useState } from "react";
import { api } from "@/lib/api";

export default function AdminBroadcast() {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [sent, setSent] = useState<number | null>(null);
  async function send() {
    const r = await api.post<{ sent: number }>("/api/admin/broadcast", { title, body }, { auth: "admin" } as any);
    setSent(r.sent); setTitle(""); setBody("");
  }
  return (
    <div className="card p-6 max-w-xl grid gap-3">
      <h1 className="font-semibold text-lg">Tarqatma yuborish</h1>
      <input className="input" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Sarlavha" />
      <textarea className="input min-h-[100px]" value={body} onChange={(e) => setBody(e.target.value)} placeholder="Matn" />
      <button onClick={send} className="btn-primary">Yuborish</button>
      {sent !== null && <div className="text-sm text-success">{sent} foydalanuvchiga yuborildi</div>}
    </div>
  );
}
