"use client";
import Link from "next/link";
import {
  ShieldCheck, Database, Lock, Share2, UserCog, Cookie, Baby,
  RefreshCw, Mail, Phone, Send, MapPin, FileText, AlertCircle,
  LifeBuoy, Instagram, Youtube,
} from "lucide-react";
import { CONTACT, SOCIAL } from "@/lib/contact";
import { ScriptToggle } from "@/components/ScriptToggle";
import { ThemeToggle } from "@/components/ThemeToggle";
import { T } from "@/components/T";
import { getAccess } from "@/lib/api";

const SECTIONS = [
  { id: "intro",     label: "Kirish",                       icon: FileText },
  { id: "collect",   label: "Qaysi ma'lumotlarni yig'amiz", icon: Database },
  { id: "use",       label: "Ma'lumotlardan foydalanish",   icon: UserCog },
  { id: "share",     label: "Uchinchi shaxslar bilan",      icon: Share2 },
  { id: "security",  label: "Xavfsizlik",                   icon: Lock },
  { id: "rights",    label: "Sizning huquqlaringiz",        icon: ShieldCheck },
  { id: "cookies",   label: "Cookie va texnologiyalar",     icon: Cookie },
  { id: "children",  label: "Bolalar maxfiyligi",           icon: Baby },
  { id: "changes",   label: "O'zgartirishlar",              icon: RefreshCw },
  { id: "contact",   label: "Aloqa",                        icon: Mail },
];

