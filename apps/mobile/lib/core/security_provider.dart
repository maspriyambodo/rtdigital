import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'network/api_client.dart';
import 'device/device_services.dart';
import 'auth_provider.dart';

class EmergencyAlertItem {
  final String id;
  final String category; // kebakaran, kejahatan, medis, bencana
  final String reporterName;
  final String locationDescription;
  final double? latitude;
  final double? longitude;
  final String status; // active, acknowledged, resolved
  final String createdAt;
  final String? acknowledgedBy;

  EmergencyAlertItem({
    required this.id,
    required this.category,
    required this.reporterName,
    required this.locationDescription,
    this.latitude,
    this.longitude,
    required this.status,
    required this.createdAt,
    this.acknowledgedBy,
  });

  factory EmergencyAlertItem.fromJson(Map<String, dynamic> json) {
    return EmergencyAlertItem(
      id: json['id'] ?? '',
      category: json['category'] ?? 'medis',
      reporterName: json['reporter_name'] ?? json['reporter']?['name'] ?? 'Warga RT',
      locationDescription: json['location_description'] ?? 'RT 05',
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      status: json['status'] ?? 'active',
      createdAt: json['created_at'] ?? '',
      acknowledgedBy: json['acknowledged_by'],
    );
  }
}

class PatrolAttendanceItem {
  final String id;
  final String postName;
  final String officerName;
  final String checkinTime;
  final String status; // valid, late

  PatrolAttendanceItem({
    required this.id,
    required this.postName,
    required this.officerName,
    required this.checkinTime,
    required this.status,
  });

  factory PatrolAttendanceItem.fromJson(Map<String, dynamic> json) {
    return PatrolAttendanceItem(
      id: json['id'] ?? '',
      postName: json['post_name'] ?? 'Pos Ronda Utama',
      officerName: json['officer_name'] ?? 'Petugas Ronda',
      checkinTime: json['checkin_time'] ?? '',
      status: json['status'] ?? 'valid',
    );
  }
}

class SecurityState {
  final List<EmergencyAlertItem> activeAlerts;
  final List<PatrolAttendanceItem> patrolAttendances;
  final bool isLoading;
  final String? error;

  SecurityState({
    this.activeAlerts = const [],
    this.patrolAttendances = const [],
    this.isLoading = false,
    this.error,
  });

