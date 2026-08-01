"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { ApiException, apiFetch } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/Button";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { TextInput } from "@/components/ui/TextInput";

interface Role {
  id: string;
  code: string;
  name: string;
}

export default function InviteUserPage() {
  const router = useRouter();
  const { getAccessToken } = useAuth();
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [roleCodes, setRoleCodes] = useState<string[]>(["warga"]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const authorized = useCallback(
    async <T,>(path: string, options: RequestInit = {}) => {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir.");
      return apiFetch<T>(path, {
        ...options,
        headers: { ...options.headers, Authorization: `Bearer ${token}` },
      });
    },
    [getAccessToken],
  );

  useEffect(() => {
    void authorized<Role[]>("roles")
      .then(setRoles)
      .catch(() => setError("Gagal memuat peran yang tersedia."));
  }, [authorized]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      await authorized("users/invite", {
        method: "POST",
        body: JSON.stringify({ email, phone, role_codes: roleCodes }),
      });
      router.replace("/pengurus/pengguna");
    } catch (cause) {
      setError(cause instanceof ApiException ? cause.message : "Gagal mengirim undangan.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={submit} style={{ display: "grid", gap: "var(--space-6)", maxWidth: 560 }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Undang Pengguna</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
          Tautan aktivasi dikirim lewat email bila tersedia.
        </p>
      </header>

      {error ? <p role="alert" style={{ background: "var(--color-danger)", borderRadius: "var(--radius-md)", color: "#fff", padding: "var(--space-3)" }}>{error}</p> : null}

      <FormField label="Email">
        {(props) => <TextInput {...props} type="email" value={email} onChange={(event) => setEmail(event.target.value)} />}
      </FormField>
      <FormField label="Nomor telepon">
        {(props) => <TextInput {...props} inputMode="tel" type="tel" value={phone} onChange={(event) => setPhone(event.target.value)} />}
      </FormField>
      <FormField label="Peran" required hint="Pilih setidaknya satu peran.">
        {(props) => (
          <Select {...props} multiple required value={roleCodes} onChange={(event) => setRoleCodes(Array.from(event.target.selectedOptions, (option) => option.value))}>
            {roles.map((role) => <option key={role.id} value={role.code}>{role.name}</option>)}
          </Select>
        )}
      </FormField>

      <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-3)" }}>
        <Button loading={loading} type="submit">Kirim undangan</Button>
        <Button disabled={loading} onClick={() => router.back()} type="button" variant="ghost">Batal</Button>
      </div>
    </form>
  );
}