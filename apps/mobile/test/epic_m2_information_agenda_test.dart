import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/core/information_provider.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/device/device_services.dart';
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


class MockDeviceServices extends DeviceServices {
  @override
  Future<bool> addToCalendar({
    required String title,
    required String description,
    required String location,
    required DateTime startDate,
    required DateTime endDate,
  }) async => true;
}

void main() {
  test('Announcement and Event model parsing', () {
    final announcement = Announcement.fromJson({
      'id': 'ann-100',
      'title': 'Pengumuman Kerja Bakti',
      'category': 'Kegiatan',
      'content': 'Isi pengumuman kerja bakti',
      'attachment_url': 'http://example.com/doc.pdf',
      'attachment_type': 'pdf',
      'published_at': '2026-08-08T10:00:00Z',
    });

    expect(announcement.id, 'ann-100');
    expect(announcement.category, 'Kegiatan');
    expect(announcement.attachmentType, 'pdf');

    final event = EventItem.fromJson({
      'id': 'evt-100',
      'title': 'Rapat Pengurus',
      'description': 'Bahasan kas RT',
      'location': 'Balai RT',
      'starts_at': '2026-08-10T19:00:00Z',
      'ends_at': '2026-08-10T21:00:00Z',
    });

    expect(event.id, 'evt-100');
    expect(event.location, 'Balai RT');
  });

  test('InformationNotifier state and filtering', () async {
    final notifier = InformationNotifier(
      apiClient: MockApiClient(),
      deviceServices: MockDeviceServices(),
    );

    await Future.delayed(const Duration(milliseconds: 100));
    expect(notifier.state.announcements.isNotEmpty, true);
    expect(notifier.state.events.isNotEmpty, true);

    notifier.filterCategory('Kegiatan');
    expect(notifier.state.categoryFilter, 'Kegiatan');

    final event = notifier.state.events.first;
    final saved = await notifier.saveEventToCalendar(event);
    expect(saved, true);
  });
}
