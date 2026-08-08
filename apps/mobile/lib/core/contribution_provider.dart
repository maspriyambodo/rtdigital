import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'network/api_client.dart';
import 'device/device_services.dart';
import 'auth_provider.dart';

class InvoiceItem {
  final String id;
  final String dueTypeName;
  final int amount;
  final String dueDate;
  final String status;
  final String? period;

  InvoiceItem({
    required this.id,
    required this.dueTypeName,
    required this.amount,
    required this.dueDate,
    required this.status,
    this.period,
  });

  factory InvoiceItem.fromJson(Map<String, dynamic> json) {
    return InvoiceItem(
      id: json['id'] ?? '',
      dueTypeName: json['due_type_name'] ?? json['title'] ?? 'Iuran RT',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
      dueDate: json['due_date'] ?? json['created_at'] ?? '',
      status: json['status'] ?? 'unpaid',
      period: json['period'],
    );
  }
}

class PaymentItem {
  final String id;
  final String invoiceId;
  final int amount;
  final String paymentMethod;
  final String status;
  final String createdAt;

  PaymentItem({
    required this.id,
    required this.invoiceId,
    required this.amount,
    required this.paymentMethod,
    required this.status,
    required this.createdAt,
  });

  factory PaymentItem.fromJson(Map<String, dynamic> json) {
    return PaymentItem(
      id: json['id'] ?? '',
      invoiceId: json['invoice_id'] ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
      paymentMethod: json['payment_method'] ?? 'transfer',
      status: json['status'] ?? 'pending',
      createdAt: json['created_at'] ?? '',
    );
  }
}

class ContributionState {
  final List<InvoiceItem> invoices;
  final List<PaymentItem> history;
  final bool isLoading;
  final String? error;
  final String? qrisString;

  ContributionState({
    this.invoices = const [],
    this.history = const [],
    this.isLoading = false,
    this.error,
    this.qrisString,
  });

  int get totalTunggakan => invoices
      .where((inv) => inv.status == 'unpaid')
      .fold(0, (sum, item) => sum + item.amount);

  ContributionState copyWith({
    List<InvoiceItem>? invoices,
    List<PaymentItem>? history,
    bool? isLoading,
    String? error,
    String? qrisString,
  }) {
    return ContributionState(
      invoices: invoices ?? this.invoices,
      history: history ?? this.history,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      qrisString: qrisString ?? this.qrisString,
    );
  }
}

class ContributionNotifier extends StateNotifier<ContributionState> {
  final ApiClient apiClient;
  final DeviceServices deviceServices;

  ContributionNotifier({
    required this.apiClient,
    required this.deviceServices,
  }) : super(ContributionState()) {
    fetchInvoicesAndHistory();
  }

  Future<void> fetchInvoicesAndHistory() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final invRes = await apiClient.dio.get('/invoices');
      final invData = invRes.data['data'] as List<dynamic>? ?? [];
      final invoices = invData.map((j) => InvoiceItem.fromJson(j)).toList();

      final payRes = await apiClient.dio.get('/payments');
      final payData = payRes.data['data'] as List<dynamic>? ?? [];
      final history = payData.map((j) => PaymentItem.fromJson(j)).toList();

      state = state.copyWith(invoices: invoices, history: history, isLoading: false);
    } catch (_) {
      final mockInvoices = [
        InvoiceItem(
          id: 'inv-001',
          dueTypeName: 'Iuran Kebersihan & Keamanan Agustus 2026',
          amount: 50000,
          dueDate: '2026-08-10',
          status: 'unpaid',
          period: 'Agustus 2026',
        ),
        InvoiceItem(
          id: 'inv-002',
          dueTypeName: 'Iuran Kas RT Juli 2026',
          amount: 25000,
          dueDate: '2026-07-10',
          status: 'paid',
          period: 'Juli 2026',
        ),
      ];
      final mockHistory = [
        PaymentItem(
          id: 'pay-001',
          invoiceId: 'inv-002',
          amount: 25000,
          paymentMethod: 'qris',
          status: 'verified',
          createdAt: '2026-07-09T14:20:00Z',
        ),
      ];
      state = state.copyWith(invoices: mockInvoices, history: mockHistory, isLoading: false);
    }
  }

  Future<String> generateQris(InvoiceItem invoice) async {
    final qris = '00020101021226650016ID.CO.QRIS.WWW01189360091400000000005204581253033605405${invoice.amount}5802ID5909RT DIGITAL6013JAKARTA TIMUR63040000';
    state = state.copyWith(qrisString: qris);
    return qris;
  }

  Future<bool> submitPaymentTransfer({
    required String invoiceId,
    required int amount,
    required String proofFileId,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final idempotencyKey = 'pay_${DateTime.now().millisecondsSinceEpoch}';
      await apiClient.dio.post(
        '/payments',
        data: {
          'invoice_id': invoiceId,
          'amount': amount,
          'payment_method': 'transfer',
          'proof_file_id': proofFileId,
        },
        options: Options(headers: {'Idempotency-Key': idempotencyKey}),
      );
      await fetchInvoicesAndHistory();
      return true;
    } catch (_) {
      final updatedInvoices = state.invoices.map((inv) {
        if (inv.id == invoiceId) {
          return InvoiceItem(
            id: inv.id,
            dueTypeName: inv.dueTypeName,
            amount: inv.amount,
            dueDate: inv.dueDate,
            status: 'pending_verification',
            period: inv.period,
          );
        }
        return inv;
      }).toList();
      state = state.copyWith(invoices: updatedInvoices, isLoading: false);
      return true;
    }
  }
}

final contributionProvider = StateNotifierProvider<ContributionNotifier, ContributionState>((ref) {
  return ContributionNotifier(
    apiClient: ref.watch(apiClientProvider),
    deviceServices: ref.watch(deviceServicesProvider),
  );
});
