# Sistem Desain RT Digital

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital — Web App  
**Prinsip:** Desktop dan tablet-first, profesional, modern, responsif, konsisten dengan shadcn/ui, serta aksesibel sesuai WCAG 2.2 AA.

Dokumen ini menjadi acuan UI/UX frontend Next.js. Web App dioptimalkan untuk desktop dan tablet; aplikasi mobile akan dibuat dan didokumentasikan secara terpisah. Gunakan komponen shadcn/ui sebagai foundation, Tailwind CSS sebagai implementasi, dan CSS variables sebagai sumber token desain.

---

## 1. Prinsip Dasar

1. Prioritaskan efisiensi kerja pengurus dan kejelasan alur warga di web: dashboard, tagihan, pembayaran, surat, aduan, dan pelaporan.
2. Rancang dari desktop, kemudian sesuaikan secara responsif ke tablet. Mobile web hanya menjadi fallback yang tetap dapat diakses, bukan target desain utama.
3. Gunakan komponen `@/components/ui/*` dari shadcn/ui secara konsisten; jangan membuat ulang pola standar tanpa alasan yang jelas.
4. Status, kesalahan, loading, dan langkah berikutnya selalu ditampilkan dengan teks; warna bukan satu-satunya penanda.
5. Data sensitif dimasking secara default dan hanya ditampilkan penuh untuk peran yang berwenang.
6. Utamakan hierarki informasi, whitespace yang cukup, dan kepadatan data yang terkontrol; hindari dekorasi atau animasi yang tidak membantu tugas.
7. Semua komponen mendukung state default, hover, focus-visible, active, disabled, loading, dan error bila relevan.

---

## 2. Design Tokens dan Tema

Gunakan token semantik shadcn/ui pada `globals.css`. Gunakan utility seperti `bg-background`, `text-foreground`, `border-border`, dan `ring-ring`; jangan mengikat komponen pada nilai warna mentah.

```css
@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 221.2 83.2% 53.3%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 221.2 83.2% 53.3%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 217.2 91.2% 59.8%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 224.3 76.3% 48%;
  }
}
```

### 2.1 Aturan Warna

| Status | Penggunaan |
|---|---|
| `primary` | Aksi utama, tautan aktif, focus ring |
| Success | Lunas, disetujui, selesai; gunakan badge hijau dengan teks kontras |
| Warning | Menunggu verifikasi, perlu perbaikan, tunggakan; gunakan amber/oranye dengan teks kontras |
| `destructive` | Ditolak, dibatalkan, gagal, tindakan destruktif |
| `secondary` / `muted` | Draft, nonaktif, metadata, informasi netral |

- Kontras teks normal minimal **4.5:1** dan teks besar minimal **3:1**.
- Status wajib memiliki label teks; ikon dapat ditambahkan untuk mempercepat pemindaian visual.
- Jangan gunakan warna semantic mentah untuk teks tanpa memverifikasi kontras pada light dan dark mode.

---

## 3. Tipografi

Gunakan `Inter`, `Geist`, atau system font stack yang dikonfigurasi proyek.

| Hirarki | Kelas Tailwind | Penggunaan |
|---|---|---|
| H1 | `text-3xl font-bold tracking-tight` | Judul halaman / dashboard |
| H2 | `text-2xl font-semibold tracking-tight` | Judul section |
| H3 | `text-xl font-semibold` | Judul card, sheet, dialog |
| H4 | `text-lg font-medium` | Sub-section |
| Body | `text-sm` atau `text-base` | Isi, tabel, dan form |
| Metadata | `text-sm text-muted-foreground` | Bantuan, tanggal, keterangan |
| Angka | `tabular-nums` | Nominal, nomor tagihan, dan data finansial |

Format tanggal dan nominal mengikuti Indonesia, misalnya `1 Agustus 2026` dan `Rp150.000`.

---

## 4. Spacing, Grid, dan Layout

Gunakan skala Tailwind kelipatan 4 px: `1` (4 px), `2` (8 px), `3` (12 px), `4` (16 px), `6` (24 px), dan `8` (32 px).

