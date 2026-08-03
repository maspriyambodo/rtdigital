"use client";

import Link from "next/link";
import { useCallback, useEffect, useState, type CSSProperties } from "react";

import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { useAuth } from "@/lib/auth-context";
import { getResidentDashboard, type ResidentDashboard } from "@/lib/dashboard";

export default function WargaHomePage() {
  const { getAccessToken } = useAuth();
  const [data, setData] = useState<ResidentDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setData(await getResidentDashboard(getAccessToken));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Gagal memuat dashboard warga.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  return (
    <div style={pageStyle}>
      <header>
        <h1 style={titleStyle}>Beranda Warga</h1>
        <p style={descriptionStyle}>Ringkasan layanan dan informasi lingkungan RT.</p>
      </header>

      {loading ? <p style={descriptionStyle}>Memuat ringkasan warga…</p> : null}
      {error ? (
        <EmptyState
          title="Gagal memuat dashboard"
          description={error}
          action={<Button variant="outline" onClick={() => void load()}>Coba lagi</Button>}
        />
      ) : null}

      {!loading && !error && data ? (
        <>
          <section style={sectionStyle}>
            <div style={sectionHeaderStyle}>
              <h2 style={sectionTitleStyle}>Tagihan Aktif</h2>
              <Link href="/warga/tagihan" style={linkStyle}>Lihat semua</Link>
            </div>
            {data.active_invoices.length ? data.active_invoices.map((invoice) => (
              <div key={invoice.id} style={rowStyle}>
                <div>
                  <strong>{invoice.due_type_name}</strong>
                  <p style={smallStyle}>Jatuh tempo: {invoice.due_date}</p>
                </div>
                <div style={{ textAlign: "right" }}>
                  <strong>Rp {(invoice.amount - invoice.paid_amount).toLocaleString("id-ID")}</strong>
                  <br />
                  <StatusBadge variant={invoice.status === "pending_verification" ? "warning" : "danger"}>
                    {invoice.status === "pending_verification" ? "Menunggu verifikasi" : "Belum lunas"}
                  </StatusBadge>
                </div>
              </div>
            )) : <p style={smallStyle}>Tidak ada tagihan aktif.</p>}
          </section>

          <section style={sectionStyle}>
            <div style={sectionHeaderStyle}>
              <h2 style={sectionTitleStyle}>Pengumuman</h2>
              <Link href="/warga/pengumuman" style={linkStyle}>Lihat semua</Link>
            </div>
            {data.announcements.length ? data.announcements.map((announcement) => (
              <div key={announcement.id} style={rowStyle}>
                <div>
                  <strong>{announcement.title}</strong>
                  <p style={smallStyle}>{announcement.category} · {formatDate(announcement.publish_at)}</p>
                </div>
                {announcement.priority === "important" ? <StatusBadge variant="danger">Penting</StatusBadge> : null}
              </div>
            )) : <p style={smallStyle}>Belum ada pengumuman aktif.</p>}
          </section>

          <section style={sectionStyle}>
            <div style={sectionHeaderStyle}>
              <h2 style={sectionTitleStyle}>Agenda Mendatang</h2>
            </div>
            {data.upcoming_events.length ? data.upcoming_events.map((event) => (
              <div key={event.id} style={rowStyle}>
                <div>
                  <strong>{event.title}</strong>
                  <p style={smallStyle}>{event.location || "Lokasi belum ditentukan"}</p>
                </div>
                <span style={smallStyle}>{formatDate(event.starts_at)}</span>
              </div>
            )) : <p style={smallStyle}>Belum ada agenda mendatang.</p>}
          </section>

          <div style={gridStyle}>
            <SummaryCard title="Pembayaran terbaru" href="/warga/tagihan" count={data.recent_payments.length} />
            <SummaryCard title="Pengajuan surat" href="/warga/surat" count={data.recent_letters.length} />
            <SummaryCard title="Aduan saya" href="/warga/aduan" count={data.recent_complaints.length} />
          </div>
        </>
      ) : null}
    </div>
  );
}

function SummaryCard({ title, href, count }: { title: string; href: string; count: number }) {
  return (
    <Link href={href} style={{ ...sectionStyle, textDecoration: "none" }}>
      <strong style={{ fontSize: "1.5rem", color: "var(--color-primary-600)" }}>{count}</strong>
      <span style={smallStyle}>{title}</span>
    </Link>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}

const pageStyle: CSSProperties = { display: "flex", flexDirection: "column", gap: "var(--space-5)" };
const titleStyle: CSSProperties = { margin: 0, fontSize: "1.875rem", fontWeight: 700, color: "var(--color-text)" };
const descriptionStyle: CSSProperties = { margin: "var(--space-1) 0 0", color: "var(--color-text-secondary)" };
const sectionStyle: CSSProperties = { display: "flex", flexDirection: "column", gap: "var(--space-3)", padding: "var(--space-4)", background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)" };
const sectionHeaderStyle: CSSProperties = { display: "flex", justifyContent: "space-between", alignItems: "center", gap: "var(--space-3)" };
const sectionTitleStyle: CSSProperties = { margin: 0, fontSize: "1.125rem" };
const rowStyle: CSSProperties = { display: "flex", justifyContent: "space-between", alignItems: "center", gap: "var(--space-3)", paddingBottom: "var(--space-2)", borderBottom: "1px solid var(--color-border)" };
const smallStyle: CSSProperties = { margin: 0, fontSize: "0.8125rem", color: "var(--color-text-secondary)" };
const linkStyle: CSSProperties = { color: "var(--color-primary-600)", textDecoration: "none", fontSize: "0.875rem", fontWeight: 600 };
const gridStyle: CSSProperties = { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(140px, 1fr))", gap: "var(--space-3)" };