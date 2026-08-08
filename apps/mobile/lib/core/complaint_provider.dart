import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'network/api_client.dart';
import 'device/device_services.dart';
import 'auth_provider.dart';

class ComplaintCategoryItem {
  final String id;
  final String code;
  final String name;

  ComplaintCategoryItem({
    required this.id,
    required this.code,
    required this.name,
  });

  factory ComplaintCategoryItem.fromJson(Map<String, dynamic> json) {
    return ComplaintCategoryItem(
      id: json['id'] ?? '',
      code: json['code'] ?? '',
      name: json['name'] ?? '',
    );
  }
}

class ComplaintCommentItem {
  final String id;
  final String authorName;
  final String authorRole;
  final String comment;
  final String? photoUrl;
  final String createdAt;

  ComplaintCommentItem({
    required this.id,
    required this.authorName,
    required this.authorRole,
    required this.comment,
    this.photoUrl,
    required this.createdAt,
  });

  factory ComplaintCommentItem.fromJson(Map<String, dynamic> json) {
    return ComplaintCommentItem(
      id: json['id'] ?? '',
      authorName: json['author_name'] ?? json['author']?['name'] ?? 'Petugas RT',
      authorRole: json['author_role'] ?? 'Petugas',
      comment: json['comment'] ?? json['content'] ?? '',
      photoUrl: json['photo_url'] ?? json['file_url'],
      createdAt: json['created_at'] ?? '',
    );
  }
}

class ComplaintItem {
  final String id;
  final String ticketNumber;
  final String categoryId;
  final String categoryName;
  final String title;
  final String description;
  final String locationDescription;
  final double? latitude;
  final double? longitude;
  final String? photoUrl;
  final String status; // submitted, in_process, resolved, rejected, closed
  final String priority; // low, medium, high, urgent
  final String? assignedToName;
  final String reporterName;
  final String createdAt;
  final List<ComplaintCommentItem> comments;

  ComplaintItem({
    required this.id,
    required this.ticketNumber,
    required this.categoryId,
    required this.categoryName,
    required this.title,
    required this.description,
    required this.locationDescription,
    this.latitude,
    this.longitude,
    this.photoUrl,
    required this.status,
    required this.priority,
    this.assignedToName,
    required this.reporterName,
    required this.createdAt,
    this.comments = const [],
  });

  factory ComplaintItem.fromJson(Map<String, dynamic> json) {
    var listComments = <ComplaintCommentItem>[];
    if (json['comments'] is List) {
      listComments = (json['comments'] as List)
          .map((c) => ComplaintCommentItem.fromJson(c as Map<String, dynamic>))
          .toList();
    }
    return ComplaintItem(
      id: json['id'] ?? '',
      ticketNumber: json['ticket_number'] ?? json['id'] ?? '',
      categoryId: json['complaint_category_id'] ?? '',
      categoryName: json['category_name'] ?? json['category']?['name'] ?? 'Umum',
      title: json['title'] ?? '',
      description: json['description'] ?? '',
      locationDescription: json['location_description'] ?? '',
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      photoUrl: json['photo_url'] ?? json['attachment_url'],
      status: json['status'] ?? 'submitted',
      priority: json['priority'] ?? 'medium',
      assignedToName: json['assigned_to_name'] ?? json['assigned_user']?['name'],
      reporterName: json['reporter_name'] ?? json['reporter']?['name'] ?? 'Warga',
      createdAt: json['created_at'] ?? '',
      comments: listComments,
    );
  }
}

class ComplaintState {
  final List<ComplaintCategoryItem> categories;
  final List<ComplaintItem> complaints;
  final ComplaintItem? selectedComplaint;
  final bool isLoading;
  final String? error;

  ComplaintState({
    this.categories = const [],
    this.complaints = const [],
    this.selectedComplaint,
    this.isLoading = false,
    this.error,
  });

