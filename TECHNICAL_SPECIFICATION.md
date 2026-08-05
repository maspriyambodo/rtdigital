# Spesifikasi Teknis RT Digital

**Status:** Draft untuk validasi  
**Cakupan:** MVP RT Digital  
**Referensi:** `PRD.md`, `SCOPE.md`, `DATABASE_DESIGN.md`, `SYSTEM_ARCHITECTURE.md`, `API_SPECIFICATION.md`, `USER_ROLES_AND_PERMISSIONS.md`, `UI_COMPONENTS.md`

Dokumen ini menetapkan standar implementasi frontend Next.js dan Go API. Backend menjadi sumber keputusan untuk autentikasi, otorisasi, validasi, transaksi, dan audit.

---

## 1. Keputusan Teknis MVP

| Area | Keputusan |
|---|---|
| Frontend | Next.js App Router, TypeScript strict, Cloudflare Workers melalui OpenNext |
| Backend | Go modular monolith di Amazon ECS Fargate |
| HTTP Go | `net/http` dan `ServeMux` standar Go |
| Database | PostgreSQL melalui `pgx`/`pgxpool`; SQL eksplisit dengan parameter binding |
| Migration | SQL migration versioned; dijalankan sebagai ECS task terpisah sebelum deploy |
| Styling | CSS custom properties/design tokens; Tailwind CSS opsional sebagai implementasi |
| State frontend | Server Components default; state lokal React/Context; tanpa Redux/Zustand MVP |
| Form | HTML native dan state lokal; `react-hook-form` hanya bila formulir kompleks terbukti sulit dikelola |
| Cache/session | PostgreSQL untuk sesi/refresh token; tanpa Redis MVP |
| File | Cloudflare R2 private bucket melalui API S3-compatible |
| Log | `log/slog` JSON ke `stdout`, diteruskan ke CloudWatch |
| API | REST JSON `/api/v1`, cursor pagination, format respons/error seragam |

Jangan menambah framework, ORM, state manager, atau library UI sebelum kebutuhan yang tidak dapat dipenuhi solusi standar terbukti.

---

## 2. Struktur Repository

```text
/
├── apps/
│   └── web/                         # Next.js
├── services/
│   └── api/                         # Go API
├── packages/
│   └── api-client/                  # Opsional: generated/typed client dari OpenAPI
├── infrastructure/
│   ├── docker/
│   ├── aws/
│   └── cloudflare/
├── docs/
│   ├── API_SPECIFICATION.md
│   ├── DATABASE_DESIGN.md
│   ├── SYSTEM_ARCHITECTURE.md
│   └── TECHNICAL_SPECIFICATION.md
├── docker-compose.yml
├── .env.example
└── README.md
```

`packages/api-client` hanya dibuat setelah OpenAPI stabil dan duplikasi tipe/request mulai nyata.

---

## 3. Struktur Go API

## 3.1 Direktori

```text
services/api/
├── cmd/
│   └── server/
│       └── main.go                  # Bootstrap, DI, graceful shutdown
├── internal/
│   ├── app/
│   │   ├── config/                  # Environment dan validasi konfigurasi
│   │   ├── auth/                    # Password, JWT, session, MFA
│   │   ├── database/                # pgxpool, transaction helper
│   │   ├── httpx/                   # Response, request ID, error mapper
│   │   ├── logging/                 # slog setup dan sanitasi log
│   │   └── middleware/              # Recover, CORS, auth, RBAC, rate limit
│   ├── modules/
│   │   ├── identity/                # User, role, permission, session
│   │   ├── resident/                # Rumah, keluarga, warga, koreksi, lookup global
│   │   ├── communication/           # Pengumuman, agenda, notifikasi
│   │   ├── finance/                 # Iuran, tagihan, payment, kas
│   │   ├── letter/                  # Jenis surat dan pengajuan
│   │   ├── complaint/               # Aduan, komentar, kategori aduan
│   │   ├── file/                    # Metadata file dan signed URL
│   │   ├── audit/                   # Audit log
│   │   ├── asset/                   # (Pasca-MVP) Aset, peminjaman, pemeliharaan
│   │   ├── security_ops/            # (Pasca-MVP) Ronda, kerja bakti, panic button, buku tamu
│   │   ├── local_services/          # (Pasca-MVP) Sampah, UMKM, Posyandu non-medis
│   │   └── governance/              # (Pasca-MVP) E-voting pengurus
│   └── infrastructure/
│       ├── s3/                      # Cloudflare R2 via AWS SDK S3-compatible
│       ├── email/                   # Resend client
│       └── whatsapp/                # SaungWA client
├── migrations/                      # SQL up/down migration
├── go.mod
└── go.sum
```

