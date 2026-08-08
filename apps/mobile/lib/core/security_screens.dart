import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'security_provider.dart';
import 'complaint_provider.dart';
import 'widgets/ui_components.dart';
import 'theme/app_theme.dart';

class PanicButtonScreen extends ConsumerStatefulWidget {
  const PanicButtonScreen({super.key});

  @override
  ConsumerState<PanicButtonScreen> createState() => _PanicButtonScreenState();
}

class _PanicButtonScreenState extends ConsumerState<PanicButtonScreen> {
  String _selectedCategory = 'kebakaran';
  final TextEditingController _locationController = TextEditingController();
  Position? _currentPosition;
  bool _isGettingLocation = false;
  bool _alertSent = false;

  @override
  void dispose() {
    _locationController.dispose();
    super.dispose();
  }

  Future<void> _fetchGPS() async {
    setState(() => _isGettingLocation = true);
    final deviceServices = ref.read(deviceServicesProvider);
    final pos = await deviceServices.getCurrentLocation();
    setState(() {
      _currentPosition = pos;
      _isGettingLocation = false;
    });
  }

  void _showConfirmationDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: Colors.red, size: 28),
            SizedBox(width: 8),
            Text('KONFIRMASI DARURAT'),
          ],
        ),
        content: Text(
          'Apakah Anda yakin ingin menyebarkan Sinyal Darurat ($_selectedCategory) ke seluruh pengurus RT & petugas siskamling?',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Batal')),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () async {
              Navigator.pop(ctx);
              final success = await ref.read(securityProvider.notifier).sendPanicAlert(
                category: _selectedCategory,
                locationDescription: _locationController.text.trim().isNotEmpty
                    ? _locationController.text.trim()
                    : 'RT 05 (Lokasi Warga)',
                latitude: _currentPosition?.latitude,
                longitude: _currentPosition?.longitude,
              );
              if (success && mounted) {
                setState(() => _alertSent = true);
              }
            },
            child: const Text('KIRIM DARURAT NOW'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final securityState = ref.watch(securityProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Tombol Darurat (Panic Button)'),
        backgroundColor: Colors.red.shade900,
        foregroundColor: Colors.white,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.space16),
        child: _alertSent
            ? Column(
                children: [
                  const SizedBox(height: 32),
                  const Icon(Icons.check_circle_outline, color: Colors.green, size: 96),
                  const SizedBox(height: 16),
                  Text('SINYAL DARURAT TERKIRIM!', style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold, color: Colors.red.shade900)),
                  const SizedBox(height: 8),
                  const Text('Sinyal dan lokasi GPS Anda telah disiarkan ke Handphone Pengurus RT & Petugas Ronda.'),
                  const SizedBox(height: 24),
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Kategori: $_selectedCategory', style: const TextStyle(fontWeight: FontWeight.bold)),
                          Text('Alamat: ${_locationController.text.isNotEmpty ? _locationController.text : "RT 05"}'),
                          if (_currentPosition != null)
                            Text('Koordinat GPS: ${_currentPosition!.latitude}, ${_currentPosition!.longitude}'),
                          const SizedBox(height: 8),
                          const StatusChip(label: 'Status: Menunggu Respon Petugas', type: StatusType.warning),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 32),
                  AppButton(
                    label: 'Kembali ke Beranda',
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              )
            : Column(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  const SizedBox(height: 16),
                  const Text(
                    'Tekan Tombol Darurat untuk Memanggil Bantuan Siskamling',
                    textAlign: TextAlign.center,
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                  const SizedBox(height: 24),
                  GestureDetector(
                    onTap: _showConfirmationDialog,
                    child: Container(
                      width: 180,
                      height: 180,
                      decoration: BoxDecoration(
                        color: Colors.red,
                        shape: BoxShape.circle,
                        boxShadow: [
                          BoxShadow(color: Colors.red.withOpacity(0.4), blurRadius: 20, spreadRadius: 5),
                        ],
                      ),
                      child: const Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.touch_app, size: 64, color: Colors.white),
                          SizedBox(height: 8),
                          Text('PANIC', style: TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold)),
                          Text('TEKAN DARURAT', style: TextStyle(color: Colors.white, fontSize: 11)),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 32),

                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(AppSpacing.space16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Pilih Jenis Kejadian Darurat:', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                          const SizedBox(height: 8),
                          DropdownButtonFormField<String>(
                            value: _selectedCategory,
                            decoration: const InputDecoration(border: OutlineInputBorder()),
                            items: const [
                              DropdownMenuItem(value: 'kebakaran', child: Text('🔥 Kebakaran')),
                              DropdownMenuItem(value: 'kejahatan', child: Text('🚨 Kejahatan / Maling')),
                              DropdownMenuItem(value: 'medis', child: Text('🚑 Darurat Medis / Ambulan')),
                              DropdownMenuItem(value: 'bencana', child: Text('🌊 Bencana Alam / Banjir')),
                            ],
                            onChanged: (val) {
                              if (val != null) setState(() => _selectedCategory = val);
                            },
                          ),
                          const SizedBox(height: 16),

                          AppTextField(
                            label: 'Petunjuk Lokasi Singkat',
                            hint: 'Contoh: Rumah Pak RT / Blok B No. 4',
                            controller: _locationController,
                          ),
                          const SizedBox(height: 12),

                          Row(
                            children: [
                              Icon(_currentPosition != null ? Icons.gps_fixed : Icons.gps_not_fixed, color: AppTheme.primaryColor),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  _currentPosition != null
                                      ? 'GPS: ${_currentPosition!.latitude.toStringAsFixed(4)}, ${_currentPosition!.longitude.toStringAsFixed(4)}'
                                      : 'Lampirkan koordinat GPS akurat',
                                  style: const TextStyle(fontSize: 12),
                                ),
                              ),
                              TextButton(
                                onPressed: _isGettingLocation ? null : _fetchGPS,
                                child: Text(_isGettingLocation ? 'Mengambil...' : 'Ambil GPS'),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
      ),
    );
  }
}

