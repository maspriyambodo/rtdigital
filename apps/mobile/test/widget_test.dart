import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/main.dart';

void main() {
  testWidgets('App renders main shell with title', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: RtDigitalMobileApp(),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    expect(find.text('RT Digital'), findsWidgets);
  });
}


