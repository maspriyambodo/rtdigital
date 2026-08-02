import { Button } from "@/components/ui/Button";
import { DatePicker } from "@/components/ui/DatePicker";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";

export default function WargaHomePage() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-8)" }}>
      <header>
        <h1
          style={{
            margin: 0,
            marginBottom: "var(--space-1)",
            fontSize: "1.875rem",
            fontWeight: 700,
            lineHeight: 1.2,
            letterSpacing: "-0.025em",
            color: "var(--color-text)",
          }}
        >
          Beranda Warga
        </h1>
        <p style={{ margin: 0, color: "var(--color-text-secondary)", fontSize: "1rem" }}>
          Shell warga siap digunakan.
        </p>
      </header>

      <EmptyState
        title="Belum ada informasi"
        description="Pengumuman, agenda, dan ringkasan layanan akan tampil di sini."
        action={<Button variant="outline">Muat ulang</Button>}
      />

      <section
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "var(--space-4)",
          padding: "var(--space-5)",
          background: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-lg)",
          boxShadow: "var(--shadow-sm)",
        }}
      >
        <h2
          style={{
            margin: 0,
            fontSize: "1.125rem",
            fontWeight: 600,
            lineHeight: 1.25,
            color: "var(--color-text)",
          }}
        >
          Komponen UI dasar
        </h2>

        <FormField label="Nama lengkap" required>
          {(props) => <TextInput {...props} defaultValue="Samsul" />}
        </FormField>

        <FormField label="Status" error="Pilihan tidak valid.">
          {(props) => (
            <Select {...props}>
              <option>Aktif</option>
              <option>Nonaktif</option>
            </Select>
          )}
        </FormField>

        <FormField label="Tanggal lahir" hint="Gunakan pemilih tanggal perangkat.">
          {(props) => <DatePicker {...props} />}
        </FormField>

        <div
          aria-label="Contoh status"
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--space-2)",
          }}
        >
          <StatusBadge variant="success">Lunas</StatusBadge>
          <StatusBadge variant="warning">Menunggu</StatusBadge>
          <StatusBadge variant="danger">Ditolak</StatusBadge>
          <StatusBadge variant="info">Baru</StatusBadge>
          <StatusBadge>Draft</StatusBadge>
        </div>
      </section>
    </div>
  );
}