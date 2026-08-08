# Variabel Environment RT Digital

**Status:** Draft  
**Cakupan:** Next.js, Go API, Docker lokal, staging, production.  
**Referensi:** `TECHNICAL_SPECIFICATION.md`, `SYSTEM_ARCHITECTURE.md`, `DOCKER_SETUP.md`, `docker-compose.yml`

Dokumen ini menetapkan konfigurasi runtime. Jangan commit `.env`, API key, password, token, atau kredensial production.

## 1. Aturan Secret

1. Commit hanya `.env.example` tanpa secret nyata.
2. Lokal: simpan nilai pada `.env` yang diabaikan Git.
3. Staging/production backend: gunakan AWS Secrets Manager atau Parameter Store.
4. Staging/production frontend: gunakan encrypted environment variables Cloudflare.
5. ECS memakai IAM Task Role untuk akses AWS; jangan menyimpan AWS access key production dalam environment variable.
6. Go API memvalidasi konfigurasi wajib saat startup dan berhenti bila nilai tidak valid.
7. Variabel `NEXT_PUBLIC_*` tersedia di browser. Jangan pernah memakai prefix ini untuk secret.

## 2. Variabel Inti

| Variabel | Komponen | Wajib | Deskripsi | Contoh lokal | Contoh production |
|---|---|---:|---|---|---|
| `DATABASE_URL` | Go API | Ya | DSN PostgreSQL. Production wajib TLS. | `postgres://rtdigital:rtdigital_local_only@postgres:5432/rtdigital?sslmode=disable` | `postgres://...rds.amazonaws.com:5432/rtdigital?sslmode=require` |
| `JWT_SECRET` | Go API | Ya | Kunci signing access token JWT. Minimum 32 byte acak. | Nilai acak development | Secret acak dari Secrets Manager |
| `APP_URL` | Go API | Ya | URL publik frontend untuk CORS dan tautan email. | `http://localhost:3000` | `https://app.domain-rt.id` |
| `API_URL` | Go API | Ya | URL publik API, termasuk prefix `/api/v1`. | `http://localhost:8080/api/v1` | `https://api.domain-rt.id/api/v1` |
| `R2_ACCOUNT_ID` | Go API | Staging/production | Cloudflare Account ID untuk endpoint R2. | Kosong | `<cloudflare_account_id>` |
| `R2_ACCESS_KEY_ID` | Go API | Ya | Access Key ID API token R2; kredensial MinIO di lokal. | `minioadmin` | Secret dari Cloudflare |
| `R2_SECRET_ACCESS_KEY` | Go API | Ya | Secret Access Key API token R2; password MinIO di lokal. | `minioadmin` | Secret dari Cloudflare |
| `R2_BUCKET` | Go API | Ya | Bucket privat lampiran dan dokumen. | `rtdigital-local` | `rtdigital-prod-documents` |
| `R2_ENDPOINT` | Go API | Ya | Endpoint S3-compatible R2 atau MinIO. | `http://minio:9000` | `https://<account_id>.r2.cloudflarestorage.com` |

## 3. Backend: Go API

| Variabel | Wajib | Default lokal | Ketentuan |
|---|---:|---|---|
| `APP_ENV` | Ya | `development` | Salah satu: `development`, `staging`, `production`. |
| `PORT` | Tidak | `8080` | Port listener HTTP API. |
| `DATABASE_URL` | Ya | Lihat bagian 2 | URI PostgreSQL valid. |
| `JWT_SECRET` | Ya | Tidak ada | Minimum 32 byte; wajib unik per environment. |
| `JWT_ACCESS_TTL` | Tidak | `15m` | Durasi Go yang valid. |
| `APP_URL` | Ya | `http://localhost:3000` | Origin frontend tanpa trailing slash. |
| `API_URL` | Ya | `http://localhost:8080/api/v1` | URL API publik dengan prefix versi. |
| `REDIS_URL` | Tidak | `redis://redis:6379/0` | Redis lokal tersedia; bukan dependency session wajib MVP. |
| `R2_ACCOUNT_ID` | Staging/production | Kosong | Cloudflare Account ID pembentuk endpoint R2. |
| `R2_ACCESS_KEY_ID` | Ya | `minioadmin` | R2 Access Key ID atau kredensial MinIO lokal. |
| `R2_SECRET_ACCESS_KEY` | Ya | `minioadmin` | R2 Secret Access Key atau password MinIO lokal. |
| `R2_BUCKET` | Ya | `rtdigital-local` | Nama bucket target. |
| `R2_ENDPOINT` | Ya | `http://minio:9000` | Endpoint S3-compatible MinIO lokal atau R2. |
| `R2_USE_PATH_STYLE` | Tidak | `true` | `true` untuk MinIO lokal; `false` untuk R2 production. |
| `RESEND_API_KEY` | Ya | Kosong | API key Resend untuk environment terkait. |
| `RESEND_FROM_EMAIL` | Staging/production | Kosong | Alamat/domain pengirim yang sudah diverifikasi Resend. |
| `SAUNGWA_API_KEY` | Staging/production | Kosong | Kredensial SaungWA: `<appkey>:<authkey>`. |
| `SAUNGWA_ENDPOINT` | Tidak | `https://app.saungwa.com/api/create-message` | Endpoint SaungWA pengiriman pesan WhatsApp. |

