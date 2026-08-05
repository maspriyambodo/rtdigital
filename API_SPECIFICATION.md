# Spesifikasi API RT Digital

**Status:** Draft untuk validasi  
**Versi API:** `v1`  
**Base URL:** `https://api.domain-rt.id/api/v1`  
**Format:** `application/json`  
**Referensi:** `PRD.md`, `SCOPE.md`, `DATABASE_DESIGN.md`, `USER_ROLES_AND_PERMISSIONS.md`, `USER_FLOW.md`

Dokumen ini mendefinisikan kontrak REST API antara frontend Next.js dan Go API. Kontrak final wajib dibuat dalam OpenAPI sebelum implementasi.

---

## 1. Konvensi Global

### 1.1 Autentikasi

Access token dikirim melalui header:

```http
Authorization: Bearer <access_token>
```

Refresh token disimpan dalam cookie `HttpOnly`, `Secure`, dan `SameSite` sesuai desain domain produksi. Backend memeriksa autentikasi, `organization_id`, permission, serta scope kepemilikan atau penugasan pada setiap endpoint privat.

### 1.2 Header Standar

```http
Content-Type: application/json
X-Request-ID: <uuid>
Idempotency-Key: <uuid>
```

- `X-Request-ID` dapat diberikan client, lalu diteruskan pada response dan log.
- `Idempotency-Key` wajib untuk pembuatan pembayaran, transaksi kas, serta operasi lain yang tidak boleh tercatat ganda.

### 1.3 Respons Sukses

```json
{
  "data": {
    "id": "uuid"
  }
}
```

Endpoint daftar memakai cursor pagination:

```json
{
  "data": [],
  "meta": {
    "next_cursor": "cursor_opsional",
    "has_more": false
  }
}
```

Parameter daftar standar:

```text
?limit=20&cursor=<cursor>&search=<kata>&status=<status>&sort=-created_at
```

- `limit` maksimum ditentukan backend.
- Prefix `-` pada `sort` berarti urutan menurun.
- Filter hanya tersedia pada endpoint yang relevan.

### 1.4 Respons Error

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Data yang dikirim tidak valid.",
    "details": [
      {
        "field": "amount",
        "issue": "Nominal harus lebih besar dari 0."
      }
    ],
    "request_id": "req_01..."
  }
}
```

Status HTTP minimum:

| Status | Makna |
|---|---|
| `400` | Request tidak valid |
| `401` | Belum login atau token tidak valid |
| `403` | Tidak memiliki izin atau scope data |
| `404` | Data tidak ditemukan |
| `409` | Konflik status, duplikasi, atau idempotency |
| `422` | Validasi bisnis gagal |
| `429` | Rate limit terlampaui |
| `500` | Kesalahan internal |

---

## 2. Endpoint API

## 2.1 Autentikasi dan Profil

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `POST` | `/auth/login` | Login email/nomor telepon dan kata sandi | Publik |
| `POST` | `/auth/refresh` | Membuat access token baru dari refresh cookie | Publik/cookie valid |
| `POST` | `/auth/logout` | Mencabut sesi perangkat saat ini | Auth |
| `POST` | `/auth/logout-all` | Mencabut seluruh sesi akun | Auth |
| `POST` | `/auth/activate` | Aktivasi akun melalui token undangan | Publik/token valid |
| `POST` | `/auth/forgot-password` | Meminta reset kata sandi | Publik |
| `POST` | `/auth/reset-password` | Reset kata sandi dengan token valid | Publik |
| `POST` | `/auth/mfa/verify` | Verifikasi MFA pengurus saat login | Auth sementara |
| `GET` | `/me` | Profil, organisasi, peran, permission efektif | Auth |
| `PATCH` | `/me` | Ubah profil dasar sendiri | Auth |
| `PATCH` | `/me/password` | Ubah kata sandi sendiri | Auth |

## 2.2 Organisasi, Pengguna, dan Peran

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/organizations/current` | Profil RT aktif | Auth |
| `PATCH` | `/organizations/current` | Ubah pengaturan RT | `organization.update` |
| `GET` | `/users` | Daftar akun pengguna | `user.read` |
| `GET` | `/users/{id}` | Detail akun | `user.read` |
| `POST` | `/users/invite` | Undang warga/pengurus | `user.invite` |
| `PATCH` | `/users/{id}` | Ubah status akun non-kredensial | `user.update` |
| `POST` | `/users/{id}/deactivate` | Nonaktifkan akun | `user.deactivate` |
| `GET` | `/roles` | Daftar peran tersedia | `role.assign` |
| `POST` | `/users/{id}/roles` | Tambahkan peran pengguna | `role.assign` |
| `DELETE` | `/users/{id}/roles/{role_id}` | Cabut peran pengguna | `role.revoke` |
| `POST` | `/office-handovers` | Inisialisasi serah-terima jabatan | `role.assign` |
| `GET` | `/office-handovers/{id}` | Detail serah-terima jabatan | `role.assign` |
| `POST` | `/office-handovers/{id}/complete` | Selesaikan serah-terima; transfer peran, cabut sesi pengurus lama | `role.assign` |

