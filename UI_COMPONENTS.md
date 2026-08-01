# Komponen UI Reusable

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Referensi:** `DESIGN_SYSTEM.md`, `INFORMATION_ARCHITECTURE.md`  
**Platform:** Next.js / React

Komponen dibangun mobile-first, aksesibel, dan mengikuti design token dalam `DESIGN_SYSTEM.md`. Gunakan elemen HTML native sebagai dasar. Jangan menambah library UI hanya untuk komponen sederhana.

---

## 1. Standar Implementasi

1. Komponen interaktif memakai `"use client"` hanya bila membutuhkan state, event browser, atau API browser.
2. Komponen presentasional tetap Server Component bila memungkinkan.
3. Tombol, input, select, dialog, dan kontrol navigasi harus mendukung keyboard serta `focus-visible`.
4. Area interaktif minimum **44 × 44 CSS pixel**.
5. Input teks minimum **16 px** untuk mencegah auto-zoom Safari iOS.
6. Semua status memakai teks; warna bukan satu-satunya penanda.
7. Props native relevan diteruskan ke elemen dasar melalui `...props`.
8. Input, select, date picker, dan tombol memakai `forwardRef`.
9. State `loading`, `disabled`, `error`, dan `empty` wajib tersedia bila relevan.
10. Data privat tidak disimpan oleh komponen UI ke `localStorage`.

---

## 2. Komponen Inti

## 2.1 `Button`

Tombol untuk aksi utama, sekunder, atau destruktif.

```ts
type ButtonProps = {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  isLoading?: boolean;
  fullWidth?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
} & React.ButtonHTMLAttributes<HTMLButtonElement>;
```

**Aturan:**

- Tinggi minimum 44 px pada ukuran `md` dan `lg`.
- `isLoading` menampilkan indikator, menonaktifkan klik ganda, serta mempertahankan label tombol.
- Hanya satu aksi `primary` dominan per section atau dialog.
- Tombol icon-only wajib memiliki `aria-label`.

---

## 2.2 `FormField`

Wrapper layout untuk label, bantuan, error, dan input form.

```ts
type FormFieldProps = {
  id: string;
  label: string;
  required?: boolean;
  hint?: string;
  error?: string;
  children: React.ReactNode;
};
```

**Aturan:**

- Label memakai `htmlFor={id}`.
- Pesan bantuan dan error dihubungkan ke input melalui `aria-describedby`.
- Saat error, input memakai `aria-invalid="true"`.
- Label berada di atas input.
- Placeholder tidak menggantikan label.
- Error ditampilkan dengan teks spesifik, misalnya: `Nominal harus lebih besar dari Rp0.`

---

## 2.3 `TextInput`

Input standar untuk teks, email, telepon, nominal, atau password.

```ts
type TextInputProps = {
  hasError?: boolean;
} & React.InputHTMLAttributes<HTMLInputElement>;
```

**Aturan:**

- Mendukung `type="text"`, `email`, `tel`, `password`, dan `number`.
- Nomor telepon memakai `type="tel"`.
- Nominal memakai `inputMode="numeric"` bila sesuai.
- Password memakai `autocomplete` tepat dan tombol tampilkan/sembunyikan password yang aksesibel.
- Tidak melakukan format nilai yang mengubah data saat pengguna sedang mengetik.

---

## 2.4 `Select`

Pemilih opsi terstruktur.

```ts
type SelectOption = {
  label: string;
  value: string;
  disabled?: boolean;
};

type SelectProps = {
  options: SelectOption[];
  placeholder?: string;
  hasError?: boolean;
} & Omit<React.SelectHTMLAttributes<HTMLSelectElement>, "children">;
```

**Aturan:**

- Gunakan elemen native `<select>` pada MVP.
- Native select memberi pengalaman pemilih terbaik pada Android dan iOS.
- Tinggi minimum 44 px, font minimum 16 px.
- Opsi placeholder memiliki `value=""` serta `disabled` bila field wajib.

---

## 2.5 `DatePicker`

Pemilih tanggal bisnis.

```ts
type DatePickerProps = {
  hasError?: boolean;
  min?: string;
  max?: string;
} & Omit<React.InputHTMLAttributes<HTMLInputElement>, "type">;
```

**Aturan:**

- Implementasi MVP memakai `<input type="date">`.
- Nilai memakai format `YYYY-MM-DD`.
- Gunakan picker native perangkat; jangan tambahkan kalender JavaScript khusus tanpa kebutuhan tervalidasi.
- Tanggal ditampilkan sesuai locale Indonesia pada area baca saja.

---

## 2.6 `FileUploader`

Unggah file privat ke S3 menggunakan pre-signed URL.

```ts
type UploadedFile = {
  id: string;
  originalName: string;
  mimeType: string;
  sizeBytes: number;
};

type FileUploaderProps = {
  accept: string;
  maxSizeBytes: number;
  value?: UploadedFile;
  capture?: "environment" | "user";
  disabled?: boolean;
  onSelect: (file: File) => void;
  onRemove?: () => void;
  progress?: number;
  error?: string;
};
```

