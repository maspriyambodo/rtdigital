import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'securityops_provider.dart';
import 'widgets/ui_components.dart';

class SecurityOpsScreen extends ConsumerStatefulWidget {
  const SecurityOpsScreen({super.key});

  @override
  ConsumerState<SecurityOpsScreen> createState() => _SecurityOpsScreenState();
}

class _SecurityOpsScreenState extends ConsumerState<SecurityOpsScreen> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _showPanicDialog() {
    final detailsController = TextEditingController();
    String selectedCategory = 'fire';

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Kirim Panggilan Darurat (SOS)'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            DropdownButtonFormField<String>(
              value: selectedCategory,
              decoration: const InputDecoration(labelText: 'Kategori Keadaan'),
              items: const [
                DropdownMenuItem(value: 'fire', child: Text('Kebakaran')),
                DropdownMenuItem(value: 'medical', child: Text('Medis / Sakit')),
                DropdownMenuItem(value: 'crime', child: Text('Kriminal / Pencurian')),
                DropdownMenuItem(value: 'accident', child: Text('Kecelakaan')),
                DropdownMenuItem(value: 'other', child: Text('Lainnya')),
              ],
              onChanged: (val) {
                if (val != null) selectedCategory = val;
              },
            ),
            const SizedBox(height: 12),
            TextField(
              controller: detailsController,
              decoration: const InputDecoration(
                labelText: 'Detail Lokasi / Kejadian',
                hintText: 'Misal: Blok A3 No 12',
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red.shade700, foregroundColor: Colors.white),
            onPressed: () async {
              Navigator.pop(ctx);
              final success = await ref.read(securityOpsProvider.notifier).triggerEmergencyAlert(
                selectedCategory,
                detailsController.text.trim(),
              );
              if (mounted && success) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Panggilan darurat berhasil dikirim!')),
                );
              }
            },
            child: const Text('KIRIM DARURAT'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(securityOpsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Keamanan & Kegiatan Warga'),
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: const [
            Tab(icon: Icon(Icons.shield_outlined), text: 'Ronda & Pos'),
            Tab(icon: Icon(Icons.warning_amber_rounded), text: 'Panggilan Darurat'),
            Tab(icon: Icon(Icons.badge_outlined), text: 'Buku Tamu'),
            Tab(icon: Icon(Icons.event_outlined), text: 'Kegiatan Warga'),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _showPanicDialog,
        backgroundColor: Colors.red.shade700,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.sos),
        label: const Text('PANIC SOS'),
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : TabBarView(
              controller: _tabController,
              children: [
                _buildPostsTab(state),
                _buildAlertsTab(state),
                _buildVisitorsTab(state),
                _buildActivitiesTab(state),
              ],
            ),
    );
  }

  Widget _buildPostsTab(SecurityOpsState state) {
    if (state.posts.isEmpty) {
      return const AppEmptyState(title: 'Kosong', description: 'Belum ada pos ronda terdaftar.');
    }
    return ListView.builder(
      itemCount: state.posts.length,
      itemBuilder: (context, index) {
        final item = state.posts[index];
        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: ListTile(
            leading: const CircleAvatar(child: Icon(Icons.local_police)),
            title: Text(item.name, style: const TextStyle(fontWeight: FontWeight.bold)),
            subtitle: Text('Kode: ${item.code} | Lokasi: ${item.location ?? "-"}'),
            trailing: Chip(
              label: Text(item.status.toUpperCase()),
              backgroundColor: Colors.green.shade100,
            ),
          ),
        );
      },
    );
  }

  Widget _buildAlertsTab(SecurityOpsState state) {
    if (state.alerts.isEmpty) {
      return const AppEmptyState(title: 'Kosong', description: 'Tidak ada panggilan darurat aktif.');
    }
    return ListView.builder(
      itemCount: state.alerts.length,
      itemBuilder: (context, index) {
        final item = state.alerts[index];
        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: Colors.red.shade100,
              child: Icon(Icons.warning, color: Colors.red.shade700),
            ),
            title: Text('Kategori: ${item.category.toUpperCase()}'),
            subtitle: Text('Lokasi: ${item.locationDetails ?? "-"}'),
            trailing: Chip(
              label: Text(item.status.toUpperCase()),
              backgroundColor: item.status == 'active' ? Colors.red.shade100 : Colors.blue.shade100,
            ),
          ),
        );
      },
    );
  }

  Widget _buildVisitorsTab(SecurityOpsState state) {
    return state.visitorLogs.isEmpty
        ? const AppEmptyState(title: 'Kosong', description: 'Belum ada catatan kunjungan tamu.')
        : ListView.builder(
            itemCount: state.visitorLogs.length,
            itemBuilder: (context, index) {
              final item = state.visitorLogs[index];
              return Card(
                margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                child: ListTile(
                  leading: const CircleAvatar(child: Icon(Icons.person)),
                  title: Text(item.visitorName, style: const TextStyle(fontWeight: FontWeight.bold)),
                  subtitle: Text('Tujuan: ${item.purpose ?? "-"} | Plat: ${item.vehiclePlate ?? "-"}'),
                  trailing: Text(item.status),
                ),
              );
            },
          );
  }

  Widget _buildActivitiesTab(SecurityOpsState state) {
    if (state.activities.isEmpty) {
      return const AppEmptyState(title: 'Kosong', description: 'Belum ada jadwal kegiatan warga.');
    }
    return ListView.builder(
      itemCount: state.activities.length,
      itemBuilder: (context, index) {
        final item = state.activities[index];
        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: ListTile(
            leading: CircleAvatar(
              child: Icon(item.isMandatory ? Icons.star : Icons.event),
            ),
            title: Text(item.title, style: const TextStyle(fontWeight: FontWeight.bold)),
            subtitle: Text('Tanggal: ${item.activityDate} | Jam: ${item.startTime}'),
            trailing: Chip(
              label: Text(item.status.toUpperCase()),
              backgroundColor: Colors.blue.shade100,
            ),
          ),
        );
      },
    );
  }
}
