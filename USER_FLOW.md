# Alur Pengguna Utama

**Status:** Draft untuk validasi  
**Berlaku:** MVP RT Digital  
**Prinsip:** Alur warga dirancang responsive desktop-first. Status, pesan kesalahan, dan tindakan berikutnya harus selalu terlihat. Setiap langkah yang memicu mutasi memakai loading state pada pemicunya sehingga klik ganda tidak menghasilkan pengajuan, pembayaran, atau persetujuan ganda. Notifikasi dikirim melalui aplikasi; email digunakan untuk aktivitas penting bila tersedia.

---

## 1. Aktivasi Akun dan Login Warga

**Aktor:** Sekretaris, Kepala Keluarga  
**Tujuan:** Warga memperoleh akses aman ke aplikasi dan memeriksa data awal keluarga.

1. Sekretaris memasukkan data awal keluarga: rumah/unit, kepala keluarga, dan nomor telepon atau email.
2. Sistem membuat token undangan sekali pakai dengan masa berlaku terbatas.
3. Sistem mengirim tautan atau kode aktivasi kepada Kepala Keluarga.
4. Kepala Keluarga membuka tautan atau halaman aktivasi dari telepon seluler.
5. Kepala Keluarga memverifikasi token, membuat kata sandi, lalu login.
6. Sistem menampilkan profil keluarga untuk diperiksa.
7. Kepala Keluarga memilih:
   - **Data benar:** akun aktif dan warga masuk ke Beranda.
   - **Data perlu diperbaiki:** warga membuat pengajuan koreksi data.
8. Sistem mencatat aktivasi dan login pada audit log.

**Gagal/pengecualian:**
- Token kedaluwarsa atau sudah digunakan: warga meminta undangan baru.
- Akun nonaktif: login ditolak dengan petunjuk menghubungi pengurus.
- Percobaan login gagal berulang: akun dikunci sementara sesuai kebijakan keamanan.

---

## 2. Pembayaran Iuran Transfer dan Verifikasi Bendahara

**Aktor:** Bendahara, Kepala Keluarga atau anggota keluarga yang diizinkan  
**Tujuan:** Tagihan dibayar, diverifikasi, tercatat pada kas, dan status diterima warga.

### 2.1 Pembuatan Tagihan

1. Bendahara membuat jenis iuran dan periode tagihan.
2. Bendahara membuat tagihan massal atau individual.
3. Sistem menerbitkan nomor tagihan dan status `Belum Dibayar`.
4. Sistem mengirim notifikasi tagihan baru kepada warga terkait.

### 2.2 Warga Mengunggah Bukti Pembayaran

1. Warga login lalu membuka menu **Tagihan**.
2. Warga memilih tagihan berstatus `Belum Dibayar` atau `Dibayar Sebagian`.
3. Sistem menampilkan nominal, jatuh tempo, rekening tujuan, dan riwayat pembayaran.
4. Warga melakukan transfer melalui aplikasi bank di luar RT Digital.
5. Warga kembali ke detail tagihan, memilih **Bayar dengan Transfer**.
6. Warga memasukkan nominal dan tanggal pembayaran.
7. Warga mengambil foto bukti transfer dengan kamera atau memilih screenshot dari galeri.
8. Sistem memvalidasi ukuran dan jenis file.
9. Warga menekan **Kirim Bukti Pembayaran**.
10. Sistem menyimpan bukti, mencatat pembayaran, mengubah status tagihan menjadi `Menunggu Verifikasi`, lalu menotifikasi Bendahara.

### 2.3 Bendahara Memverifikasi

1. Bendahara login lalu membuka menu **Pembayaran Menunggu Verifikasi**.
2. Bendahara memilih pembayaran dan melihat detail tagihan, nominal, serta bukti transfer.
3. Bendahara mencocokkan bukti dengan mutasi rekening atau pembayaran nyata.
4. Bendahara memilih:
   - **Verifikasi:** sistem menandai pembayaran tervalidasi.
   - **Tolak:** Bendahara wajib mengisi alasan penolakan.
