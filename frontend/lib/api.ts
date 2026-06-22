"use client";

export const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";
export const WS_BASE = process.env.NEXT_PUBLIC_WS_BASE || "ws://localhost:8080";

const ACCESS_KEY = "ib-access";
const REFRESH_KEY = "ib-refresh";
const ADMIN_KEY = "ib-admin";

export function getAccess(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS_KEY);
}
export function setAccess(t: string | null) {
  if (typeof window === "undefined") return;
  if (t) localStorage.setItem(ACCESS_KEY, t);
  else localStorage.removeItem(ACCESS_KEY);
}
export function setRefresh(t: string | null) {
  if (typeof window === "undefined") return;
  if (t) localStorage.setItem(REFRESH_KEY, t);
  else localStorage.removeItem(REFRESH_KEY);
}
export function getRefresh(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(REFRESH_KEY);
}

export function setAdminToken(t: string | null) {
  if (typeof window === "undefined") return;
  if (t) localStorage.setItem(ADMIN_KEY, t);
  else localStorage.removeItem(ADMIN_KEY);
}
export function getAdminToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ADMIN_KEY);
}

export interface APIError {
  code: string;
  message: string;
}

async function request<T>(
  path: string,
  opts: RequestInit & { auth?: "user" | "admin" | "none" } = {}
): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const auth = opts.auth ?? "user";
  if (auth === "user") {
    const t = getAccess();
    if (t) headers["Authorization"] = `Bearer ${t}`;
  } else if (auth === "admin") {
    const t = getAdminToken();
    if (t) headers["Authorization"] = `Bearer ${t}`;
  }
  Object.assign(headers, opts.headers || {});
  const res = await fetch(`${API_BASE}${path}`, { ...opts, headers });
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;
  if (!res.ok) {
    const err: APIError = (data && data.error) || { code: "http", message: `HTTP ${res.status}` };
    throw err;
  }
  return data as T;
}

export const api = {
  get: <T>(p: string, opts?: any) => request<T>(p, { ...(opts || {}), method: "GET" }),
  post: <T>(p: string, body?: any, opts?: any) =>
    request<T>(p, { ...(opts || {}), method: "POST", body: body !== undefined ? JSON.stringify(body) : undefined }),
  patch: <T>(p: string, body?: any, opts?: any) =>
    request<T>(p, { ...(opts || {}), method: "PATCH", body: body !== undefined ? JSON.stringify(body) : undefined }),
  delete: <T>(p: string, opts?: any) => request<T>(p, { ...(opts || {}), method: "DELETE" }),
};

// ----- domain types -----
export type ID = string;

export interface User {
  id: ID;
  telegramId?: number;
  phone: string;
  firstName: string;
  lastName: string;
  avatarUrl?: string;
  region?: string;
  district?: string;
  bio?: string;
  skills?: string[];
  rating: number;
  reviewsCount: number;
  completedJobsCount: number;
  isPhoneVerified: boolean;
  isPremium: boolean;
  isBlocked: boolean;
  langPref?: "latin" | "cyrillic";
  themePref?: "light" | "dark";
  onboardingCompleted?: boolean;
}
export interface Category {
  id: ID;
  name: string;
  slug: string;
  icon?: string;
  isActive: boolean;
  usageCount: number;
}
export interface Elon {
  id: ID;
  ownerId: ID;
  title: string;
  categoryId: ID;
  categoryName: string;
  description: string;
  locationUrl?: string;
  locationText?: string;
  region?: string;
  district?: string;
  workersNeeded: number;
  pricingType: "per_worker" | "total" | "negotiable";
  priceAmount: number;
  perWorkerAmount: number;
  startDate?: string;
  workTimeFrom?: string;
  workTimeTo?: string;
  contactPhone?: string;
  status: "draft" | "recruiting" | "filled" | "in_progress" | "completed" | "cancelled";
  acceptedCount: number;
  publishedAt?: string;
  createdAt: string;
  ownerName?: string;
  ownerRating?: number;
  images?: string[];
}
export interface Application {
  id: ID;
  elonId: ID;
  elonTitle: string;
  workerId: ID;
  employerId: ID;
  workerPhone: string;
  amount: number;
  isNegotiable: boolean;
  status: "pending" | "accepted" | "rejected" | "cancelled" | "completed";
  employerConfirmedDone?: boolean;
  workerConfirmedDone?: boolean;
  cancelledBy?: string;
  appliedAt: string;
  decidedAt?: string;
  completedAt?: string;
}
export interface Notification {
  id: ID;
  type: string;
  title: string;
  body: string;
  isRead: boolean;
  createdAt: string;
  relatedEntity?: { type: string; id: ID };
}
export interface Conversation {
  id: ID;
  participantIds: ID[];
  lastMessageText: string;
  lastMessageAt: string;
  lastSenderId?: ID;
  unread: Record<string, number>;
}
export interface MessageAttachment {
  url: string;
  name?: string;
  size?: number;
  mime?: string;
}
export interface Message {
  id: ID;
  conversationId: ID;
  senderId: ID;
  text: string;
  attachments?: MessageAttachment[];
  isRead: boolean;
  createdAt: string;
}
export interface FinanceEntry {
  id: ID;
  userId: ID;
  role: "worker" | "employer";
  applicationId: ID;
  elonId: ID;
  elonTitle: string;
  counterpartyId: ID;
  type: "earned" | "spent";
  status: "completed" | "cancelled";
  amount: number;
  isNegotiable: boolean;
  createdAt: string;
}
export interface FinanceSummary {
  earnedSum: number;
  spentSum: number;
  completedCount: number;
  negotiableCount: number;
  cancelledCount: number;
  entries: FinanceEntry[];
}
export interface Review {
  id: ID;
  applicationId: ID;
  elonId: ID;
  fromUserId: ID;
  toUserId: ID;
  direction: "employer_to_worker" | "worker_to_employer";
  rating: number;
  comment?: string;
  createdAt: string;
}
