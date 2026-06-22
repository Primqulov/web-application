"use client";
import { useEffect, useState } from "react";
import { api, Review, User } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { Star, CheckCircle2 } from "lucide-react";
import { T } from "@/components/T";

export default function MyProfile() {
  const [me, setMe] = useState<User | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  useEffect(() => {
    api.get<User>("/api/me").then((u) => {
      setMe(u);
      api.get<Review[]>(`/api/users/${u.id}/reviews`).then(setReviews).catch(() => {});
    });
  }, []);
  if (!me) return <div className="p-6">Yuklanmoqda…</div>;
  return (
    <Shell title="Mening profilim">
      <div className="card p-6 flex gap-4 items-start">
        <div className="h-20 w-20 rounded-full bg-brand-navy text-white grid place-items-center text-2xl font-bold">{me.firstName?.[0]?.toUpperCase()}</div>
        <div className="flex-1">
          <h1 className="text-xl font-bold">{me.firstName} {me.lastName} {me.isPhoneVerified && <span className="ml-1 align-middle text-success"><CheckCircle2 size={16} className="inline" /></span>}</h1>
          <div className="text-sm text-[color:var(--text-muted)] mt-1 flex items-center gap-2">
            <span className="inline-flex items-center gap-1"><Star size={14} className="text-accent-amber" />{me.rating.toFixed(1)} ({me.reviewsCount} <T>baho</T>)</span>
            <span>•</span><span>{me.region}{me.district ? `, ${me.district}` : ""}</span>
            <span>•</span><span>{me.completedJobsCount} <T>bajarilgan ish</T></span>
          </div>
          {me.skills && me.skills.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-2">
              {me.skills.map((s, i) => <span key={i} className="chip">{s}</span>)}
            </div>
          )}
          {me.bio && <p className="mt-3 text-sm"><T>{me.bio}</T></p>}
        </div>
      </div>
      <div className="card p-6">
        <h2 className="font-semibold mb-3"><T>Baholar va sharhlar</T></h2>
        {reviews.length === 0 && <div className="text-[color:var(--text-muted)]"><T>Hozircha sharhlar yo'q.</T></div>}
        <div className="grid gap-3">
          {reviews.map((r) => (
            <div key={r.id} className="border rounded-lg p-3" style={{ borderColor: "var(--border)" }}>
              <div className="flex items-center gap-1 text-accent-amber">
                {Array.from({ length: r.rating }).map((_, i) => <Star key={i} size={14} fill="currentColor" />)}
              </div>
              {r.comment && <p className="text-sm mt-1">{r.comment}</p>}
            </div>
          ))}
        </div>
      </div>
    </Shell>
  );
}
