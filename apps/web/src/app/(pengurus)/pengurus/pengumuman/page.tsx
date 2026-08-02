"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FileUploader } from "@/components/ui/FileUploader";
import { FormField } from "@/components/ui/FormField";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";
import { ApiException } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import {
  type AnnouncementCategory,
  type AnnouncementItem,
  type AnnouncementPriority,
  type AnnouncementStatus,
  type EventItem,
  type EventStatus,
  archiveAnnouncement,
  cancelEvent,
  createAnnouncement,
  createEvent,
  getAnnouncementReadStats,
  listAnnouncements,
  listEvents,
  publishAnnouncement,
  updateAnnouncement,
  updateEvent,
} from "@/lib/communication";

const announcementCategories: Record<AnnouncementCategory, string> = {
  general: "Umum",
  security: "Keamanan",
  health: "Kesehatan",
  billing: "Iuran",
  event: "Kegiatan",
  emergency: "Darurat",
};

const announcementStatuses: Record<AnnouncementStatus, string> = {
  draft: "Draft",
  scheduled: "Terjadwal",
  published: "Terbit",
  archived: "Arsip",
};

const eventStatuses: Record<EventStatus, string> = {
  planned: "Direncanakan",
  ongoing: "Berlangsung",
  completed: "Selesai",
  cancelled: "Dibatalkan",
};

const cardStyle = {
  display: "grid",
  gap: "var(--space-3)",
  padding: "var(--space-4)",
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-lg)",
  background: "var(--color-surface)",
} as const;

const localDateTime = (value?: string) =>
  value ? new Date(value).toISOString().slice(0, 16) : "";

