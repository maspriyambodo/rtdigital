"use client";

import { useCallback, useEffect, useState } from "react";

import { Button } from "../../../../components/ui/Button";
import { EmptyState } from "../../../../components/ui/EmptyState";
import { FormField } from "../../../../components/ui/FormField";
import { Select } from "../../../../components/ui/Select";
import { StatusBadge, type StatusBadgeProps } from "../../../../components/ui/StatusBadge";
import { TextInput } from "../../../../components/ui/TextInput";
import { useAuth } from "../../../../lib/auth-context";
import {
  createComplaintCategory,
  listComplaintCategories,
  updateComplaintCategory,
  type ComplaintCategory,
} from "../../../../lib/complaints";

function statusVariant(status: ComplaintCategory["status"]): StatusBadgeProps["variant"] {
  return status === "active" ? "success" : "info";
}

export default function KategoriAduanPage() {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<ComplaintCategory[]>([]);
  const [editing, setEditing] = useState<ComplaintCategory>();
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [status, setStatus] = useState<ComplaintCategory["status"]>("active");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      setItems(await listComplaintCategories(token, false));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Data kategori gagal dimuat.");
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const resetForm = () => {
    setEditing(undefined);
    setCode("");
    setName("");
    setStatus("active");
  };

  const startEdit = (item: ComplaintCategory) => {
    setEditing(item);
    setCode(item.code);
    setName(item.name);
    setStatus(item.status);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!name.trim() || (!editing && !code.trim())) {
      setError("Isi kode dan nama kategori.");
      return;
    }

    setSaving(true);
    setError("");
    setNotice("");
    try {
      const token = await getAccessToken();
      if (!token) return;
      if (editing) {
        await updateComplaintCategory(token, editing.id, { name, status });
        setNotice("Kategori diperbarui.");
      } else {
        await createComplaintCategory(token, { code, name });
        setNotice("Kategori dibuat.");
      }
      resetForm();
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Kategori gagal disimpan.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div style={{ display: "grid", gap: "var(--space-4)", margin: "0 auto", maxWidth: 760, padding: "var(--space-4)" }}>
      <header>
        <h1 style={{ fontSize: "1.25rem", margin: 0 }}>Kategori Aduan</h1>
        <p style={{ color: "var(--color-text-secondary)", margin: "var(--space-1) 0 0" }}>Kelola data master kategori aduan.</p>
      </header>

      {error ? <p role="alert" style={{ color: "var(--color-danger)", margin: 0 }}>{error}</p> : null}
      {notice ? <p role="status" style={{ color: "var(--color-success)", margin: 0 }}>{notice}</p> : null}

      <section style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-lg)", padding: "var(--space-4)" }}>
        <h2 style={{ fontSize: "1rem", marginTop: 0 }}>{editing ? "Ubah" : "Tambah"} kategori</h2>
        <form onSubmit={submit} style={{ display: "grid", gap: "var(--space-3)" }}>
          <FormField label="Kode" required={!editing}>
            {(props) => <TextInput {...props} required={!editing} disabled={editing !== undefined} value={code} onChange={(event) => setCode(event.target.value)} />}
          </FormField>
          <FormField label="Nama" required>
            {(props) => <TextInput {...props} required value={name} onChange={(event) => setName(event.target.value)} />}
          </FormField>
          {editing ? (
            <FormField label="Status">
              {(props) => <Select {...props} value={status} onChange={(event) => setStatus(event.target.value as ComplaintCategory["status"])}>
                <option value="active">Aktif</option>
                <option value="inactive">Nonaktif</option>
              </Select>}
            </FormField>
          ) : null}
          <div style={{ display: "flex", gap: "var(--space-2)" }}>
            <Button type="submit" loading={saving} disabled={saving}>{editing ? "Simpan" : "Tambah"}</Button>
            {editing ? <Button type="button" variant="outline" onClick={resetForm}>Batal</Button> : null}
          </div>
        </form>
      </section>

      <section>
        <h2 style={{ fontSize: "1rem" }}>Daftar kategori</h2>
        {loading ? <p>Memuat…</p> : items.length === 0 ? <EmptyState title="Belum ada kategori" description="Kategori aduan akan muncul di sini." /> : <div style={{ display: "grid", gap: "var(--space-3)" }}>{items.map((item) => (
          <article key={item.id} style={{ background: "var(--color-surface)", border: "1px solid var(--color-border)", borderRadius: "var(--radius-md)", display: "grid", gap: "var(--space-2)", padding: "var(--space-3)" }}>
            <div style={{ alignItems: "start", display: "flex", justifyContent: "space-between", gap: "var(--space-2)" }}>
              <div><strong>{item.name}</strong><br /><small>{item.code}</small></div>
              <StatusBadge variant={statusVariant(item.status)}>{item.status === "active" ? "Aktif" : "Nonaktif"}</StatusBadge>
            </div>
            <div><Button type="button" variant="outline" onClick={() => startEdit(item)}>Ubah</Button></div>
          </article>
        ))}</div>}
      </section>
    </div>
  );
}