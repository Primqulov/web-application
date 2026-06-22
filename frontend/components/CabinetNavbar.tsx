"use client";
import Link from "next/link";
import { Bell } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { api, Notification, User } from "@/lib/api";
import { ThemeToggle } from "./ThemeToggle";
import { ScriptToggle } from "./ScriptToggle";
import { T } from "./T";

export function CabinetNavbar({ title }: { title: string }) {
  const { data: notifs } = useQuery<Notification[]>({
    queryKey: ["notifications"],
    queryFn: () => api.get<Notification[]>("/api/notifications"),
  });
  const { data: me } = useQuery<User>({ queryKey: ["me"], queryFn: () => api.get<User>("/api/me") });
  const unread = (notifs || []).filter((n) => !n.isRead).length;
  return (
    <div className="card flex items-center justify-between p-3">
      <h1 className="text-lg font-semibold"><T>{title}</T></h1>
      <div className="flex items-center gap-2">
        <ScriptToggle />
        <ThemeToggle />
        <Link href="/notifications" className="relative grid h-9 w-9 place-items-center rounded-lg border hover:bg-black/5"
          style={{ borderColor: "var(--border)" }} aria-label="Bildirishnomalar">
          <Bell size={18} />
          {unread > 0 && <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-danger" />}
        </Link>
        <Link href="/profile" className="hidden sm:flex items-center gap-2 rounded-lg border px-2 py-1.5 hover:bg-black/5"
          style={{ borderColor: "var(--border)" }}>
          <div className="grid h-7 w-7 place-items-center rounded-full bg-brand-navy text-white text-xs">
            {(me?.firstName?.[0] || "?").toUpperCase()}
          </div>
          <span className="text-sm"><T>Mening profilim</T></span>
        </Link>
      </div>
    </div>
  );
}