**State:**

- `idle`: belum memilih file.
- `selected`: file lolos validasi lokal.
- `uploading`: progress terlihat.
- `uploaded`: metadata file tersedia.
- `error`: alasan gagal dan tombol coba lagi.
- `disabled`: unggah tidak tersedia.

**Aturan:**

- Validasi MIME type, ekstensi, dan ukuran di client untuk UX; backend tetap memvalidasi kembali.
- Untuk foto, dukung tombol **Ambil Foto** dengan `capture="environment"` bila browser mendukung.
- Tampilkan nama file, ukuran, progress, tombol hapus, dan preview gambar bila aman.
- File biner diunggah browser langsung ke S3; komponen hanya memegang metadata hasil unggah.
- Jangan menyimpan file, URL signed, NIK, atau dokumen di `localStorage`.

---

## 2.7 `StatusBadge`

Status ringkas untuk tagihan, surat, aduan, akun, dan proses lain.

```ts
type StatusBadgeProps = {
  variant: "success" | "warning" | "danger" | "info" | "neutral";
  label: string;
  icon?: React.ReactNode;
};
```

| Variant | Contoh |
|---|---|
| `success` | Lunas, Disetujui, Diterbitkan, Selesai |
| `warning` | Menunggu Verifikasi, Perlu Perbaikan, Dibayar Sebagian |
| `danger` | Ditolak, Dibatalkan, Gagal |
| `info` | Baru, Diajukan, Diproses |
| `neutral` | Draft, Nonaktif |

**Aturan:**

- Tinggi minimum 24 px.
- Padding horizontal 8 px.
- Ukuran teks 12 px, berat 500.
- Label wajib selalu terlihat.

---

## 2.8 `EmptyState`

Tampilan saat daftar atau pencarian tidak menghasilkan data.

```ts
type EmptyStateProps = {
  title: string;
  description?: string;
  icon?: React.ReactNode;
  action?: React.ReactNode;
};
```

**Contoh:**

- `Belum ada tagihan bulan ini.`
- `Belum ada pengajuan surat.`
- `Tidak ditemukan warga dengan pencarian tersebut.`

**Aturan:**

- Beri satu aksi lanjutan bila pengguna memiliki izin.
- Hindari ilustrasi besar atau aset berat.
- Jangan tampilkan tombol aksi tanpa permission.

---

## 2.9 `Pagination`

Navigasi daftar berbasis cursor.

```ts
type PaginationProps = {
  hasNext: boolean;
  hasPrevious: boolean;
  onNext: () => void;
  onPrevious: () => void;
  isLoading?: boolean;
  label?: string;
};
```

**Aturan:**

- MVP mengutamakan cursor pagination API.
- Tombol **Sebelumnya** dan **Berikutnya** minimum 44 px.
- Tombol nonaktif bila cursor tidak tersedia.
- Tampilkan `aria-live="polite"` saat daftar berhasil diperbarui.
- Jangan membuat nomor halaman palsu bila API memakai cursor.

---

## 2.10 `DataTable`

Tabel administrasi pengurus untuk desktop. Tampilan mobile memakai kartu melalui `renderMobileCard`.

```ts
type DataTableColumn<T> = {
  id: string;
  header: string;
  sortable?: boolean;
  className?: string;
  render: (row: T) => React.ReactNode;
};

type DataTableProps<T> = {
  columns: DataTableColumn<T>[];
  data: T[];
  getRowId: (row: T) => string;
  renderMobileCard: (row: T) => React.ReactNode;
  emptyState?: React.ReactNode;
  isLoading?: boolean;
  sort?: { field: string; direction: "asc" | "desc" };
  onSort?: (field: string, direction: "asc" | "desc") => void;
  pagination?: React.ReactNode;
};
```

**Aturan:**

- Desktop memakai elemen semantik `<table>`, `<thead>`, `<tbody>`, `<th scope="col">`.
- Kolom hanya boleh diurutkan bila API mendukung sorting tersebut.
- Mobile `< 768 px`: gunakan `renderMobileCard`; jangan mengecilkan tabel sampai tidak terbaca.
- Bila kartu belum dapat dibuat, gunakan scroll horizontal yang jelas sebagai fallback.
- Data sensitif tetap dimasking pada tabel dan kartu.
- `isLoading` menampilkan skeleton, bukan tabel kosong.

---

## 2.11 `CardList`

Daftar kartu untuk layar seluler.

```ts
type CardListProps<T> = {
  items: T[];
  getItemId: (item: T) => string;
  renderItem: (item: T) => React.ReactNode;
  isLoading?: boolean;
  emptyState?: React.ReactNode;
};
```

**Aturan:**