export default function PengumumanAgendaPengurusPage() {
  const { getAccessToken } = useAuth();
  const [tab, setTab] = useState<"announcements" | "events">("announcements");
  const [announcements, setAnnouncements] = useState<AnnouncementItem[]>([]);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [editingAnnouncement, setEditingAnnouncement] = useState<AnnouncementItem>();
  const [editingEvent, setEditingEvent] = useState<EventItem>();
  const [readStats, setReadStats] = useState<{ id: string; audience: number; reads: number }>();

  const [announcement, setAnnouncement] = useState({
    title: "",
    content: "",
    category: "general" as AnnouncementCategory,
    priority: "normal" as AnnouncementPriority,
    status: "draft" as Exclude<AnnouncementStatus, "archived">,
    publishAt: "",
    expireAt: "",
    targetType: "all" as "all" | "role" | "household" | "house_unit",
    targetID: "",
    attachmentIDs: [] as string[],
  });
  const [announcementUploadID, setAnnouncementUploadID] = useState(() => crypto.randomUUID());

  const [event, setEvent] = useState({
    title: "",
    description: "",
    location: "",
    startsAt: "",
    endsAt: "",
    status: "planned" as EventStatus,
    attachmentIDs: [] as string[],
  });
  const [eventUploadID, setEventUploadID] = useState(() => crypto.randomUUID());

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      const [announcementList, eventList] = await Promise.all([
        listAnnouncements(token),
        listEvents(token),
      ]);
      setAnnouncements(announcementList);
      setEvents(eventList);
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat komunikasi.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const resetAnnouncement = () => {
    setEditingAnnouncement(undefined);
    setAnnouncement({
      title: "",
      content: "",
      category: "general",
      priority: "normal",
      status: "draft",
      publishAt: "",
      expireAt: "",
      targetType: "all",
      targetID: "",
      attachmentIDs: [],
    });
    setAnnouncementUploadID(crypto.randomUUID());
  };

  const resetEvent = () => {
    setEditingEvent(undefined);
    setEvent({
      title: "",
      description: "",
      location: "",
      startsAt: "",
      endsAt: "",
      status: "planned",
      attachmentIDs: [],
    });
    setEventUploadID(crypto.randomUUID());
  };

  const saveAnnouncement = async (formEvent: React.FormEvent) => {
    formEvent.preventDefault();
    setSaving(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      const payload = {
        title: announcement.title.trim(),
        content: announcement.content.trim(),
        category: announcement.category,
        priority: announcement.priority,
        status: announcement.status,
        publish_at: announcement.publishAt ? new Date(announcement.publishAt).toISOString() : undefined,
        expire_at: announcement.expireAt ? new Date(announcement.expireAt).toISOString() : undefined,
        targets: [{
          target_type: announcement.targetType,
          target_id: announcement.targetType === "all" ? undefined : announcement.targetID.trim(),
        }],
        attachment_file_ids: announcement.attachmentIDs,
      };
      if (editingAnnouncement) {
        await updateAnnouncement(token, editingAnnouncement.id, payload);
        setMessage("Pengumuman diperbarui.");
      } else {
        await createAnnouncement(token, payload);
        setMessage("Pengumuman disimpan.");
      }
      resetAnnouncement();
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal menyimpan pengumuman.");
    } finally {
      setSaving(false);
    }
  };

  const saveEvent = async (formEvent: React.FormEvent) => {
    formEvent.preventDefault();
    setSaving(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      const payload = {
        title: event.title.trim(),
        description: event.description.trim() || undefined,
        location: event.location.trim() || undefined,
        starts_at: new Date(event.startsAt).toISOString(),
        ends_at: event.endsAt ? new Date(event.endsAt).toISOString() : undefined,
        status: event.status,
        attachment_file_ids: event.attachmentIDs,
      };
      if (editingEvent) {
        await updateEvent(token, editingEvent.id, payload);
        setMessage("Agenda diperbarui.");
      } else {
        await createEvent(token, payload);
        setMessage("Agenda disimpan.");
      }
      resetEvent();
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal menyimpan agenda.");
    } finally {
      setSaving(false);
    }
  };

  const act = async (action: () => Promise<unknown>, success: string) => {
    setSaving(true);
    setError("");
    try {
      await action();
      setMessage(success);
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Tindakan gagal.");
    } finally {
      setSaving(false);
    }
  };

  const editAnnouncement = (item: AnnouncementItem) => {
    const target = item.targets[0];
    setEditingAnnouncement(item);
    setAnnouncement({
      title: item.title,
      content: item.content,
      category: item.category,
      priority: item.priority,
      status: item.status === "archived" ? "draft" : item.status,
      publishAt: localDateTime(item.publish_at),
      expireAt: localDateTime(item.expire_at),
      targetType: target?.target_type ?? "all",
      targetID: target?.target_id ?? "",
      attachmentIDs: item.attachments.map((attachment) => attachment.file_id),
    });
    setAnnouncementUploadID(item.id);
    setTab("announcements");
  };

  const editEvent = (item: EventItem) => {
    setEditingEvent(item);
    setEvent({
      title: item.title,
      description: item.description ?? "",
      location: item.location ?? "",
      startsAt: localDateTime(item.starts_at),
      endsAt: localDateTime(item.ends_at),
      status: item.status,
      attachmentIDs: item.attachments.map((attachment) => attachment.file_id),
    });
    setEventUploadID(item.id);
    setTab("events");
  };

  const showStats = async (id: string) => {
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      const stats = await getAnnouncementReadStats(token, id);
      setReadStats({ id, audience: stats.total_audience, reads: stats.read_count });
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat statistik baca.");
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Pengumuman & Agenda</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
          Kelola informasi dan kegiatan RT.
        </p>
      </header>

      <div role="tablist" aria-label="Komunikasi RT" style={{ display: "flex", gap: "var(--space-2)" }}>
        <Button type="button" variant={tab === "announcements" ? "primary" : "secondary"} onClick={() => setTab("announcements")}>
          Pengumuman
        </Button>
        <Button type="button" variant={tab === "events" ? "primary" : "secondary"} onClick={() => setTab("events")}>
          Agenda
        </Button>
      </div>

      {message ? <p role="status" style={{ color: "var(--color-success-text)" }}>{message}</p> : null}
      {error ? <p role="alert" style={{ color: "var(--color-danger-text)" }}>{error}</p> : null}

      {tab === "announcements" ? (
        <form onSubmit={(formEvent) => void saveAnnouncement(formEvent)} style={cardStyle}>
          <h2 style={{ fontSize: "1.125rem" }}>{editingAnnouncement ? "Ubah Pengumuman" : "Buat Pengumuman"}</h2>
          <FormField label="Judul">{({ id }) => <TextInput id={id} value={announcement.title} onChange={(event) => setAnnouncement((current) => ({ ...current, title: event.target.value }))} required />}</FormField>
          <FormField label="Isi pengumuman">
            {({ id }) => <textarea id={id} value={announcement.content} onChange={(event) => setAnnouncement((current) => ({ ...current, content: event.target.value }))} required style={{ minHeight: 120 }} />}
          </FormField>
          <div style={formGrid}>
            <FormField label="Kategori">{({ id }) => <select id={id} value={announcement.category} onChange={(event) => setAnnouncement((current) => ({ ...current, category: event.target.value as AnnouncementCategory }))}>{Object.entries(announcementCategories).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select>}</FormField>
            <FormField label="Prioritas">{({ id }) => <select id={id} value={announcement.priority} onChange={(event) => setAnnouncement((current) => ({ ...current, priority: event.target.value as AnnouncementPriority }))}><option value="normal">Normal</option><option value="important">Penting</option></select>}</FormField>
            <FormField label="Status">{({ id }) => <select id={id} value={announcement.status} onChange={(event) => setAnnouncement((current) => ({ ...current, status: event.target.value as Exclude<AnnouncementStatus, "archived"> }))}><option value="draft">Draft</option><option value="scheduled">Terjadwal</option><option value="published">Terbit sekarang</option></select>}</FormField>
          </div>
          {announcement.status === "scheduled" ? <FormField label="Jadwal terbit">{({ id }) => <TextInput id={id} type="datetime-local" value={announcement.publishAt} onChange={(event) => setAnnouncement((current) => ({ ...current, publishAt: event.target.value }))} required />}</FormField> : null}
          <FormField label="Kedaluwarsa (opsional)">{({ id }) => <TextInput id={id} type="datetime-local" value={announcement.expireAt} onChange={(event) => setAnnouncement((current) => ({ ...current, expireAt: event.target.value }))} />}</FormField>
          <div style={formGrid}>
            <FormField label="Target">{({ id }) => <select id={id} value={announcement.targetType} onChange={(event) => setAnnouncement((current) => ({ ...current, targetType: event.target.value as typeof current.targetType, targetID: "" }))}><option value="all">Semua warga</option><option value="role">Peran</option><option value="household">Keluarga</option><option value="house_unit">Rumah/unit</option></select>}</FormField>
            {announcement.targetType !== "all" ? <FormField label="ID target">{({ id }) => <TextInput id={id} value={announcement.targetID} onChange={(event) => setAnnouncement((current) => ({ ...current, targetID: event.target.value }))} required />}</FormField> : null}
          </div>
          <FormField label="Lampiran (opsional)">
            {() => <FileUploader entityType="announcement" entityId={announcementUploadID} disabled={saving} onChange={(fileID) => fileID && setAnnouncement((current) => ({ ...current, attachmentIDs: [...current.attachmentIDs, fileID] }))} />}
          </FormField>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
            <Button type="submit" disabled={saving}>{editingAnnouncement ? "Simpan perubahan" : "Simpan pengumuman"}</Button>
            {editingAnnouncement ? <Button type="button" variant="secondary" onClick={resetAnnouncement}>Batal</Button> : null}
          </div>
        </form>
      ) : (
        <form onSubmit={(formEvent) => void saveEvent(formEvent)} style={cardStyle}>
          <h2 style={{ fontSize: "1.125rem" }}>{editingEvent ? "Ubah Agenda" : "Buat Agenda"}</h2>
          <FormField label="Nama kegiatan">{({ id }) => <TextInput id={id} value={event.title} onChange={(inputEvent) => setEvent((current) => ({ ...current, title: inputEvent.target.value }))} required />}</FormField>
          <FormField label="Deskripsi (opsional)">{({ id }) => <textarea id={id} value={event.description} onChange={(inputEvent) => setEvent((current) => ({ ...current, description: inputEvent.target.value }))} style={{ minHeight: 96 }} />}</FormField>
          <div style={formGrid}>
            <FormField label="Lokasi">{({ id }) => <TextInput id={id} value={event.location} onChange={(inputEvent) => setEvent((current) => ({ ...current, location: inputEvent.target.value }))} />}</FormField>
            <FormField label="Status">{({ id }) => <select id={id} value={event.status} onChange={(inputEvent) => setEvent((current) => ({ ...current, status: inputEvent.target.value as EventStatus }))}>{Object.entries(eventStatuses).filter(([value]) => value !== "cancelled").map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select>}</FormField>
            <FormField label="Mulai">{({ id }) => <TextInput id={id} type="datetime-local" value={event.startsAt} onChange={(inputEvent) => setEvent((current) => ({ ...current, startsAt: inputEvent.target.value }))} required />}</FormField>
            <FormField label="Selesai">{({ id }) => <TextInput id={id} type="datetime-local" value={event.endsAt} onChange={(inputEvent) => setEvent((current) => ({ ...current, endsAt: inputEvent.target.value }))} />}</FormField>
          </div>
          <FormField label="Lampiran (opsional)">
            {() => <FileUploader entityType="event" entityId={eventUploadID} disabled={saving} onChange={(fileID) => fileID && setEvent((current) => ({ ...current, attachmentIDs: [...current.attachmentIDs, fileID] }))} />}
          </FormField>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
            <Button type="submit" disabled={saving}>{editingEvent ? "Simpan perubahan" : "Simpan agenda"}</Button>
            {editingEvent ? <Button type="button" variant="secondary" onClick={resetEvent}>Batal</Button> : null}
          </div>
        </form>
      )}

      {readStats ? <section style={cardStyle}><strong>Statistik baca</strong><span>{readStats.reads} dari {readStats.audience} penerima membaca pengumuman.</span><Button type="button" variant="secondary" onClick={() => setReadStats(undefined)}>Tutup</Button></section> : null}

      <section style={{ display: "grid", gap: "var(--space-3)" }}>
        <h2 style={{ fontSize: "1.125rem" }}>{tab === "announcements" ? "Daftar Pengumuman" : "Daftar Agenda"}</h2>
        {loading ? <p>Memuat data…</p> : tab === "announcements" ? (
          announcements.length === 0 ? <EmptyState title="Belum ada pengumuman" description="Buat pengumuman pertama untuk warga." /> : announcements.map((item) => (
            <article key={item.id} style={cardStyle}>
              <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}><strong>{item.title}</strong><StatusBadge variant={item.status === "published" ? "success" : item.status === "archived" ? "neutral" : "warning"}>{announcementStatuses[item.status]}</StatusBadge></div>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>{announcementCategories[item.category]} · Dibaca {item.read_count}</span>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
                {item.status !== "archived" ? <Button type="button" variant="secondary" onClick={() => editAnnouncement(item)}>Ubah</Button> : null}
                {item.status === "draft" || item.status === "scheduled" ? <Button type="button" disabled={saving} onClick={() => void act(async () => { const token = await getAccessToken(); if (!token) throw new Error("Sesi telah berakhir."); await publishAnnouncement(token, item.id); }, "Pengumuman diterbitkan.")}>Terbitkan</Button> : null}
                {item.status !== "archived" ? <Button type="button" variant="danger" disabled={saving} onClick={() => void act(async () => { const token = await getAccessToken(); if (!token) throw new Error("Sesi telah berakhir."); await archiveAnnouncement(token, item.id); }, "Pengumuman diarsipkan.")}>Arsipkan</Button> : null}
                <Button type="button" variant="outline" onClick={() => void showStats(item.id)}>Statistik baca</Button>
              </div>
            </article>
          ))
        ) : events.length === 0 ? <EmptyState title="Belum ada agenda" description="Buat agenda kegiatan RT." /> : events.map((item) => (
          <article key={item.id} style={cardStyle}>
            <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}><strong>{item.title}</strong><StatusBadge variant={item.status === "cancelled" ? "danger" : "neutral"}>{eventStatuses[item.status]}</StatusBadge></div>
            <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>{new Date(item.starts_at).toLocaleString("id-ID")} · {item.location ?? "Lokasi belum ditentukan"}</span>
            {item.status !== "cancelled" && item.status !== "completed" ? <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}><Button type="button" variant="secondary" onClick={() => editEvent(item)}>Ubah</Button><Button type="button" variant="danger" disabled={saving} onClick={() => void act(async () => { const token = await getAccessToken(); if (!token) throw new Error("Sesi telah berakhir."); await cancelEvent(token, item.id); }, "Agenda dibatalkan.")}>Batalkan</Button></div> : null}
          </article>
        ))}
      </section>
    </div>
  );
}

const formGrid = {
  display: "grid",
  gap: "var(--space-3)",
  gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 200px), 1fr))",
} as const;