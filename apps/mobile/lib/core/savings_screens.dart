import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'savings_provider.dart';

class SavingsScreen extends ConsumerWidget {
  const SavingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final savingsState = ref.watch(savingsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Tabungan Warga'),
      ),
      body: savingsState.isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: () => ref.read(savingsProvider.notifier).fetchSavingsData(),
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  if (savingsState.error != null)
                    Card(
                      color: Colors.red.shade50,
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Text(
                          savingsState.error!,
                          style: TextStyle(color: Colors.red.shade900),
                        ),
                      ),
                    ),
                  const Text(
                    'Akun Tabungan Warga',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  if (savingsState.accounts.isEmpty)
                    const Center(child: Text('Belum ada akun tabungan aktif.'))
                  else
                    ...savingsState.accounts.map((acc) => Card(
                          child: ListTile(
                            title: Text(acc.productName ?? 'Tabungan'),
                            subtitle: Text('Status: ${acc.status}'),
                            trailing: Text(
                              'Rp ${acc.balance.toStringAsFixed(0)}',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Colors.green,
                              ),
                            ),
                          ),
                        )),
                  const SizedBox(height: 24),
                  const Text(
                    'Riwayat Mutasi',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  if (savingsState.transactions.isEmpty)
                    const Center(child: Text('Belum ada transaksi tabungan.'))
                  else
                    ...savingsState.transactions.map((tx) => Card(
                          child: ListTile(
                            leading: Icon(
                              tx.type == 'deposit'
                                  ? Icons.arrow_downward
                                  : Icons.arrow_upward,
                              color: tx.type == 'deposit' ? Colors.green : Colors.red,
                            ),
                            title: Text('${tx.type.toUpperCase()} - ${tx.transactionNumber}'),
                            subtitle: Text(
                              '${tx.description}\nStatus: ${tx.verificationStatus}',
                            ),
                            trailing: Text(
                              'Rp ${tx.amount.toStringAsFixed(0)}',
                              style: const TextStyle(fontWeight: FontWeight.bold),
                            ),
                          ),
                        )),
                ],
              ),
            ),
    );
  }
}

