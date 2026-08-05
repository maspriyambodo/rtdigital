# Peran dan Hak Akses Pengguna

**Status:** Draft untuk validasi  
**Berlaku untuk:** MVP RT Digital  
**Model akses:** Role-Based Access Control (RBAC) dengan pembatasan data berdasarkan organisasi, keluarga, kepemilikan data, dan penugasan.

Dokumen ini melengkapi `PRD.md` dan `SCOPE.md`. Hak akses selalu diperiksa oleh backend API; penyembunyian menu frontend bukan kontrol keamanan.

---

## 1. Prinsip Akses

1. **Least privilege:** setiap peran hanya menerima izin minimum untuk tugasnya.
2. **Scope data:** seluruh akses dibatasi pada `organization_id` aktif.
3. **Data sendiri:** Warga hanya dapat mengakses data diri, keluarga, tagihan, pembayaran, surat, dan aduan yang menjadi miliknya sesuai perannya dalam keluarga.
4. **Data sensitif:** NIK, nomor KK, nomor telepon, dokumen pribadi, dan data khusus dimasking secara default. Pembukaan data utuh memerlukan izin khusus serta audit log.
5. **Multi-peran:** satu akun dapat memiliki lebih dari satu peran. Izin efektif adalah gabungan izin peran, tetap tunduk pada batas organisasi dan kepemilikan data.
6. **Pemisahan tugas:** pembuat atau pemilik transaksi tidak boleh memverifikasi, menyetujui, atau membatalkan transaksinya sendiri.
7. **Tanpa hard delete:** data penting menggunakan penonaktifan, arsip, pembatalan, atau transaksi pembalik. Transaksi keuangan dan audit log tidak dapat dihapus dari antarmuka.
8. **Audit:** akses data sensitif, perubahan peran, ekspor, tindakan keuangan, dan persetujuan surat wajib tercatat.
9. **MFA wajib:** Super Admin, Ketua RT, Sekretaris, dan Bendahara wajib menggunakan MFA.

---

## 2. Definisi Peran

| Peran | Tujuan | Batas Utama |
|---|---|---|
| **Super Admin** | Mengelola platform dan organisasi secara teknis. | Tidak mengakses data operasional atau dokumen warga kecuali pemulihan/investigasi darurat yang disetujui dan diaudit. |
| **Ketua RT** | Memantau operasional, menetapkan pengurus, menyetujui keputusan penting dan surat. | Akses keuangan terutama baca/ekspor; tidak melakukan verifikasi pembayaran rutin. |
| **Sekretaris** | Mengelola administrasi rumah, keluarga, warga, pengumuman, agenda, dan proses surat. | Tidak mengelola kas, tagihan, atau verifikasi pembayaran tanpa penugasan darurat eksplisit. |
| **Bendahara** | Mengelola iuran, tagihan, pembayaran, kas, serta laporan keuangan. | Tidak mengubah data kependudukan utama atau menyetujui surat. |
| **Pengurus** | Menjalankan tugas operasional tertentu, misalnya keamanan, kebersihan, atau koordinator wilayah. | Tidak memiliki akses bawaan ke seluruh data warga atau keuangan; izin diberikan per tugas. |
| **Warga** | Mengakses layanan dan informasi RT untuk diri atau keluarganya. | Tidak melihat data warga lain, data keuangan internal, audit log, atau konfigurasi RT. |

### 2.1 Variasi Peran Warga

| Jenis | Hak Tambahan |
|---|---|
| **Kepala Keluarga** | Melihat tagihan dan pembayaran keluarga; mengunggah bukti pembayaran; mengajukan koreksi data keluarga; mengelola akses anggota keluarga yang diizinkan. |
| **Anggota Keluarga** | Melihat informasi umum; melihat data diri dan data keluarga sesuai izin; mengajukan surat atau aduan bila diaktifkan. Tidak mengubah data inti keluarga atau mengelola pembayaran keluarga secara default. |

---

## 3. Matriks Hak Akses MVP

**Keterangan:**  
**C** membuat, **R** melihat, **U** mengubah, **A** mengarsipkan/menonaktifkan/membatalkan, **V** memverifikasi/menyetujui, **E** ekspor.  
`R*` berarti hanya data dalam scope sendiri, keluarga sendiri, atau penugasan sendiri.

### 3.1 Sistem, Organisasi, dan Akun

| Modul | Super Admin | Ketua RT | Sekretaris | Bendahara | Pengurus | Warga |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Organisasi RT | C, R, U, A | R | R | R | - | R informasi publik |
| Pengaturan profil RT | R | R, U | R, U | R | R terbatas | R informasi publik |
| Pengguna akun | C, R, U, A teknis | R | C, R, U, A undangan warga | R terbatas | - | R, U akun sendiri |
| Peran dan penugasan pengurus | R | C, R, U, A | - | - | - | - |
| Audit log teknis | R, E | - | - | - | - | - |
| Audit log operasional | - | R, E | R* administrasi | R* keuangan | - | - |