export default function PrivacyPage() {
  const ctaHref = getAccess() ? "/dashboard" : "/login";

  return (
    <div className="min-h-screen flex flex-col">
      {/* ── Header ─────────────────── */}
      <header className="border-b" style={{ borderColor: "var(--border)", background: "var(--card)" }}>
        <div className="mx-auto max-w-6xl flex items-center justify-between px-4 py-3">
          <Link href="/" className="font-extrabold text-xl heading">Ishchi Bormi</Link>
          <div className="flex items-center gap-3">
            <ScriptToggle />
            <ThemeToggle />
            <Link href={ctaHref} className="btn-primary"><T>Kirish</T></Link>
          </div>
        </div>
      </header>

      {/* ── Hero ───────────────────── */}
      <section className="px-4 pt-6">
        <div className="mx-auto max-w-6xl card p-8 sm:p-12 text-center">
          <div className="mx-auto h-14 w-14 grid place-items-center rounded-2xl bg-brand-navy text-white">
            <ShieldCheck size={26} />
          </div>
          <h1 className="mt-4 text-2xl sm:text-3xl font-extrabold heading">
            <T>Maxfiylik siyosati</T>
          </h1>
          <p className="mt-2 text-sm muted max-w-2xl mx-auto">
            <T>Ishchi Bormi sizning shaxsiy ma'lumotlaringizni qanday yig'ishi, ishlatishi va himoya qilishini batafsil tushuntiramiz.</T>
          </p>
          <p className="mt-3 text-xs muted"><T>Oxirgi yangilanish</T>: 22.06.2026</p>
        </div>
      </section>

      <main className="flex-1 mx-auto max-w-6xl w-full px-4 mt-6 pb-12 grid lg:grid-cols-[260px_1fr] gap-6">
        {/* ── TOC ─────────────────── */}
        <aside className="card p-4 self-start lg:sticky lg:top-4">
          <h3 className="text-xs uppercase tracking-wider muted px-2"><T>Mundarija</T></h3>
          <nav className="mt-3 grid gap-1">
            {SECTIONS.map(({ id, label, icon: Icon }) => (
              <a key={id} href={`#${id}`} className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm hover:bg-black/5">
                <Icon size={15} className="muted" />
                <span><T>{label}</T></span>
              </a>
            ))}
          </nav>
        </aside>

        {/* ── Content ─────────────── */}
        <article className="grid gap-4">
          <Note>
            <T>Sizning ishonchingiz biz uchun muhim. Quyida bizning maxfiylikka oid majburiyatlarimiz keltirilgan. Iltimos, diqqat bilan o'qing.</T>
          </Note>

          <Sec id="intro" icon={<FileText size={18} />} title="Kirish">
            <P>
              <T>Ushbu Maxfiylik siyosati Ishchi Bormi platformasi ("Biz", "Platforma") tomonidan foydalanuvchilarning shaxsiy ma'lumotlarini qayta ishlash tartibini belgilaydi. Platformadan foydalanish orqali siz ushbu siyosat shartlariga rozilik bildirasiz.</T>
            </P>
          </Sec>

          <Sec id="collect" icon={<Database size={18} />} title="Qaysi ma'lumotlarni yig'amiz">
            <P><T>Biz quyidagi shaxsiy ma'lumotlarni yig'amiz:</T></P>
            <ul className="mt-2 space-y-1.5 text-sm muted list-disc pl-5">
              <li><b className="heading"><T>Hisob ma'lumotlari</T>:</b> <T>telefon raqami, Telegram identifikatori, ism va familiya.</T></li>
              <li><b className="heading"><T>Profil ma'lumotlari</T>:</b> <T>avatar, viloyat va tuman, qisqacha bio, ko'nikmalar.</T></li>
              <li><b className="heading"><T>Faoliyat</T>:</b> <T>e'lonlar, arizalar, baholar, sharhlar, xabarlar va moliyaviy yozuvlar.</T></li>
              <li><b className="heading"><T>Texnik ma'lumot</T>:</b> <T>brauzer turi, IP manzil, qurilma turi va kirish vaqti (xavfsizlik uchun).</T></li>
            </ul>
          </Sec>

          <Sec id="use" icon={<UserCog size={18} />} title="Ma'lumotlardan foydalanish">
            <P><T>Yig'ilgan ma'lumotlardan quyidagi maqsadlarda foydalanamiz:</T></P>
            <ul className="mt-2 space-y-1.5 text-sm muted list-disc pl-5">
              <li><T>Hisobingizni yaratish va Telegram OTP orqali tasdiqlash.</T></li>
              <li><T>Sizga mos e'lonlar va ishchilarni topishingizga yordam berish.</T></li>
              <li><T>Bildirishnomalar yuborish (ariza qabul qilindi, yangi xabar, va h.k.).</T></li>
              <li><T>Platforma xavfsizligini ta'minlash va firibgarlikning oldini olish.</T></li>
              <li><T>Xizmatlarimizni doimiy ravishda yaxshilash uchun statistik tahlil.</T></li>
            </ul>
          </Sec>

          <Sec id="share" icon={<Share2 size={18} />} title="Uchinchi shaxslar bilan bo'lishish">
            <P>
              <T>Biz sizning shaxsiy ma'lumotlaringizni sotmaymiz va reklama maqsadida uchinchi shaxslarga bermaymiz. Faqat quyidagi hollarda ma'lumot oshkor qilinishi mumkin:</T>
            </P>
            <ul className="mt-2 space-y-1.5 text-sm muted list-disc pl-5">
              <li><T>Boshqa foydalanuvchilar bilan — ariza topshirganingizda telefon raqamingiz e'lon egasiga, e'lon egasining raqami arizangiz qabul qilinganda sizga ko'rsatiladi.</T></li>
              <li><T>Qonuniy talab asosida — vakolatli davlat organlari rasmiy so'rovi bo'yicha.</T></li>
              <li><T>Texnik xizmat ko'rsatuvchilar bilan — faqat zaruriy doirada (masalan, server xosting).</T></li>
            </ul>
          </Sec>

          <Sec id="security" icon={<Lock size={18} />} title="Xavfsizlik choralari">
            <P>
              <T>Ma'lumotlaringizni himoyalash uchun zamonaviy shifrlash, autentifikatsiya (JWT), HTTPS aloqasi, parollarni bcrypt orqali xeshlash va kirish huquqlarini cheklash kabi texnologiyalardan foydalanamiz.</T>
            </P>
            <P>
              <T>Shunga qaramay, hech bir tizim 100% xavfsiz emas. Iltimos, parol va kodlaringizni hech kim bilan baham ko'rmang.</T>
            </P>
          </Sec>

          <Sec id="rights" icon={<ShieldCheck size={18} />} title="Sizning huquqlaringiz">
            <P><T>Sizda quyidagi huquqlar mavjud:</T></P>
            <ul className="mt-2 space-y-1.5 text-sm muted list-disc pl-5">
              <li><T>Ma'lumotlaringizga kirish va ularni ko'rib chiqish.</T></li>
              <li><T>Ma'lumotlarni tahrirlash yoki yangilash.</T></li>
              <li><T>Hisobingizni o'chirish (Sozlamalar bo'limidan).</T></li>
              <li><T>Bildirishnomalardan voz kechish.</T></li>
              <li><T>Boshqa foydalanuvchini bloklash yoki shikoyat qilish.</T></li>
            </ul>
          </Sec>

          <Sec id="cookies" icon={<Cookie size={18} />} title="Cookie va shunga o'xshash texnologiyalar">
            <P>
              <T>Saytimiz hisobni eslab qolish va sozlamalarni saqlash uchun localStorage va sessiya tokenlaridan foydalanadi. Bu fayllar reklama maqsadida ishlatilmaydi.</T>
            </P>
          </Sec>

          <Sec id="children" icon={<Baby size={18} />} title="Bolalar maxfiyligi">
            <P>
              <T>Platforma 18 yoshdan kichik foydalanuvchilarga mo'ljallanmagan. Agar bolaning ma'lumoti tasodifan yig'ilgani aniqlansa, biz uni darhol o'chiramiz.</T>
            </P>
          </Sec>

          <Sec id="changes" icon={<RefreshCw size={18} />} title="Siyosatga o'zgartirishlar">
            <P>
              <T>Biz vaqti-vaqti bilan ushbu siyosatni yangilab turishimiz mumkin. Muhim o'zgarishlar bo'lganda sizni bildirishnoma orqali xabardor qilamiz.</T>
            </P>
          </Sec>

          <Sec id="contact" icon={<Mail size={18} />} title="Biz bilan bog'lanish">
            <P><T>Savollar yoki shikoyatlar yuzasidan biz bilan quyidagi yo'llar orqali bog'lanishingiz mumkin:</T></P>
            <div className="mt-3 grid sm:grid-cols-3 gap-3 text-sm">
              <Contact icon={<Mail size={16} />} text={CONTACT.email} href={CONTACT.emailHref} />
              <Contact icon={<Phone size={16} />} text={CONTACT.phone} href={CONTACT.phoneHref} />
              <Contact icon={<Send size={16} />} text={SOCIAL.telegram.label} href={SOCIAL.telegram.href} />
              <Contact icon={<LifeBuoy size={16} />} text={SOCIAL.support.label} href={SOCIAL.support.href} />
              <Contact icon={<Instagram size={16} />} text={SOCIAL.instagram.label} href={SOCIAL.instagram.href} />
              <Contact icon={<Youtube size={16} />} text={SOCIAL.youtube.label} href={SOCIAL.youtube.href} />
            </div>
          </Sec>

          <div className="card p-5 mt-2 text-center">
            <p className="text-sm muted">
              <T>Bu siyosatni o'qib chiqqaningiz uchun rahmat. Bizning maqsadimiz — sizning ma'lumotlaringizni xavfsiz saqlash.</T>
            </p>
            <Link href={ctaHref} className="btn-primary mt-4"><T>Davom etish</T></Link>
          </div>
        </article>
      </main>

      {/* ── Footer ─────────────── */}
      <footer className="mt-auto border-t" style={{ borderColor: "var(--border)", background: "var(--card)" }}>
        <div className="mx-auto max-w-6xl px-4 py-6 grid md:grid-cols-2 gap-4 text-sm">
          <div className="flex items-center gap-2 muted">
            <MapPin size={14} /><T>Toshkent sh., Yunusobod tumani</T>
            <span className="ml-auto md:ml-0">· © 2026 Ishchi Bormi</span>
          </div>
          <div className="flex md:justify-end gap-5 muted">
            <Link href="/biz-haqimizda"><T>Biz haqimizda</T></Link>
            <Link href="/maxfiylik-siyosati" className="heading"><T>Maxfiylik siyosati</T></Link>
            <Link href="/foydalanish-shartlari"><T>Foydalanish shartlari</T></Link>
            <Link href="/yordam"><T>Yordam</T></Link>
          </div>
        </div>
      </footer>
    </div>
  );
}

