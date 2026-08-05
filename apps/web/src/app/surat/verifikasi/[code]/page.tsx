"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";

import { StatusBadge } from "@/components/ui/StatusBadge";
import { ApiException, apiFetch } from "@/lib/api";

interface Verification {
  letter_number: string;
  letter_type: string;
  issued_at: string | null;
  status: "issued" | "cancelled";
}

export default function VerifyLetterPage() {
  const params = useParams<{ code: string | string[] }>();
  const code = Array.isArray(params.code) ? params.code[0] : params.code;
  const [data, setData] = useState<Verification | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!code) return;

    const verify = async () => {
      try {
        const result = await apiFetch<Verification>(`letters/verify/${encodeURIComponent(code)}`);
        setData(result);
      } catch (cause) {
        setError(cause instanceof ApiException ? cause.message : "Surat tidak ditemukan atau kode verifikasi tidak valid.");
      } finally {
        setLoading(false);
      }
    };
    void verify();
  }, [code]);

  if (!code) {
    return (
      <main style={{ margin: "0 auto", maxWidth: "32rem", padding: "var(--space-4)" }}>
        <h1 style={{ fontSize: "1.5rem", margin: 0 }}>Verifikasi Surat</h1>
        <p role="alert" style={{ color: "var(--color-danger)" }}>Kode verifikasi tidak valid.</p>
      </main>
    );
  }

  if (loading) {
    return <main style={{ margin: "0 auto", maxWidth: "32rem", padding: "var(--space-4)" }}>Memverifikasi surat…</main>;
  }

  if (error || !data) {
    return (
      <main style={{ margin: "0 auto", maxWidth: "32rem", padding: "var(--space-4)" }}>
        <h1 style={{ fontSize: "1.5rem", margin: 0 }}>Verifikasi Surat</h1>
        <p role="alert" style={{ color: "var(--color-danger)" }}>{error || "Surat tidak ditemukan."}</p>
      </main>
    );
  }

  const valid = data.status === "issued";

  return (
    <main style={{ display: "grid", gap: "var(--space-4)", margin: "0 auto", maxWidth: "32rem", padding: "var(--space-4)" }}>
      <div>
        <h1 style={{ fontSize: "1.5rem", margin: 0 }}>Verifikasi Surat</h1>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>
          Hasil verifikasi dokumen RT Digital.
        </p>
      </div>

      <section
        aria-label="Hasil verifikasi surat"
        style={{
          display: "grid",
          gap: "var(--space-3)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-lg)",
          padding: "var(--space-4)",
        }}
      >
        <p style={{ margin: 0 }}>
          <strong>Status: </strong>
          <StatusBadge variant={valid ? "success" : "danger"}>{valid ? "Valid" : "Dibatalkan"}</StatusBadge>
        </p>
        <p style={{ margin: 0 }}><strong>Nomor surat:</strong> {data.letter_number || "-"}</p>
        <p style={{ margin: 0 }}><strong>Jenis surat:</strong> {data.letter_type}</p>
        <p style={{ margin: 0 }}>
          <strong>Tanggal terbit:</strong>{" "}
          {data.issued_at ? new Date(data.issued_at).toLocaleDateString("id-ID") : "-"}
        </p>
      </section>
    </main>
  );
}