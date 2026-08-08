import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'asset_provider.dart';

class AssetsScreen extends ConsumerStatefulWidget {
  const AssetsScreen({super.key});

  @override
  ConsumerState<AssetsScreen> createState() => _AssetsScreenState();
}

class _AssetsScreenState extends ConsumerState<AssetsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(assetProvider.notifier).loadAllData());
  }

  @override
  Widget build(BuildContext context) {
    final assetState = ref.watch(assetProvider);

    return DefaultTabController(
      length: 3,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Manajemen Aset RT'),
          bottom: const TabBar(
            tabs: [
              Tab(icon: Icon(Icons.inventory_2), text: 'Daftar Aset'),
              Tab(icon: Icon(Icons.assignment), text: 'Peminjaman'),
              Tab(icon: Icon(Icons.build), text: 'Maintenance'),
            ],
          ),
        ),
        body: assetState.isLoading
            ? const Center(child: CircularProgressIndicator())
            : TabBarView(
                children: [
                  _buildAssetsTab(context, ref, assetState),
                  _buildLoansTab(context, ref, assetState),
                  _buildMaintenanceTab(context, ref, assetState),
                ],
              ),
      ),
    );
  }

  Widget _buildAssetsTab(BuildContext context, WidgetRef ref, AssetState state) {
    return RefreshIndicator(
      onRefresh: () => ref.read(assetProvider.notifier).loadAllData(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (state.error != null) _buildErrorCard(state.error!),
          if (state.assets.isEmpty)
            const Center(child: Text('Belum ada data aset RT.'))
          else
            ...state.assets.map((asset) => Card(
                  margin: const EdgeInsets.only(bottom: 12),
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor: _getStatusColor(asset.status).withOpacity(0.2),
                      child: Icon(Icons.inventory, color: _getStatusColor(asset.status)),
                    ),
                    title: Text(asset.name, style: const TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text(
                      'Kode: ${asset.code} | Kondisi: ${asset.condition.toUpperCase()}\nStatus: ${asset.status.toUpperCase()}',
                    ),
                    trailing: asset.status == 'available'
                        ? ElevatedButton(
                            onPressed: () => _showCreateLoanDialog(context, ref, asset),
                            child: const Text('Pinjam'),
                          )
                        : Chip(
                            label: Text(asset.status),
                            backgroundColor: _getStatusColor(asset.status).withOpacity(0.1),
                          ),
                  ),
                )),
        ],
      ),
    );
  }

  Widget _buildLoansTab(BuildContext context, WidgetRef ref, AssetState state) {
    return RefreshIndicator(
      onRefresh: () => ref.read(assetProvider.notifier).loadAllData(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (state.loans.isEmpty)
            const Center(child: Text('Belum ada riwayat peminjaman.'))
          else
            ...state.loans.map((loan) => Card(
                  margin: const EdgeInsets.only(bottom: 12),
                  child: ListTile(
                    title: Text(loan.assetName ?? 'Aset RT', style: const TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text(
                      'Peminjam: ${loan.borrowerName ?? loan.borrowerId}\nTgl Pinjam: ${loan.loanDate} s/d ${loan.dueDate}\nStatus: ${loan.status.toUpperCase()}',
                    ),
                    trailing: loan.status == 'pending'
                        ? Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              IconButton(
                                icon: const Icon(Icons.check_circle, color: Colors.green),
                                onPressed: () => ref.read(assetProvider.notifier).reviewLoan(loan.id, 'approve'),
                              ),
                              IconButton(
                                icon: const Icon(Icons.cancel, color: Colors.red),
                                onPressed: () => ref.read(assetProvider.notifier).reviewLoan(loan.id, 'reject'),
                              ),
                            ],
                          )
                        : loan.status == 'approved'
                            ? TextButton(
                                onPressed: () => _showReturnDialog(context, ref, loan),
                                child: const Text('Kembalikan'),
                              )
                            : Chip(label: Text(loan.status)),
                  ),
                )),
        ],
      ),
    );
  }

  Widget _buildMaintenanceTab(BuildContext context, WidgetRef ref, AssetState state) {
    return RefreshIndicator(
      onRefresh: () => ref.read(assetProvider.notifier).loadAllData(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (state.maintenances.isEmpty)
            const Center(child: Text('Belum ada catatan pemeliharaan.'))
          else
            ...state.maintenances.map((m) => Card(
                  margin: const EdgeInsets.only(bottom: 12),
                  child: ListTile(
                    leading: const Icon(Icons.build_circle, color: Colors.orange),
                    title: Text(m.maintenanceType, style: const TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text(
                      'Aset: ${m.assetName ?? m.assetId}\nTgl: ${m.maintenanceDate} | Biaya: Rp ${(m.cost ?? 0).toStringAsFixed(0)}\nKondisi Sesudah: ${m.conditionAfter.toUpperCase()}',
                    ),
                  ),
                )),
        ],
      ),
    );
  }

  Widget _buildErrorCard(String error) {
    return Card(
      color: Colors.red.shade50,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Text(error, style: TextStyle(color: Colors.red.shade900)),
      ),
    );
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'available':
        return Colors.green;
      case 'borrowed':
        return Colors.orange;
      case 'maintenance':
        return Colors.blue;
      default:
        return Colors.grey;
    }
  }

  void _showCreateLoanDialog(BuildContext context, WidgetRef ref, AssetItem asset) {
    final loanDateCtrl = TextEditingController(text: DateTime.now().toString().substring(0, 10));
    final dueDateCtrl = TextEditingController(
      text: DateTime.now().add(const Duration(days: 2)).toString().substring(0, 10),
    );
    final notesCtrl = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Pinjam ${asset.name}'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: loanDateCtrl,
              decoration: const InputDecoration(labelText: 'Tgl Pinjam (YYYY-MM-DD)'),
            ),
            TextField(
              controller: dueDateCtrl,
              decoration: const InputDecoration(labelText: 'Tgl Kembali (YYYY-MM-DD)'),
            ),
            TextField(
              controller: notesCtrl,
              decoration: const InputDecoration(labelText: 'Catatan / Keperluan'),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Batal')),
          ElevatedButton(
            onPressed: () async {
              final ok = await ref.read(assetProvider.notifier).createLoan(
                    assetId: asset.id,
                    loanDate: loanDateCtrl.text,
                    dueDate: dueDateCtrl.text,
                    conditionOut: asset.condition,
                    notes: notesCtrl.text.isNotEmpty ? notesCtrl.text : null,
                  );
              if (ok && ctx.mounted) Navigator.pop(ctx);
            },
            child: const Text('Ajukan'),
          ),
        ],
      ),
    );
  }

  void _showReturnDialog(BuildContext context, WidgetRef ref, AssetLoanItem loan) {
    final conditionCtrl = TextEditingController(text: loan.conditionOut);
    final notesCtrl = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Pengembalian Aset'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: conditionCtrl,
              decoration: const InputDecoration(labelText: 'Kondisi Saat Kembali (good/fair/poor/broken)'),
            ),
            TextField(
              controller: notesCtrl,
              decoration: const InputDecoration(labelText: 'Catatan Pengembalian'),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Batal')),
          ElevatedButton(
            onPressed: () async {
              final ok = await ref.read(assetProvider.notifier).returnLoan(
                    loan.id,
                    conditionCtrl.text,
                    notes: notesCtrl.text.isNotEmpty ? notesCtrl.text : null,
                  );
              if (ok && ctx.mounted) Navigator.pop(ctx);
            },
            child: const Text('Proses Kembali'),
          ),
        ],
      ),
    );
  }
}

