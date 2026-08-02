"use client";

import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";

type InvoiceStatus =
  | "unpaid"
  | "pending_verification"
  | "partial"
  | "paid"
  | "cancelled";

interface Invoice {
  id: string;
  invoice_number: string;
  household_number: string;
  house_unit_code: string;
  due_type_name: string;
  period_start: string;
  period_end: string;
  due_date: string;
  amount: number;
  paid_amount: number;
  adjustment_amount: number;
  adjustment_reason: string | null;
  status: InvoiceStatus;
  cancelled_at: string | null;
  cancellation_reason: string | null;
}

const status: Record<
  InvoiceStatus,
  { label: string; variant: "neutral" | "warning" | "success" | "danger" }
> = {
  unpaid: { label: "Belum bayar", variant: "warning" },
  pending_verification: { label: "Menunggu verifikasi", variant: "neutral" },
  partial: { label: "Dibayar sebagian", variant: "warning" },
  paid: { label: "Lunas", variant: "success" },
  cancelled: { label: "Dibatalkan", variant: "danger" },
};

const rupiah = (amount: number) =>
  new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);

export default function TagihanWargaPage() {
  const { getAccessToken } = useAuth();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [arrears, setArrears] = useState(false);
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null);

  const loadInvoices = useCallback(async () => {
    setLoading(true);
    setError("");

    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      }

      const query = arrears ? "?arrears=true" : "";
      const data = await apiFetch<Invoice[]>(`invoices${query}`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      setInvoices(data);
    } catch (cause) {
      setError(
        cause instanceof ApiException || cause instanceof Error
          ? cause.message
          : "Gagal memuat tagihan.",
      );
    } finally {
      setLoading(false);
    }
  }, [arrears, getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadInvoices();
  }, [loadInvoices]);

  if (loading) {
    return (
      <p style={{ color: "var(--color-text-secondary)" }}>
        Memuat tagihan…
      </p>
    );
  }

  if (error) {
    return (
      <EmptyState
        title="Gagal memuat tagihan"
        description={error}
        action={
          <Button variant="secondary" onClick={() => void loadInvoices()}>
            Coba lagi
          </Button>
        }
      />
    );
  }

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>
          Tagihan Keluarga
        </h1>
        <p
          style={{
            color: "var(--color-text-secondary)",
            marginTop: "var(--space-2)",
          }}
        >
          Tagihan aktif, riwayat, dan tunggakan keluarga Anda.
        </p>
      </header>

      <section style={{ display: "grid", gap: "var(--space-3)" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            gap: "var(--space-3)",
            alignItems: "center",
          }}
        >
          <h2 style={{ fontSize: "1.125rem" }}>Daftar Tagihan</h2>
          <label>
            <input
              type="checkbox"
              checked={arrears}
              onChange={(event) => setArrears(event.target.checked)}
            />{" "}
            Tunggakan saja
          </label>
        </div>

        {invoices.length === 0 ? (
          <EmptyState
            title="Tidak ada tagihan"
            description={
              arrears
                ? "Anda tidak memiliki tunggakan."
                : "Belum ada tagihan untuk keluarga Anda."
            }
          />
        ) : (
          invoices.map((item) => (
            <article
              key={item.id}
              style={{
                display: "grid",
                gap: "var(--space-2)",
                padding: "var(--space-4)",
                border: "1px solid var(--color-border)",
                borderRadius: "var(--radius-lg)",
                background: "var(--color-surface)",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "flex-start",
                  gap: "var(--space-3)",
                }}
              >
                <strong>{item.due_type_name}</strong>
                <StatusBadge variant={status[item.status].variant}>
                  {status[item.status].label}
                </StatusBadge>
              </div>
              <span
                style={{
                  color: "var(--color-text-secondary)",
                  fontSize: "0.875rem",
                }}
              >
                {item.invoice_number} · Jatuh tempo {item.due_date}
              </span>
              <span>
                {rupiah(item.amount - item.adjustment_amount)} · terbayar{" "}
                {rupiah(item.paid_amount)}
              </span>
              <Button
                variant="secondary"
                onClick={() => setSelectedInvoice(item)}
              >
                Lihat detail
              </Button>
            </article>
          ))
        )}
      </section>

      {selectedInvoice ? (
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="invoice-detail-title"
          style={{
            position: "fixed",
            inset: 0,
            zIndex: 50,
            display: "grid",
            placeItems: "center",
            padding: "var(--space-4)",
            background: "rgba(0, 0, 0, 0.5)",
          }}
        >
          <section
            style={{
              display: "grid",
              width: "min(100%, 32rem)",
              maxHeight: "calc(100vh - var(--space-8))",
              gap: "var(--space-3)",
              overflowY: "auto",
              padding: "var(--space-4)",
              borderRadius: "var(--radius-lg)",
              background: "var(--color-surface)",
            }}
          >
            <h2 id="invoice-detail-title" style={{ fontSize: "1.25rem" }}>
              Detail Tagihan
            </h2>
            <p>
              <strong>Nomor:</strong> {selectedInvoice.invoice_number}
            </p>
            <p>
              <strong>Iuran:</strong> {selectedInvoice.due_type_name}
            </p>
            <p>
              <strong>Unit/Keluarga:</strong> {selectedInvoice.house_unit_code}{" "}
              · {selectedInvoice.household_number}
            </p>
            <p>
              <strong>Periode:</strong> {selectedInvoice.period_start} s.d.{" "}
              {selectedInvoice.period_end}
            </p>
            <p>
              <strong>Jatuh tempo:</strong> {selectedInvoice.due_date}
            </p>
            <p>
              <strong>Nilai tagihan:</strong> {rupiah(selectedInvoice.amount)}
            </p>
            {selectedInvoice.adjustment_amount > 0 ? (
              <p>
                <strong>Diskon/penyesuaian:</strong>{" "}
                {rupiah(selectedInvoice.adjustment_amount)}
                {selectedInvoice.adjustment_reason
                  ? ` · ${selectedInvoice.adjustment_reason}`
                  : ""}
              </p>
            ) : null}
            <p>
              <strong>Total dibayar:</strong>{" "}
              {rupiah(selectedInvoice.paid_amount)}
            </p>
            <p>
              <strong>Status:</strong>{" "}
              <StatusBadge variant={status[selectedInvoice.status].variant}>
                {status[selectedInvoice.status].label}
              </StatusBadge>
            </p>
            {selectedInvoice.cancelled_at ? (
              <p>
                <strong>Pembatalan:</strong>{" "}
                {selectedInvoice.cancellation_reason ?? "Tidak tersedia"}
              </p>
            ) : null}
            <Button onClick={() => setSelectedInvoice(null)}>Tutup</Button>
          </section>
        </div>
      ) : null}
    </div>
  );
}