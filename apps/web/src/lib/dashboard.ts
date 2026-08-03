import { apiFetch } from "@/lib/api";

export interface ResidentDashboard {
  active_invoices: InvoiceSummary[];
  recent_payments: PaymentSummary[];
  recent_letters: LetterSummary[];
  recent_complaints: ComplaintSummary[];
  announcements: AnnouncementSummary[];
  upcoming_events: EventSummary[];
}

export interface AdminDashboard {
  active_households: number;
  active_residents: number;
  active_invoices: number;
  outstanding_amount: number;
  pending_payments: number;
  cash_balance: number;
  pending_letters: number;
  open_complaints: number;
  recent_payments: PaymentSummary[];
  recent_letters: LetterSummary[];
  recent_complaints: ComplaintSummary[];
}

export interface InvoiceSummary {
  id: string;
  invoice_number: string;
  due_type_name: string;
  amount: number;
  paid_amount: number;
  due_date: string;
  status: string;
}

export interface PaymentSummary {
  id: string;
  payment_number: string;
  invoice_number: string;
  amount: number;
  paid_at: string;
  verification_status: string;
}

export interface LetterSummary {
  id: string;
  request_number: string;
  letter_type: string;
  status: string;
  updated_at: string;
}

export interface ComplaintSummary {
  id: string;
  ticket_number: string;
  title: string;
  priority: string;
  status: string;
  updated_at: string;
}

export interface AnnouncementSummary {
  id: string;
  title: string;
  category: string;
  priority: string;
  publish_at: string;
}

export interface EventSummary {
  id: string;
  title: string;
  starts_at: string;
  location: string;
}

type AccessToken = () => Promise<string | null>;

async function authorizedFetch<T>(path: string, getAccessToken: AccessToken): Promise<T> {
  const token = await getAccessToken();
  if (!token) {
    throw new Error("Sesi tidak tersedia.");
  }
  return apiFetch<T>(path, { headers: { Authorization: `Bearer ${token}` } });
}

export function getResidentDashboard(getAccessToken: AccessToken): Promise<ResidentDashboard> {
  return authorizedFetch("dashboard/resident", getAccessToken);
}

export function getAdminDashboard(getAccessToken: AccessToken): Promise<AdminDashboard> {
  return authorizedFetch("dashboard/admin", getAccessToken);
}