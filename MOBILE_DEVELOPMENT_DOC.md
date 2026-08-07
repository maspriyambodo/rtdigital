# Dokumentasi Pengembangan Aplikasi Mobile RT Digital (Flutter Cross-Platform)

Dokumen ini mendefinisikan ruang lingkup, arsitektur teknis, arsitektur modul, dan development backlog aplikasi RT Digital berbasis **Flutter** untuk satu codebase **Android dan iOS**.

---

## 1. Stack Teknologi & Prinsip Arsitektur

| Area | Standar Flutter | Tujuan |
|---|---|---|
| Framework | Flutter dan Dart | UI lintas platform Android/iOS dari satu codebase. |
| Struktur aplikasi | Feature-first dengan lapisan presentation, application, domain, dan data | Memisahkan UI, aturan bisnis, dan akses API/storage agar mudah diuji. |
| State management & DI | `flutter_riverpod` | State async, dependency injection, serta state yang mudah diuji. |
| Routing | `go_router` | Deep link, redirect/auth guard, dan shell navigation. |
| API | `dio` | Interceptor autentikasi, refresh token terkoordinasi, retry terkontrol, dan multipart upload. |
| Penyimpanan aman | `flutter_secure_storage` | Menyimpan token di Android Keystore dan iOS Keychain. |
| Data offline | `drift` (SQLite) | Cache data terstruktur dan antrean sinkronisasi terbatas saat koneksi kembali. |
| Notifikasi | `firebase_messaging` + `flutter_local_notifications` | FCM untuk Android dan APNs melalui FCM untuk iOS; tampilan notifikasi saat aplikasi aktif. |
| Tema & UI | Material 3 `ThemeData` | Design token, mode terang/gelap, serta komponen UI konsisten. |
| Integrasi perangkat | `permission_handler`, `local_auth`, `image_picker`/`camera`, `geolocator`, `mobile_scanner` | Izin perangkat, biometrik, media, lokasi, dan QR code. |

**Ketentuan keamanan:** access token dan refresh token tidak disimpan di shared preferences/cache biasa; NIK dan nomor KK wajib dimasking di UI serta tidak dicatat pada log; izin kamera, foto, lokasi, biometrik, notifikasi, dan kalender diminta hanya saat fitur dipakai. Konfigurasi `Info.plist`, Android manifest, APNs, serta Firebase wajib diuji pada perangkat fisik Android dan iOS.

---

## 2. Arsitektur Modul Mobile

Aplikasi mobile RT Digital dibagi berdasarkan dua kelompok pengguna utama: **Warga** (akses mobile-first/layanan mandiri) dan **Pengurus/Petugas** (layanan operasional & persetujuan cepat).

### 2.0 Kriteria Modul Wajib Masuk Mobile

Tidak semua epic pada `DEVELOPMENT_BACKLOG.md` perlu dibawa ke mobile. Sebuah modul **wajib** masuk aplikasi mobile bila memenuhi minimal satu kriteria berikut:

1. **Membutuhkan kapabilitas native**: kamera, GPS, biometrik, pemindai QR, push notification, kalender perangkat.
2. **Bersifat mendesak/di lapangan**: aksi harus dilakukan saat kejadian berlangsung, bukan di depan desktop (panic button, absensi ronda, check-in tamu, disposisi aduan).
3. **Layanan mandiri warga bervolume tinggi**: warga adalah pengguna mobile-first, sehingga seluruh alur warga pada MVP wajib tersedia di mobile.
4. **Persetujuan yang memblokir alur**: approval pengurus yang menahan layanan warga bila tidak segera diproses.

Aturan interaksi wajib untuk seluruh modul mobile: setiap pemicu aksi mutasi (kirim, bayar, unggah bukti, setuju, tolak, absensi QR, panic button) masuk ke loading state pada ketukan pertama dan tidak dapat diketuk ulang sampai request selesai, gagal, atau timeout. Layar dengan koneksi lemah adalah kondisi paling rawan ketukan berulang, sehingga label tombol tetap terlihat selama proses dan hasil akhir dikonfirmasi eksplisit.

Modul yang **tidak wajib** masuk mobile (tetap dikerjakan di web pengurus): laporan analitis dan ekspor CSV/PDF (Epic 11 sisi pengurus), UI audit log (Task 12.3), master data administratif (Epic 13 sisi pengelolaan), konfigurasi RT dan operasional rilis (Epic 12), serta import CSV data warga (Task 3.7). Modul ini hanya ditampilkan dalam bentuk ringkasan baca-saja bila memang dibutuhkan.

### 2.1 Modul Warga (Citizen Services)

