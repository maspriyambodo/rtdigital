"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { StatusBadge } from "@/components/ui/StatusBadge";

interface User {
  id: string;
  email: string | null;
  phone: string | null;
  status: "invited" | "active" | "inactive" | "locked";
  roles: string[];
  last_login_at: string | null;
  created_at: string;
}

const status: Record<User["status"], { label: string; variant: "success" | "warning" | "danger" | "neutral" }> = {
  active: { label: "Aktif", variant: "success" },
  invited: { label: "Diundang", variant: "warning" },
  inactive: { label: "Nonaktif", variant: "neutral" },
  locked: { label: "Terkunci", variant: "danger" },
};

function formatRoles(roles: string[]) {
  return roles.length ? roles.map((role) => role.replaceAll("_", " ")).join(", ") : "Tanpa peran";
}

export default function PenggunaPage() {
  const { getAccessToken } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadUsers = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const accessToken = await getAccessToken();
      if (!accessToken) {
        setError("Sesi telah berakhir. Silakan masuk kembali.");
        return;
      }
      setUsers(await apiFetch<User[]>("users", { headers: { Authorization: `Bearer ${accessToken}` } }));
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat pengguna.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // Async API synchronization; state updates occur after the request resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadUsers();
  }, [loadUsers]);

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header style={{ display: "flex", flexWrap: "wrap", justifyContent: "space-between", alignItems: "flex-start", gap: "var(--space-4)" }}>
        <div>
          <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Manajemen Pengguna</h1>
          <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
            Undangan, perubahan status, dan peran akun organisasi.
          </p>
        </div>
        <Link
          href="/pengurus/pengguna/invite"
          style={{
            alignItems: "center",
            background: "var(--color-primary-600)",
            border: "1px solid var(--color-primary-600)",
            borderRadius: "var(--radius-md)",
            color: "#ffffff",
            display: "inline-flex",
            fontWeight: 600,
            justifyContent: "center",
            minHeight: 44,
            padding: "var(--space-2) var(--space-4)",
            textDecoration: "none",
          }}
        >
          Undang pengguna
        </Link>
      </header>

      {loading ? (
        <p style={{ color: "var(--color-text-secondary)" }}>Memuat pengguna…</p>
      ) : error ? (
        <EmptyState title="Gagal memuat pengguna" description={error} action={<Button variant="secondary" onClick={() => void loadUsers()}>Coba lagi</Button>} />
      ) : users.length === 0 ? (
        <EmptyState title="Belum ada pengguna" description="Undang pengguna untuk memulai." />
      ) : (
        <div style={{ display: "grid", gap: "var(--space-3)" }}>
          {users.map((user) => (
            <article key={user.id} style={{ display: "grid", gap: "var(--space-2)", padding: "var(--space-4)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", background: "var(--color-surface)" }}>
              <div style={{ display: "flex", gap: "var(--space-3)", justifyContent: "space-between", alignItems: "flex-start" }}>
                <strong>{user.email ?? user.phone ?? "Kontak tidak tersedia"}</strong>
                <StatusBadge variant={status[user.status].variant}>{status[user.status].label}</StatusBadge>
              </div>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem", textTransform: "capitalize" }}>
                {formatRoles(user.roles)}
              </span>
              <div style={{ marginTop: "var(--space-2)" }}>
                <Link
                  href={`/pengurus/pengguna/${user.id}`}
                  style={{ color: "var(--color-primary-600)", fontSize: "0.875rem", fontWeight: 600, textDecoration: "none" }}
                >
                  Kelola pengguna →
                </Link>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}