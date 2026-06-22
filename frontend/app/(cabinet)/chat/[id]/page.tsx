"use client";
import { useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { api, Message, MessageAttachment, User, WS_BASE, getAccess } from "@/lib/api";
import { Shell } from "@/components/Shell";
import { Send, Paperclip, X, Loader2, FileText } from "lucide-react";
import { T, useT } from "@/components/T";
import { uploadFile } from "@/lib/upload";

export default function ChatRoom() {
  const t = useT();
  const { id } = useParams<{ id: string }>();
  const [me, setMe] = useState<User | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState("");
  const [pendingFiles, setPendingFiles] = useState<MessageAttachment[]>([]);
  const [uploading, setUploading] = useState(false);
  const endRef = useRef<HTMLDivElement>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => { api.get<User>("/api/me").then(setMe); }, []);
  useEffect(() => {
    if (!id) return;
    api.get<Message[]>(`/api/conversations/${id}/messages`).then((m) => {
      setMessages(m);
      setTimeout(() => endRef.current?.scrollIntoView({ behavior: "smooth" }), 50);
    });
  }, [id]);

  useEffect(() => {
    const tok = getAccess();
    if (!tok) return;
    const ws = new WebSocket(`${WS_BASE}/api/ws?token=${tok}`);
    ws.onmessage = (ev) => {
      try {
        const env = JSON.parse(ev.data);
        if (env.kind === "message") {
          const m = env.payload as Message;
          if (m.conversationId === id) {
            setMessages((prev) => prev.find((x) => x.id === m.id) ? prev : [...prev, m]);
            setTimeout(() => endRef.current?.scrollIntoView({ behavior: "smooth" }), 30);
          }
        }
      } catch {}
    };
    return () => ws.close();
  }, [id]);

  async function pickFiles(files: FileList | null) {
    if (!files || files.length === 0) return;
    setUploading(true);
    const added: MessageAttachment[] = [];
    for (const f of Array.from(files).slice(0, 4)) {
      try {
        const r = await uploadFile(f, "chat", { scope: String(id) });
        added.push({ url: r.url, name: f.name, size: f.size, mime: f.type });
      } catch {}
    }
    setPendingFiles((prev) => [...prev, ...added]);
    setUploading(false);
  }

  async function send(e: React.FormEvent) {
    e.preventDefault();
    if (!text.trim() && pendingFiles.length === 0) return;
    const body = { text: text.trim(), attachments: pendingFiles };
    setText("");
    setPendingFiles([]);
    await api.post(`/api/conversations/${id}/messages`, body);
  }

  return (
    <Shell title="Suhbat">
      <div className="card flex flex-col h-[calc(100vh-180px)] min-h-[440px] overflow-hidden">
        {/* Messages */}
        <div className="flex-1 overflow-y-auto scroll-y-auto p-4 space-y-2">
          {messages.length === 0 && (
            <div className="grid place-items-center h-full muted text-sm"><T>Suhbat boshlanmagan</T></div>
          )}
          {messages.map((m) => {
            const mine = m.senderId === me?.id;
            return (
              <div key={m.id} className={`flex ${mine ? "justify-end" : "justify-start"} animate-fade-in`}>
                <div className={`max-w-[75%] rounded-2xl px-3.5 py-2 text-sm ${
                  mine
                    ? "bg-brand-navy text-white rounded-br-md"
                    : "bg-[color:var(--bg-subtle)] rounded-bl-md"
                }`}>
                  {m.attachments?.map((a, i) => <Attachment key={i} a={a} mine={mine} />)}
                  {m.text && <div className={m.attachments?.length ? "mt-2" : ""}>{m.text}</div>}
                </div>
              </div>
            );
          })}
          <div ref={endRef} />
        </div>

        {/* Pending file previews */}
        {pendingFiles.length > 0 && (
          <div className="px-3 pb-2 flex gap-2 flex-wrap border-t" style={{ borderColor: "var(--border)" }}>
            {pendingFiles.map((a, i) => (
              <div key={i} className="relative">
                {a.mime?.startsWith("image/")
                  ? <img src={a.url} alt="" className="h-14 w-14 object-cover rounded-lg" />
                  : <div className="h-14 w-14 rounded-lg grid place-items-center surface text-xs muted px-1 text-center truncate">{a.name}</div>
                }
                <button
                  type="button"
                  onClick={() => setPendingFiles((p) => p.filter((_, j) => j !== i))}
                  className="absolute -top-1 -right-1 grid place-items-center h-5 w-5 rounded-full bg-danger text-white"
                >
                  <X size={11} />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Composer */}
        <form onSubmit={send} className="flex gap-2 border-t p-3" style={{ borderColor: "var(--border)" }}>
          <button
            type="button"
            onClick={() => fileRef.current?.click()}
            disabled={uploading}
            className="btn btn-ghost px-3"
            aria-label={t("Fayl biriktirish")}
          >
            {uploading ? <Loader2 size={16} className="animate-spin" /> : <Paperclip size={16} />}
          </button>
          <input
            ref={fileRef}
            type="file"
            multiple
            accept="image/*,application/pdf,application/zip"
            className="hidden"
            onChange={(e) => { pickFiles(e.target.files); e.target.value = ""; }}
          />
          <input
            className="input"
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder={t("Xabar yozing…")}
          />
          <button className="btn btn-primary gap-2" disabled={uploading}>
            <Send size={16} />
          </button>
        </form>
      </div>
    </Shell>
  );
}

function Attachment({ a, mine }: { a: MessageAttachment; mine: boolean }) {
  if (a.mime?.startsWith("image/")) {
    return (
      <a href={a.url} target="_blank" rel="noreferrer" className="block">
        <img src={a.url} alt={a.name || ""} className="max-h-64 rounded-lg object-cover" />
      </a>
    );
  }
  return (
    <a
      href={a.url}
      target="_blank"
      rel="noreferrer"
      className={`inline-flex items-center gap-2 rounded-lg px-3 py-2 text-xs ${
        mine ? "bg-white/15 text-white" : "bg-black/5"
      }`}
    >
      <FileText size={14} />
      <span className="truncate max-w-[180px]">{a.name || "fayl"}</span>
    </a>
  );
}