## 3.2 Struktur Modul

```text
internal/modules/finance/
├── handler.go                       # HTTP parsing dan response
├── service.go                       # Aturan bisnis dan transaksi
├── repository.go                    # Query PostgreSQL
├── types.go                         # DTO, model domain, filter
└── service_test.go                  # Unit test aturan bisnis
```

| Lapisan | Tanggung jawab | Dilarang |
|---|---|---|
| Handler | Parse request, validasi struktur dasar, panggil service, tulis JSON | Query DB, aturan bisnis |
| Service | Otorisasi bisnis, transaksi, state transition, audit, orchestration | Parsing HTTP |
| Repository | Query parameterized, scan data, lock row bila perlu | Aturan bisnis |
| Infrastructure | Cloudflare R2, Resend, SaungWA, external API | Menentukan hak akses |
| Middleware | Request ID, recover, CORS, auth, logging | Aturan modul bisnis |

## 3.3 Aturan Go

- Dependency injection manual melalui constructor dari `cmd/server/main.go`.
- Tidak ada global mutable state atau `init()` untuk konfigurasi.
- Teruskan `context.Context` dari handler sampai repository dan external client.
- Semua request memiliki timeout; outbound request memakai timeout eksplisit.
- Repository menggunakan `pgxpool.Pool`; query selalu parameterized.
- Transaksi dimulai di service untuk use case atomik, terutama payment verification, kas, status surat, dan audit.
- Gunakan `SELECT ... FOR UPDATE` untuk perubahan state yang berisiko race condition.
- `panic` hanya untuk kondisi tidak dapat dipulihkan; middleware recovery mengembalikan error generik `500`.
- Graceful shutdown menghentikan penerimaan request baru, menunggu request aktif dalam batas timeout, lalu menutup pool.

---

## 4. Struktur Next.js

## 4.1 Direktori

```text
apps/web/
├── src/
│   ├── app/
│   │   ├── (auth)/                  # Login, aktivasi, reset password
│   │   ├── app/                     # Area Warga
│   │   ├── admin/                   # Area Pengurus
│   │   ├── layout.tsx
│   │   ├── error.tsx
│   │   └── not-found.tsx
│   ├── components/
│   │   ├── ui/                      # Dari UI_COMPONENTS.md
│   │   ├── layout/
│   │   └── feedback/
│   ├── features/
│   │   ├── finance/
│   │   ├── resident/
│   │   ├── letter/
│   │   ├── complaint/
│   │   ├── communication/
│   │   ├── asset/                   # Dibuat saat Epic 16 dimulai
│   │   ├── security-ops/            # Dibuat saat Epic 17 dimulai
│   │   ├── local-services/          # Dibuat saat Epic 18 dimulai
│   │   └── governance/              # Dibuat saat Epic 19 dimulai
│   ├── lib/
│   │   ├── api.ts                   # Wrapper fetch tipis
│   │   ├── auth.ts                  # Token memory dan refresh satu kali
│   │   ├── format.ts                # Rupiah, tanggal, masking
│   │   └── errors.ts                # API error type
│   ├── types/
│   └── styles/
│       └── globals.css
├── public/
│   ├── manifest.webmanifest
│   └── icons/
├── next.config.mjs
└── tsconfig.json
```

## 4.2 Aturan Rendering dan Data

