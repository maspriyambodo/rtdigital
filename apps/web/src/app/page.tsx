import Link from "next/link";

export default function HomePage() {
  return (
    <main
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        padding: "var(--space-4)",
      }}
    >
      <section
        aria-labelledby="home-title"
        style={{
          width: "100%",
          maxWidth: 480,
          padding: "var(--space-6)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-lg)",
          background: "var(--color-surface)",
          boxShadow: "var(--shadow-md)",
        }}
      >
        <p
          style={{
            marginBottom: "var(--space-2)",
            color: "var(--color-primary-700)",
            fontWeight: 700,
          }}
        >
          RT Digital
        </p>
        <h1
          id="home-title"
          style={{
            marginBottom: "var(--space-2)",
            fontSize: "1.5rem",
            lineHeight: 1.2,
          }}
        >
          Fondasi aplikasi siap
        </h1>
        <p
          style={{
            marginBottom: "var(--space-6)",
            color: "var(--color-text-secondary)",
          }}
        >
          Pilih shell sesuai peran untuk melihat struktur antarmuka.
        </p>
        <div
          style={{
            display: "grid",
            gap: "var(--space-3)",
          }}
        >
          <Link
            href="/warga"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              minHeight: 44,
              padding: "var(--space-2) var(--space-4)",
              borderRadius: "var(--radius-md)",
              background: "var(--color-primary-600)",
              color: "#ffffff",
              fontWeight: 600,
            }}
          >
            Shell warga
          </Link>
          <Link
            href="/pengurus"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              minHeight: 44,
              padding: "var(--space-2) var(--space-4)",
              border: "1px solid var(--color-border-strong)",
              borderRadius: "var(--radius-md)",
              color: "var(--color-text)",
              fontWeight: 600,
            }}
          >
            Shell pengurus
          </Link>
        </div>
      </section>
    </main>
  );
}