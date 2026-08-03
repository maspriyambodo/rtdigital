"use client";

import Link from "next/link";
import { useCallback, useEffect, useState, type CSSProperties } from "react";

import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { useAuth } from "@/lib/auth-context";
import { getAdminDashboard, type AdminDashboard } from "@/lib/dashboard";

export default function PengurusHomePage() {
  const { getAccessToken } = useAuth();
  const [data, setData] = useState<AdminDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setData(await getAdminDashboard(getAccessToken));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Gagal memuat dashboard pengurus.");
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
      <header style={headerStyle}>
        <div>
          <h1 style={titleStyle}>Dashboard Pengurus</h1>
          <p style={descriptionStyle}>Ringkasan operasional dan keuangan lingkungan RT.</p>
        </div>
        <Link href="/pengurus/laporan" style={reportLinkStyle}>
          Menu Laporan & Ekspor
        </Link>
      </header>

      {loading ? <p style={descriptionStyle}>Memuat ringkasan pengurus…</p> : null}
      {error ? (
        <EmptyState
          title="Gagal memuat dashboard"
          description={error}
          action={
            <Button variant="outline" onClick={() => void load()}>
              Coba lagi
            </Button>
          }
        />
      ) : null}

      {!loading && !error && data ? (
        <>
          <div style={gridStyle}>
            <StatCard title="Keluarga Aktif" value={data.active_households} />
            <StatCard title="Warga Aktif" value={data.active_residents} />
            <StatCard title="Tagihan Aktif" value={data.active_invoices} />
            <StatCard title="Total Tunggakan" value={`Rp ${data.outstanding_amount.toLocaleString("id-ID")}`} />
            <StatCard title="Menunggu Verifikasi" value={data.pending_payments} />
            <StatCard title="Saldo Kas" value={`Rp ${data.cash_balance.toLocaleString("id-ID")}`} />
            <StatCard title="Antrean Surat" value={data.pending_letters} />
            <StatCard title="Aduan Terbuka" value={data.open_complaints} />
          </div>

          <section style={sectionStyle}>
            <div style={sectionHeaderStyle}>
              <h2 style={sectionTitleStyle}>Pembayaran Terbaru</h2>
              <Link href="/pengurus/tagihan" style={linkStyle}>Urus Pembayaran</Link>
            </div>
            {data.recent_payments.length ? data.recent_payments.map((payment) => (
              <div key={payment.id} style={rowStyle}>
                <div>
                  <strong>{payment.payment_number}</strong>
                  <p style={smallStyle}>Tagihan: {payment.invoice_number} · {formatDate(payment.paid_at)}</p>
                </div>
                <div style={{ textAlign: "right" }}>
                  <strong>Rp {payment.amount.toLocaleString("id-ID")}</strong>
                  <br />
                  <StatusBadge variant={payment.verification_status === "verified" ? "success" : payment.verification_status === "pending" ? "warning" : "danger"}>
                    {payment.verification_status}
                  </StatusBadge>
                </div>
              </div>
            )) : <p style={smallStyle}>Belum ada riwayat pembayaran.</p>}
          </section>

          <section style={sectionStyle}>
            <div style={sectionHeaderStyle}>
              <h2 style={sectionTitleStyle}>Antrean Surat</h2>
              <Link href="/pengurus/surat" style={linkStyle}>Proses Surat</Link>
            </div>
            {data.recent_letters.length ? data.recent_letters.map((letter) => (
              <div key={letter.id} style={rowStyle}>
                <div>
                  <strong>{letter.request_number}</strong>
                  <p style={smallStyle}>{letter.letter_type} · {formatDate(letter.updated_at)}</p>
                </div>
                <StatusBadge variant={letter.status === "issued" ? "success" : letter.status === "rejected" ? "danger" : "warning"}>
                  {letter.status}
                </StatusBadge>
              </div>
            )) : <p style={smallStyle}>Antrean surat kosong.</p>}
          </section>

          <section style={sectionStyle}>
            <div style={sectionHeaderStyle}>
              <h2 style={sectionTitleStyle}>Aduan Terbaru</h2>
              <Link href="/pengurus/aduan" style={linkStyle}>Tinjau Aduan</Link>
            </div>
            {data.recent_complaints.length ? data.recent_complaints.map((complaint) => (
              <div key={complaint.id} style={rowStyle}>
                <div>
                  <strong>[{complaint.ticket_number}] {complaint.title}</strong>
                  <p style={smallStyle}>Prioritas: {complaint.priority} · {formatDate(complaint.updated_at)}</p>
                </div>
                <StatusBadge variant={complaint.status === "resolved" || complaint.status === "closed" ? "success" : complaint.status === "rejected" ? "danger" : "warning"}>
                  {complaint.status}
                </StatusBadge>
              </div>
            )) : <p style={smallStyle}>Antrean aduan kosong.</p>}
          </section>
        </>
      ) : null}
    </div>
  );
}

function StatCard({ title, value }: { title: string; value: string | number }) {
  return (
    <div style={statCardStyle}>
      <span style={smallStyle}>{title}</span>
      <strong style={{ fontSize: "1.25rem", color: "var(--color-text)" }}>{value}</strong>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}

const pageStyle: CSSProperties = { display: "flex", flexDirection: "column", gap: "var(--space-5)" };
const headerStyle: CSSProperties = { display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-3)", flexWrap: "wrap" };
const titleStyle: CSSProperties = { margin: 0, fontSize: "1.875rem", fontWeight: 700, color: "var(--color-text)" };
const descriptionStyle: CSSProperties = { margin: "var(--space-1) 0 0", color: "var(--color-text-secondary)" };
const sectionStyle: CSSProperties = { display: "flex", flexDirection: "column", gap: "var(--space-3)", padding: "var(--space-4)", background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)" };
const sectionHeaderStyle: CSSProperties = { display: "flex", justifyContent: "space-between", alignItems: "center", gap: "var(--space-3)" };
const sectionTitleStyle: CSSProperties = { margin: 0, fontSize: "1.125rem" };
const rowStyle: CSSProperties = { display: "flex", justifyContent: "space-between", alignItems: "center", gap: "var(--space-3)", paddingBottom: "var(--space-2)", borderBottom: "1px solid var(--color-border)" };
const smallStyle: CSSProperties = { margin: 0, fontSize: "0.8125rem", color: "var(--color-text-secondary)" };
const linkStyle: CSSProperties = { color: "var(--color-primary-600)", textDecoration: "none", fontSize: "0.875rem", fontWeight: 600 };
const reportLinkStyle: CSSProperties = { ...linkStyle, display: "inline-flex", alignItems: "center", minHeight: 44, padding: "var(--space-2) var(--space-4)", border: "1px solid var(--color-primary-600)", borderRadius: "var(--radius-md)" };
const gridStyle: CSSProperties = { display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))", gap: "var(--space-3)" };
const statCardStyle: CSSProperties = { display: "flex", flexDirection: "column", gap: "var(--space-1)", padding: "var(--space-4)", background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)" };