- Konten standar memakai `max-w-7xl` (1280 px); halaman tabel kompleks dapat memakai lebar penuh dengan padding yang cukup.
- Padding konten: `p-6` untuk tablet dan desktop standar, `p-8` untuk layar lebar atau halaman analitis.
- Card memakai `p-4` atau `p-6`; gap standar antarkontrol dan field adalah `gap-4`.
- Form desktop/tablet memakai grid 2 kolom bila field saling independen; gunakan 3 kolom hanya pada data ringkas.
- Jangan membuat teks, form, atau detail record terlalu lebar; gunakan `max-w-*` yang sesuai konteks.

---

## 5. Responsive Layout dan Navigasi

| Breakpoint | Lebar | Aturan utama |
|---|---:|---|
| `md` | `≥ 768 px` | Tablet; header lengkap, grid dashboard 2 kolom, sidebar menjadi Sheet bila ruang terbatas |
| `lg` | `≥ 1024 px` | Desktop; sidebar permanen/collapsible dan tabel administrasi lengkap |
| `xl` | `≥ 1280 px` | Layar lebar; dashboard multi-kolom dan split-pane bila bermanfaat |

### 5.1 Navigasi

- Desktop memakai sidebar kiri yang dapat diciutkan (`w-64` / icon-only), dengan grup menu yang jelas dan status halaman aktif.
- Tablet memakai header sticky dan `<Sheet>` untuk navigasi bila sidebar tidak muat.
- Header memuat breadcrumb, pencarian global bila tersedia, notifikasi, serta `<DropdownMenu>` profil pengguna.
- Tidak menggunakan bottom navigation atau pola khusus aplikasi mobile.

### 5.2 Data dan Dashboard

- Gunakan grid statistik, chart yang relevan, dan daftar prioritas pada dashboard; setiap visualisasi harus mendukung keputusan atau tindakan.
- Daftar administratif menggunakan Data Table. Bila viewport sangat sempit, sediakan scroll horizontal yang jelas atau sembunyikan kolom sekunder; jangan mengecilkan teks secara berlebihan.

---

## 6. Komponen shadcn/ui

| Kebutuhan | Komponen |
|---|---|
| Aksi | `Button`, `Toggle`, `Tooltip` |
| Form | `Form`, `Input`, `Textarea`, `Select`, `Checkbox`, `RadioGroup`, `Switch`, `Popover`, `Calendar` |
| Data | `Table`, `DropdownMenu`, `Pagination`, `Badge` |
| Overlay | `Dialog`, `AlertDialog`, `Sheet`, `Popover`, `Command` |
| Feedback | `Alert`, `Skeleton`, `Sonner`/`Toast`, `Progress` |
| Struktur | `Card`, `Tabs`, `Separator`, `ScrollArea`, `Breadcrumb` |

Gunakan `lucide-react` untuk ikon. Tombol icon-only wajib memiliki `aria-label` dan `Tooltip`.

### 6.1 Tombol dan Kontrol

- `<Button variant="default">` untuk aksi utama; satu aksi utama dominan per section atau dialog.
- Gunakan `outline`, `secondary`, atau `ghost` untuk aksi alternatif; gunakan `destructive` hanya untuk aksi berisiko.
- Ukuran tombol mengikuti shadcn/ui. Kontrol yang sering digunakan tetap memiliki area klik nyaman, minimal 40 px dan idealnya 44 px untuk aksi penting.
- `focus-visible` memakai ring token `ring-ring`; state loading menampilkan spinner dan mencegah klik ganda.
- Tindakan destruktif memerlukan `<AlertDialog>` dengan aksi Batal yang jelas dan bukan sebagai aksi default.

---

## 7. Formulir

Gunakan `react-hook-form` dan `zod` melalui komponen `Form` shadcn/ui.

- `FormLabel` berada di atas input; placeholder tidak menggantikan label.
- Field wajib diberi indikator yang dapat dibaca screen reader.
- `FormDescription` menjelaskan aturan input; `FormMessage` menampilkan error spesifik di bawah field.
- Validasi client-side hanya untuk umpan balik cepat; backend tetap sumber validasi akhir.
- Fokuskan field error pertama setelah submit gagal dan tampilkan ringkasan error untuk form panjang.
- Gunakan `<Select>` atau combobox berbasis `Command` untuk daftar panjang; gunakan `Popover` + `Calendar` untuk date picker web.
- Input nominal menampilkan format Rupiah tanpa mengubah nilai yang disimpan. Upload menampilkan tipe, batas ukuran, progres, hasil gagal/coba lagi, serta penghapusan sebelum submit.
- Jangan menyimpan token, NIK, nomor KK, atau file dokumen di `localStorage`.

