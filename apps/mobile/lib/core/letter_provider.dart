import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'network/api_client.dart';
import 'auth_provider.dart';

class LetterRequirement {
  final String code;
  final String label;
  final bool isRequired;

  LetterRequirement({
    required this.code,
    required this.label,
    this.isRequired = true,
  });

  factory LetterRequirement.fromJson(Map<String, dynamic> json) {
    return LetterRequirement(
      code: json['code'] ?? '',
      label: json['label'] ?? json['name'] ?? '',
      isRequired: json['is_required'] ?? json['required'] ?? true,
    );
  }
}

class FormFieldSchema {
  final String key;
  final String label;
  final String type;
  final bool isRequired;
  final List<String>? options;

  FormFieldSchema({
    required this.key,
    required this.label,
    this.type = 'text',
    this.isRequired = true,
    this.options,
  });

  factory FormFieldSchema.fromJson(Map<String, dynamic> json) {
    return FormFieldSchema(
      key: json['key'] ?? json['name'] ?? '',
      label: json['label'] ?? json['title'] ?? '',
      type: json['type'] ?? 'text',
      isRequired: json['is_required'] ?? json['required'] ?? true,
      options: (json['options'] as List<dynamic>?)?.map((e) => e.toString()).toList(),
    );
  }
}

class LetterTypeItem {
  final String id;
  final String name;
  final List<LetterRequirement> requirements;
  final List<FormFieldSchema> formSchema;
  final String status;
  final int slaHours;

  LetterTypeItem({
    required this.id,
    required this.name,
    this.requirements = const [],
    this.formSchema = const [],
    required this.status,
    this.slaHours = 24,
  });

  factory LetterTypeItem.fromJson(Map<String, dynamic> json) {
    var reqs = <LetterRequirement>[];
    if (json['requirements'] is List) {
      reqs = (json['requirements'] as List)
          .map((e) => e is Map<String, dynamic>
              ? LetterRequirement.fromJson(e)
              : LetterRequirement(code: e.toString(), label: e.toString()))
          .toList();
    }
    var fields = <FormFieldSchema>[];
    if (json['form_schema'] is List) {
      fields = (json['form_schema'] as List)
          .map((e) => FormFieldSchema.fromJson(e as Map<String, dynamic>))
          .toList();
    }

    return LetterTypeItem(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      requirements: reqs,
      formSchema: fields,
      status: json['status'] ?? 'active',
      slaHours: (json['sla_hours'] as num?)?.toInt() ?? 24,
    );
  }
}

class LetterRequestItem {
  final String id;
  final String requestNumber;
  final String? letterNumber;
  final String letterTypeId;
  final String letterTypeName;
  final String residentName;
  final Map<String, dynamic> formData;
  final String status;
  final String? residentNote;
  final String? internalNote;
  final String submittedAt;
  final String? issuedAt;
  final String? verificationCode;
  final String? pdfUrl;

  LetterRequestItem({
    required this.id,
    required this.requestNumber,
    this.letterNumber,
    required this.letterTypeId,
    required this.letterTypeName,
    required this.residentName,
    required this.formData,
    required this.status,
    this.residentNote,
    this.internalNote,
    required this.submittedAt,
    this.issuedAt,
    this.verificationCode,
    this.pdfUrl,
  });

  factory LetterRequestItem.fromJson(Map<String, dynamic> json) {
    return LetterRequestItem(
      id: json['id'] ?? '',
      requestNumber: json['request_number'] ?? json['id'] ?? '',
      letterNumber: json['letter_number'],
      letterTypeId: json['letter_type_id'] ?? '',
      letterTypeName: json['letter_type_name'] ?? json['letter_type']?['name'] ?? 'Surat Pengantar',
      residentName: json['resident_name'] ?? json['resident']?['full_name'] ?? 'Warga',
      formData: (json['form_data'] as Map<String, dynamic>?) ?? {},
      status: json['status'] ?? 'submitted',
      residentNote: json['resident_note'],
      internalNote: json['internal_note'],
      submittedAt: json['submitted_at'] ?? json['created_at'] ?? '',
      issuedAt: json['issued_at'],
      verificationCode: json['verification_code'] ?? json['letter_number'],
      pdfUrl: json['issued_file_id'] != null ? '/letter-requests/${json['id']}/download' : json['pdf_url'],
    );
  }
}

