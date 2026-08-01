"use client";

import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";

interface HouseUnit {
  id: string;
  code: string;
  address_detail: string | null;
  occupancy_status: "owned" | "rented" | "contract" | "empty";
  status: "active" | "inactive";
}

const occupancyLabels: Record<HouseUnit["occupancy_status"], string> = {
  owned: "Milik Sendiri",
  rented: "Sewa/Kost",
  contract: "Kontrak",
  empty: "Kosong",
};

export default function RumahPage() {
  const { getAccessToken } = useAuth();
  const [units, setUnits] = useState<HouseUnit[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [code, setCode] = useState("");
  const [addressDetail, setAddressDetail] = useState("");
  const [occupancyStatus, setOccupancyStatus] = useState<HouseUnit["occupancy_status"]>("owned");
  const [creating, setCreating] = useState(false);

  const loadUnits = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      setUnits(await apiFetch<HouseUnit[]>("house-units", { headers: { Authorization: `Bearer ${accessToken}` } }));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat rumah/unit.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Async API synchronization; state updates occur after the request resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadUnits();
  }, [loadUnits]);

  const createUnit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!code.trim()) return;

    setCreating(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      await apiFetch("house-units", {
        method: "POST",
        headers: { Authorization: `Bearer ${accessToken}`, "Content-Type": "application/json" },
        body: JSON.stringify({
          code: code.trim(),
          address_detail: addressDetail.trim() || undefined,
          occupancy_status: occupancyStatus,
        }),
      });
      setCode("");
      setAddressDetail("");
      await loadUnits();
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal membuat unit rumah.");
    } finally {
      setCreating(false);
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Manajemen Rumah & Unit</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>Pendataan blok, nomor rumah, dan status hunian.</p>
      </header>

      <form onSubmit={(event) => void createUnit(event)} style={{ display: "grid", gap: "var(--space-4)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
        <h2 style={{ fontSize: "1.125rem" }}>Tambah Rumah/Unit</h2>
        <FormField label="Kode/Nomor Unit">
          {({ id, "aria-describedby": ariaDescribedBy, "aria-invalid": ariaInvalid }) => (
            <TextInput id={id} aria-describedby={ariaDescribedBy} aria-invalid={ariaInvalid} name="code" value={code} onChange={(event) => setCode(event.target.value)} required />
          )}
        </FormField>
        <FormField label="Detail Alamat">
          {({ id, "aria-describedby": ariaDescribedBy, "aria-invalid": ariaInvalid }) => (
            <TextInput id={id} aria-describedby={ariaDescribedBy} aria-invalid={ariaInvalid} name="addressDetail" value={addressDetail} onChange={(event) => setAddressDetail(event.target.value)} />
          )}
        </FormField>
        <FormField label="Status Hunian">
          {({ id, "aria-describedby": ariaDescribedBy }) => (
            <select id={id} aria-describedby={ariaDescribedBy} value={occupancyStatus} onChange={(event) => setOccupancyStatus(event.target.value as HouseUnit["occupancy_status"])} style={{ minHeight: 44, border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", background: "var(--color-surface)", color: "var(--color-text)", padding: "0 var(--space-3)" }}>
              <option value="owned">Milik Sendiri</option>
              <option value="rented">Sewa/Kost</option>
              <option value="contract">Kontrak</option>
              <option value="empty">Kosong</option>
            </select>
          )}
        </FormField>
        <Button type="submit" disabled={creating}>{creating ? "Menyimpan…" : "Tambah unit"}</Button>
      </form>

      {loading ? <p style={{ color: "var(--color-text-secondary)" }}>Memuat unit…</p> : error ? (
        <EmptyState title="Gagal memuat data" description={error} action={<Button variant="secondary" onClick={() => void loadUnits()}>Coba lagi</Button>} />
      ) : units.length === 0 ? (
        <EmptyState title="Belum ada unit" description="Tambahkan unit pertama untuk memulai." />
      ) : (
        <div style={{ display: "grid", gap: "var(--space-3)" }}>
          {units.map((unit) => (
            <article key={unit.id} style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-3)" }}>
                <strong>Unit {unit.code}</strong>
                <StatusBadge variant={unit.status === "active" ? "success" : "neutral"}>{unit.status === "active" ? "Aktif" : "Nonaktif"}</StatusBadge>
              </div>
              <p style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>Hunian: {occupancyLabels[unit.occupancy_status]} · {unit.address_detail ?? "Alamat detail belum diisi"}</p>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}