- Server Components adalah default.
- `"use client"` hanya untuk form, upload, dialog, navigasi interaktif, browser API, dan state UI.
- Halaman privat tidak dicache: `Cache-Control: private, no-store`.
- Data awal halaman privat dapat diambil Server Component bila autentikasi server-side tersedia; bila token hanya ada di memory browser, gunakan client fetch setelah halaman shell dirender.
- Mutasi selalu memakai `fetch` wrapper tipis, loading state, error terstruktur, dan idempotency key bila diperlukan.
- Jangan membuat API publik kedua di Next.js untuk meneruskan seluruh request ke Go API.
- Next.js Route Handler hanya digunakan untuk kebutuhan frontend internal yang tidak dapat dilakukan browser secara aman.
- Form panjang boleh menyimpan draft non-sensitif secara lokal. Jangan simpan access token, refresh token, NIK, nomor KK, dokumen, atau signed URL di `localStorage`.

---

## 5. Authentication

## 5.1 Token dan Session

| Artefak | Bentuk | Penyimpanan | Masa berlaku |
|---|---|---|---|
| Access token | JWT signed | Memory browser | 15 menit |
| Refresh token | Random opaque token, bukan JWT wajib | Cookie `HttpOnly`, `Secure`, `SameSite=Lax` | 7 hari, dapat diperpanjang terkontrol |
| Session record | Hash refresh token dan metadata | PostgreSQL | Mengikuti refresh token |

Refresh token tidak disimpan mentah di database. Simpan hash aman; token mentah hanya berada dalam cookie.

JWT minimum memuat:

```json
{
  "sub": "user_uuid",
  "org": "organization_uuid",
  "sid": "session_uuid",
  "exp": 1785600900
}
```

Jangan memasukkan daftar permission lengkap ke access token. Permission/peran dapat berubah sebelum token kedaluwarsa; backend memuat permission efektif dari database atau cache aman per request/interval pendek bila kelak terbukti perlu.

## 5.2 Login

1. Backend memvalidasi email/telepon dan kata sandi.
2. Backend memeriksa status akun, lockout, serta MFA jika peran pengurus.
3. Backend membuat session PostgreSQL berisi hash refresh token, waktu kedaluwarsa, user agent, dan IP sesuai kebijakan privasi.
4. Backend mengirim access token melalui JSON.
5. Backend mengirim refresh token melalui `Set-Cookie`.
6. Backend mencatat login penting pada audit log.

Cookie production:

```http
Set-Cookie: rt_refresh=<token>; HttpOnly; Secure; SameSite=Lax; Path=/api/v1/auth
```

Gunakan domain host-only bila frontend dan API memakai subdomain berbeda; browser mengirim cookie hanya ke `api.domain-rt.id`.

## 5.3 Refresh, Logout, dan CSRF

- `/auth/refresh` memvalidasi cookie, session, expiry, dan hash token.
- Refresh token dirotasi saat dipakai; token lama segera dicabut.
- `/auth/logout` mencabut session aktif dan menghapus cookie.
- `/auth/logout-all` mencabut seluruh session akun.
- Endpoint yang memakai cookie refresh harus memvalidasi `Origin` terhadap origin frontend resmi. Terapkan proteksi CSRF tambahan bila endpoint cookie-authenticated bertambah.
- Endpoint bisnis memakai Bearer access token; jangan mengandalkan cookie refresh untuk otorisasi aksi bisnis.

---

## 6. Authorization

## 6.1 Urutan Wajib

Setiap endpoint privat memeriksa:

1. Access token valid dan belum kedaluwarsa.
2. User serta organization masih aktif.
3. Permission code yang dibutuhkan.
4. `organization_id` objek sama dengan organization session. Lookup global read-only seperti `education_levels` dan `marital_statuses` tidak memiliki scope tenant.
5. Scope kepemilikan/penugasan.
6. Aturan pemisahan tugas dan transisi status.

Contoh pembayaran:

