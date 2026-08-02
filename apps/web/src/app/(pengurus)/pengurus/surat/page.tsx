"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "../../../../components/ui/Button";
import { EmptyState } from "../../../../components/ui/EmptyState";
import { FormField } from "../../../../components/ui/FormField";
import { StatusBadge, type StatusBadgeProps } from "../../../../components/ui/StatusBadge";
import { TextInput } from "../../../../components/ui/TextInput";
import { useAuth } from "../../../../lib/auth-context";
import {
  approveLetterRequest,
  issueLetter,
  listLetterRequests,
  processLetterRequest,
  rejectLetterRequest,
  requestRevision,
  type LetterRequestItem,
  type ReviewLetterRequest,
} from "../../../../lib/letters";

export default function PengurusSuratPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<LetterRequestItem[]>([]);
  const [selected, setSelected] = useState<LetterRequestItem>();
  const [residentNote, setResidentNote] = useState("");
  const [internalNote, setInternalNote] = useState("");
  const [loading, setLoading] = useState(true);
  const [actioning, setActioning] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      const requests = await listLetterRequests(token);
      setItems(requests);
      setSelected((current) => current ? requests.find((item) => item.id === current.id) : undefined);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Antrean surat gagal dimuat.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Data loading belongs to this page's mount lifecycle.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const select = (item: LetterRequestItem) => {
    setSelected(item);
    setResidentNote(item.resident_note ?? "");
    setInternalNote(item.internal_note ?? "");
    setError("");
    setNotice("");
  };

  const handleAction = async (
    action: (token: string, id: string, data: ReviewLetterRequest) => Promise<unknown>,
    requireResidentNote = false,
  ) => {
    if (!selected) return;
    if (requireResidentNote && !residentNote.trim()) {
      setError("Catatan untuk warga wajib diisi.");
      return;
    }

    setActioning(true);
    setError("");
    setNotice("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      await action(token, selected.id, {
        resident_note: residentNote || undefined,
        internal_note: internalNote || undefined,
      });
      setNotice("Tindakan surat berhasil diproses.");
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Tindakan surat gagal diproses.");
    } finally {
      setActioning(false);
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

  const handleIssue = async () => {
    if (!selected) return;
    setActioning(true);
    setError("");
    setNotice("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      await issueLetter(token, selected.id);
      setNotice("Surat berhasil diterbitkan.");
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Surat gagal diterbitkan.");
    } finally {
      setActioning(false);
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-4)", gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))", padding: "var(--space-4)" }}>
      <section style={{ display: "grid", gap: "var(--space-3)" }}>
        <header>
          <h1 style={{ fontSize: "1.25rem", margin: 0 }}>Antrean Pengajuan Surat</h1>
          <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>
            Tinjau, minta revisi, setujui, lalu terbitkan surat warga.
          </p>
        </header>
        {notice ? <p role="status" style={{ color: "var(--color-success)", margin: 0 }}>{notice}</p> : null}
        {loading ? <p>Memuat antrean…</p> : items.length === 0 ? <EmptyState title="Antrean kosong" description="Belum ada pengajuan surat yang masuk." /> : (
          <div style={{ display: "grid", gap: "var(--space-2)" }}>
            {items.map((item) => (
              <button key={item.id} type="button" onClick={() => select(item)} style={{ background: selected?.id === item.id ? "var(--color-primary-50)" : "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", cursor: "pointer", display: "grid", gap: "var(--space-1)", padding: "var(--space-3)", textAlign: "left", width: "100%" }}>
                <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-2)" }}>
                  <strong>{item.resident_name}</strong><StatusBadge variant={getStatusVariant(item.status)}>{item.status.toUpperCase()}</StatusBadge>
                </div>
                <small style={{ color: "var(--color-text-secondary)" }}>{item.letter_type_name} · {item.request_number}</small>
              </button>
            ))}
          </div>
        )}
      </section>

      <section>
        {!selected ? <EmptyState title="Belum memilih surat" description="Pilih surat dari antrean untuk memproses." /> : (
          <div style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", display: "grid", gap: "var(--space-4)", padding: "var(--space-4)" }}>
            <header style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-2)" }}>
              <div><h2 style={{ fontSize: "1.1rem", margin: 0 }}>Detail Pengajuan</h2><small>{selected.request_number}</small></div>
              <StatusBadge variant={getStatusVariant(selected.status)}>{selected.status.toUpperCase()}</StatusBadge>
            </header>
            {error ? <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>{error}</p> : null}
            <div style={{ display: "grid", gap: "var(--space-1)" }}>
              <div><small>Pemohon:</small> <strong>{selected.requester_name}</strong></div>
              <div><small>Subjek:</small> <strong>{selected.resident_name}</strong></div>
              <div><small>Jenis:</small> <strong>{selected.letter_type_name}</strong></div>
              {selected.letter_number ? <div><small>Nomor surat:</small> <strong>{selected.letter_number}</strong></div> : null}
            </div>
            <div><h3>Isian formulir</h3><pre style={{ background: "var(--color-surface-muted)", margin: 0, overflow: "auto", padding: "var(--space-2)", whiteSpace: "pre-wrap" }}>{JSON.stringify(selected.form_data, null, 2)}</pre></div>
            {selected.attachments.length ? <div><h3>Lampiran</h3><ul>{selected.attachments.map((file) => <li key={file.attachment_id}>{file.original_name}</li>)}</ul></div> : null}
            <div style={{ display: "grid", gap: "var(--space-3)" }}>
              <FormField label="Catatan untuk warga (wajib untuk revisi/penolakan)">
                {(props) => <TextInput {...props} value={residentNote} onChange={(event) => setResidentNote(event.target.value)} />}
              </FormField>
              <FormField label="Catatan internal">
                {(props) => <TextInput {...props} value={internalNote} onChange={(event) => setInternalNote(event.target.value)} />}
              </FormField>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
                {selected.status === "submitted" ? <Button disabled={actioning} loading={actioning} onClick={() => void handleAction(processLetterRequest)}>Mulai review</Button> : null}
                {["submitted", "under_review", "awaiting_approval"].includes(selected.status) ? <><Button disabled={actioning} loading={actioning} variant="secondary" onClick={() => void handleAction(requestRevision, true)}>Minta revisi</Button><Button disabled={actioning} loading={actioning} variant="outline" onClick={() => void handleAction(rejectLetterRequest, true)}>Tolak</Button></> : null}
                {["under_review", "awaiting_approval"].includes(selected.status) ? <Button disabled={actioning} loading={actioning} onClick={() => void handleAction(approveLetterRequest)}>Setujui</Button> : null}
                {selected.status === "approved" ? <Button disabled={actioning} loading={actioning} onClick={() => void handleIssue()}>Terbitkan PDF</Button> : null}
              </div>
            </div>
          </div>
        )}
      </section>
    </div>
  );
}