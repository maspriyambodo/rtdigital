import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'screens.dart';
import '../contribution_screens.dart';
import 'attachment_viewer_screen.dart';
import 'login_screen.dart';
import '../letter_screens.dart';
import '../complaint_screens.dart';
import '../security_screens.dart';
import '../savings_screens.dart';
import '../asset_screens.dart';

final GlobalKey<NavigatorState> _rootNavigatorKey = GlobalKey<NavigatorState>();

final appRouter = GoRouter(
  navigatorKey: _rootNavigatorKey,
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: '/attachment-viewer',
      builder: (context, state) {
        final url = state.uri.queryParameters['url'] ?? '';
        final title = state.uri.queryParameters['title'] ?? 'Lampiran';
        final type = state.uri.queryParameters['type'] ?? 'image';
        return AttachmentViewerScreen(url: url, title: title, type: type);
      },
    ),
    GoRoute(
      path: '/warga/iuran',
      builder: (context, state) => const WargaIuranScreen(),
    ),
    GoRoute(
      path: '/warga/kas-rt',
      builder: (context, state) => const WargaKasRtScreen(),
    ),
    GoRoute(
      path: '/warga/tabungan',
      builder: (context, state) => const SavingsScreen(),
    ),
    GoRoute(
      path: '/warga/aset',
      builder: (context, state) => const AssetsScreen(),
    ),
    GoRoute(
      path: '/warga/surat/baru',
      builder: (context, state) => const WargaLetterFormScreen(),
    ),
    GoRoute(
      path: '/warga/surat/tracking',
      builder: (context, state) => const WargaLetterTrackingScreen(),
    ),
    GoRoute(
      path: '/warga/surat/viewer',
      builder: (context, state) {
        final id = state.uri.queryParameters['id'] ?? '';
        final number = state.uri.queryParameters['number'] ?? '001';
        return WargaLetterPdfViewerScreen(requestId: id, letterNumber: number);
      },
    ),
    GoRoute(
      path: '/warga/aduan/baru',
      builder: (context, state) => const WargaComplaintFormScreen(),
    ),
    GoRoute(
      path: '/warga/aduan/list',
      builder: (context, state) => const WargaComplaintListScreen(),
    ),
    GoRoute(
      path: '/warga/aduan/timeline',
      builder: (context, state) {
        final id = state.uri.queryParameters['id'] ?? '';
        return WargaComplaintTimelineScreen(complaintId: id);
      },
    ),
    GoRoute(
      path: '/warga/panic',
      builder: (context, state) => const PanicButtonScreen(),
    ),
    GoRoute(
      path: '/pengurus/surat/approval',
      builder: (context, state) => const PengurusApprovalScreen(),
    ),
    GoRoute(
      path: '/pengurus/aduan/dashboard',
      builder: (context, state) => const PengurusComplaintDashboardScreen(),
    ),
    GoRoute(
      path: '/pengurus/keamanan/dashboard',
      builder: (context, state) => const SecurityDashboardAlertsScreen(),
    ),

    ShellRoute(
      builder: (context, state, child) => WargaShell(child: child),
      routes: [
        GoRoute(
          path: '/warga/beranda',
          builder: (context, state) => const WargaHomeScreen(),
        ),
        GoRoute(
          path: '/warga/layanan',
          builder: (context, state) => const WargaLayananScreen(),
        ),
        GoRoute(
          path: '/warga/aktivitas',
          builder: (context, state) => const WargaAktivitasScreen(),
        ),
        GoRoute(
          path: '/warga/notifikasi',
          builder: (context, state) => const WargaNotifikasiScreen(),
        ),
        GoRoute(
          path: '/warga/profil',
          builder: (context, state) => const WargaProfilScreen(),
        ),
      ],
    ),
    ShellRoute(
      builder: (context, state, child) => PengurusShell(child: child),
      routes: [
        GoRoute(
          path: '/pengurus/dashboard',
          builder: (context, state) => const PengurusDashboardScreen(),
        ),
        GoRoute(
          path: '/pengurus/verifikasi',
          builder: (context, state) => const PengurusVerifikasiScreen(),
        ),
      ],
    ),
  ],
);