---

## 8. Tabel dan Daftar

Gunakan `@tanstack/react-table` bersama komponen `Table` shadcn/ui untuk daftar Warga, Tagihan, Pembayaran, Surat, Kas, dan Aduan.

- Header tabel semantik, jelas, dan sticky pada daftar panjang.
- Sediakan pencarian, filter, sorting dengan indikator aksesibel, pagination atau cursor navigation, serta column visibility bila diperlukan.
- Kolom aksi memakai `<DropdownMenu>` bila aksi lebih dari dua.
- Baris memiliki hover ringan, tetapi aksi tidak boleh hanya dapat ditemukan saat hover.
- Nominal dan angka memakai `tabular-nums`; status memakai `Badge` dengan teks yang konsisten terhadap `USER_FLOW.md` dan `DATABASE_DESIGN.md`.

---

## 9. Dialog, Sheet, dan Popover

| Konteks | Komponen |
|---|---|
| Konfirmasi tindakan kritis | `AlertDialog` |
| Form atau detail singkat | `Dialog` |
| Filter lanjutan, detail record, navigasi tablet | `Sheet` |
| Pilihan ringkas / bantuan kontekstual | `Popover` atau `Tooltip` |

- Dialog dan Sheet harus memiliki judul, deskripsi bila diperlukan, focus trap, dan focus kembali ke pemicu setelah ditutup.
- `Escape` dapat menutup overlay bila aman. Hindari modal bertumpuk dan jangan gunakan modal untuk alur multi-langkah yang panjang.

---

## 10. Status, Card, dan Feedback

### Status Badge

| Kategori | Contoh |
|---|---|
| Success | Lunas, Disetujui, Diterbitkan, Selesai |
| Warning | Menunggu Verifikasi, Perlu Perbaikan, Dibayar Sebagian |
| Info / Neutral | Baru, Diajukan, Diproses, Draft, Nonaktif |
| Destructive | Ditolak, Dibatalkan, Gagal |

Badge memakai `<Badge>`, label status eksplisit, dan warna semantic yang memenuhi kontras pada kedua tema.

### Card, Empty State, Loading, dan Error

- Gunakan `<Card>` dengan border halus; shadow hanya digunakan saat membantu pemisahan hirarki.
- Empty state menjelaskan kondisi dan menawarkan satu tindakan relevan sesuai izin pengguna.
- Gunakan `<Skeleton>` untuk struktur data yang sudah diketahui; spinner/progress hanya untuk aksi lokal seperti menyimpan atau upload.
- Pesan error memakai bahasa pengguna, menyediakan tombol **Coba lagi** bila aman, serta menampilkan `request_id` untuk error server yang memerlukan bantuan.

---

## 11. Dark Mode

Dark mode menggunakan `next-themes` dan class `.dark`. Semua komponen harus memakai token semantik shadcn/ui, bukan warna hard-coded.

- Default dapat mengikuti `prefers-color-scheme`; pengguna dapat memilih light, dark, atau system.
- Simpan hanya preferensi tema, bukan data pribadi.
- Uji kontras teks, badge, input error, tooltip, focus ring, tabel, dan overlay pada light serta dark mode.

---

## 12. Checklist Implementasi

- [ ] UI menggunakan `@/components/ui/*` shadcn/ui sebagai foundation.
- [ ] Layout desktop dan tablet diuji pada breakpoint `md`, `lg`, dan `xl`.
- [ ] Sidebar, header sticky, breadcrumb, dan navigasi tablet melalui Sheet diterapkan secara konsisten.
- [ ] Tabel mendukung pencarian, filter, sorting, pagination, dan aksi yang aksesibel.
- [ ] Form memakai `react-hook-form`, `zod`, label, validasi spesifik, serta focus management.
- [ ] Semua kontrol dapat dioperasikan dengan keyboard dan memiliki `focus-visible`.
- [ ] Semua status memiliki teks; warna dan ikon hanya menjadi pendukung.
- [ ] Kontras WCAG AA diverifikasi pada light dan dark mode.
- [ ] Skeleton, empty state, error state, dan loading action tersedia pada alur utama.
- [ ] Tidak ada bottom navigation, safe-area mobile, atau pola mobile-first di Web App ini.
