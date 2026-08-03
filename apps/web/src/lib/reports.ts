import { apiFetch } from "@/lib/api";

type AccessToken = () => Promise<string | null>;

export type ReportType =
  | "residents"
  | "mutations"
  | "households"
  | "invoices"
  | "arrears"
  | "payments"
  | "cash"
  | "letters"
  | "complaints";

export interface ReportFilter {
  start_date?: string;
  end_date?: string;
  status?: string;
}

function buildQuery(filter: ReportFilter = {}, format?: "csv"): string {
  const query = new URLSearchParams();
  if (filter.start_date) query.set("start_date", filter.start_date);
  if (filter.end_date) query.set("end_date", filter.end_date);
  if (filter.status) query.set("status", filter.status);
  if (format) query.set("format", format);

  const value = query.toString();
  return value ? `?${value}` : "";
}

async function authorization(getAccessToken: AccessToken): Promise<HeadersInit> {
  const token = await getAccessToken();
  if (!token) {
    throw new Error("Sesi tidak tersedia.");
  }
  return { Authorization: `Bearer ${token}` };
}

export async function fetchReport<T>(
  type: ReportType,
  filter: ReportFilter = {},
  getAccessToken: AccessToken,
): Promise<T[]> {
  return apiFetch<T[]>(`reports/${type}${buildQuery(filter)}`, {
    headers: await authorization(getAccessToken),
  });
}

export async function downloadReportCSV(
  type: ReportType,
  filter: ReportFilter = {},
  getAccessToken: AccessToken,
  filename: string = type,
): Promise<void> {
  const apiBase = (process.env.NEXT_PUBLIC_API_URL || "/api/v1").replace(/\/+$/, "");
  const response = await fetch(`${apiBase}/reports/${type}${buildQuery(filter, "csv")}`, {
    headers: await authorization(getAccessToken),
    credentials: "include",
    cache: "no-store",
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}));
    throw new Error(payload.error?.message || "Gagal mengunduh berkas CSV laporan.");
  }

  const objectURL = URL.createObjectURL(await response.blob());
  const link = document.createElement("a");
  link.href = objectURL;
  link.download = `${filename}-${new Date().toISOString().slice(0, 10)}.csv`;
  document.body.append(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(objectURL);
}