class SecurityDashboardAlertsScreen extends ConsumerWidget {
  const SecurityDashboardAlertsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final securityState = ref.watch(securityProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Dashboard Alert Panic & Ronda'),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(securityProvider.notifier).fetchAlerts();
          await ref.read(securityProvider.notifier).fetchPatrolAttendances();
        },
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.space16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('🚨 Sinyal Darurat Masuk (Panic Alerts)', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
              const SizedBox(height: 12),
              if (securityState.activeAlerts.isEmpty)
                const AppEmptyState(
                  title: 'Situasi Kondusif',
                  description: 'Tidak ada sinyal darurat aktif saat ini di lingkungan RT.',
                )
              else
                ...securityState.activeAlerts.map((alert) {
                  return Card(
                    margin: const EdgeInsets.only(bottom: 12),
                    color: alert.status == 'active' ? Colors.red.shade50 : null,
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Text('Kejadian: ${alert.category.toUpperCase()}', style: TextStyle(fontWeight: FontWeight.bold, color: alert.status == 'active' ? Colors.red.shade900 : Colors.black)),
                              const Spacer(),
                              StatusChip(
                                label: alert.status,
                                type: alert.status == 'active' ? StatusType.error : StatusType.success,
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          Text('Pelapor: ${alert.reporterName}'),
                          Text('Lokasi: ${alert.locationDescription}'),
                          if (alert.latitude != null)
                            Text('GPS: ${alert.latitude}, ${alert.longitude}', style: const TextStyle(fontSize: 12, color: Colors.blue)),
                          if (alert.acknowledgedBy != null)
                            Text('Ditanggapi oleh: ${alert.acknowledgedBy}', style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.primaryColor)),
                          const SizedBox(height: 12),
                          Row(
                            children: [
                              if (alert.status == 'active')
                                Expanded(
                                  child: OutlinedButton(
                                    onPressed: () {
                                      ref.read(securityProvider.notifier).acknowledgeAlert(alert.id, 'Pak Joko (Petugas Ronda)');
                                    },
                                    child: const Text('Respon & Datangi'),
                                  ),
                                ),
                              if (alert.status == 'active') const SizedBox(width: 8),
                              Expanded(
                                child: FilledButton(
                                  onPressed: () {
                                    ref.read(securityProvider.notifier).resolveAlert(alert.id);
                                  },
                                  child: const Text('Tutup Selesai'),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  );
                }),
              const SizedBox(height: 24),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('🛡️ Rekap Absensi Pos Ronda', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                  IconButton(
                    icon: const Icon(Icons.qr_code_scanner),
                    tooltip: 'Check-in Pos Ronda',
                    onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const PatrolScanQRScreen())),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              if (securityState.patrolAttendances.isEmpty)
                const AppEmptyState(title: 'Belum Ada Absensi Ronda', description: 'Scan QR di pos ronda untuk mencatat kehadiran.')
              else
                ...securityState.patrolAttendances.map((patrol) {
                  return Card(
                    margin: const EdgeInsets.only(bottom: 8),
                    child: ListTile(
                      leading: const Icon(Icons.security, color: AppTheme.primaryColor),
                      title: Text(patrol.postName, style: const TextStyle(fontWeight: FontWeight.bold)),
                      subtitle: Text('Petugas: ${patrol.officerName} | ${patrol.checkinTime.split("T").first}'),
                      trailing: const StatusChip(label: 'Hadir Valid', type: StatusType.success),
                    ),
                  );
                }),
            ],
          ),
        ),
      ),
    );
  }
}

class PatrolScanQRScreen extends StatefulWidget {
  const PatrolScanQRScreen({super.key});

  @override
  State<PatrolScanQRScreen> createState() => _PatrolScanQRScreenState();
}

class _PatrolScanQRScreenState extends State<PatrolScanQRScreen> {
  bool _scanned = false;

  @override
  Widget build(BuildContext context) {
    return Consumer(
      builder: (context, ref, child) => Scaffold(
        appBar: AppBar(
          title: const Text('Scan QR Code Pos Ronda'),
        ),
        body: MobileScanner(
          onDetect: (capture) async {
            if (_scanned) return;
            final List<Barcode> barcodes = capture.barcodes;
            for (final barcode in barcodes) {
              if (barcode.rawValue != null) {
                setState(() => _scanned = true);
                final code = barcode.rawValue!;
                await ref.read(securityProvider.notifier).checkinPatrolQR(code);
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Absensi Pos Ronda Berhasil: $code')),
                  );
                  Navigator.pop(context);
                }
                break;
              }
            }
          },
        ),
      ),
    );
  }
}
