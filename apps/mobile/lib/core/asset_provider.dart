import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'network/api_client.dart';
import 'auth_provider.dart';

class AssetItem {
  final String id;
  final String categoryId;
  final String? categoryName;
  final String locationId;
  final String? locationName;
  final String code;
  final String name;
  final String? description;
  final String condition;
  final String status;
  final String? acquisitionDate;
  final double? acquisitionValue;
  final String? picId;
  final String? picName;
  final String? fileObjectId;

  AssetItem({
    required this.id,
    required this.categoryId,
    this.categoryName,
    required this.locationId,
    this.locationName,
    required this.code,
    required this.name,
    this.description,
    required this.condition,
    required this.status,
    this.acquisitionDate,
    this.acquisitionValue,
    this.picId,
    this.picName,
    this.fileObjectId,
  });

  factory AssetItem.fromJson(Map<String, dynamic> json) {
    return AssetItem(
      id: json['id'] ?? '',
      categoryId: json['category_id'] ?? '',
      categoryName: json['category_name'],
      locationId: json['location_id'] ?? '',
      locationName: json['location_name'],
      code: json['code'] ?? '',
      name: json['name'] ?? '',
      description: json['description'],
      condition: json['condition'] ?? 'good',
      status: json['status'] ?? 'available',
      acquisitionDate: json['acquisition_date'],
      acquisitionValue: (json['acquisition_value'] as num?)?.toDouble(),
      picId: json['pic_id'],
      picName: json['pic_name'],
      fileObjectId: json['file_object_id'],
    );
  }
}

class AssetCategoryItem {
  final String id;
  final String code;
  final String name;
  final String status;

  AssetCategoryItem({
    required this.id,
    required this.code,
    required this.name,
    required this.status,
  });

  factory AssetCategoryItem.fromJson(Map<String, dynamic> json) {
    return AssetCategoryItem(
      id: json['id'] ?? '',
      code: json['code'] ?? '',
      name: json['name'] ?? '',
      status: json['status'] ?? 'active',
    );
  }
}

class AssetLocationItem {
  final String id;
  final String code;
  final String name;
  final String status;

  AssetLocationItem({
    required this.id,
    required this.code,
    required this.name,
    required this.status,
  });

  factory AssetLocationItem.fromJson(Map<String, dynamic> json) {
    return AssetLocationItem(
      id: json['id'] ?? '',
      code: json['code'] ?? '',
      name: json['name'] ?? '',
      status: json['status'] ?? 'active',
    );
  }
}

class AssetLoanItem {
  final String id;
  final String assetId;
  final String? assetName;
  final String borrowerId;
  final String? borrowerName;
  final String? approverId;
  final String? approverName;
  final String loanDate;
  final String dueDate;
  final String? returnDate;
  final String conditionOut;
  final String? conditionIn;
  final String status;
  final String? notes;

  AssetLoanItem({
    required this.id,
    required this.assetId,
    this.assetName,
    required this.borrowerId,
    this.borrowerName,
    this.approverId,
    this.approverName,
    required this.loanDate,
    required this.dueDate,
    this.returnDate,
    required this.conditionOut,
    this.conditionIn,
    required this.status,
    this.notes,
  });

  factory AssetLoanItem.fromJson(Map<String, dynamic> json) {
    return AssetLoanItem(
      id: json['id'] ?? '',
      assetId: json['asset_id'] ?? '',
      assetName: json['asset_name'],
      borrowerId: json['borrower_id'] ?? '',
      borrowerName: json['borrower_name'],
      approverId: json['approver_id'],
      approverName: json['approver_name'],
      loanDate: json['loan_date'] ?? '',
      dueDate: json['due_date'] ?? '',
      returnDate: json['return_date'],
      conditionOut: json['condition_out'] ?? 'good',
      conditionIn: json['condition_in'],
      status: json['status'] ?? 'pending',
      notes: json['notes'],
    );
  }
}

class AssetMaintenanceItem {
  final String id;
  final String assetId;
  final String? assetName;
  final String maintenanceDate;
  final String maintenanceType;
  final double? cost;
  final String? technician;
  final String? notes;
  final String conditionAfter;
  final String statusAfter;

  AssetMaintenanceItem({
    required this.id,
    required this.assetId,
    this.assetName,
    required this.maintenanceDate,
    required this.maintenanceType,
    this.cost,
    this.technician,
    this.notes,
    required this.conditionAfter,
    required this.statusAfter,
  });

