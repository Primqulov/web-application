"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api, Review, User } from "@/lib/api";
import { Star, CheckCircle2 } from "lucide-react";
import { T } from "@/components/T";
import { ScriptToggle } from "@/components/ScriptToggle";

export default function PublicProfile() {
  const { id } = useParams<{ id: string }>();
  const [u, setU] = useState<User | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  useEffect(() => {
    api.get<User>(`/api/users/${id}`, { auth: "none" } as any).then(setU);
    api.get<Review[]>(`/api/users/${id}/reviews`, { auth: "none" } as any).then(setReviews).catch(() => {});
  }, [id]);
  if (!u) return <div className="p-6">Yuklanmoqda…</div>;
  return (
    <div className="min-h-screen p-4 mx-auto max-w-3xl">
      <div className="flex items-center justify-between mb-4">
        <Link href="/" className="heading font-extrabold text-xl">Ishchi Bormi</Link>
        <ScriptToggle />
      </div>
      <div className="card p-6 flex gap-4">
        <div className="h-20 w-20 rounded-full bg-brand-navy text-white grid place-items-center text-2xl font-bold">{u.firstName?.[0]?.toUpperCase()}</div>
        <div className="flex-1">
          <h1 className="text-xl font-bold">{u.firstName} {u.lastName} {u.isPhoneVerified && <CheckCircle2 size={16} className="inline text-success" />}</h1>
          <div className="text-sm text-[color:var(--text-muted)] mt-1">
            <span className="inline-flex items-center gap-1"><Star size={14} className="text-accent-amber" />{u.rating.toFixed(1)} ({u.reviewsCount})</span>
            {" "}• {u.region}{u.district ? `, ${u.district}` : ""}
            {" "}• {u.completedJobsCount} <T>bajarilgan ish</T>
          </div>
          {u.skills && u.skills.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-2">
              {u.skills.map((s, i) => <span key={i} className="chip">{s}</span>)}
            </div>
          )}
          {u.bio && <p className="mt-3 text-sm">{u.bio}</p>}
        </div>
      </div>
      <div className="card p-6 mt-4">
        <h2 className="font-semibold mb-3"><T>Baholar</T></h2>
        {reviews.length === 0 && <div className="text-[color:var(--text-muted)]"><T>Sharhlar yo'q.</T></div>}
        <div className="grid gap-2">
          {reviews.map((r) => (
            <div key={r.id} className="border rounded-lg p-3" style={{ borderColor: "var(--border)" }}>
              <div className="text-accent-amber">{"★".repeat(r.rating)}</div>
              {r.comment && <p className="text-sm">{r.comment}</p>}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
