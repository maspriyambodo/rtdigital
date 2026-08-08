import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:geolocator/geolocator.dart';
import 'complaint_provider.dart';
import 'complaint_provider.dart';
import 'device/device_services.dart';
import 'widgets/ui_components.dart';
import 'theme/app_theme.dart';

class WargaComplaintFormScreen extends ConsumerStatefulWidget {
  const WargaComplaintFormScreen({super.key});

  @override
  ConsumerState<WargaComplaintFormScreen> createState() => _WargaComplaintFormScreenState();
}

class _WargaComplaintFormScreenState extends ConsumerState<WargaComplaintFormScreen> {
  final _formKey = GlobalKey<FormState>();
  String? _selectedCategoryId;
  final TextEditingController _titleController = TextEditingController();
  final TextEditingController _descController = TextEditingController();
  final TextEditingController _locationController = TextEditingController();

  XFile? _capturedImage;
  Position? _currentPosition;
  bool _isGettingLocation = false;

  @override
  void dispose() {
    _titleController.dispose();
    _descController.dispose();
    _locationController.dispose();
    super.dispose();
  }

  Future<void> _takePhoto() async {
    final deviceServices = ref.read(deviceServicesProvider);
    final image = await deviceServices.pickImageFromCamera();
    if (image != null) {
      setState(() {
        _capturedImage = image;
      });
    }
  }