class WargaShell extends StatelessWidget {
  final Widget child;

  const WargaShell({super.key, required this.child});

  int _calculateSelectedIndex(BuildContext context) {
    final String location = GoRouterState.of(context).uri.path;
    if (location.startsWith('/warga/layanan')) return 1;
    if (location.startsWith('/warga/aktivitas')) return 2;
    if (location.startsWith('/warga/notifikasi')) return 3;
    if (location.startsWith('/warga/profil')) return 4;
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    final currentIndex = _calculateSelectedIndex(context);
    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: currentIndex,
        onDestinationSelected: (int index) {
          switch (index) {
            case 0:
              context.go('/warga/beranda');
              break;
            case 1:
              context.go('/warga/layanan');
              break;
            case 2:
              context.go('/warga/aktivitas');
              break;
            case 3:
              context.go('/warga/notifikasi');
              break;
            case 4:
              context.go('/warga/profil');
              break;
          }
        },
        destinations: const [
          NavigationDestination(icon: Icon(Icons.home_outlined), selectedIcon: Icon(Icons.home), label: 'Beranda'),
          NavigationDestination(icon: Icon(Icons.grid_view_outlined), selectedIcon: Icon(Icons.grid_view), label: 'Layanan'),
          NavigationDestination(icon: Icon(Icons.receipt_long_outlined), selectedIcon: Icon(Icons.receipt_long), label: 'Aktivitas'),
          NavigationDestination(icon: Icon(Icons.notifications_outlined), selectedIcon: Icon(Icons.notifications), label: 'Notifikasi'),
          NavigationDestination(icon: Icon(Icons.person_outline), selectedIcon: Icon(Icons.person), label: 'Profil'),
        ],
      ),
    );
  }
}

class PengurusShell extends StatelessWidget {
  final Widget child;
  const PengurusShell({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Panel Operasional Pengurus'),
        actions: [
          IconButton(
            icon: const Icon(Icons.swap_horiz),
            tooltip: 'Ke Tampilan Warga',
            onPressed: () => context.go('/warga/beranda'),
          ),
        ],
      ),
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            DrawerHeader(
              decoration: const BoxDecoration(color: Color(0xFF1B5E20)),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Image.asset(
                    'assets/images/logo_256x256.png',
                    width: 48,
                    height: 48,
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Menu Pengurus RT',
                    style: TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
            ),
            ListTile(
              leading: const Icon(Icons.dashboard),
              title: const Text('Dashboard Approval'),
              onTap: () {
                Navigator.pop(context);
                context.go('/pengurus/dashboard');
              },
            ),
            ListTile(
              leading: const Icon(Icons.qr_code_scanner),
              title: const Text('Verifikasi Lapangan'),
              onTap: () {
                Navigator.pop(context);
                context.go('/pengurus/verifikasi');
              },
            ),
            ListTile(
              leading: const Icon(Icons.assignment_turned_in),
              title: const Text('Approval Surat Warga'),
              onTap: () {
                Navigator.pop(context);
                context.go('/pengurus/surat/approval');
              },
            ),
            ListTile(
              leading: const Icon(Icons.report_problem_outlined),
              title: const Text('Dashboard Tiket Aduan'),
              onTap: () {
                Navigator.pop(context);
                context.go('/pengurus/aduan/dashboard');
              },
            ),
            ListTile(
              leading: const Icon(Icons.security),
              title: const Text('Dashboard Siskamling & Alert'),
              onTap: () {
                Navigator.pop(context);
                context.go('/pengurus/keamanan/dashboard');
              },
            ),
          ],
        ),
      ),
      body: child,
    );
  }
}
