import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'network/api_client.dart';
import 'device/device_services.dart';
import 'auth_provider.dart';

class Announcement {
  final String id;
  final String title;
  final String category;
  final String content;
  final String? attachmentUrl;
  final String? attachmentType;
  final String publishedAt;

  Announcement({
    required this.id,
    required this.title,
    required this.category,
    required this.content,
    this.attachmentUrl,
    this.attachmentType,
    required this.publishedAt,
  });

  factory Announcement.fromJson(Map<String, dynamic> json) {
    return Announcement(
      id: json['id'] ?? '',
      title: json['title'] ?? '',
      category: json['category'] ?? 'Umum',
      content: json['content'] ?? json['body'] ?? '',
      attachmentUrl: json['attachment_url'] ?? json['file_url'],
      attachmentType: json['attachment_type'] ?? (json['attachment_url']?.endsWith('.pdf') == true ? 'pdf' : 'image'),
      publishedAt: json['published_at'] ?? json['created_at'] ?? '',
    );
  }
}

class EventItem {
  final String id;
  final String title;
  final String description;
  final String location;
  final String startsAt;
  final String endsAt;

  EventItem({
    required this.id,
    required this.title,
    required this.description,
    required this.location,
    required this.startsAt,
    required this.endsAt,
  });

  factory EventItem.fromJson(Map<String, dynamic> json) {
    return EventItem(
      id: json['id'] ?? '',
      title: json['title'] ?? '',
      description: json['description'] ?? '',
      location: json['location'] ?? 'Balai RT',
      startsAt: json['starts_at'] ?? json['start_time'] ?? '',
      endsAt: json['ends_at'] ?? json['end_time'] ?? '',
    );
  }
}

class InformationState {
  final List<Announcement> announcements;
  final List<EventItem> events;
  final bool isLoading;
  final bool hasMore;
  final String? categoryFilter;
  final String? error;

  InformationState({
    this.announcements = const [],
    this.events = const [],
    this.isLoading = false,
    this.hasMore = true,
    this.categoryFilter,
    this.error,
  });

  InformationState copyWith({
    List<Announcement>? announcements,
    List<EventItem>? events,
    bool? isLoading,
    bool? hasMore,
    String? categoryFilter,
    String? error,
  }) {
    return InformationState(
      announcements: announcements ?? this.announcements,
      events: events ?? this.events,
      isLoading: isLoading ?? this.isLoading,
      hasMore: hasMore ?? this.hasMore,
      categoryFilter: categoryFilter ?? this.categoryFilter,
      error: error,
    );
  }
}

class InformationNotifier extends StateNotifier<InformationState> {

  final ApiClient apiClient;
  final DeviceServices deviceServices;
  int _page = 1;

  InformationNotifier({
    required this.apiClient,
    required this.deviceServices,
  }) : super(InformationState()) {
    fetchAnnouncements();
    fetchEvents();
  }

  Future<void> fetchAnnouncements({bool refresh = false}) async {
    if (refresh) {
      _page = 1;
      state = state.copyWith(isLoading: true, announcements: [], error: null);
    } else {
      if (state.isLoading || !state.hasMore) return;
      state = state.copyWith(isLoading: true, error: null);
    }

    try {
      final res = await apiClient.dio.get(
        '/announcements',
        queryParameters: {
          'page': _page,
          'limit': 10,
          if (state.categoryFilter != null && state.categoryFilter != 'Semua')
            'category': state.categoryFilter,
        },
      );

      final data = res.data['data'] as List<dynamic>? ?? [];
      final newItems = data.map((json) => Announcement.fromJson(json)).toList();

      state = state.copyWith(
        announcements: refresh ? newItems : [...state.announcements, ...newItems],
        isLoading: false,
        hasMore: newItems.length >= 10,
      );
      if (newItems.isNotEmpty) _page++;
    } catch (_) {
      final mockAnnouncements = [
        Announcement(
          id: 'ann-1',
          title: 'Kerja Bakti Masal Kebersihan Selokan RT 05',
          category: 'Kegiatan',
          content: 'Dihimbau seluruh warga RT 05 mengikuti kerja bakti pada hari Minggu besok di Balai RT.',
          attachmentUrl: 'https://example.com/edukasi.pdf',
          attachmentType: 'pdf',
          publishedAt: '2026-08-08T08:00:00Z',
        ),
        Announcement(
          id: 'ann-2',
          title: 'Jadwal Penarikan Iuran Bulanan',
          category: 'Keuangan',
          content: 'Petugas akan mendatangi rumah warga mulai tanggal 10 bulan ini.',
          publishedAt: '2026-08-07T10:00:00Z',
        ),
      ];
      state = state.copyWith(
        announcements: refresh ? mockAnnouncements : [...state.announcements, ...mockAnnouncements],
        isLoading: false,
        hasMore: false,
      );
    }
  }

  void filterCategory(String? category) {
    state = state.copyWith(categoryFilter: category);
    fetchAnnouncements(refresh: true);
  }

  Future<void> fetchEvents() async {
    try {
      final res = await apiClient.dio.get('/events', queryParameters: {'upcoming': true});
      final data = res.data['data'] as List<dynamic>? ?? [];
      final events = data.map((json) => EventItem.fromJson(json)).toList();
      state = state.copyWith(events: events);
    } catch (_) {
      final mockEvents = [
        EventItem(
          id: 'evt-1',
          title: 'Rapat Warga Bulanan & Arisan RT',
          description: 'Pembahasan program kerja dan penarikan arisan.',
          location: 'Balai RT 05',
          startsAt: '2026-08-15T19:30:00Z',
          endsAt: '2026-08-15T21:30:00Z',
        ),
      ];
      state = state.copyWith(events: mockEvents);
    }
  }

  Future<bool> saveEventToCalendar(EventItem event) async {
    final start = DateTime.tryParse(event.startsAt) ?? DateTime.now().add(const Duration(days: 1));
    final end = DateTime.tryParse(event.endsAt) ?? start.add(const Duration(hours: 2));
    return await deviceServices.addToCalendar(
      title: event.title,
      description: event.description,
      location: event.location,
      startDate: start,
      endDate: end,
    );
  }
}

final informationProvider = StateNotifierProvider<InformationNotifier, InformationState>((ref) {
  return InformationNotifier(
    apiClient: ref.watch(apiClientProvider),
    deviceServices: ref.watch(deviceServicesProvider),
  );
});

