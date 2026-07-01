import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/uz";

dayjs.extend(relativeTime);
dayjs.locale("uz");

// fmtSum minglik xonalarni probel bilan ajratadi: 150000 -> "150 000".
// toLocaleString muhitga bog'liq bo'lgani uchun qo'lda guruhlaymiz.
export function fmtSum(n: number): string {
  if (!n && n !== 0) return "0";
  const neg = n < 0;
  const s = Math.abs(Math.trunc(n)).toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  return neg ? "-" + s : s;
}

export function fmtSumSom(n: number, negotiable?: boolean): string {
  if (negotiable) return "Kelishiladi";
  return `${fmtSum(n)} so'm`;
}

// onlyDigits faqat raqamlarni qoldiradi (harf/belgilarni olib tashlaydi).
export function onlyDigits(s: string): string {
  return (s || "").replace(/\D/g, "");
}

// fmtThousands input uchun: matndagi raqamlarni "150 000" ko'rinishida qaytaradi.
export function fmtThousands(s: string): string {
  const d = onlyDigits(s).replace(/^0+(?=\d)/, "");
  if (!d) return "";
  return d.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}

// fmtPhone O'zbekiston raqamini "+998 90 020 25 35" ko'rinishida formatlaydi.
export function fmtPhone(raw: string): string {
  let d = onlyDigits(raw);
  if (d.startsWith("998")) d = d.slice(3);
  d = d.slice(0, 9); // 2 + 3 + 2 + 2
  const parts = [d.slice(0, 2), d.slice(2, 5), d.slice(5, 7), d.slice(7, 9)].filter(Boolean);
  return "+998" + (parts.length ? " " + parts.join(" ") : "");
}

// phoneDigits: faqat 9 xonali milliy qism (998siz).
export function phoneDigits(raw: string): string {
  let d = onlyDigits(raw);
  if (d.startsWith("998")) d = d.slice(3);
  return d.slice(0, 9);
}

export function fromNow(iso: string): string {
  return dayjs(iso).fromNow();
}

export function fmtDate(iso?: string): string {
  if (!iso) return "";
  return dayjs(iso).format("D MMMM YYYY");
}
