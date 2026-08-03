"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { useAuth } from "@/lib/auth-context";
import { ApiException } from "@/lib/api";
import {
  fetchNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
  type Notification,
} from "@/lib/notifications";

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export default function NotifikasiPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<Notification[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    const token = await getAccessToken();
    if (!token) {
      setError("Autentikasi dibutuhkan.");
      setLoading(false);
      return;
    }

    try {
      setItems(await fetchNotifications(token));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat notifikasi.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Initial API synchronization; load updates state after authentication resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  async function markRead(item: Notification) {
    if (item.read_at || updating) return;

    const token = await getAccessToken();
    if (!token) return;

    setUpdating(true);
    try {
      const updated = await markNotificationAsRead(token, item.id);
      setItems((current) => current.map((value) => (value.id === updated.id ? updated : value)));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memperbarui notifikasi.");
    } finally {
      setUpdating(false);
    }
  }

  async function markAllRead() {
    const token = await getAccessToken();
    if (!token) return;

    setUpdating(true);
    try {
      await markAllNotificationsAsRead(token);
      setItems((current) =>
        current.map((item) => (item.read_at ? item : { ...item, read_at: new Date().toISOString() })),
      );
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memperbarui notifikasi.");
    } finally {
      setUpdating(false);
    }
  }

  const unreadCount = items.filter((item) => !item.read_at).length;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
      <header
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "var(--space-3)",
        }}
      >
        <div>
          <h1 style={{ margin: 0, fontSize: "1.5rem", fontWeight: 700 }}>Notifikasi</h1>
          <p style={{ margin: "var(--space-1) 0 0", color: "var(--color-text-secondary)" }}>
            {unreadCount ? `${unreadCount} belum dibaca` : "Semua sudah dibaca"}
          </p>
        </div>
        {unreadCount > 0 && (
          <Button variant="outline" onClick={markAllRead} disabled={updating}>
            Tandai semua dibaca
          </Button>
        )}
      </header>

      {loading && <p aria-live="polite">Memuat notifikasi…</p>}
      {error && (
        <div role="alert" style={{ color: "var(--color-danger-700)" }}>
          <p>{error}</p>
          <Button variant="outline" onClick={() => void load()}>
            Coba lagi
          </Button>
        </div>
      )}
      {!loading && !error && items.length === 0 && (
        <EmptyState
          title="Belum ada notifikasi"
          description="Pembaruan penting terkait layanan RT akan tampil di sini."
        />
      )}
      {!loading && !error && items.length > 0 && (
        <ul
          aria-label="Daftar notifikasi"
          style={{ display: "flex", flexDirection: "column", gap: "var(--space-3)", margin: 0, padding: 0, listStyle: "none" }}
        >
          {items.map((item) => (
            <li key={item.id}>
              <button
                type="button"
                onClick={() => void markRead(item)}
                disabled={updating}
                aria-label={`${item.read_at ? "Dibaca" : "Belum dibaca"}: ${item.title}`}
                style={{
                  width: "100%",
                  padding: "var(--space-4)",
                  border: "1px solid var(--color-border)",
                  borderRadius: "var(--radius-md)",
                  background: item.read_at ? "var(--color-surface)" : "var(--color-primary-50)",
                  color: "var(--color-text)",
                  cursor: item.read_at ? "default" : "pointer",
                  textAlign: "left",
                }}
              >
                <strong style={{ display: "block", fontSize: "1rem" }}>{item.title}</strong>
                {item.body && <span style={{ display: "block", marginTop: "var(--space-1)" }}>{item.body}</span>}
                <time
                  dateTime={item.created_at}
                  style={{ display: "block", marginTop: "var(--space-2)", color: "var(--color-text-secondary)", fontSize: "0.8125rem" }}
                >
                  {formatDate(item.created_at)}
                </time>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}