"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "../../../../components/ui/Button";
import { EmptyState } from "../../../../components/ui/EmptyState";
import { FormField } from "../../../../components/ui/FormField";
import { Select } from "../../../../components/ui/Select";
import { StatusBadge, type StatusBadgeProps } from "../../../../components/ui/StatusBadge";
import { TextInput } from "../../../../components/ui/TextInput";
import { apiFetch } from "../../../../lib/api";
import { useAuth } from "../../../../lib/auth-context";
import {
  addComment,
  assignComplaint,
  listComplaints,
  updateStatus,
  type ComplaintItem,
  type ComplaintStatus,
} from "../../../../lib/complaints";

const statusLabel: Record<ComplaintStatus, string> = {
  new: "Baru",
  reviewed: "Ditinjau",
  in_progress: "Diproses",
  waiting_information: "Menunggu informasi",
  resolved: "Selesai",
  rejected: "Ditolak",
  closed: "Ditutup",
};

function statusVariant(status: ComplaintStatus): StatusBadgeProps["variant"] {
  switch (status) {
    case "resolved":
    case "closed":
      return "success";
    case "rejected":
      return "danger";
    case "waiting_information":
      return "warning";
    default:
      return "info";
  }
}

type UserItem = {
  id: string;
  email?: string;
  phone?: string;
  status: string;
};