**Aturan:** Ketua RT tidak dapat menaikkan hak akses dirinya sendiri atau menetapkan peran Super Admin.

## 2.3 Rumah, Keluarga, dan Warga

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/house-units` | Daftar rumah/unit | `house_unit.read` |
| `POST` | `/house-units` | Tambah rumah/unit | `house_unit.create` |
| `GET` | `/house-units/{id}` | Detail rumah/unit | `house_unit.read` |
| `PATCH` | `/house-units/{id}` | Ubah rumah/unit | `house_unit.update` |
| `POST` | `/house-units/{id}/deactivate` | Nonaktifkan rumah/unit | `house_unit.deactivate` |
| `GET` | `/households` | Daftar keluarga | `household.read` |
| `POST` | `/households` | Tambah keluarga | `household.create` |
| `GET` | `/households/health-scores` | Daftar kerja kualitas data keluarga | `resident.read` |
| `GET` | `/households/{id}` | Detail keluarga dan anggota | `household.read` + scope |
| `PATCH` | `/households/{id}` | Ubah keluarga | `household.update` + scope |
| `POST` | `/households/{id}/verify` | Verifikasi keluarga | `household.verify` |
| `GET` | `/residents` | Daftar warga; nilai sensitif dimasking | `resident.read` |
| `POST` | `/residents` | Tambah warga | `resident.create` |
| `GET` | `/residents/{id}` | Detail warga; masking aktif | `resident.read` + scope |
| `PATCH` | `/residents/{id}` | Ubah warga | `resident.update` |
| `POST` | `/residents/{id}/verify` | Verifikasi data warga | `resident.verify` |
| `POST` | `/residents/corrections` | Ajukan koreksi data | `resident.correction.submit` + scope diri/keluarga |
| `GET` | `/resident-corrections` | Daftar usulan koreksi | `resident.correction.review` |
| `GET` | `/resident-corrections/{id}` | Detail usulan koreksi | `resident.correction.review` + scope |
| `POST` | `/resident-corrections/{id}/approve` | Setujui koreksi | `resident.correction.review` |
| `POST` | `/resident-corrections/{id}/reject` | Tolak koreksi; alasan wajib | `resident.correction.review` |
| `POST` | `/resident-corrections/{id}/request-revision` | Minta perbaikan koreksi | `resident.correction.review` |
| `POST` | `/residents/import` | Validasi/impor CSV warga | `resident.create` |
| `GET` | `/reports/residents` | Ekspor laporan warga CSV/PDF | `resident.export` |
| `GET` | `/reports/mutations` | Laporan mutasi warga CSV/PDF | `resident.export` |
| `GET` | `/reports/households` | Laporan keluarga CSV/PDF | `household.export` |

## 2.4 Pengumuman, Agenda, dan Notifikasi

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/announcements` | Pengumuman sesuai target pengguna | `announcement.read` |
| `POST` | `/announcements` | Buat pengumuman draft/terjadwal | `announcement.create` |
| `GET` | `/announcements/{id}` | Detail pengumuman | `announcement.read` + scope |
| `PATCH` | `/announcements/{id}` | Ubah pengumuman | `announcement.update` |
| `POST` | `/announcements/{id}/publish` | Terbitkan pengumuman | `announcement.update` |
| `POST` | `/announcements/{id}/archive` | Arsipkan pengumuman | `announcement.archive` |
| `GET` | `/events` | Daftar agenda | `event.read` |
| `POST` | `/events` | Buat agenda | `event.create` |
| `GET` | `/events/{id}` | Detail agenda | `event.read` |
| `PATCH` | `/events/{id}` | Ubah agenda | `event.update` |
| `POST` | `/events/{id}/cancel` | Batalkan agenda | `event.cancel` |
| `GET` | `/notifications` | Notifikasi akun sendiri | Auth |
| `PATCH` | `/notifications/{id}/read` | Tandai notifikasi dibaca | Auth + pemilik |
| `POST` | `/notifications/read-all` | Tandai seluruh notifikasi dibaca | Auth |

