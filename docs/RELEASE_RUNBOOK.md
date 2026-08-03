# Runbook Rilis dan Operasional RT Digital

Cakupan: Task 12.5, 12.6, 12.7, 12.8. Jangan rilis produksi sebelum seluruh checklist staging selesai.

## Observabilitas

API mengeluarkan JSON log `slog` ke stdout. Setiap request memiliki `X-Request-ID`; jangan log request body, `Authorization`, cookie, token, kata sandi, NIK, nomor KK, atau dokumen warga.

| Sinyal | Ambang alarm | Tindakan |
| --- | --- | --- |
| HTTP 5xx | >1% selama 5 menit | Cek log dengan `request_id`, rollback bila rilis penyebabnya |
| Latensi P95 | >2 detik selama 5 menit | Cek DB pool, query lambat, R2/provider |
| `/readyz` | 503 selama 2 menit | Periksa RDS/network, hentikan rollout |
| DB connection pool | idle <5 atau koneksi gagal | Skalakan/perbaiki konfigurasi pool |
| Upload/R2 error | kenaikan berkelanjutan 5 menit | Periksa kredensial, bucket, quota |

- `GET /healthz`: proses API hidup.
- `GET /readyz`: koneksi PostgreSQL tervalidasi dengan batas 3 detik.
- CloudWatch: tangkap stdout container sebagai JSON; buat metric filter/alarm untuk 5xx dan kegagalan readiness.
- Error tracker pihak ketiga tidak dipakai pada MVP. Tambahkan hanya setelah scrubber PII diverifikasi.

## Backup dan Restore

### PostgreSQL/RDS

- Aktifkan automated backup, PITR, delete protection.
- Retensi minimal 14 hari.
- Ambil snapshot sebelum migrasi skema besar.
- Uji restore ke staging minimal setiap kuartal, tidak pernah menimpa produksi.

Prosedur restore staging:

1. Restore snapshot/PITR ke instance PostgreSQL staging baru.
2. Set `DATABASE_URL` staging ke instance hasil restore.
3. Jalankan migrasi aplikasi yang belum diterapkan memakai mekanisme migration proyek.
4. Jalankan `GET /readyz`, login akun uji, cek data warga tersamarkan.
5. Catat waktu recovery, RPO, RTO, hasil, operator pada audit operasional.

### Cloudflare R2

- Aktifkan object versioning bila tersedia pada bucket produksi.
- Tetapkan lifecycle untuk objek sementara sesuai kebijakan produk.
- Simpan salinan objek kritis pada bucket/account terpisah bila kebutuhan DR meningkat.
- Uji signed download dari data hasil restore staging. Jangan memakai bukti pembayaran nyata dalam UAT.

## Checklist Staging dan Soft Launch

### Sebelum UAT

- [ ] Variabel produksi/staging berbeda; secret tidak masuk repository.
- [ ] Migration database berhasil.
- [ ] `/healthz` dan `/readyz` lulus.
- [ ] CORS hanya mengizinkan origin aplikasi.
- [ ] Header `X-Content-Type-Options`, `X-Frame-Options`, CSP, HSTS produksi terverifikasi.
- [ ] Backup RDS tersedia; restore staging diuji.
- [ ] R2 upload/download signed URL lulus.
- [ ] Semua test API, web lint/type check/build lulus.

### UAT Pengurus

- [ ] Ketua RT: pengaturan RT, audit log, MFA, pengumuman, surat.
- [ ] Bendahara: tagihan, pembayaran, pembalikan kas.
- [ ] Sekretaris: surat, warga, laporan.
- [ ] Warga: profil, koreksi data, pembayaran, surat, aduan.
- [ ] Import CSV final: dry-run, duplikasi NIK, hasil import, audit.
- [ ] Uji viewport 320 px, 360 px, 390 px, desktop.
- [ ] Uji keyboard, fokus, label form, dialog audit log.
- [ ] Uji tenant isolation dan permission utama.
- [ ] Kebijakan privasi, runbook, kontak insiden disetujui.

### Soft Launch

1. Buat snapshot database.
2. Terapkan migration.
3. Deploy API dan web.
4. Jalankan smoke test login, dashboard, upload, pembayaran, surat, audit log.
5. Pantau alarm dan error selama 24 jam.
6. Bila ada insiden data/keamanan: hentikan perubahan, simpan `request_id`, rollback aplikasi atau restore sesuai hasil triage.

`ponytail:` alarm didefinisikan sebagai runbook CloudWatch, bukan IaC. Tambahkan Terraform/CloudFormation saat target AWS produksi disahkan.