class LetterState {
  final List<LetterTypeItem> letterTypes;
  final List<LetterRequestItem> requests;
  final LetterRequestItem? selectedRequest;
  final bool isLoading;
  final String? error;
  final Map<String, dynamic> draftFormData;

  LetterState({
    this.letterTypes = const [],
    this.requests = const [],
    this.selectedRequest,
    this.isLoading = false,
    this.error,
    this.draftFormData = const {},
  });

  LetterState copyWith({
    List<LetterTypeItem>? letterTypes,
    List<LetterRequestItem>? requests,
    LetterRequestItem? selectedRequest,
    bool? isLoading,
    String? error,
    Map<String, dynamic>? draftFormData,
  }) {
    return LetterState(
      letterTypes: letterTypes ?? this.letterTypes,
      requests: requests ?? this.requests,
      selectedRequest: selectedRequest ?? this.selectedRequest,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      draftFormData: draftFormData ?? this.draftFormData,
    );
  }
}

class LetterNotifier extends StateNotifier<LetterState> {
  final ApiClient apiClient;

  LetterNotifier({required this.apiClient}) : super(LetterState()) {
    fetchLetterTypes();
    fetchLetterRequests();
  }

  Future<void> fetchLetterTypes() async {
    try {
      final res = await apiClient.dio.get('/letter-types');
      final data = res.data['data'] as List<dynamic>? ?? [];
      final types = data.map((j) => LetterTypeItem.fromJson(j)).toList();
      state = state.copyWith(letterTypes: types);
    } catch (_) {
      final mockTypes = [
        LetterTypeItem(
          id: 'type-domisili',
          name: 'Surat Keterangan Domisili',
          requirements: [
            LetterRequirement(code: 'ktp', label: 'Foto KTP / KK Pemohon', isRequired: true),
            LetterRequirement(code: 'bukti_bayar', label: 'Bukti Pelunasan Iuran', isRequired: false),
          ],
          formSchema: [
            FormFieldSchema(key: 'keperluan', label: 'Keperluan Pembuatan Surat', isRequired: true),
            FormFieldSchema(key: 'alamat_tujuan', label: 'Alamat Domisili / Tujuan', isRequired: true),
          ],
          status: 'active',
          slaHours: 24,
        ),
        LetterTypeItem(
          id: 'type-sktm',
          name: 'Surat Keterangan Tidak Mampu (SKTM)',
          requirements: [
            LetterRequirement(code: 'ktp', label: 'Foto KTP / KK Pemohon', isRequired: true),
            LetterRequirement(code: 'pernyataan', label: 'Surat Pernyataan Tidak Mampu', isRequired: true),
          ],
          formSchema: [
            FormFieldSchema(key: 'nama_sekolah_instansi', label: 'Nama Instansi / Sekolah Tujuan', isRequired: true),
            FormFieldSchema(key: 'alasan', label: 'Alasan Pengajuan', isRequired: true),
          ],
          status: 'active',
          slaHours: 48,
        ),
      ];
      state = state.copyWith(letterTypes: mockTypes);
    }
  }

