import 'package:flutter_test/flutter_test.dart';
import 'package:drift/native.dart';
import 'package:mobile/core/storage/app_database.dart';
import 'package:mobile/core/notifications/push_notification_service.dart';

void main() {
  group('Epic M0 Foundation Tests', () {
    test('M0.3: AppDatabase queue sync offline works correctly', () async {
      final db = AppDatabase(NativeDatabase.memory());

      await db.enqueueSyncItem('/api/v1/aduan', 'POST', '{"title": "Lampu Rusak"}');

      final pending = await db.getPendingSyncItems();
      expect(pending.length, equals(1));
      expect(pending.first.endpoint, equals('/api/v1/aduan'));
      expect(pending.first.isSynced, isFalse);

      await db.markSynced(pending.first.id);
      final pendingAfter = await db.getPendingSyncItems();
      expect(pendingAfter.length, equals(0));
    });

    test('M0.2: PushNotificationService streams messages and gets FCM token', () async {
      final pushService = PushNotificationService();
      final token = await pushService.getToken();
      expect(token, contains('mock_fcm_token'));

      expect(
        pushService.onMessageReceived,
        emits(predicate<NotificationPayload>((p) => p.title == 'Pengumuman RT' && p.body == 'Kerja bakti besok pagi')),
      );

      pushService.simulateIncomingNotification(
        title: 'Pengumuman RT',
        body: 'Kerja bakti besok pagi',
      );
    });
  });
}