5. Jika jumlah pembayaran memenuhi tagihan, sistem mengubah tagihan menjadi `Lunas`.
6. Jika jumlah belum memenuhi tagihan, sistem mengubah tagihan menjadi `Dibayar Sebagian`.
7. Sistem otomatis membuat pemasukan Buku Kas untuk pembayaran yang diverifikasi.
8. Sistem membuat audit log untuk verifikasi atau penolakan.
9. Sistem mengirim notifikasi status kepada warga.
10. Untuk pembayaran `Lunas`, warga dapat membuka atau mengunduh tanda terima.

**Aturan:**
- Bendahara tidak dapat memverifikasi, menolak, atau membatalkan pembayaran miliknya sendiri.
- Penolakan tidak menghapus bukti atau riwayat pembayaran.
- Pembatalan pembayaran tervalidasi memerlukan alasan dan audit log.

---

## 3. Pembayaran Iuran Tunai

**Aktor:** Bendahara, Warga  
**Tujuan:** Pembayaran tunai tetap tercatat dengan bukti dan status yang jelas.

1. Warga membayar tunai kepada Bendahara.
2. Bendahara membuka detail tagihan warga.
3. Bendahara memilih **Catat Pembayaran Tunai**.
4. Bendahara memasukkan nominal, tanggal pembayaran, dan catatan bila diperlukan.
5. Sistem mencatat pembayaran, memperbarui status tagihan menjadi `Lunas` atau `Dibayar Sebagian`, serta membuat pemasukan kas.
6. Sistem menghasilkan tanda terima.
7. Warga menerima notifikasi dan dapat melihat tanda terima.
8. Sistem mencatat tindakan pada audit log.

---

## 4. Pengajuan dan Penerbitan Surat

**Aktor:** Warga, Sekretaris, Ketua RT  
**Tujuan:** Warga mengajukan surat administrasi dan menerima PDF surat yang telah diterbitkan.

### 4.1 Pengajuan oleh Warga

1. Warga login lalu membuka **Pengajuan Surat**.
2. Warga memilih **Buat Pengajuan** dan memilih jenis surat.
3. Sistem menampilkan persyaratan, formulir, serta data warga yang dapat diisi otomatis.
4. Warga melengkapi formulir dan mengunggah lampiran dari kamera atau galeri bila disyaratkan.
5. Warga meninjau ringkasan lalu menekan **Ajukan Surat**.
6. Sistem membuat nomor pengajuan, menyimpan data, mengubah status menjadi `Diajukan`, dan menotifikasi Sekretaris.

### 4.2 Pemeriksaan oleh Sekretaris

1. Sekretaris membuka daftar surat berstatus `Diajukan`.
2. Sekretaris memeriksa data dan lampiran.
3. Sekretaris memilih:
   - **Perlu Perbaikan:** wajib memberi catatan; status menjadi `Perlu Perbaikan`.
   - **Lanjutkan Proses:** sistem membuat atau memperbarui draft; status menjadi `Menunggu Persetujuan`.
   - **Tolak:** wajib memberi alasan; status menjadi `Ditolak`.
4. Sistem menotifikasi warga untuk setiap perubahan status.

### 4.3 Persetujuan dan Penerbitan

1. Ketua RT membuka daftar surat `Menunggu Persetujuan`.
2. Ketua RT memeriksa draft surat.
3. Ketua RT memilih:
   - **Setujui:** sistem menerbitkan nomor surat unik dan PDF final; status menjadi `Diterbitkan`.
   - **Kembalikan:** Ketua RT memberi catatan; status kembali ke proses Sekretaris.
   - **Tolak:** Ketua RT memberi alasan; status menjadi `Ditolak`.
4. Sistem mencatat persetujuan atau penolakan pada audit log.
5. Sistem menotifikasi warga.
6. Warga membuka detail pengajuan lalu mengunduh PDF surat yang diterbitkan.

