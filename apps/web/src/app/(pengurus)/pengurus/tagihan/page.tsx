"use client";

import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";

import { PaymentQueue } from "./PaymentQueue";

type Frequency = "once" | "monthly" | "quarterly" | "yearly";
type InvoiceStatus = "unpaid" | "pending_verification" | "partial" | "paid" | "cancelled";

interface DueType {
  id: string;
  name: string;
  description: string | null;
  amount: number | null;
  frequency: Frequency;
  due_day: number | null;
  status: "active" | "inactive";
}

interface Household {
  id: string;
  internal_number: string;
}

interface Invoice {
  id: string;
  invoice_number: string;
  household_number: string;
  house_unit_code: string;
  due_type_name: string;
  due_date: string;
  amount: number;
  paid_amount: number;
  adjustment_amount: number;
  status: InvoiceStatus;
}

const status: Record<InvoiceStatus, { label: string; variant: "neutral" | "warning" | "success" | "danger" }> = {
  unpaid: { label: "Belum bayar", variant: "warning" },
  pending_verification: { label: "Menunggu verifikasi", variant: "neutral" },
  partial: { label: "Dibayar sebagian", variant: "warning" },
  paid: { label: "Lunas", variant: "success" },
  cancelled: { label: "Dibatalkan", variant: "danger" },
};

const rupiah = (amount: number) =>
  new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(amount);

export default function TagihanPengurusPage() {
  const { getAccessToken } = useAuth();
  const [dueTypes, setDueTypes] = useState<DueType[]>([]);
  const [households, setHouseholds] = useState<Household[]>([]);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [arrears, setArrears] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [dueName, setDueName] = useState("");
  const [dueAmount, setDueAmount] = useState("");
  const [dueFrequency, setDueFrequency] = useState<Frequency>("monthly");
  const [dueDay, setDueDay] = useState("");
  const [householdID, setHouseholdID] = useState("");
  const [dueTypeID, setDueTypeID] = useState("");
  const [periodStart, setPeriodStart] = useState("");
  const [periodEnd, setPeriodEnd] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [amount, setAmount] = useState("");
  const [cancelID, setCancelID] = useState<string | null>(null);
  const [cancelReason, setCancelReason] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      const headers = { Authorization: `Bearer ${token}` };
      const [types, families, bills] = await Promise.all([
        apiFetch<DueType[]>("due-types", { headers }),
        apiFetch<Household[]>("households", { headers }),
        apiFetch<Invoice[]>(`invoices${arrears ? "?arrears=true" : ""}`, { headers }),
      ]);
      setDueTypes(types);
      setHouseholds(families);
      setInvoices(bills);
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat data iuran.");
    } finally {
      setLoading(false);
    }
  }, [arrears, getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const request = async (path: string, body: object, headers: HeadersInit = {}) => {
    const token = await getAccessToken();
    if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
    return apiFetch(path, { method: "POST", headers: { Authorization: `Bearer ${token}`, ...headers }, body: JSON.stringify(body) });
  };

  const createDueType = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true); setError(""); setMessage("");
    try {
      await request("due-types", {
        name: dueName.trim(), amount: dueAmount ? Number(dueAmount) : undefined,
        frequency: dueFrequency, due_day: dueDay ? Number(dueDay) : undefined,
      });
      setDueName(""); setDueAmount(""); setDueDay(""); setMessage("Jenis iuran ditambahkan."); await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal menambah iuran.");
    } finally { setSaving(false); }
  };

  const createInvoice = async (event: React.FormEvent, bulk: boolean) => {
    event.preventDefault();
    setSaving(true); setError(""); setMessage("");
    try {
      const body = { due_type_id: dueTypeID, period_start: periodStart, period_end: periodEnd, due_date: dueDate, amount: amount ? Number(amount) : undefined };
      if (bulk) {
        const result = await request("invoices/generate", { ...body, household_ids: [] }, { "Idempotency-Key": crypto.randomUUID() }) as { total_created: number; total_skipped: number };
        setMessage(`${result.total_created} tagihan dibuat, ${result.total_skipped} dilewati.`);
      } else {
        await request("invoices", { ...body, household_id: householdID });
        setMessage("Tagihan individual diterbitkan.");
      }
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal menerbitkan tagihan.");
    } finally { setSaving(false); }
  };

  const cancelInvoice = async () => {
    if (!cancelID || !cancelReason.trim()) return;
    setSaving(true); setError("");
    try {
      await request(`invoices/${cancelID}/cancel`, { reason: cancelReason.trim() });
      setCancelID(null); setCancelReason(""); setMessage("Tagihan dibatalkan."); await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal membatalkan tagihan.");
    } finally { setSaving(false); }
  };

  const activeTypes = dueTypes.filter((item) => item.status === "active");

  return <div style={{ display: "grid", gap: "var(--space-6)" }}>
    <header><h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Iuran & Tagihan</h1><p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>Jenis iuran, penerbitan tagihan, dan tunggakan keluarga.</p></header>
    {message ? <p role="status" style={{ color: "var(--color-success-text)" }}>{message}</p> : null}
    {error ? <p role="alert" style={{ color: "var(--color-danger-text)" }}>{error}</p> : null}

    <PaymentQueue onActionComplete={() => void load()} />

    <form onSubmit={(event) => void createDueType(event)} style={{ display: "grid", gap: "var(--space-3)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)" }}>
      <h2 style={{ fontSize: "1.125rem" }}>Tambah Jenis Iuran</h2>
      <FormField label="Nama iuran">{({ id }) => <TextInput id={id} value={dueName} onChange={(event) => setDueName(event.target.value)} required />}</FormField>
      <FormField label="Nominal tetap (opsional)">{({ id }) => <TextInput id={id} type="number" min="1" value={dueAmount} onChange={(event) => setDueAmount(event.target.value)} />}</FormField>
      <FormField label="Frekuensi">{({ id }) => <Select id={id} value={dueFrequency} onChange={(event) => setDueFrequency(event.target.value as Frequency)}><option value="monthly">Bulanan</option><option value="quarterly">Triwulan</option><option value="yearly">Tahunan</option><option value="once">Sekali bayar</option></Select>}</FormField>
      <FormField label="Tanggal jatuh tempo 1–31 (opsional)">{({ id }) => <TextInput id={id} type="number" min="1" max="31" value={dueDay} onChange={(event) => setDueDay(event.target.value)} />}</FormField>
      <Button type="submit" disabled={saving}>{saving ? "Menyimpan…" : "Simpan iuran"}</Button>
    </form>

    <section style={{ display: "grid", gap: "var(--space-3)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)" }}>
      <h2 style={{ fontSize: "1.125rem" }}>Terbitkan Tagihan</h2>
      <form onSubmit={(event) => void createInvoice(event, false)} style={{ display: "grid", gap: "var(--space-3)" }}>
        <FormField label="Keluarga">{({ id }) => <Select id={id} value={householdID} onChange={(event) => setHouseholdID(event.target.value)} required><option value="">Pilih keluarga</option>{households.map((item) => <option key={item.id} value={item.id}>{item.internal_number}</option>)}</Select>}</FormField>
        <InvoiceFields dueTypes={activeTypes} dueTypeID={dueTypeID} setDueTypeID={setDueTypeID} periodStart={periodStart} setPeriodStart={setPeriodStart} periodEnd={periodEnd} setPeriodEnd={setPeriodEnd} dueDate={dueDate} setDueDate={setDueDate} amount={amount} setAmount={setAmount} />
        <Button type="submit" disabled={saving}>Buat individual</Button>
      </form>
      <form onSubmit={(event) => void createInvoice(event, true)} style={{ display: "grid", gap: "var(--space-3)", borderTop: "1px solid var(--color-border)", paddingTop: "var(--space-4)" }}>
        <strong>Tagihan massal</strong>
        <InvoiceFields dueTypes={activeTypes} dueTypeID={dueTypeID} setDueTypeID={setDueTypeID} periodStart={periodStart} setPeriodStart={setPeriodStart} periodEnd={periodEnd} setPeriodEnd={setPeriodEnd} dueDate={dueDate} setDueDate={setDueDate} amount={amount} setAmount={setAmount} />
        <Button type="submit" disabled={saving}>Terbitkan ke semua keluarga</Button>
      </form>
    </section>

    <section style={{ display: "grid", gap: "var(--space-3)" }}>
      <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)", alignItems: "center" }}><h2 style={{ fontSize: "1.125rem" }}>Daftar Tagihan</h2><label><input type="checkbox" checked={arrears} onChange={(event) => setArrears(event.target.checked)} /> Tunggakan saja</label></div>
      {loading ? <p>Memuat tagihan…</p> : invoices.length === 0 ? <EmptyState title="Tidak ada tagihan" description="Belum ada tagihan pada filter ini." /> : invoices.map((item) => <article key={item.id} style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)" }}>
        <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}><strong>{item.invoice_number} · {item.due_type_name}</strong><StatusBadge variant={status[item.status].variant}>{status[item.status].label}</StatusBadge></div>
        <span style={{ color: "var(--color-text-secondary)" }}>Unit {item.house_unit_code} · KK {item.household_number} · Jatuh tempo {item.due_date}</span>
        <span>{rupiah(item.amount - item.adjustment_amount)} · terbayar {rupiah(item.paid_amount)}</span>
        {item.status === "unpaid" ? <Button variant="secondary" onClick={() => setCancelID(item.id)}>Batalkan</Button> : null}
        {cancelID === item.id ? <div style={{ display: "flex", gap: "var(--space-2)", flexWrap: "wrap" }}><TextInput aria-label="Alasan pembatalan" value={cancelReason} onChange={(event) => setCancelReason(event.target.value)} /><Button variant="danger" disabled={saving} onClick={() => void cancelInvoice()}>Konfirmasi</Button><Button variant="secondary" onClick={() => setCancelID(null)}>Tutup</Button></div> : null}
      </article>)}
    </section>
  </div>;
}