  Future<void> _getLocation() async {
    setState(() {
      _isGettingLocation = true;
    });
    final deviceServices = ref.read(deviceServicesProvider);
    final pos = await deviceServices.getCurrentLocation();
    setState(() {
      _currentPosition = pos;
      _isGettingLocation = false;
    });
    if (pos != null && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Lokasi GPS berhasil didapatkan: ${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)}')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final complaintState = ref.watch(complaintProvider);
    final complaintNotifier = ref.read(complaintProvider.notifier);

    if (_selectedCategoryId == null && complaintState.categories.isNotEmpty) {
      _selectedCategoryId = complaintState.categories.first.id;
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Buat Laporan Aduan'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.space16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Kategori Masalah', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
              const SizedBox(height: AppSpacing.space8),
              DropdownButtonFormField<String>(
                value: _selectedCategoryId,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                ),
                items: complaintState.categories.map((c) {
                  return DropdownMenuItem(value: c.id, child: Text(c.name));
                }).toList(),
                onChanged: (val) {
                  setState(() {
                    _selectedCategoryId = val;
                  });
                },
              ),
              const SizedBox(height: AppSpacing.space16),

              AppTextField(
                label: 'Judul Laporan *',
                hint: 'Ringkasan singkat masalah',
                controller: _titleController,
                validator: (val) => val == null || val.trim().isEmpty ? 'Judul laporan wajib diisi' : null,
              ),
              const SizedBox(height: AppSpacing.space16),

              AppTextField(
                label: 'Deskripsi Kejadian *',
                hint: 'Jelaskan detail masalah yang Anda temukan...',
                controller: _descController,
                validator: (val) => val == null || val.trim().isEmpty ? 'Deskripsi wajib diisi' : null,
              ),
              const SizedBox(height: AppSpacing.space16),

              AppTextField(
                label: 'Petunjuk Alamat / Lokasi Lapangan *',
                hint: 'Contoh: Depan Rumah No. 12 RT 05',
                controller: _locationController,
                validator: (val) => val == null || val.trim().isEmpty ? 'Petunjuk lokasi wajib diisi' : null,
              ),
              const SizedBox(height: AppSpacing.space20),

              Text('Foto Bukti Lapangan (Kamera)', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
              const SizedBox(height: AppSpacing.space8),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(AppSpacing.space12),
                  child: Row(
                    children: [
                      Icon(
                        _capturedImage != null ? Icons.check_circle : Icons.camera_alt,
                        color: _capturedImage != null ? AppTheme.successColor : Colors.grey,
                        size: 32,
                      ),
                      const SizedBox(width: AppSpacing.space12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _capturedImage != null ? 'Foto Terambil' : 'Belum Ada Foto',
                              style: const TextStyle(fontWeight: FontWeight.bold),
                            ),
                            Text(
                              _capturedImage != null ? _capturedImage!.name : 'Ambil foto kondisi di lokasi',
                              style: const TextStyle(fontSize: 12, color: Colors.grey),
                            ),
                          ],
                        ),
                      ),
                      OutlinedButton.icon(
                        icon: const Icon(Icons.camera_alt, size: 18),
                        label: Text(_capturedImage != null ? 'Ulangi' : 'Foto'),
                        onPressed: _takePhoto,
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.space20),

              Text('Geotagging Koordinat GPS', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
              const SizedBox(height: AppSpacing.space8),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(AppSpacing.space12),
                  child: Row(
                    children: [
                      Icon(
                        _currentPosition != null ? Icons.my_location : Icons.location_off,
                        color: _currentPosition != null ? AppTheme.primaryColor : Colors.grey,
                        size: 32,
                      ),
                      const SizedBox(width: AppSpacing.space12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _currentPosition != null
                                  ? '${_currentPosition!.latitude.toStringAsFixed(5)}, ${_currentPosition!.longitude.toStringAsFixed(5)}'
                                  : 'Lokasi GPS Belum Ditentukan',
                              style: const TextStyle(fontWeight: FontWeight.bold),
                            ),
                            const Text('Geotagging otomatis dari GPS HP', style: TextStyle(fontSize: 12, color: Colors.grey)),
                          ],
                        ),
                      ),
                      OutlinedButton.icon(
                        icon: const Icon(Icons.gps_fixed, size: 18),
                        label: Text(_isGettingLocation ? 'Ambil...' : 'Ambil GPS'),
                        onPressed: _isGettingLocation ? null : _getLocation,
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.space24),

              AppButton(
                label: 'Kirim Aduan Lapangan',
                icon: Icons.send_rounded,
                isLoading: complaintState.isLoading,
                onPressed: () async {
                  if (_formKey.currentState?.validate() ?? false) {
                    final success = await complaintNotifier.createComplaint(
                      categoryId: _selectedCategoryId ?? 'cat-lainnya',
                      title: _titleController.text.trim(),
                      description: _descController.text.trim(),
                      locationDescription: _locationController.text.trim(),
                      latitude: _currentPosition?.latitude,
                      longitude: _currentPosition?.longitude,
                      photoPath: _capturedImage?.path ?? 'https://example.com/aduan-kamera.jpg',
                    );

                    if (context.mounted && success) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Laporan aduan berhasil dikirim!')),
                      );
                      context.go('/warga/aduan/list');
                    }
                  }
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class WargaComplaintListScreen extends ConsumerWidget {
  const WargaComplaintListScreen({super.key});

  StatusType _getStatusType(String status) {
    switch (status) {
      case 'resolved':
      case 'closed':
        return StatusType.success;
      case 'in_process':
        return StatusType.info;
      case 'submitted':
        return StatusType.warning;
      case 'rejected':
        return StatusType.error;
      default:
        return StatusType.info;
    }
  }

  String _getStatusLabel(String status) {
    switch (status) {
      case 'submitted':
        return 'Diajukan';
      case 'in_process':
        return 'Sedang Diproses';
      case 'resolved':
        return 'Selesai Penanganan';
      case 'closed':
        return 'Ditutup';
      case 'rejected':
        return 'Ditolak';
      default:
        return status;
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final complaintState = ref.watch(complaintProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Daftar Aduan Saya'),
      ),
      floatingActionButton: FloatingActionButton.extended(
        icon: const Icon(Icons.add),
        label: const Text('Buat Aduan'),
        onPressed: () => context.push('/warga/aduan/baru'),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(complaintProvider.notifier).fetchComplaints();
        },
        child: complaintState.complaints.isEmpty
            ? const AppEmptyState(
                title: 'Belum Ada Aduan Lapangan',
                description: 'Laporan aduan yang Anda kirim akan melacak timeline penanganannya di sini.',
              )
            : ListView.builder(
                padding: const EdgeInsets.all(AppSpacing.space16),
                itemCount: complaintState.complaints.length,
                itemBuilder: (context, index) {
                  final item = complaintState.complaints[index];
                  return Card(
                    margin: const EdgeInsets.only(bottom: AppSpacing.space12),
                    child: ListTile(
                      contentPadding: const EdgeInsets.all(AppSpacing.space16),
                      title: Row(
                        children: [
                          Expanded(
                            child: Text(item.title, style: const TextStyle(fontWeight: FontWeight.bold)),
                          ),
                          StatusChip(
                            label: _getStatusLabel(item.status),
                            type: _getStatusType(item.status),
                          ),
                        ],
                      ),
                      subtitle: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const SizedBox(height: AppSpacing.space8),
                          Text('No. Tiket: ${item.ticketNumber}', style: const TextStyle(color: Colors.grey, fontSize: 13)),
                          Text('Kategori: ${item.categoryName}', style: const TextStyle(fontSize: 13)),
                          Text('Lokasi: ${item.locationDescription}', style: const TextStyle(fontSize: 13)),
                          if (item.assignedToName != null)
                            Text('Petugas: ${item.assignedToName}', style: const TextStyle(fontWeight: FontWeight.w600, color: AppTheme.primaryColor)),
                        ],
                      ),
                      onTap: () {
                        context.push('/warga/aduan/timeline?id=${item.id}');
                      },
                    ),
                  );
                },
              ),
      ),
    );
  }
}

