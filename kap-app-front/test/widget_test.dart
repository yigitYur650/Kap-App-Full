import 'package:flutter_test/flutter_test.dart';
import 'package:kapp_app/core/network/api_client.dart';
import 'package:kapp_app/features/auth/data/auth_repository.dart';
import 'package:kapp_app/features/product/data/product_repository.dart';
import 'package:kapp_app/main.dart';

void main() {
  testWidgets('App load test', (WidgetTester tester) async {
    final apiClient = ApiClient();
    final authRepo = AuthRepository(apiClient);
    final productRepo = ProductRepository(apiClient);
    await tester.pumpWidget(KappApp(apiClient: apiClient, authRepo: authRepo, productRepo: productRepo));
    expect(find.byType(KappApp), findsOneWidget);
  });
}
