"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { api, setAdminToken } from "@/lib/api";

export default function AdminLogin() {
  const router = useRouter();
  const [u, setU] = useState("admin");
  const [p, setP] = useState("");
  const [err, setErr] = useState("");
  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    try {
      const r = await api.post<{ accessToken: string }>("/api/admin/login", { username: u, password: p }, { auth: "none" });
      setAdminToken(r.accessToken);
      router.replace("/admin");
    } catch (e: any) {
      setErr(e?.message || "Xatolik");
    }
  }
  return (
    <div className="min-h-screen grid place-items-center p-4">
      <form onSubmit={submit} className="card w-full max-w-sm p-6 grid gap-3">
        <h1 className="text-xl font-bold heading">Admin kirish</h1>
        <input className="input" value={u} onChange={(e) => setU(e.target.value)} placeholder="username" required />
        <input className="input" type="password" value={p} onChange={(e) => setP(e.target.value)} placeholder="parol" required />
        {err && <div className="text-danger text-sm">{err}</div>}
        <button className="btn-primary">Kirish</button>
      </form>
    </div>
  );
}
