"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { TextInput } from "@/components/ui/TextInput";

type LoginResponse = {
  access_token: string;
  expires_at: string;
  mfa_required: boolean;
};

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [mfaCode, setMfaCode] = useState("");
  const [mfaToken, setMfaToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");

    try {
      if (mfaToken) {
        const result = await apiFetch<{ access_token: string; expires_at: string }>("auth/mfa/verify", {
          method: "POST",
          headers: { Authorization: `Bearer ${mfaToken}` },
          body: JSON.stringify({ code: mfaCode }),
        });
        login(result.access_token, result.expires_at);
        router.replace("/warga");
        return;
      }

      const result = await apiFetch<LoginResponse>("auth/login", {
        method: "POST",
        body: JSON.stringify({ identifier, password }),
      });

      if (result.mfa_required) {
        setMfaToken(result.access_token);
        setMfaCode("");
        return;
      }

      login(result.access_token, result.expires_at);
      router.replace("/warga");
    } catch (cause) {
      setError(
        cause instanceof ApiException
          ? cause.message
          : "Terjadi kesalahan. Silakan coba lagi.",
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
        aria-labelledby="login-title"
        style={{
          background: "var(--color-surface)",
          borderRadius: "var(--radius-lg)",
          boxShadow: "0 4px 12px rgb(0 0 0 / 0.1)",
          maxWidth: 400,
          padding: "var(--space-6)",
          width: "100%",
        }}
      >
        <h1 id="login-title" style={{ fontSize: "1.5rem", marginBottom: "var(--space-2)", textAlign: "center" }}>
          {mfaToken ? "Verifikasi MFA" : "Masuk RT Digital"}
        </h1>
        <p style={{ color: "var(--color-text-secondary)", marginBottom: "var(--space-6)", textAlign: "center" }}>
          {mfaToken
            ? "Masukkan kode dari aplikasi authenticator."
            : "Akses layanan warga dan pengurus RT."}
        </p>

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
          {mfaToken ? (
            <FormField label="Kode autentikasi" hint="Enam digit, berlaku singkat." required>
              {(props) => (
                <TextInput
                  {...props}
                  autoComplete="one-time-code"
                  autoFocus
                  inputMode="numeric"
                  maxLength={6}
                  onChange={(event) => setMfaCode(event.target.value.replace(/\D/g, ""))}
                  pattern="[0-9]{6}"
                  required
                  value={mfaCode}
                />
              )}
            </FormField>
          ) : (
            <>
              <FormField label="Email atau nomor telepon" required>
                {(props) => (
                  <TextInput
                    {...props}
                    autoComplete="username"
                    onChange={(event) => setIdentifier(event.target.value)}
                    placeholder="nama@email.id atau 081234567890"
                    required
                    value={identifier}
                  />
                )}
              </FormField>

              <FormField label="Kata sandi" required>
                {(props) => (
                  <TextInput
                    {...props}
                    autoComplete="current-password"
                    onChange={(event) => setPassword(event.target.value)}
                    required
                    type="password"
                    value={password}
                  />
                )}
              </FormField>

              <Link
                href="/forgot-password"
                style={{ alignSelf: "flex-end", color: "var(--color-primary-600)", fontSize: "0.875rem" }}
              >
                Lupa kata sandi?
              </Link>
            </>
          )}

          <Button loading={loading} type="submit">
            {mfaToken ? "Verifikasi & masuk" : "Masuk"}
          </Button>
        </form>
      </section>
    </main>
  );
}