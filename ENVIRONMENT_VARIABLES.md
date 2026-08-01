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
| `AWS_REGION` | Go API | Ya | Region AWS untuk S3 dan layanan AWS lain. | `us-east-1` | `ap-southeast-3` |
| `AWS_S3_BUCKET` | Go API | Ya | Bucket privat lampiran dan dokumen. | `rtdigital-local` | `rtdigital-prod-documents` |

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
| `AWS_REGION` | Ya | `us-east-1` | Region SDK AWS atau region kompatibel MinIO. |
| `AWS_S3_BUCKET` | Ya | `rtdigital-local` | Nama bucket target. |
| `S3_ENDPOINT` | Tidak | `http://minio:9000` | Endpoint MinIO lokal; kosongkan untuk AWS S3. |
| `S3_USE_PATH_STYLE` | Tidak | `true` | `true` untuk MinIO; `false` untuk AWS S3. |
| `AWS_ACCESS_KEY_ID` | Lokal saja | `minioadmin` | Kredensial MinIO. ECS production memakai IAM Task Role. |
| `AWS_SECRET_ACCESS_KEY` | Lokal saja | `minioadmin` | Kredensial MinIO. ECS production memakai IAM Task Role. |
| `SMTP_HOST` | Lokal saja | `mailpit` | Host Mailpit untuk pengujian email. |
| `SMTP_PORT` | Lokal saja | `1025` | Port SMTP Mailpit. |
| `RESEND_API_KEY` | Staging/production | Kosong | API key Resend. Jangan gunakan Mailpit untuk pengiriman nyata. |
| `RESEND_FROM_EMAIL` | Staging/production | Kosong | Alamat/domain pengirim yang sudah diverifikasi Resend. |
| `SAUNGWA_API_KEY` | Staging/production | Kosong | API key SaungWA. |
| `SAUNGWA_ENDPOINT` | Tidak | Sesuai dokumentasi SaungWA | Endpoint API SaungWA. Jangan menebak atau hardcode endpoint tanpa verifikasi dokumentasi provider. |

## 4. Frontend: Next.js

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

# Object storage: MinIO local only
AWS_REGION=us-east-1
AWS_S3_BUCKET=rtdigital-local
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
S3_ENDPOINT=http://minio:9000
S3_USE_PATH_STYLE=true

# Local email testing
SMTP_HOST=mailpit
SMTP_PORT=1025

# External notification providers: leave empty locally unless deliberately tested
RESEND_API_KEY=
RESEND_FROM_EMAIL=
SAUNGWA_API_KEY=
SAUNGWA_ENDPOINT=
```

## 6. Matriks Environment

| Area | Local | Staging | Production |
|---|---|---|---|
| PostgreSQL | Container Compose | RDS staging | RDS private subnet |
| Redis | Container Compose | Sesuai kebutuhan | Sesuai kebutuhan |
| File | MinIO | S3 bucket staging | S3 bucket production |
| Email | Mailpit | Resend | Resend |
| WhatsApp | Adapter mock/no-op | SaungWA test environment bila tersedia | SaungWA production |
| Secret backend | `.env` lokal | AWS Secrets Manager | AWS Secrets Manager |
| AWS credential | MinIO credential lokal | IAM Task Role | IAM Task Role |

## 7. Checklist Sebelum Deploy

- [ ] `DATABASE_URL` memakai koneksi TLS di staging/production.
- [ ] `JWT_SECRET` unik, acak, minimal 32 byte, dan tidak tercatat pada log.
- [ ] `APP_URL`, `API_URL`, serta URL `NEXT_PUBLIC_*` memakai domain environment yang benar.
- [ ] `AWS_S3_BUCKET` privat dan berbeda untuk staging/production.
- [ ] ECS Task Role memiliki least-privilege access ke bucket dan CloudWatch.
- [ ] `RESEND_FROM_EMAIL` telah diverifikasi.
- [ ] Credential Resend dan SaungWA tersimpan sebagai secret, bukan source code.