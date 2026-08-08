import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'network/api_client.dart';
import 'auth_provider.dart';

class PatrolPostItem {
  final String id;
  final String code;
  final String name;
  final String? location;
  final String status;

  PatrolPostItem({
    required this.id,
    required this.code,
    required this.name,
    this.location,
    required this.status,
  });

  factory PatrolPostItem.fromJson(Map<String, dynamic> json) {
    return PatrolPostItem(
      id: json['id'] as String? ?? '',
      code: json['code'] as String? ?? '',
      name: json['name'] as String? ?? '',
      location: json['location'] as String?,
      status: json['status'] as String? ?? 'active',
    );
  }
}

class EmergencyAlertItem {
  final String id;
  final String reporterId;
  final String? reporterName;
  final String category;
  final double? latitude;
  final double? longitude;
  final String? locationDetails;
  final String status;
  final String createdAt;

  EmergencyAlertItem({
    required this.id,
    required this.reporterId,
    this.reporterName,
    required this.category,
    this.latitude,
    this.longitude,
    this.locationDetails,
    required this.status,
    required this.createdAt,
  });

  factory EmergencyAlertItem.fromJson(Map<String, dynamic> json) {
    return EmergencyAlertItem(
      id: json['id'] as String? ?? '',
      reporterId: json['reporter_id'] as String? ?? '',
      reporterName: json['reporter_name'] as String?,
      category: json['category'] as String? ?? 'other',
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      locationDetails: json['location_details'] as String?,
      status: json['status'] as String? ?? 'active',
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}

class VisitorInviteItem {
  final String id;
  final String visitorName;
  final String? purpose;
  final String validFrom;
  final String validUntil;
  final String qrCodeHash;
  final String status;

  VisitorInviteItem({
    required this.id,
    required this.visitorName,
    this.purpose,
    required this.validFrom,
    required this.validUntil,
    required this.qrCodeHash,
    required this.status,
  });

  factory VisitorInviteItem.fromJson(Map<String, dynamic> json) {
    return VisitorInviteItem(
      id: json['id'] as String? ?? '',
      visitorName: json['visitor_name'] as String? ?? '',
      purpose: json['purpose'] as String?,
      validFrom: json['valid_from'] as String? ?? '',
      validUntil: json['valid_until'] as String? ?? '',
      qrCodeHash: json['qr_code_hash'] as String? ?? '',
      status: json['status'] as String? ?? 'active',
    );
  }
}

class VisitorLogItem {
  final String id;
  final String visitorName;
  final String? identityType;
  final String? identityNumber;
  final String? vehiclePlate;
  final String? purpose;
  final String checkInTime;
  final String status;

  VisitorLogItem({
    required this.id,
    required this.visitorName,
    this.identityType,
    this.identityNumber,
    this.vehiclePlate,
    this.purpose,
    required this.checkInTime,
    required this.status,
  });

  factory VisitorLogItem.fromJson(Map<String, dynamic> json) {
    return VisitorLogItem(
      id: json['id'] as String? ?? '',
      visitorName: json['visitor_name'] as String? ?? '',
      identityType: json['identity_type'] as String?,
      identityNumber: json['identity_number'] as String?,
      vehiclePlate: json['vehicle_plate'] as String?,
      purpose: json['purpose'] as String?,
      checkInTime: json['check_in_time'] as String? ?? '',
      status: json['status'] as String? ?? 'checked_in',
    );
  }
}

class CommunityActivityItem {
  final String id;
  final String title;
  final String? description;
  final String activityDate;
  final String startTime;
  final String? endTime;
  final String? location;
  final String targetType;
  final bool isMandatory;
  final String status;

  CommunityActivityItem({
    required this.id,
    required this.title,
    this.description,
    required this.activityDate,
    required this.startTime,
    this.endTime,
    this.location,
    required this.targetType,
    required this.isMandatory,
    required this.status,
  });

  factory CommunityActivityItem.fromJson(Map<String, dynamic> json) {
    return CommunityActivityItem(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      description: json['description'] as String?,
      activityDate: json['activity_date'] as String? ?? '',
      startTime: json['start_time'] as String? ?? '',
      endTime: json['end_time'] as String?,
      location: json['location'] as String?,
      targetType: json['target_type'] as String? ?? 'all',
      isMandatory: json['is_mandatory'] as bool? ?? false,
      status: json['status'] as String? ?? 'scheduled',
    );
  }
}

class SecurityOpsState {
  final bool isLoading;
  final String? errorMessage;
  final List<PatrolPostItem> posts;
  final List<EmergencyAlertItem> alerts;
  final List<VisitorLogItem> visitorLogs;
  final List<CommunityActivityItem> activities;

  const SecurityOpsState({
    this.isLoading = false,
    this.errorMessage,
    this.posts = const [],
    this.alerts = const [],
    this.visitorLogs = const [],
    this.activities = const [],
  });

  SecurityOpsState copyWith({
    bool? isLoading,
    String? errorMessage,
    List<PatrolPostItem>? posts,
    List<EmergencyAlertItem>? alerts,
    List<VisitorLogItem>? visitorLogs,
    List<CommunityActivityItem>? activities,
  }) {
    return SecurityOpsState(
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      posts: posts ?? this.posts,
      alerts: alerts ?? this.alerts,
      visitorLogs: visitorLogs ?? this.visitorLogs,
      activities: activities ?? this.activities,
    );
  }
}

class SecurityOpsNotifier extends StateNotifier<SecurityOpsState> {
  final ApiClient _apiClient;

  SecurityOpsNotifier(this._apiClient) : super(const SecurityOpsState()) {
    fetchInitialData();
  }

  Future<void> fetchInitialData() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final postsRes = await _apiClient.dio.get('/patrol-posts');
      final alertsRes = await _apiClient.dio.get('/emergency-alerts');
      final visitorLogsRes = await _apiClient.dio.get('/visitor-logs');
      final activitiesRes = await _apiClient.dio.get('/community-activities');

      final postsList = ((postsRes.data['data'] ?? []) as List)
          .map((e) => PatrolPostItem.fromJson(e as Map<String, dynamic>))
          .toList();
      final alertsList = ((alertsRes.data['data'] ?? []) as List)
          .map((e) => EmergencyAlertItem.fromJson(e as Map<String, dynamic>))
          .toList();
      final logsList = ((visitorLogsRes.data['data'] ?? []) as List)
          .map((e) => VisitorLogItem.fromJson(e as Map<String, dynamic>))
          .toList();
      final actList = ((activitiesRes.data['data'] ?? []) as List)
          .map((e) => CommunityActivityItem.fromJson(e as Map<String, dynamic>))
          .toList();

      state = state.copyWith(
        isLoading: false,
        posts: postsList,
        alerts: alertsList,
        visitorLogs: logsList,
        activities: actList,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, errorMessage: e.toString());
    }
  }

  Future<bool> triggerEmergencyAlert(String category, String details) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      await _apiClient.dio.post('/emergency-alerts', data: {
        'category': category,
        'location_details': details,
      });
      await fetchInitialData();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, errorMessage: e.toString());
      return false;
    }
  }

  Future<bool> createVisitorInvite(String name, String purpose, String validFrom, String validUntil) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      await _apiClient.dio.post('/visitor-invites', data: {
        'visitor_name': name,
        'purpose': purpose,
        'valid_from': validFrom,
        'valid_until': validUntil,
      });
      await fetchInitialData();
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, errorMessage: e.toString());
      return false;
    }
  }
}

final securityOpsProvider = StateNotifierProvider<SecurityOpsStateNotifier, SecurityOpsState>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return SecurityOpsStateNotifier(apiClient);
});

typedef SecurityOpsStateNotifier = SecurityOpsNotifier;
