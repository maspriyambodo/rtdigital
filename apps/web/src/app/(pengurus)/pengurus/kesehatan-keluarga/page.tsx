"use client";

import { useCallback, useEffect, useState } from "react";

import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

interface HouseholdHealthScore {
  household_id: string;
  internal_number: string;
  score: number;
  missing_items: string[];
  domicile_due_at?: string;
}

const missingLabels: Record<string, string> = {
  active_members: "anggota aktif",
  contact: "kontak",
  verification: "verifikasi",
  data_update: "pembaruan data",
  domicile_confirmation: "konfirmasi domisili",
};

function scoreVariant(score: number): "success" | "warning" | "danger" {
  if (score >= 80) return "success";
  if (score >= 50) return "warning";
  return "danger";
}

export default function HouseholdHealthPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<HouseholdHealthScore[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");

      const result = await apiFetch<HouseholdHealthScore[]>("households/health-scores", {
        headers: { Authorization: `Bearer ${token}` },
      });
      setItems(result);
    } catch (cause) {
      setError(
        cause instanceof ApiException || cause instanceof Error
          ? cause.message
          : "Gagal memuat daftar kerja keluarga.",
      );
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Initial remote-data synchronization; load handles cancellation-safe UI state.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  return (
    <main style={{ display: "grid", gap: "var(--space-4)", maxWidth: "52rem" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", margin: 0 }}>Kesehatan Data Keluarga</h1>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>
          Daftar kerja sekretaris. Skor membantu prioritas pembaruan, bukan penilaian atau sanksi warga.
        </p>
      </header>

      {error ? (
        <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>
          {error}
        </p>
      ) : null}

      {loading ? (
        <p style={{ color: "var(--color-text-secondary)", margin: 0 }}>Memuat daftar kerja…</p>
      ) : items.length === 0 ? (
        <EmptyState title="Belum ada keluarga aktif" description="Daftar kualitas data akan muncul saat keluarga aktif tersedia." />
      ) : (
        <ul style={{ display: "grid", gap: "var(--space-3)", listStyle: "none", margin: 0, padding: 0 }}>
          {items.map((item) => (
            <li
              key={item.household_id}
              style={{
                display: "grid",
                gap: "var(--space-2)",
                border: "1px solid var(--color-border)",
                borderRadius: "var(--radius-lg)",
                padding: "var(--space-4)",
              }}
            >
              <div style={{ alignItems: "center", display: "flex", gap: "var(--space-3)", justifyContent: "space-between" }}>
                <strong>Keluarga {item.internal_number}</strong>
                <StatusBadge variant={scoreVariant(item.score)}>{item.score}% lengkap</StatusBadge>
              </div>
              <p style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem", margin: 0 }}>
                {item.missing_items.length
                  ? `Perlu ditindaklanjuti: ${item.missing_items.map((item) => missingLabels[item] ?? item).join(", ")}.`
                  : "Data operasional sudah lengkap."}
              </p>
              {item.domicile_due_at ? (
                <p style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem", margin: 0 }}>
                  Evaluasi domisili: {new Date(item.domicile_due_at).toLocaleDateString("id-ID")}
                </p>
              ) : null}
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}