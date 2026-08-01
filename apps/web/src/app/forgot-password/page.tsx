"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { TextInput } from "@/components/ui/TextInput";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");

    try {
      await apiFetch("auth/forgot-password", {
        method: "POST",
        body: JSON.stringify({ email }),
      });
      setSubmitted(true);
    } catch (cause) {
      setError(
        cause instanceof ApiException
          ? cause.message
          : "Permintaan gagal diproses. Silakan coba lagi.",
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
        aria-labelledby="forgot-title"
        style={{
          background: "var(--color-surface)",
          borderRadius: "var(--radius-lg)",
          boxShadow: "0 4px 12px rgb(0 0 0 / 0.1)",
          maxWidth: 400,
          padding: "var(--space-6)",
          width: "100%",
        }}
      >
        <h1
          id="forgot-title"
          style={{
            fontSize: "1.5rem",
            marginBottom: "var(--space-2)",
            textAlign: "center",
          }}
        >
          Lupa Kata Sandi
        </h1>
        <p
          style={{
            color: "var(--color-text-secondary)",
            marginBottom: "var(--space-6)",
            textAlign: "center",
          }}
        >
          Tautan reset akan dikirimkan bila email terdaftar.
        </p>

        {submitted ? (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "var(--space-4)",
            }}
          >
            <p
              role="status"
              style={{
                background: "var(--color-primary-600)",
                borderRadius: "var(--radius-md)",
                color: "#ffffff",
                padding: "var(--space-3)",
              }}
            >
              Bila akun terdaftar, petunjuk reset telah dikirim ke email Anda.
            </p>
            <Link
              href="/login"
              style={{
                color: "var(--color-primary-600)",
                textAlign: "center",
              }}
            >
              Kembali ke login
            </Link>
          </div>
        ) : (
          <form
            onSubmit={handleSubmit}
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "var(--space-4)",
            }}
          >
            {error ? (
              <p
                role="alert"
                style={{
                  background: "var(--color-danger)",
                  borderRadius: "var(--radius-md)",
                  color: "#ffffff",
                  padding: "var(--space-3)",
                }}
              >
                {error}
              </p>
            ) : null}

            <FormField label="Email terdaftar" required>
              {(props) => (
                <TextInput
                  {...props}
                  autoComplete="email"
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="nama@email.id"
                  required
                  type="email"
                  value={email}
                />
              )}
            </FormField>

            <Button loading={loading} type="submit">
              Kirim tautan reset
            </Button>
            <Link
              href="/login"
              style={{
                color: "var(--color-primary-600)",
                textAlign: "center",
              }}
            >
              Kembali ke login
            </Link>
          </form>
        )}
      </section>
    </main>
  );
}