```text
payment.verify
+ payment.organization_id = session.organization_id
+ payment.created_by != session.user_id
+ payment.verification_status = pending
```

## 6.2 Implementasi

- Middleware mengautentikasi access token dan menyimpan principal pada `context.Context`.
- Handler mendeklarasikan permission endpoint.
- Service melakukan pemeriksaan ownership, penugasan, state transition, serta separation of duties.
- Semua query tenant memakai `WHERE organization_id = $1`.
- Frontend menyembunyikan menu tanpa permission, tetapi tidak pernah menjadi kontrol keamanan.

## 6.3 Data Sensitif

- Daftar dan detail biasa mengirim nilai yang dimasking.
- Nilai NIK/KK utuh memerlukan `resident.read_sensitive`.
- Akses utuh memerlukan alasan akses, permission, scope, serta audit log.
- Password hash, refresh token, secret, dan data MFA tidak pernah dikembalikan API.

---

## 7. Validation

## 7.1 Frontend

- Gunakan validasi native HTML untuk wajib isi, tipe, panjang, dan batas angka.
- Tambahkan validasi UX untuk nominal, tanggal, file, dan field bergantung.
- Tampilkan error berbahasa Indonesia per field.
- Frontend tidak boleh dianggap sumber validasi keamanan.

## 7.2 Backend

Validasi berlangsung dalam tiga tahap:

| Tahap | Contoh |
|---|---|
| Struktur | JSON valid, UUID, field wajib, string length, format email |
| Semantik | Status yang diizinkan, nominal lebih dari nol, tanggal valid |
| Bisnis | Tagihan milik keluarga, transfer memiliki bukti, payment belum diverifikasi, nomor surat unik |

- Tolak field JSON yang tidak dikenal untuk endpoint mutasi, kecuali field ekspansibel yang didokumentasikan.
- Batasi ukuran body JSON dan multipart request sebelum decode.
- Validasi file dilakukan sebelum presign dan setelah confirm upload.
- Error validasi tidak boleh membocorkan data milik pengguna lain.

---

## 8. Error Handling dan Response API

## 8.1 Respons Sukses

```json
{
  "data": {
    "id": "uuid"
  }
}
```

Respons daftar:

```json
{
  "data": [],
  "meta": {
    "next_cursor": "opaque_cursor",
    "has_more": false
  }
}
```

## 8.2 Respons Error

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Data yang dikirim tidak valid.",
    "details": [
      {
        "field": "amount",
        "issue": "Nominal harus lebih besar dari Rp0."
      }
    ],
    "request_id": "req_01..."
  }
}
```

| HTTP | Kode contoh | Penggunaan |
|---:|---|---|
| 400 | `BAD_REQUEST` | JSON/header/query tidak valid |
| 401 | `UNAUTHORIZED` | Tidak login, token tidak valid, sesi dicabut |
| 403 | `FORBIDDEN` | Permission/scope tidak cukup |
| 404 | `NOT_FOUND` | Objek tidak ada atau tidak boleh diungkapkan |
| 409 | `CONFLICT`, `PAYMENT_ALREADY_VERIFIED` | Duplikasi, idempotency, konflik state |
| 422 | `VALIDATION_FAILED` | Validasi aturan bisnis gagal |
| 429 | `RATE_LIMITED` | Batas request |
| 500 | `INTERNAL_ERROR` | Kesalahan tak terduga |

- Gunakan tipe `AppError` untuk status, code, pesan pengguna, detail aman, serta cause internal.
- Jangan kirim SQL error, stack trace, token, secret, atau detail infrastruktur ke client.
- Error tak terduga dicatat `ERROR` dengan request ID; client hanya menerima `INTERNAL_ERROR`.

---

## 9. Logging, Audit, dan Observabilitas

## 9.1 Application Log

Gunakan `log/slog` JSON.

Atribut minimum:

```json
{
  "time": "2026-08-01T12:00:00Z",
  "level": "INFO",
  "msg": "request completed",
  "request_id": "req_01...",
  "method": "POST",
  "route": "/api/v1/payments/{id}/verify",
  "status": 200,
  "latency_ms": 45,
  "user_id": "uuid",
  "organization_id": "uuid"
}
```

Aturan:

- `INFO`: request sukses dan event operasional normal.
- `WARN`: `4xx`, retry, kondisi tidak normal yang dapat dipulihkan.
- `ERROR`: `5xx`, dependency gagal, panic.
- Log tidak memuat password, raw token, cookie, NIK, KK, nomor rekening, alamat detail, isi dokumen, bukti transfer, atau signed URL.
- Query parameter/header disanitasi sebelum logging.
- Gunakan `X-Request-ID` valid dari client atau buat ID baru; selalu kembalikan pada response.

## 9.2 Audit Log

Audit log berbeda dari application log.

Audit wajib untuk:

- Login gagal signifikan dan login berhasil.
- Perubahan akun, peran, dan permission.
- Akses data sensitif.
- Perubahan warga/keluarga.
- Pembuatan, verifikasi, penolakan, atau pembatalan pembayaran.
- Perubahan kas.
- Persetujuan/penerbitan surat.
- Ekspor laporan.
- Perubahan konfigurasi RT.

Audit bersifat append-only. Simpan nilai sebelum/sesudah yang sudah disanitasi.

---

## 10. Pagination, Search, dan Filter

- Endpoint daftar besar menggunakan cursor pagination.
- Sorting default: `created_at DESC, id DESC`.
- Cursor adalah opaque Base64URL dari pasangan nilai sort dan ID; cursor harus divalidasi sebelum dipakai.
- Query keyset pagination:

```sql
WHERE organization_id = $1
  AND (created_at, id) < ($2, $3)
