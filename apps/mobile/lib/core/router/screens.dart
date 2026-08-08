import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../widgets/ui_components.dart';

class WargaHomeScreen extends StatelessWidget {
  const WargaHomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Image.asset(
              'assets/images/logo_256x256.png',
              width: 32,
              height: 32,
            ),
            const SizedBox(width: 8),
            const Text('RT Digital'),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.admin_panel_settings_outlined),
            tooltip: 'Mode Pengurus',
            onPressed: () => context.go('/pengurus/dashboard'),
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Card(
              child: Padding(
                padding: EdgeInsets.all(16),
                child: Row(
                  children: [
                    CircleAvatar(child: Icon(Icons.person)),
                    SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Selamat Datang, Warga', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        Text('RT 05 / RW 02 - Penggilingan', style: TextStyle(color: Colors.grey)),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            const Row(
              children: [
                StatusChip(label: 'Iuran Lunas', type: StatusType.success),
                SizedBox(width: 8),
                StatusChip(label: '1 Aduan Diproses', type: StatusType.warning),
              ],
            ),
            const SizedBox(height: 24),
            AppButton(
              label: 'Buka Pengajuan Surat',
              onPressed: () {},
            ),
          ],
        ),
      ),
    );
  }
}

class WargaLayananScreen extends StatelessWidget {
  const WargaLayananScreen({super.key});
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Layanan Mandiri')),
      body: const AppEmptyState(
        title: 'Katalog Layanan RT',
        description: 'Seluruh permohonan surat, aduan, dan pembayaran dapat diakses melalui menu ini.',
      ),
    );
  }
}

class WargaAktivitasScreen extends StatelessWidget {
  const WargaAktivitasScreen({super.key});
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Aktivitas Saya')),
      body: const AppEmptyState(
        title: 'Belum Ada Aktivitas Terkini',
        description: 'Riwayat pengajuan surat, iuran, dan aduan Anda akan muncul di sini.',
      ),
    );
  }
}

class WargaNotifikasiScreen extends StatelessWidget {
  const WargaNotifikasiScreen({super.key});
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Pemberitahuan')),
      body: const AppEmptyState(
        title: 'Kotak Masuk Kosong',
        description: 'Pengumuman dan pemberitahuan status akan ditampilkan di sini.',
      ),
    );
  }
}

class WargaProfilScreen extends StatelessWidget {
  const WargaProfilScreen({super.key});
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Profil Keluarga')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            const ListTile(
              leading: Icon(Icons.badge),
              title: Text('NIK Ter-masking'),
              subtitle: Text('3175************'),
            ),
            const ListTile(
              leading: Icon(Icons.home),
              title: Text('Nomor KK Ter-masking'),
              subtitle: Text('3175************'),
            ),
            const Spacer(),
            AppButton(
              label: 'Kembali ke Beranda Warga',
              isSecondary: true,
              onPressed: () => context.go('/warga/beranda'),
            ),
          ],
        ),
      ),
    );
  }
}

class PengurusDashboardScreen extends StatelessWidget {
  const PengurusDashboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Persetujuan Menunggu', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
          SizedBox(height: 12),
          AppEmptyState(
            title: 'Tidak Ada Antrean Approval',
            description: 'Semua permohonan surat dan aduan sudah diverifikasi.',
          ),
        ],
      ),
    );
  }
}

class PengurusVerifikasiScreen extends StatelessWidget {
  const PengurusVerifikasiScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Verifikasi QR & Setoran Lapangan', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
          const SizedBox(height: 16),
          AppButton(
            label: 'Mulai Pindai QR Code',
            icon: Icons.qr_code_scanner,
            onPressed: () {},
          ),
        ],
      ),
    );
  }
}
