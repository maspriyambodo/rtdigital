"use client";

import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";

interface Resident {
  id: string;
  full_name: string;
  national_id: string | null;
  phone: string | null;
  email: string | null;
  birth_place: string | null;
  birth_date: string | null;
  gender: "male" | "female" | null;
  marital_status: string | null;
  occupation: string | null;
  education: string | null;
  resident_status: "active" | "moved_out" | "deceased";
  verification_status: "unverified" | "verified";
}

const residentStatus: Record<Resident["resident_status"], { label: string; variant: "success" | "neutral" | "danger" }> = {
  active: { label: "Aktif", variant: "success" },
  moved_out: { label: "Pindah", variant: "neutral" },
  deceased: { label: "Meninggal", variant: "danger" },
};

export default function WargaDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { getAccessToken } = useAuth();
  const [id, setID] = useState("");
  const [resident, setResident] = useState<Resident | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    void params.then(({ id: residentID }) => setID(residentID));
  }, [params]);

  const loadResident = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      setResident(await apiFetch<Resident>(`residents/${id}`, { headers: { Authorization: `Bearer ${accessToken}` } }));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat detail warga.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken, id]);

  useEffect(() => {
    // Async API synchronization; state updates occur after the request resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadResident();
  }, [loadResident]);

  const verify = async () => {
    if (!id) return;
    setLoading(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      setResident(await apiFetch<Resident>(`residents/${id}/verify`, { method: "POST", headers: { Authorization: `Bearer ${accessToken}` } }));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memverifikasi warga.");
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <p style={{ color: "var(--color-text-secondary)" }}>Memuat detail warga…</p>;
  if (error) return <EmptyState title="Gagal memuat detail warga" description={error} action={<Button variant="secondary" onClick={() => void loadResident()}>Coba lagi</Button>} />;
  if (!resident) return <EmptyState title="Warga tidak ditemukan" />;

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header style={{ display: "flex", flexWrap: "wrap", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-4)" }}>
        <div>
          <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>{resident.full_name}</h1>
          <div style={{ display: "flex", gap: "var(--space-3)", marginTop: "var(--space-2)" }}>
            <StatusBadge variant={residentStatus[resident.resident_status].variant}>{residentStatus[resident.resident_status].label}</StatusBadge>
            <StatusBadge variant={resident.verification_status === "verified" ? "success" : "warning"}>
              {resident.verification_status === "verified" ? "Terverifikasi" : "Belum diverifikasi"}
            </StatusBadge>
          </div>
        </div>
        {resident.verification_status === "unverified" && <Button onClick={() => void verify()}>Verifikasi data</Button>}
      </header>

      <section style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
        <p><strong>NIK:</strong> {resident.national_id ?? "Belum diisi"}</p>
        <p><strong>Telepon:</strong> {resident.phone ?? "Belum diisi"}</p>
        <p><strong>Email:</strong> {resident.email ?? "Belum diisi"}</p>
        <p><strong>Tempat/tanggal lahir:</strong> {resident.birth_place ?? "Belum diisi"}, {resident.birth_date ?? "Belum diisi"}</p>
        <p><strong>Jenis kelamin:</strong> {resident.gender === "male" ? "Laki-laki" : resident.gender === "female" ? "Perempuan" : "Belum diisi"}</p>
        <p><strong>Status perkawinan:</strong> {resident.marital_status ?? "Belum diisi"}</p>
        <p><strong>Pekerjaan:</strong> {resident.occupation ?? "Belum diisi"}</p>
        <p><strong>Pendidikan:</strong> {resident.education ?? "Belum diisi"}</p>
      </section>
    </div>
  );
}