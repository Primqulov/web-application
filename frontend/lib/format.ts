import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import "dayjs/locale/uz";

dayjs.extend(relativeTime);
dayjs.locale("uz");

export function fmtSum(n: number): string {
  if (!n) return "0";
  return n.toLocaleString("uz-UZ").replace(/ /g, " ");
}

export function fmtSumSom(n: number, negotiable?: boolean): string {
  if (negotiable) return "Kelishiladi";
  return `${fmtSum(n)} so'm`;
}

export function fromNow(iso: string): string {
  return dayjs(iso).fromNow();
}

export function fmtDate(iso?: string): string {
  if (!iso) return "";
  return dayjs(iso).format("D MMMM YYYY");
}
