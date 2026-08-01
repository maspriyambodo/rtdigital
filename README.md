# RT Digital

Aplikasi web **mobile-first** untuk administrasi dan komunikasi Rukun Tetangga (RT). RT Digital membantu pengurus mengelola data warga, iuran, kas, surat, aduan, pengumuman, serta agenda secara terpusat, aman, dan terlacak.

## Status

Tahap perencanaan dan desain MVP. Repository ini saat ini berisi dokumentasi produk, UX, arsitektur, data, API, serta struktur implementasi.

## Fokus MVP

- Data rumah, keluarga, dan warga.
- RBAC, MFA pengurus, audit log, serta perlindungan data sensitif.
- Pengumuman, agenda, notifikasi dalam aplikasi.
- Notifikasi email melalui [Resend](https://resend.com).
- Notifikasi WhatsApp melalui [SaungWA](https://saungwa.com/).
- Iuran, tagihan, pembayaran manual, bukti transfer, dan buku kas.
- Pengajuan serta penerbitan surat PDF.
- Aduan warga dan tindak lanjut pengurus.
- PWA dasar dan UX mobile-first.

## Arsitektur Target

| Area | Teknologi |
|---|---|
| Frontend | Next.js App Router, TypeScript, OpenNext, Cloudflare Workers |
| Backend | Go modular monolith, REST API |
| Database | PostgreSQL 18.4, Amazon RDS |
| File privat | Amazon S3 dengan pre-signed URL |
| Email | Resend |
| WhatsApp | SaungWA |
| API runtime | Amazon ECS Fargate |
| Edge, DNS, WAF | Cloudflare |
| Lokal | Docker Compose, pnpm workspace, Go modules |

## Dokumentasi

### Produk dan UX

- [PRD](PRD.md)
- [Ruang Lingkup MVP](SCOPE.md)
- [Arsitektur Informasi](INFORMATION_ARCHITECTURE.md)
- [Alur Pengguna](USER_FLOW.md)
- [Design System](DESIGN_SYSTEM.md)
- [Komponen UI](UI_COMPONENTS.md)

### Teknis dan Infrastruktur

- [Arsitektur Sistem](SYSTEM_ARCHITECTURE.md)
- [Spesifikasi Teknis](TECHNICAL_SPECIFICATION.md)
- [Struktur Repository](REPOSITORY_STRUCTURE.md)
- [Desain Database](DATABASE_DESIGN.md)
- [Spesifikasi API](API_SPECIFICATION.md)
- [Peran dan Hak Akses](USER_ROLES_AND_PERMISSIONS.md)
- [Panduan Docker](DOCKER_SETUP.md)
- [Variabel Environment](ENVIRONMENT_VARIABLES.md)

## Pengembangan Lokal

Scaffolding aplikasi belum dibuat. Struktur target, perintah Docker Compose, environment, CI/CD, serta aturan implementasi tersedia di [REPOSITORY_STRUCTURE.md](REPOSITORY_STRUCTURE.md) dan [TECHNICAL_SPECIFICATION.md](TECHNICAL_SPECIFICATION.md).

Setelah scaffolding tersedia:

```bash
pnpm install
docker compose up --build
```

Frontend target: `http://localhost:3000`  
API target: `http://localhost:8080`

## Prinsip Utama

- Mobile-first untuk warga.
- Satu RT pada MVP; `organization_id` disiapkan untuk masa depan.
- Backend menjadi sumber keputusan autentikasi, otorisasi, validasi, transaksi, serta audit.
- Data privat tidak dicache secara publik.
- Transaksi keuangan tidak dihapus; koreksi menggunakan pembatalan atau pembalikan.

## Lisensi

Belum ditentukan.