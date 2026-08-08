import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/main.dart';
import 'package:mobile/core/router/login_screen.dart';

void main() {
  testWidgets('App renders main shell with title', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: RtDigitalMobileApp(),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    expect(find.byType(LoginScreen), findsOneWidget);
  });
}