### 3.2 Rumah, Keluarga, dan Warga

| Modul | Super Admin | Ketua RT | Sekretaris | Bendahara | Pengurus | Warga |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Rumah/unit | - | R | C, R, U, A, E | R | R* bila ditugaskan | R* sendiri |
| Keluarga | - | R, E | C, R, U, A, V, E | R | R* data minimum bila ditugaskan | R* sendiri; U pengajuan koreksi |
| Warga | - | R, E | C, R, U, A, V, E | R terbatas | R* data minimum bila ditugaskan | R* diri/keluarga sesuai izin; U pengajuan koreksi |
| Koreksi data warga | - | R | R, U, V | - | - | C, R* |
| Impor CSV | - | R | C, R, V | - | - | - |
| Lookup pendidikan/status perkawinan | - | R | R | R | R | R |
| Ekspor data kependudukan | - | R, E | R, E | - | - | - |

### 3.3 Pengumuman, Agenda, dan Notifikasi

| Modul | Super Admin | Ketua RT | Sekretaris | Bendahara | Pengurus | Warga |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Pengumuman | - | C, R, U, A | C, R, U, A | R | C, R, U, A bila ditugaskan | R* |
| Agenda | - | C, R, U, A | C, R, U, A | R | C, R, U, A bila ditugaskan | R* |
| Notifikasi | - | R* | R* | R* | R* | R, U sendiri |

### 3.4 Iuran, Pembayaran, dan Kas

| Modul | Super Admin | Ketua RT | Sekretaris | Bendahara | Pengurus | Warga |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Jenis iuran | - | R | R | C, R, U, A | - | R* terkait |
| Tagihan | - | R, E | R | C, R, U, A, E | - | R* keluarga |
| Bukti pembayaran | - | R | R terbatas | C, R, V, A | - | C, R* keluarga sesuai izin |
| Tanda terima | - | R | R | C, R | - | R* keluarga |
| Buku kas | - | R, E | R | C, R, U, E | - | R ringkasan yang dipublikasikan |
| Transaksi pembalik | - | R | - | C, R, V | - | - |
| Laporan keuangan | - | R, E | - | R, E | - | R ringkasan yang dipublikasikan |

**Catatan keuangan:**
- `A` pada tagihan atau pembayaran berarti pembatalan beralasan, bukan penghapusan.
- Koreksi kas dilakukan melalui transaksi pembalik.
- Bendahara tidak dapat memverifikasi, menolak, atau membatalkan pembayaran miliknya sendiri.
- Fallback `payment.verify` hanya dapat diberikan Ketua RT atau Sekretaris melalui penugasan terdokumentasi; tetap dilarang memverifikasi transaksi sendiri.

### 3.5 Surat dan Aduan

| Modul | Super Admin | Ketua RT | Sekretaris | Bendahara | Pengurus | Warga |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Jenis/template surat | - | R | C, R, U, A | - | - | - |
| Pengajuan surat | - | R, V | R, U, V proses | - | R* bila ditugaskan | C, R* sendiri |
| Surat terbit/PDF | - | R | C, R | - | - | R, E* sendiri |
| Kategori aduan | - | R | C, R, U, A | - | - | - |
| Aduan | - | R, E | R, U penugasan | R terbatas bila terkait | R, U* bila ditugaskan | C, R, U* milik sendiri |
| Komentar/pembaruan aduan | - | C, R | C, R, U | R terbatas | C, R, U* | C, R* pada aduan sendiri |

**Catatan persuratan:**
- Sekretaris memeriksa kelengkapan serta meminta perbaikan.
- Ketua RT memberi persetujuan akhir sebelum surat diterbitkan.
- Pemohon tidak dapat menyetujui atau menerbitkan suratnya sendiri.

---

## 4. Kode Izin Teknis

Format: `resource.action`. Scope seperti `self`, `assigned`, atau `organization` wajib diperiksa di backend.

### 4.1 Sistem

```text
organization.create
organization.read
organization.update
organization.deactivate
user.invite
user.read
user.update
user.deactivate
role.assign
role.revoke
audit.read
audit.export
```

### 4.2 Kependudukan

```text
house_unit.read
house_unit.create
house_unit.update
house_unit.deactivate
household.read
household.create
household.update
household.deactivate
household.verify
household.export
resident.read
resident.create
resident.update
resident.deactivate
resident.verify
resident.export
resident.read_sensitive
resident.correction.submit
resident.correction.review
```