/* ── helpers ───────────────────── */

function Sec({ id, icon, title, children }: { id: string; icon: React.ReactNode; title: string; children: React.ReactNode }) {
  return (
    <section id={id} className="card p-6 scroll-mt-20">
      <h2 className="font-semibold heading flex items-center gap-2 text-lg">
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-navy text-white">{icon}</span>
        <T>{title}</T>
      </h2>
      <div className="mt-3">{children}</div>
    </section>
  );
}

function P({ children }: { children: React.ReactNode }) {
  return <p className="text-sm leading-relaxed muted">{children}</p>;
}

function Note({ children }: { children: React.ReactNode }) {
  return (
    <div className="card p-4 flex gap-3" style={{ background: "rgba(232,146,10,0.08)" }}>
      <AlertCircle size={18} className="text-accent-amber shrink-0 mt-0.5" />
      <p className="text-sm muted">{children}</p>
    </div>
  );
}

function Contact({ icon, text, href }: { icon: React.ReactNode; text: string; href?: string }) {
  const inner = (
    <>
      <span className="text-accent-amber">{icon}</span>
      <span className="text-sm break-all">{text}</span>
    </>
  );
  const cls = "rounded-xl border p-3 flex items-center gap-2";
  return href ? (
    <a href={href} target={href.startsWith("http") ? "_blank" : undefined} rel="noreferrer"
       className={`${cls} hover:shadow-md transition`} style={{ borderColor: "var(--border)" }}>{inner}</a>
  ) : (
    <div className={cls} style={{ borderColor: "var(--border)" }}>{inner}</div>
  );
}
