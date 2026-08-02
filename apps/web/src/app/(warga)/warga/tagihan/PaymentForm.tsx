"use client";

import { useState, type FormEvent } from "react";

import { Button } from "@/components/ui/Button";
import { DatePicker } from "@/components/ui/DatePicker";
import { FileUploader } from "@/components/ui/FileUploader";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { TextInput } from "@/components/ui/TextInput";
import { ApiException, apiFetch } from "@/lib/api";

interface SubmitPaymentResponse {
  id: string;
  payment_number: string;
  verification_status: string;
  invoice_status: string;
}

export interface PaymentFormProps {
  invoiceId: string;
  maxAmount: number;
  onSuccess: (paymentId: string) => void;
  onCancel: () => void;
}

function newIdempotencyKey() {
  return crypto.randomUUID();
}

export function PaymentForm({
  invoiceId,
  maxAmount,
  onSuccess,
  onCancel,
}: PaymentFormProps) {
  const [method, setMethod] = useState("transfer");
  const [amount, setAmount] = useState(maxAmount.toString());
  const [paidAt, setPaidAt] = useState(() => new Date().toISOString().slice(0, 10));
  const [proofFileId, setProofFileId] = useState<string>();
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [idempotencyKey] = useState(newIdempotencyKey);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");

    const parsedAmount = Number(amount);
    if (!Number.isFinite(parsedAmount) || parsedAmount <= 0 || parsedAmount > maxAmount) {
      setError(`Nominal harus antara Rp1 dan Rp${maxAmount.toLocaleString("id-ID")}.`);
      return;
    }
    if (method === "transfer" && !proofFileId) {
      setError("Bukti transfer wajib diunggah.");
      return;
    }

    setSubmitting(true);
    try {
      const response = await apiFetch<SubmitPaymentResponse>("payments", {
        method: "POST",
        headers: { "Idempotency-Key": idempotencyKey },
        body: JSON.stringify({
          invoice_id: invoiceId,
          method,
          amount: parsedAmount,
          paid_at: new Date(`${paidAt}T12:00:00`).toISOString(),
          proof_file_id: method === "transfer" ? proofFileId : undefined,
        }),
      });
      onSuccess(response.id);
    } catch (cause) {
      setError(
        cause instanceof ApiException || cause instanceof Error
          ? cause.message
          : "Gagal mengirim pembayaran.",
      );
      setSubmitting(false);
    }
  };

  return (
    <form noValidate onSubmit={handleSubmit} style={{ display: "grid", gap: "var(--space-4)" }}>
      {error ? (
        <p role="alert" style={{ color: "var(--color-danger)", fontWeight: 500, margin: 0 }}>
          {error}
        </p>
      ) : null}

      <FormField label="Metode Pembayaran" required>
        {(props) => (
          <Select
            {...props}
            disabled={submitting}
            onChange={(event) => {
              const value = event.target.value;
              setMethod(value);
              if (value !== "transfer") {
                setProofFileId(undefined);
              }
            }}
            value={method}
          >
            <option value="transfer">Transfer bank / dompet digital</option>
            <option value="cash">Tunai ke pengurus</option>
            <option value="other">Metode lainnya</option>
          </Select>
        )}
      </FormField>

      <FormField
        hint={`Sisa tagihan: Rp${maxAmount.toLocaleString("id-ID")}`}
        label="Nominal pembayaran"
        required
      >
        {(props) => (
          <TextInput
            {...props}
            disabled={submitting}
            inputMode="numeric"
            max={maxAmount}
            min={1}
            onChange={(event) => setAmount(event.target.value)}
            type="number"
            value={amount}
          />
        )}
      </FormField>

      <FormField label="Tanggal pembayaran" required>
        {(props) => (
          <DatePicker
            {...props}
            disabled={submitting}
            max={new Date().toISOString().slice(0, 10)}
            onChange={(event) => setPaidAt(event.target.value)}
            value={paidAt}
          />
        )}
      </FormField>

      {method === "transfer" ? (
        <FormField label="Bukti transfer" required>
          {(props) => (
            <div id={props.id}>
              <FileUploader
                disabled={submitting}
                entityId={invoiceId}
                onChange={setProofFileId}
              />
            </div>
          )}
        </FormField>
      ) : null}

      <div style={{ display: "flex", gap: "var(--space-3)", marginTop: "var(--space-2)" }}>
        <Button disabled={submitting} onClick={onCancel} type="button" variant="secondary">
          Batal
        </Button>
        <Button loading={submitting} type="submit">
          Kirim pembayaran
        </Button>
      </div>
    </form>
  );
}