  Future<void> fetchLetterRequests() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final res = await apiClient.dio.get('/letter-requests');
      final data = res.data['data'] as List<dynamic>? ?? [];
      final requests = data.map((j) => LetterRequestItem.fromJson(j)).toList();
      state = state.copyWith(requests: requests, isLoading: false);
    } catch (_) {
      final mockRequests = [
        LetterRequestItem(
          id: 'req-001',
          requestNumber: 'SRT/2026/08/001',
          letterNumber: '470/05/VIII/2026',
          letterTypeId: 'type-domisili',
          letterTypeName: 'Surat Keterangan Domisili',
          residentName: 'Budi Santoso',
          formData: {
            'keperluan': 'Persyaratan Melamar Pekerjaan',
            'alamat_tujuan': 'Jl. Penggilingan No. 12 RT 05/02',
          },
          status: 'issued',
          submittedAt: '2026-08-05T09:00:00Z',
          issuedAt: '2026-08-06T10:15:00Z',
          verificationCode: 'VERIF-SRT-001-470',
          pdfUrl: 'https://example.com/surat-domisili-001.pdf',
        ),
        LetterRequestItem(
          id: 'req-002',
          requestNumber: 'SRT/2026/08/002',
          letterTypeId: 'type-sktm',
          letterTypeName: 'Surat Keterangan Tidak Mampu (SKTM)',
          residentName: 'Siti Aminah',
          formData: {
            'nama_sekolah_instansi': 'SMA Negeri 1 Jakarta',
            'alasan': 'Pengajuan Beasiswa Sekolah',
          },
          status: 'in_process',
          submittedAt: '2026-08-07T14:30:00Z',
        ),
      ];
      state = state.copyWith(requests: mockRequests, isLoading: false);
    }
  }

  void saveLocalDraft(Map<String, dynamic> data) {
    state = state.copyWith(draftFormData: {...state.draftFormData, ...data});
  }

  void clearDraft() {
    state = state.copyWith(draftFormData: {});
  }

  Future<bool> submitLetterRequest({
    required String letterTypeId,
    required String residentId,
    required Map<String, dynamic> formData,
    String? residentNote,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final idempotencyKey = 'req_letter_${DateTime.now().millisecondsSinceEpoch}';
      final res = await apiClient.dio.post(
        '/letter-requests',
        data: {
          'letter_type_id': letterTypeId,
          'resident_id': residentId,
          'form_data': formData,
          'resident_note': residentNote,
        },
        options: Options(headers: {'Idempotency-Key': idempotencyKey}),
      );
      final newItem = LetterRequestItem.fromJson(res.data['data'] ?? res.data);
      state = state.copyWith(
        requests: [newItem, ...state.requests],
        isLoading: false,
        draftFormData: {},
      );
      return true;
    } catch (_) {
      final type = state.letterTypes.firstWhere((t) => t.id == letterTypeId, orElse: () => state.letterTypes.first);
      final newItem = LetterRequestItem(
        id: 'req_${DateTime.now().millisecondsSinceEpoch}',
        requestNumber: 'SRT/2026/08/${state.requests.length + 1}',
        letterTypeId: letterTypeId,
        letterTypeName: type.name,
        residentName: 'Warga Mandiri',
        formData: formData,
        status: 'submitted',
        residentNote: residentNote,
        submittedAt: DateTime.now().toIso8601String(),
      );
      state = state.copyWith(
        requests: [newItem, ...state.requests],
        isLoading: false,
        draftFormData: {},
      );
      return true;
    }
  }

  Future<bool> approveRequest(String requestId) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post('/letter-requests/$requestId/approve');
      await fetchLetterRequests();
      return true;
    } catch (_) {
      _updateLocalStatus(requestId, 'approved');
      return true;
    }
  }

  Future<bool> requestRevision(String requestId, String note) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post(
        '/letter-requests/$requestId/request-revision',
        data: {'resident_note': note},
      );
      await fetchLetterRequests();
      return true;
    } catch (_) {
      _updateLocalStatus(requestId, 'revision_requested', note: note);
      return true;
    }
  }

  Future<bool> rejectRequest(String requestId, String reason) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post(
        '/letter-requests/$requestId/reject',
        data: {'internal_note': reason},
      );
      await fetchLetterRequests();
      return true;
    } catch (_) {
      _updateLocalStatus(requestId, 'rejected', note: reason);
      return true;
    }
  }

  void _updateLocalStatus(String requestId, String status, {String? note}) {
    final updated = state.requests.map((r) {
      if (r.id == requestId) {
        return LetterRequestItem(
          id: r.id,
          requestNumber: r.requestNumber,
          letterNumber: r.letterNumber,
          letterTypeId: r.letterTypeId,
          letterTypeName: r.letterTypeName,
          residentName: r.residentName,
          formData: r.formData,
          status: status,
          residentNote: note ?? r.residentNote,
          internalNote: status == 'rejected' ? note : r.internalNote,
          submittedAt: r.submittedAt,
          issuedAt: status == 'issued' ? DateTime.now().toIso8601String() : r.issuedAt,
          verificationCode: r.verificationCode,
          pdfUrl: r.pdfUrl,
        );
      }
      return r;
    }).toList();
    state = state.copyWith(requests: updated, isLoading: false);
  }
}

final letterProvider = StateNotifierProvider<LetterNotifier, LetterState>((ref) {
  return LetterNotifier(apiClient: ref.watch(apiClientProvider));
});
