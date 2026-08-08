import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'letter_provider.dart';
import 'widgets/ui_components.dart';
import 'theme/app_theme.dart';

class WargaLetterFormScreen extends ConsumerStatefulWidget {
  const WargaLetterFormScreen({super.key});

  @override
  ConsumerState<WargaLetterFormScreen> createState() => _WargaLetterFormScreenState();
}

class _WargaLetterFormScreenState extends ConsumerState<WargaLetterFormScreen> {
  final _formKey = GlobalKey<FormState>();
  String? _selectedTypeId;
  final Map<String, TextEditingController> _controllers = {};
  final TextEditingController _noteController = TextEditingController();
  final Map<String, bool> _uploadedRequirements = {};

  @override
  void dispose() {
    for (var c in _controllers.values) {
      c.dispose();
    }
    _noteController.dispose();
    super.dispose();
  }

  void _initFormForType(LetterTypeItem type) {
    _controllers.clear();
    for (var field in type.formSchema) {
      _controllers[field.key] = TextEditingController();
    }
    _uploadedRequirements.clear();
    for (var req in type.requirements) {
      _uploadedRequirements[req.code] = false;
    }
  }

  @override
  Widget build(BuildContext context) {
    final letterState = ref.watch(letterProvider);
    final letterNotifier = ref.read(letterProvider.notifier);

    LetterTypeItem? selectedType;
    if (_selectedTypeId != null) {
      selectedType = letterState.letterTypes.firstWhere(
        (t) => t.id == _selectedTypeId,
        orElse: () => letterState.letterTypes.first,
      );
    } else if (letterState.letterTypes.isNotEmpty) {
      selectedType = letterState.letterTypes.first;
      _selectedTypeId = selectedType.id;
      _initFormForType(selectedType);
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Ajukan Surat Pengantar'),
      ),
      body: selectedType == null
          ? const Center(child: CircularProgressIndicator())
          : Form(
              key: _formKey,
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(AppSpacing.space16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Pilih Jenis Surat', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                    const SizedBox(height: AppSpacing.space8),
                    DropdownButtonFormField<String>(
                      value: _selectedTypeId,
                      decoration: const InputDecoration(
                        border: OutlineInputBorder(),
                        contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                      ),
                      items: letterState.letterTypes.map((t) {
                        return DropdownMenuItem(
                          value: t.id,
                          child: Text(t.name),
                        );
                      }).toList(),
                      onChanged: (val) {
                        if (val != null) {
                          setState(() {
                            _selectedTypeId = val;
                            final type = letterState.letterTypes.firstWhere((t) => t.id == val);
                            _initFormForType(type);
                          });
                        }
                      },
                    ),
                    const SizedBox(height: AppSpacing.space20),
                    if (selectedType.requirements.isNotEmpty) ...[
                      Text('Persyaratan & Dokumen Pendukung', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                      const SizedBox(height: AppSpacing.space8),
                      ...selectedType.requirements.map((req) {
                        final isUploaded = _uploadedRequirements[req.code] ?? false;
                        return Card(
                          margin: const EdgeInsets.only(bottom: AppSpacing.space8),
                          child: ListTile(
                            leading: Icon(
                              isUploaded ? Icons.check_circle : Icons.upload_file,
                              color: isUploaded ? AppTheme.successColor : Colors.grey,
                            ),
                            title: Text(req.label),
                            subtitle: Text(req.isRequired ? 'Wajib' : 'Opsional'),
                            trailing: OutlinedButton(
                              onPressed: () {
                                setState(() {
                                  _uploadedRequirements[req.code] = !isUploaded;
                                });
                              },
                              child: Text(isUploaded ? 'Terunggah' : 'Unggah'),
                            ),
                          ),
                        );
                      }),
                      const SizedBox(height: AppSpacing.space20),
                    ],
                    Text('Isi Data Pengajuan', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                    const SizedBox(height: AppSpacing.space8),
                    ...selectedType.formSchema.map((field) {
                      final controller = _controllers[field.key] ??= TextEditingController();
                      return Padding(
                        padding: const EdgeInsets.only(bottom: AppSpacing.space16),
                        child: AppTextField(
                          label: '${field.label}${field.isRequired ? ' *' : ''}',
                          hint: 'Masukkan ${field.label.toLowerCase()}',
                          controller: controller,
                          validator: (val) {
                            if (field.isRequired && (val == null || val.trim().isEmpty)) {
                              return '${field.label} wajib diisi';
                            }
                            return null;
                          },
                        ),
                      );
                    }),
                    AppTextField(
                      label: 'Catatan Tambahan (Opsional)',
                      hint: 'Tambahkan keterangan bila ada',
                      controller: _noteController,
                    ),
                    const SizedBox(height: AppSpacing.space24),
                    AppButton(
                      label: 'Kirim Pengajuan Surat',
                      icon: Icons.send_rounded,
                      isLoading: letterState.isLoading,
                      onPressed: () async {
                        if (_formKey.currentState?.validate() ?? false) {
                          final formData = <String, dynamic>{};
                          _controllers.forEach((k, v) {
                            formData[k] = v.text.trim();
                          });
                          letterNotifier.saveLocalDraft(formData);

                          final success = await letterNotifier.submitLetterRequest(
                            letterTypeId: selectedType!.id,
                            residentId: 'res-current-user',
                            formData: formData,
                            residentNote: _noteController.text.trim().isNotEmpty ? _noteController.text.trim() : null,
                          );

                          if (context.mounted && success) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Pengajuan surat berhasil dikirim!')),
                            );
                            context.go('/warga/surat/tracking');
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

class WargaLetterTrackingScreen extends ConsumerWidget {
  const WargaLetterTrackingScreen({super.key});

  StatusType _getStatusType(String status) {
    switch (status) {
      case 'issued':
      case 'approved':
        return StatusType.success;
      case 'submitted':
      case 'in_process':
        return StatusType.info;
      case 'revision_requested':
        return StatusType.warning;
      case 'rejected':
      case 'cancelled':
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
        return 'Diproses Sekretaris';
      case 'revision_requested':
        return 'Perlu Perbaikan';
      case 'approved':
        return 'Disetujui Ketua RT';
      case 'issued':
        return 'Surat Terbit';
      case 'rejected':
        return 'Ditolak';
      case 'cancelled':
        return 'Dibatalkan';
      default:
        return status;
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final letterState = ref.watch(letterProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Status Tracking Surat'),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(letterProvider.notifier).fetchLetterRequests();
        },
        child: letterState.requests.isEmpty
            ? const AppEmptyState(
                title: 'Belum Ada Pengajuan Surat',
                description: 'Surat pengantar yang Anda ajukan akan melacak statusnya di sini.',
              )
            : ListView.builder(
                padding: const EdgeInsets.all(AppSpacing.space16),
                itemCount: letterState.requests.length,
                itemBuilder: (context, index) {
                  final req = letterState.requests[index];
                  return Card(
                    margin: const EdgeInsets.only(bottom: AppSpacing.space12),
                    child: Padding(
                      padding: const EdgeInsets.all(AppSpacing.space16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(
                                child: Text(
                                  req.letterTypeName,
                                  style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
                                ),
                              ),
                              StatusChip(
                                label: _getStatusLabel(req.status),
                                type: _getStatusType(req.status),
                              ),
                            ],
                          ),
                          const SizedBox(height: AppSpacing.space8),
                          Text('No. Request: ${req.requestNumber}', style: const TextStyle(color: Colors.grey, fontSize: 13)),
                          if (req.letterNumber != null)
                            Text('No. Surat: ${req.letterNumber}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: AppTheme.primaryColor)),
                          const SizedBox(height: AppSpacing.space12),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              if (req.status == 'issued') ...[
                                OutlinedButton.icon(
                                  icon: const Icon(Icons.picture_as_pdf, size: 18),
                                  label: const Text('Lihat PDF'),
                                  onPressed: () {
                                    context.push('/warga/surat/viewer?id=${req.id}&number=${req.letterNumber ?? req.requestNumber}');
                                  },
                                ),
                              ],
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

class WargaLetterPdfViewerScreen extends StatelessWidget {
  final String requestId;
  final String letterNumber;

  const WargaLetterPdfViewerScreen({
    super.key,
    required this.requestId,
    required this.letterNumber,
  });

  @override
  Widget build(BuildContext context) {
    final verifCode = 'VERIF-$requestId-QR';
    return Scaffold(
      appBar: AppBar(
        title: Text('Surat Digital #$letterNumber'),
        actions: [
          IconButton(
            icon: const Icon(Icons.download_rounded),
            tooltip: 'Unduh PDF',
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Mengunduh dokumen PDF surat...')),
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.space20),
        child: Container(
          width: double.infinity,
          padding: const EdgeInsets.all(AppSpacing.space20),
          decoration: BoxDecoration(
            color: Theme.of(context).cardColor,
            borderRadius: BorderRadius.circular(AppRadius.card),
            border: Border.all(color: Colors.grey.shade300),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              const Text('PEMERINTAH KOTA JAKARTA TIMUR', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
              const Text('KECAMATAN CAKUNG - KELURAHAN PENGGILINGAN', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
              const Text('RUKUN TETANGGA 005 RUKUN WARGA 002', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
              const Divider(thickness: 2, height: 24),
              const SizedBox(height: 12),
              const Text(
                'SURAT PENGANTAR RT',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16, decoration: TextDecoration.underline),
              ),
              Text('Nomor: $letterNumber', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500)),
              const SizedBox(height: 24),
              const Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Yang bertanda tangan di bawah ini Ketua RT 005 / RW 002 Kelurahan Penggilingan, dengan ini menerangkan bahwa warga berikut adalah benar penduduk domisili RT kami.',
                  textAlign: TextAlign.justify,
                ),
              ),
              const SizedBox(height: 32),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          border: Border.all(color: Colors.black87, width: 2),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: const Icon(Icons.qr_code_2, size: 80),
                      ),
                      const SizedBox(height: 4),
                      Text(verifCode, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold)),
                      const Text('Validasi Keaslian QR', style: TextStyle(fontSize: 10, color: Colors.grey)),
                    ],
                  ),
                  const Column(
                    children: [
                      Text('Jakarta, 08 Agustus 2026'),
                      Text('Ketua RT 005 / RW 002'),
                      SizedBox(height: 48),
                      Text('H. Ahmad Dahlan', style: TextStyle(fontWeight: FontWeight.bold, decoration: TextDecoration.underline)),
                    ],
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class PengurusApprovalScreen extends ConsumerWidget {
  const PengurusApprovalScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final letterState = ref.watch(letterProvider);
    final pendingRequests = letterState.requests.where((r) => r.status == 'submitted' || r.status == 'in_process').toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Persetujuan Surat Warga'),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(letterProvider.notifier).fetchLetterRequests();
        },
        child: pendingRequests.isEmpty
            ? const AppEmptyState(
                title: 'Tidak Ada Permohonan Menunggu',
                description: 'Semua pengajuan surat warga telah diproses.',
              )
            : ListView.builder(
                padding: const EdgeInsets.all(AppSpacing.space16),
                itemCount: pendingRequests.length,
                itemBuilder: (context, index) {
                  final req = pendingRequests[index];
                  return Card(
                    margin: const EdgeInsets.only(bottom: AppSpacing.space16),
                    child: Padding(
                      padding: const EdgeInsets.all(AppSpacing.space16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Expanded(child: Text(req.letterTypeName, style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold))),
                              const StatusChip(label: 'SLA: 24 Jam', type: StatusType.warning),
                            ],
                          ),
                          const SizedBox(height: AppSpacing.space8),
                          Text('Pemohon: ${req.residentName}', style: const TextStyle(fontWeight: FontWeight.w600)),
                          Text('No. Pengajuan: ${req.requestNumber}', style: const TextStyle(color: Colors.grey, fontSize: 13)),
                          const SizedBox(height: AppSpacing.space16),
                          Row(
                            children: [
                              Expanded(
                                child: OutlinedButton(
                                  onPressed: () {
                                    ref.read(letterProvider.notifier).requestRevision(req.id, 'Perlu dokumen tambahan');
                                  },
                                  child: const Text('Revisi'),
                                ),
                              ),
                              const SizedBox(width: AppSpacing.space8),
                              Expanded(
                                child: OutlinedButton(
                                  style: OutlinedButton.styleFrom(foregroundColor: AppTheme.errorColor),
                                  onPressed: () {
                                    ref.read(letterProvider.notifier).rejectRequest(req.id, 'Persyaratan tidak valid');
                                  },
                                  child: const Text('Tolak'),
                                ),
                              ),
                              const SizedBox(width: AppSpacing.space8),
                              Expanded(
                                child: FilledButton(
                                  onPressed: () {
                                    ref.read(letterProvider.notifier).approveRequest(req.id);
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(content: Text('Surat berhasil disetujui!')),
                                    );
                                  },
                                  child: const Text('Setujui'),
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
