"use client";

import { useCallback, useEffect, useState } from "react";

import { EmptyState } from "../../../../components/ui/EmptyState";
import { FileUploader } from "../../../../components/ui/FileUploader";
import { FormField } from "../../../../components/ui/FormField";
import { Select } from "../../../../components/ui/Select";
import { StatusBadge, type StatusBadgeProps } from "../../../../components/ui/StatusBadge";
import { TextInput } from "../../../../components/ui/TextInput";
import { Button } from "../../../../components/ui/Button";
import { apiFetch } from "../../../../lib/api";
import { useAuth } from "../../../../lib/auth-context";
import {
  downloadLetter,
  listLetterRequests,
  listLetterTypes,
  submitLetterRequest,
  updateLetterRequest,
  type LetterRequestItem,
  type LetterTypeItem,
} from "../../../../lib/letters";

type ProfileResponse = {
  resident?: { id: string; full_name: string };
};

export default function WargaSuratPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<LetterRequestItem[]>([]);
  const [types, setTypes] = useState<LetterTypeItem[]>([]);
  const [resident, setResident] = useState<ProfileResponse["resident"]>();
  const [selectedTypeID, setSelectedTypeID] = useState("");
  const [formData, setFormData] = useState<Record<string, string>>({});
  const [attachmentIDs, setAttachmentIDs] = useState<string[]>([]);
  const [residentNote, setResidentNote] = useState("");
  const [editing, setEditing] = useState<LetterRequestItem>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const selectedType = types.find((item) => item.id === selectedTypeID);

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;

      const [profile, letterTypes, requests] = await Promise.all([
        apiFetch<ProfileResponse>("me", { headers: { Authorization: `Bearer ${token}` } }),
        listLetterTypes(token),
        listLetterRequests(token),
      ]);
      setResident(profile.resident);
      setTypes(letterTypes);
      setItems(requests);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Data surat gagal dimuat.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Data loading belongs to this page's mount lifecycle.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const resetForm = () => {
    setEditing(undefined);
    setSelectedTypeID("");
    setFormData({});
    setAttachmentIDs([]);
    setResidentNote("");
  };

  const edit = (request: LetterRequestItem) => {
    setEditing(request);
    setSelectedTypeID(request.letter_type_id);
    setFormData(
      Object.fromEntries(
        Object.entries(request.form_data).map(([key, value]) => [key, String(value ?? "")]),
      ),
    );
    setAttachmentIDs(request.attachments.filter((item) => item.purpose === "attachment").map((item) => item.file_id));
    setResidentNote(request.resident_note ?? "");
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!resident || !selectedType) {
      setError("Pilih jenis surat. Akun juga harus terhubung ke data warga.");
      return;
    }

    const missingRequired = (selectedType.form_schema.fields ?? []).some(
      (field) => field.required && !formData[field.name]?.trim(),
    );
    if (missingRequired) {
      setError("Lengkapi seluruh isian wajib.");
      return;
    }
    if (selectedType.requirements.some((requirement) => requirement.required) && attachmentIDs.length === 0) {
      setError("Unggah lampiran persyaratan wajib.");
      return;
    }

    setSaving(true);
    setError("");
    setNotice("");
    try {
      const token = await getAccessToken();
      if (!token) return;

      const payload = {
        letter_type_id: selectedType.id,
        resident_id: resident.id,
        form_data: formData,
        attachment_file_ids: attachmentIDs,
        resident_note: residentNote || undefined,
      };
      if (editing) {
        await updateLetterRequest(token, editing.id, payload);
        setNotice("Perbaikan surat dikirim kembali.");
      } else {
        await submitLetterRequest(token, payload);
        setNotice("Pengajuan surat dikirim.");
      }
      resetForm();
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Pengajuan surat gagal disimpan.");
    } finally {
      setSaving(false);
    }
  };

  const getStatusVariant = (status: LetterRequestItem["status"]): StatusBadgeProps["variant"] => {
    switch (status) {
      case "approved":
      case "issued":
        return "success";
      case "submitted":
      case "under_review":
      case "awaiting_approval":
        return "info";
      case "needs_revision":
        return "warning";
      case "rejected":
      case "cancelled":
        return "danger";
      default:
        return "neutral";
    }
  };

  const download = async (id: string) => {
    try {
      const token = await getAccessToken();
      if (!token) return;
      const response = await downloadLetter(token, id);
      window.open(response.download_url, "_blank", "noopener,noreferrer");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "PDF surat gagal diunduh.");
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-4)", margin: "0 auto", maxWidth: 760, padding: "var(--space-4)", paddingBottom: 100 }}>
      <header>
        <h1 style={{ fontSize: "1.25rem", margin: 0 }}>Layanan Persuratan</h1>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>
          Ajukan surat, perbaiki bila diminta, unduh PDF setelah diterbitkan.
        </p>
      </header>

      {error ? <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>{error}</p> : null}
      {notice ? <p role="status" style={{ color: "var(--color-success)", margin: 0 }}>{notice}</p> : null}

      <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", padding: "var(--space-4)" }}>
        <h2 style={{ fontSize: "1rem", marginTop: 0 }}>{editing ? `Perbaikan ${editing.request_number}` : "Ajukan surat"}</h2>
        {editing?.resident_note ? <p style={{ background: "var(--color-surface-muted)", padding: "var(--space-2)" }}>Catatan pengurus: {editing.resident_note}</p> : null}

        <form onSubmit={submit} style={{ display: "grid", gap: "var(--space-3)" }}>
          <FormField label="Warga">
            {(props) => <TextInput {...props} disabled value={resident?.full_name ?? "Akun belum terhubung ke data warga"} />}
          </FormField>
          <FormField label="Jenis surat" required>
            {(props) => <Select {...props} value={selectedTypeID} disabled={Boolean(editing)} onChange={(event) => {
              setSelectedTypeID(event.target.value);
              setFormData({});
              setAttachmentIDs([]);
            }}>
              <option value="">Pilih jenis surat</option>
              {types.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </Select>}
          </FormField>

          {selectedType?.form_schema.fields?.map((field) => (
            <FormField key={field.name} label={field.label} required={field.required}>
              {(props) => field.type === "select" ? (
                <Select {...props} required={field.required} value={formData[field.name] ?? ""} onChange={(event) => setFormData((current) => ({ ...current, [field.name]: event.target.value }))}>
                  <option value="">Pilih {field.label}</option>
                  {(field.options ?? []).map((option) => <option key={option} value={option}>{option}</option>)}
                </Select>
              ) : (
                <TextInput {...props} required={field.required} type={field.type === "date" || field.type === "number" ? field.type : "text"} placeholder={field.placeholder} value={formData[field.name] ?? ""} onChange={(event) => setFormData((current) => ({ ...current, [field.name]: event.target.value }))} />
              )}
            </FormField>
          ))}

          {selectedType ? (
            <div>
              <strong>Persyaratan</strong>
              {selectedType.requirements.length ? (
                <ul>{selectedType.requirements.map((item, index) => <li key={`${item.name ?? item.label}-${index}`}>{item.label ?? item.name ?? "Lampiran"}{item.required ? " (wajib)" : ""}</li>)}</ul>
              ) : <p>Tidak ada lampiran wajib.</p>}
              <FileUploader entityType="letter_request" entityId={editing?.id ?? "new-letter-request"} onChange={(fileID) => {
                if (fileID) setAttachmentIDs((current) => current.includes(fileID) ? current : [...current, fileID]);
              }} />
            </div>
          ) : null}

          <FormField label="Catatan untuk pengurus">
            {(props) => <TextInput {...props} value={residentNote} onChange={(event) => setResidentNote(event.target.value)} placeholder="Opsional" />}
          </FormField>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
            <Button type="submit" disabled={!selectedTypeID || saving} loading={saving}>{editing ? "Kirim perbaikan" : "Ajukan surat"}</Button>
            {editing ? <Button type="button" variant="outline" onClick={resetForm}>Batal</Button> : null}
          </div>
        </form>
      </section>

      <section>
        <h2 style={{ fontSize: "1rem" }}>Riwayat surat</h2>
        {loading ? <p>Memuat…</p> : items.length === 0 ? <EmptyState title="Belum ada pengajuan" description="Pengajuan surat Anda akan muncul di sini." /> : (
          <div style={{ display: "grid", gap: "var(--space-3)" }}>
            {items.map((item) => (
              <article key={item.id} style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", display: "grid", gap: "var(--space-2)", padding: "var(--space-3)" }}>
                <div style={{ alignItems: "start", display: "flex", gap: "var(--space-2)", justifyContent: "space-between" }}>
                  <div><strong>{item.letter_type_name}</strong><br /><small>{item.request_number}{item.letter_number ? ` · ${item.letter_number}` : ""}</small></div>
                  <StatusBadge variant={getStatusVariant(item.status)}>{item.status.toUpperCase()}</StatusBadge>
                </div>
                {item.resident_note ? <small>Catatan: {item.resident_note}</small> : null}
                <div style={{ display: "flex", gap: "var(--space-2)", justifyContent: "flex-end" }}>
                  {item.status === "needs_revision" ? <Button type="button" onClick={() => edit(item)}>Perbaiki</Button> : null}
                  {item.status === "issued" ? <Button type="button" variant="secondary" onClick={() => void download(item.id)}>Unduh PDF</Button> : null}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}