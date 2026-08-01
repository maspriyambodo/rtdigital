"use client";

import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";

interface Household {
  id: string;
  house_unit_code: string | null;
  family_card_number: string | null;
  internal_number: string | null;
  domicile_status: "permanent" | "temporary";
  status: "active" | "inactive";
  members: { id: string; full_name: string; relationship: string; is_active: boolean }[];
}

export default function KeluargaPengurusPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<Household[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      setItems(await apiFetch<Household[]>("households", { headers: { Authorization: `Bearer ${token}` } }));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat keluarga.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Async API synchronization; state updates occur after the request resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  if (loading) return <p style={{ color: "var(--color-text-secondary)" }}>Memuat keluarga…</p>;
  if (error) return <EmptyState title="Gagal memuat keluarga" description={error} action={<Button variant="secondary" onClick={() => void load()}>Coba lagi</Button>} />;
  if (!items.length) return <EmptyState title="Belum ada keluarga" description="Buat keluarga melalui API setelah rumah/unit dan warga tersedia." />;

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Data Keluarga</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>Keluarga, unit hunian, dan anggota aktif.</p>
      </header>
      <div style={{ display: "grid", gap: "var(--space-3)" }}>
        {items.map((household) => (
          <article key={household.id} style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
            <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}>
              <strong>{household.family_card_number ?? household.internal_number ?? "Keluarga tanpa nomor"}</strong>
              <StatusBadge variant={household.status === "active" ? "success" : "neutral"}>{household.status === "active" ? "Aktif" : "Nonaktif"}</StatusBadge>
            </div>
            <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
              Unit {household.house_unit_code ?? "-"} · {household.domicile_status === "permanent" ? "Tetap" : "Sementara"} · {household.members?.filter((member) => member.is_active).length ?? 0} anggota aktif
            </span>
          </article>
        ))}
      </div>
    </div>
  );
}