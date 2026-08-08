import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/letter_provider.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/storage/secure_storage_service.dart';
import 'package:dio/dio.dart';

class MockApiClient implements ApiClient {
  @override
  final Dio dio = Dio();
  @override
  final String baseUrl = 'http://localhost';
  @override
  final SecureStorageService storageService = SecureStorageService();
  @override
  void setAuthToken(String token) {}
  @override
  void clearAuthToken() {}
}

void main() {
  test('LetterTypeItem and LetterRequestItem parsing', () {
    final type = LetterTypeItem.fromJson({
      'id': 'type-100',
      'name': 'Surat Keterangan Domisili',
      'status': 'active',
      'sla_hours': 24,
      'requirements': [
        {'code': 'ktp', 'label': 'KTP', 'is_required': true}
      ],
      'form_schema': [
        {'key': 'keperluan', 'label': 'Keperluan', 'type': 'text', 'is_required': true}
      ],
    });

    expect(type.id, 'type-100');
    expect(type.name, 'Surat Keterangan Domisili');
    expect(type.requirements.length, 1);
    expect(type.formSchema.length, 1);

    final req = LetterRequestItem.fromJson({
      'id': 'req-100',
      'request_number': 'SRT/001',
      'letter_number': '470/01/2026',
      'letter_type_id': 'type-100',
      'letter_type_name': 'Surat Keterangan Domisili',
      'resident_name': 'Budi',
      'form_data': {'keperluan': 'Kerja'},
      'status': 'issued',
      'submitted_at': '2026-08-08T10:00:00Z',
    });

    expect(req.id, 'req-100');
    expect(req.letterNumber, '470/01/2026');
    expect(req.status, 'issued');
  });

  test('LetterNotifier state and actions', () async {
    final notifier = LetterNotifier(apiClient: MockApiClient());

    await Future.delayed(const Duration(milliseconds: 100));
    expect(notifier.state.letterTypes.isNotEmpty, true);
    expect(notifier.state.requests.isNotEmpty, true);

    // Save draft
    notifier.saveLocalDraft({'keperluan': 'Test Draft'});
    expect(notifier.state.draftFormData['keperluan'], 'Test Draft');

    // Submit request (Task M4.1)
    final submitSuccess = await notifier.submitLetterRequest(
      letterTypeId: notifier.state.letterTypes.first.id,
      residentId: 'res-1',
      formData: {'keperluan': 'Surat Nikah'},
    );
    expect(submitSuccess, true);

    final newReqId = notifier.state.requests.first.id;

    // Approve (Task M4.4)
    final approveSuccess = await notifier.approveRequest(newReqId);
    expect(approveSuccess, true);
    expect(notifier.state.requests.firstWhere((r) => r.id == newReqId).status, 'approved');

    // Request revision (Task M4.4)
    final revSuccess = await notifier.requestRevision(newReqId, 'Lengkapi KTP');
    expect(revSuccess, true);
    expect(notifier.state.requests.firstWhere((r) => r.id == newReqId).status, 'revision_requested');
  });
}
