import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../information_provider.dart';
import '../auth_provider.dart';
import '../widgets/ui_components.dart';

class WargaHomeScreen extends ConsumerStatefulWidget {
  const WargaHomeScreen({super.key});

  @override
  ConsumerState<WargaHomeScreen> createState() => _WargaHomeScreenState();
}

class _WargaHomeScreenState extends ConsumerState<WargaHomeScreen> {
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      ref.read(informationProvider.notifier).fetchAnnouncements();
    }
  }



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
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(informationProvider.notifier).fetchAnnouncements(refresh: true);
          await ref.read(informationProvider.notifier).fetchEvents();
        },
        child: SingleChildScrollView(
          controller: _scrollController,
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [

              Consumer(
                builder: (context, ref, child) {
                  final authState = ref.watch(authProvider);
                  final userName = authState.user?.name ?? 'Warga';
                  return Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Row(
                        children: [
                          const CircleAvatar(child: Icon(Icons.person)),
                          const SizedBox(width: 12),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text('Selamat Datang, $userName', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                              const Text('RT 05 / RW 02 - Penggilingan', style: TextStyle(color: Colors.grey)),
                            ],
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
              const SizedBox(height: 16),
              Consumer(
                builder: (context, ref, child) {
                  final infoState = ref.watch(informationProvider);
                  if (infoState.events.isEmpty) return const SizedBox.shrink();
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text('Agenda RT Mendatang', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                          const StatusChip(label: 'Mendatang', type: StatusType.info),
                        ],
                      ),
                      const SizedBox(height: 8),
                      ...infoState.events.map((evt) => Card(
                            child: ListTile(
                              leading: const Icon(Icons.event_available, color: Colors.green),
                              title: Text(evt.title, style: const TextStyle(fontWeight: FontWeight.bold)),
                              subtitle: Text('${evt.location} • ${evt.startsAt.split('T').first}'),
                              trailing: IconButton(
                                icon: const Icon(Icons.calendar_month),
                                tooltip: 'Simpan ke Kalender HP',
                                onPressed: () async {
                                  final ok = await ref.read(informationProvider.notifier).saveEventToCalendar(evt);
                                  if (context.mounted) {
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      SnackBar(content: Text(ok ? 'Izin kalender disetujui, agenda disimpan' : 'Gagal menyimpan ke kalender')),
                                    );
                                  }
                                },
                              ),
                            ),
                          )),
                      const SizedBox(height: 16),
                    ],
                  );
                },
              ),
              Text('Feed Pengumuman RT', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              Consumer(
                builder: (context, ref, child) {
                  final infoState = ref.watch(informationProvider);
                  final categories = ['Semua', 'Kegiatan', 'Keuangan', 'Keamanan', 'Umum'];
                  return Column(
                    children: [
                      SingleChildScrollView(
                        scrollDirection: Axis.horizontal,
                        child: Row(
                          children: categories.map((cat) {
                            final selected = infoState.categoryFilter == cat || (infoState.categoryFilter == null && cat == 'Semua');
                            return Padding(
                              padding: const EdgeInsets.only(right: 8),
                              child: ChoiceChip(
                                label: Text(cat),
                                selected: selected,
                                onSelected: (val) {
                                  if (val) {
                                    ref.read(informationProvider.notifier).filterCategory(cat == 'Semua' ? null : cat);
                                  }
                                },
                              ),
                            );
                          }).toList(),
                        ),
                      ),
                      const SizedBox(height: 12),
                      if (infoState.announcements.isEmpty && !infoState.isLoading)
                        const AppEmptyState(
                          title: 'Belum Ada Pengumuman',
                          description: 'Tidak ada pengumuman terbaru untuk kategori ini.',
                        )
                      else
                        ...infoState.announcements.map((ann) => Card(
                              margin: const EdgeInsets.only(bottom: 12),
                              child: Padding(
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      children: [
                                        StatusChip(label: ann.category, type: StatusType.info),
                                        Text(ann.publishedAt.split('T').first, style: const TextStyle(color: Colors.grey, fontSize: 12)),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Text(ann.title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                                    const SizedBox(height: 6),
                                    Text(ann.content, style: const TextStyle(fontSize: 14)),
                                    if (ann.attachmentUrl != null) ...[
                                      const SizedBox(height: 12),
                                      OutlinedButton.icon(
                                        icon: Icon(ann.attachmentType == 'pdf' ? Icons.picture_as_pdf : Icons.image),
                                        label: Text('Buka Lampiran (${ann.attachmentType?.toUpperCase()})'),
                                        onPressed: () {
                                          context.push(
                                            '/attachment-viewer?url=${Uri.encodeComponent(ann.attachmentUrl!)}&title=${Uri.encodeComponent(ann.title)}&type=${ann.attachmentType}',
                                          );
                                        },
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                            )),
                      if (infoState.isLoading)
                        const Padding(
                          padding: EdgeInsets.symmetric(vertical: 16),
                          child: Center(child: CircularProgressIndicator()),
                        ),
                    ],
                  );
                },
              ),
            ],
          ),
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
      appBar: AppBar(title: const Text('Layanan Mandiri Warga')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: ListTile(
              leading: const Icon(Icons.payments_outlined, color: Color(0xFF1B5E20)),
              title: const Text('Pembayaran & Iuran Warga', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: const Text('Cek tagihan berjalan, riwayat bayar & QRIS'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/warga/iuran'),
            ),
          ),
          Card(
            child: ListTile(
              leading: const Icon(Icons.account_balance_outlined, color: Color(0xFF1B5E20)),
              title: const Text('Transparansi Kas RT', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: const Text('Laporan saldo kas & grafik pengeluaran'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/warga/kas-rt'),
            ),
          ),
        ],
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
