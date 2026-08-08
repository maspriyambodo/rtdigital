import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'contribution_provider.dart';
import 'auth_provider.dart';
import 'device/device_services.dart';
import 'device/device_services.dart';
import 'widgets/ui_components.dart';

class WargaIuranScreen extends ConsumerWidget {
  const WargaIuranScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(contributionProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Pembayaran & Iuran Warga'),
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: () => ref.read(contributionProvider.notifier).fetchInvoicesAndHistory(),
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  Card(
                    color: Theme.of(context).colorScheme.primaryContainer,
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text('Total Tunggakan Saya', style: TextStyle(fontSize: 14)),
                          const SizedBox(height: 4),
                          Text(
                            'Rp ${state.totalTunggakan}',
                            style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  const Text('Tagihan Berjalan', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  if (state.invoices.isEmpty)
                    const AppEmptyState(title: 'Tidak Ada Tagihan', description: 'Semua iuran RT sudah lunas.')
                  else
                    ...state.invoices.map((inv) => Card(
                          child: ListTile(
                            title: Text(inv.dueTypeName, style: const TextStyle(fontWeight: FontWeight.bold)),
                            subtitle: Text('Jatuh tempo: ${inv.dueDate}'),
                            trailing: Text('Rp ${inv.amount}'),
                            onTap: inv.status == 'unpaid' ? () => _showPaymentOptions(context, ref, inv) : null,
                          ),
                        )),
                  const SizedBox(height: 24),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('Riwayat Pembayaran', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                      TextButton.icon(
                        icon: const Icon(Icons.bar_chart),
                        label: const Text('Transparansi Kas'),
                        onPressed: () => context.push('/warga/kas-rt'),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  if (state.history.isEmpty)
                    const AppEmptyState(title: 'Belum Ada Riwayat', description: 'Transaksi pembayaran yang berhasil akan tampil di sini.')
                  else
                    ...state.history.map((pay) => ListTile(
                          leading: const Icon(Icons.check_circle, color: Colors.green),
                          title: Text('Pembayaran #${pay.id}'),
                          subtitle: Text('Metode: ${pay.paymentMethod}'),
                          trailing: Text('Rp ${pay.amount}'),
                        )),
                ],
              ),
            ),
    );
  }

  void _showPaymentOptions(BuildContext context, WidgetRef ref, InvoiceItem inv) {
    showModalBottomSheet(
      context: context,
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('Bayar ${inv.dueTypeName}', style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            const SizedBox(height: 16),
            ListTile(
              leading: const Icon(Icons.qr_code),
              title: const Text('Bayar via QRIS Dynamic'),
              onTap: () async {
                Navigator.pop(ctx);
                final qris = await ref.read(contributionProvider.notifier).generateQris(inv);
                _showQrisDialog(context, inv, qris);
              },
            ),
            ListTile(
              leading: const Icon(Icons.upload_file),
              title: const Text('Upload Bukti Transfer / Kamera'),
              onTap: () async {
                Navigator.pop(ctx);
                final path = await ref.read(deviceServicesProvider).pickImageFromCamera();
                if (path != null) {
                  await ref.read(contributionProvider.notifier).submitPaymentTransfer(
                        invoiceId: inv.id,
                        amount: inv.amount,
                        proofFileId: 'file_${DateTime.now().millisecondsSinceEpoch}',
                      );
                }
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showQrisDialog(BuildContext context, InvoiceItem inv, String qrisPayload) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('QRIS Pembayaran'),
        content: Text('Jumlah: Rp ${inv.amount}'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Tutup')),
        ],
      ),
    );
  }
}

class WargaKasRtScreen extends StatelessWidget {
  const WargaKasRtScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Laporan Transparansi Kas RT')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Ringkasan Saldo Kas RT 05', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  const SizedBox(height: 12),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: const [
                      Text('Total Pemasukan Bulan Ini:'),
                      Text('Rp 4.500.000', style: TextStyle(color: Colors.green, fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: const [
                      Text('Total Pengeluaran Bulan Ini:'),
                      Text('Rp 1.200.000', style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const Divider(height: 24),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: const [
                      Text('Saldo Akhir Kas:', style: TextStyle(fontWeight: FontWeight.bold)),
                      Text('Rp 12.850.000', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          const Text('Grafik Mutasi Kas (Data Teragregasi)', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          Container(
            height: 140,
            decoration: BoxDecoration(
              color: Colors.blue.shade50,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.blue.shade200),
            ),
            alignment: Alignment.center,
            child: const Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                _BarChartColumn(label: 'Mei', heightRatio: 0.5),
                _BarChartColumn(label: 'Jun', heightRatio: 0.7),
                _BarChartColumn(label: 'Jul', heightRatio: 0.6),
                _BarChartColumn(label: 'Agu', heightRatio: 0.9),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _BarChartColumn extends StatelessWidget {
  final String label;
  final double heightRatio;
  const _BarChartColumn({required this.label, required this.heightRatio});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        Container(
          width: 24,
          height: 80 * heightRatio,
          color: Theme.of(context).colorScheme.primary,
        ),
        const SizedBox(height: 4),
        Text(label, style: const TextStyle(fontSize: 12)),
        const SizedBox(height: 8),
      ],
    );
  }
}
