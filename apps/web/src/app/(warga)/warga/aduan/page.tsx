"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "../../../../components/ui/Button";
import { EmptyState } from "../../../../components/ui/EmptyState";
import { FileUploader } from "../../../../components/ui/FileUploader";
import { FormField } from "../../../../components/ui/FormField";
import { Select } from "../../../../components/ui/Select";
import { StatusBadge, type StatusBadgeProps } from "../../../../components/ui/StatusBadge";
import { TextArea } from "../../../../components/ui/TextArea";
import { TextInput } from "../../../../components/ui/TextInput";
import { useAuth } from "../../../../lib/auth-context";
import {
  addComment,
  createComplaint,
  listComplaints,
  updateComplaint,
  updateStatus,
  type ComplaintItem,
  type ComplaintPriority,
} from "../../../../lib/complaints";

const statusLabel: Record<ComplaintItem["status"], string> = {
  new: "Baru",
  reviewed: "Ditinjau",
  in_progress: "Diproses",
  waiting_information: "Menunggu informasi",
  resolved: "Selesai",
  rejected: "Ditolak",
  closed: "Ditutup",
};

function statusVariant(status: ComplaintItem["status"]): StatusBadgeProps["variant"] {
  switch (status) {
    case "resolved":
    case "closed":
      return "success";
    case "rejected":
      return "danger";
    case "waiting_information":
      return "warning";
    case "new":
    case "reviewed":
    case "in_progress":
      return "info";
  }
}