class WargaComplaintTimelineScreen extends ConsumerStatefulWidget {
  final String complaintId;

  const WargaComplaintTimelineScreen({super.key, required this.complaintId});

  @override
  ConsumerState<WargaComplaintTimelineScreen> createState() => _WargaComplaintTimelineScreenState();
}

class _WargaComplaintTimelineScreenState extends ConsumerState<WargaComplaintTimelineScreen> {
  final TextEditingController _commentController = TextEditingController();

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final complaintState = ref.watch(complaintProvider);
    final complaint = complaintState.complaints.firstWhere(
      (c) => c.id == widget.complaintId,
      orElse: () => complaintState.complaints.first,
    );

    return Scaffold(
      appBar: AppBar(
        title: Text('Tiket #${complaint.ticketNumber}'),
      ),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(AppSpacing.space16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(AppSpacing.space16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(complaint.title, style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                          const SizedBox(height: AppSpacing.space8),
                          Text('Kategori: ${complaint.categoryName}', style: const TextStyle(color: Colors.grey)),
                          Text('Lokasi: ${complaint.locationDescription}'),
                          if (complaint.latitude != null && complaint.longitude != null)
                            Text('GPS: ${complaint.latitude}, ${complaint.longitude}', style: const TextStyle(fontSize: 12, color: AppTheme.primaryColor)),
                          const SizedBox(height: AppSpacing.space12),
                          Text(complaint.description),
                          if (complaint.photoUrl != null) ...[
                            const SizedBox(height: AppSpacing.space12),
                            ClipRRect(
                              borderRadius: BorderRadius.circular(AppRadius.card),
                              child: Container(
                                height: 160,
                                width: double.infinity,
                                color: Colors.grey.shade200,
                                child: const Icon(Icons.image, size: 64, color: Colors.grey),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.space20),
                  Text('Timeline & Perkembangan Lapangan', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                  const SizedBox(height: AppSpacing.space12),

                  if (complaint.comments.isEmpty)
                    const AppEmptyState(
                      title: 'Belum Ada Perkembangan Baru',
                      description: 'Petugas sedang menindaklanjuti laporan aduan ini di lapangan.',
                    )
                  else
                    ...complaint.comments.map((comment) {
                      return Card(
                        margin: const EdgeInsets.only(bottom: AppSpacing.space12),
                        child: Padding(
                          padding: const EdgeInsets.all(AppSpacing.space12),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  Text('${comment.authorName} (${comment.authorRole})', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                                  Text(comment.createdAt.split('T').first, style: const TextStyle(color: Colors.grey, fontSize: 11)),
                                ],
                              ),
                              const SizedBox(height: 6),
                              Text(comment.comment),
                            ],
                          ),
                        ),
                      );
                    }),
                ],
              ),
            ),
          ),
          Container(
            padding: const EdgeInsets.all(AppSpacing.space12),
            decoration: BoxDecoration(
              color: Theme.of(context).cardColor,
              boxShadow: [
                BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 4, offset: const Offset(0, -2)),
              ],
            ),
            child: Row(
              children: [
                Expanded(
                  child: AppTextField(
                    label: '',
                    hint: 'Tulis tanggapan atau komentar...',
                    controller: _commentController,
                  ),
                ),
                const SizedBox(width: AppSpacing.space8),
                IconButton.filled(
                  icon: const Icon(Icons.send),
                  onPressed: () async {
                    if (_commentController.text.trim().isNotEmpty) {
                      final text = _commentController.text.trim();
                      _commentController.clear();
                      await ref.read(complaintProvider.notifier).addComment(complaint.id, text);
                    }
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class PengurusComplaintDashboardScreen extends ConsumerStatefulWidget {
  const PengurusComplaintDashboardScreen({super.key});

  @override
  ConsumerState<PengurusComplaintDashboardScreen> createState() => _PengurusComplaintDashboardScreenState();
}

class _PengurusComplaintDashboardScreenState extends ConsumerState<PengurusComplaintDashboardScreen> {
  final TextEditingController _officerController = TextEditingController();
  final TextEditingController _commentController = TextEditingController();

  @override
  void dispose() {
    _officerController.dispose();
    _commentController.dispose();
    super.dispose();
  }

  void _showAssignDialog(BuildContext context, ComplaintItem complaint) {
    _officerController.text = complaint.assignedToName ?? '';
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Disposisi Petugas Lapangan'),
        content: AppTextField(
          label: 'Nama Petugas Lapangan',
          hint: 'Contoh: Pak Joko (Teknik)',
          controller: _officerController,
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Batal')),
          FilledButton(
            onPressed: () {
              if (_officerController.text.trim().isNotEmpty) {
                ref.read(complaintProvider.notifier).assignOfficer(complaint.id, _officerController.text.trim());
                Navigator.pop(ctx);
              }
            },
            child: const Text('Simpan Disposisi'),
          ),
        ],
      ),
    );
  }

  void _showUpdateStatusDialog(BuildContext context, ComplaintItem complaint) {
    _commentController.clear();
    String selectedStatus = 'in_process';
    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Update Status & Perkembangan'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField<String>(
                value: selectedStatus,
                items: const [
                  DropdownMenuItem(value: 'in_process', child: Text('Sedang Diproses')),
                  DropdownMenuItem(value: 'resolved', child: Text('Selesai Penanganan')),
                  DropdownMenuItem(value: 'rejected', child: Text('Ditolak')),
                ],
                onChanged: (val) {
                  if (val != null) setDialogState(() => selectedStatus = val);
                },
              ),
              const SizedBox(height: 12),
              AppTextField(
                label: 'Catatan Penanganan',
                hint: 'Catatan tindakan petugas di lapangan...',
                controller: _commentController,
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Batal')),
            FilledButton(
              onPressed: () {
                ref.read(complaintProvider.notifier).updateStatus(
                  complaint.id,
                  selectedStatus,
                  comment: _commentController.text.trim().isNotEmpty ? _commentController.text.trim() : null,
                );
                Navigator.pop(ctx);
              },
              child: const Text('Simpan Status'),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final complaintState = ref.watch(complaintProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Dashboard Tiket Aduan Masuk'),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(complaintProvider.notifier).fetchComplaints();
        },
        child: complaintState.complaints.isEmpty
            ? const AppEmptyState(
                title: 'Tidak Ada Tiket Aduan',
                description: 'Seluruh aduan warga telah selesai ditangani.',
              )
            : ListView.builder(
                padding: const EdgeInsets.all(AppSpacing.space16),
                itemCount: complaintState.complaints.length,
                itemBuilder: (context, index) {
                  final complaint = complaintState.complaints[index];
                  return Card(
                    margin: const EdgeInsets.only(bottom: AppSpacing.space16),
                    child: Padding(
                      padding: const EdgeInsets.all(AppSpacing.space16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(complaint.title, style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                              ),
                              StatusChip(
                                label: complaint.status,
                                type: complaint.status == 'resolved' ? StatusType.success : StatusType.warning,
                              ),
                            ],
                          ),
                          const SizedBox(height: AppSpacing.space8),
                          Text('Pelapor: ${complaint.reporterName} | No: ${complaint.ticketNumber}'),
                          Text('Lokasi: ${complaint.locationDescription}'),
                          if (complaint.assignedToName != null)
                            Text('Disposisi: ${complaint.assignedToName}', style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.primaryColor)),
                          const SizedBox(height: AppSpacing.space16),
                          Row(
                            children: [
                              Expanded(
                                child: OutlinedButton.icon(
                                  icon: const Icon(Icons.person_add, size: 16),
                                  label: const Text('Disposisi'),
                                  onPressed: () => _showAssignDialog(context, complaint),
                                ),
                              ),
                              const SizedBox(width: AppSpacing.space8),
                              Expanded(
                                child: FilledButton.icon(
                                  icon: const Icon(Icons.edit_note, size: 16),
                                  label: const Text('Update Status'),
                                  onPressed: () => _showUpdateStatusDialog(context, complaint),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
      ),
    );
  }
}
