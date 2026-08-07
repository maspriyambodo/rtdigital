# Dokumentasi Pengembangan Aplikasi Mobile RT Digital

Dokumen ini mendefinisikan ruang lingkup, arsitektur modul, dan development backlog untuk aplikasi RT Digital versi mobile (Native Android/iOS atau PWA Native Hybrid).

---

## 1. Arsitektur Modul Mobile

Aplikasi mobile RT Digital dibagi berdasarkan dua kelompok pengguna utama: **Warga** (akses mobile-first/layanan mandiri) dan **Pengurus/Petugas** (layanan operasional & persetujuan cepat).

### 1.1 Modul Warga (Citizen Services)

| Modul | Deskripsi | Fitur Utama |
|---|---|---|
| **Auth & Profile** | Akses masuk & manajemen profil aman. | - Login (Nomor Telepon/Email/Kata Sandi).<br>- Autentikasi Biometrik (Face ID/Fingerprint).<br>- Manajemen Keluarga (Lihat data anggota, KK, NIK ter-masking).<br>- Pengajuan Koreksi Data Kependudukan. |
| **Feed & Events** | Portal informasi & pengumuman lingkungan. | - Feed Pengumuman (Filter kategori, lampiran gambar/PDF).<br>- Agenda Kegiatan RT (Jadwal, detail lokasi, penanggung jawab).<br>- Integrasi Kalender Perangkat. |
| **Dues & Payments** | Pengelolaan pembayaran iuran warga. | - Rincian Tagihan Aktif & Tunggakan.<br>- Upload Bukti Pembayaran Manual (Kamera/Galeri).<br>- Transparansi Kas RT (Laporan ringkas kas). |
| **E-Surat** | Pengajuan dokumen administrasi mandiri. | - Form Permohonan Surat Pengantar RT.<br>- Tracking Status Pengajuan Real-time.<br>- Viewer & Unduh Surat Digital (Format PDF + QR Code Validasi). |
| **Aduan & Laporan** | Kanal pelaporan masalah lingkungan. | - Formulir Aduan (Upload foto kamera langsung, koordinat GPS).<br>- Tracking Timeline Penanganan (Tiket aduan). |
| **Keamanan** | Fitur sosial & proteksi lingkungan. | - Panic Button (Kirim alert bahaya + GPS ke petugas ronda).<br>- Jadwal Ronda Malam & Absensi Siskamling. |

### 1.2 Modul Pengurus & Petugas (Field Administration)

| Modul | Deskripsi | Fitur Utama |
|---|---|---|
| **Mobile Approvals** | Persetujuan dokumen & tindak lanjut aduan. | - Approval Surat Pengantar RT sekali ketuk (One-click approve).<br>- Tindak Lanjut Aduan (Update status tiket lapangan, upload foto). |
| **Field Verification** | Validasi data fisik di lapangan. | - Pemindai QR Code (Validasi fisik surat warga).<br>- Cek Pembayaran Tunai & Upload Bukti Setoran di tempat. |
| **Emergency** | Diseminasi informasi darurat cepat. | - Kirim Notifikasi Darurat (Broadcast push notification massal).<br>- Penerima Isyarat Panic Button Warga. |

---

## 2. Peta Jalan & Backlog Pengembangan (Mobile App Backlog)

### Epic M0: Fondasi & Arsitektur Mobile App
Tujuan: Menyiapkan framework, routing, push engine, storage offline, dan design token.
- [ ] **Task M0.1:** Inisialisasi Mobile App Project (React Native / Flutter / PWA Hybrid) dengan TypeScript strict.
- [ ] **Task M0.2:** Setup Push Notification Service (Firebase Cloud Messaging / APNs).
- [ ] **Task M0.3:** Implementasi SQLite/WatermelonDB/React Query untuk local offline cache (skema *offline-first* terbatas).
- [ ] **Task M0.4:** Integrasi API Perangkat (Permissions manager, Kamera, Galeri, Geolocation/GPS).
- [ ] **Task M0.5:** Setup Shell UI: Bottom navigation warga (maksimal 5 item), Sidebar pengurus, Design Token (Dark/Light mode).