export default function WargaAduanPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<ComplaintItem[]>([]);
  const [title, setTitle] = useState("");
  const [category, setCategory] = useState("");
  const [description, setDescription] = useState("");
  const [locationDescription, setLocationDescription] = useState("");
  const [priority, setPriority] = useState<ComplaintPriority>("normal");
  const [attachmentIDs, setAttachmentIDs] = useState<string[]>([]);
  const [editing, setEditing] = useState<ComplaintItem>();
  const [activeComplaint, setActiveComplaint] = useState<ComplaintItem>();
  const [commentBody, setCommentBody] = useState("");
  const [closingNote, setClosingNote] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      const complaints = await listComplaints(token);
      setItems(complaints);
      if (activeComplaint) setActiveComplaint(complaints.find((item) => item.id === activeComplaint.id));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Data aduan gagal dimuat.");
    } finally {
      setLoading(false);
    }
  }, [activeComplaint, getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const resetForm = () => {
    setEditing(undefined);
    setTitle("");
    setCategory("");
    setDescription("");
    setLocationDescription("");
    setPriority("normal");
    setAttachmentIDs([]);
  };

  const startEdit = (item: ComplaintItem) => {
    setEditing(item);
    setTitle(item.title);
    setCategory(item.category);
    setDescription(item.description);
    setLocationDescription(item.location_description ?? "");
    setPriority(item.priority);
    setAttachmentIDs(item.attachments.map((attachment) => attachment.file_id));
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const submitComplaint = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!category.trim() || !title.trim() || !description.trim()) {
      setError("Isi kategori, judul, dan deskripsi aduan.");
      return;
    }

    setSaving(true);
    setError("");
    setNotice("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      const payload = {
        category,
        title,
        description,
        location_description: locationDescription || undefined,
        priority,
        attachment_file_ids: attachmentIDs,
      };
      if (editing) {
        await updateComplaint(token, editing.id, payload);
        setNotice("Aduan diperbarui.");
      } else {
        await createComplaint(token, payload);
        setNotice("Aduan dikirim.");
      }
      resetForm();
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Aduan gagal disimpan.");
    } finally {
      setSaving(false);
    }
  };

  const submitComment = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!activeComplaint || !commentBody.trim()) return;
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      await addComment(token, activeComplaint.id, { body: commentBody, is_internal: false });
      setCommentBody("");
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Komentar gagal dikirim.");
    }
  };

  const closeComplaint = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!activeComplaint || !closingNote.trim()) return;
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      await updateStatus(token, activeComplaint.id, { status: "closed", resolution_note: closingNote });
      setClosingNote("");
      setNotice("Aduan ditutup.");
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Aduan gagal ditutup.");
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-4)", margin: "0 auto", maxWidth: 760, padding: "var(--space-4)", paddingBottom: 100 }}>
      <header>
        <h1 style={{ fontSize: "1.25rem", margin: 0 }}>Aduan Warga</h1>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>Laporkan masalah lingkungan, pantau tindak lanjutnya.</p>
      </header>

      {error ? <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>{error}</p> : null}
      {notice ? <p role="status" style={{ color: "var(--color-success)", margin: 0 }}>{notice}</p> : null}

      {activeComplaint ? (
        <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", padding: "var(--space-4)" }}>
          <div style={{ alignItems: "start", display: "flex", gap: "var(--space-2)", justifyContent: "space-between" }}>
            <div><h2 style={{ fontSize: "1rem", margin: 0 }}>{activeComplaint.title}</h2><small>{activeComplaint.ticket_number} · {activeComplaint.category}</small></div>
            <StatusBadge variant={statusVariant(activeComplaint.status)}>{statusLabel[activeComplaint.status]}</StatusBadge>
          </div>
          <p>{activeComplaint.description}</p>
          {activeComplaint.location_description ? <p><strong>Lokasi:</strong> {activeComplaint.location_description}</p> : null}
          {activeComplaint.assigned_to_name ? <p><strong>Petugas:</strong> {activeComplaint.assigned_to_name}</p> : null}
          {activeComplaint.resolution_note ? <p style={{ background: "var(--color-surface-muted)", padding: "var(--space-2)" }}><strong>Catatan resolusi:</strong> {activeComplaint.resolution_note}</p> : null}

          {activeComplaint.status === "resolved" ? <form onSubmit={closeComplaint} style={{ display: "grid", gap: "var(--space-2)", marginTop: "var(--space-3)" }}>
            <FormField label="Umpan balik penutupan" required>{(props) => <TextInput {...props} required value={closingNote} onChange={(event) => setClosingNote(event.target.value)} />}</FormField>
            <div><Button type="submit">Tutup aduan</Button></div>
          </form> : null}

          <div style={{ borderTop: "1px solid var(--color-border)", display: "grid", gap: "var(--space-3)", marginTop: "var(--space-4)", paddingTop: "var(--space-4)" }}>
            <h3 style={{ fontSize: "1rem", margin: 0 }}>Timeline dan komentar</h3>
            {activeComplaint.comments.length === 0 ? <p style={{ margin: 0 }}>Belum ada pembaruan.</p> : activeComplaint.comments.map((comment) => (
              <article key={comment.id} style={{ background: "var(--color-surface-muted)", borderRadius: "var(--radius-md)", padding: "var(--space-2)" }}>
                <strong>{comment.author_name}</strong><br /><small>{new Date(comment.created_at).toLocaleString("id-ID")}</small><p style={{ marginBottom: 0 }}>{comment.body}</p>
              </article>
            ))}
            <form onSubmit={submitComment} style={{ display: "flex", gap: "var(--space-2)" }}>
              <TextInput required aria-label="Komentar" placeholder="Tulis komentar…" value={commentBody} onChange={(event) => setCommentBody(event.target.value)} />
              <Button type="submit">Kirim</Button>
            </form>
            <div><Button type="button" variant="outline" onClick={() => setActiveComplaint(undefined)}>Kembali</Button></div>
          </div>
        </section>
      ) : (
        <>
          <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", padding: "var(--space-4)" }}>
            <h2 style={{ fontSize: "1rem", marginTop: 0 }}>{editing ? `Ubah ${editing.ticket_number}` : "Buat aduan"}</h2>
            <form onSubmit={submitComplaint} style={{ display: "grid", gap: "var(--space-3)" }}>
              <FormField label="Kategori" required>{(props) => <Select {...props} required value={category} onChange={(event) => setCategory(event.target.value)}><option value="">Pilih kategori</option><option value="keamanan">Keamanan</option><option value="kebersihan">Kebersihan</option><option value="infrastruktur">Infrastruktur</option><option value="fasilitas_umum">Fasilitas umum</option><option value="lainnya">Lainnya</option></Select>}</FormField>
              <FormField label="Judul" required>{(props) => <TextInput {...props} required value={title} onChange={(event) => setTitle(event.target.value)} />}</FormField>
              <FormField label="Deskripsi" required>{(props) => <TextArea {...props} required rows={4} value={description} onChange={(event) => setDescription(event.target.value)} style={{ minHeight: 96 }} />}</FormField>
              <FormField label="Lokasi umum">{(props) => <TextInput {...props} value={locationDescription} onChange={(event) => setLocationDescription(event.target.value)} />}</FormField>
              <FormField label="Prioritas">{(props) => <Select {...props} value={priority} onChange={(event) => setPriority(event.target.value as ComplaintPriority)}><option value="low">Rendah</option><option value="normal">Normal</option><option value="high">Tinggi</option></Select>}</FormField>
              <div><strong>Lampiran opsional</strong><FileUploader entityType="complaint" entityId={editing?.id ?? "new-complaint"} onChange={(fileID) => fileID && setAttachmentIDs((current) => current.includes(fileID) ? current : [...current, fileID])} /></div>
              <div style={{ display: "flex", gap: "var(--space-2)" }}><Button type="submit" loading={saving} disabled={saving}>{editing ? "Simpan" : "Kirim aduan"}</Button>{editing ? <Button type="button" variant="outline" onClick={resetForm}>Batal</Button> : null}</div>
            </form>
          </section>

          <section>
            <h2 style={{ fontSize: "1rem" }}>Aduan saya</h2>
            {loading ? <p>Memuat…</p> : items.length === 0 ? <EmptyState title="Belum ada aduan" description="Aduan yang Anda buat muncul di sini." /> : <div style={{ display: "grid", gap: "var(--space-3)" }}>{items.map((item) => <article key={item.id} style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", display: "grid", gap: "var(--space-2)", padding: "var(--space-3)" }}><div style={{ alignItems: "start", display: "flex", justifyContent: "space-between", gap: "var(--space-2)" }}><div><strong>{item.title}</strong><br /><small>{item.ticket_number} · {item.category}</small></div><StatusBadge variant={statusVariant(item.status)}>{statusLabel[item.status]}</StatusBadge></div><div style={{ display: "flex", gap: "var(--space-2)", justifyContent: "flex-end" }}>{item.status === "new" ? <Button type="button" variant="outline" onClick={() => startEdit(item)}>Ubah</Button> : null}<Button type="button" onClick={() => setActiveComplaint(item)}>Detail</Button></div></article>)}</div>}
          </section>
        </>
      )}
    </div>
  );
}