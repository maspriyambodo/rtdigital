import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'storage/secure_storage_service.dart';
import 'device/device_services.dart';
import 'network/api_client.dart';

class UserProfile {
  final String id;
  final String name;
  final String email;
  final String phone;
  final String role;

  UserProfile({
    required this.id,
    required this.name,
    required this.email,
    required this.phone,
    required this.role,
  });

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    return UserProfile(
      id: json['id'] ?? '',
      name: json['full_name'] ?? json['name'] ?? '',
      email: json['email'] ?? '',
      phone: json['phone_number'] ?? json['phone'] ?? '',
      role: json['role'] ?? 'citizen',
    );
  }
}

class HouseholdMember {
  final String id;
  final String fullName;
  final String nik;
  final String relationship;

  HouseholdMember({
    required this.id,
    required this.fullName,
    required this.nik,
    required this.relationship,
  });

  factory HouseholdMember.fromJson(Map<String, dynamic> json) {
    return HouseholdMember(
      id: json['id'] ?? '',
      fullName: json['full_name'] ?? json['name'] ?? '',
      nik: json['nik'] ?? '',
      relationship: json['relationship'] ?? 'member',
    );
  }
}

class HouseholdInfo {
  final String id;
  final String kkNumber;
  final String address;
  final List<HouseholdMember> members;

  HouseholdInfo({
    required this.id,
    required this.kkNumber,
    required this.address,
    required this.members,
  });

  factory HouseholdInfo.fromJson(Map<String, dynamic> json) {
    final rawMembers = json['members'] as List<dynamic>? ?? [];
    return HouseholdInfo(
      id: json['id'] ?? '',
      kkNumber: json['kk_number'] ?? json['internal_number'] ?? '',
      address: json['address'] ?? '',
      members: rawMembers.map((m) => HouseholdMember.fromJson(m)).toList(),
    );
  }
}

class AuthState {
  final bool isAuthenticated;
  final bool isLoading;
  final String? error;
  final UserProfile? user;
  final HouseholdInfo? household;

  AuthState({
    this.isAuthenticated = false,
    this.isLoading = false,
    this.error,
    this.user,
    this.household,
  });

  AuthState copyWith({
    bool? isAuthenticated,
    bool? isLoading,
    String? error,
    UserProfile? user,
    HouseholdInfo? household,
  }) {
    return AuthState(
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      user: user ?? this.user,
      household: household ?? this.household,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  final SecureStorageService storage;
  final DeviceServices deviceServices;
  final ApiClient apiClient;


  AuthNotifier({
    required this.storage,
    required this.deviceServices,
    required this.apiClient,
  }) : super(AuthState()) {
    checkAuthStatus();
  }

  Future<void> checkAuthStatus() async {
    final token = await storage.getAccessToken();
    if (token != null && token.isNotEmpty) {
      state = state.copyWith(isAuthenticated: true);
      await fetchProfile();
    } else {
      state = state.copyWith(isAuthenticated: false);
    }
  }

  Future<bool> login(String identifier, String password) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final res = await apiClient.dio.post('/auth/login', data: {
        'login': identifier,
        'password': password,
      });
      final data = res.data['data'] ?? res.data;
      final accessToken = data['access_token'] ?? data['accessToken'];
      final refreshToken = data['refresh_token'] ?? data['refreshToken'];

      if (accessToken != null && refreshToken != null) {
        await storage.saveTokens(accessToken: accessToken, refreshToken: refreshToken);
        state = state.copyWith(isAuthenticated: true, isLoading: false);
        await fetchProfile();
        return true;
      }
      state = state.copyWith(isLoading: false, error: 'Respon login tidak valid');
      return false;
    } on DioException catch (e) {
      final msg = e.response?.data?['message'] ?? 'Login gagal. Periksa kredensial Anda.';
      state = state.copyWith(isLoading: false, error: msg);
      return false;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: 'Terjadi kesalahan sistem');
      return false;
    }
  }

  Future<bool> loginWithBiometrics() async {
    final enabled = await storage.isBiometricEnabled();
    if (!enabled) {
      state = state.copyWith(error: 'Biometrik belum diaktifkan');
      return false;
    }
    final canCheck = await deviceServices.canCheckBiometrics();
    if (!canCheck) {
      state = state.copyWith(error: 'Perangkat tidak mendukung biometrik');
      return false;
    }
    final authenticated = await deviceServices.authenticateBiometric(
      reason: 'Masuk ke akun RT Digital',
    );
    if (!authenticated) {
      return false;
    }
    final token = await storage.getAccessToken();
    if (token != null && token.isNotEmpty) {
      state = state.copyWith(isAuthenticated: true);
      await fetchProfile();
      return true;
    }
    state = state.copyWith(error: 'Sesi kedaluwarsa, silakan login dengan kata sandi');
    return false;
  }

  Future<bool> verifyPin(String pin) async {
    final savedPin = await storage.getPin();
    if (savedPin == pin) {
      final token = await storage.getAccessToken();
      if (token != null && token.isNotEmpty) {
        state = state.copyWith(isAuthenticated: true);
        await fetchProfile();
        return true;
      }
    }
    state = state.copyWith(error: 'PIN salah atau tidak ditemukan');
    return false;
  }

  Future<void> fetchProfile() async {
    try {
      final res = await apiClient.dio.get('/me');
      final data = res.data['data'] ?? res.data;
      final user = UserProfile.fromJson(data);
      state = state.copyWith(user: user);

      if (data['household'] != null) {
        state = state.copyWith(household: HouseholdInfo.fromJson(data['household']));
      } else {
        try {
          final hhRes = await apiClient.dio.get('/households');
          final hhData = hhRes.data['data'] as List<dynamic>?;
          if (hhData != null && hhData.isNotEmpty) {
            state = state.copyWith(household: HouseholdInfo.fromJson(hhData.first));
          }
        } catch (_) {}
      }
    } catch (e) {
      // Ignore background fetch error
    }
  }

  Future<bool> submitCorrection({
    required String residentId,
    required String field,
    required String newValue,
    required String reason,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      await apiClient.dio.post('/residents/corrections', data: {
        'resident_id': residentId,
        'field': field,
        'proposed_value': newValue,
        'reason': reason,
      });
      state = state.copyWith(isLoading: false);
      return true;
    } on DioException catch (e) {
      final msg = e.response?.data?['message'] ?? 'Gagal mengajukan koreksi';
      state = state.copyWith(isLoading: false, error: msg);
      return false;
    } catch (_) {
      state = state.copyWith(isLoading: false, error: 'Terjadi kesalahan sistem');
      return false;
    }
  }

  Future<void> logout({bool allDevices = false}) async {
    try {
      final endpoint = allDevices ? '/auth/logout-all' : '/auth/logout';
      await apiClient.dio.post(endpoint);
    } catch (_) {}
    await storage.clearTokens();
    state = AuthState(isAuthenticated: false);
  }
}

final secureStorageProvider = Provider<SecureStorageService>((ref) => SecureStorageService());
final deviceServicesProvider = Provider<DeviceServices>((ref) => DeviceServices());
final apiClientProvider = Provider<ApiClient>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return ApiClient(
    baseUrl: 'http://localhost:8080/api/v1',
    storageService: storage,
  );
});

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(
    storage: ref.watch(secureStorageProvider),
    deviceServices: ref.watch(deviceServicesProvider),
    apiClient: ref.watch(apiClientProvider),
  );
});

String maskSensitiveText(String? value) {
  if (value == null || value.isEmpty) return '-';
  if (value.length <= 4) return '*' * value.length;
  return '${value.substring(0, 4)}${'*' * (value.length - 4)}';
}