## 4. Web App: Next.js

| Variabel | Wajib | Diekspos ke browser | Deskripsi |
|---|---:|---:|---|
| `NEXT_PUBLIC_APP_URL` | Ya | Ya | URL basis aplikasi untuk metadata, PWA, dan tautan publik. |
| `NEXT_PUBLIC_API_URL` | Ya | Ya | URL basis API yang dipanggil browser. |
| `API_INTERNAL_URL` | Tidak | Tidak | URL internal API untuk Server Component bila diperlukan. Jangan gunakan bila browser fetch sudah memadai. |

`APP_URL` dan `API_URL` adalah konfigurasi Go API. `NEXT_PUBLIC_APP_URL` dan `NEXT_PUBLIC_API_URL` adalah konfigurasi frontend. Nilai lokalnya boleh sama, tetapi nama variabel tidak saling menggantikan.

## 5. Template `.env.example`

```env
# Runtime
APP_ENV=development
PORT=8080

# Public URLs
APP_URL=http://localhost:3000
API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# Database and cache
DATABASE_URL=postgres://rtdigital:rtdigital_local_only@postgres:5432/rtdigital?sslmode=disable
REDIS_URL=redis://redis:6379/0

# Security: generate a unique value, e.g. `openssl rand -base64 48`
JWT_SECRET=replace_with_a_local_random_secret_at_least_32_bytes

# Object storage: Cloudflare R2 production; MinIO local
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=minioadmin
R2_SECRET_ACCESS_KEY=minioadmin
R2_BUCKET=rtdigital-local
R2_ENDPOINT=http://minio:9000
R2_USE_PATH_STYLE=true

# Notification providers: Resend for email, SaungWA for WhatsApp.
# Leave empty locally unless deliberately tested.
RESEND_API_KEY=
RESEND_FROM_EMAIL=
# Format: <appkey>:<authkey>
SAUNGWA_API_KEY=
# Optional. Default: https://app.saungwa.com/api/create-message
SAUNGWA_ENDPOINT=
```

## 6. Matriks Environment

| Area | Local | Staging | Production |
|---|---|---|---|
| PostgreSQL | Container Compose | RDS staging | RDS private subnet |
| Redis | Container Compose | Sesuai kebutuhan | Sesuai kebutuhan |
| File | MinIO | Cloudflare R2 staging | Cloudflare R2 production |
| Email | Resend development key | Resend staging key | Resend production key |
| WhatsApp | Adapter mock/no-op | SaungWA test environment bila tersedia | SaungWA production |
| Secret backend | `.env` lokal | AWS Secrets Manager | AWS Secrets Manager |
| R2 credential | MinIO credential lokal | R2 API token secret | R2 API token secret |

## 7. Checklist Sebelum Deploy

- [ ] `DATABASE_URL` memakai koneksi TLS di staging/production.
- [ ] `JWT_SECRET` unik, acak, minimal 32 byte, dan tidak tercatat pada log.
- [ ] `APP_URL`, `API_URL`, serta URL `NEXT_PUBLIC_*` memakai domain environment yang benar.
- [ ] `R2_BUCKET` privat dan berbeda untuk staging/production.
- [ ] API token R2 hanya memiliki izin bucket yang diperlukan.
- [ ] ECS Task Role memiliki izin CloudWatch Logs dan AWS Secrets Manager untuk membaca secret R2.
- [ ] `RESEND_FROM_EMAIL` telah diverifikasi.
- [ ] Credential Resend dan SaungWA tersimpan sebagai secret, bukan source code.