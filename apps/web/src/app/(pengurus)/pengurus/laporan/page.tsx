"use client";

import { useState, useTransition, type CSSProperties } from "react";

import { Button } from "@/components/ui/Button";
import { DatePicker } from "@/components/ui/DatePicker";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { TextInput } from "@/components/ui/TextInput";
import { useAuth } from "@/lib/auth-context";
import {
  downloadReportCSV,
  fetchReport,
  type ReportFilter,
  type ReportType,
} from "@/lib/reports";

const reportOptions: Array<{ label: string; value: ReportType }> = [
  { label: "Daftar Warga", value: "residents" },
  { label: "Mutasi Warga", value: "mutations" },
  { label: "Daftar Keluarga", value: "households" },
  { label: "Daftar Tagihan", value: "invoices" },
  { label: "Rekap Tunggakan", value: "arrears" },
  { label: "Daftar Pembayaran", value: "payments" },
  { label: "Buku Kas", value: "cash" },
  { label: "Pengajuan Surat", value: "letters" },
  { label: "Aduan Warga", value: "complaints" },
];

type ReportRow = Record<string, unknown>;

export default function LaporanPage() {
  const { getAccessToken } = useAuth();
  const [isPending, startTransition] = useTransition();
  const [reportType, setReportType] = useState<ReportType>("residents");
  const [filter, setFilter] = useState<ReportFilter>({
    start_date: "",
    end_date: "",
    status: "",
  });
  const [items, setItems] = useState<ReportRow[] | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const activeFilter = (): ReportFilter => ({
    start_date: filter.start_date || undefined,
    end_date: filter.end_date || undefined,
    status: filter.status || undefined,
  });

  const preview = () => {
    setMessage(null);
    startTransition(async () => {
      try {
        setItems(await fetchReport<ReportRow>(reportType, activeFilter(), getAccessToken));
      } catch (reason) {
        setItems(null);
        setMessage(reason instanceof Error ? reason.message : "Gagal memuat pratinjau laporan.");
      }
    });
  };

  const exportCSV = () => {
    setMessage(null);
    startTransition(async () => {
      try {
        await downloadReportCSV(
          reportType,
          activeFilter(),
          getAccessToken,
          `laporan_${reportType}`,
        );
        setMessage("Laporan berhasil diekspor. Aktivitas tercatat pada audit log.");
      } catch (reason) {
        setMessage(reason instanceof Error ? reason.message : "Gagal mengekspor laporan.");
      }
    });
  };

  const changeReportType = (value: ReportType) => {
    setReportType(value);
    setItems(null);
    setMessage(null);
    setFilter({ start_date: "", end_date: "", status: "" });
  };

  return (
    <div style={pageStyle}>
      <header>
        <h1 style={titleStyle}>Laporan & Ekspor</h1>
        <p style={descriptionStyle}>
          Pratinjau dan ekspor CSV data kependudukan, keuangan, surat, serta aduan.
        </p>
      </header>

      <section style={filterPanelStyle}>
        <FormField label="Jenis Laporan" required>
          {(props) => (
            <Select
              {...props}
              value={reportType}
              onChange={(event) => changeReportType(event.target.value as ReportType)}
            >
              {reportOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
          )}
        </FormField>

        <div style={filterGridStyle}>
          <FormField label="Tanggal Mulai">
            {(props) => (
              <DatePicker
                {...props}
                value={filter.start_date}
                onChange={(event) =>
                  setFilter((current) => ({ ...current, start_date: event.target.value }))
                }
              />
            )}
          </FormField>

          <FormField label="Tanggal Akhir">
            {(props) => (
              <DatePicker
                {...props}
                value={filter.end_date}
                onChange={(event) =>
                  setFilter((current) => ({ ...current, end_date: event.target.value }))
                }
              />
            )}
          </FormField>

          <StatusFilter
            reportType={reportType}
            value={filter.status || ""}
            onChange={(status) => setFilter((current) => ({ ...current, status }))}
          />
        </div>

        <div style={actionRowStyle}>
          <Button variant="secondary" onClick={preview} loading={isPending}>
            Pratinjau
          </Button>
          <Button onClick={exportCSV} loading={isPending}>
            Ekspor CSV
          </Button>
        </div>
      </section>

      {message ? <p style={messageStyle}>{message}</p> : null}

      {items?.length === 0 ? (
        <EmptyState
          title="Tidak ada data"
          description="Tidak ada data yang sesuai dengan filter dipilih."
        />
      ) : null}

      {items?.length ? <PreviewTable items={items} /> : null}
    </div>
  );
}

function StatusFilter({
  reportType,
  value,
  onChange,
}: {
  reportType: ReportType;
  value: string;
  onChange: (value: string) => void;
}) {
  if (reportType === "residents") {
    return (
      <FormField label="Status Warga">
        {(props) => (
          <Select {...props} value={value} onChange={(event) => onChange(event.target.value)}>
            <option value="">Semua status</option>
            <option value="active">Aktif</option>
            <option value="moved">Pindah</option>
            <option value="deceased">Meninggal</option>
            <option value="inactive">Nonaktif</option>
          </Select>
        )}
      </FormField>
    );
  }

  if (reportType === "mutations") {
    return (
      <FormField label="Jenis Mutasi">
        {(props) => (
          <Select {...props} value={value} onChange={(event) => onChange(event.target.value)}>
            <option value="">Semua mutasi</option>
            <option value="active">Anggota aktif</option>
            <option value="ended">Anggota keluar</option>
          </Select>
        )}
      </FormField>
    );
  }

  if (reportType === "arrears") {
    return null;
  }

  return (
    <FormField label="Status">
      {(props) => (
        <TextInput
          {...props}
          value={value}
          placeholder="Contoh: verified, unpaid, issued"
          onChange={(event) => onChange(event.target.value)}
        />
      )}
    </FormField>
  );
}

function PreviewTable({ items }: { items: ReportRow[] }) {
  const headers = Object.keys(items[0]);
  const visibleItems = items.slice(0, 10);

  return (
    <section style={tableCardStyle}>
      <h2 style={sectionTitleStyle}>Pratinjau ({items.length} baris)</h2>
      <div style={tableWrapperStyle}>
        <table style={tableStyle}>
          <thead>
            <tr>
              {headers.map((header) => (
                <th key={header} style={headerStyle}>
                  {header.replaceAll("_", " ")}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {visibleItems.map((item, rowIndex) => (
              <tr key={rowIndex}>
                {headers.map((header) => (
                  <td key={header} style={cellStyle}>
                    {formatValue(item[header])}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {items.length > visibleItems.length ? (
        <p style={smallStyle}>
          Menampilkan {visibleItems.length} dari {items.length} baris. Ekspor CSV untuk data lengkap.
        </p>
      ) : null}
    </section>
  );
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
}

const pageStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-5)",
};
const titleStyle: CSSProperties = {
  margin: 0,
  fontSize: "1.875rem",
  fontWeight: 700,
  color: "var(--color-text)",
};
const descriptionStyle: CSSProperties = {
  margin: "var(--space-1) 0 0",
  color: "var(--color-text-secondary)",
};
const filterPanelStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-4)",
  padding: "var(--space-4)",
  background: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-lg)",
};
const filterGridStyle: CSSProperties = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))",
  gap: "var(--space-3)",
};
const actionRowStyle: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--space-3)",
};
const messageStyle: CSSProperties = {
  margin: 0,
  color: "var(--color-primary-600)",
  fontWeight: 600,
};
const tableCardStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--space-3)",
  padding: "var(--space-4)",
  background: "var(--color-surface)",
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-lg)",
};
const sectionTitleStyle: CSSProperties = {
  margin: 0,
  fontSize: "1.125rem",
};
const tableWrapperStyle: CSSProperties = {
  overflowX: "auto",
};
const tableStyle: CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
  fontSize: "0.875rem",
};
const headerStyle: CSSProperties = {
  padding: "var(--space-2)",
  textAlign: "left",
  textTransform: "capitalize",
  borderBottom: "2px solid var(--color-border)",
  color: "var(--color-text-secondary)",
};
const cellStyle: CSSProperties = {
  padding: "var(--space-2)",
  borderBottom: "1px solid var(--color-border)",
  whiteSpace: "nowrap",
};
const smallStyle: CSSProperties = {
  margin: 0,
  fontSize: "0.8125rem",
  color: "var(--color-text-secondary)",
};