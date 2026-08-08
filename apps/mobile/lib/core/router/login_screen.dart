import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../auth_provider.dart';
import '../widgets/ui_components.dart';


class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _identifierController = TextEditingController();
  final _passwordController = TextEditingController();
  final _pinController = TextEditingController();
  bool _usePinFallback = false;

  @override
  void dispose() {
    _identifierController.dispose();
    _passwordController.dispose();
    _pinController.dispose();
    super.dispose();
  }

  Future<void> _handleLogin() async {
    if (!_formKey.currentState!.validate()) return;
    final success = await ref.read(authProvider.notifier).login(
          _identifierController.text.trim(),
          _passwordController.text,
        );
    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Berhasil masuk')),
      );
    }
  }

  Future<void> _handleBiometricLogin() async {
    final success = await ref.read(authProvider.notifier).loginWithBiometrics();
    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Berhasil masuk via Biometrik')),
      );
    }
  }

  Future<void> _handlePinLogin() async {
    if (_pinController.text.length < 4) return;
    final success = await ref.read(authProvider.notifier).verifyPin(_pinController.text);
    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Berhasil masuk via PIN')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {

    final authState = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Masuk RT Digital')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const SizedBox(height: 16),
              Image.asset(
                'assets/images/logo_256x256.png',
                height: 72,
                width: 72,
              ),
              const SizedBox(height: 16),
              Text(
                'Selamat Datang',
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                'Masukkan nomor telepon/email dan kata sandi Anda',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: Colors.grey),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),
              if (authState.error != null) ...[
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.red.shade50,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.red.shade200),
                  ),
                  child: Text(
                    authState.error!,
                    style: TextStyle(color: Colors.red.shade800),
                  ),
                ),
                const SizedBox(height: 16),
              ],
              if (!_usePinFallback) ...[
                AppTextField(
                  label: 'Nomor Telepon / Email',
                  hint: 'Contoh: 08123456789 atau warga@rt.id',
                  controller: _identifierController,
                  keyboardType: TextInputType.emailAddress,
                  validator: (val) {
                    if (val == null || val.trim().isEmpty) return 'Nomor telepon/email wajib diisi';
                    return null;
                  },
                ),
                const SizedBox(height: 16),
                AppTextField(
                  label: 'Kata Sandi',
                  hint: 'Masukkan kata sandi',
                  obscureText: true,
                  controller: _passwordController,
                  validator: (val) {
                    if (val == null || val.isEmpty) return 'Kata sandi wajib diisi';
                    return null;
                  },
                ),
                const SizedBox(height: 24),
                AppButton(
                  label: 'Masuk Akun',
                  isLoading: authState.isLoading,
                  onPressed: _handleLogin,
                ),
                const SizedBox(height: 16),
                AppButton(
                  label: 'Masuk dengan Biometrik',
                  isSecondary: true,
                  icon: Icons.fingerprint,
                  onPressed: _handleBiometricLogin,
                ),
                const SizedBox(height: 8),
                TextButton(
                  onPressed: () {
                    setState(() {
                      _usePinFallback = true;
                    });
                  },
                  child: const Text('Gunakan Fallback PIN Perangkat'),
                ),
              ] else ...[
                AppTextField(
                  label: 'PIN Keamanan (Fallback)',
                  hint: 'Masukkan 6 digit PIN',
                  obscureText: true,
                  keyboardType: TextInputType.number,
                  controller: _pinController,
                  validator: (val) {
                    if (val == null || val.length < 4) return 'PIN tidak valid';
                    return null;
                  },
                ),
                const SizedBox(height: 24),
                AppButton(
                  label: 'Verifikasi PIN',
                  isLoading: authState.isLoading,
                  onPressed: _handlePinLogin,
                ),
                const SizedBox(height: 12),
                TextButton(
                  onPressed: () {
                    setState(() {
                      _usePinFallback = false;
                    });
                  },
                  child: const Text('Kembali ke Login Kata Sandi'),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

