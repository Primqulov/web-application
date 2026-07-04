// Butun sayt uchun HAQIQIY aloqa va ijtimoiy tarmoq ma'lumotlari — yagona manba (DRY).
// O'zgartirish kerak bo'lsa faqat shu yerni yangilang; barcha sahifalar shundan oladi.

export const CONTACT = {
  phone: "+998 90 020 25 35",
  phoneHref: "tel:+998900202535",
  email: "ishchibormi@gmail.com",
  emailHref: "mailto:ishchibormi@gmail.com",
} as const;

export type SocialLink = { label: string; href: string };

export const SOCIAL = {
  // Rasmiy Telegram kanali
  telegram: { label: "@Ishchibormi", href: "https://t.me/Ishchibormi" },
  // Qo'llab-quvvatlash (Telegram)
  support: { label: "@Ishchi_bormi_support", href: "https://t.me/Ishchi_bormi_support" },
  instagram: { label: "@ishchi_bormi", href: "https://instagram.com/ishchi_bormi" },
  youtube: { label: "@Ishchi_bormi", href: "https://youtube.com/@Ishchi_bormi" },
} as const satisfies Record<string, SocialLink>;

// Google/schema.org "sameAs" uchun ochiq ijtimoiy tarmoq profillari.
export const SOCIAL_SAMEAS: string[] = [
  SOCIAL.telegram.href,
  SOCIAL.instagram.href,
  SOCIAL.youtube.href,
];