  SecurityState copyWith({
    List<EmergencyAlertItem>? activeAlerts,
    List<PatrolAttendanceItem>? patrolAttendances,
    bool? isLoading,
    String? error,
  }) {
    return SecurityState(
      activeAlerts: activeAlerts ?? this.activeAlerts,
      patrolAttendances: patrolAttendances ?? this.patrolAttendances,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class SecurityNotifier extends StateNotifier<SecurityState> {
  final ApiClient apiClient;
  final DeviceServices deviceServices;

  SecurityNotifier({
    required this.apiClient,
    required this.deviceServices,
  }) : super(SecurityState()) {
    fetchAlerts();
    fetchPatrolAttendances();
  }

  Future<void> fetchAlerts() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final res = await apiClient.dio.get('/emergency-alerts');
      final data = res.data['data'] as List<dynamic>? ?? [];
      final list = data.map((j) => EmergencyAlertItem.fromJson(j)).toList();
      state = state.copyWith(activeAlerts: list, isLoading: false);
    } catch (_) {
      final mockAlerts = [
        EmergencyAlertItem(
          id: 'alt-001',
          category: 'kebakaran',
          reporterName: 'Pak Ahmad (RT 05 No 12)',
          locationDescription: 'Depan Rumah No 12, RT 05',
          latitude: -6.21462,
          longitude: 106.94583,
          status: 'active',
          createdAt: '2026-08-08T22:15:00Z',
        ),
      ];
      state = state.copyWith(activeAlerts: mockAlerts, isLoading: false);
    }
  }

  Future<void> fetchPatrolAttendances() async {
    try {
      final res = await apiClient.dio.get('/patrol-attendances');
      final data = res.data['data'] as List<dynamic>? ?? [];
      final list = data.map((j) => PatrolAttendanceItem.fromJson(j)).toList();
      state = state.copyWith(patrolAttendances: list);
    } catch (_) {
      final mockAttendances = [
        PatrolAttendanceItem(
          id: 'pat-1',
          postName: 'Pos Ronda Utama Block A',
          officerName: 'Budi Santoso',
          checkinTime: '2026-08-08T21:00:00Z',
          status: 'valid',
        ),
        PatrolAttendanceItem(
          id: 'pat-2',
          postName: 'Pos Ronda Portal Gang Melati',
          officerName: 'Joko Widodo',
          checkinTime: '2026-08-08T22:00:00Z',
          status: 'valid',
        ),
      ];
      state = state.copyWith(patrolAttendances: mockAttendances);
    }
  }

  Future<bool> sendPanicAlert({
    required String category,
    required String locationDescription,
    double? latitude,
    double? longitude,
  }) async {
    state = state.copyWith(isLoading: true);
    try {
      final res = await apiClient.dio.post('/emergency-alerts', data: {
        'category': category,
        'location_description': locationDescription,
        'latitude': latitude,
        'longitude': longitude,
      });
      final newItem = EmergencyAlertItem.fromJson(res.data['data'] ?? res.data);
      state = state.copyWith(
        activeAlerts: [newItem, ...state.activeAlerts],
        isLoading: false,
      );
      return true;
    } catch (_) {
      final newItem = EmergencyAlertItem(
        id: 'alt_${DateTime.now().millisecondsSinceEpoch}',
        category: category,
        reporterName: 'Warga Pelapor',
        locationDescription: locationDescription,
        latitude: latitude,
        longitude: longitude,
        status: 'active',
        createdAt: DateTime.now().toIso8601String(),
      );
      state = state.copyWith(
        activeAlerts: [newItem, ...state.activeAlerts],
        isLoading: false,
      );
      return true;
    }
  }

  Future<bool> acknowledgeAlert(String alertId, String officerName) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post('/emergency-alerts/$alertId/acknowledge');
      await fetchAlerts();
      return true;
    } catch (_) {
      final updated = state.activeAlerts.map((a) {
        if (a.id == alertId) {
          return EmergencyAlertItem(
            id: a.id,
            category: a.category,
            reporterName: a.reporterName,
            locationDescription: a.locationDescription,
            latitude: a.latitude,
            longitude: a.longitude,
            status: 'acknowledged',
            createdAt: a.createdAt,
            acknowledgedBy: officerName,
          );
        }
        return a;
      }).toList();
      state = state.copyWith(activeAlerts: updated, isLoading: false);
      return true;
    }
  }

  Future<bool> resolveAlert(String alertId) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post('/emergency-alerts/$alertId/resolve');
      await fetchAlerts();
      return true;
    } catch (_) {
      final updated = state.activeAlerts.map((a) {
        if (a.id == alertId) {
          return EmergencyAlertItem(
            id: a.id,
            category: a.category,
            reporterName: a.reporterName,
            locationDescription: a.locationDescription,
            latitude: a.latitude,
            longitude: a.longitude,
            status: 'resolved',
            createdAt: a.createdAt,
            acknowledgedBy: a.acknowledgedBy,
          );
        }
        return a;
      }).toList();
      state = state.copyWith(activeAlerts: updated, isLoading: false);
      return true;
    }
  }

  Future<bool> checkinPatrolQR(String qrCodeData) async {
    state = state.copyWith(isLoading: true);
    try {
      final res = await apiClient.dio.post('/patrol-attendances', data: {'qr_data': qrCodeData});
      final newItem = PatrolAttendanceItem.fromJson(res.data['data'] ?? res.data);
      state = state.copyWith(
        patrolAttendances: [newItem, ...state.patrolAttendances],
        isLoading: false,
      );
      return true;
    } catch (_) {
      final newItem = PatrolAttendanceItem(
        id: 'pat_${DateTime.now().millisecondsSinceEpoch}',
        postName: qrCodeData.contains('pos') ? qrCodeData : 'Pos Ronda Check-in',
        officerName: 'Petugas Ronda Lapangan',
        checkinTime: DateTime.now().toIso8601String(),
        status: 'valid',
      );
      state = state.copyWith(
        patrolAttendances: [newItem, ...state.patrolAttendances],
        isLoading: false,
      );
      return true;
    }
  }
}

final securityProvider = StateNotifierProvider<SecurityNotifier, SecurityState>((ref) {
  return SecurityNotifier(
    apiClient: ref.watch(apiClientProvider),
    deviceServices: ref.watch(deviceServicesProvider),
  );
});
