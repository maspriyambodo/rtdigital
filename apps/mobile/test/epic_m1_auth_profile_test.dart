import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/core/auth_provider.dart';
import 'package:mobile/core/storage/secure_storage_service.dart';
import 'package:mobile/core/device/device_services.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('Epic M1 Authentication & Profile Tests', () {
    test('M1.1: Masking NIK and No KK formats correctly', () {
      expect(maskSensitiveText('3175010101010001'), equals('3175************'));
      expect(maskSensitiveText('3175050505050002'), equals('3175************'));
      expect(maskSensitiveText('123'), equals('***'));
      expect(maskSensitiveText(null), equals('-'));
    });

    test('M1.2: SecureStorageService pin and biometric state persistence', () async {
      FlutterSecureStorage.setMockInitialValues({
        'biometric_enabled': 'true',
        'user_pin': '123456',
        'access_token': 'mock_access_token',
        'refresh_token': 'mock_refresh_token',
      });

      final storage = SecureStorageService();
      expect(await storage.isBiometricEnabled(), isTrue);
      expect(await storage.getPin(), equals('123456'));
      expect(await storage.getAccessToken(), equals('mock_access_token'));
      expect(await storage.getRefreshToken(), equals('mock_refresh_token'));
    });

    test('M1.3: UserProfile and HouseholdInfo JSON parsing', () {
      final userJson = {
        'id': 'user-1',
        'full_name': 'Ahmad Fauzi',
        'email': 'ahmad@example.com',
        'phone_number': '08123456789',
        'role': 'citizen',
      };
      final user = UserProfile.fromJson(userJson);
      expect(user.id, equals('user-1'));
      expect(user.name, equals('Ahmad Fauzi'));
      expect(user.email, equals('ahmad@example.com'));
      expect(user.phone, equals('08123456789'));

      final hhJson = {
        'id': 'hh-1',
        'kk_number': '3175000000000000',
        'address': 'Jl. Penggilingan No. 5',
        'members': [
          {
            'id': 'res-1',
            'full_name': 'Ahmad Fauzi',
            'nik': '3175010101010001',
            'relationship': 'head',
          }
        ]
      };
      final hh = HouseholdInfo.fromJson(hhJson);
      expect(hh.id, equals('hh-1'));
      expect(hh.kkNumber, equals('3175000000000000'));
      expect(hh.members.length, equals(1));
      expect(hh.members.first.fullName, equals('Ahmad Fauzi'));
    });
  });
}
