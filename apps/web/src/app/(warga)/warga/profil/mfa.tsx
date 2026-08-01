"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { TextInput } from "@/components/ui/TextInput";

export function MFASettings() {
  const { getAccessToken } = useAuth();
  const [enrollment, setEnrollment] = useState<{ uri: string; secret: string } | null>(null);
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [enabled, setEnabled] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  const authorizedRequest = useCallback(async <T,>(path: string, options: RequestInit = {}): Promise<T> => {
    const token = await getAccessToken();
    if (!token) {
      throw new Error("Sesi berakhir. Silakan masuk kembali.");
    }
    return apiFetch<T>(path, {
      ...options,
      headers: { ...options.headers, Authorization: `Bearer ${token}` },
    });
  }, [getAccessToken]);

  useEffect(() => {
    void authorizedRequest<{ user: { mfa_active: boolean } }>("me")
      .then((data) => setEnabled(data.user.mfa_active))
      .catch(() => undefined);
  }, [authorizedRequest]);

  async function handleGenerate() {
    setLoading(true);
    setError("");
    setMessage("");

    try {
      const data = await authorizedRequest<{ uri: string; secret: string }>("me/mfa/generate", {
        method: "POST",
      });
      setEnrollment(data);
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal membuat setup MFA.");
    } finally {
      setLoading(false);
    }
  }

  async function handleEnable(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");

    try {
      await authorizedRequest("me/mfa/enable", {
        method: "POST",
        body: JSON.stringify({ code }),
      });
      setEnrollment(null);
      setCode("");
      setEnabled(true);
      setMessage("MFA aktif untuk akun ini.");
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Kode MFA tidak valid.");
    } finally {
      setLoading(false);
    }
  }

  async function handleDisable() {
    if (!code) {
      setError("Masukkan kode MFA aktif untuk menonaktifkan.");
      return;
    }

    setLoading(true);
    setError("");
    setMessage("");

    try {
      await authorizedRequest("me/mfa/disable", {
        method: "POST",
        body: JSON.stringify({ code }),
      });
      setCode("");
      setEnabled(false);
      setMessage("MFA dinonaktifkan.");
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal menonaktifkan MFA.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <section
      aria-labelledby="mfa-heading"
      style={{
        background: "var(--color-surface)",
        border: "1px solid var(--color-border)",
        borderRadius: "var(--radius-lg)",
        padding: "var(--space-4)",
      }}
    >
      <h2 id="mfa-heading" style={{ fontSize: "1.125rem", marginBottom: "var(--space-2)" }}>
        Autentikasi Dua Langkah
      </h2>
      <p style={{ color: "var(--color-text-secondary)", marginBottom: "var(--space-4)" }}>
        Gunakan aplikasi authenticator untuk meningkatkan keamanan akun.
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
      {message ? (
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
          {message}
        </p>
      ) : null}

      {enrollment ? (
        <form onSubmit={handleEnable} style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
          <p style={{ color: "var(--color-text-secondary)" }}>
            Tambahkan akun manual pada aplikasi authenticator memakai setup key berikut. Rahasia ini hanya ditampilkan sekali.
          </p>
          <output
            aria-label="Kunci setup MFA"
            style={{
              background: "var(--color-surface-muted)",
              borderRadius: "var(--radius-md)",
              fontFamily: "monospace",
              letterSpacing: "0.08em",
              overflowWrap: "anywhere",
              padding: "var(--space-3)",
            }}
          >
            {enrollment.secret}
          </output>
          <FormField label="Kode enam digit dari authenticator" required>
            {(props) => (
              <TextInput
                {...props}
                autoComplete="one-time-code"
                inputMode="numeric"
                maxLength={6}
                onChange={(event) => setCode(event.target.value.replace(/\D/g, ""))}
                pattern="[0-9]{6}"
                required
                value={code}
              />
            )}
          </FormField>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-2)" }}>
            <Button loading={loading} type="submit">
              Verifikasi dan aktifkan
            </Button>
            <Button disabled={loading} onClick={() => setEnrollment(null)} type="button" variant="ghost">
              Batal
            </Button>
          </div>
        </form>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
          <Button loading={loading} onClick={handleGenerate} style={{ alignSelf: "flex-start" }} variant="outline">
            {enabled ? "Buat setup MFA baru" : "Aktifkan MFA"}
          </Button>
          {enabled ? (
            <>
              <FormField label="Kode MFA aktif untuk menonaktifkan">
                {(props) => (
                  <TextInput
                    {...props}
                    autoComplete="one-time-code"
                    inputMode="numeric"
                    maxLength={6}
                    onChange={(event) => setCode(event.target.value.replace(/\D/g, ""))}
                    pattern="[0-9]{6}"
                    value={code}
                  />
                )}
              </FormField>
              <Button loading={loading} onClick={handleDisable} style={{ alignSelf: "flex-start" }} variant="danger">
                Nonaktifkan MFA
              </Button>
            </>
          ) : null}
        </div>
      )}
    </section>
  );
}