"use client";

import { useState, type FormEvent } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { TextInput } from "@/components/ui/TextInput";
import { MFASettings } from "./mfa";

export default function ProfilPage() {
  const { getAccessToken, logout } = useAuth();
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [loading, setLoading] = useState(false);
  const [logoutLoading, setLogoutLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  async function handlePasswordChange(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (newPassword !== confirmation) {
      setError("Konfirmasi kata sandi tidak cocok.");
      return;
    }
    if (newPassword.length < 8) {
      setError("Kata sandi baru minimal 8 karakter.");
      return;
    }

    setLoading(true);
    setError("");
    setSuccess(false);

    try {
      const token = await getAccessToken();
      if (!token) {
        await logout();
        return;
      }

      await apiFetch("me/password", {
        method: "PATCH",
        headers: { Authorization: `Bearer ${token}` },
        body: JSON.stringify({
          old_password: oldPassword,
          new_password: newPassword,
        }),
      });

      setSuccess(true);
      setOldPassword("");
      setNewPassword("");
      setConfirmation("");
    } catch (cause) {
      setError(
        cause instanceof ApiException
          ? cause.message
          : "Gagal mengubah kata sandi. Periksa kata sandi lama Anda.",
      );
    } finally {
      setLoading(false);
    }
  }

  async function handleLogout() {
    setLogoutLoading(true);
    await logout();
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-6)" }}>
      <div>
        <h1 style={{ fontSize: "1.5rem", fontWeight: 700 }}>Profil & Akun</h1>
        <p style={{ color: "var(--color-text-secondary)" }}>
          Pengaturan kredensial keamanan akun Anda.
        </p>
      </div>

      <section
        aria-labelledby="password-heading"
        style={{
          background: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-lg)",
          padding: "var(--space-4)",
        }}
      >
        <h2 id="password-heading" style={{ fontSize: "1.125rem", marginBottom: "var(--space-4)" }}>
          Ubah Kata Sandi
        </h2>

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
            Kata sandi berhasil diperbarui.
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

        <form onSubmit={handlePasswordChange} style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
          <FormField label="Kata sandi saat ini" required>
            {(props) => (
              <TextInput
                {...props}
                autoComplete="current-password"
                onChange={(event) => setOldPassword(event.target.value)}
                required
                type="password"
                value={oldPassword}
              />
            )}
          </FormField>

          <FormField label="Kata sandi baru" hint="Minimal 8 karakter." required>
            {(props) => (
              <TextInput
                {...props}
                autoComplete="new-password"
                onChange={(event) => setNewPassword(event.target.value)}
                required
                type="password"
                value={newPassword}
              />
            )}
          </FormField>

          <FormField label="Ulangi kata sandi baru" required>
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

          <Button loading={loading} style={{ alignSelf: "flex-start" }} type="submit">
            Simpan kata sandi baru
          </Button>
        </form>
      </section>

      <MFASettings />

      <section
        aria-labelledby="logout-heading"
        style={{
          background: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-lg)",
          padding: "var(--space-4)",
        }}
      >
        <h2 id="logout-heading" style={{ fontSize: "1.125rem", marginBottom: "var(--space-2)" }}>
          Keluar
        </h2>
        <p style={{ color: "var(--color-text-secondary)", marginBottom: "var(--space-4)" }}>
          Akhiri sesi aman pada perangkat ini.
        </p>
        <Button loading={logoutLoading} onClick={handleLogout} variant="danger">
          Keluar dari aplikasi
        </Button>
      </section>
    </div>
  );
}