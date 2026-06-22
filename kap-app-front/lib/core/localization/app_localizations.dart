// lib/core/localization/app_localizations.dart
//
// Basit JSON tabanlı lokalizasyon servisi.
// flutter_localizations paketi ile entegre çalışır.

import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class AppLocalizations {
  final Locale locale;
  late Map<String, dynamic> _strings;

  AppLocalizations(this.locale);

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  Future<bool> load() async {
    final jsonStr = await rootBundle
        .loadString('assets/lang/${locale.languageCode}.json');
    _strings = json.decode(jsonStr) as Map<String, dynamic>;
    return true;
  }

  /// Nokta notasyonunu destekler: t('auth.login')
  String t(String key) {
    final parts = key.split('.');
    dynamic current = _strings;
    for (final part in parts) {
      if (current is Map<String, dynamic>) {
        current = current[part];
      } else {
        return key;
      }
    }
    return current?.toString() ?? key;
  }
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) =>
      ['tr', 'en'].contains(locale.languageCode);

  @override
  Future<AppLocalizations> load(Locale locale) async {
    final localizations = AppLocalizations(locale);
    await localizations.load();
    return localizations;
  }

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}
