"use client";
import Link from "next/link";
import { Star, MapPin, Clock, Share2 } from "lucide-react";
import { Elon } from "@/lib/api";
import { fmtSumSom, fromNow } from "@/lib/format";
import { T } from "./T";

export function JobCard({ e }: { e: Elon }) {
  const negotiable = e.pricingType === "negotiable";
  return (
    <Link href={`/elon/${e.id}`} className="card block p-4 hover:shadow-md transition">
      <div className="flex items-start justify-between gap-2">
        <h3 className="font-semibold text-base"><T>{e.title}</T></h3>
        <button
          onClick={(ev) => { ev.preventDefault(); navigator.share?.({ url: location.origin + `/elon/${e.id}`, title: e.title }).catch(() => {}); }}
          className="text-accent-amber hover:opacity-80"
          aria-label="Ulashish"
        >
          <Share2 size={18} />
        </button>
      </div>
      <div className="mt-1 flex items-center gap-2 text-sm text-[color:var(--text-muted)]">
        <Star size={14} className="text-accent-amber" />
        <span>{(e.ownerRating ?? 0).toFixed(1)}</span>
        <span>•</span>
        <span>{e.ownerName || "Foydalanuvchi"}</span>
      </div>
      <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-[color:var(--text-muted)]">
        {e.locationText && (<span className="inline-flex items-center gap-1"><MapPin size={14} />{e.locationText}</span>)}
        {e.publishedAt && (<span className="inline-flex items-center gap-1"><Clock size={14} />{fromNow(e.publishedAt)}</span>)}
      </div>
      <div className="mt-3 flex items-center justify-between">
        <span className="text-xs uppercase text-[color:var(--text-muted)]">
          <T>{e.categoryName || ""}</T>
        </span>
        <span className="font-semibold text-accent-amber">
          {fmtSumSom(e.perWorkerAmount || e.priceAmount, negotiable)}
        </span>
      </div>
    </Link>
  );
}
