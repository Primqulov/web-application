"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, Notification } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { Bell } from "lucide-react";
import { T } from "@/components/T";
import { fromNow } from "@/lib/format";

export default function Notifications() {
  const qc = useQueryClient();
  const { data } = useQuery<Notification[]>({
    queryKey: ["notifications"],
    queryFn: () => api.get<Notification[]>("/api/notifications"),
  });
  const read = useMutation({
    mutationFn: () => api.post("/api/notifications/read-all"),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notifications"] }),
  });
  return (
    <Shell title="Bildirishnomalar">
      <div className="card p-4 flex items-center justify-between">
        <h2 className="font-semibold"><T>Bildirishnomalar</T></h2>
        <button onClick={() => read.mutate()} className="btn-secondary"><T>Hammasini o'qildi</T></button>
      </div>
      <div className="grid gap-2">
        {(data || []).length === 0 && <div className="card p-8 text-center text-[color:var(--text-muted)]"><T>Bildirishnomalar yo'q.</T></div>}
        {(data || []).map((n) => (
          <div key={n.id} className={`card p-4 flex gap-3 ${!n.isRead ? "ring-2 ring-brand-navy/20" : ""}`}>
            <Bell size={18} className="heading mt-0.5" />
            <div className="flex-1">
              <div className="font-medium"><T>{n.title}</T></div>
              <div className="text-sm text-[color:var(--text-muted)]"><T>{n.body}</T></div>
            </div>
            <div className="text-xs text-[color:var(--text-muted)]">{fromNow(n.createdAt)}</div>
          </div>
        ))}
      </div>
    </Shell>
  );
}
