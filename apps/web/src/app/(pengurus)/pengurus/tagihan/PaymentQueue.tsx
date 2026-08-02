"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

type PaymentStatus = "pending" | "verified" | "rejected" | "cancelled";

interface Payment {
  id: string;
  invoice_number: string;
  payment_number: string;
  method: string;
  amount: number;
  paid_at: string;
  proof_file_id: string | null;
  verification_status: PaymentStatus;
  rejection_reason: string | null;
  cancellation_reason: string | null;
}

type Action = "verify" | "reject" | "cancel";

const methodLabels: Record<string, string> = {
  cash: "Tunai",
  transfer: "Transfer",
  other: "Lainnya",
};

const statusLabels: Record<
  PaymentStatus,
  { label: string; variant: "neutral" | "warning" | "success" | "danger" }
> = {
  pending: { label: "Menunggu", variant: "neutral" },
  verified: { label: "Diterima", variant: "success" },
  rejected: { label: "Ditolak", variant: "danger" },
  cancelled: { label: "Dibatalkan", variant: "danger" },
};

const rupiah = (amount: number) =>
  new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);

export function PaymentQueue({ onActionComplete }: { onActionComplete: () => void }) {
  const { getAccessToken } = useAuth();
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [action, setAction] = useState<{ id: string; type: Action }>();
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");

      const data = await apiFetch<Payment[]>("payments", {
        headers: { Authorization: `Bearer ${token}` },
      });
      setPayments(data);
    } catch (cause) {
      setError(
        cause instanceof ApiException || cause instanceof Error
          ? cause.message
          : "Gagal memuat antrean pembayaran.",
      );
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const closeAction = () => {
    setAction(undefined);
    setReason("");
  };

  const downloadProof = async (fileID: string) => {
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");

      const response = await apiFetch<{ download_url: string }>(`files/${fileID}/download`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      window.open(response.download_url, "_blank", "noopener,noreferrer");
    } catch (cause) {
      setError(
        cause instanceof ApiException || cause instanceof Error
          ? cause.message
          : "Gagal mengunduh bukti pembayaran.",
      );
    }
  };

  const submitAction = async () => {
    if (!action) return;
    if ((action.type === "reject" || action.type === "cancel") && !reason.trim()) {
      setError("Alasan wajib diisi.");
      return;
    }

    setSubmitting(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");

      const body =
        action.type === "verify"
          ? { note: "Diverifikasi bendahara." }
          : { reason: reason.trim() };
      await apiFetch(`payments/${action.id}/${action.type}`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify(body),
      });
      closeAction();
      await load();
      onActionComplete();
    } catch (cause) {
      setError(
        cause instanceof ApiException || cause instanceof Error
          ? cause.message
          : "Gagal memproses pembayaran.",
      );
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <p style={{ color: "var(--color-text-secondary)", margin: 0 }}>Memuat antrean pembayaran…</p>;
  }

  return (
    <section style={{ display: "grid", gap: "var(--space-3)" }}>
      <div>
        <h2 style={{ fontSize: "1.125rem", margin: 0 }}>Pembayaran</h2>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>
          Tinjau bukti sebelum menerima atau menolak. Pembayaran diterima dapat dibatalkan dengan alasan.
        </p>
      </div>

      {error ? (
        <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>
          {error}
        </p>
      ) : null}

      {payments.length === 0 ? (
        <EmptyState
          title="Tidak ada antrean"
          description="Semua laporan pembayaran sudah diproses."
        />
      ) : (
        <ul style={{ display: "grid", gap: "var(--space-3)", listStyle: "none", margin: 0, padding: 0 }}>
          {payments.map((payment) => {
            const isActive = action?.id === payment.id;
            return (
              <li
                key={payment.id}
                style={{
                  display: "grid",
                  gap: "var(--space-3)",
                  padding: "var(--space-4)",
                  border: "1px solid var(--color-border)",
                  borderRadius: "var(--radius-lg)",
                  background: "var(--color-surface-muted)",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "flex-start",
                    justifyContent: "space-between",
                    gap: "var(--space-3)",
                  }}
                >
                  <div>
                    <strong>{payment.payment_number}</strong>
                    <p style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem", margin: "var(--space-1) 0 0" }}>
                      Tagihan {payment.invoice_number}
                    </p>
                  </div>
                  <StatusBadge variant={statusLabels[payment.verification_status].variant}>
                    {statusLabels[payment.verification_status].label}
                  </StatusBadge>
                </div>

                <p style={{ margin: 0 }}>
                  {rupiah(payment.amount)} · {methodLabels[payment.method] ?? payment.method} ·{" "}
                  {new Date(payment.paid_at).toLocaleDateString("id-ID")}
                </p>

                <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
                  {payment.proof_file_id ? (
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => void downloadProof(payment.proof_file_id as string)}
                    >
                      Lihat bukti
                    </Button>
                  ) : null}
                  {!isActive && payment.verification_status === "pending" ? (
                    <>
                      <Button type="button" onClick={() => setAction({ id: payment.id, type: "verify" })}>
                        Terima
                      </Button>
                      <Button type="button" variant="outline" onClick={() => setAction({ id: payment.id, type: "reject" })}>
                        Tolak
                      </Button>
                    </>
                  ) : null}
                  {!isActive && payment.verification_status === "verified" ? (
                    <Button type="button" variant="danger" onClick={() => setAction({ id: payment.id, type: "cancel" })}>
                      Batalkan pembayaran
                    </Button>
                  ) : null}
                </div>

                {isActive ? (
                  <div style={{ display: "grid", gap: "var(--space-3)", borderTop: "1px solid var(--color-border)", paddingTop: "var(--space-3)" }}>
                    {action.type === "reject" || action.type === "cancel" ? (
                      <FormField label={action.type === "reject" ? "Alasan penolakan" : "Alasan pembatalan"} required>
                        {(props) => (
                          <TextInput
                            {...props}
                            disabled={submitting}
                            onChange={(event) => setReason(event.target.value)}
                            value={reason}
                          />
                        )}
                      </FormField>
                    ) : (
                      <p style={{ margin: 0 }}>
                        Terima pembayaran ini? Tindakan hanya dapat dibatalkan dengan alasan dan tercatat di audit.
                      </p>
                    )}
                    <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
                      <Button
                        disabled={submitting || ((action.type === "reject" || action.type === "cancel") && !reason.trim())}
                        loading={submitting}
                        onClick={() => void submitAction()}
                        type="button"
                        variant={action.type === "verify" ? "primary" : "danger"}
                      >
                        {action.type === "verify"
                          ? "Konfirmasi terima"
                          : action.type === "reject"
                            ? "Konfirmasi tolak"
                            : "Konfirmasi pembatalan"}
                      </Button>
                      <Button disabled={submitting} onClick={closeAction} type="button" variant="secondary">
                        Batal
                      </Button>
                    </div>
                  </div>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}