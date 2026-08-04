"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { TextInput } from "@/components/ui/TextInput";
import { useAuth } from "@/lib/auth-context";
import { AuditLogItem, getAuditLogs } from "@/lib/settings";

export default function AuditLogsPage() {
  const { getAccessToken } = useAuth();
  const [logs, setLogs] = useState<AuditLogItem[]>([]);
  const [nextCursor, setNextCursor] = useState<number | undefined>();
  const [hasMore, setHasMore] = useState(false);
  const [action, setAction] = useState("");
  const [actorUserID, setActorUserID] = useState("");
  const [entityType, setEntityType] = useState("");
  const [entityID, setEntityID] = useState("");
  const [filters, setFilters] = useState({
    action: "",
    actorUserID: "",
    entityType: "",
    entityID: "",
  });
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState<AuditLogItem | null>(null);

  const load = useCallback(async (cursor?: number) => {
    if (cursor) setLoadingMore(true);
    else setLoading(true);
    setError("");

    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi tidak valid.");

      const result = await getAuditLogs(token, {
        action: filters.action || undefined,
        actor_user_id: filters.actorUserID || undefined,
        entity_type: filters.entityType || undefined,
        entity_id: filters.entityID || undefined,
        cursor,
        limit: 20,
      });

      const data = Array.isArray(result?.data) ? result.data : [];
      setLogs((current) => (cursor ? [...current, ...data] : data));
      setNextCursor(result?.meta?.next_cursor);
      setHasMore(result?.meta?.has_more ?? false);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Gagal memuat audit log.");
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [filters, getAccessToken]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void load();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  function submitFilters(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFilters({
      action: action.trim(),
      actorUserID: actorUserID.trim(),
      entityType: entityType.trim(),
      entityID: entityID.trim(),
    });
  }

  const card: React.CSSProperties = {
    padding: "var(--space-4)",
    border: "1px solid var(--color-border)",
    borderRadius: "var(--radius-md)",
    background: "var(--color-surface)",
  };

  const payload = (value: Record<string, unknown>) => JSON.stringify(value, null, 2);

  return (
    <div style={{ maxWidth: 960, display: "flex", flexDirection: "column", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", margin: 0 }}>Audit Log</h1>
        <p style={{ margin: "var(--space-1) 0 0", color: "var(--color-text-secondary)" }}>
          Riwayat aktivitas penting. Read-only, append-only, data sensitif tersanitasi.
        </p>
      </header>

      <form onSubmit={submitFilters} style={{ ...card, display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))", gap: "var(--space-3)", alignItems: "end" }}>
        <FormField label="Aksi" id="filter-action">
          {(props) => <TextInput {...props} value={action} onChange={(event) => setAction(event.target.value)} placeholder="Contoh: payment.verify" />}
        </FormField>
        <FormField label="ID aktor" id="filter-actor">
          {(props) => <TextInput {...props} value={actorUserID} onChange={(event) => setActorUserID(event.target.value)} placeholder="UUID aktor" />}
        </FormField>
        <FormField label="Tipe entitas" id="filter-entity-type">
          {(props) => <TextInput {...props} value={entityType} onChange={(event) => setEntityType(event.target.value)} placeholder="Contoh: payment" />}
        </FormField>
        <FormField label="ID entitas" id="filter-entity-id">
          {(props) => <TextInput {...props} value={entityID} onChange={(event) => setEntityID(event.target.value)} placeholder="UUID entitas" />}
        </FormField>
        <Button type="submit">Terapkan filter</Button>
      </form>

      {error ? <p role="alert" style={{ ...card, margin: 0, color: "var(--color-danger-700)", background: "var(--color-danger-50)" }}>{error}</p> : null}

      {loading ? <p aria-live="polite">Memuat audit log…</p> : (
        <section style={{ display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
          {logs.length === 0 ? <p style={{ color: "var(--color-text-secondary)" }}>Tidak ada audit log yang sesuai.</p> : logs.map((log) => (
            <article key={log.id} style={{ ...card, display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
              <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-2)", flexWrap: "wrap" }}>
                <div>
                  <strong>{log.action}</strong>
                  <span style={{ marginLeft: "var(--space-2)", color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
                    {log.entity_type}{log.entity_id ? ` · ${log.entity_id.slice(0, 8)}` : ""}
                  </span>
                </div>
                <time dateTime={log.created_at} style={{ color: "var(--color-text-secondary)", fontSize: "0.8125rem" }}>
                  {new Date(log.created_at).toLocaleString("id-ID")}
                </time>
              </div>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
                Pelaku: {log.actor_name ?? log.actor_user_id ?? "Sistem"}{log.actor_role_codes.length ? ` [${log.actor_role_codes.join(", ")}]` : ""}
              </span>
              <div style={{ display: "flex", justifyContent: "flex-end" }}>
                <Button type="button" variant="secondary" onClick={() => setSelected(log)}>Lihat detail</Button>
              </div>
            </article>
          ))}

          {hasMore && nextCursor ? <Button type="button" variant="secondary" loading={loadingMore} onClick={() => void load(nextCursor)}>Muat lebih banyak</Button> : null}
        </section>
      )}

      {selected ? (
        <div role="dialog" aria-modal="true" aria-labelledby="audit-detail-title" style={{ position: "fixed", inset: 0, zIndex: 100, display: "flex", alignItems: "center", justifyContent: "center", padding: "var(--space-4)", background: "rgba(0, 0, 0, 0.5)" }}>
          <div style={{ ...card, width: "100%", maxWidth: 640, maxHeight: "90vh", overflowY: "auto", display: "flex", flexDirection: "column", gap: "var(--space-3)" }}>
            <h2 id="audit-detail-title" style={{ fontSize: "1.25rem", margin: 0 }}>Detail Audit #{selected.id}</h2>
            <p style={{ margin: 0, color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>Request ID: {selected.request_id ?? "Tidak tersedia"}</p>
            <Payload title="Metadata" value={payload(selected.metadata)} />
            <Payload title="Sebelum" value={payload(selected.before_data)} />
            <Payload title="Sesudah" value={payload(selected.after_data)} />
            <Button type="button" onClick={() => setSelected(null)}>Tutup</Button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function Payload({ title, value }: { title: string; value: string }) {
  return (
    <section>
      <h3 style={{ fontSize: "1rem", margin: "0 0 var(--space-2)" }}>{title}</h3>
      <pre style={{ margin: 0, padding: "var(--space-3)", overflowX: "auto", borderRadius: "var(--radius-sm)", background: "var(--color-surface-muted)", fontSize: "0.8125rem" }}>{value}</pre>
    </section>
  );
}