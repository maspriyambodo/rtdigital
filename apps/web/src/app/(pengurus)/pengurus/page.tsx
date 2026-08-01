import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { TextInput } from "@/components/ui/TextInput";

export default function PengurusHomePage() {
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
          Dashboard Pengurus
        </h1>
        <p style={{ color: "var(--color-text-secondary)" }}>
          Shell administrasi siap digunakan.
        </p>
      </section>

      <EmptyState
        title="Ringkasan operasional kosong"
        description="Statistik warga, keuangan, surat, dan aduan akan tampil di sini."
        action={
          <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-3)" }}>
            <Button variant="secondary">Ekspor</Button>
            <Button>Buat pengumuman</Button>
          </div>
        }
      />

      <section style={{ display: "grid", gap: "var(--space-4)" }}>
        <h2 style={{ fontSize: "1.25rem", lineHeight: 1.25 }}>
          Komponen form
        </h2>

        <FormField label="Judul" required>
          {(props) => <TextInput {...props} placeholder="Rapat kerja bakti" />}
        </FormField>

        <FormField label="Kategori" error="Kategori wajib dipilih.">
          {(props) => (
            <Select {...props}>
              <option value="">Pilih kategori</option>
              <option value="iuran">Iuran</option>
              <option value="administrasi">Administrasi</option>
            </Select>
          )}
        </FormField>
      </section>
    </div>
  );
}