### 4.3 Komunikasi

```text
announcement.read
announcement.create
announcement.update
announcement.archive
event.read
event.create
event.update
event.cancel
notification.read_self
notification.mark_read_self
```

### 4.4 Keuangan

```text
due_type.read
due_type.create
due_type.update
due_type.deactivate
invoice.read
invoice.create
invoice.update
invoice.cancel
invoice.export
payment.read
payment.submit
payment.verify
payment.reject
payment.cancel
cash.read
cash.create
cash.update
cash.reverse
finance.export
```

### 4.5 Surat dan Aduan

```text
letter_type.read
letter_type.create
letter_type.update
letter_type.deactivate
letter_request.read
letter_request.submit
letter_request.process
letter_request.request_revision
letter_request.approve
letter_request.issue
letter_request.download
complaint.read
complaint.submit
complaint.assign
complaint.update_status
complaint.comment
complaint.export
complaint_category.read
complaint_category.create
complaint_category.update
```

### 4.6 Rencana Izin Pasca-MVP (Epic 16–19)

```text
asset.read
asset.manage
asset.loan
asset.maintain
patrol.read
patrol.manage
patrol.checkin
patrol.incident
activity.read
activity.manage
emergency.alert
visitor.invite
visitor.manage
waste.read
waste.manage
business.read
business.submit
posyandu.read
posyandu.manage
election.read
election.manage
election.vote
```

### 4.7 Rencana Akses Pasca-MVP (Epic 16–19)

*Hak akses berikut bukan bagian MVP. Finalisasi peran, pemisahan tugas, scope, retensi, dan audit dilakukan bersama acceptance criteria tiap epic.*

| Modul | Peran utama | Batas akses |
|---|---|---|
| Aset | Pengurus, Warga | Warga hanya melihat aset yang dapat dipinjam serta pengajuan/riwayat sendiri; pengurus terotorisasi mengelola inventaris, peminjaman, dan pemeliharaan. |
| Siskamling, kerja bakti, panic button, buku tamu | Warga, Petugas keamanan, Pengurus | Warga hanya mengakses jadwal, absensi, laporan, alert, atau undangan miliknya; lokasi darurat dan log tamu dibatasi petugas/pengurus berwenang. |
| Sampah, UMKM, Posyandu | Warga, Bendahara, Pengurus | Retribusi sampah mengikuti otorisasi `due_types`/`invoices`; kontak UMKM hanya yang disetujui pemilik; data Posyandu non-medis dibatasi petugas berwenang. |
| E-voting | Warga, Pengurus pemilihan | Pemilih hanya dapat menggunakan satu hak suara per KK; admin pemilihan tidak dapat melihat pilihan individual; hasil yang dipublikasikan hanya agregat. |

---

## 5. Aturan Implementasi Wajib

1. Semua endpoint memeriksa autentikasi, `organization_id`, permission code, serta scope kepemilikan/penugasan.
2. Hak baca data sensitif tidak diberikan melalui akses daftar biasa. Tindakan membuka nilai utuh memakai `resident.read_sensitive`, alasan akses, dan audit log.
3. Super Admin tidak menerima `resident.read`, `finance.read`, `letter_request.read`, atau `complaint.read` sebagai izin default.
4. Ketua RT tidak dapat menaikkan hak akses dirinya sendiri atau menetapkan peran Super Admin.
5. Pengurus tidak memiliki izin bawaan selain autentikasi, profil sendiri, pengumuman/agenda yang diterbitkan, serta izin eksplisit dari Ketua RT.
6. Perubahan peran, ekspor, pembukaan data sensitif, persetujuan surat, dan tindakan keuangan menghasilkan audit log berisi pelaku, peran aktif, objek, waktu, IP yang diperlakukan sesuai kebijakan privasi, user agent, dan request ID.
7. Akun nonaktif kehilangan seluruh akses baru, tetapi riwayat aktivitasnya tetap tersimpan.
8. Lookup pendidikan dan status perkawinan memakai `resident.read`; keduanya global serta read-only, tanpa permission CRUD terpisah.
9. Kategori aduan tidak dihapus; `complaint_category.update` dapat mengubah status menjadi `inactive`.
10. Izin frontend harus mengikuti izin backend, tetapi backend tetap menjadi sumber keputusan akhir.
11. Epic 16–19 memakai izin rencana pada §4.6 hanya setelah permission di-seed, matriks peran disetujui, serta authorization test tersedia.
12. Alert darurat, lokasi pengguna, log tamu, data Posyandu, dan ballot tidak boleh diekspor atau diakses lintas scope tanpa dasar operasional, permission khusus, dan audit.
