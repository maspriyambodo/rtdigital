import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:mobile/core/complaint_provider.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/device/device_services.dart';
import 'package:mobile/core/storage/secure_storage_service.dart';

class FakeApiClient extends ApiClient {
  final Dio _customDio;

  FakeApiClient(this._customDio) : super(baseUrl: 'https://api.example.com', storageService: SecureStorageService());

  @override
  Dio get dio => _customDio;
}

class _FakeAdapter implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path.contains('/complaint-categories')) {
      final jsonStr = '''
      {
        "data": [
          {"id": "cat-kebersihan", "code": "kebersihan", "name": "Kebersihan"},
          {"id": "cat-keamanan", "code": "keamanan", "name": "Keamanan"}
        ]
      }
      ''';
      return ResponseBody.fromString(jsonStr, 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    if (options.path.endsWith('/complaints') && options.method == 'GET') {
      final jsonStr = '''
      {
        "data": [
          {
            "id": "cmp-101",
            "ticket_number": "TKT/2026/08/101",
            "complaint_category_id": "cat-kebersihan",
            "category_name": "Kebersihan",
            "title": "Sampah Berhamburan",
            "description": "Tong sampah TPS meluap",
            "location_description": "Depan Blok A",
            "status": "submitted",
            "priority": "medium",
            "reporter_name": "Pelapor Test",
            "created_at": "2026-08-08T10:00:00Z"
          }
        ]
      }
      ''';
      return ResponseBody.fromString(jsonStr, 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    if (options.path.endsWith('/complaints') && options.method == 'POST') {
      final jsonStr = '''
      {
        "data": {
          "id": "cmp-102",
          "ticket_number": "TKT/2026/08/102",
          "complaint_category_id": "cat-keamanan",
          "category_name": "Keamanan",
          "title": "Penerangan Jalan Mati",
          "description": "Lampu pos 2 mati",
          "location_description": "Pos 2 RT 05",
          "status": "submitted",
          "priority": "high",
          "reporter_name": "Pelapor Test",
          "created_at": "2026-08-08T11:00:00Z"
        }
      }
      ''';
      return ResponseBody.fromString(jsonStr, 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    if (options.path.contains('/assign') || options.path.contains('/status') || options.path.contains('/comments')) {
      return ResponseBody.fromString('{"success": true}', 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    return ResponseBody.fromString('{}', 200);
  }

  @override
  void close({bool force = false}) {}
}

void main() {
  late ComplaintNotifier notifier;
  late FakeApiClient fakeApiClient;

  setUp(() {
    final testDio = Dio();
    testDio.httpClientAdapter = _FakeAdapter();
    fakeApiClient = FakeApiClient(testDio);
    notifier = ComplaintNotifier(
      apiClient: fakeApiClient,
      deviceServices: DeviceServices(),
    );
  });

  test('ComplaintNotifier fetches categories & initial complaints', () async {
    await Future.delayed(const Duration(milliseconds: 100));
    expect(notifier.state.categories.isNotEmpty, isTrue);
    expect(notifier.state.complaints.isNotEmpty, isTrue);
  });

  test('ComplaintNotifier creates complaint with Idempotency header', () async {
    final result = await notifier.createComplaint(
      categoryId: 'cat-keamanan',
      title: 'Penerangan Jalan Mati',
      description: 'Lampu pos 2 mati',
      locationDescription: 'Pos 2 RT 05',
    );
    expect(result, isTrue);
    expect(notifier.state.complaints.length, greaterThan(1));
  });

  test('ComplaintNotifier updates status and adds comment', () async {
    final statusResult = await notifier.updateStatus('cmp-101', 'in_process', comment: 'Sudah ditangani');
    expect(statusResult, isTrue);

    final commentResult = await notifier.addComment('cmp-101', 'Catatan tambahan dari warga');
    expect(commentResult, isTrue);
  });
}
