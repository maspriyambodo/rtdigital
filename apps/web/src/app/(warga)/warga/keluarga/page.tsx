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

interface HouseholdMember {
  id: string;
  resident_id: string;
  full_name: string;
  relationship: "head" | "spouse" | "child" | "parent" | "other";
  is_active: boolean;
}

interface Household {
  id: string;
  house_unit_code: string | null;
  family_card_number: string | null;
  domicile_status: "permanent" | "temporary";
  members: HouseholdMember[];
}

const relationshipLabels: Record<HouseholdMember["relationship"], string> = {
  head: "Kepala Keluarga",
  spouse: "Pasangan",
  child: "Anak",
  parent: "Orang Tua",
  other: "Lainnya",
};

export default function KeluargaPage() {
  const { getAccessToken } = useAuth();
  const [household, setHousehold] = useState<Household | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedResidentID, setSelectedResidentID] = useState<string | null>(null);
  const [field, setField] = useState("phone");
  const [value, setValue] = useState("");
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState("");

  const loadHousehold = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      const households = await apiFetch<Household[]>("households", {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (!households.length) {
        setHousehold(null);
        return;
      }
      setHousehold(
        await apiFetch<Household>(`households/${households[0].id}`, {
          headers: { Authorization: `Bearer ${accessToken}` },
        }),
      );
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat profil keluarga.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Async API synchronization; state updates occur after the request resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadHousehold();
  }, [loadHousehold]);

  const submitCorrection = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selectedResidentID || !reason.trim() || !value.trim()) return;

    setSubmitting(true);
    setError("");
    setMessage("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      await apiFetch(`residents/${selectedResidentID}/corrections`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${accessToken}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          requested_changes: { [field]: value.trim() },
          reason: reason.trim(),
        }),
      });
      setMessage("Pengajuan koreksi data berhasil dikirim.");
      setValue("");
      setReason("");
      setSelectedResidentID(null);
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal mengajukan koreksi data.");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <p style={{ color: "var(--color-text-secondary)" }}>Memuat keluarga…</p>;
  if (error) return <EmptyState title="Gagal memuat profil keluarga" description={error} action={<Button variant="secondary" onClick={() => void loadHousehold()}>Coba lagi</Button>} />;
  if (!household) return <EmptyState title="Belum terhubung ke keluarga" description="Hubungi pengurus RT untuk menghubungkan akun Anda ke kartu keluarga." />;

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Profil Keluarga</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
          Unit {household.house_unit_code ?? "-"} · No. KK {household.family_card_number ?? "-"}
        </p>
      </header>

      {message && <p role="status" style={{ color: "var(--color-success-text)" }}>{message}</p>}

      <section style={{ display: "grid", gap: "var(--space-3)" }}>
        <h2 style={{ fontSize: "1.125rem" }}>Anggota Keluarga</h2>
        {household.members.map((member) => (
          <article key={member.id} style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-3)" }}>
              <strong>{member.full_name}</strong>
              <StatusBadge variant={member.relationship === "head" ? "success" : "neutral"}>
                {relationshipLabels[member.relationship]}
              </StatusBadge>
            </div>
            {!member.is_active && <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>Anggota nonaktif</span>}
            <div>
              <Button variant="secondary" onClick={() => setSelectedResidentID(member.resident_id)}>
                Ajukan koreksi data
              </Button>
            </div>
          </article>
        ))}
      </section>

      {selectedResidentID && (
        <form onSubmit={(event) => void submitCorrection(event)} style={{ display: "grid", gap: "var(--space-4)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
          <h2 style={{ fontSize: "1.125rem" }}>Formulir Koreksi Data</h2>
          <FormField label="Bidang Data">
            {({ id, "aria-describedby": ariaDescribedBy }) => (
              <Select id={id} aria-describedby={ariaDescribedBy} value={field} onChange={(event) => setField(event.target.value)}>
                <option value="phone">Nomor Telepon</option>
                <option value="email">Email</option>
                <option value="occupation">Pekerjaan</option>
                <option value="education">Pendidikan</option>
                <option value="marital_status">Status Perkawinan</option>
              </Select>
            )}
          </FormField>
          <FormField label="Nilai Baru">
            {({ id, "aria-describedby": ariaDescribedBy, "aria-invalid": ariaInvalid }) => (
              <TextInput id={id} aria-describedby={ariaDescribedBy} aria-invalid={ariaInvalid} name="value" value={value} onChange={(event) => setValue(event.target.value)} required />
            )}
          </FormField>
          <FormField label="Alasan Koreksi">
            {({ id, "aria-describedby": ariaDescribedBy, "aria-invalid": ariaInvalid }) => (
              <TextInput id={id} aria-describedby={ariaDescribedBy} aria-invalid={ariaInvalid} name="reason" value={reason} onChange={(event) => setReason(event.target.value)} required />
            )}
          </FormField>
          <div style={{ display: "flex", gap: "var(--space-2)" }}>
            <Button type="submit" disabled={submitting}>{submitting ? "Mengirim…" : "Kirim koreksi"}</Button>
            <Button type="button" variant="secondary" onClick={() => setSelectedResidentID(null)}>Batal</Button>
          </div>
        </form>
      )}
    </div>
  );
}