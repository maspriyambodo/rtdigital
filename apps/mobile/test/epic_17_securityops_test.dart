import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/securityops_provider.dart';

void main() {
  group('Epic 17 SecurityOps Models Tests', () {
    test('PatrolPostItem fromJson parses correctly', () {
      final json = {
        'id': 'post-1',
        'code': 'POS-01',
        'name': 'Pos Utama',
        'location': 'Gerbang Utama',
        'status': 'active',
      };
      final post = PatrolPostItem.fromJson(json);
      expect(post.id, 'post-1');
      expect(post.code, 'POS-01');
      expect(post.name, 'Pos Utama');
      expect(post.status, 'active');
    });

    test('EmergencyAlertItem fromJson parses correctly', () {
      final json = {
        'id': 'alert-1',
        'reporter_id': 'res-1',
        'reporter_name': 'Pak Budi',
        'category': 'fire',
        'latitude': -6.1234,
        'longitude': 106.5678,
        'location_details': 'Rumah No. 12',
        'status': 'active',
        'created_at': '2026-08-09T10:00:00Z',
      };
      final alert = EmergencyAlertItem.fromJson(json);
      expect(alert.id, 'alert-1');
      expect(alert.category, 'fire');
      expect(alert.latitude, -6.1234);
      expect(alert.status, 'active');
    });

    test('VisitorLogItem fromJson parses correctly', () {
      final json = {
        'id': 'log-1',
        'visitor_name': 'Budi Tamu',
        'identity_type': 'KTP',
        'identity_number': '31712345678',
        'vehicle_plate': 'B 1234 CD',
        'purpose': 'Kunjungan keluarga',
        'check_in_time': '2026-08-09T14:00:00Z',
        'status': 'checked_in',
      };
      final log = VisitorLogItem.fromJson(json);
      expect(log.id, 'log-1');
      expect(log.visitorName, 'Budi Tamu');
      expect(log.vehiclePlate, 'B 1234 CD');
      expect(log.status, 'checked_in');
    });

    test('CommunityActivityItem fromJson parses correctly', () {
      final json = {
        'id': 'act-1',
        'title': 'Gotong Royong',
        'description': 'Pembersihan selokan',
        'activity_date': '2026-08-17',
        'start_time': '07:00',
        'end_time': '10:00',
        'location': 'Lapangan RT',
        'target_type': 'all',
        'is_mandatory': true,
        'status': 'scheduled',
      };
      final act = CommunityActivityItem.fromJson(json);
      expect(act.id, 'act-1');
      expect(act.title, 'Gotong Royong');
      expect(act.isMandatory, isTrue);
      expect(act.status, 'scheduled');
    });
  });
}

