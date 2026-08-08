import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/asset_provider.dart';

void main() {
  group('Epic 16 Assets Models Tests', () {
    test('AssetItem fromJson parses all properties correctly', () {
      final json = {
        'id': 'ast-1',
        'category_id': 'cat-1',
        'category_name': 'Tenda & Meja',
        'location_id': 'loc-1',
        'location_name': 'Gudang Balai RT',
        'code': 'AST-001',
        'name': 'Tenda Lipat 3x3m',
        'description': 'Tenda besi serbaguna',
        'condition': 'good',
        'status': 'available',
        'acquisition_date': '2026-01-15',
        'acquisition_value': 1500000.0,
        'pic_id': 'user-1',
        'pic_name': 'Pak RT',
      };

      final asset = AssetItem.fromJson(json);

      expect(asset.id, 'ast-1');
      expect(asset.categoryId, 'cat-1');
      expect(asset.categoryName, 'Tenda & Meja');
      expect(asset.code, 'AST-001');
      expect(asset.name, 'Tenda Lipat 3x3m');
      expect(asset.condition, 'good');
      expect(asset.status, 'available');
      expect(asset.acquisitionValue, 1500000.0);
    });

    test('AssetCategoryItem and AssetLocationItem fromJson parse correctly', () {
      final cat = AssetCategoryItem.fromJson({
        'id': 'cat-1',
        'code': 'CAT-01',
        'name': 'Tenda',
        'status': 'active',
      });
      final loc = AssetLocationItem.fromJson({
        'id': 'loc-1',
        'code': 'LOC-01',
        'name': 'Gudang',
        'status': 'active',
      });

      expect(cat.name, 'Tenda');
      expect(loc.name, 'Gudang');
    });

    test('AssetLoanItem fromJson parses correctly', () {
      final json = {
        'id': 'loan-1',
        'asset_id': 'ast-1',
        'asset_name': 'Tenda Lipat',
        'borrower_id': 'warga-1',
        'borrower_name': 'Ahmad Warga',
        'loan_date': '2026-08-10',
        'due_date': '2026-08-12',
        'condition_out': 'good',
        'status': 'pending',
        'notes': 'Acara syukuran',
      };

      final loan = AssetLoanItem.fromJson(json);

      expect(loan.id, 'loan-1');
      expect(loan.assetName, 'Tenda Lipat');
      expect(loan.borrowerName, 'Ahmad Warga');
      expect(loan.status, 'pending');
    });

    test('AssetMaintenanceItem fromJson parses correctly', () {
      final json = {
        'id': 'maint-1',
        'asset_id': 'ast-1',
        'asset_name': 'Tenda Lipat',
        'maintenance_date': '2026-08-15',
        'maintenance_type': 'Perbaikan Rangka',
        'cost': 75000.0,
        'condition_after': 'good',
        'status_after': 'available',
      };

      final maint = AssetMaintenanceItem.fromJson(json);

      expect(maint.id, 'maint-1');
      expect(maint.cost, 75000.0);
      expect(maint.conditionAfter, 'good');
    });
  });
}
