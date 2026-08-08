import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:mobile/core/security_provider.dart';
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
    if (options.path.contains('/emergency-alerts') && options.method == 'GET') {
      final jsonStr = '''
      {
        "data": [
          {
            "id": "alt-test",
            "category": "kebakaran",
            "reporter_name": "Testing Warga",
            "location_description": "RT 05 No 1",
            "status": "active",
            "created_at": "2026-08-08T12:00:00Z"
          }
        ]
      }
      ''';
      return ResponseBody.fromString(jsonStr, 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    if (options.path.contains('/emergency-alerts') && options.method == 'POST') {
      final jsonStr = '''
      {
        "data": {
          "id": "alt-new",
          "category": "medis",
          "reporter_name": "Testing Warga",
          "location_description": "RT 05 Pos 1",
          "status": "active",
          "created_at": "2026-08-08T12:05:00Z"
        }
      }
      ''';
      return ResponseBody.fromString(jsonStr, 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    if (options.path.contains('/patrol-attendances')) {
      final jsonStr = '''
      {
        "data": [
          {
            "id": "pat-1",
            "post_name": "Pos Ronda Utama",
            "officer_name": "Petugas 1",
            "checkin_time": "2026-08-08T20:00:00Z",
            "status": "valid"
          }
        ]
      }
      ''';
      return ResponseBody.fromString(jsonStr, 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }

    return ResponseBody.fromString('{"success": true}', 200);
  }

  @override
  void close({bool force = false}) {}
}

void main() {
  late SecurityNotifier notifier;
  late FakeApiClient fakeApiClient;

  setUp(() {
    final testDio = Dio();
    testDio.httpClientAdapter = _FakeAdapter();
    fakeApiClient = FakeApiClient(testDio);
    notifier = SecurityNotifier(
      apiClient: fakeApiClient,
      deviceServices: DeviceServices(),
    );
  });

  test('SecurityNotifier fetches active alerts & patrol attendances', () async {
    await Future.delayed(const Duration(milliseconds: 100));
    expect(notifier.state.activeAlerts.isNotEmpty, isTrue);
    expect(notifier.state.patrolAttendances.isNotEmpty, isTrue);
  });

  test('SecurityNotifier sends panic alert', () async {
    final result = await notifier.sendPanicAlert(
      category: 'medis',
      locationDescription: 'Pos 1',
    );
    expect(result, isTrue);
    expect(notifier.state.activeAlerts.length, greaterThan(1));
  });

  test('SecurityNotifier acknowledges & resolves alert', () async {
    final ackResult = await notifier.acknowledgeAlert('alt-test', 'Pak Satpam');
    expect(ackResult, isTrue);

    final resolveResult = await notifier.resolveAlert('alt-test');
    expect(resolveResult, isTrue);
  });

  test('SecurityNotifier checks in patrol QR', () async {
    final result = await notifier.checkinPatrolQR('pos-ronda-barat');
    expect(result, isTrue);
    expect(notifier.state.patrolAttendances.isNotEmpty, isTrue);
  });
}
