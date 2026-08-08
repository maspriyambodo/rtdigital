import 'package:permission_handler/permission_handler.dart';
import 'package:local_auth/local_auth.dart';
import 'package:image_picker/image_picker.dart';
import 'package:geolocator/geolocator.dart';

class DeviceServices {
  final LocalAuthentication _localAuth;
  final ImagePicker _imagePicker;

  DeviceServices({
    LocalAuthentication? localAuth,
    ImagePicker? imagePicker,
  })  : _localAuth = localAuth ?? LocalAuthentication(),
        _imagePicker = imagePicker ?? ImagePicker();

  Future<bool> requestCameraPermission() async {
    final status = await Permission.camera.request();
    return status.isGranted;
  }

  Future<bool> requestLocationPermission() async {
    final status = await Permission.locationWhenInUse.request();
    return status.isGranted;
  }

  Future<XFile?> pickImageFromCamera() async {
    return await _imagePicker.pickImage(source: ImageSource.camera);
  }

  Future<XFile?> pickImageFromGallery() async {
    return await _imagePicker.pickImage(source: ImageSource.gallery);
  }

  Future<Position?> getCurrentLocation() async {
    final isEnabled = await Geolocator.isLocationServiceEnabled();
    if (!isEnabled) return null;
    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) return null;
    }
    if (permission == LocationPermission.deniedForever) return null;
    return await Geolocator.getCurrentPosition();
  }

  Future<bool> canCheckBiometrics() async {
    return (await _localAuth.canCheckBiometrics) || (await _localAuth.isDeviceSupported());
  }

  Future<bool> authenticateBiometric({String reason = 'Verifikasi identitas Anda'}) async {
    try {
      return await _localAuth.authenticate(
        localizedReason: reason,
        options: const AuthenticationOptions(
          stickyAuth: true,
          biometricOnly: false,
        ),
      );
    } catch (_) {
      return false;
    }
  }
}
