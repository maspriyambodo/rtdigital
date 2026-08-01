"use client";

export default function ErrorPage({
  reset,
}: Readonly<{
  error: Error & { digest?: string };
  reset: () => void;
}>) {
  return (
    <main
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        padding: "var(--space-4)",
        textAlign: "center",
      }}
    >
      <section
        aria-labelledby="error-title"
        style={{
          width: "100%",
          maxWidth: 400,
          padding: "var(--space-6)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-lg)",
          background: "var(--color-surface)",
          boxShadow: "var(--shadow-md)",
        }}
      >
        <p
          aria-hidden="true"
          style={{
            marginBottom: "var(--space-3)",
            color: "var(--color-danger)",
            fontSize: "1.5rem",
            fontWeight: 700,
          }}
        >
          !
        </p>
        <h1
          id="error-title"
          style={{
            marginBottom: "var(--space-2)",
            fontSize: "1.5rem",
            lineHeight: 1.2,
          }}
        >
          Terjadi kesalahan
        </h1>
        <p
          style={{
            marginBottom: "var(--space-6)",
            color: "var(--color-text-secondary)",
          }}
        >
          Halaman tidak dapat dimuat. Coba lagi. Hubungi pengurus RT bila
          masalah berlanjut.
        </p>
        <button
          type="button"
          onClick={reset}
          style={{
            width: "100%",
            minHeight: 44,
            border: 0,
            borderRadius: "var(--radius-md)",
            background: "var(--color-primary-600)",
            color: "#ffffff",
            cursor: "pointer",
            fontWeight: 600,
          }}
        >
          Coba lagi
        </button>
      </section>
    </main>
  );
}