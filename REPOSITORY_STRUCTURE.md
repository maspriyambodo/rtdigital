# Struktur Repository RT Digital

**Status:** Draft untuk validasi  
**Jenis:** Monorepo  
**Cakupan:** MVP RT Digital  
**Tooling:** `pnpm` workspace untuk TypeScript/Next.js, Go modules untuk API, Docker Compose, Makefile ringan.  
**Referensi:** `TECHNICAL_SPECIFICATION.md`, `SYSTEM_ARCHITECTURE.md`, `API_SPECIFICATION.md`

Monorepo menyatukan frontend, backend, kontrak API, infrastruktur, serta dokumentasi. Deployment frontend Cloudflare dan backend AWS tetap independen.

Tidak menggunakan Turborepo, Nx, Bazel, ORM, atau package shared tambahan pada awal MVP. Tambahkan hanya bila waktu build atau duplikasi kontrak terbukti menjadi masalah.

---

## 1. Struktur Direktori Target

```text
/
├── apps/
│   └── web/                         # Next.js frontend
├── services/
│   └── api/                         # Go modular monolith API
├── packages/
│   └── api-client/                  # Opsional, generated dari OpenAPI setelah kontrak stabil
├── infrastructure/
│   ├── docker/                      # Dockerfile dan konfigurasi lokal
│   ├── aws/                         # IaC/deployment AWS
│   └── cloudflare/                  # Wrangler/konfigurasi Cloudflare
├── docs/                            # Seluruh dokumen produk dan teknis
├── scripts/                         # Script utilitas non-rahasia
├── .github/
│   └── workflows/                   # CI/CD GitHub Actions
├── .editorconfig
├── .env.example                     # Template tanpa secret
├── .gitignore
├── docker-compose.yml
├── Makefile
├── package.json                     # Root pnpm workspace
├── pnpm-workspace.yaml
└── README.md
```

**Catatan migrasi dokumen:** file Markdown saat ini dapat tetap berada di root selama tahap perencanaan. Saat scaffold aplikasi dibuat, pindahkan seluruh dokumen ke `docs/` dalam satu commit khusus agar referensi internal diperbarui konsisten.

---

## 2. Frontend: `apps/web`

```text
apps/web/
├── public/
│   ├── icons/
│   ├── manifest.webmanifest
│   └── robots.txt
├── src/
│   ├── app/
│   │   ├── (auth)/                  # Login, aktivasi, reset password
│   │   ├── app/                     # Area Warga
│   │   ├── admin/                   # Area Pengurus
│   │   ├── layout.tsx
│   │   ├── error.tsx
│   │   └── not-found.tsx
│   ├── components/
│   │   ├── ui/                      # Button, FormField, DataTable, dll.
│   │   ├── layout/                  # AppShell, Sidebar, BottomNavigation
│   │   └── feedback/                # Skeleton, ErrorState, OfflineNotice
│   ├── features/
│   │   ├── identity/
│   │   ├── resident/
│   │   ├── finance/
│   │   ├── communication/
│   │   ├── letter/
│   │   └── complaint/
│   ├── lib/
│   │   ├── api.ts                   # Wrapper fetch tipis
│   │   ├── auth.ts                  # Access token memory dan refresh
│   │   ├── errors.ts                # Tipe API error
│   │   └── format.ts                # Rupiah, tanggal, masking
│   ├── types/                       # Tipe spesifik frontend
│   └── styles/
│       └── globals.css
├── .env.example
├── next.config.mjs
├── opennext.config.ts
├── package.json
└── tsconfig.json
```

Aturan:

- App Router, TypeScript strict, Server Components default.
- `components/ui` mengikuti `UI_COMPONENTS.md`.
- `features` memuat logika domain frontend, bukan komponen generik.
- Tidak membuat Next.js proxy API untuk seluruh Go API.
- Build/deploy: `pnpm --filter web build`, kemudian deploy OpenNext ke Cloudflare Workers.

---

## 3. Backend: `services/api`

```text
services/api/
├── cmd/
│   └── server/
│       └── main.go                  # Bootstrap, DI, graceful shutdown
├── internal/
│   ├── app/
│   │   ├── auth/                    # Password, JWT, session, MFA
│   │   ├── config/                  # Validasi environment
│   │   ├── database/                # pgxpool dan transaction helper
│   │   ├── httpx/                   # JSON response dan AppError
│   │   ├── logging/                 # slog dan sanitasi log
│   │   └── middleware/              # Request ID, recover, CORS, auth, RBAC
│   ├── modules/
│   │   ├── identity/                # User, role, permission, session
│   │   ├── resident/                # Rumah, keluarga, warga, koreksi
│   │   ├── communication/           # Pengumuman, agenda, notifikasi
│   │   ├── finance/                 # Iuran, invoice, payment, kas
│   │   ├── letter/                  # Jenis dan pengajuan surat
│   │   ├── complaint/               # Aduan dan komentar
│   │   ├── file/                    # Metadata file dan signed URL
│   │   └── audit/                   # Audit append-only
│   └── infrastructure/
│       ├── email/                   # Resend client
│       ├── whatsapp/                # SaungWA client
│       └── r2/                      # Cloudflare R2 S3-compatible client
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── .env.example
├── Dockerfile
├── go.mod
└── go.sum
```

Struktur setiap modul:

```text
internal/modules/finance/
├── handler.go                       # HTTP request/response
├── service.go                       # Aturan bisnis dan transaksi
├── repository.go                    # SQL parameterized PostgreSQL
├── types.go                         # DTO, filter, model domain
└── service_test.go                  # Test aturan bisnis
```

Klien eksternal:

