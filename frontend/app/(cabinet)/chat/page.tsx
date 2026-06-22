"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Conversation, User } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { Search } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { T, useT } from "@/components/T";
import { fromNow } from "@/lib/format";

export default function ChatList() {
  const t = useT();
  const [q, setQ] = useState("");
  const [results, setResults] = useState<User[]>([]);
  const { data: convs } = useQuery<Conversation[]>({
    queryKey: ["conversations"],
    queryFn: () => api.get<Conversation[]>("/api/conversations"),
  });
  useEffect(() => {
    if (!q) { setResults([]); return; }
    const t = setTimeout(() => {
      api.get<User[]>(`/api/users?q=${encodeURIComponent(q)}`).then(setResults).catch(() => {});
    }, 250);
    return () => clearTimeout(t);
  }, [q]);
  return (
    <Shell title="Xabarlar">
      <div className="card p-4">
        <div className="relative">
          <Search size={18} className="absolute left-3 top-2.5 text-[color:var(--text-muted)]" />
          <input className="input pl-9" placeholder={t("Foydalanuvchi qidirish…")} value={q} onChange={(e) => setQ(e.target.value)} />
        </div>
        {results.length > 0 && (
          <div className="mt-2 grid gap-1">
            {results.map((u) => (
              <button key={u.id} className="text-left rounded-lg px-3 py-2 hover:bg-black/5" onClick={async () => {
                const c = await api.post<{ id: string }>("/api/conversations", { userId: u.id });
                location.href = `/chat/${c.id}`;
              }}>
                {u.firstName} {u.lastName} <span className="text-xs text-[color:var(--text-muted)]">{u.region}</span>
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="grid gap-2">
        {(convs || []).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)]"><T>Xabarlar yo'q.</T></div>}
        {(convs || []).map((c) => (
          <Link key={c.id} href={`/chat/${c.id}`} className="card p-4 flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-brand-navy text-white grid place-items-center">?</div>
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate">{c.lastMessageText}</div>
              <div className="text-xs text-[color:var(--text-muted)]">{fromNow(c.lastMessageAt)}</div>
            </div>
          </Link>
        ))}
      </div>
    </Shell>
  );
}
