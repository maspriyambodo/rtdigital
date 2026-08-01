"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";

interface Resident {
  id: string;
  full_name: string;
  national_id: string | null;
  phone: string | null;
  resident_status: "active" | "moved_out" | "deceased";
  verification_status: "unverified" | "verified";
}

const residentStatus: Record<Resident["resident_status"], { label: string; variant: "success" | "neutral" | "danger" }> = {
  active: { label: "Aktif", variant: "success" },
  moved_out: { label: "Pindah", variant: "neutral" },
  deceased: { label: "Meninggal", variant: "danger" },
};

export default function WargaPage() {
  const { getAccessToken } = useAuth();
  const [residents, setResidents] = useState<Resident[]>([]);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadResidents = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      const params = new URLSearchParams();
      if (query.trim()) params.set("q", query.trim());
      if (status) params.set("status", status);
      setResidents(await apiFetch<Resident[]>(`residents?${params}`, { headers: { Authorization: `Bearer ${accessToken}` } }));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat warga.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken, query, status]);

  useEffect(() => {
    const timer = window.setTimeout(() => void loadResidents(), 250);
    return () => window.clearTimeout(timer);
  }, [loadResidents]);

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header style={{ display: "flex", flexWrap: "wrap", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-4)" }}>
        <div>
          <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Data Warga</h1>
          <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
            Pencarian, filter status, dan verifikasi data warga.
          </p>
        </div>
      </header>

      <div style={{ display: "grid", gap: "var(--space-3)", gridTemplateColumns: "minmax(0, 1fr) minmax(9rem, 12rem)" }}>
        <TextInput name="q" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Cari nama atau NIK…" />
        <select aria-label="Filter status warga" value={status} onChange={(event) => setStatus(event.target.value)} style={{ minHeight: 44, border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", background: "var(--color-surface)", color: "var(--color-text)", padding: "0 var(--space-3)" }}>
          <option value="">Semua status</option>
          <option value="active">Aktif</option>
          <option value="moved_out">Pindah</option>
          <option value="deceased">Meninggal</option>
        </select>
      </div>

      {loading ? (
        <p style={{ color: "var(--color-text-secondary)" }}>Memuat warga…</p>
      ) : error ? (
        <EmptyState title="Gagal memuat warga" description={error} action={<Button variant="secondary" onClick={() => void loadResidents()}>Coba lagi</Button>} />
      ) : residents.length === 0 ? (
        <EmptyState title="Warga tidak ditemukan" description="Ubah kata kunci atau filter pencarian." />
      ) : (
        <div style={{ display: "grid", gap: "var(--space-3)" }}>
          {residents.map((resident) => (
            <article key={resident.id} style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
              <div style={{ display: "flex", gap: "var(--space-3)", justifyContent: "space-between", alignItems: "flex-start" }}>
                <strong>{resident.full_name}</strong>
                <StatusBadge variant={residentStatus[resident.resident_status].variant}>{residentStatus[resident.resident_status].label}</StatusBadge>
              </div>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
                NIK: {resident.national_id ?? "Belum diisi"} · {resident.verification_status === "verified" ? "Terverifikasi" : "Belum diverifikasi"}
              </span>
              <Link href={`/pengurus/warga/${resident.id}`} style={{ color: "var(--color-primary-600)", fontSize: "0.875rem", fontWeight: 600, textDecoration: "none", marginTop: "var(--space-2)" }}>
                Detail warga →
              </Link>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}