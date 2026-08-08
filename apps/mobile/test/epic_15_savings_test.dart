import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/savings_provider.dart';

void main() {
  group('Epic 15 Savings Tests', () {
    test('SavingsProduct fromJson works correctly', () {
      final json = {
        'id': 'sp-1',
        'code': 'SP-01',
        'name': 'Tabungan Sukarela',
        'description': 'Deskripsi tabungan',
        'minimum_deposit': 50000,
        'status': 'active',
      };

      final product = SavingsProduct.fromJson(json);

      expect(product.id, 'sp-1');
      expect(product.code, 'SP-01');
      expect(product.name, 'Tabungan Sukarela');
      expect(product.minimumDeposit, 50000.0);
      expect(product.status, 'active');
    });

    test('SavingsAccount fromJson works correctly', () {
      final json = {
        'id': 'sa-1',
        'savings_product_id': 'sp-1',
        'household_id': 'hh-1',
        'status': 'active',
        'balance': 150000,
        'product_name': 'Tabungan Sukarela',
      };

      final account = SavingsAccount.fromJson(json);

      expect(account.id, 'sa-1');
      expect(account.balance, 150000.0);
      expect(account.productName, 'Tabungan Sukarela');
    });

    test('SavingsTransaction fromJson works correctly', () {
      final json = {
        'id': 'st-1',
        'account_id': 'sa-1',
        'transaction_number': 'TX-001',
        'type': 'deposit',
        'amount': 100000,
        'balance_after': 150000,
        'transaction_date': '2026-08-08',
        'description': 'Setoran awal',
        'verification_status': 'verified',
        'created_by': 'user-1',
      };

      final tx = SavingsTransaction.fromJson(json);

      expect(tx.id, 'st-1');
      expect(tx.amount, 100000.0);
      expect(tx.type, 'deposit');
      expect(tx.verificationStatus, 'verified');
    });
  });
}
