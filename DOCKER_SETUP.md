# Panduan Docker Development Lokal

**Status:** Draft  
**Proyek:** RT Digital  
**Cakupan:** Lingkungan pengembangan lokal. Bukan konfigurasi production.  
**Referensi:** `docker-compose.yml`, `REPOSITORY_STRUCTURE.md`, `TECHNICAL_SPECIFICATION.md`

Docker Compose menjalankan frontend, Go API, serta seluruh dependensi lokal agar development konsisten tanpa memasang PostgreSQL, Redis, Mailpit, atau MinIO langsung pada host.

## 1. Prasyarat

- Docker Desktop atau Docker Engine dengan Docker Compose plugin v2.
- `pnpm` bila menjalankan frontend dari host atau mengelola dependency workspace.
- Git.

Go, PostgreSQL, Redis, Mailpit, dan MinIO tidak perlu dipasang pada host.

> `apps/web`, `services/api`, serta Dockerfile development belum di-scaffold pada repository ini. Karena itu `web` dan `api` belum dapat di-build sampai struktur tersebut dibuat sesuai `REPOSITORY_STRUCTURE.md`.

## 2. Layanan

| Layanan | Port host | Fungsi |
|---|---:|---|
| `web` | `3000` | Next.js development server dengan hot reload. |
| `api` | `8080` | Go REST API. |
| `postgres` | `5432` | PostgreSQL 18.4; database utama. |
| `redis` | `6379` | Cache/rate limiting/job lokal bila diperlukan. |
| `mailpit` | `8025`, `1025` | UI email dan SMTP catcher lokal. Menggantikan Resend saat development. |
| `minio` | `9000`, `9001` | Emulator S3-compatible untuk file privat. |
| `minio-init` | — | Container sementara pembuat bucket `rtdigital-local`, lalu berhenti. |

## 3. Menjalankan Setelah Scaffolding

Dari root repository:

```bash
docker compose up --build
```

Akses layanan:

| Layanan | URL / koneksi |
|---|---|
| Frontend | http://localhost:3000 |
| API | http://localhost:8080 |
| API health | http://localhost:8080/healthz |
| Mailpit | http://localhost:8025 |
| MinIO Console | http://localhost:9001 |
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |

Container `api` menunggu PostgreSQL, Redis, MinIO, dan inisialisasi bucket sebelum mulai. Container `web` menunggu API siap.

## 4. Konfigurasi Lokal

Kredensial berikut hanya untuk development lokal. Jangan dipakai untuk staging atau production.

### PostgreSQL

```text
Host:     localhost
Port:     5432
User:     rtdigital
Password: rtdigital_local_only
Database: rtdigital
```

DSN dari container API:

```text
postgres://rtdigital:rtdigital_local_only@postgres:5432/rtdigital?sslmode=disable
```

### Redis

```text
Host: localhost
Port: 6379
URL internal: redis://redis:6379/0
```

### Mailpit

```text
SMTP host internal: mailpit
SMTP port:          1025
Web UI:             http://localhost:8025
```

Mailpit menangkap email lokal. Tidak meneruskan email ke penerima asli.

### MinIO

```text
API endpoint: http://localhost:9000
Console:      http://localhost:9001
Access key:   minioadmin
Secret key:   minioadmin
Bucket:       rtdigital-local
Path style:   true
```

Endpoint dari container API:

```text
http://minio:9000
```

## 5. Menghentikan dan Mereset

Hentikan layanan tanpa menghapus data:

```bash
docker compose down
```

Hapus container beserta seluruh volume lokal:

```bash
docker compose down -v
```

Peringatan: perintah kedua menghapus data PostgreSQL, Redis, cache module Go, `node_modules` container, serta data MinIO lokal.

## 6. Operasi Berguna

Akses PostgreSQL:

```bash
docker compose exec postgres psql -U rtdigital rtdigital
```

Akses Redis CLI:

```bash
docker compose exec redis redis-cli
```

Lihat status layanan:

```bash
docker compose ps
```

Lihat log API:

```bash
docker compose logs -f api
```

## 7. Menjalankan Frontend dari Host

Untuk HMR lebih cepat, jalankan dependency dan API dalam Docker, lalu Next.js dari host:

```bash
docker compose up api postgres redis mailpit minio minio-init
```

Kemudian:

```bash
cd apps/web
pnpm install
pnpm dev
```

Atur endpoint browser:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## 8. Batasan

- Mailpit menggantikan Resend hanya di lokal.
- MinIO adalah emulator S3-compatible yang digunakan untuk meniru Cloudflare R2 di lokal.
- Redis disediakan untuk keseragaman development; arsitektur MVP tidak menjadikannya dependency session wajib.
- SaungWA tidak diemulasikan. Integrasi WhatsApp lokal memakai adapter/no-op atau sandbox provider saat implementasi tersedia.
- Secret production tetap menggunakan AWS Secrets Manager; tidak masuk `.env`, Docker image, atau repository.