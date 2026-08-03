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
      setError(
        reason instanceof Error
          ? reason.message
          : "Gagal memuat dashboard pengurus.",
      );
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
          <p style={descriptionStyle}>
            Ringkasan operasional dan keuangan lingkungan RT.
          </p>
        </div>
        <Link href="/pengurus/laporan" style={reportLinkStyle}>
          Laporan & Ekspor
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
            <StatCard
              title="Total Tunggakan"
              value={`Rp ${data.outstanding_amount.toLocaleString("id-ID")}`}
            />
            <StatCard
              title="Menunggu Verifikasi"
              value={data.pending_payments}
            />
            <StatCard
              title="Saldo Kas"
              value={`Rp ${data.cash_balance.toLocaleString("id-ID")}`}
            />
            <StatCard title="Antrean Surat" value={data.pending_letters} />
            <StatCard title="Aduan Terbuka" value={data.open_complaints} />
          </div>

          <div style={feedGridStyle}>
            <section style={sectionStyle}>
              <div style={sectionHeaderStyle}>
                <div>
                  <h2 style={sectionTitleStyle}>Pembayaran Terbaru</h2>
                  <p style={sectionSubtitleStyle}>
                    Daftar transaksi setoran iuran warga
                  </p>
                </div>
                <Link href="/pengurus/tagihan" style={linkStyle}>
                  Urus Pembayaran
                </Link>
              </div>
              {data.recent_payments.length ? (
                <div style={listStyle}>
                  {data.recent_payments.map((payment) => (
                    <div key={payment.id} style={rowStyle}>
                      <div style={rowContentStyle}>
                        <strong style={strongStyle}>
                          {payment.payment_number}
                        </strong>
                        <p style={smallStyle}>
                          Tagihan: {payment.invoice_number} ·{" "}
                          {formatDate(payment.paid_at)}
                        </p>
                      </div>
                      <div style={paymentMetaStyle}>
                        <strong style={amountStyle}>
                          Rp {payment.amount.toLocaleString("id-ID")}
                        </strong>
                        <StatusBadge
                          variant={
                            payment.verification_status === "verified"
                              ? "success"
                              : payment.verification_status === "pending"
                                ? "warning"
                                : "danger"
                          }
                        >
                          {payment.verification_status}
                        </StatusBadge>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p style={smallStyle}>Belum ada riwayat pembayaran.</p>
              )}
            </section>

            <section style={sectionStyle}>
              <div style={sectionHeaderStyle}>
                <div>
                  <h2 style={sectionTitleStyle}>Antrean Surat</h2>
                  <p style={sectionSubtitleStyle}>
                    Permohonan surat pengantar dari warga
                  </p>
                </div>
                <Link href="/pengurus/surat" style={linkStyle}>
                  Proses Surat
                </Link>
              </div>
              {data.recent_letters.length ? (
                <div style={listStyle}>
                  {data.recent_letters.map((letter) => (
                    <div key={letter.id} style={rowStyle}>
                      <div style={rowContentStyle}>
                        <strong style={strongStyle}>
                          {letter.request_number}
                        </strong>
                        <p style={smallStyle}>
                          {letter.letter_type} · {formatDate(letter.updated_at)}
                        </p>
                      </div>
                      <StatusBadge
                        variant={
                          letter.status === "issued"
                            ? "success"
                            : letter.status === "rejected"
                              ? "danger"
                              : "warning"
                        }
                      >
                        {letter.status}
                      </StatusBadge>
                    </div>
                  ))}
                </div>
              ) : (
                <p style={smallStyle}>Antrean surat kosong.</p>
              )}
            </section>

            <section style={sectionStyle}>
              <div style={sectionHeaderStyle}>
                <div>
                  <h2 style={sectionTitleStyle}>Aduan Terbaru</h2>
                  <p style={sectionSubtitleStyle}>
                    Laporan aspirasi dan masalah lingkungan
                  </p>
                </div>
                <Link href="/pengurus/aduan" style={linkStyle}>
                  Tinjau Aduan
                </Link>
              </div>
              {data.recent_complaints.length ? (
                <div style={listStyle}>
                  {data.recent_complaints.map((complaint) => (
                    <div key={complaint.id} style={rowStyle}>
                      <div style={rowContentStyle}>
                        <strong style={strongStyle}>
                          [{complaint.ticket_number}] {complaint.title}
                        </strong>
                        <p style={smallStyle}>
                          Prioritas: {complaint.priority} ·{" "}
                          {formatDate(complaint.updated_at)}
                        </p>
                      </div>
                      <StatusBadge
                        variant={
                          complaint.status === "resolved" ||
                          complaint.status === "closed"
                            ? "success"
                            : complaint.status === "rejected"
                              ? "danger"
                              : "warning"
                        }
                      >
                        {complaint.status}
                      </StatusBadge>
                    </div>
                  ))}
                </div>
              ) : (
                <p style={smallStyle}>Antrean aduan kosong.</p>
              )}
            </section>
          </div>
        </>
      ) : null}
    </div>
  );
}

