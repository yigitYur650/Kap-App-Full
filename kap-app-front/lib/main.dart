// lib/main.dart

import 'package:flutter/material.dart' as m;
import 'package:shadcn_flutter/shadcn_flutter.dart';
import 'package:flutter_localizations/flutter_localizations.dart' as loc;
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'core/localization/app_localizations.dart';
import 'core/constants/app_constants.dart';
import 'core/network/api_client.dart';
import 'features/product/data/product_repository.dart';
import 'features/product/providers/product_provider.dart';
import 'features/auth/data/auth_repository.dart';
import 'features/auth/providers/auth_provider.dart';
import 'features/auth/presentation/screens/login_screen.dart';
import 'features/settings/presentation/screens/settings_screen.dart';
import 'features/home/presentation/screens/main_layout.dart';

void main() async {
  m.WidgetsFlutterBinding.ensureInitialized();
  
  final prefs = await SharedPreferences.getInstance();
  final token = prefs.getString(kTokenKey);
  final userId = prefs.getString('user_id');

  final apiClient = ApiClient();
  if (token != null) {
    apiClient.setToken(token);
  }

  final authRepo = AuthRepository(apiClient);
  final productRepo = ProductRepository(apiClient);
  final tenantId = prefs.getString('current_tenant_id');

  runApp(KappApp(
    apiClient: apiClient,
    authRepo: authRepo,
    productRepo: productRepo,
    initialUserId: token != null ? userId : null,
    initialTenantId: tenantId,
  ));
}

class KappApp extends m.StatelessWidget {
  final ApiClient apiClient;
  final AuthRepository authRepo;
  final ProductRepository productRepo;
  final String? initialUserId;
  final String? initialTenantId;

  const KappApp({
    super.key,
    required this.apiClient,
    required this.authRepo,
    required this.productRepo,
    this.initialUserId,
    this.initialTenantId,
  });

  @override
  m.Widget build(m.BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(
          create: (_) => AuthProvider(
            authRepo,
            apiClient,
            initialUserId: initialUserId,
            initialTenantId: initialTenantId,
          ),
        ),
        ChangeNotifierProvider(
          create: (_) => ProductProvider(productRepo),
        ),
      ],
      child: m.ValueListenableBuilder<String>(
        valueListenable: globalSelectedLang,
        builder: (context, lang, child) {
          final authProvider = Provider.of<AuthProvider>(context);
          
          return ShadcnApp(
            title: 'Kap-App',
            debugShowCheckedModeBanner: false,
            
            // Shadcn Özel Teması
            theme: ThemeData(
              colorScheme: ColorSchemes.darkZinc, 
            ),

            // ── Lokalizasyon ─────────────────────────────────────────────────────
            locale: m.Locale(lang),
            supportedLocales: const [m.Locale('tr'), m.Locale('en')],
            localizationsDelegates: const [
              AppLocalizations.delegate,
              _TrShadcnLocalizationsDelegate(),
              loc.GlobalMaterialLocalizations.delegate,
              loc.GlobalWidgetsLocalizations.delegate,
              loc.GlobalCupertinoLocalizations.delegate,
            ],

            // ── Ana Ekran (Kimlik doğrulamaya göre yönlendirme) ──────────────────
            home: authProvider.isAuthenticated
                ? const MainLayout()
                : const LoginScreen(),
          );
        },
      ),
    );
  }
}

class _TrShadcnLocalizationsDelegate extends m.LocalizationsDelegate<ShadcnLocalizations> {
  const _TrShadcnLocalizationsDelegate();

  @override
  bool isSupported(m.Locale locale) => locale.languageCode == 'tr';

  @override
  Future<ShadcnLocalizations> load(m.Locale locale) {
    return ShadcnLocalizations.delegate.load(const m.Locale('en'));
  }

  @override
  bool shouldReload(m.LocalizationsDelegate<ShadcnLocalizations> old) => false;
}