## 2.5 Iuran, Tagihan, Pembayaran, dan Kas

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/due-types` | Daftar jenis iuran | `due_type.read` |
| `POST` | `/due-types` | Buat jenis iuran | `due_type.create` |
| `PATCH` | `/due-types/{id}` | Ubah jenis iuran | `due_type.update` |
| `POST` | `/due-types/{id}/deactivate` | Nonaktifkan jenis iuran | `due_type.deactivate` |
| `GET` | `/invoices` | Daftar tagihan sesuai scope | `invoice.read` + scope |
| `POST` | `/invoices` | Buat tagihan individual | `invoice.create` |
| `POST` | `/invoices/generate` | Buat tagihan massal | `invoice.create` |
| `GET` | `/invoices/{id}` | Detail tagihan dan riwayat pembayaran | `invoice.read` + scope |
| `PATCH` | `/invoices/{id}` | Ubah tagihan sebelum pembayaran tervalidasi | `invoice.update` |
| `POST` | `/invoices/{id}/cancel` | Batalkan tagihan; alasan wajib | `invoice.cancel` |
| `GET` | `/payments` | Daftar pembayaran sesuai scope | `payment.read` + scope |
| `POST` | `/payments` | Kirim pembayaran transfer atau catat tunai | `payment.submit` |
| `GET` | `/payments/{id}` | Detail pembayaran | `payment.read` + scope |
| `POST` | `/payments/{id}/verify` | Verifikasi pembayaran | `payment.verify` |
| `POST` | `/payments/{id}/reject` | Tolak pembayaran; alasan wajib | `payment.reject` |
| `POST` | `/payments/{id}/cancel` | Batalkan pembayaran tervalidasi; alasan wajib | `payment.cancel` |
| `GET` | `/cash-categories` | Daftar kategori kas | `cash.read` |
| `POST` | `/cash-categories` | Buat kategori kas | `cash.create` |
| `GET` | `/cash-transactions` | Riwayat buku kas | `cash.read` |
| `POST` | `/cash-transactions` | Catat kas manual non-iuran | `cash.create` |
| `GET` | `/cash-transactions/{id}` | Detail transaksi kas | `cash.read` |
| `POST` | `/cash-transactions/{id}/reverse` | Buat transaksi pembalik | `cash.reverse` |
| `GET` | `/reports/invoices` | Laporan tagihan CSV/PDF | `invoice.export` |
| `GET` | `/reports/payments` | Laporan pembayaran CSV/PDF | `finance.export` |
| `GET` | `/reports/arrears` | Laporan tunggakan CSV/PDF | `finance.export` |
| `GET` | `/reports/cash` | Laporan buku kas CSV/PDF | `finance.export` |

**Aturan:**
- `POST /payments` dan `POST /cash-transactions` wajib memakai `Idempotency-Key`.
- Pembayaran transfer harus mempunyai `proof_file_id`.
- `verified_by` tidak boleh sama dengan `created_by`.
- Pembatalan bukan penghapusan data.
- Verifikasi pembayaran membuat transaksi pemasukan kas atomik.

## 2.6 Persuratan

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/letter-types` | Daftar jenis surat aktif | `letter_type.read` |
| `POST` | `/letter-types` | Buat jenis/template surat | `letter_type.create` |
| `PATCH` | `/letter-types/{id}` | Ubah jenis/template surat | `letter_type.update` |
| `POST` | `/letter-types/{id}/deactivate` | Nonaktifkan jenis surat | `letter_type.deactivate` |
| `GET` | `/letter-requests` | Daftar pengajuan sesuai scope | `letter_request.read` + scope |
| `POST` | `/letter-requests` | Ajukan surat baru | `letter_request.submit` |
| `GET` | `/letter-requests/{id}` | Detail pengajuan surat | `letter_request.read` + scope |
| `PATCH` | `/letter-requests/{id}` | Ubah draft atau perbaikan pemohon | `letter_request.submit` + scope |
| `POST` | `/letter-requests/{id}/request-revision` | Minta perbaikan pemohon | `letter_request.request_revision` |
| `POST` | `/letter-requests/{id}/process` | Lanjutkan pemeriksaan/draft surat | `letter_request.process` |
| `POST` | `/letter-requests/{id}/approve` | Persetujuan akhir Ketua RT | `letter_request.approve` |
| `POST` | `/letter-requests/{id}/reject` | Tolak pengajuan; alasan wajib | `letter_request.process` |
| `POST` | `/letter-requests/{id}/issue` | Terbitkan nomor dan PDF surat | `letter_request.issue` |
| `GET` | `/letter-requests/{id}/download` | URL download PDF surat | `letter_request.download` + scope |
| `GET` | `/letters/verify/{code}` | Verifikasi publik nomor, jenis, tanggal terbit, status surat; tanpa data pribadi | Publik |