  factory AssetMaintenanceItem.fromJson(Map<String, dynamic> json) {
    return AssetMaintenanceItem(
      id: json['id'] ?? '',
      assetId: json['asset_id'] ?? '',
      assetName: json['asset_name'],
      maintenanceDate: json['maintenance_date'] ?? '',
      maintenanceType: json['maintenance_type'] ?? '',
      cost: (json['cost'] as num?)?.toDouble(),
      technician: json['technician'],
      notes: json['notes'],
      conditionAfter: json['condition_after'] ?? 'good',
      statusAfter: json['status_after'] ?? 'available',
    );
  }
}

class AssetState {
  final bool isLoading;
  final String? error;
  final List<AssetItem> assets;
  final List<AssetCategoryItem> categories;
  final List<AssetLocationItem> locations;
  final List<AssetLoanItem> loans;
  final List<AssetMaintenanceItem> maintenances;

  AssetState({
    this.isLoading = false,
    this.error,
    this.assets = const [],
    this.categories = const [],
    this.locations = const [],
    this.loans = const [],
    this.maintenances = const [],
  });

  AssetState copyWith({
    bool? isLoading,
    String? error,
    List<AssetItem>? assets,
    List<AssetCategoryItem>? categories,
    List<AssetLocationItem>? locations,
    List<AssetLoanItem>? loans,
    List<AssetMaintenanceItem>? maintenances,
  }) {
    return AssetState(
      isLoading: isLoading ?? this.isLoading,
      error: error,
      assets: assets ?? this.assets,
      categories: categories ?? this.categories,
      locations: locations ?? this.locations,
      loans: loans ?? this.loans,
      maintenances: maintenances ?? this.maintenances,
    );
  }
}

class AssetNotifier extends StateNotifier<AssetState> {
  final ApiClient _apiClient;

  AssetNotifier(this._apiClient) : super(AssetState());

  Future<void> loadAllData() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final resAssets = await _apiClient.dio.get('/assets');
      final resCategories = await _apiClient.dio.get('/asset-categories');
      final resLocations = await _apiClient.dio.get('/asset-locations');
      final resLoans = await _apiClient.dio.get('/asset-loans');
      final resMaintenances = await _apiClient.dio.get('/asset-maintenances');

      final assetList = ((resAssets.data['data'] ?? []) as List)
          .map((e) => AssetItem.fromJson(e))
          .toList();
      final catList = ((resCategories.data['data'] ?? []) as List)
          .map((e) => AssetCategoryItem.fromJson(e))
          .toList();
      final locList = ((resLocations.data['data'] ?? []) as List)
          .map((e) => AssetLocationItem.fromJson(e))
          .toList();
      final loanList = ((resLoans.data['data'] ?? []) as List)
          .map((e) => AssetLoanItem.fromJson(e))
          .toList();
      final maintList = ((resMaintenances.data['data'] ?? []) as List)
          .map((e) => AssetMaintenanceItem.fromJson(e))
          .toList();

      state = state.copyWith(
        isLoading: false,
        assets: assetList,
        categories: catList,
        locations: locList,
        loans: loanList,
        maintenances: maintList,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<bool> createLoan({
    required String assetId,
    required String loanDate,
    required String dueDate,
    required String conditionOut,
    String? notes,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      await _apiClient.dio.post('/asset-loans', data: {
        'asset_id': assetId,
        'loan_date': loanDate,
        'due_date': dueDate,
        'condition_out': conditionOut,
        if (notes != null) 'notes': notes,
      });
      await loadAllData();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }

  Future<bool> reviewLoan(String loanId, String action, {String? notes}) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      await _apiClient.dio.post('/asset-loans/$loanId/review', data: {
        'action': action,
        if (notes != null) 'notes': notes,
      });
      await loadAllData();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }

  Future<bool> returnLoan(String loanId, String conditionIn, {String? notes}) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      await _apiClient.dio.post('/asset-loans/$loanId/return', data: {
        'condition_in': conditionIn,
        if (notes != null) 'notes': notes,
      });
      await loadAllData();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }

  Future<bool> createMaintenance({
    required String assetId,
    required String maintenanceDate,
    required String maintenanceType,
    double? cost,
    String? technician,
    String? notes,
    required String conditionAfter,
    required String statusAfter,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      await _apiClient.dio.post('/asset-maintenances', data: {
        'asset_id': assetId,
        'maintenance_date': maintenanceDate,
        'maintenance_type': maintenanceType,
        if (cost != null) 'cost': cost,
        if (technician != null) 'technician': technician,
        if (notes != null) 'notes': notes,
        'condition_after': conditionAfter,
        'status_after': statusAfter,
      });
      await loadAllData();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }
}

final assetProvider = StateNotifierProvider<AssetNotifier, AssetState>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return AssetNotifier(apiClient);
});
