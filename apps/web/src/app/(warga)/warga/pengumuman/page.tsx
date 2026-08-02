"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import {
  type AnnouncementCategory,
  type AnnouncementItem,
  type EventItem,
  type EventStatus,
  getAnnouncement,
  getEvent,
  listAnnouncements,
  listEvents,
} from "@/lib/communication";

const announcementCategories: Record<AnnouncementCategory, string> = {
  general: "Umum",
  security: "Keamanan",
  health: "Kesehatan",
  billing: "Iuran",
  event: "Kegiatan",
  emergency: "Darurat",
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

export default function PengumumanAgendaWargaPage() {
  const { getAccessToken } = useAuth();
  const [tab, setTab] = useState<"announcements" | "events">("announcements");
  const [announcements, setAnnouncements] = useState<AnnouncementItem[]>([]);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [detailAnnouncement, setDetailAnnouncement] = useState<AnnouncementItem>();
  const [detailEvent, setDetailEvent] = useState<EventItem>();

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      const [announcementList, eventList] = await Promise.all([
        listAnnouncements(token),
        listEvents(token, { upcoming: true }),
      ]);
      setAnnouncements(announcementList);
      setEvents(eventList);
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat informasi.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const viewAnnouncement = async (id: string) => {
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      setDetailAnnouncement(await getAnnouncement(token, id));
      setDetailEvent(undefined);
      await load(); // Refresh penanda telah dibaca.
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat detail pengumuman.");
    }
  };

  const viewEvent = async (id: string) => {
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");
      setDetailEvent(await getEvent(token, id));
      setDetailAnnouncement(undefined);
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat detail agenda.");
    }
  };

  const downloadAttachment = async (fileID: string) => {
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir.");
      const { download_url } = await apiFetch<{ download_url: string }>(`files/${fileID}/download`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      window.open(download_url, "_blank", "noopener,noreferrer");
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal mengunduh lampiran.");
    }
  };

  const attachments = (items: AnnouncementItem["attachments"]) => (
    items.length ? (
      <div style={{ borderTop: "1px solid var(--color-border)", paddingTop: "var(--space-3)" }}>
        <strong>Lampiran</strong>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)", marginTop: "var(--space-2)" }}>
          {items.map((item) => (
            <Button key={item.attachment_id} type="button" variant="outline" onClick={() => void downloadAttachment(item.file_id)}>
              Unduh: {item.original_name}
            </Button>
          ))}
        </div>
      </div>
    ) : null
  );

  const detail = detailAnnouncement ?? detailEvent;
  const isAnnouncement = !!detailAnnouncement;

  return (
    <div style={{ display: "grid", gap: "var(--space-6)", paddingBottom: 80 }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Pusat Informasi RT</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
          Pengumuman resmi dan agenda kegiatan warga.
        </p>
      </header>

      <div role="tablist" aria-label="Informasi warga" style={{ display: "flex", gap: "var(--space-2)" }}>
        <Button type="button" variant={tab === "announcements" ? "primary" : "secondary"} onClick={() => { setTab("announcements"); setDetailAnnouncement(undefined); setDetailEvent(undefined); }}>
          Pengumuman{announcements.some((item) => !item.is_read) ? " • Baru" : ""}
        </Button>
        <Button type="button" variant={tab === "events" ? "primary" : "secondary"} onClick={() => { setTab("events"); setDetailAnnouncement(undefined); setDetailEvent(undefined); }}>
          Agenda Mendatang
        </Button>
      </div>

      {error ? <p role="alert" style={{ color: "var(--color-danger-text)" }}>{error}</p> : null}

      {detail ? (
        <article style={cardStyle}>
          <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}>
            <h2 style={{ fontSize: "1.25rem", margin: 0 }}>{detail.title}</h2>
            <Button type="button" variant="outline" onClick={() => { setDetailAnnouncement(undefined); setDetailEvent(undefined); }}>Tutup</Button>
          </div>
          {isAnnouncement ? (
            <>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
                {announcementCategories[detailAnnouncement.category]} · {detailAnnouncement.author_name} · {new Date(detailAnnouncement.publish_at ?? detailAnnouncement.created_at).toLocaleString("id-ID")}
              </span>
              <p style={{ whiteSpace: "pre-wrap", margin: 0 }}>{detailAnnouncement.content}</p>
              {attachments(detailAnnouncement.attachments)}
            </>
          ) : detailEvent ? (
            <>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
                {eventStatuses[detailEvent.status]} · {detailEvent.author_name}
              </span>
              {detailEvent.description ? <p style={{ whiteSpace: "pre-wrap", margin: 0 }}>{detailEvent.description}</p> : null}
              <span><strong>Lokasi:</strong> {detailEvent.location ?? "Belum ditentukan"}</span>
              <span><strong>Mulai:</strong> {new Date(detailEvent.starts_at).toLocaleString("id-ID")}</span>
              {detailEvent.ends_at ? <span><strong>Selesai:</strong> {new Date(detailEvent.ends_at).toLocaleString("id-ID")}</span> : null}
              {attachments(detailEvent.attachments)}
            </>
          ) : null}
        </article>
      ) : null}

      <section style={{ display: "grid", gap: "var(--space-3)" }}>
        <h2 style={{ fontSize: "1.125rem" }}>{tab === "announcements" ? "Semua Pengumuman" : "Agenda Kegiatan"}</h2>
        {loading ? <p>Memuat informasi…</p> : tab === "announcements" ? (
          announcements.length === 0 ? <EmptyState title="Belum ada pengumuman" description="Informasi resmi RT akan muncul di sini." /> : announcements.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => void viewAnnouncement(item.id)}
              style={{ ...cardStyle, cursor: "pointer", textAlign: "left", borderLeft: item.priority === "important" ? "4px solid var(--color-danger)" : undefined, opacity: item.is_read ? 0.75 : 1 }}
            >
              <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}>
                <strong>{item.title}{!item.is_read ? " • Baru" : ""}</strong>
                <StatusBadge variant={item.priority === "important" ? "danger" : "neutral"}>{item.priority === "important" ? "Penting" : announcementCategories[item.category]}</StatusBadge>
              </div>
              <span style={{ fontSize: "0.875rem", color: "var(--color-text-secondary)" }}>
                {new Date(item.publish_at ?? item.created_at).toLocaleDateString("id-ID")} · {item.author_name}
              </span>
            </button>
          ))
        ) : events.length === 0 ? <EmptyState title="Tidak ada kegiatan terdekat" description="Agenda kegiatan RT mendatang akan muncul di sini." /> : events.map((item) => (
          <button key={item.id} type="button" onClick={() => void viewEvent(item.id)} style={{ ...cardStyle, cursor: "pointer", textAlign: "left" }}>
            <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)" }}>
              <strong>{item.title}</strong>
              <StatusBadge variant={item.status === "ongoing" ? "success" : "neutral"}>{eventStatuses[item.status]}</StatusBadge>
            </div>
            <span style={{ fontSize: "0.875rem", color: "var(--color-text-secondary)" }}>
              {new Date(item.starts_at).toLocaleString("id-ID")} · {item.location ?? "Lokasi belum ditentukan"}
            </span>
          </button>
        ))}
      </section>
    </div>
  );
}