- `email/` menggunakan [Resend](https://resend.com) untuk email transaksional MVP.
- `whatsapp/` menggunakan [SaungWA](https://saungwa.com/) untuk notifikasi WhatsApp pada MVP.
- `r2/` menggunakan AWS SDK Go melalui API S3-compatible untuk pre-signed URL dan `HeadObject` ke Cloudflare R2 atau MinIO lokal.

Aturan:

- `handler` tidak menjalankan query database.
- `service` mengelola transaksi, state transition, audit, serta otorisasi bisnis.
- `repository` hanya menjalankan query parameterized.
- Migration berada di `services/api/migrations` agar ikut image/task migration yang sama.
- Build/deploy: Docker image → Amazon ECR → Amazon ECS Fargate.

---

## 4. Package Bersama: `packages`

```text
packages/
└── api-client/                      # Opsional
    ├── src/
    ├── package.json
    └── tsconfig.json
```

`api-client` dibuat hanya setelah OpenAPI menjadi kontrak stabil dan request/type frontend mulai berulang.

Aturan:

- Sumber kontrak API adalah OpenAPI, bukan model database.
- Jangan membuat `shared-types` pada MVP awal. DTO Go dan TypeScript dapat dipisahkan sampai generator OpenAPI diperlukan.
- Tidak ada logika bisnis frontend di `packages/`.

---

## 5. Infrastruktur: `infrastructure`

```text
infrastructure/
├── docker/
│   ├── api/
│   │   └── Dockerfile.dev
│   └── web/
│       └── Dockerfile.dev
├── aws/
│   ├── ecs/
│   ├── rds/
│   ├── r2/
│   └── README.md
└── cloudflare/
    ├── wrangler.jsonc
    └── README.md
```

Aturan:

- Dockerfile production dapat berada dekat service (`services/api/Dockerfile`) agar build context sederhana.
- IaC AWS/Cloudflare ditambahkan saat environment staging/production diotomasi.
- Jangan menyimpan secret, state Terraform, kredensial AWS, atau token Cloudflare di repository.

---

## 6. Docker Compose Development

Root `docker-compose.yml` menjalankan dependensi lokal:

```text
web       Next.js development server
api       Go API
postgres  PostgreSQL 18.4
mailpit   SMTP testing lokal untuk simulasi email sebelum integrasi Resend
```

Opsional:

```text
minio     Emulator S3-compatible lokal untuk Cloudflare R2, hanya saat menguji upload/download
```

Prinsip:

- Jalankan dengan `docker compose up`.
- PostgreSQL memakai volume persisten.
- `mailpit` tidak digunakan pada production.
- MinIO bukan dependency wajib sebelum fitur file dikerjakan.
- Hot reload frontend/backend tersedia.
- Secret lokal hanya berada di `.env`, tidak masuk Git.

---

## 7. Konfigurasi Root

### `pnpm-workspace.yaml`

```yaml
packages:
  - "apps/*"
  - "packages/*"
```

Go tidak dikelola pnpm. `services/api/go.mod` tetap module independen.

### `package.json`

Root hanya mengelola script workspace TypeScript dan tooling bersama:

```json
{
  "private": true,
  "packageManager": "pnpm@10",
  "scripts": {
    "dev:web": "pnpm --filter web dev",
    "build:web": "pnpm --filter web build",
    "lint:web": "pnpm --filter web lint",
    "test:web": "pnpm --filter web test",
    "test:api": "cd services/api && go test ./...",
    "lint:api": "cd services/api && gofmt -w . && go vet ./..."
  }
}
```

### `Makefile`

```makefile
.PHONY: up down test lint

up:
	docker compose up --build

down:
	docker compose down

test:
	pnpm test:web
	cd services/api && go test ./...

lint:
	pnpm lint:web
	cd services/api && gofmt -w . && go vet ./...
```

`gofmt -w` memodifikasi file. CI harus memakai pemeriksaan format non-mutatif, misalnya `test -z "$$(gofmt -l .)"`.

---

## 8. CI/CD dan Deployment Independen

| Perubahan | CI minimum | Deploy |
|---|---|---|
| `apps/web/**` | Lint, type check, test, build OpenNext | Cloudflare Workers |
| `services/api/**` | `gofmt` check, `go vet`, test, build image, image scan | ECR lalu ECS Fargate |
| `services/api/migrations/**` | Migration rehearsal | ECS migration task sebelum deploy API |
| `infrastructure/**` | Validasi IaC | Manual approval staging/production |
| `docs/**` | Markdown/link check bila tersedia | Tidak ada deploy aplikasi |

- Frontend dan API memakai pipeline, artifact, serta rollback sendiri.
- Perubahan kontrak OpenAPI memicu validasi API client sebelum frontend deploy.
- Migration dijalankan sekali sebagai task terpisah; tidak saat startup semua replica API.

---

## 9. Alur Kerja Developer

1. Clone repository.
2. Salin `.env.example` menjadi `.env`, isi secret lokal.
3. Jalankan `pnpm install`.
4. Jalankan `docker compose up --build`.
5. Akses frontend `http://localhost:3000`, API `http://localhost:8080`.
6. Jalankan `make test` sebelum pull request.
7. Perubahan kontrak API: perbarui OpenAPI, regenerate `packages/api-client` bila package tersebut sudah digunakan.

---

## 10. Batas MVP

- Tidak ada Turborepo/Nx/Bazel.
- Tidak ada package UI terpisah; komponen berada di `apps/web`.
- Tidak ada Redis lokal.
- Tidak ada microservice tambahan.
- Tidak ada shared model database antara Go dan TypeScript.
- Tidak ada deploy otomatis dari perubahan dokumentasi.