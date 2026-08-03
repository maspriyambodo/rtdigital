"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { TextInput } from "@/components/ui/TextInput";
import { useAuth } from "@/lib/auth-context";
import {
  getOrganizationSettings,
  OrganizationSettings,
  updateOrganizationSettings,
} from "@/lib/settings";

export default function PengaturanRTPage() {
  const { getAccessToken } = useAuth();
  const [settings, setSettings] = useState<OrganizationSettings | null>(null);
  const [name, setName] = useState("");
  const [rtNumber, setRTNumber] = useState("");
  const [rwNumber, setRWNumber] = useState("");
  const [address, setAddress] = useState("");
  const [timezone, setTimezone] = useState("Asia/Jakarta");
  const [logoFileID, setLogoFileID] = useState("");
  const [bankName, setBankName] = useState("");
  const [bankAccountNumber, setBankAccountNumber] = useState("");
  const [bankAccountHolder, setBankAccountHolder] = useState("");
  const [maxUploadMB, setMaxUploadMB] = useState(10);
  const [letterPattern, setLetterPattern] = useState("SK/{YEAR}/{MONTH}/{SEQ}");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi tidak valid.");
      const data = await getOrganizationSettings(token);
      setSettings(data);
      setName(data.name);
      setRTNumber(data.rt_number);
      setRWNumber(data.rw_number);
      setAddress(data.address ?? "");
      setTimezone(data.timezone);
      setLogoFileID(data.logo_file_id ?? "");
      setBankName(data.bank_name ?? "");
      setBankAccountNumber(data.bank_account_number ?? "");
      setBankAccountHolder(data.bank_account_holder ?? "");
      setMaxUploadMB(data.max_upload_size_bytes / (1024 * 1024));
      setLetterPattern(data.default_letter_number_pattern);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Gagal memuat pengaturan RT.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void load();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");
    setSuccess("");

    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi tidak valid.");
      const updated = await updateOrganizationSettings(token, {
        name,
        rt_number: rtNumber,
        rw_number: rwNumber,
        address: address || null,
        timezone,
        logo_file_id: logoFileID || null,
        bank_name: bankName || null,
        bank_account_number: bankAccountNumber || null,
        bank_account_holder: bankAccountHolder || null,
        max_upload_size_bytes: Math.round(maxUploadMB * 1024 * 1024),
        default_letter_number_pattern: letterPattern,
      });
      setSettings(updated);
      setSuccess("Pengaturan RT tersimpan. Perubahan tercatat pada audit log.");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Gagal menyimpan pengaturan RT.");
    } finally {
      setSaving(false);
    }
  }

  const card: React.CSSProperties = {
    display: "flex",
    flexDirection: "column",
    gap: "var(--space-3)",
    padding: "var(--space-4)",
    border: "1px solid var(--color-border)",
    borderRadius: "var(--radius-md)",
    background: "var(--color-surface)",
  };

  return (
    <div style={{ maxWidth: 840, display: "flex", flexDirection: "column", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", margin: 0 }}>Pengaturan RT</h1>
        <p style={{ margin: "var(--space-1) 0 0", color: "var(--color-text-secondary)" }}>
          Identitas, rekening, zona waktu, unggahan, nomor surat, dan template.
        </p>
      </header>

      {error ? <p role="alert" style={{ margin: 0, padding: "var(--space-3)", color: "var(--color-danger-700)", background: "var(--color-danger-50)", borderRadius: "var(--radius-md)" }}>{error}</p> : null}
      {success ? <p role="status" style={{ margin: 0, padding: "var(--space-3)", color: "var(--color-success-700)", background: "var(--color-success-50)", borderRadius: "var(--radius-md)" }}>{success}</p> : null}

      {loading ? <p aria-live="polite">Memuat pengaturan…</p> : (
        <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
          <section style={card}>
            <h2 style={{ fontSize: "1.125rem", margin: 0 }}>Identitas</h2>
            <FormField label="Nama RT" id="organization-name" required>
              {(props) => <TextInput {...props} value={name} onChange={(event) => setName(event.target.value)} required />}
            </FormField>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(140px, 1fr))", gap: "var(--space-3)" }}>
              <FormField label="Nomor RT" id="rt-number" required>
                {(props) => <TextInput {...props} value={rtNumber} onChange={(event) => setRTNumber(event.target.value)} required />}
              </FormField>
              <FormField label="Nomor RW" id="rw-number" required>
                {(props) => <TextInput {...props} value={rwNumber} onChange={(event) => setRWNumber(event.target.value)} required />}
              </FormField>
            </div>
            <FormField label="Alamat sekretariat" id="address">
              {(props) => <TextInput {...props} value={address} onChange={(event) => setAddress(event.target.value)} />}
            </FormField>
            <FormField label="Zona waktu" id="timezone" required>
              {(props) => (
                <Select {...props} value={timezone} onChange={(event) => setTimezone(event.target.value)}>
                  <option value="Asia/Jakarta">WIB — Asia/Jakarta</option>
                  <option value="Asia/Makassar">WITA — Asia/Makassar</option>
                  <option value="Asia/Jayapura">WIT — Asia/Jayapura</option>
                </Select>
              )}
            </FormField>
            <FormField label="ID file logo" id="logo-file-id">
              {(props) => <TextInput {...props} value={logoFileID} onChange={(event) => setLogoFileID(event.target.value)} placeholder="Unggah logo melalui layanan file, lalu masukkan ID." />}
            </FormField>
          </section>

          <section style={card}>
            <h2 style={{ fontSize: "1.125rem", margin: 0 }}>Rekening iuran</h2>
            <FormField label="Nama bank" id="bank-name">
              {(props) => <TextInput {...props} value={bankName} onChange={(event) => setBankName(event.target.value)} />}
            </FormField>
            <FormField label="Nomor rekening" id="bank-account-number">
              {(props) => <TextInput {...props} value={bankAccountNumber} onChange={(event) => setBankAccountNumber(event.target.value)} />}
            </FormField>
            <FormField label="Nama pemilik rekening" id="bank-account-holder">
              {(props) => <TextInput {...props} value={bankAccountHolder} onChange={(event) => setBankAccountHolder(event.target.value)} />}
            </FormField>
          </section>

          <section style={card}>
            <h2 style={{ fontSize: "1.125rem", margin: 0 }}>Surat dan unggahan</h2>
            <FormField label="Batas unggahan (MB)" id="max-upload" required>
              {(props) => <TextInput {...props} type="number" min="1" max="50" step="1" value={maxUploadMB} onChange={(event) => setMaxUploadMB(Number(event.target.value))} required />}
            </FormField>
            <FormField label="Pola nomor surat" id="letter-pattern" required>
              {(props) => <TextInput {...props} value={letterPattern} onChange={(event) => setLetterPattern(event.target.value)} required />}
            </FormField>
            <p style={{ margin: 0, fontSize: "0.8125rem", color: "var(--color-text-secondary)" }}>Gunakan {"{YEAR}"}, {"{MONTH}"}, dan {"{SEQ}"} dalam pola.</p>
          </section>

          <Button type="submit" loading={saving}>Simpan pengaturan</Button>
          {settings ? <small style={{ color: "var(--color-text-secondary)" }}>Diperbarui: {new Date(settings.updated_at).toLocaleString("id-ID")}</small> : null}
        </form>
      )}
    </div>
  );
}