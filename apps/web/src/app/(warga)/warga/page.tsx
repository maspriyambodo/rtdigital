import { Button } from "@/components/ui/Button";
import { DatePicker } from "@/components/ui/DatePicker";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";

export default function WargaHomePage() {
  return (
    <div style={{ display: "grid", gap: "var(--space-8)" }}>
      <section>
        <h1
          style={{
            marginBottom: "var(--space-2)",
            fontSize: "1.5rem",
            lineHeight: 1.2,
          }}
        >
          Beranda Warga
        </h1>
        <p style={{ color: "var(--color-text-secondary)" }}>
          Shell warga siap digunakan.
        </p>
      </section>

      <EmptyState
        title="Belum ada informasi"
        description="Pengumuman, agenda, dan ringkasan layanan akan tampil di sini."
        action={<Button variant="outline">Muat ulang</Button>}
      />

      <section style={{ display: "grid", gap: "var(--space-4)" }}>
        <h2 style={{ fontSize: "1.25rem", lineHeight: 1.25 }}>
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