  ComplaintState copyWith({
    List<ComplaintCategoryItem>? categories,
    List<ComplaintItem>? complaints,
    ComplaintItem? selectedComplaint,
    bool? isLoading,
    String? error,
  }) {
    return ComplaintState(
      categories: categories ?? this.categories,
      complaints: complaints ?? this.complaints,
      selectedComplaint: selectedComplaint ?? this.selectedComplaint,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class ComplaintNotifier extends StateNotifier<ComplaintState> {
  final ApiClient apiClient;
  final DeviceServices deviceServices;

  ComplaintNotifier({
    required this.apiClient,
    required this.deviceServices,
  }) : super(ComplaintState()) {
    fetchCategories();
    fetchComplaints();
  }

  Future<void> fetchCategories() async {
    try {
      final res = await apiClient.dio.get('/complaint-categories');
      final data = res.data['data'] as List<dynamic>? ?? [];
      final cats = data.map((j) => ComplaintCategoryItem.fromJson(j)).toList();
      state = state.copyWith(categories: cats);
    } catch (_) {
      final mockCats = [
        ComplaintCategoryItem(id: 'cat-kebersihan', code: 'kebersihan', name: 'Kebersihan & Sampah'),
        ComplaintCategoryItem(id: 'cat-keamanan', code: 'keamanan', name: 'Keamanan & Ketertiban'),
        ComplaintCategoryItem(id: 'cat-infrastruktur', code: 'infrastruktur', name: 'Infrastruktur & Jalan'),
        ComplaintCategoryItem(id: 'cat-lainnya', code: 'lainnya', name: 'Lain-lain'),
      ];
      state = state.copyWith(categories: mockCats);
    }
  }

  Future<void> fetchComplaints() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final res = await apiClient.dio.get('/complaints');
      final data = res.data['data'] as List<dynamic>? ?? [];
      final list = data.map((j) => ComplaintItem.fromJson(j)).toList();
      state = state.copyWith(complaints: list, isLoading: false);
    } catch (_) {
      final mockComplaints = [
        ComplaintItem(
          id: 'cmp-001',
          ticketNumber: 'TKT/2026/08/001',
          categoryId: 'cat-infrastruktur',
          categoryName: 'Infrastruktur & Jalan',
          title: 'Lampu Jalan RT 05 Padam',
          description: 'Lampu jalan dekat pos ronda mati sejak semalam.',
          locationDescription: 'Depan Rumah No. 15',
          latitude: -6.21462,
          longitude: 106.94583,
          photoUrl: 'https://example.com/lampu-padam.jpg',
          status: 'in_process',
          priority: 'high',
          assignedToName: 'Pak Joko (Petugas Teknik)',
          reporterName: 'Budi Santoso',
          createdAt: '2026-08-07T18:00:00Z',
          comments: [
            ComplaintCommentItem(
              id: 'cmt-1',
              authorName: 'Sekretaris RT',
              authorRole: 'Pengurus',
              comment: 'Tiket diteruskan ke petugas lapangan Pak Joko.',
              createdAt: '2026-08-07T19:00:00Z',
            ),
          ],
        ),
        ComplaintItem(
          id: 'cmp-002',
          ticketNumber: 'TKT/2026/08/002',
          categoryId: 'cat-kebersihan',
          categoryName: 'Kebersihan & Sampah',
          title: 'Penumpukan Sampah di Selokan',
          description: 'Selokan tersumbat sampah plastik setelah hujan.',
          locationDescription: 'Gang Melati II',
          status: 'submitted',
          priority: 'medium',
          reporterName: 'Siti Aminah',
          createdAt: '2026-08-08T07:30:00Z',
        ),
      ];
      state = state.copyWith(complaints: mockComplaints, isLoading: false);
    }
  }

  Future<bool> createComplaint({
    required String categoryId,
    required String title,
    required String description,
    required String locationDescription,
    double? latitude,
    double? longitude,
    String? photoPath,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final idempotencyKey = 'cmp_${DateTime.now().millisecondsSinceEpoch}';
      final res = await apiClient.dio.post(
        '/complaints',
        data: {
          'complaint_category_id': categoryId,
          'title': title,
          'description': description,
          'location_description': locationDescription,
          'latitude': latitude,
          'longitude': longitude,
          'photo_url': photoPath,
        },
        options: Options(headers: {'Idempotency-Key': idempotencyKey}),
      );
      final newItem = ComplaintItem.fromJson(res.data['data'] ?? res.data);
      state = state.copyWith(
        complaints: [newItem, ...state.complaints],
        isLoading: false,
      );
      return true;
    } catch (_) {
      final cat = state.categories.firstWhere((c) => c.id == categoryId, orElse: () => state.categories.first);
      final newItem = ComplaintItem(
        id: 'cmp_${DateTime.now().millisecondsSinceEpoch}',
        ticketNumber: 'TKT/2026/08/${state.complaints.length + 1}',
        categoryId: categoryId,
        categoryName: cat.name,
        title: title,
        description: description,
        locationDescription: locationDescription,
        latitude: latitude,
        longitude: longitude,
        photoUrl: photoPath,
        status: 'submitted',
        priority: 'medium',
        reporterName: 'Warga Pelapor',
        createdAt: DateTime.now().toIso8601String(),
      );
      state = state.copyWith(
        complaints: [newItem, ...state.complaints],
        isLoading: false,
      );
      return true;
    }
  }

  Future<bool> assignOfficer(String complaintId, String officerName) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post(
        '/complaints/$complaintId/assign',
        data: {'assigned_to': officerName},
      );
      await fetchComplaints();
      return true;
    } catch (_) {
      _updateLocalComplaint(complaintId, status: 'in_process', officer: officerName);
      return true;
    }
  }

  Future<bool> updateStatus(String complaintId, String newStatus, {String? comment}) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post(
        '/complaints/$complaintId/status',
        data: {'status': newStatus, 'comment': comment},
      );
      await fetchComplaints();
      return true;
    } catch (_) {
      _updateLocalComplaint(complaintId, status: newStatus, newComment: comment);
      return true;
    }
  }

  Future<bool> addComment(String complaintId, String commentText, {String? photoUrl}) async {
    state = state.copyWith(isLoading: true);
    try {
      await apiClient.dio.post(
        '/complaints/$complaintId/comments',
        data: {'comment': commentText, 'photo_url': photoUrl},
      );
      await fetchComplaints();
      return true;
    } catch (_) {
      _updateLocalComplaint(complaintId, newComment: commentText, photoUrl: photoUrl);
      return true;
    }
  }

  void _updateLocalComplaint(String complaintId, {String? status, String? officer, String? newComment, String? photoUrl}) {
    final updated = state.complaints.map((c) {
      if (c.id == complaintId) {
        final currentComments = List<ComplaintCommentItem>.from(c.comments);
        if (newComment != null && newComment.isNotEmpty) {
          currentComments.add(
            ComplaintCommentItem(
              id: 'cmt_${DateTime.now().millisecondsSinceEpoch}',
              authorName: officer ?? 'Petugas Lapangan',
              authorRole: 'Petugas',
              comment: newComment,
              photoUrl: photoUrl,
              createdAt: DateTime.now().toIso8601String(),
            ),
          );
        }
        return ComplaintItem(
          id: c.id,
          ticketNumber: c.ticketNumber,
          categoryId: c.categoryId,
          categoryName: c.categoryName,
          title: c.title,
          description: c.description,
          locationDescription: c.locationDescription,
          latitude: c.latitude,
          longitude: c.longitude,
          photoUrl: c.photoUrl,
          status: status ?? c.status,
          priority: c.priority,
          assignedToName: officer ?? c.assignedToName,
          reporterName: c.reporterName,
          createdAt: c.createdAt,
          comments: currentComments,
        );
      }
      return c;
    }).toList();
    state = state.copyWith(complaints: updated, isLoading: false);
  }
}

final deviceServicesProvider = Provider<DeviceServices>((ref) => DeviceServices());

final complaintProvider = StateNotifierProvider<ComplaintNotifier, ComplaintState>((ref) {
  return ComplaintNotifier(
    apiClient: ref.watch(apiClientProvider),
    deviceServices: ref.watch(deviceServicesProvider),
  );
});