ORDER BY created_at DESC, id DESC
LIMIT $4;
```

- `limit` default `20`; maksimum `100`.
- Filter dan field sort memakai allowlist server-side; jangan memasukkan nama kolom langsung dari query parameter.
- Daftar referensi kecil dapat memakai limit tetap atau pagination halaman sederhana bila benar-benar lebih sederhana.
- Pencarian nama warga menggunakan parameter binding, batas panjang query, dan index yang sesuai.

---

## 11. Upload dan Download File

## 11.1 Upload

1. Browser memilih file.
2. Frontend memvalidasi tipe, ekstensi, ukuran, dan jumlah file untuk UX.
3. Frontend meminta `POST /files/presign-upload`.
4. Backend memeriksa authentication, permission, ownership `entity_id`, `entity_type`, purpose, MIME allowlist, dan ukuran.
5. Backend membuat record upload sementara dan pre-signed `PUT` URL S3-compatible untuk R2 berumur maksimal 5 menit.
6. Browser mengunggah langsung ke Cloudflare R2.
7. Frontend memanggil `POST /files/confirm-upload`.
8. Backend menjalankan `HeadObject`, memverifikasi file ada, ukuran, metadata, lalu menandai record siap dipakai.
9. Endpoint bisnis hanya menerima `file_id` yang ready, dimiliki organisasi sama, dan sesuai purpose.

## 11.2 Download

1. Browser meminta download ke API.
2. Backend memeriksa permission dan scope objek.
3. Backend menerbitkan pre-signed `GET` URL pendek.
4. Browser mengunduh langsung dari Cloudflare R2.

## 11.3 Aturan

- File biner tidak pernah melewati Go API.
- Bucket Cloudflare R2 dibuat private; public bucket/domain tidak digunakan untuk file privat.
- `storage_key` dibuat server-side, tidak dari `original_name`.
- Validasi MIME client tidak cukup; backend memvalidasi deklarasi dan metadata. Malware scan ditunda sesuai `SCOPE.md`.
- File belum dikonfirmasi dibersihkan oleh lifecycle rule/job terkontrol.
- Signed URL tidak dicatat dalam log, analytics, atau audit detail.

---

## 12. Idempotency dan Transaksi

## 12.1 Idempotency

Wajib untuk:

- `POST /payments`
- `POST /cash-transactions`
- `POST /invoices/generate`
- Endpoint lain yang berpotensi diproses dua kali karena retry jaringan.

Aturan:

1. Frontend membuat UUID baru sekali untuk satu aksi pengguna.
2. Client mengirim `Idempotency-Key`.
3. Backend menyimpan key bersama `organization_id`, user, request hash, status, serta response aman.
4. Request ulang dengan key dan payload sama mengembalikan response awal.
5. Request ulang dengan key sama dan payload berbeda ditolak `409 IDEMPOTENCY_KEY_REUSED`.
6. Unique constraint database mencegah race condition.

## 12.2 Transaksi Atomik

Verifikasi pembayaran harus atomik:

1. Lock payment dan invoice.
2. Validasi status serta pemisahan tugas.
3. Tandai payment `verified`.
4. Perbarui `paid_amount` dan status invoice.
5. Buat `cash_transactions` pemasukan.
6. Buat audit log serta notification record.
7. Commit.

Pengiriman email melalui Resend dan notifikasi WhatsApp melalui SaungWA terjadi asinkron setelah transaksi utama di-commit. Kegagalan notifikasi dicatat tanpa membatalkan transaksi keuangan; retry ditambahkan bila diperlukan.

---

## 13. Database Migration

- Migration memakai SQL versioned `up`/`down`.
- Migration dijalankan sekali sebagai ECS task sebelum API baru menerima traffic.
- API container tidak menjalankan auto-migrate saat startup.
- Migration additive lebih dahulu: tambah tabel/kolom/index, deploy kode kompatibel, migrasikan data, baru hapus kolom lama pada rilis berikutnya.
- Migration destruktif memerlukan snapshot RDS, rehearsal staging, review, serta rollback procedure.
- Untuk Epic 13: migration `0016` menambah master data dan memigrasikan nilai legacy; `0017` menolak penghapusan kolom teks bila masih ada nilai warga yang belum dipetakan ke lookup global.
- Jangan mengandalkan migration down untuk memulihkan data yang sudah berubah; gunakan backup/restore atau migration kompensasi.

---

## 14. Waktu dan Zona Waktu

- PostgreSQL menyimpan waktu sebagai `TIMESTAMPTZ` dalam UTC.
- Go memakai `time.Time` UTC untuk event bertimestamp.
- API mengirim ISO 8601 UTC: `2026-08-01T10:00:00Z`.
- Tanggal bisnis tanpa jam memakai PostgreSQL `DATE`.
- Frontend menampilkan waktu dalam `organization.timezone`, default `Asia/Jakarta`.
- Batas periode tagihan, laporan, dan kas dihitung menggunakan zona waktu organisasi sebelum dikonversi ke UTC bila dibutuhkan.

---

## 15. Checklist Implementasi

- [ ] Struktur Go modular monolith dan Next.js App Router diterapkan.
- [ ] Konfigurasi tervalidasi saat startup; secret tidak ada di repository.
- [ ] Semua endpoint privat memeriksa token, user aktif, permission, tenant, scope, dan aturan bisnis.
- [ ] Password memakai Argon2id; refresh token di-hash dan dirotasi.
- [ ] Semua mutasi memakai validasi struktur, semantik, dan bisnis.
- [ ] Error API mengikuti format standar tanpa detail internal.
- [ ] Application log JSON serta audit log terpisah dan tersanitasi.
- [ ] Endpoint daftar memakai cursor/keyset pagination.
- [ ] Upload/download memakai Cloudflare R2 signed URL S3-compatible dan pemeriksaan permission.
- [ ] Payment, kas, dan pembuatan massal dilindungi idempotency.
- [ ] Modul Epic 16–19 menerapkan idempotency pada mutasi yang dapat diproses ulang, isolasi tenant, kontrol akses, audit, dan pengujian privasi sesuai backlog.
- [ ] Migration dijalankan terpisah, teruji di staging, dan kompatibel lintas rilis.
- [ ] Unit, integration, authorization, serta E2E test meliputi alur utama.