- Kartu menampilkan informasi utama, metadata, dan `StatusBadge`.
- Seluruh kartu dapat diketuk hanya bila membuka satu detail yang jelas.
- Aksi lebih dari dua disimpan dalam menu overflow berlabel.
- Urutan informasi penting: status, judul/nomor, nominal atau tanggal, metadata.

---

## 2.12 `ConfirmationDialog`

Dialog konfirmasi untuk tindakan penting dan destruktif.

```ts
type ConfirmationDialogProps = {
  open: boolean;
  title: string;
  description: React.ReactNode;
  confirmText: string;
  cancelText?: string;
  confirmVariant?: "primary" | "danger";
  isLoading?: boolean;
  onConfirm: () => void | Promise<void>;
  onOpenChange: (open: boolean) => void;
};
```

**Gunakan untuk:**

- Menolak atau membatalkan pembayaran.
- Membatalkan tagihan.
- Membuat transaksi pembalik kas.
- Menolak surat atau aduan.
- Menonaktifkan akun, warga, keluarga, atau rumah/unit.

**Aturan:**

- Fokus berpindah ke dialog saat terbuka dan kembali ke pemicu saat tertutup.
- Fokus tidak boleh keluar dari dialog saat aktif.
- `Escape` menutup dialog hanya bila tidak sedang memproses aksi.
- Tombol **Batal** selalu tersedia dan tidak diberi gaya destruktif.
- Latar belakang tidak dapat di-scroll selama dialog terbuka.
- Tindakan berisiko menjelaskan dampak dan ketidakdapatbalikan.

---

## 2.13 `BottomSheet`

Panel dari bawah untuk filter, pilihan singkat, atau aksi sekunder di mobile.

```ts
type BottomSheetProps = {
  open: boolean;
  title: string;
  children: React.ReactNode;
  onOpenChange: (open: boolean) => void;
};
```

**Aturan:**

- Memiliki handle, judul, tombol tutup, dan focus management.
- Menghormati `safe-area-inset-bottom`.
- Untuk form panjang gunakan halaman tersendiri, bukan bottom sheet.
- Pada desktop dapat ditampilkan sebagai dialog kecil.

---

## 3. Komponen Pendukung

| Komponen | Fungsi |
|---|---|
| `AppShell` | Layout area Warga/Pengurus, header, outlet konten, safe area. |
| `BottomNavigation` | Navigasi Warga maksimal lima tujuan utama. |
| `Sidebar` | Navigasi Pengurus desktop berbasis permission. |
| `NavigationDrawer` | Navigasi Pengurus mobile/tablet. |
| `PageHeader` | Judul halaman, deskripsi, breadcrumb desktop, aksi utama. |
| `SearchInput` | Input pencarian dengan debounce dikendalikan halaman. |
| `FilterBar` | Filter daftar; mobile dibuka melalui `BottomSheet`. |
| `Skeleton` | Placeholder loading untuk card, tabel, dan detail. |
| `Timeline` | Riwayat status surat, aduan, pembayaran, atau koreksi data. |
| `OfflineNotice` | Indikator jaringan lemah/offline dan status retry. |
| `ErrorState` | Pesan error, tombol coba lagi, serta `request_id` bila tersedia. |

---

## 4. Struktur Direktori Rekomendasi

```text
apps/web/src/components/
├── ui/
│   ├── button.tsx
│   ├── form-field.tsx
│   ├── text-input.tsx
│   ├── select.tsx
│   ├── date-picker.tsx
│   ├── file-uploader.tsx
│   ├── status-badge.tsx
│   ├── empty-state.tsx
│   ├── pagination.tsx
│   ├── data-table.tsx
│   ├── card-list.tsx
│   ├── confirmation-dialog.tsx
│   └── bottom-sheet.tsx
├── layout/
│   ├── app-shell.tsx
│   ├── bottom-navigation.tsx
│   ├── sidebar.tsx
│   └── navigation-drawer.tsx
└── feedback/
    ├── skeleton.tsx
    ├── error-state.tsx
    └── offline-notice.tsx
```

Jangan membuat komponen domain seperti `InvoiceTable` atau `ComplaintStatus` sebelum pola generik di atas terbukti tidak mencukupi.

---

## 5. Checklist Penerimaan

- [ ] Komponen memakai token dari `DESIGN_SYSTEM.md`.
- [ ] Kontrol interaktif minimum 44 × 44 px.
- [ ] Input teks minimum 16 px.
- [ ] Fokus keyboard dan `focus-visible` bekerja.
- [ ] Form menautkan label, hint, dan error secara aksesibel.
- [ ] `DataTable` memiliki alternatif kartu mobile.
- [ ] `FileUploader` menangani validasi, progress, gagal, coba lagi, dan hapus.
- [ ] `ConfirmationDialog` memiliki focus trap dan tombol Batal.
- [ ] `StatusBadge` selalu memiliki label teks.
- [ ] Komponen tidak menyimpan data pribadi atau token secara tidak aman.