import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'network/api_client.dart';
import 'auth_provider.dart';

class SavingsProduct {
  final String id;
  final String code;
  final String name;
  final String? description;
  final double minimumDeposit;
  final String status;

  SavingsProduct({
    required this.id,
    required this.code,
    required this.name,
    this.description,
    required this.minimumDeposit,
    required this.status,
  });

  factory SavingsProduct.fromJson(Map<String, dynamic> json) {
    return SavingsProduct(
      id: json['id'] ?? '',
      code: json['code'] ?? '',
      name: json['name'] ?? '',
      description: json['description'],
      minimumDeposit: (json['minimum_deposit'] as num?)?.toDouble() ?? 0.0,
      status: json['status'] ?? 'active',
    );
  }
}

class SavingsAccount {
  final String id;
  final String savingsProductId;
  final String householdId;
  final String status;
  final double balance;
  final String? productName;

  SavingsAccount({
    required this.id,
    required this.savingsProductId,
    required this.householdId,
    required this.status,
    required this.balance,
    this.productName,
  });

  factory SavingsAccount.fromJson(Map<String, dynamic> json) {
    return SavingsAccount(
      id: json['id'] ?? '',
      savingsProductId: json['savings_product_id'] ?? '',
      householdId: json['household_id'] ?? '',
      status: json['status'] ?? 'active',
      balance: (json['balance'] as num?)?.toDouble() ?? 0.0,
      productName: json['product_name'],
    );
  }
}

class SavingsTransaction {
  final String id;
  final String accountId;
  final String transactionNumber;
  final String type;
  final double amount;
  final double balanceAfter;
  final String transactionDate;
  final String description;
  final String verificationStatus;
  final String? rejectionReason;
  final String createdBy;

  SavingsTransaction({
    required this.id,
    required this.accountId,
    required this.transactionNumber,
    required this.type,
    required this.amount,
    required this.balanceAfter,
    required this.transactionDate,
    required this.description,
    required this.verificationStatus,
    this.rejectionReason,
    required this.createdBy,
  });

  factory SavingsTransaction.fromJson(Map<String, dynamic> json) {
    return SavingsTransaction(
      id: json['id'] ?? '',
      accountId: json['account_id'] ?? '',
      transactionNumber: json['transaction_number'] ?? '',
      type: json['type'] ?? 'deposit',
      amount: (json['amount'] as num?)?.toDouble() ?? 0.0,
      balanceAfter: (json['balance_after'] as num?)?.toDouble() ?? 0.0,
      transactionDate: json['transaction_date'] ?? '',
      description: json['description'] ?? '',
      verificationStatus: json['verification_status'] ?? 'pending',
      rejectionReason: json['rejection_reason'],
      createdBy: json['created_by'] ?? '',
    );
  }
}

class SavingsState {
  final List<SavingsProduct> products;
  final List<SavingsAccount> accounts;
  final List<SavingsTransaction> transactions;
  final bool isLoading;
  final String? error;

  SavingsState({
    this.products = const [],
    this.accounts = const [],
    this.transactions = const [],
    this.isLoading = false,
    this.error,
  });

  SavingsState copyWith({
    List<SavingsProduct>? products,
    List<SavingsAccount>? accounts,
    List<SavingsTransaction>? transactions,
    bool? isLoading,
    String? error,
  }) {
    return SavingsState(
      products: products ?? this.products,
      accounts: accounts ?? this.accounts,
      transactions: transactions ?? this.transactions,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class SavingsNotifier extends StateNotifier<SavingsState> {
  final ApiClient _apiClient;

  SavingsNotifier(this._apiClient) : super(SavingsState()) {
    fetchSavingsData();
  }

  Future<void> fetchSavingsData() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final productsRes = await _apiClient.dio.get('/savings/products');
      final accountsRes = await _apiClient.dio.get('/savings/accounts');
      final txsRes = await _apiClient.dio.get('/savings/transactions');

      final products = (productsRes.data['data'] as List? ?? [])
          .map((e) => SavingsProduct.fromJson(e))
          .toList();
      final accounts = (accountsRes.data['data'] as List? ?? [])
          .map((e) => SavingsAccount.fromJson(e))
          .toList();
      final transactions = (txsRes.data['data'] as List? ?? [])
          .map((e) => SavingsTransaction.fromJson(e))
          .toList();

      state = state.copyWith(
        products: products,
        accounts: accounts,
        transactions: transactions,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }
}

final savingsProvider = StateNotifierProvider<SavingsNotifier, SavingsState>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return SavingsNotifier(apiClient);
});
