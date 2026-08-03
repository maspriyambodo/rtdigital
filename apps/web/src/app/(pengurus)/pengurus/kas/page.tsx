"use client";

import { useCallback, useEffect, useState } from "react";

import { ApiException } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import {
  type CashBook,
  type CashCategory,
  type CashType,
  createCashCategory,
  getCashBook,
  getCashCategories,
  recordCashTransaction,
  reverseCashTransaction,
  updateCashCategory,
} from "@/lib/cash";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { FileUploader } from "@/components/ui/FileUploader";
import { FormField } from "@/components/ui/FormField";
import { Select } from "@/components/ui/Select";
import { StatusBadge } from "@/components/ui/StatusBadge";
import { TextInput } from "@/components/ui/TextInput";

const rupiah = (amount: number) =>
  new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);

const today = () => new Date().toLocaleDateString("en-CA");

export default function BukuKasPengurusPage() {
  const { getAccessToken } = useAuth();
  const [book, setBook] = useState<CashBook>();
  const [categories, setCategories] = useState<CashCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  const [filters, setFilters] = useState({
    startDate: "",
    endDate: "",
    type: "" as CashType | "",
    categoryId: "",
  });
  const [categoryName, setCategoryName] = useState("");
  const [categoryType, setCategoryType] = useState<CashType>("expense");
  const [transaction, setTransaction] = useState({
    type: "expense" as CashType,
    categoryId: "",
    amount: "",
    transactionDate: today(),
    description: "",
  });
  const [proofFileId, setProofFileId] = useState<string>();
  const [draftTransactionId] = useState(() => crypto.randomUUID());
  const [reversalId, setReversalId] = useState<string>();
  const [reversalReason, setReversalReason] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const token = await getAccessToken();
      if (!token) throw new Error("Sesi telah berakhir. Silakan masuk kembali.");

      const [cashBook, cashCategories] = await Promise.all([
        getCashBook({
          startDate: filters.startDate || undefined,
          endDate: filters.endDate || undefined,
          type: filters.type || undefined,
          categoryId: filters.categoryId || undefined,
        }),
        getCashCategories(),
      ]);
      setBook(cashBook);
      setCategories(cashCategories);
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memuat buku kas.");
    } finally {
      setLoading(false);
    }
  }, [filters, getAccessToken]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
  }, [load]);

  const createCategory = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    setMessage("");
    try {
      await createCashCategory({ name: categoryName.trim(), type: categoryType });
      setCategoryName("");
      setMessage("Kategori kas ditambahkan.");
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal menambah kategori.");
    } finally {
      setSaving(false);
    }
  };

  const changeCategoryStatus = async (category: CashCategory) => {
    setSaving(true);
    setError("");
    try {
      await updateCashCategory(category.id, {
        status: category.status === "active" ? "inactive" : "active",
      });
      setMessage(`Kategori ${category.status === "active" ? "dinonaktifkan" : "diaktifkan"}.`);
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal memperbarui kategori.");
    } finally {
      setSaving(false);
    }
  };

  const recordTransaction = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    setMessage("");
    try {
      await recordCashTransaction({
        type: transaction.type,
        categoryId: transaction.categoryId,
        amount: Number(transaction.amount),
        transactionDate: transaction.transactionDate,
        description: transaction.description.trim(),
        proofFileId,
      });
      setTransaction({
        type: "expense",
        categoryId: "",
        amount: "",
        transactionDate: today(),
        description: "",
      });
      setProofFileId(undefined);
      setMessage("Transaksi kas dicatat.");
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal mencatat transaksi.");
    } finally {
      setSaving(false);
    }
  };

  const reverseTransaction = async () => {
    if (!reversalId || !reversalReason.trim()) return;

    setSaving(true);
    setError("");
    try {
      await reverseCashTransaction(reversalId, reversalReason.trim());
      setReversalId(undefined);
      setReversalReason("");
      setMessage("Transaksi pembalik berhasil dibuat.");
      await load();
    } catch (cause) {
      setError(cause instanceof ApiException || cause instanceof Error ? cause.message : "Gagal membalik transaksi.");
    } finally {
      setSaving(false);
    }
  };

  const activeCategories = categories.filter(
    (category) => category.status === "active" && category.type === transaction.type,
  );

  return (
    <div style={{ display: "grid", gap: "var(--space-6)" }}>
      <header>
        <h1 style={{ fontSize: "1.5rem", lineHeight: 1.2 }}>Buku Kas</h1>
        <p style={{ color: "var(--color-text-secondary)", marginTop: "var(--space-2)" }}>
          Catat mutasi kas tanpa menghapus riwayat transaksi.
        </p>
      </header>

      {message ? <p role="status" style={{ color: "var(--color-success-text)" }}>{message}</p> : null}
      {error ? <p role="alert" style={{ color: "var(--color-danger-text)" }}>{error}</p> : null}

      <section
        style={{
          display: "grid",
          gap: "var(--space-4)",
          gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 280px), 1fr))",
        }}
      >
        <form onSubmit={(event) => void recordTransaction(event)} style={cardStyle}>
          <h2 style={{ fontSize: "1.125rem" }}>Catat Transaksi</h2>
          <FormField label="Jenis transaksi">
            {({ id }) => (
              <Select
                id={id}
                value={transaction.type}
                onChange={(event) =>
                  setTransaction((current) => ({
                    ...current,
                    type: event.target.value as CashType,
                    categoryId: "",
                  }))
                }
              >
                <option value="expense">Pengeluaran</option>
                <option value="income">Pemasukan manual</option>
              </Select>
            )}
          </FormField>
          <FormField label="Kategori">
            {({ id }) => (
              <Select
                id={id}
                value={transaction.categoryId}
                onChange={(event) => setTransaction((current) => ({ ...current, categoryId: event.target.value }))}
                required
              >
                <option value="">Pilih kategori</option>
                {activeCategories.map((category) => (
                  <option key={category.id} value={category.id}>{category.name}</option>
                ))}
              </Select>
            )}
          </FormField>
          <FormField label="Nominal">
            {({ id }) => (
              <TextInput
                id={id}
                type="number"
                min="1"
                inputMode="numeric"
                value={transaction.amount}
                onChange={(event) => setTransaction((current) => ({ ...current, amount: event.target.value }))}
                required
              />
            )}
          </FormField>
          <FormField label="Tanggal transaksi">
            {({ id }) => (
              <TextInput
                id={id}
                type="date"
                value={transaction.transactionDate}
                onChange={(event) => setTransaction((current) => ({ ...current, transactionDate: event.target.value }))}
                required
              />
            )}
          </FormField>
          <FormField label="Keterangan">
            {({ id }) => (
              <TextInput
                id={id}
                value={transaction.description}
                onChange={(event) => setTransaction((current) => ({ ...current, description: event.target.value }))}
                required
              />
            )}
          </FormField>
          <FormField label="Bukti transaksi (opsional)">
            {() => (
              <FileUploader
                entityType="cash_transaction"
                entityId={draftTransactionId}
                onChange={setProofFileId}
                disabled={saving}
              />
            )}
          </FormField>
          <Button type="submit" disabled={saving}>{saving ? "Menyimpan…" : "Catat transaksi"}</Button>
        </form>

        <div style={cardStyle}>
          <h2 style={{ fontSize: "1.125rem" }}>Kategori Kas</h2>
          <form onSubmit={(event) => void createCategory(event)} style={{ display: "grid", gap: "var(--space-3)" }}>
            <FormField label="Nama kategori">
              {({ id }) => <TextInput id={id} value={categoryName} onChange={(event) => setCategoryName(event.target.value)} required />}
            </FormField>
            <FormField label="Jenis kategori">
              {({ id }) => (
                <Select id={id} value={categoryType} onChange={(event) => setCategoryType(event.target.value as CashType)}>
                  <option value="expense">Pengeluaran</option>
                  <option value="income">Pemasukan</option>
                </Select>
              )}
            </FormField>
            <Button type="submit" variant="secondary" disabled={saving}>Tambah kategori</Button>
          </form>

          <div style={{ display: "grid", gap: "var(--space-2)", marginTop: "var(--space-4)" }}>
            {categories.map((category) => (
              <div key={category.id} style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", justifyContent: "space-between", borderTop: "1px solid var(--color-border)", paddingTop: "var(--space-2)" }}>
                <span>{category.name} · {category.type === "income" ? "Pemasukan" : "Pengeluaran"}</span>
                <Button type="button" variant="secondary" disabled={saving} onClick={() => void changeCategoryStatus(category)}>
                  {category.status === "active" ? "Nonaktifkan" : "Aktifkan"}
                </Button>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section style={cardStyle}>
        <h2 style={{ fontSize: "1.125rem" }}>Buku Kas</h2>
        <form onSubmit={(event) => { event.preventDefault(); void load(); }} style={{ display: "grid", gap: "var(--space-3)", gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 160px), 1fr))" }}>
          <FormField label="Dari">{({ id }) => <TextInput id={id} type="date" value={filters.startDate} onChange={(event) => setFilters((current) => ({ ...current, startDate: event.target.value }))} />}</FormField>
          <FormField label="Sampai">{({ id }) => <TextInput id={id} type="date" value={filters.endDate} onChange={(event) => setFilters((current) => ({ ...current, endDate: event.target.value }))} />}</FormField>
          <FormField label="Jenis">{({ id }) => <Select id={id} value={filters.type} onChange={(event) => setFilters((current) => ({ ...current, type: event.target.value as CashType | "" }))}><option value="">Semua</option><option value="income">Pemasukan</option><option value="expense">Pengeluaran</option></Select>}</FormField>
          <FormField label="Kategori">{({ id }) => <Select id={id} value={filters.categoryId} onChange={(event) => setFilters((current) => ({ ...current, categoryId: event.target.value }))}><option value="">Semua</option>{categories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</Select>}</FormField>
          <Button type="submit" disabled={loading}>Terapkan filter</Button>
        </form>
      </section>

      {book ? (
        <section
          aria-label="Ringkasan buku kas"
          style={{
            display: "grid",
            gap: "var(--space-3)",
            gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 180px), 1fr))",
          }}
        >
          <Summary label="Pemasukan" amount={book.total_income} color="var(--color-success)" />
          <Summary label="Pengeluaran" amount={book.total_expense} color="var(--color-danger)" />
          <Summary label="Saldo" amount={book.balance} color="var(--color-primary-700)" />
        </section>
      ) : null}

      <section style={{ display: "grid", gap: "var(--space-3)" }}>
        <h2 style={{ fontSize: "1.125rem" }}>Mutasi Kas</h2>
        {loading ? <p>Memuat buku kas…</p> : !book || book.transactions.length === 0 ? (
          <EmptyState title="Tidak ada transaksi" description="Belum ada transaksi pada filter ini." />
        ) : (
          book.transactions.map((item) => (
            <article key={item.id} style={{ ...cardStyle, opacity: item.status === "reversed" ? 0.7 : 1 }}>
              <div style={{ display: "flex", justifyContent: "space-between", gap: "var(--space-3)", alignItems: "start" }}>
                <strong>{item.transaction_number}</strong>
                <StatusBadge variant={item.type === "income" ? "success" : "danger"}>
                  {item.type === "income" ? "+" : "-"} {rupiah(item.amount)}
                </StatusBadge>
              </div>
              <span>{item.description}</span>
              <span style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}>
                {item.transaction_date} · {item.category_name ?? "Sistem"} · Saldo: {rupiah(item.running_balance)}
              </span>
              {item.status === "reversed" ? <StatusBadge variant="neutral">Sudah dibalik</StatusBadge> : null}
              {item.status === "active" && !item.reversal_of_id && item.reference_type !== "payment" ? (
                reversalId === item.id ? (
                  <div style={{ display: "flex", gap: "var(--space-2)", flexWrap: "wrap" }}>
                    <TextInput aria-label="Alasan pembalikan" placeholder="Alasan pembalikan" value={reversalReason} onChange={(event) => setReversalReason(event.target.value)} />
                    <Button type="button" variant="danger" disabled={saving || !reversalReason.trim()} onClick={() => void reverseTransaction()}>Konfirmasi</Button>
                    <Button type="button" variant="secondary" onClick={() => setReversalId(undefined)}>Batal</Button>
                  </div>
                ) : (
                  <Button type="button" variant="secondary" onClick={() => setReversalId(item.id)}>Balikkan transaksi</Button>
                )
              ) : null}
            </article>
          ))
        )}
      </section>
    </div>
  );
}

function Summary({ label, amount, color }: { label: string; amount: number; color: string }) {
  return (
    <div style={{ ...cardStyle, gap: "var(--space-1)" }}>
      <span style={{ color: "var(--color-text-secondary)" }}>{label}</span>
      <strong style={{ color, fontSize: "1.25rem" }}>{rupiah(amount)}</strong>
    </div>
  );
}

const cardStyle = {
  display: "grid",
  gap: "var(--space-3)",
  padding: "var(--space-4)",
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-lg)",
} as const;