## 2.7 Aduan

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/complaint-categories` | Daftar kategori aduan per organisasi | `complaint_category.read` |
| `POST` | `/complaint-categories` | Buat kategori aduan baru | `complaint_category.create` |
| `GET` | `/complaint-categories/{id}` | Detail kategori aduan | `complaint_category.read` |
| `PATCH` | `/complaint-categories/{id}` | Ubah nama, kode, atau status kategori aduan | `complaint_category.update` |
| `GET` | `/complaints` | Daftar aduan sesuai scope; filter `complaint_category_id` tersedia | `complaint.read` + scope |
| `POST` | `/complaints` | Buat aduan dengan `complaint_category_id` | `complaint.submit` |
| `GET` | `/complaints/{id}` | Detail aduan dan komentar | `complaint.read` + scope |
| `PATCH` | `/complaints/{id}` | Ubah aduan milik sendiri sebelum diproses | `complaint.submit` + scope |
| `POST` | `/complaints/{id}/assign` | Tetapkan penanggung jawab | `complaint.assign` |
| `POST` | `/complaints/{id}/status` | Ubah status aduan | `complaint.update_status` + scope |
| `POST` | `/complaints/{id}/comments` | Tambah komentar atau perkembangan | `complaint.comment` + scope |
| `GET` | `/reports/letters` | Laporan pengajuan surat CSV/PDF | `letter_request.export` |
| `GET` | `/reports/complaints` | Rekap aduan CSV/PDF, termasuk kategori master | `complaint.export` |

## 2.8 Master Data Global

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `GET` | `/education-levels` | Daftar tingkat pendidikan standar read-only | `resident.read` |
| `GET` | `/marital-statuses` | Daftar status perkawinan standar read-only | `resident.read` |

## 2.9 File, Dashboard, dan Audit

| Method | Endpoint | Keterangan | Otorisasi |
|---|---|---|---|
| `POST` | `/files/presign-upload` | Minta URL upload Cloudflare R2 S3-compatible | Auth + validasi scope |
| `POST` | `/files/confirm-upload` | Konfirmasi unggah Cloudflare R2 selesai | Auth + pemilik unggahan |
| `GET` | `/files/{id}/download` | Minta URL download Cloudflare R2 sementara | Auth + scope |
| `GET` | `/dashboard/admin` | Ringkasan operasional pengurus | Pengurus berwenang |
| `GET` | `/dashboard/resident` | Ringkasan tagihan, surat, aduan warga | Warga |
| `GET` | `/audit-logs` | Daftar audit log operasional | `audit.read` |
| `GET` | `/audit-logs/{id}` | Detail audit log | `audit.read` |

---

## 3. Contoh Payload Inti

### 3.1 `POST /payments`

Header:

```http
Idempotency-Key: 2f7fe63a-341a-46dd-8fb6-c32466501011
```

Request:

```json
{
  "invoice_id": "5ac7fe90-5acd-4878-a2fb-a61f206e6d0b",
  "method": "transfer",
  "amount": 150000.00,
  "paid_at": "2026-08-01T10:00:00Z",
  "proof_file_id": "e6832489-8928-49b5-a8f8-8483a99195e0"
}
```

Response:

```json
{
  "data": {
    "id": "b657d74c-5916-4214-983b-b365af4c87fa",
    "payment_number": "PAY-2608-0001",
    "verification_status": "pending",
    "invoice_status": "pending_verification",
    "created_at": "2026-08-01T10:05:00Z"
  }
}
```

### 3.2 `POST /payments/{id}/verify`

Request:

```json
{
  "note": "Sesuai mutasi rekening."
}
```

Response:

```json
{
  "data": {
    "id": "b657d74c-5916-4214-983b-b365af4c87fa",
    "verification_status": "verified",
    "verified_at": "2026-08-02T09:00:00Z",
    "invoice_status": "paid",
    "cash_transaction_id": "8fea9e3c-4f2d-4344-92fd-5ef568091b7b"
  }
}
```

### 3.3 `POST /letter-requests`

```json
{
  "letter_type_id": "4adf9dfd-4ebf-426a-92e5-e009830e0e3d",
  "resident_id": "66fdf983-7032-4e1b-b3af-8f62c2d2830e",
  "form_data": {
    "purpose": "Pembuatan KTP Baru"
  },
  "attachment_file_ids": [
    "3f01e5d0-0f9f-47b4-9adc-a41e5e6a014b"
  ]
}
```

### 3.4 `POST /files/presign-upload`

```json
{
  "entity_type": "payment",
  "entity_id": "5ac7fe90-5acd-4878-a2fb-a61f206e6d0b",
  "purpose": "payment_proof",
  "original_name": "bukti-transfer.jpg",
  "mime_type": "image/jpeg",
  "size_bytes": 248321
}
```

Response:

```json
{
  "data": {
    "file_id": "e6832489-8928-49b5-a8f8-8483a99195e0",
    "upload_url": "https://<account_id>.r2.cloudflarestorage.com/...",
    "upload_headers": {
      "Content-Type": "image/jpeg"
    },
    "expires_at": "2026-08-01T10:10:00Z"
  }
}
```

---

## 4. Ketentuan Frontend

1. Gunakan `fetch` native atau wrapper tipis. Jangan tambahkan library HTTP tanpa kebutuhan nyata.
2. Saat menerima `401`, frontend boleh mencoba satu kali `POST /auth/refresh`; bila gagal, hapus state privat lalu arahkan ke login.
3. Upload file selalu langsung browser → Cloudflare R2 memakai `upload_url`; file biner tidak melewati Go API.
4. Simpan draft formulir panjang secara lokal dengan aman. Jangan menyimpan token, NIK, nomor KK, atau dokumen dalam local storage.
5. Tampilkan `message` error API kepada pengguna; catat `request_id` untuk bantuan pengurus.
6. Semua halaman atau response data privat menggunakan kebijakan cache `private, no-store`.

---

## 5. Keamanan API

- CORS hanya mengizinkan origin frontend resmi.
- Endpoint login, reset password, presign upload, dan endpoint sensitif memakai rate limit.
- Validasi schema, tipe, ukuran, dan kepemilikan data dilakukan backend.
- Endpoint daftar tidak mengirim data sensitif utuh secara default.
- Pembukaan data sensitif memerlukan `resident.read_sensitive`, alasan akses, serta audit log.
- Token, password hash, NIK, nomor KK, bukti transfer, dan dokumen pribadi tidak boleh muncul pada log atau response yang tidak berwenang.
- Semua perubahan peran, ekspor, tindakan keuangan, persetujuan surat, dan akses sensitif dicatat pada audit log.