| Modul | Deskripsi | Fitur Utama |
|---|---|---|
| **Auth & Profile** | Akses masuk & manajemen profil aman. | - Login (Nomor Telepon/Email/Kata Sandi).<br>- Autentikasi Biometrik (Face ID/Fingerprint).<br>- Manajemen Keluarga (Lihat data anggota, KK, NIK ter-masking).<br>- Pengajuan Koreksi Data Kependudukan. |
| **Feed & Events** | Portal informasi & pengumuman lingkungan. | - Feed Pengumuman (Filter kategori, lampiran gambar/PDF).<br>- Agenda Kegiatan RT (Jadwal, detail lokasi, penanggung jawab).<br>- Integrasi Kalender Perangkat. |
| **Dues & Payments** | Pengelolaan pembayaran iuran warga. | - Rincian Tagihan Aktif & Tunggakan.<br>- Upload Bukti Pembayaran Manual (Kamera/Galeri).<br>- Transparansi Kas RT (Laporan ringkas kas). |
| **E-Surat** | Pengajuan dokumen administrasi mandiri. | - Form Permohonan Surat Pengantar RT.<br>- Tracking Status Pengajuan Real-time.<br>- Viewer & Unduh Surat Digital (Format PDF + QR Code Validasi). |
| **Aduan & Laporan** | Kanal pelaporan masalah lingkungan. | - Formulir Aduan (Upload foto kamera langsung, koordinat GPS).<br>- Tracking Timeline Penanganan (Tiket aduan). |
| **Notifikasi** | Pusat pemberitahuan personal. | - Inbox notifikasi in-app (indikator belum dibaca, tandai dibaca).<br>- Deep link push ke detail tagihan/surat/aduan. |
| **Keamanan & Ronda** | Fitur proteksi lingkungan & siskamling. | - Panic Button (alert bahaya + GPS, dengan konfirmasi kirim).<br>- Jadwal ronda pribadi, tukar jadwal, absensi via QR pos.<br>- Laporan kejadian patroli.<br>- Undangan tamu berbasis QR berumur pendek. |
| **Kegiatan Warga** | Partisipasi kerja bakti & kegiatan RT. | - Jadwal kerja bakti & status kehadiran.<br>- Absensi kegiatan per KK. |
| **Tabungan** | Dana titipan terarah (mis. Qurban). | - Saldo & mutasi tabungan keluarga.<br>- Lapor setoran + bukti foto.<br>- Pengajuan penarikan sesuai aturan produk. |
| **Aset** | Peminjaman fasilitas milik RT. | - Katalog aset yang dapat dipinjam & ketersediaan.<br>- Pengajuan pinjam, status, dan riwayat sendiri. |
| **Layanan Lingkungan** | Layanan rutin & ekonomi lokal. | - Kalender jadwal pengangkutan sampah + pengingat.<br>- Direktori UMKM warga & pengajuan profil usaha.<br>- Jadwal/pengingat Posyandu (non-medis). |
| **E-Voting** | Partisipasi pemilihan pengurus RT. | - Informasi pemilihan & profil calon.<br>- Voting rahasia satu suara per KK + tanda terima.<br>- Hasil agregat setelah pengesahan. |

### 2.2 Modul Pengurus & Petugas (Field Administration)

| Modul | Deskripsi | Fitur Utama |
|---|---|---|
| **Mobile Approvals** | Persetujuan dokumen & tindak lanjut aduan. | - Approval Surat Pengantar RT sekali ketuk (One-click approve).<br>- Tindak Lanjut Aduan (Update status tiket lapangan, upload foto).<br>- Indikator SLA jatuh tempo/terlambat. |
| **Field Verification** | Validasi data fisik di lapangan. | - Pemindai QR Code (validasi surat warga, check-in tamu, absensi pos ronda).<br>- Cek Pembayaran Tunai & Upload Bukti Setoran di tempat.<br>- Verifikasi setoran/penarikan tabungan warga. |
| **Emergency** | Diseminasi informasi darurat cepat. | - Kirim Notifikasi Darurat (Broadcast push notification massal).<br>- Penerima Isyarat Panic Button Warga + acknowledgement dan penutupan. |
| **Aset & Inventaris** | Pengelolaan aset fisik di lapangan. | - Serah-terima & pengembalian pinjaman aset beserta kondisi fisik.<br>- Pencatatan pemeliharaan langsung di lokasi. |
| **Tata Kelola** | Peralihan wewenang pengurus. | - Checklist serah-terima jabatan & konfirmasi penurunan/penambahan akses. |

---

## 3. Peta Jalan & Backlog Pengembangan (Mobile App Backlog)

### Epic M0: Fondasi & Arsitektur Mobile App
Tujuan: Menyiapkan framework, routing, push engine, storage offline, dan design token.
- [ ] **Task M0.1:** Inisialisasi Mobile App Project Flutter dengan struktur feature-first.
- [ ] **Task M0.2:** Setup Push Notification Service (Firebase Cloud Messaging / APNs).
- [ ] **Task M0.3:** Implementasi `drift` (SQLite) untuk cache data terstruktur dan antrean sinkronisasi terbatas.
- [ ] **Task M0.4:** Integrasi API Perangkat (Permissions manager, Kamera, Galeri, Geolocation/GPS, Biometrik, Kalender).
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

