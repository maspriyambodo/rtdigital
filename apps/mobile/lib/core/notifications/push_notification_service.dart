import 'dart:async';

class NotificationPayload {
  final String id;
  final String title;
  final String body;
  final Map<String, String>? data;

  NotificationPayload({
    required this.id,
    required this.title,
    required this.body,
    this.data,
  });
}

class PushNotificationService {
  final _messageController = StreamController<NotificationPayload>.broadcast();

  Stream<NotificationPayload> get onMessageReceived => _messageController.stream;

  Future<void> initialize() async {
    // Inisialisasi FCM & Push Notification handler
  }

  Future<String?> getToken() async {
    return 'mock_fcm_token_rt_digital_12345';
  }

  void simulateIncomingNotification({
    required String title,
    required String body,
    Map<String, String>? data,
  }) {
    _messageController.add(
      NotificationPayload(
        id: DateTime.now().millisecondsSinceEpoch.toString(),
        title: title,
        body: body,
        data: data,
      ),
    );
  }

  void dispose() {
    _messageController.close();
  }
}