---

## 5. Aduan Warga sampai Penyelesaian

**Aktor:** Warga, Ketua RT atau Sekretaris, Pengurus/Petugas RT  
**Tujuan:** Aduan memiliki nomor tiket, penanggung jawab, perkembangan, dan hasil penyelesaian.

1. Warga login lalu membuka menu **Aduan**.
2. Warga memilih **Buat Aduan Baru**.
3. Warga memilih kategori, mengisi judul, deskripsi, lokasi umum, serta lampiran foto bila ada.
4. Warga menekan **Kirim Aduan**.
5. Sistem membuat nomor tiket, menetapkan status `Baru`, dan menotifikasi pengurus.
6. Ketua RT atau Sekretaris meninjau aduan lalu memilih:
   - **Tugaskan:** menetapkan Pengurus/Petugas RT; status `Ditinjau` atau `Diproses`.
   - **Minta Informasi:** status `Menunggu Informasi`; warga diberi catatan.
   - **Tolak:** wajib memberi alasan; status `Ditolak`.
7. Pengurus/Petugas yang ditugaskan memperbarui status dan menambahkan komentar perkembangan.
8. Warga menerima notifikasi, dapat membaca perkembangan, lalu menambahkan informasi bila diminta.
9. Setelah masalah selesai, petugas menambahkan catatan penyelesaian dan lampiran hasil bila relevan, lalu memilih **Tandai Selesai**.
10. Sistem mengubah status menjadi `Selesai` dan menotifikasi warga.
11. Warga dapat memberi umpan balik sederhana atau menutup aduan.
12. Sistem mencatat penugasan dan perubahan status pada audit log.

---

## 6. Pengajuan Koreksi Data Keluarga

**Aktor:** Kepala Keluarga, Sekretaris  
**Tujuan:** Warga mengoreksi data tanpa mengubah data aktif secara langsung.

1. Kepala Keluarga login lalu membuka **Profil Keluarga** atau **Data Anggota Keluarga**.
2. Kepala Keluarga memilih data yang keliru dan menekan **Ajukan Koreksi**.
3. Sistem menampilkan nilai saat ini dan formulir nilai usulan.
4. Kepala Keluarga memasukkan data baru, alasan koreksi, serta dokumen pendukung bila diperlukan.
5. Kepala Keluarga menekan **Kirim Koreksi**.
6. Sistem menyimpan usulan tanpa mengubah data aktif, menetapkan status `Menunggu Verifikasi`, dan menotifikasi Sekretaris.
7. Sekretaris membandingkan nilai lama, nilai usulan, dan lampiran.
8. Sekretaris memilih:
   - **Setujui:** sistem memperbarui data aktif.
   - **Tolak:** Sekretaris memberi alasan.
   - **Minta Perbaikan:** Sekretaris memberi catatan tambahan.
9. Sistem mencatat keputusan pada audit log dan menotifikasi Kepala Keluarga.

**Aturan:**
- Perubahan NIK, nomor KK, status meninggal, dan status pindah wajib melalui verifikasi pengurus.
- Riwayat nilai sebelum dan sesudah disimpan pada audit log.
- Data sensitif tetap dimasking dalam daftar dan notifikasi.

---

## 7. Status Utama

| Modul | Status |
|---|---|
| Tagihan | `Belum Dibayar`, `Menunggu Verifikasi`, `Dibayar Sebagian`, `Lunas`, `Dibatalkan` |
| Pengajuan surat | `Draft`, `Diajukan`, `Diperiksa`, `Perlu Perbaikan`, `Menunggu Persetujuan`, `Disetujui`, `Diterbitkan`, `Ditolak`, `Dibatalkan` |
| Aduan | `Baru`, `Ditinjau`, `Diproses`, `Menunggu Informasi`, `Selesai`, `Ditolak`, `Ditutup` |
| Koreksi data | `Menunggu Verifikasi`, `Disetujui`, `Ditolak`, `Perlu Perbaikan` |