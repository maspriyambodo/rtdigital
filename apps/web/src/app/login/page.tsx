"use client";

import Image from "next/image";
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
        const result = await apiFetch<{
          access_token: string;
          expires_at: string;
        }>("auth/mfa/verify", {
          method: "POST",
          headers: { Authorization: `Bearer ${mfaToken}` },
          body: JSON.stringify({ code: mfaCode }),
        });
        login(result.access_token, result.expires_at);
        await redirectByRole(result.access_token);
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
      await redirectByRole(result.access_token);
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

  async function redirectByRole(accessToken: string) {
    try {
      const principal = await apiFetch<{ roles: string[] }>("me", {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      const isPengurus = principal.roles.some((role) =>
        [
          "super_admin",
          "ketua_rt",
          "sekretaris",
          "bendahara",
          "pengurus",
        ].includes(role),
      );
      router.replace(isPengurus ? "/pengurus" : "/warga");
    } catch {
      router.replace("/warga");
    }
  }

  return (
    <main
      style={{
        alignItems: "center",
        background: "var(--color-bg)",
        display: "flex",
        justifyContent: "center",
        minHeight: "100vh",
        padding: "var(--space-6) var(--space-4)",
      }}
    >
      <section
        aria-labelledby="login-title"
        style={{
          background: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-xl)",
          boxShadow: "var(--shadow-lg)",
          maxWidth: 420,
          padding: "var(--space-8) var(--space-6)",
          width: "100%",
        }}
      >
        <div
          style={{
            alignItems: "center",
            display: "flex",
            flexDirection: "column",
            gap: "var(--space-3)",
            marginBottom: "var(--space-8)",
          }}
        >
          <Image
            src="/logo.png"
            alt=""
            width={56}
            height={56}
            priority
            style={{
              flexShrink: 0,
              width: "auto",
              height: "3.5rem",
              maxWidth: "100%",
              objectFit: "contain",
            }}
          />
          <h1
            id="login-title"
            style={{
              fontSize: "1.375rem",
              fontWeight: 700,
              letterSpacing: "-0.02em",
              textAlign: "center",
            }}
          >
            {mfaToken ? "Verifikasi MFA" : "Masuk RT Digital"}
          </h1>
          <p
            style={{
              color: "var(--color-text-secondary)",
              fontSize: "0.875rem",
              textAlign: "center",
            }}
          >
            {mfaToken
              ? "Masukkan kode dari aplikasi authenticator."
              : "Akses layanan warga dan pengurus RT."}
          </p>
        </div>

        {error ? (
          <div
            role="alert"
            style={{
              background: "var(--color-danger-bg)",
              border: "1px solid rgb(185 28 28 / 0.15)",
              borderRadius: "var(--radius-md)",
              color: "var(--color-danger)",
              fontSize: "0.875rem",
              fontWeight: 500,
              marginBottom: "var(--space-5)",
              padding: "var(--space-3) var(--space-4)",
            }}
          >
            {error}
          </div>
        ) : null}

        <form
          onSubmit={handleSubmit}
          style={{
            display: "flex",
            flexDirection: "column",
            gap: "var(--space-5)",
          }}
        >
          {mfaToken ? (
            <FormField
              label="Kode autentikasi"
              hint="Enam digit, berlaku singkat."
              required
            >
              {(props) => (
                <TextInput
                  {...props}
                  autoComplete="one-time-code"
                  autoFocus
                  inputMode="numeric"
                  maxLength={6}
                  onChange={(event) =>
                    setMfaCode(event.target.value.replace(/\D/g, ""))
                  }
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
                style={{
                  alignSelf: "flex-end",
                  color: "var(--color-primary-600)",
                  fontSize: "0.8125rem",
                  fontWeight: 600,
                }}
              >
                Lupa kata sandi?
              </Link>
            </>
          )}

          <Button
            loading={loading}
            type="submit"
            style={{ width: "100%", marginTop: "var(--space-1)" }}
          >
            {mfaToken ? "Verifikasi & masuk" : "Masuk"}
          </Button>
        </form>
      </section>
    </main>
  );
}