function StatCard({ title, value }: { title: string; value: string | number }) {
  return (
    <div style={statCardStyle}>
      <span style={statTitleStyle}>{title}</span>
      <strong style={statValueStyle}>{value}</strong>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

const pageStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-6)",
};
const headerStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "flex-start",
  gap: "var(--space-4)",
  flexWrap: "wrap",
};
const titleStyle: CSSProperties = {
  margin: 0,
  fontSize: "clamp(1.5rem, 4vw, 1.875rem)",
  fontWeight: 700,
  letterSpacing: "-0.025em",
  color: "var(--color-text)",
};
const descriptionStyle: CSSProperties = {
  margin: "var(--space-1) 0 0",
  color: "var(--color-text-secondary)",
  fontSize: "0.9375rem",
};
const gridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))",
  gap: "var(--space-3)",
};
const statCardStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-2)",
  minWidth: 0,
  padding: "var(--space-4) var(--space-5)",
  background: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-lg)",
  boxShadow: "var(--shadow-sm)",
};
const statTitleStyle: CSSProperties = {
  color: "var(--color-text-muted)",
  fontSize: "0.75rem",
  fontWeight: 600,
  letterSpacing: "0.05em",
  textTransform: "uppercase",
};
const statValueStyle: CSSProperties = {
  overflowWrap: "anywhere",
  color: "var(--color-text)",
  fontSize: "1.375rem",
  fontWeight: 700,
  letterSpacing: "-0.02em",
};
const feedGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))",
  gap: "var(--space-4)",
};
const sectionStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-4)",
  minWidth: 0,
  padding: "var(--space-5)",
  background: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-lg)",
  boxShadow: "var(--shadow-sm)",
};
const sectionHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "flex-start",
  gap: "var(--space-3)",
  paddingBottom: "var(--space-3)",
  borderBottom: "1px solid var(--color-border)",
};
const sectionTitleStyle: CSSProperties = {
  margin: 0,
  color: "var(--color-text)",
  fontSize: "1.0625rem",
  fontWeight: 700,
  letterSpacing: "-0.01em",
};
const sectionSubtitleStyle: CSSProperties = {
  margin: "2px 0 0",
  color: "var(--color-text-secondary)",
  fontSize: "0.8125rem",
};
const listStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-3)",
};
const rowStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "var(--space-3)",
  minWidth: 0,
  paddingBottom: "var(--space-3)",
  borderBottom: "1px solid var(--color-border)",
};
const rowContentStyle: CSSProperties = {
  display: "flex",
  flex: 1,
  flexDirection: "column",
  gap: "2px",
  minWidth: 0,
};
const paymentMetaStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  flexShrink: 0,
  alignItems: "flex-end",
  gap: "var(--space-1)",
  textAlign: "right",
};
const strongStyle: CSSProperties = {
  overflow: "hidden",
  color: "var(--color-text)",
  fontSize: "0.875rem",
  fontWeight: 600,
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};
const amountStyle: CSSProperties = {
  color: "var(--color-text)",
  fontSize: "0.875rem",
  fontWeight: 600,
};
const smallStyle: CSSProperties = {
  margin: 0,
  overflow: "hidden",
  color: "var(--color-text-secondary)",
  fontSize: "0.8125rem",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};
const linkStyle: CSSProperties = {
  flexShrink: 0,
  color: "var(--color-primary-600)",
  fontSize: "0.8125rem",
  fontWeight: 600,
};
const reportLinkStyle: CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  minHeight: 40,
  padding: "var(--space-2) var(--space-4)",
  borderRadius: "var(--radius-md)",
  background: "var(--color-primary-600)",
  color: "#ffffff",
  fontSize: "0.875rem",
  fontWeight: 600,
  boxShadow: "0 1px 2px rgb(37 99 235 / 0.2)",
};