function InvoiceFields(props: { dueTypes: DueType[]; dueTypeID: string; setDueTypeID: (value: string) => void; periodStart: string; setPeriodStart: (value: string) => void; periodEnd: string; setPeriodEnd: (value: string) => void; dueDate: string; setDueDate: (value: string) => void; amount: string; setAmount: (value: string) => void }) {
  return <><FormField label="Jenis iuran">{({ id }) => <Select id={id} value={props.dueTypeID} onChange={(event) => props.setDueTypeID(event.target.value)} required><option value="">Pilih iuran</option>{props.dueTypes.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</Select>}</FormField><FormField label="Awal periode">{({ id }) => <TextInput id={id} type="date" value={props.periodStart} onChange={(event) => props.setPeriodStart(event.target.value)} required />}</FormField><FormField label="Akhir periode">{({ id }) => <TextInput id={id} type="date" value={props.periodEnd} onChange={(event) => props.setPeriodEnd(event.target.value)} required />}</FormField><FormField label="Jatuh tempo">{({ id }) => <TextInput id={id} type="date" value={props.dueDate} onChange={(event) => props.setDueDate(event.target.value)} required />}</FormField><FormField label="Nominal khusus (opsional)">{({ id }) => <TextInput id={id} type="number" min="1" value={props.amount} onChange={(event) => props.setAmount(event.target.value)} />}</FormField></>;
}
