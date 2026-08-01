"use client";

import { use, useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { StatusBadge } from "@/components/ui/StatusBadge";

interface Role {
  id: string;
  code: string;
  name: string;
}

interface UserDetail {
  id: string;
  email: string | null;
  phone: string | null;
  status: "invited" | "active" | "inactive" | "locked";
  roles: Role[];
  mfa_enabled_at: string | null;
  last_login_at: string | null;
  created_at: string;
}

const status: Record<UserDetail["status"], { label: string; variant: "success" | "warning" | "danger" | "neutral" }> = {
  active: { label: "Aktif", variant: "success" },
  invited: { label: "Diundang", variant: "warning" },
  inactive: { label: "Nonaktif", variant: "neutral" },
  locked: { label: "Terkunci", variant: "danger" },
};

export default function UserDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const { getAccessToken } = useAuth();
  const [user, setUser] = useState<UserDetail | null>(null);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [acting, setActing] = useState(false);
  const [error, setError] = useState("");

  const authorized = useCallback(async <T,>(path: string, options: RequestInit = {}) => {
    const token = await getAccessToken();
    if (!token) throw new Error("Sesi telah berakhir.");
    return apiFetch<T>(path, {
      ...options,
      headers: { ...options.headers, Authorization: `Bearer ${token}` },
    });
  }, [getAccessToken]);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [userData, roleData] = await Promise.all([
        authorized<UserDetail>(`users/${id}`),
        authorized<Role[]>("roles"),
      ]);
      setUser(userData);
      setRoles(roleData);
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal memuat detail pengguna.");
    } finally {
      setLoading(false);
    }
  }, [authorized, id]);

  useEffect(() => {
    // Async API synchronization; state updates occur after the request resolves.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadData();
  }, [loadData]);

  async function act(request: () => Promise<unknown>, fallback: string) {
    setActing(true);
    setError("");
    try {
      await request();
      await loadData();
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : fallback);
    } finally {
      setActing(false);
    }
  }

  async function updateStatus(nextStatus: "active" | "inactive") {
    if (!confirm(`Ubah status pengguna menjadi ${status[nextStatus].label}?`)) return;
    await act(
      () => authorized(`users/${id}`, { method: "PATCH", body: JSON.stringify({ status: nextStatus }) }),
      "Gagal memperbarui status pengguna.",
    );
  }

  async function assignRole(roleID: string) {
    if (!roleID) return;
    await act(
      () => authorized(`users/${id}/roles`, { method: "POST", body: JSON.stringify({ role_id: roleID }) }),
      "Gagal menetapkan peran.",
    );
  }

  async function revokeRole(role: Role) {
    if (!confirm(`Cabut peran ${role.name}?`)) return;
    await act(
      () => authorized(`users/${id}/roles/${role.id}`, { method: "DELETE" }),
      "Gagal mencabut peran.",
    );
  }

  if (loading && !user) return <p style={{ color: "var(--color-text-secondary)" }}>Memuat detail pengguna…</p>;
  if (!user) return <EmptyState title="Gagal memuat pengguna" description={error} action={<Button variant="secondary" onClick={() => router.back()}>Kembali</Button>} />;

  const unassignedRoles = roles.filter((role) => !user.roles.some((assigned) => assigned.id === role.id));

  return (
    <div style={{ display: "grid", gap: "var(--space-6)", maxWidth: 720 }}>
      <header style={{ display: "flex", alignItems: "flex-start", flexWrap: "wrap", gap: "var(--space-4)", justifyContent: "space-between" }}>
        <div>
          <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>{user.email ?? user.phone ?? "Detail Pengguna"}</h1>
          <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>Kelola status akun dan peran organisasi.</p>
        </div>
        <Button onClick={() => router.back()} variant="secondary">Kembali</Button>
      </header>

      {error ? <p role="alert" style={{ background: "var(--color-danger)", borderRadius: "var(--radius-md)", color: "#fff", padding: "var(--space-3)" }}>{error}</p> : null}

      <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", display: "grid", gap: "var(--space-4)", padding: "var(--space-4)" }}>
        <div style={{ alignItems: "center", display: "flex", justifyContent: "space-between" }}>
          <h2 style={{ fontSize: "1.125rem" }}>Informasi Akun</h2>
          <StatusBadge variant={status[user.status].variant}>{status[user.status].label}</StatusBadge>
        </div>
        <dl style={{ display: "grid", gap: "var(--space-3)", gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))" }}>
          <div><dt style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>Email</dt><dd>{user.email ?? "-"}</dd></div>
          <div><dt style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>Telepon</dt><dd>{user.phone ?? "-"}</dd></div>
          <div><dt style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>MFA</dt><dd>{user.mfa_enabled_at ? "Aktif" : "Tidak aktif"}</dd></div>
          <div><dt style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>Login terakhir</dt><dd>{user.last_login_at ? new Date(user.last_login_at).toLocaleString("id-ID") : "Belum pernah"}</dd></div>
        </dl>
        {user.status !== "invited" ? (
          <div>
            {user.status === "active" ? (
              <Button disabled={acting} onClick={() => void updateStatus("inactive")} variant="danger">Nonaktifkan akun</Button>
            ) : (
              <Button disabled={acting} onClick={() => void updateStatus("active")} variant="outline">Aktifkan akun</Button>
            )}
          </div>
        ) : null}
      </section>

      <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", display: "grid", gap: "var(--space-4)", padding: "var(--space-4)" }}>
        <h2 style={{ fontSize: "1.125rem" }}>Peran</h2>
        {user.roles.length ? (
          <ul style={{ display: "grid", gap: "var(--space-2)", listStyle: "none", margin: 0, padding: 0 }}>
            {user.roles.map((role) => (
              <li key={role.id} style={{ alignItems: "center", background: "var(--color-surface-muted)", borderRadius: "var(--radius-md)", display: "flex", gap: "var(--space-3)", justifyContent: "space-between", padding: "var(--space-3)" }}>
                <span><strong>{role.name}</strong><br /><code style={{ color: "var(--color-text-secondary)", fontSize: "0.8125rem" }}>{role.code}</code></span>
                <Button aria-label={`Cabut peran ${role.name}`} disabled={acting} onClick={() => void revokeRole(role)} variant="ghost">Cabut</Button>
              </li>
            ))}
          </ul>
        ) : <p style={{ color: "var(--color-text-secondary)" }}>Belum ada peran.</p>}
        {unassignedRoles.length ? (
          <FormField label="Tambah peran">
            {(props) => (
              <Select {...props} disabled={acting} onChange={(event) => void assignRole(event.target.value)} value="">
                <option disabled value="">Pilih peran…</option>
                {unassignedRoles.map((role) => <option key={role.id} value={role.id}>{role.name}</option>)}
              </Select>
            )}
          </FormField>
        ) : null}
      </section>
    </div>
  );
}