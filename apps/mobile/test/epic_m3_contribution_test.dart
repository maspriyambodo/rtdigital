import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/contribution_provider.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/device/device_services.dart';
import 'package:mobile/core/storage/secure_storage_service.dart';
import 'package:image_picker/image_picker.dart';
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
  Future<XFile?> pickImageFromCamera() async => XFile('/tmp/mock_camera.jpg');
}

void main() {
  test('InvoiceItem and PaymentItem parsing', () {
    final invoice = InvoiceItem.fromJson({
      'id': 'inv-100',
      'due_type_name': 'Iuran Kebersihan',
      'amount': 50000,
      'due_date': '2026-08-10',
      'status': 'unpaid',
    });

    expect(invoice.id, 'inv-100');
    expect(invoice.amount, 50000);
    expect(invoice.status, 'unpaid');

    final payment = PaymentItem.fromJson({
      'id': 'pay-100',
      'invoice_id': 'inv-100',
      'amount': 50000,
      'payment_method': 'qris',
      'status': 'verified',
      'created_at': '2026-08-08T10:00:00Z',
    });

    expect(payment.id, 'pay-100');
    expect(payment.paymentMethod, 'qris');
  });

  test('ContributionNotifier state and flow', () async {
    final notifier = ContributionNotifier(
      apiClient: MockApiClient(),
      deviceServices: MockDeviceServices(),
    );

    await Future.delayed(const Duration(milliseconds: 100));
    expect(notifier.state.invoices.isNotEmpty, true);
    expect(notifier.state.totalTunggakan, 50000);

    final qris = await notifier.generateQris(notifier.state.invoices.first);
    expect(qris.contains('ID.CO.QRIS'), true);

    final success = await notifier.submitPaymentTransfer(
      invoiceId: 'inv-001',
      amount: 50000,
      proofFileId: 'file-123',
    );
    expect(success, true);
  });
}

