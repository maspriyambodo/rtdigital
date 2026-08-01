"use client";

import Link from "next/link";
import { Suspense, useState, type FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ApiException, apiFetch } from "@/lib/api";
import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { TextInput } from "@/components/ui/TextInput";

export default function ActivatePage() {
  return (
    <Suspense fallback={<p style={{ padding: "var(--space-4)", textAlign: "center" }}>Memuat…</p>}>
      <ActivateForm />
    </Suspense>
  );
}

function ActivateForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [token, setToken] = useState(searchParams.get("token") ?? "");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (password !== confirmation) {
      setError("Konfirmasi kata sandi tidak cocok.");
      return;
    }
    if (password.length < 8) {
      setError("Kata sandi minimal 8 karakter.");
      return;
    }

    setLoading(true);
    setError("");
    try {
      await apiFetch("auth/activate", {
        method: "POST",
        body: JSON.stringify({ token, password }),
      });
      setSuccess(true);
      window.setTimeout(() => router.replace("/login"), 2_000);
    } catch (cause) {
      setError(
        cause instanceof ApiException
          ? cause.message
          : "Aktivasi gagal. Periksa token undangan Anda.",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <main
      style={{
        alignItems: "center",
        background: "var(--color-surface-muted)",
        display: "flex",
        justifyContent: "center",
        minHeight: "100vh",
        padding: "var(--space-4)",
      }}
    >
      <section
        aria-labelledby="activate-title"
        style={{
          background: "var(--color-surface)",
          borderRadius: "var(--radius-lg)",
          boxShadow: "0 4px 12px rgb(0 0 0 / 0.1)",
          maxWidth: 400,
          padding: "var(--space-6)",
          width: "100%",
        }}
      >
        <h1 id="activate-title" style={{ fontSize: "1.5rem", marginBottom: "var(--space-2)", textAlign: "center" }}>
          Aktivasi Akun
        </h1>
        <p style={{ color: "var(--color-text-secondary)", marginBottom: "var(--space-6)", textAlign: "center" }}>
          Atur kata sandi untuk mengaktifkan akun RT Digital.
        </p>

        {success ? (
          <p
            role="status"
            style={{
              background: "var(--color-primary-600)",
              borderRadius: "var(--radius-md)",
              color: "#ffffff",
              marginBottom: "var(--space-4)",
              padding: "var(--space-3)",
            }}
          >
            Akun aktif. Mengalihkan ke halaman login…
          </p>
        ) : null}
        {error ? (
          <p
            role="alert"
            style={{
              background: "var(--color-danger)",
              borderRadius: "var(--radius-md)",
              color: "#ffffff",
              marginBottom: "var(--space-4)",
              padding: "var(--space-3)",
            }}
          >
            {error}
          </p>
        ) : null}

        <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
          <FormField label="Token aktivasi" required>
            {(props) => (
              <TextInput
                {...props}
                autoComplete="off"
                onChange={(event) => setToken(event.target.value)}
                required
                value={token}
              />
            )}
          </FormField>
          <FormField label="Kata sandi baru" hint="Minimal 8 karakter." required>
            {(props) => (
              <TextInput
                {...props}
                autoComplete="new-password"
                onChange={(event) => setPassword(event.target.value)}
                required
                type="password"
                value={password}
              />
            )}
          </FormField>
          <FormField label="Ulangi kata sandi" required>
            {(props) => (
              <TextInput
                {...props}
                autoComplete="new-password"
                onChange={(event) => setConfirmation(event.target.value)}
                required
                type="password"
                value={confirmation}
              />
            )}
          </FormField>
          <Button disabled={success} loading={loading} type="submit">
            Aktifkan akun
          </Button>
          <Link href="/login" style={{ color: "var(--color-primary-600)", textAlign: "center" }}>
            Kembali ke login
          </Link>
        </form>
      </section>
    </main>
  );
}