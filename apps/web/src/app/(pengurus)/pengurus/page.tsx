import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { TextInput } from "@/components/ui/TextInput";

export default function PengurusHomePage() {
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
          Dashboard Pengurus
        </h1>
        <p style={{ margin: 0, color: "var(--color-text-secondary)", fontSize: "1rem" }}>
          Shell administrasi siap digunakan.
        </p>
      </header>

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