export default function PengurusAduanPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<ComplaintItem[]>([]);
  const [users, setUsers] = useState<UserItem[]>([]);
  const [status, setStatus] = useState<ComplaintStatus | "">("");
  const [category, setCategory] = useState("");
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<ComplaintItem>();
  const [assigneeID, setAssigneeID] = useState("");
  const [nextStatus, setNextStatus] = useState<ComplaintStatus | "">("");
  const [resolutionNote, setResolutionNote] = useState("");
  const [comment, setComment] = useState("");
  const [internal, setInternal] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      const [complaints, activeUsers] = await Promise.all([
        listComplaints(token, { status: status || undefined, category: category || undefined, search: search || undefined }),
        apiFetch<UserItem[]>("users", { headers: { Authorization: `Bearer ${token}` } }),
      ]);
      setItems(complaints);
      setUsers(activeUsers.filter((user) => user.status === "active"));
      if (selected) setSelected(complaints.find((item) => item.id === selected.id));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Data aduan gagal dimuat.");
    } finally {
      setLoading(false);
    }
  }, [category, getAccessToken, search, selected, status]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const assign = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected || !assigneeID) return;
    try {
      const token = await getAccessToken();
      if (!token) return;
      await assignComplaint(token, selected.id, { assigned_to: assigneeID });
      setNotice("Petugas ditugaskan.");
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Penugasan gagal.");
    }
  };

  const changeStatus = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected || !nextStatus) return;
    try {
      const token = await getAccessToken();
      if (!token) return;
      await updateStatus(token, selected.id, { status: nextStatus, resolution_note: resolutionNote || undefined });
      setNotice("Status aduan diperbarui.");
      setNextStatus("");
      setResolutionNote("");
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Status aduan gagal diperbarui.");
    }
  };

  const submitComment = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected || !comment.trim()) return;
    try {
      const token = await getAccessToken();
      if (!token) return;
      await addComment(token, selected.id, { body: comment, is_internal: internal });
      setComment("");
      setInternal(false);
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Komentar gagal dikirim.");
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-4)", margin: "0 auto", maxWidth: 760, padding: "var(--space-4)" }}>
      <header>
        <h1 style={{ fontSize: "1.25rem", margin: 0 }}>Antrean Aduan</h1>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>Tinjau, tugaskan, tindak lanjuti aduan warga.</p>
      </header>
      {error ? <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>{error}</p> : null}
      {notice ? <p role="status" style={{ color: "var(--color-success)", margin: 0 }}>{notice}</p> : null}

      {selected ? (
        <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", display: "grid", gap: "var(--space-3)", padding: "var(--space-4)" }}>
          <div style={{ alignItems: "start", display: "flex", gap: "var(--space-2)", justifyContent: "space-between" }}>
            <div><h2 style={{ fontSize: "1rem", margin: 0 }}>{selected.title}</h2><small>{selected.ticket_number} · Pelapor: {selected.reporter_name}</small></div>
            <StatusBadge variant={statusVariant(selected.status)}>{statusLabel[selected.status]}</StatusBadge>
          </div>
          <p style={{ margin: 0 }}>{selected.description}</p>
          {selected.location_description ? <p style={{ margin: 0 }}><strong>Lokasi:</strong> {selected.location_description}</p> : null}
          {selected.resolution_note ? <p style={{ background: "var(--color-surface-muted)", margin: 0, padding: "var(--space-2)" }}><strong>Resolusi:</strong> {selected.resolution_note}</p> : null}

          {!["resolved", "rejected", "closed"].includes(selected.status) ? <form onSubmit={assign} style={{ display: "grid", gap: "var(--space-2)" }}>
            <FormField label="Petugas tertugaskan">{(props) => <Select {...props} value={assigneeID || selected.assigned_to || ""} onChange={(event) => setAssigneeID(event.target.value)}><option value="">Pilih petugas</option>{users.map((user) => <option key={user.id} value={user.id}>{user.email ?? user.phone ?? user.id}</option>)}</Select>}</FormField>
            <div><Button type="submit" disabled={!assigneeID}>Simpan penugasan</Button></div>
          </form> : null}

          {!["resolved", "rejected", "closed"].includes(selected.status) ? <form onSubmit={changeStatus} style={{ display: "grid", gap: "var(--space-2)" }}>
            <FormField label="Status berikutnya">{(props) => <Select {...props} value={nextStatus} onChange={(event) => setNextStatus(event.target.value as ComplaintStatus)}><option value="">Pilih status</option><option value="reviewed">Ditinjau</option><option value="in_progress">Diproses</option><option value="waiting_information">Menunggu informasi</option><option value="resolved">Selesai</option><option value="rejected">Ditolak</option></Select>}</FormField>
            {nextStatus === "resolved" || nextStatus === "rejected" ? <FormField label="Catatan resolusi" required>{(props) => <TextInput {...props} required value={resolutionNote} onChange={(event) => setResolutionNote(event.target.value)} />}</FormField> : null}
            <div><Button type="submit" disabled={!nextStatus}>Perbarui status</Button></div>
          </form> : null}

          <div style={{ borderTop: "1px solid var(--color-border)", display: "grid", gap: "var(--space-2)", paddingTop: "var(--space-3)" }}>
            <h3 style={{ fontSize: "1rem", margin: 0 }}>Timeline</h3>
            {selected.comments.map((item) => <article key={item.id} style={{ background: item.is_internal ? "#fef3c7" : "var(--color-surface-muted)", borderRadius: "var(--radius-md)", padding: "var(--space-2)" }}><strong>{item.author_name}{item.is_internal ? " · Internal" : ""}</strong><br /><small>{new Date(item.created_at).toLocaleString("id-ID")}</small><p style={{ marginBottom: 0 }}>{item.body}</p></article>)}
            <form onSubmit={submitComment} style={{ display: "grid", gap: "var(--space-2)" }}>
              <TextInput required aria-label="Komentar" placeholder="Tulis pembaruan…" value={comment} onChange={(event) => setComment(event.target.value)} />
              <label style={{ display: "flex", gap: "var(--space-2)" }}><input type="checkbox" checked={internal} onChange={(event) => setInternal(event.target.checked)} /> Catatan internal</label>
              <div><Button type="submit">Kirim komentar</Button></div>
            </form>
          </div>
          <div><Button type="button" variant="outline" onClick={() => setSelected(undefined)}>Kembali</Button></div>
        </section>
      ) : (
        <>
          <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", display: "grid", gap: "var(--space-2)", padding: "var(--space-4)" }}>
            <Select aria-label="Filter status" value={status} onChange={(event) => setStatus(event.target.value as ComplaintStatus)}><option value="">Semua status</option>{Object.entries(statusLabel).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</Select>
            <TextInput aria-label="Filter kategori" placeholder="Kategori" value={category} onChange={(event) => setCategory(event.target.value)} />
            <TextInput aria-label="Cari aduan" placeholder="Cari tiket atau judul" value={search} onChange={(event) => setSearch(event.target.value)} />
          </section>
          <section>
            <h2 style={{ fontSize: "1rem" }}>Daftar aduan</h2>
            {loading ? <p>Memuat…</p> : items.length === 0 ? <EmptyState title="Tidak ada aduan" description="Aduan sesuai filter akan muncul di sini." /> : <div style={{ display: "grid", gap: "var(--space-3)" }}>{items.map((item) => <article key={item.id} style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", display: "grid", gap: "var(--space-2)", padding: "var(--space-3)" }}><div style={{ alignItems: "start", display: "flex", justifyContent: "space-between", gap: "var(--space-2)" }}><div><strong>{item.title}</strong><br /><small>{item.ticket_number} · {item.category} · {item.reporter_name}</small></div><StatusBadge variant={statusVariant(item.status)}>{statusLabel[item.status]}</StatusBadge></div><div><Button type="button" onClick={() => setSelected(item)}>Detail</Button></div></article>)}</div>}
          </section>
        </>
      )}
    </div>
  );
}