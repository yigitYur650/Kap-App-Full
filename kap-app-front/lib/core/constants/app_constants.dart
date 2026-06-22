// lib/core/constants/app_constants.dart

const String kAppName = 'Kap-App';
const String kAppVersion = '0.1.0';

// API
const String kApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://192.168.0.8:8080/api/v1',
);

// Supported locales
const List<String> kSupportedLocales = ['tr', 'en'];
const String kDefaultLocale = 'tr';

// Storage keys
const String kTokenKey = 'access_token';
const String kRefreshTokenKey = 'refresh_token';
const String kLocaleKey = 'app_locale';
const String kThemeKey = 'app_theme';