### Epic M1: Autentikasi & Profil Pengguna
Tujuan: Mengamankan session di perangkat mobile dan melacak data warga.
- [ ] **Task M1.1:** Screen Login, Integrasi Refresh Token Rotation (HttpOnly/Secure Storage), & logout multi-device.
- [ ] **Task M1.2:** Setup Biometric Login (Face ID/Fingerprint) & fallback PIN.
- [ ] **Task M1.3:** Screen Profil Keluarga, Masking NIK/No KK di UI, dan Form Pengisian Usulan Koreksi Data.

### Epic M2: Informasi & Agenda RT
Tujuan: Menyediakan feed berita dan pengumuman real-time ke warga.
- [ ] **Task M2.1:** Screen Feed Pengumuman RT (Infinite scroll, Pull-to-refresh).
- [ ] **Task M2.2:** Viewer Dokumen Lampiran (Gambar & PDF viewer in-app).
- [ ] **Task M2.3:** Integrasi Push Notification untuk pengumuman baru berdasarkan target sasaran.
- [ ] **Task M2.4:** Screen Agenda RT & Integrasi "Simpan ke Kalender HP".

### Epic M3: Pembayaran & Manajemen Iuran
Tujuan: Memfasilitasi warga membayar iuran lewat mobile secara transparan.
- [ ] **Task M3.1:** Screen Rincian Iuran Warga (Riwayat bayar, Tagihan berjalan, Total tunggakan).
- [ ] **Task M3.2:** Integrasi Kamera untuk Capture Bukti Transfer fisik / Screen Capture.
- [ ] **Task M3.3:** Screen Laporan Transparansi Kas RT (Grafik ringkas, data teragregasi).
- [ ] **Task M3.4:** *(Khusus Pengurus)* Quick Action Verifikasi Bayar: Upload foto setoran tunai & verifikasi bukti transfer.

### Epic M4: Pengajuan Surat & Approval
Tujuan: Mempercepat layanan surat pengantar RT lewat handphone.
- [ ] **Task M4.1:** Form Pengajuan Surat Pengantar RT dengan validasi kelengkapan dokumen pendukung.
- [ ] **Task M4.2:** Screen Status Tracking pengajuan surat.
- [ ] **Task M4.3:** Viewer Surat Digital (PDF) dengan QR Code verifikasi.
- [ ] **Task M4.4:** *(Khusus Pengurus)* Screen Approval Surat: Ketua RT menyetujui atau meminta revisi pengajuan surat warga.

### Epic M5: Pelaporan Aduan & Penanganan Lapangan
Tujuan: Melaporkan kejadian dan mengelola tindak lanjut lapangan.
- [ ] **Task M5.1:** Form Buat Aduan dengan upload foto langsung dari kamera + Geotagging otomatis.
- [ ] **Task M5.2:** Timeline Status Penanganan Aduan (Tiket: Diajukan -> Diproses -> Selesai).
- [ ] **Task M5.3:** Push notification saat ada pembaruan komentar/status aduan.
- [ ] **Task M5.4:** *(Khusus Pengurus)* Dashboard Tiket Aduan Masuk & Tombol Disposisi Petugas.

### Epic M6: Fitur Keamanan Eksklusif Mobile (Pasca-MVP)
Tujuan: Menunjang keamanan lingkungan yang responsif menggunakan kapabilitas native handphone.
- [ ] **Task M6.1:** Tombol Panic Button (Widget/Shortcut layar utama) untuk mengirim pesan darurat & GPS.
- [ ] **Task M6.2:** Dashboard Penerima Alert Panic Button (Untuk Petugas Ronda/Pengurus).
- [ ] **Task M6.3:** Modul Ronda: QRCode check-in di pos/ronda malam & rekap absensi.

